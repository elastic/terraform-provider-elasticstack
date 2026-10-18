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

package integrationpolicy

import (
	"context"
	"testing"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpgradeV2ToV3_DropsEnabled verifies that the V2→V3 upgrader drops the
// removed top-level `enabled` attribute and carries every other scalar field
// through unchanged.
func TestUpgradeV2ToV3_DropsEnabled(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		enabled bool
	}{
		{name: "enabled true", enabled: true},
		{name: "enabled false (the case that previously broke apply)", enabled: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priorSchema := getSchemaV2()
			liveSchema := getSchemaV3()

			prior := integrationPolicyModelV2{
				ID:                 types.StringValue("policy-id-1"),
				KibanaConnection:   providerschema.KibanaConnectionNullList(),
				PolicyID:           types.StringValue("policy-id-1"),
				Name:               types.StringValue("test-policy"),
				Namespace:          types.StringValue("default"),
				AgentPolicyID:      types.StringValue("agent-1"),
				AgentPolicyIDs:     types.ListNull(types.StringType),
				Description:        types.StringValue("a description"),
				Enabled:            types.BoolValue(tt.enabled),
				Force:              types.BoolValue(false),
				IntegrationName:    types.StringValue("tcp"),
				IntegrationVersion: types.StringValue("1.16.0"),
				OutputID:           types.StringNull(),
				// V2's frozen inputs element type predates the `condition`
				// attribute added to the live schema; use the frozen V2 type
				// here so this fixture matches what real V2 state looked
				// like on the wire.
				Inputs:   NewInputsNull(NewInputType(getInputsAttributeTypesV2())),
				VarsJSON: NewVarsJSONNull(),
				SpaceIDs: types.SetNull(types.StringType),
			}

			rawState := tfsdk.State{Schema: priorSchema}
			diags := rawState.Set(ctx, &prior)
			require.False(t, diags.HasError(), "set prior state: %v", diags)

			req := resource.UpgradeStateRequest{State: &rawState}
			resp := resource.UpgradeStateResponse{
				State: tfsdk.State{
					Schema: liveSchema,
					Raw:    tftypes.NewValue(liveSchema.Type().TerraformType(ctx), nil),
				},
			}

			upgradeV2ToV3(ctx, req, &resp)
			require.False(t, resp.Diagnostics.HasError(), "upgrade diagnostics: %v", resp.Diagnostics)

			var got integrationPolicyModel
			diags = resp.State.Get(ctx, &got)
			require.False(t, diags.HasError(), "decode upgraded state: %v", diags)

			assert.Equal(t, prior.ID, got.ID)
			assert.Equal(t, prior.PolicyID, got.PolicyID)
			assert.Equal(t, prior.Name, got.Name)
			assert.Equal(t, prior.Namespace, got.Namespace)
			assert.Equal(t, prior.AgentPolicyID, got.AgentPolicyID)
			assert.Equal(t, prior.AgentPolicyIDs, got.AgentPolicyIDs)
			assert.Equal(t, prior.Description, got.Description)
			assert.Equal(t, prior.Force, got.Force)
			assert.Equal(t, prior.IntegrationName, got.IntegrationName)
			assert.Equal(t, prior.IntegrationVersion, got.IntegrationVersion)
			assert.Equal(t, prior.OutputID, got.OutputID)
			assert.Equal(t, prior.SpaceIDs, got.SpaceIDs)

			// `enabled` is no longer part of the live schema; the only proof
			// that it was dropped is that the upgraded raw state has no such
			// attribute. The live schema's TerraformType must not contain it.
			liveType := liveSchema.Type().TerraformType(ctx)
			objType, ok := liveType.(tftypes.Object)
			require.True(t, ok, "live schema root must be an object type")
			_, hasEnabled := objType.AttributeTypes["enabled"]
			assert.False(t, hasEnabled, "live V3 schema must not expose top-level `enabled`")
		})
	}

	// Defensive: the prior schema MUST still expose `enabled`, otherwise the
	// upgrader has nothing to drop and this test isn't actually proving what
	// it claims.
	priorType := getSchemaV2().Type().TerraformType(ctx)
	priorObj, ok := priorType.(tftypes.Object)
	require.True(t, ok)
	_, hadEnabled := priorObj.AttributeTypes["enabled"]
	require.True(t, hadEnabled, "prior V2 schema is expected to retain `enabled` as the field being dropped")
}

// TestUpgradeV2ToV3_PreservesPopulatedInputs exercises the actual conversion
// logic in convertInputsV2ToV3/convertStreamsV2ToV3 against a populated
// `inputs` map (including a nested `streams` map and vars) decoded against
// the frozen V2 element type and rebuilt against the live V3 element type.
// TestUpgradeV2ToV3_DropsEnabled only ever passes NewInputsNull(...), so it
// never touches this path; a bug in rebuilding a populated map here would
// corrupt `terraform apply` for every pre-existing V2 resource with
// configured inputs.
func TestUpgradeV2ToV3_PreservesPopulatedInputs(t *testing.T) {
	ctx := context.Background()

	v2Types := getInputsAttributeTypesV2()
	streamMapType, ok := v2Types[attrStreams].(types.MapType)
	require.True(t, ok, "V2 inputs `streams` attribute must be a map type")
	streamObjType, ok := streamMapType.ElemType.(types.ObjectType)
	require.True(t, ok, "V2 `streams` element type must be an object type")
	defaultsObjType, ok := v2Types[attrDefaults].(types.ObjectType)
	require.True(t, ok, "V2 inputs `defaults` attribute must be an object type")

	streamsV2 := map[string]integrationPolicyInputStreamModelV2{
		"test.stream": {
			Enabled: types.BoolValue(true),
			Vars:    jsontypes.NewNormalizedValue(`{"stream_var":"value"}`),
		},
	}
	streams, diags := types.MapValueFrom(ctx, streamObjType, streamsV2)
	require.False(t, diags.HasError(), "build V2 streams map: %v", diags)

	inputsV2 := map[string]integrationPolicyInputsModelV2{
		"test-input": {
			Enabled:  types.BoolValue(true),
			Vars:     jsontypes.NewNormalizedValue(`{"input_var":"value"}`),
			Defaults: types.ObjectNull(defaultsObjType.AttrTypes),
			Streams:  streams,
		},
	}
	priorInputs, diags := NewInputsValueFrom(ctx, NewInputType(v2Types), inputsV2)
	require.False(t, diags.HasError(), "build V2 inputs map: %v", diags)

	priorSchema := getSchemaV2()
	liveSchema := getSchemaV3()

	prior := integrationPolicyModelV2{
		ID:                 types.StringValue("policy-id-1"),
		KibanaConnection:   providerschema.KibanaConnectionNullList(),
		PolicyID:           types.StringValue("policy-id-1"),
		Name:               types.StringValue("test-policy"),
		Namespace:          types.StringValue("default"),
		AgentPolicyID:      types.StringValue("agent-1"),
		AgentPolicyIDs:     types.ListNull(types.StringType),
		Description:        types.StringValue("a description"),
		Enabled:            types.BoolValue(true),
		Force:              types.BoolValue(false),
		IntegrationName:    types.StringValue("tcp"),
		IntegrationVersion: types.StringValue("1.16.0"),
		OutputID:           types.StringNull(),
		Inputs:             priorInputs,
		VarsJSON:           NewVarsJSONNull(),
		SpaceIDs:           types.SetNull(types.StringType),
	}

	rawState := tfsdk.State{Schema: priorSchema}
	diags = rawState.Set(ctx, &prior)
	require.False(t, diags.HasError(), "set prior state: %v", diags)

	req := resource.UpgradeStateRequest{State: &rawState}
	resp := resource.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: liveSchema,
			Raw:    tftypes.NewValue(liveSchema.Type().TerraformType(ctx), nil),
		},
	}

	upgradeV2ToV3(ctx, req, &resp)
	require.False(t, resp.Diagnostics.HasError(), "upgrade diagnostics: %v", resp.Diagnostics)

	var got integrationPolicyModel
	diags = resp.State.Get(ctx, &got)
	require.False(t, diags.HasError(), "decode upgraded state: %v", diags)

	require.False(t, got.Inputs.IsNull(), "upgraded inputs must not be null")

	var gotInputs map[string]integrationPolicyInputsModel
	diags = got.Inputs.ElementsAs(ctx, &gotInputs, false)
	require.False(t, diags.HasError(), "decode upgraded inputs: %v", diags)
	require.Len(t, gotInputs, 1)
	require.Contains(t, gotInputs, "test-input")

	input := gotInputs["test-input"]
	assert.Equal(t, types.BoolValue(true), input.Enabled)
	assert.True(t, input.Condition.IsNull(), "condition must be present-and-null on an input upgraded from V2")
	assert.JSONEq(t, `{"input_var":"value"}`, input.Vars.ValueString())
	require.False(t, input.Streams.IsNull(), "streams must be preserved through the upgrade")

	var gotStreams map[string]integrationPolicyInputStreamModel
	diags = input.Streams.ElementsAs(ctx, &gotStreams, false)
	require.False(t, diags.HasError(), "decode upgraded streams: %v", diags)
	require.Len(t, gotStreams, 1)
	require.Contains(t, gotStreams, "test.stream")

	stream := gotStreams["test.stream"]
	assert.Equal(t, types.BoolValue(true), stream.Enabled)
	assert.True(t, stream.Condition.IsNull(), "condition must be present-and-null on a stream upgraded from V2")
	assert.JSONEq(t, `{"stream_var":"value"}`, stream.Vars.ValueString())
}
