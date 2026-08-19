// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package entitycore

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// idParamTestModel is a minimal KibanaResourceModel fixture for exercising
// [DeleteByIDParams], [ReadByIDParams], and
// [ReadByIDParamsWithAgnosticNamespaceRetry] without a security exception/list
// package's full model.
type idParamTestModel struct {
	ResourceTimeoutsField
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	SpaceID       types.String `tfsdk:"space_id"`
	NamespaceType types.String `tfsdk:"namespace_type"`
	Populated     types.String `tfsdk:"populated"`
}

func (m idParamTestModel) GetID() types.String             { return m.ID }
func (m idParamTestModel) GetResourceID() types.String     { return m.Name }
func (m idParamTestModel) GetSpaceID() types.String        { return m.SpaceID }
func (m idParamTestModel) GetKibanaConnection() types.List { return types.ListNull(types.StringType) }

type idParamTestParams struct {
	ID            *string
	NamespaceType *string
}

type idParamTestResponse struct {
	ID string
}

func newIDParamTestClient() *clients.KibanaScopedClient {
	return &clients.KibanaScopedClient{}
}

func TestDeleteByIDParams(t *testing.T) {
	t.Run("builds params from resourceID and model and delegates to apiDelete", func(t *testing.T) {
		var gotSpaceID string
		var gotParams *idParamTestParams

		apiDelete := func(_ context.Context, _ *kibanaoapi.Client, spaceID string, params *idParamTestParams) diag.Diagnostics {
			gotSpaceID = spaceID
			gotParams = params
			return nil
		}

		deleteFn := DeleteByIDParams[idParamTestModel](
			func(id string, model idParamTestModel) *idParamTestParams {
				nsType := model.NamespaceType.ValueString()
				return &idParamTestParams{ID: &id, NamespaceType: &nsType}
			},
			apiDelete,
		)

		model := idParamTestModel{NamespaceType: types.StringValue("agnostic")}
		diags := deleteFn(context.Background(), newIDParamTestClient(), "item-1", "default", model)

		require.False(t, diags.HasError())
		require.Equal(t, "default", gotSpaceID)
		require.Equal(t, "item-1", *gotParams.ID)
		require.Equal(t, "agnostic", *gotParams.NamespaceType)
	})

	t.Run("propagates diagnostics from apiDelete", func(t *testing.T) {
		apiDelete := func(_ context.Context, _ *kibanaoapi.Client, _ string, _ *idParamTestParams) diag.Diagnostics {
			var diags diag.Diagnostics
			diags.AddError("boom", "delete failed")
			return diags
		}

		deleteFn := DeleteByIDParams[idParamTestModel](
			func(id string, _ idParamTestModel) *idParamTestParams { return &idParamTestParams{ID: &id} },
			apiDelete,
		)

		diags := deleteFn(context.Background(), newIDParamTestClient(), "item-1", "default", idParamTestModel{})

		require.True(t, diags.HasError())
	})
}

func TestReadByIDParams(t *testing.T) {
	t.Run("not found returns false without calling populate", func(t *testing.T) {
		apiGet := func(_ context.Context, _ *kibanaoapi.Client, _ string, _ *idParamTestParams) (*idParamTestResponse, diag.Diagnostics) {
			return nil, nil
		}
		populateCalled := false

		readFn := ReadByIDParams[idParamTestModel, idParamTestParams, idParamTestResponse](
			func(id string) *idParamTestParams { return &idParamTestParams{ID: &id} },
			apiGet,
			func(_ *idParamTestModel, _ context.Context, _ string, _ *idParamTestResponse) diag.Diagnostics {
				populateCalled = true
				return nil
			},
		)

		_, found, diags := readFn(context.Background(), newIDParamTestClient(), "item-1", "default", idParamTestModel{})

		require.False(t, diags.HasError())
		require.False(t, found)
		require.False(t, populateCalled)
	})

	t.Run("found calls populate with spaceID and data", func(t *testing.T) {
		apiGet := func(_ context.Context, _ *kibanaoapi.Client, _ string, params *idParamTestParams) (*idParamTestResponse, diag.Diagnostics) {
			return &idParamTestResponse{ID: *params.ID}, nil
		}

		readFn := ReadByIDParams[idParamTestModel, idParamTestParams, idParamTestResponse](
			func(id string) *idParamTestParams { return &idParamTestParams{ID: &id} },
			apiGet,
			func(model *idParamTestModel, _ context.Context, spaceID string, data *idParamTestResponse) diag.Diagnostics {
				model.SpaceID = types.StringValue(spaceID)
				model.Populated = types.StringValue(data.ID)
				return nil
			},
		)

		got, found, diags := readFn(context.Background(), newIDParamTestClient(), "item-1", "default", idParamTestModel{})

		require.False(t, diags.HasError())
		require.True(t, found)
		require.Equal(t, "default", got.SpaceID.ValueString())
		require.Equal(t, "item-1", got.Populated.ValueString())
	})
}

func TestReadByIDParamsWithAgnosticNamespaceRetry(t *testing.T) {
	newParams := func(id string, namespaceType types.String) *idParamTestParams {
		params := &idParamTestParams{ID: &id}
		if !namespaceType.IsNull() && !namespaceType.IsUnknown() && namespaceType.ValueString() != "" {
			nsType := namespaceType.ValueString()
			params.NamespaceType = &nsType
		}
		return params
	}
	setAgnostic := func(params *idParamTestParams) {
		agnostic := "agnostic"
		params.NamespaceType = &agnostic
	}
	populate := func(model *idParamTestModel, _ context.Context, spaceID string, data *idParamTestResponse) diag.Diagnostics {
		model.SpaceID = types.StringValue(spaceID)
		model.Populated = types.StringValue(data.ID)
		return nil
	}

	t.Run("known namespace type not found does not retry", func(t *testing.T) {
		callCount := 0
		apiGet := func(_ context.Context, _ *kibanaoapi.Client, _ string, _ *idParamTestParams) (*idParamTestResponse, diag.Diagnostics) {
			callCount++
			return nil, nil
		}

		readFn := ReadByIDParamsWithAgnosticNamespaceRetry[idParamTestModel, idParamTestParams, idParamTestResponse](
			func(model idParamTestModel) types.String { return model.NamespaceType },
			newParams,
			setAgnostic,
			apiGet,
			populate,
		)

		_, found, diags := readFn(context.Background(), newIDParamTestClient(), "item-1", "default", idParamTestModel{NamespaceType: types.StringValue("single")})

		require.False(t, diags.HasError())
		require.False(t, found)
		require.Equal(t, 1, callCount)
	})

	t.Run("unknown namespace type not found retries with agnostic", func(t *testing.T) {
		var namespaceTypesSeen []string
		apiGet := func(_ context.Context, _ *kibanaoapi.Client, _ string, params *idParamTestParams) (*idParamTestResponse, diag.Diagnostics) {
			if params.NamespaceType != nil {
				namespaceTypesSeen = append(namespaceTypesSeen, *params.NamespaceType)
			} else {
				namespaceTypesSeen = append(namespaceTypesSeen, "")
			}
			if len(namespaceTypesSeen) == 1 {
				return nil, nil
			}
			return &idParamTestResponse{ID: *params.ID}, nil
		}

		readFn := ReadByIDParamsWithAgnosticNamespaceRetry[idParamTestModel, idParamTestParams, idParamTestResponse](
			func(model idParamTestModel) types.String { return model.NamespaceType },
			newParams,
			setAgnostic,
			apiGet,
			populate,
		)

		got, found, diags := readFn(context.Background(), newIDParamTestClient(), "item-1", "default", idParamTestModel{})

		require.False(t, diags.HasError())
		require.True(t, found)
		require.Equal(t, []string{"", "agnostic"}, namespaceTypesSeen)
		require.Equal(t, "item-1", got.Populated.ValueString())
	})

	t.Run("unknown namespace type still not found after retry returns not found", func(t *testing.T) {
		apiGet := func(_ context.Context, _ *kibanaoapi.Client, _ string, _ *idParamTestParams) (*idParamTestResponse, diag.Diagnostics) {
			return nil, nil
		}

		readFn := ReadByIDParamsWithAgnosticNamespaceRetry[idParamTestModel, idParamTestParams, idParamTestResponse](
			func(model idParamTestModel) types.String { return model.NamespaceType },
			newParams,
			setAgnostic,
			apiGet,
			populate,
		)

		_, found, diags := readFn(context.Background(), newIDParamTestClient(), "item-1", "default", idParamTestModel{})

		require.False(t, diags.HasError())
		require.False(t, found)
	})

	t.Run("error on initial read short-circuits without retry", func(t *testing.T) {
		callCount := 0
		apiGet := func(_ context.Context, _ *kibanaoapi.Client, _ string, _ *idParamTestParams) (*idParamTestResponse, diag.Diagnostics) {
			callCount++
			var diags diag.Diagnostics
			diags.AddError("boom", "read failed")
			return nil, diags
		}

		readFn := ReadByIDParamsWithAgnosticNamespaceRetry[idParamTestModel, idParamTestParams, idParamTestResponse](
			func(model idParamTestModel) types.String { return model.NamespaceType },
			newParams,
			setAgnostic,
			apiGet,
			populate,
		)

		_, found, diags := readFn(context.Background(), newIDParamTestClient(), "item-1", "default", idParamTestModel{})

		require.True(t, diags.HasError())
		require.False(t, found)
		require.Equal(t, 1, callCount)
	})
}
