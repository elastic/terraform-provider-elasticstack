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
	"context"
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

// testKibanaConnectionAttrTypes mirrors config.KibanaConnection's tfsdk tags so
// tests can build a types.List matching what GetKibanaClient expects, without
// depending on the clients package's unexported test helpers.
func testKibanaConnectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"username":     types.StringType,
		"password":     types.StringType,
		"api_key":      types.StringType,
		"bearer_token": types.StringType,
		"endpoints":    types.ListType{ElemType: types.StringType},
		"insecure":     types.BoolType,
		"ca_certs":     types.ListType{ElemType: types.StringType},
	}
}

func newTestProviderClientFactory(t *testing.T) *clients.ProviderClientFactory {
	t.Helper()

	factory, diags := clients.NewProviderClientFactoryFromFramework(context.Background(), config.ProviderConfiguration{
		Kibana: []config.KibanaConnection{
			{
				Username: types.StringValue("elastic"),
				Password: types.StringValue("changeme"),
				Endpoints: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("http://localhost:5601"),
				}),
				CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
				Insecure: types.BoolValue(true),
			},
		},
	}, "test")
	require.False(t, diags.HasError(), "failed to create test factory: %v", diags)
	require.NotNil(t, factory)
	return factory
}

func emptyKibanaConnectionList(t *testing.T) types.List {
	t.Helper()

	list, diags := types.ListValueFrom(context.Background(),
		types.ObjectType{AttrTypes: testKibanaConnectionAttrTypes()},
		[]config.KibanaConnection{},
	)
	require.False(t, diags.HasError())
	return list
}

// stateSchemaWithSpaceIDs returns a minimal resource schema exposing only the
// space_ids attribute that GetOperationalSpaceFromState reads from state.
func stateSchemaWithSpaceIDs() rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"space_ids": rschema.SetAttribute{ElementType: types.StringType, Optional: true},
		},
	}
}

func stateWithSpaceIDs(t *testing.T, spaceIDs []string) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	s := stateSchemaWithSpaceIDs()

	var spaceIDsValue tftypes.Value
	if spaceIDs == nil {
		spaceIDsValue = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil)
	} else {
		elems := make([]tftypes.Value, 0, len(spaceIDs))
		for _, id := range spaceIDs {
			elems = append(elems, tftypes.NewValue(tftypes.String, id))
		}
		spaceIDsValue = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, elems)
	}

	raw := tftypes.NewValue(s.Type().TerraformType(ctx), map[string]tftypes.Value{
		"space_ids": spaceIDsValue,
	})

	return tfsdk.State{Raw: raw, Schema: s}
}

// stateMissingSpaceIDs returns a state whose schema does not declare space_ids,
// used to exercise the error path when GetOperationalSpaceFromState cannot find
// the attribute.
func stateMissingSpaceIDs() tfsdk.State {
	ctx := context.Background()
	s := rschema.Schema{Attributes: map[string]rschema.Attribute{}}
	raw := tftypes.NewValue(s.Type().TerraformType(ctx), map[string]tftypes.Value{})
	return tfsdk.State{Raw: raw, Schema: s}
}

func TestResolveFleetClientAndSpace_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestProviderClientFactory(t)
	kibanaConnection := emptyKibanaConnectionList(t)
	state := stateWithSpaceIDs(t, []string{"custom-space"})

	fleetClient, spaceID, diags := ResolveFleetClientAndSpace(ctx, state, factory, kibanaConnection)

	require.False(t, diags.HasError(), "unexpected error diagnostics: %v", diags)
	require.NotNil(t, fleetClient)
	require.Equal(t, "custom-space", spaceID)
}

func TestResolveFleetClientAndSpace_DefaultSpaceWhenSpaceIDsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestProviderClientFactory(t)
	kibanaConnection := emptyKibanaConnectionList(t)
	state := stateWithSpaceIDs(t, []string{})

	fleetClient, spaceID, diags := ResolveFleetClientAndSpace(ctx, state, factory, kibanaConnection)

	require.False(t, diags.HasError(), "unexpected error diagnostics: %v", diags)
	require.NotNil(t, fleetClient)
	require.Empty(t, spaceID)
}

func TestResolveFleetClientAndSpace_ClientResolutionErrorSkipsSpaceLookup(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	kibanaConnection := emptyKibanaConnectionList(t)
	// A state missing space_ids would fail GetOperationalSpaceFromState if it
	// were reached; a nil factory must short-circuit before that happens.
	state := stateMissingSpaceIDs()

	fleetClient, spaceID, diags := ResolveFleetClientAndSpace(ctx, state, nil, kibanaConnection)

	require.True(t, diags.HasError())
	require.Nil(t, fleetClient)
	require.Empty(t, spaceID)
}

func TestResolveFleetClientAndSpace_SpaceResolutionErrorPropagates(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	factory := newTestProviderClientFactory(t)
	kibanaConnection := emptyKibanaConnectionList(t)
	state := stateMissingSpaceIDs()

	fleetClient, spaceID, diags := ResolveFleetClientAndSpace(ctx, state, factory, kibanaConnection)

	require.True(t, diags.HasError())
	require.Nil(t, fleetClient)
	require.Empty(t, spaceID)
}
