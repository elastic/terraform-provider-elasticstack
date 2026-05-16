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

package dashboard

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_alignPanelStateFromPlan_preservesCommonPanelFields(t *testing.T) {
	planPanels := []models.PanelModel{
		{
			VisConfig: &models.VisConfigModel{
				ByValue: &models.VisByValueModel{
					LensByValueChartBlocks: models.LensByValueChartBlocks{
						MosaicConfig: &models.MosaicConfigModel{
							Title:       types.StringValue("Sample Mosaic"),
							Description: types.StringValue("Test mosaic visualization"),
						},
					},
				},
			},
		},
		{
			Type: types.StringValue("esql_control"),
			EsqlControlConfig: &models.EsqlControlConfigModel{
				EsqlQuery:        types.StringValue("FROM logs-* | KEEP host.name"),
				Title:            types.StringValue("Fields Control"),
				AvailableOptions: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("option_a")}),
			},
		},
		{
			VisConfig: &models.VisConfigModel{
				ByValue: &models.VisByValueModel{
					LensByValueChartBlocks: models.LensByValueChartBlocks{
						TagcloudConfig: &models.TagcloudConfigModel{
							Title:       types.StringValue("Sample Tagcloud"),
							Description: types.StringValue("Test tagcloud visualization"),
							TagByJSON:   mustTagcloudJSON(`{"operation":"terms","fields":["host.name"],"limit":10}`),
						},
					},
				},
			},
		},
	}

	statePanels := []models.PanelModel{
		{
			VisConfig: &models.VisConfigModel{
				ByValue: &models.VisByValueModel{
					LensByValueChartBlocks: models.LensByValueChartBlocks{
						MosaicConfig: &models.MosaicConfigModel{
							Title:       types.StringValue(""),
							Description: types.StringValue(""),
						},
					},
				},
			},
		},
		{
			Type: types.StringValue("esql_control"),
			EsqlControlConfig: &models.EsqlControlConfigModel{
				EsqlQuery:        types.StringValue(""),
				Title:            types.StringValue(""),
				AvailableOptions: types.ListNull(types.StringType),
			},
		},
		{
			VisConfig: &models.VisConfigModel{
				ByValue: &models.VisByValueModel{
					LensByValueChartBlocks: models.LensByValueChartBlocks{
						TagcloudConfig: &models.TagcloudConfigModel{
							Title:       types.StringValue(""),
							Description: types.StringValue(""),
							TagByJSON: mustTagcloudJSON(
								`{"operation":"terms","fields":["host.name"],"limit":10,"rank_by":{"type":"metric","metric_index":0,"direction":"desc"}}`,
							),
						},
					},
				},
			},
		},
	}

	for i := range min(len(planPanels), len(statePanels)) {
		alignPanelStateFromPlan(t.Context(), &planPanels[i], &statePanels[i])
	}

	assert.Equal(t, planPanels[0].VisConfig.ByValue.MosaicConfig.Title, statePanels[0].VisConfig.ByValue.MosaicConfig.Title)
	assert.Equal(t, planPanels[0].VisConfig.ByValue.MosaicConfig.Description, statePanels[0].VisConfig.ByValue.MosaicConfig.Description)
	assert.Equal(t, planPanels[1].EsqlControlConfig.EsqlQuery, statePanels[1].EsqlControlConfig.EsqlQuery)
	assert.Equal(t, planPanels[1].EsqlControlConfig.Title, statePanels[1].EsqlControlConfig.Title)
	assert.Equal(t, planPanels[1].EsqlControlConfig.AvailableOptions, statePanels[1].EsqlControlConfig.AvailableOptions)
	assert.Equal(t, planPanels[2].VisConfig.ByValue.TagcloudConfig.Title, statePanels[2].VisConfig.ByValue.TagcloudConfig.Title)
	assert.Equal(t, planPanels[2].VisConfig.ByValue.TagcloudConfig.Description, statePanels[2].VisConfig.ByValue.TagcloudConfig.Description)
	assert.Equal(t, planPanels[2].VisConfig.ByValue.TagcloudConfig.TagByJSON.ValueString(), statePanels[2].VisConfig.ByValue.TagcloudConfig.TagByJSON.ValueString())
}

func Test_alignPanelStateFromPlan_dispatchesRegisteredHandlers(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "esql_control_invokes_Handler_AlignStateFromPlan",
			run: func(t *testing.T) {
				t.Helper()
				plan := models.PanelModel{
					Type: types.StringValue("esql_control"),
					EsqlControlConfig: &models.EsqlControlConfigModel{
						EsqlQuery:        types.StringValue("FROM logs-*"),
						Title:            types.StringValue("t"),
						AvailableOptions: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
					},
				}
				state := models.PanelModel{
					Type: types.StringValue("esql_control"),
					EsqlControlConfig: &models.EsqlControlConfigModel{
						EsqlQuery:        types.StringValue(""),
						Title:            types.StringValue(""),
						AvailableOptions: types.ListNull(types.StringType),
					},
				}
				alignPanelStateFromPlan(ctx, &plan, &state)
				assert.Equal(t, plan.EsqlControlConfig.EsqlQuery, state.EsqlControlConfig.EsqlQuery)
				assert.Equal(t, plan.EsqlControlConfig.Title, state.EsqlControlConfig.Title)
				assert.Equal(t, plan.EsqlControlConfig.AvailableOptions, state.EsqlControlConfig.AvailableOptions)
			},
		},
		{
			name: "unregistered_discriminator_skips_handler_specific_alignment",
			run: func(t *testing.T) {
				t.Helper()
				plan := models.PanelModel{
					Type: types.StringValue("unknown_panel_xyz"),
					EsqlControlConfig: &models.EsqlControlConfigModel{
						EsqlQuery: types.StringValue("FROM logs-*"),
						Title:     types.StringValue("would_align_if_registered"),
					},
				}
				state := models.PanelModel{
					Type: types.StringValue("unknown_panel_xyz"),
					EsqlControlConfig: &models.EsqlControlConfigModel{
						EsqlQuery: types.StringValue(""),
						Title:     types.StringValue(""),
					},
				}
				alignPanelStateFromPlan(ctx, &plan, &state)
				assert.Empty(t, state.EsqlControlConfig.EsqlQuery.ValueString())
				assert.Empty(t, state.EsqlControlConfig.Title.ValueString())
			},
		},
		{
			name: "registered_no_op_handler",
			run: func(t *testing.T) {
				t.Helper()
				plan := models.PanelModel{Type: types.StringValue("slo_burn_rate")}
				state := models.PanelModel{Type: types.StringValue("slo_burn_rate")}
				alignPanelStateFromPlan(ctx, &plan, &state)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func Test_alignPanelStateFromPlan_preservesMosaicTreemapPartitionSnapshots(t *testing.T) {
	plan := models.PanelModel{
		VisConfig: &models.VisConfigModel{
			ByValue: &models.VisByValueModel{
				LensByValueChartBlocks: models.LensByValueChartBlocks{
					MosaicConfig: &models.MosaicConfigModel{
						Title:               types.StringValue("M"),
						Description:         types.StringValue("d"),
						IgnoreGlobalFilters: types.BoolValue(true),
						Sampling:            types.Float64Value(0.5),
					},
					TreemapConfig: &models.TreemapConfigModel{
						Title:               types.StringValue("T"),
						Description:         types.StringValue("d"),
						IgnoreGlobalFilters: types.BoolValue(true),
						Sampling:            types.Float64Value(0.5),
					},
				},
			},
		},
	}
	state := models.PanelModel{
		VisConfig: &models.VisConfigModel{
			ByValue: &models.VisByValueModel{
				LensByValueChartBlocks: models.LensByValueChartBlocks{
					MosaicConfig: &models.MosaicConfigModel{
						Title:               types.StringValue("M"),
						Description:         types.StringValue("d"),
						IgnoreGlobalFilters: types.BoolNull(),
						Sampling:            types.Float64Null(),
					},
					TreemapConfig: &models.TreemapConfigModel{
						Title:               types.StringValue("T"),
						Description:         types.StringValue("d"),
						IgnoreGlobalFilters: types.BoolNull(),
						Sampling:            types.Float64Null(),
					},
				},
			},
		},
	}

	alignPanelStateFromPlan(t.Context(), &plan, &state)

	assert.True(t, state.VisConfig.ByValue.MosaicConfig.IgnoreGlobalFilters.ValueBool())
	assert.InEpsilon(t, 0.5, state.VisConfig.ByValue.MosaicConfig.Sampling.ValueFloat64(), 1e-9)
	assert.True(t, state.VisConfig.ByValue.TreemapConfig.IgnoreGlobalFilters.ValueBool())
	assert.InEpsilon(t, 0.5, state.VisConfig.ByValue.TreemapConfig.Sampling.ValueFloat64(), 1e-9)
}

func mustTagcloudJSON(v string) customtypes.JSONWithDefaultsValue[map[string]any] {
	return customtypes.NewJSONWithDefaultsValue(v, populateTagcloudTagByDefaults)
}

// Test_alignPanelStateFromPlan_pinnedPanel_xyChart_appliesAlignment verifies XY drift alignment runs through
// alignPanelStateFromPlan (the same helper alignDashboardStateFromPlanPinnedPanels delegates to once pinned panels
// expose Lens vis blocks on their synthetic PanelModel surface).
func Test_alignPanelStateFromPlan_pinnedPanel_xyChart_appliesAlignment(t *testing.T) {
	t.Parallel()

	plan := models.PanelModel{
		VisConfig: &models.VisConfigModel{
			ByValue: &models.VisByValueModel{
				LensByValueChartBlocks: models.LensByValueChartBlocks{
					XYChartConfig: &models.XYChartConfigModel{
						Legend: &models.XYLegendModel{
							Visibility: types.StringValue("visible"),
							Inside:     types.BoolValue(false),
							Position:   types.StringValue("right"),
						},
					},
				},
			},
		},
	}
	state := models.PanelModel{
		VisConfig: &models.VisConfigModel{
			ByValue: &models.VisByValueModel{
				LensByValueChartBlocks: models.LensByValueChartBlocks{
					XYChartConfig: &models.XYChartConfigModel{
						Legend: &models.XYLegendModel{
							Visibility: types.StringValue("visible"),
							Inside:     types.BoolNull(),
							Position:   types.StringNull(),
						},
					},
				},
			},
		},
	}

	alignPanelStateFromPlan(t.Context(), &plan, &state)

	got := state.VisConfig.ByValue.XYChartConfig.Legend
	require.False(t, got.Inside.IsNull())
	require.False(t, got.Inside.ValueBool())
	require.Equal(t, "right", got.Position.ValueString())
}
