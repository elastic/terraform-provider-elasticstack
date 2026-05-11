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
				Inputs:             NewInputsNull(getInputsElementType()),
				VarsJSON:           NewVarsJSONNull(),
				SpaceIDs:           types.SetNull(types.StringType),
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
