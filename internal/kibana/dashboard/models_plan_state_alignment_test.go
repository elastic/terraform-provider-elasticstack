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
