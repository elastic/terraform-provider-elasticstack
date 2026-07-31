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

package fleet

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// kibanaConnectionAttrTypes mirrors the object shape of config.KibanaConnection, used to
// build a types.List of kibana_connection blocks for these tests without depending on the
// unexported helper of the same name in package clients.
func kibanaConnectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"username":     types.StringType,
		"password":     types.StringType,
		"api_key":      types.StringType,
		"bearer_token": types.StringType,
		"endpoints":    types.ListType{ElemType: types.StringType},
		"ca_certs":     types.ListType{ElemType: types.StringType},
		"insecure":     types.BoolType,
	}
}

func newTestKibanaConnectionList(t *testing.T, endpoint string) types.List {
	t.Helper()
	ctx := t.Context()

	conn := config.KibanaConnection{
		Username:    types.StringValue("elastic"),
		Password:    types.StringValue("changeme"),
		APIKey:      types.StringValue(""),
		BearerToken: types.StringValue(""),
		Endpoints: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue(endpoint),
		}),
		CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
		Insecure: types.BoolValue(false),
	}

	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()}, []config.KibanaConnection{conn})
	require.False(t, diags.HasError())
	return list
}

// stateWithSpaceIDs builds a minimal tfsdk.State containing only a space_ids attribute, for
// exercising the space-resolution half of ResolveReadDeleteContext.
func stateWithSpaceIDs(t *testing.T, spaceIDs []string) tfsdk.State {
	t.Helper()

	schema := rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"space_ids": rschema.SetAttribute{ElementType: types.StringType, Optional: true},
		},
	}

	setType := tftypes.Set{ElementType: tftypes.String}
	values := make([]tftypes.Value, 0, len(spaceIDs))
	for _, id := range spaceIDs {
		values = append(values, tftypes.NewValue(tftypes.String, id))
	}

	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"space_ids": setType}}
	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"space_ids": tftypes.NewValue(setType, values),
	})

	return tfsdk.State{Raw: objValue, Schema: schema}
}

// TestResolveReadDeleteContext_NilFactory verifies that a nil client factory (an unconfigured
// resource) surfaces a diagnostic instead of panicking, matching the error behavior of
// ProviderClientFactory.GetKibanaClient.
func TestResolveReadDeleteContext_NilFactory(t *testing.T) {
	ctx := t.Context()

	fleetClient, spaceID, diags := ResolveReadDeleteContext(ctx, nil, types.ListNull(types.ObjectType{AttrTypes: kibanaConnectionAttrTypes()}), tfsdk.State{})

	require.True(t, diags.HasError())
	require.Nil(t, fleetClient)
	require.Empty(t, spaceID)
}

// TestResolveReadDeleteContext_ResolvesClientAndSpace verifies that a configured factory and a
// state with space_ids set produces a non-nil Fleet client and the expected operational space,
// mirroring the preamble previously duplicated across the policy resources' Read/Delete methods.
func TestResolveReadDeleteContext_ResolvesClientAndSpace(t *testing.T) {
	ctx := t.Context()

	factory, diags := clients.NewProviderClientFactoryFromFramework(ctx, config.ProviderConfiguration{}, "test")
	require.False(t, diags.HasError())

	kibanaConnection := newTestKibanaConnectionList(t, "http://127.0.0.1:0")
	state := stateWithSpaceIDs(t, []string{"custom-space"})

	fleetClient, spaceID, diags := ResolveReadDeleteContext(ctx, factory, kibanaConnection, state)

	require.False(t, diags.HasError())
	require.NotNil(t, fleetClient)
	require.Equal(t, "custom-space", spaceID)
}

// TestResolveReadDeleteContext_DefaultSpace verifies that an empty space_ids set resolves to
// the default space (empty string), matching GetOperationalSpaceFromState's contract.
func TestResolveReadDeleteContext_DefaultSpace(t *testing.T) {
	ctx := t.Context()

	factory, diags := clients.NewProviderClientFactoryFromFramework(ctx, config.ProviderConfiguration{}, "test")
	require.False(t, diags.HasError())

	kibanaConnection := newTestKibanaConnectionList(t, "http://127.0.0.1:0")
	state := stateWithSpaceIDs(t, nil)

	fleetClient, spaceID, diags := ResolveReadDeleteContext(ctx, factory, kibanaConnection, state)

	require.False(t, diags.HasError())
	require.NotNil(t, fleetClient)
	require.Empty(t, spaceID)
}
