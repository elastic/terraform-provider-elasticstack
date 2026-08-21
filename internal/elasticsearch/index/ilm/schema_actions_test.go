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

package ilm

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

func TestNullForceMergeOnCloneWhenIndexDisabled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			attrForceMergeIndex:   schema.BoolAttribute{Optional: true},
			attrForceMergeOnClone: schema.BoolAttribute{Optional: true, Computed: true},
		},
	}
	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			attrForceMergeIndex:   tftypes.Bool,
			attrForceMergeOnClone: tftypes.Bool,
		},
	}

	boolTF := func(t *testing.T, v types.Bool) tftypes.Value {
		t.Helper()
		tv, err := v.ToTerraformValue(ctx)
		require.NoError(t, err)
		return tv
	}

	tests := []struct {
		name         string
		configClone  types.Bool
		planClone    types.Bool
		configIndex  types.Bool
		expectedPlan types.Bool
	}{
		{
			name:         "unset with force_merge_index false nulls the computed default",
			configClone:  types.BoolNull(),
			planClone:    types.BoolValue(true),
			configIndex:  types.BoolValue(false),
			expectedPlan: types.BoolNull(),
		},
		{
			name:         "unknown config with force_merge_index false is left unknown",
			configClone:  types.BoolUnknown(),
			planClone:    types.BoolUnknown(),
			configIndex:  types.BoolValue(false),
			expectedPlan: types.BoolUnknown(),
		},
		{
			name:         "explicit true is unchanged",
			configClone:  types.BoolValue(true),
			planClone:    types.BoolValue(true),
			configIndex:  types.BoolValue(false),
			expectedPlan: types.BoolValue(true),
		},
		{
			name:         "explicit false is unchanged",
			configClone:  types.BoolValue(false),
			planClone:    types.BoolValue(false),
			configIndex:  types.BoolValue(false),
			expectedPlan: types.BoolValue(false),
		},
		{
			name:         "unset with force_merge_index true leaves the default",
			configClone:  types.BoolNull(),
			planClone:    types.BoolValue(true),
			configIndex:  types.BoolValue(true),
			expectedPlan: types.BoolValue(true),
		},
		{
			name:         "unset with force_merge_index unset leaves the default",
			configClone:  types.BoolNull(),
			planClone:    types.BoolValue(true),
			configIndex:  types.BoolNull(),
			expectedPlan: types.BoolValue(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := tfsdk.Config{
				Schema: testSchema,
				Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
					attrForceMergeIndex:   boolTF(t, tt.configIndex),
					attrForceMergeOnClone: boolTF(t, tt.configClone),
				}),
			}
			req := planmodifier.BoolRequest{
				Path:        path.Root(attrForceMergeOnClone),
				Config:      config,
				ConfigValue: tt.configClone,
				PlanValue:   tt.planClone,
			}
			resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}
			nullForceMergeOnCloneWhenIndexDisabled{}.PlanModifyBool(ctx, req, resp)
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
			assert.Equal(t, tt.expectedPlan, resp.PlanValue)
		})
	}
}
