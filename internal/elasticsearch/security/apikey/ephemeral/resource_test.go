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

package ephemeral

import (
	"context"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwephemeral "github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestEphemeralSchemaNameValidation(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())
	nameAttr, diags := s.AttributeAtPath(context.Background(), path.Root("name"))
	require.False(t, diags.HasError())

	stringAttr, ok := nameAttr.(eschema.StringAttribute)
	require.True(t, ok)
	require.Len(t, stringAttr.Validators, 2)

	testCases := []struct {
		name        string
		value       string
		expectError bool
	}{
		{name: "valid name", value: "app-key", expectError: false},
		{name: "empty name", value: "", expectError: true},
		{name: "too long name", value: strings.Repeat("a", 1025), expectError: true},
		{name: "non-printable name", value: "bad\tname", expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			config := tfsdk.Config{
				Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				}}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, testCase.value),
				}),
				Schema: eschema.Schema{
					Attributes: map[string]eschema.Attribute{
						"name": stringAttr,
					},
				},
			}

			request := validator.StringRequest{
				Path:        path.Root("name"),
				ConfigValue: types.StringValue(testCase.value),
				Config:      config,
			}
			response := &validator.StringResponse{}
			for _, v := range stringAttr.Validators {
				v.ValidateString(context.Background(), request, response)
			}

			if testCase.expectError {
				require.True(t, response.Diagnostics.HasError())
				return
			}
			require.False(t, response.Diagnostics.HasError())
		})
	}
}

func TestEphemeralSchemaTypeValidation(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())
	typeAttr, diags := s.AttributeAtPath(context.Background(), path.Root("type"))
	require.False(t, diags.HasError())

	stringAttr := typeAttr.(eschema.StringAttribute)

	testCases := []struct {
		name        string
		value       string
		expectError bool
	}{
		{name: "rest", value: "rest", expectError: false},
		{name: "cross_cluster", value: "cross_cluster", expectError: false},
		{name: "invalid", value: "invalid", expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				Path:        path.Root("type"),
				ConfigValue: types.StringValue(testCase.value),
			}
			response := &validator.StringResponse{}
			for _, v := range stringAttr.Validators {
				v.ValidateString(context.Background(), request, response)
			}

			if testCase.expectError {
				require.True(t, response.Diagnostics.HasError())
				return
			}
			require.False(t, response.Diagnostics.HasError())
		})
	}
}

func TestEphemeralSchemaRequiresTypeValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		typeValue   string
		attrValue   string
		expectError bool
	}{
		{
			name:        "role_descriptors with rest",
			typeValue:   "rest",
			attrValue:   `{"role": {"cluster": ["all"]}}`,
			expectError: false,
		},
		{
			name:        "role_descriptors with cross_cluster",
			typeValue:   "cross_cluster",
			attrValue:   `{"role": {"cluster": ["all"]}}`,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			config := tfsdk.Config{
				Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"type":             tftypes.String,
					"role_descriptors": tftypes.String,
				}}, map[string]tftypes.Value{
					"type":             tftypes.NewValue(tftypes.String, testCase.typeValue),
					"role_descriptors": tftypes.NewValue(tftypes.String, testCase.attrValue),
				}),
				Schema: eschema.Schema{
					Attributes: map[string]eschema.Attribute{
						"type":             eschema.StringAttribute{},
						"role_descriptors": eschema.StringAttribute{},
					},
				},
			}

			request := validator.StringRequest{
				Path:        path.Root("role_descriptors"),
				ConfigValue: types.StringValue(testCase.attrValue),
				Config:      config,
			}
			response := &validator.StringResponse{}
			apikey.RequiresType("rest").ValidateString(context.Background(), request, response)
			if testCase.expectError {
				require.True(t, response.Diagnostics.HasError())
				return
			}
			require.False(t, response.Diagnostics.HasError())
		})
	}
}

func TestEphemeralSchemaAccessRequiresTypeValidation(t *testing.T) {
	t.Parallel()

	accessValue := types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{})

	testCases := []struct {
		name        string
		typeValue   *string
		expectError bool
	}{
		{name: "access with rest", typeValue: new("rest"), expectError: true},
		{name: "access with unset type", typeValue: nil, expectError: true},
		{name: "access with cross_cluster", typeValue: new("cross_cluster"), expectError: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			configValues := map[string]tftypes.Value{
				"access": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
			}
			attrTypes := map[string]tftypes.Type{
				"access": tftypes.Object{AttributeTypes: map[string]tftypes.Type{}},
			}
			if testCase.typeValue != nil {
				configValues["type"] = tftypes.NewValue(tftypes.String, *testCase.typeValue)
				attrTypes["type"] = tftypes.String
			}

			config := tfsdk.Config{
				Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypes}, configValues),
				Schema: eschema.Schema{
					Attributes: map[string]eschema.Attribute{
						"type":   eschema.StringAttribute{Optional: true},
						"access": eschema.SingleNestedAttribute{},
					},
				},
			}

			request := validator.ObjectRequest{
				Path:        path.Root("access"),
				ConfigValue: accessValue,
				Config:      config,
			}
			response := &validator.ObjectResponse{}
			apikey.RequiresType(apikey.CrossClusterAPIKeyType).ValidateObject(context.Background(), request, response)

			if testCase.expectError {
				require.True(t, response.Diagnostics.HasError())
				return
			}
			require.False(t, response.Diagnostics.HasError())
		})
	}
}

func TestCloseAPIKey(t *testing.T) {
	t.Parallel()

	t.Run("does not call delete when invalidate_on_close is false", func(t *testing.T) {
		t.Parallel()

		originalDelete := deleteAPIKeyFn
		t.Cleanup(func() { deleteAPIKeyFn = originalDelete })

		deleteCalled := false
		deleteAPIKeyFn = func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string) diag.Diagnostics {
			deleteCalled = true
			return nil
		}

		_, diags := closeAPIKey(context.Background(), &clients.ElasticsearchScopedClient{}, entitycore.CloseRequest[closeState]{
			State: closeState{KeyID: "key-id", InvalidateOnClose: false},
		})
		require.False(t, diags.HasError())
		require.False(t, deleteCalled)
	})

	t.Run("does not call delete when key id is empty", func(t *testing.T) {
		t.Parallel()

		originalDelete := deleteAPIKeyFn
		t.Cleanup(func() { deleteAPIKeyFn = originalDelete })

		deleteCalled := false
		deleteAPIKeyFn = func(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string) diag.Diagnostics {
			deleteCalled = true
			return nil
		}

		_, diags := closeAPIKey(context.Background(), &clients.ElasticsearchScopedClient{}, entitycore.CloseRequest[closeState]{
			State: closeState{KeyID: "", InvalidateOnClose: true},
		})
		require.False(t, diags.HasError())
		require.False(t, deleteCalled)
	})

	t.Run("calls delete when invalidate_on_close is true", func(t *testing.T) {
		t.Parallel()

		originalDelete := deleteAPIKeyFn
		t.Cleanup(func() { deleteAPIKeyFn = originalDelete })

		var (
			deleteCalled bool
			deleteKeyID  string
		)
		deleteAPIKeyFn = func(_ context.Context, _ *clients.ElasticsearchScopedClient, keyID string) diag.Diagnostics {
			deleteCalled = true
			deleteKeyID = keyID
			return nil
		}

		_, diags := closeAPIKey(context.Background(), &clients.ElasticsearchScopedClient{}, entitycore.CloseRequest[closeState]{
			State: closeState{KeyID: "key-to-delete", InvalidateOnClose: true},
		})
		require.False(t, diags.HasError())
		require.True(t, deleteCalled)
		require.Equal(t, "key-to-delete", deleteKeyID)
	})
}

func TestInvalidateOnCloseValue(t *testing.T) {
	t.Parallel()

	require.False(t, invalidateOnCloseValue(types.BoolNull()))
	require.False(t, invalidateOnCloseValue(types.BoolUnknown()))
	require.False(t, invalidateOnCloseValue(types.BoolValue(false)))
	require.True(t, invalidateOnCloseValue(types.BoolValue(true)))
}

func TestEffectiveAPIKeyType(t *testing.T) {
	t.Parallel()

	require.Equal(t, apikey.DefaultAPIKeyType, effectiveAPIKeyType(types.StringNull()).ValueString())
	require.Equal(t, apikey.DefaultAPIKeyType, effectiveAPIKeyType(types.StringValue("")).ValueString())
	require.Equal(t, apikey.CrossClusterAPIKeyType, effectiveAPIKeyType(types.StringValue(apikey.CrossClusterAPIKeyType)).ValueString())
}

func TestNewResourceImplementsInterfaces(t *testing.T) {
	t.Parallel()

	r := NewResource()
	require.Implements(t, (*fwephemeral.EphemeralResource)(nil), r)
	require.Implements(t, (*fwephemeral.EphemeralResourceWithConfigure)(nil), r)
	require.Implements(t, (*fwephemeral.EphemeralResourceWithClose)(nil), r)
}
