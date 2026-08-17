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

package datastreamlifecycle

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDPlanModifiers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	idAttr := getSchemaFactory(ctx).Attributes["id"].(schema.StringAttribute)
	priorID := types.StringValue("stale-uuid/old-name")

	tests := []struct {
		name        string
		configName  tftypes.Value
		wantUnknown bool
	}{
		{
			name:        "name change sets planned id unknown",
			configName:  tftypes.NewValue(tftypes.String, "new-name"),
			wantUnknown: true,
		},
		{
			name:        "name unchanged keeps prior id after UseStateForUnknown",
			configName:  tftypes.NewValue(tftypes.String, "old-name"),
			wantUnknown: false,
		},
		{
			name:        "unknown config name sets planned id unknown",
			configName:  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			wantUnknown: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			state, config := testNameStateConfig(t, "old-name", tt.configName)
			req := planmodifier.StringRequest{
				Path:        path.Root("id"),
				State:       state,
				Config:      config,
				StateValue:  priorID,
				PlanValue:   types.StringUnknown(),
				ConfigValue: types.StringNull(),
			}
			resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
			for _, m := range idAttr.PlanModifiers {
				m.PlanModifyString(ctx, req, resp)
			}
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
			if tt.wantUnknown {
				assert.True(t, resp.PlanValue.IsUnknown())
				return
			}
			assert.Equal(t, priorID, resp.PlanValue)
		})
	}
}

func testNameStateConfig(t *testing.T, stateName string, configName tftypes.Value) (tfsdk.State, tfsdk.Config) {
	t.Helper()
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true},
			attrName: schema.StringAttribute{Required: true},
		},
	}
	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":     tftypes.String,
			attrName: tftypes.String,
		},
	}
	stateRaw := tftypes.NewValue(objectType, map[string]tftypes.Value{
		"id":     tftypes.NewValue(tftypes.String, "stale-uuid/old-name"),
		attrName: tftypes.NewValue(tftypes.String, stateName),
	})
	configRaw := tftypes.NewValue(objectType, map[string]tftypes.Value{
		"id":     tftypes.NewValue(tftypes.String, nil),
		attrName: configName,
	})
	return tfsdk.State{Raw: stateRaw, Schema: testSchema}, tfsdk.Config{Raw: configRaw, Schema: testSchema}
}
