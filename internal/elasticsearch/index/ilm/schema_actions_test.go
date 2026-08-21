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

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
		planIndex    *types.Bool
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
			name:         "unset with config index null uses plan force_merge_index false",
			configClone:  types.BoolNull(),
			planClone:    types.BoolValue(true),
			configIndex:  types.BoolNull(),
			planIndex:    boolPtr(types.BoolValue(false)),
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
			if tt.planIndex != nil {
				req.Plan = tfsdk.Plan{
					Schema: testSchema,
					Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
						attrForceMergeIndex:   boolTF(t, *tt.planIndex),
						attrForceMergeOnClone: boolTF(t, tt.planClone),
					}),
				}
			}
			resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}
			nullForceMergeOnCloneWhenIndexDisabled{}.PlanModifyBool(ctx, req, resp)
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
			assert.Equal(t, tt.expectedPlan, resp.PlanValue)
		})
	}
}

// Production attributes live under hot|cold|frozen.searchable_snapshot blocks.
// A flat schema can hide GetAttribute path bugs that only show up with nested blocks.
func TestNullForceMergeOnCloneWhenIndexDisabled_nestedBlock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testSchema := schema.Schema{
		Blocks: map[string]schema.Block{
			ilmPhaseFrozen: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					attrMinAge: schema.StringAttribute{Optional: true},
				},
				Blocks: map[string]schema.Block{
					ilmActionSearchableSnapshot: schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							attrSnapshotRepository: schema.StringAttribute{Optional: true},
							attrForceMergeIndex:    schema.BoolAttribute{Optional: true, Computed: true},
							attrForceMergeOnClone:  schema.BoolAttribute{Optional: true, Computed: true},
						},
					},
				},
			},
		},
	}

	tfType := testSchema.Type().TerraformType(ctx)
	boolTF := func(t *testing.T, v types.Bool) tftypes.Value {
		t.Helper()
		tv, err := v.ToTerraformValue(ctx)
		require.NoError(t, err)
		return tv
	}
	strTF := func(t *testing.T, v types.String) tftypes.Value {
		t.Helper()
		tv, err := v.ToTerraformValue(ctx)
		require.NoError(t, err)
		return tv
	}

	ssObj := tftypes.NewValue(tfType.(tftypes.Object).AttributeTypes[ilmPhaseFrozen].(tftypes.Object).AttributeTypes[ilmActionSearchableSnapshot], map[string]tftypes.Value{
		attrSnapshotRepository: strTF(t, types.StringValue("repo-a")),
		attrForceMergeIndex:    boolTF(t, types.BoolValue(false)),
		attrForceMergeOnClone:  boolTF(t, types.BoolNull()),
	})
	frozenObj := tftypes.NewValue(tfType.(tftypes.Object).AttributeTypes[ilmPhaseFrozen], map[string]tftypes.Value{
		attrMinAge:                  strTF(t, types.StringValue("30d")),
		ilmActionSearchableSnapshot: ssObj,
	})
	raw := tftypes.NewValue(tfType, map[string]tftypes.Value{
		ilmPhaseFrozen: frozenObj,
	})

	config := tfsdk.Config{Schema: testSchema, Raw: raw}
	plan := tfsdk.Plan{Schema: testSchema, Raw: raw}
	attrPath := path.Root(ilmPhaseFrozen).AtName(ilmActionSearchableSnapshot).AtName(attrForceMergeOnClone)

	req := planmodifier.BoolRequest{
		Path:        attrPath,
		Config:      config,
		Plan:        plan,
		ConfigValue: types.BoolNull(),
		PlanValue:   types.BoolValue(true),
	}
	resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}
	nullForceMergeOnCloneWhenIndexDisabled{}.PlanModifyBool(ctx, req, resp)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	assert.Equal(t, types.BoolNull(), resp.PlanValue)
}

func TestNullForceMergeOnCloneInSearchableSnapshot(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	attrTypes := searchableSnapshotObjectType().AttrTypes
	ssObj := func(index, clone types.Bool) types.Object {
		return types.ObjectValueMust(attrTypes, map[string]attr.Value{
			attrSnapshotRepository: types.StringValue("repo-a"),
			attrForceMergeIndex:    index,
			attrForceMergeOnClone:  clone,
		})
	}

	tests := []struct {
		name          string
		configIndex   types.Bool
		configClone   types.Bool
		planIndex     types.Bool
		planClone     types.Bool
		expectedClone types.Bool
	}{
		{
			name:          "unset with force_merge_index false nulls the computed default",
			configIndex:   types.BoolValue(false),
			configClone:   types.BoolNull(),
			planIndex:     types.BoolValue(false),
			planClone:     types.BoolValue(true),
			expectedClone: types.BoolNull(),
		},
		{
			name:          "unset with plan force_merge_index false when config index is null",
			configIndex:   types.BoolNull(),
			configClone:   types.BoolNull(),
			planIndex:     types.BoolValue(false),
			planClone:     types.BoolValue(true),
			expectedClone: types.BoolNull(),
		},
		{
			name:          "unset with force_merge_index true leaves the default",
			configIndex:   types.BoolValue(true),
			configClone:   types.BoolNull(),
			planIndex:     types.BoolValue(true),
			planClone:     types.BoolValue(true),
			expectedClone: types.BoolValue(true),
		},
		{
			name:          "explicit false is unchanged",
			configIndex:   types.BoolValue(true),
			configClone:   types.BoolValue(false),
			planIndex:     types.BoolValue(true),
			planClone:     types.BoolValue(false),
			expectedClone: types.BoolValue(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := planmodifier.ObjectRequest{
				ConfigValue: ssObj(tt.configIndex, tt.configClone),
				PlanValue:   ssObj(tt.planIndex, tt.planClone),
			}
			resp := &planmodifier.ObjectResponse{PlanValue: req.PlanValue}
			nullForceMergeOnCloneInSearchableSnapshot{}.PlanModifyObject(ctx, req, resp)
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
			got, ok := resp.PlanValue.Attributes()[attrForceMergeOnClone].(types.Bool)
			require.True(t, ok)
			assert.Equal(t, tt.expectedClone, got)
		})
	}
}

func boolPtr(v types.Bool) *types.Bool {
	return &v
}
