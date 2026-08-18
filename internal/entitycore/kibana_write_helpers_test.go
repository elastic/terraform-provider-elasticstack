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

type simpleKibanaWriteTestBody struct {
	Name string
}

type simpleKibanaWriteTestResp struct {
	ID string
}

func simpleKibanaWriteTestBuildRequest(diags diag.Diagnostics) func(context.Context, KibanaWriteRequest[testKibanaResourceModel]) (*simpleKibanaWriteTestBody, diag.Diagnostics) {
	return func(_ context.Context, req KibanaWriteRequest[testKibanaResourceModel]) (*simpleKibanaWriteTestBody, diag.Diagnostics) {
		if diags.HasError() {
			return nil, diags
		}
		return &simpleKibanaWriteTestBody{Name: req.Plan.Name.ValueString()}, nil
	}
}

type simpleKibanaWriteTestAPICallFunc func(context.Context, *kibanaoapi.Client, string, simpleKibanaWriteTestBody) (*simpleKibanaWriteTestResp, diag.Diagnostics)

func simpleKibanaWriteTestAPICall(resp *simpleKibanaWriteTestResp, diags diag.Diagnostics) simpleKibanaWriteTestAPICallFunc {
	return func(_ context.Context, _ *kibanaoapi.Client, _ string, _ simpleKibanaWriteTestBody) (*simpleKibanaWriteTestResp, diag.Diagnostics) {
		return resp, diags
	}
}

func TestRequireNonNilKibanaWriteResponse(t *testing.T) {
	t.Run("non-nil response reports false and appends no diagnostics", func(t *testing.T) {
		var diags diag.Diagnostics
		resp := "created"

		nilResp := RequireNonNilKibanaWriteResponse(&diags, &resp, "create", "security list")

		require.False(t, nilResp)
		require.False(t, diags.HasError())
	})

	t.Run("nil response reports true and appends the standard error", func(t *testing.T) {
		var diags diag.Diagnostics
		var resp *string

		nilResp := RequireNonNilKibanaWriteResponse(&diags, resp, "update", "exception item")

		require.True(t, nilResp)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Failed to update exception item", diags[0].Summary())
		require.Equal(t, "API returned empty response", diags[0].Detail())
	})
}

func TestKibanaResourceID(t *testing.T) {
	got := KibanaResourceID("default", "abc-123")
	require.Equal(t, types.StringValue("default/abc-123"), got)
}

func TestSimpleKibanaCreate(t *testing.T) {
	writeReq := KibanaWriteRequest[testKibanaResourceModel]{
		Plan:    testKibanaResourceModel{Name: types.StringValue("widget-1")},
		SpaceID: "default",
	}

	t.Run("populates the model from the response on success", func(t *testing.T) {
		writeFn := SimpleKibanaCreate(
			"widget",
			simpleKibanaWriteTestBuildRequest(nil),
			simpleKibanaWriteTestAPICall(&simpleKibanaWriteTestResp{ID: "created-1"}, nil),
			func(m *testKibanaResourceModel, spaceID string, resp *simpleKibanaWriteTestResp) {
				m.ID = KibanaResourceID(spaceID, resp.ID)
			},
		)

		result, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, writeReq)

		require.False(t, diags.HasError())
		require.Equal(t, types.StringValue("default/created-1"), result.Model.ID)
	})

	t.Run("allows a nil populate when the response carries nothing the model needs", func(t *testing.T) {
		writeFn := SimpleKibanaCreate(
			"widget",
			simpleKibanaWriteTestBuildRequest(nil),
			simpleKibanaWriteTestAPICall(&simpleKibanaWriteTestResp{ID: "created-1"}, nil),
			nil,
		)

		result, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, writeReq)

		require.False(t, diags.HasError())
		require.Equal(t, writeReq.Plan, result.Model)
	})

	t.Run("returns early with buildRequest diagnostics", func(t *testing.T) {
		var buildDiags diag.Diagnostics
		buildDiags.AddError("bad plan", "could not build request")

		called := false
		writeFn := SimpleKibanaCreate(
			"widget",
			simpleKibanaWriteTestBuildRequest(buildDiags),
			func(_ context.Context, _ *kibanaoapi.Client, _ string, _ simpleKibanaWriteTestBody) (*simpleKibanaWriteTestResp, diag.Diagnostics) {
				called = true
				return &simpleKibanaWriteTestResp{}, nil
			},
			nil,
		)

		result, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, writeReq)

		require.False(t, called)
		require.True(t, diags.HasError())
		require.Equal(t, KibanaWriteResult[testKibanaResourceModel]{}, result)
	})

	t.Run("returns early with apiCreate diagnostics", func(t *testing.T) {
		var apiDiags diag.Diagnostics
		apiDiags.AddError("api error", "create failed")

		writeFn := SimpleKibanaCreate(
			"widget",
			simpleKibanaWriteTestBuildRequest(nil),
			simpleKibanaWriteTestAPICall(nil, apiDiags),
			nil,
		)

		result, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, writeReq)

		require.True(t, diags.HasError())
		require.Equal(t, KibanaWriteResult[testKibanaResourceModel]{}, result)
	})

	t.Run("surfaces the standard error when the API returns a nil response", func(t *testing.T) {
		writeFn := SimpleKibanaCreate(
			"widget",
			simpleKibanaWriteTestBuildRequest(nil),
			simpleKibanaWriteTestAPICall(nil, nil),
			nil,
		)

		result, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, writeReq)

		require.True(t, diags.HasError())
		require.Equal(t, "Failed to create widget", diags[0].Summary())
		require.Equal(t, KibanaWriteResult[testKibanaResourceModel]{}, result)
	})
}

func TestSimpleKibanaUpdate(t *testing.T) {
	writeReq := KibanaWriteRequest[testKibanaResourceModel]{
		Plan:    testKibanaResourceModel{Name: types.StringValue("widget-1")},
		SpaceID: "default",
	}

	t.Run("populates the model from the response on success", func(t *testing.T) {
		writeFn := SimpleKibanaUpdate(
			"widget",
			simpleKibanaWriteTestBuildRequest(nil),
			simpleKibanaWriteTestAPICall(&simpleKibanaWriteTestResp{ID: "updated-1"}, nil),
			func(m *testKibanaResourceModel, spaceID string, resp *simpleKibanaWriteTestResp) {
				m.ID = KibanaResourceID(spaceID, resp.ID)
			},
		)

		result, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, writeReq)

		require.False(t, diags.HasError())
		require.Equal(t, types.StringValue("default/updated-1"), result.Model.ID)
	})

	t.Run("surfaces the standard error when the API returns a nil response", func(t *testing.T) {
		writeFn := SimpleKibanaUpdate(
			"widget",
			simpleKibanaWriteTestBuildRequest(nil),
			simpleKibanaWriteTestAPICall(nil, nil),
			nil,
		)

		result, diags := writeFn(context.Background(), &clients.KibanaScopedClient{}, writeReq)

		require.True(t, diags.HasError())
		require.Equal(t, "Failed to update widget", diags[0].Summary())
		require.Equal(t, KibanaWriteResult[testKibanaResourceModel]{}, result)
	})
}
