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
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// testEnvelopeModel is a minimal state model exercising the fields
// ReadPolicyEnvelope depends on: a policy_id attribute, a space_ids
// attribute (consumed via GetOperationalSpaceFromState), and a
// kibana_connection block (consumed via the provider client factory).
type testEnvelopeModel struct {
	PolicyID         types.String `tfsdk:"policy_id"`
	SpaceIDs         types.Set    `tfsdk:"space_ids"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`
}

// testEnvelopePolicy stands in for a package-specific Fleet policy payload
// (e.g. kbapi.PackagePolicy, kbapi.KibanaHTTPAPIsAgentPolicyResponse).
type testEnvelopePolicy struct {
	Name string
}

func testEnvelopeSchema() rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"policy_id": rschema.StringAttribute{Required: true},
			"space_ids": rschema.SetAttribute{Optional: true, ElementType: types.StringType},
		},
		Blocks: map[string]rschema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}
}

func testEnvelopeObjectType(ctx context.Context) tftypes.Type {
	kibanaConnType := types.ListType{ElemType: providerschema.KibanaConnectionObjectType()}.TerraformType(ctx)
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"policy_id":         tftypes.String,
			"space_ids":         tftypes.Set{ElementType: tftypes.String},
			"kibana_connection": kibanaConnType,
		},
	}
}

// makeTestEnvelopeState builds a tfsdk.State with the given policy ID and
// space IDs, and a null kibana_connection block (so the provider client
// factory falls back to its configured defaults).
func makeTestEnvelopeState(ctx context.Context, t *testing.T, policyID string, spaceIDs []string) tfsdk.State {
	t.Helper()

	objType := testEnvelopeObjectType(ctx)
	kibanaConnType := objType.(tftypes.Object).AttributeTypes["kibana_connection"]

	spaceIDValues := make([]tftypes.Value, 0, len(spaceIDs))
	for _, id := range spaceIDs {
		spaceIDValues = append(spaceIDValues, tftypes.NewValue(tftypes.String, id))
	}
	var spaceIDsValue tftypes.Value
	if spaceIDs == nil {
		spaceIDsValue = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil)
	} else {
		spaceIDsValue = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, spaceIDValues)
	}

	objValue := tftypes.NewValue(objType, map[string]tftypes.Value{
		"policy_id":         tftypes.NewValue(tftypes.String, policyID),
		"space_ids":         spaceIDsValue,
		"kibana_connection": tftypes.NewValue(kibanaConnType, nil),
	})

	return tfsdk.State{
		Raw:    objValue,
		Schema: testEnvelopeSchema(),
	}
}

func newTestEnvelopeFactory(ctx context.Context, t *testing.T) *clients.ProviderClientFactory {
	t.Helper()

	cfg := config.ProviderConfiguration{
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
	}
	factory, diags := clients.NewProviderClientFactoryFromFramework(ctx, cfg, "test")
	require.False(t, diags.HasError(), "failed to create test factory: %v", diags)
	require.NotNil(t, factory)
	return factory
}

func kibanaConnectionOf(m *testEnvelopeModel) types.List { return m.KibanaConnection }
func policyIDOf(m *testEnvelopeModel) string             { return m.PolicyID.ValueString() }

func TestReadPolicyEnvelope_Found(t *testing.T) {
	ctx := t.Context()
	factory := newTestEnvelopeFactory(ctx, t)
	state := makeTestEnvelopeState(ctx, t, "policy-1", []string{"custom-space"})

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}

	var fetchCalledWith struct {
		policyID string
		spaceID  string
	}
	fetch := func(_ context.Context, client *fleetclient.Client, policyID string, spaceID string) (*testEnvelopePolicy, diag.Diagnostics) {
		require.NotNil(t, client)
		fetchCalledWith.policyID = policyID
		fetchCalledWith.spaceID = spaceID
		return &testEnvelopePolicy{Name: "found"}, nil
	}

	var stateModel testEnvelopeModel
	fleetClient, resolvedPolicyID, spaceID, policy, ok := ReadPolicyEnvelope(
		ctx, req, resp, factory, &stateModel, kibanaConnectionOf, policyIDOf, fetch,
	)

	require.True(t, ok)
	require.False(t, resp.Diagnostics.HasError())
	require.NotNil(t, fleetClient)
	require.Equal(t, "policy-1", resolvedPolicyID)
	require.Equal(t, "custom-space", spaceID)
	require.NotNil(t, policy)
	require.Equal(t, "found", policy.Name)
	require.Equal(t, "policy-1", fetchCalledWith.policyID)
	require.Equal(t, "custom-space", fetchCalledWith.spaceID)
}

func TestReadPolicyEnvelope_NotFoundRemovesResource(t *testing.T) {
	ctx := t.Context()
	factory := newTestEnvelopeFactory(ctx, t)
	state := makeTestEnvelopeState(ctx, t, "policy-missing", nil)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}

	fetch := func(_ context.Context, _ *fleetclient.Client, _ string, _ string) (*testEnvelopePolicy, diag.Diagnostics) {
		return nil, nil
	}

	var stateModel testEnvelopeModel
	_, _, _, policy, ok := ReadPolicyEnvelope(
		ctx, req, resp, factory, &stateModel, kibanaConnectionOf, policyIDOf, fetch,
	)

	require.False(t, ok)
	require.Nil(t, policy)
	require.False(t, resp.Diagnostics.HasError())
	require.True(t, resp.State.Raw.IsNull(), "expected RemoveResource to null out state")
}

func TestReadPolicyEnvelope_FetchErrorStopsShort(t *testing.T) {
	ctx := t.Context()
	factory := newTestEnvelopeFactory(ctx, t)
	state := makeTestEnvelopeState(ctx, t, "policy-1", []string{"default"})

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}

	fetch := func(_ context.Context, _ *fleetclient.Client, _ string, _ string) (*testEnvelopePolicy, diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError("fetch failed", "simulated API error")
		return nil, diags
	}

	var stateModel testEnvelopeModel
	_, _, _, policy, ok := ReadPolicyEnvelope(
		ctx, req, resp, factory, &stateModel, kibanaConnectionOf, policyIDOf, fetch,
	)

	require.False(t, ok)
	require.Nil(t, policy)
	require.True(t, resp.Diagnostics.HasError())
	require.False(t, resp.State.Raw.IsNull(), "state should not be removed on a fetch error")
}
