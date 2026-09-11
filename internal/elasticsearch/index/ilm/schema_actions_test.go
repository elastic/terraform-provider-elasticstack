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

func TestDefaultForceMergeOnClone_bool(t *testing.T) {
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

	tests := []struct {
		name         string
		configClone  types.Bool
		planClone    types.Bool
		configIndex  types.Bool
		planIndex    types.Bool
		expectedPlan types.Bool
	}{
		{
			name:         "unset with force_merge_index false stays null",
			configClone:  types.BoolNull(),
			planClone:    types.BoolNull(),
			configIndex:  types.BoolValue(false),
			planIndex:    types.BoolValue(false),
			expectedPlan: types.BoolNull(),
		},
		{
			name:         "unset with config index null uses plan force_merge_index false",
			configClone:  types.BoolNull(),
			planClone:    types.BoolNull(),
			configIndex:  types.BoolNull(),
			planIndex:    types.BoolValue(false),
			expectedPlan: types.BoolNull(),
		},
		{
			name:         "unknown config with force_merge_index false is left unknown",
			configClone:  types.BoolUnknown(),
			planClone:    types.BoolUnknown(),
			configIndex:  types.BoolValue(false),
			planIndex:    types.BoolValue(false),
			expectedPlan: types.BoolUnknown(),
		},
		{
			name:         "explicit true is unchanged",
			configClone:  types.BoolValue(true),
			planClone:    types.BoolValue(true),
			configIndex:  types.BoolValue(true),
			planIndex:    types.BoolValue(true),
			expectedPlan: types.BoolValue(true),
		},
		{
			name:         "explicit false is unchanged",
			configClone:  types.BoolValue(false),
			planClone:    types.BoolValue(false),
			configIndex:  types.BoolValue(true),
			planIndex:    types.BoolValue(true),
			expectedPlan: types.BoolValue(false),
		},
		{
			name:         "unset with force_merge_index true defaults to true",
			configClone:  types.BoolNull(),
			planClone:    types.BoolNull(),
			configIndex:  types.BoolValue(true),
			planIndex:    types.BoolValue(true),
			expectedPlan: types.BoolValue(true),
		},
		{
			name:         "unset with force_merge_index unset leaves plan when sibling is unknown",
			configClone:  types.BoolNull(),
			planClone:    types.BoolNull(),
			configIndex:  types.BoolNull(),
			planIndex:    types.BoolNull(),
			expectedPlan: types.BoolNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := tfsdk.Config{
				Schema: testSchema,
				Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
					attrForceMergeIndex:   tfVal(ctx, t, tt.configIndex),
					attrForceMergeOnClone: tfVal(ctx, t, tt.configClone),
				}),
			}
			plan := tfsdk.Plan{
				Schema: testSchema,
				Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
					attrForceMergeIndex:   tfVal(ctx, t, tt.planIndex),
					attrForceMergeOnClone: tfVal(ctx, t, tt.planClone),
				}),
			}
			req := planmodifier.BoolRequest{
				Path:        path.Root(attrForceMergeOnClone),
				Config:      config,
				Plan:        plan,
				ConfigValue: tt.configClone,
				PlanValue:   tt.planClone,
			}
			resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}
			defaultForceMergeOnClone{}.PlanModifyBool(ctx, req, resp)
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
			assert.Equal(t, tt.expectedPlan, resp.PlanValue)
		})
	}
}

func TestDefaultForceMergeOnClone_nestedProductionBlocks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name         string
		phase        string
		blockSchema  schema.Block
		index        types.Bool
		configClone  types.Bool
		planClone    types.Bool
		expectedPlan types.Bool
	}{
		{
			name:         "frozen searchable_snapshot nulls clone when force_merge_index is false",
			phase:        ilmPhaseFrozen,
			blockSchema:  phaseFrozenBlock(),
			index:        types.BoolValue(false),
			configClone:  types.BoolNull(),
			planClone:    types.BoolUnknown(),
			expectedPlan: types.BoolNull(),
		},
		{
			name:         "frozen searchable_snapshot defaults clone to true when force_merge_index is true",
			phase:        ilmPhaseFrozen,
			blockSchema:  phaseFrozenBlock(),
			index:        types.BoolValue(true),
			configClone:  types.BoolNull(),
			planClone:    types.BoolNull(),
			expectedPlan: types.BoolValue(true),
		},
		{
			name:         "hot searchable_snapshot nulls clone when force_merge_index is false",
			phase:        ilmPhaseHot,
			blockSchema:  phaseHotBlock(),
			index:        types.BoolValue(false),
			configClone:  types.BoolNull(),
			planClone:    types.BoolUnknown(),
			expectedPlan: types.BoolNull(),
		},
		{
			name:         "hot searchable_snapshot defaults clone to true when force_merge_index is true",
			phase:        ilmPhaseHot,
			blockSchema:  phaseHotBlock(),
			index:        types.BoolValue(true),
			configClone:  types.BoolNull(),
			planClone:    types.BoolNull(),
			expectedPlan: types.BoolValue(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testSchema := schema.Schema{
				Blocks: map[string]schema.Block{
					tt.phase: tt.blockSchema,
				},
			}
			configRaw := nestedSearchableSnapshotRaw(ctx, t, testSchema, tt.phase, tt.index, tt.configClone)
			planRaw := nestedSearchableSnapshotRaw(ctx, t, testSchema, tt.phase, tt.index, tt.planClone)
			attrPath := path.Root(tt.phase).AtName(ilmActionSearchableSnapshot).AtName(attrForceMergeOnClone)

			req := planmodifier.BoolRequest{
				Path:        attrPath,
				Config:      tfsdk.Config{Schema: testSchema, Raw: configRaw},
				Plan:        tfsdk.Plan{Schema: testSchema, Raw: planRaw},
				ConfigValue: tt.configClone,
				PlanValue:   tt.planClone,
			}
			resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}
			defaultForceMergeOnClone{}.PlanModifyBool(ctx, req, resp)
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
			assert.Equal(t, tt.expectedPlan, resp.PlanValue)
		})
	}
}

func TestDefaultForceMergeOnClone_object(t *testing.T) {
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
			name:          "unset with force_merge_index false stays null",
			configIndex:   types.BoolValue(false),
			configClone:   types.BoolNull(),
			planIndex:     types.BoolValue(false),
			planClone:     types.BoolNull(),
			expectedClone: types.BoolNull(),
		},
		{
			name:          "unset with plan force_merge_index false when config index is null",
			configIndex:   types.BoolNull(),
			configClone:   types.BoolNull(),
			planIndex:     types.BoolValue(false),
			planClone:     types.BoolNull(),
			expectedClone: types.BoolNull(),
		},
		{
			name:          "unset with force_merge_index true defaults to true",
			configIndex:   types.BoolValue(true),
			configClone:   types.BoolNull(),
			planIndex:     types.BoolValue(true),
			planClone:     types.BoolNull(),
			expectedClone: types.BoolValue(true),
		},
		{
			name:          "unset with force_merge_index unset defaults to true",
			configIndex:   types.BoolNull(),
			configClone:   types.BoolNull(),
			planIndex:     types.BoolNull(),
			planClone:     types.BoolNull(),
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
		{
			name:          "unknown config clone is left unknown",
			configIndex:   types.BoolValue(false),
			configClone:   types.BoolUnknown(),
			planIndex:     types.BoolValue(false),
			planClone:     types.BoolUnknown(),
			expectedClone: types.BoolUnknown(),
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
			defaultForceMergeOnClone{}.PlanModifyObject(ctx, req, resp)
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
			got, ok := resp.PlanValue.Attributes()[attrForceMergeOnClone].(types.Bool)
			require.True(t, ok)
			assert.Equal(t, tt.expectedClone, got)
		})
	}
}

func nestedSearchableSnapshotRaw(ctx context.Context, t *testing.T, s schema.Schema, phaseName string, index, clone types.Bool) tftypes.Value {
	t.Helper()

	rootType, ok := s.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)
	phaseType, ok := rootType.AttributeTypes[phaseName].(tftypes.Object)
	require.True(t, ok)
	ssType, ok := phaseType.AttributeTypes[ilmActionSearchableSnapshot].(tftypes.Object)
	require.True(t, ok)

	ssVal := tfObjectFill(ssType, map[string]tftypes.Value{
		attrSnapshotRepository: tfVal(ctx, t, types.StringValue("repo-a")),
		attrForceMergeIndex:    tfVal(ctx, t, index),
		attrForceMergeOnClone:  tfVal(ctx, t, clone),
	})
	phaseVal := tfObjectFill(phaseType, map[string]tftypes.Value{
		attrMinAge:                  tfVal(ctx, t, types.StringValue("30d")),
		ilmActionSearchableSnapshot: ssVal,
	})
	return tfObjectFill(rootType, map[string]tftypes.Value{
		phaseName: phaseVal,
	})
}

func tfObjectFill(typ tftypes.Object, attrs map[string]tftypes.Value) tftypes.Value {
	full := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for name, attrType := range typ.AttributeTypes {
		if v, ok := attrs[name]; ok {
			full[name] = v
			continue
		}
		full[name] = tftypes.NewValue(attrType, nil)
	}
	return tftypes.NewValue(typ, full)
}

func tfVal(ctx context.Context, t *testing.T, v attr.Value) tftypes.Value {
	t.Helper()
	tv, err := v.ToTerraformValue(ctx)
	require.NoError(t, err)
	return tv
}
