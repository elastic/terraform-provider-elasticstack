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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_alignPanelStateFromPlan_preservesCommonPanelFields(t *testing.T) {
	planPanels := []panelModel{
		{
			VizConfig: &vizConfigModel{
				ByValue: &vizByValueModel{
					lensByValueChartBlocks: lensByValueChartBlocks{
						MosaicConfig: &mosaicConfigModel{
							Title:       types.StringValue("Sample Mosaic"),
							Description: types.StringValue("Test mosaic visualization"),
						},
					},
				},
			},
		},
		{
			EsqlControlConfig: &esqlControlConfigModel{
				EsqlQuery:        types.StringValue("FROM logs-* | KEEP host.name"),
				Title:            types.StringValue("Fields Control"),
				AvailableOptions: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("option_a")}),
			},
		},
		{
			VizConfig: &vizConfigModel{
				ByValue: &vizByValueModel{
					lensByValueChartBlocks: lensByValueChartBlocks{
						TagcloudConfig: &tagcloudConfigModel{
							Title:       types.StringValue("Sample Tagcloud"),
							Description: types.StringValue("Test tagcloud visualization"),
							TagByJSON:   mustTagcloudJSON(`{"operation":"terms","fields":["host.name"],"limit":10}`),
						},
					},
				},
			},
		},
	}

	statePanels := []panelModel{
		{
			VizConfig: &vizConfigModel{
				ByValue: &vizByValueModel{
					lensByValueChartBlocks: lensByValueChartBlocks{
						MosaicConfig: &mosaicConfigModel{
							Title:       types.StringValue(""),
							Description: types.StringValue(""),
						},
					},
				},
			},
		},
		{
			EsqlControlConfig: &esqlControlConfigModel{
				EsqlQuery:        types.StringValue(""),
				Title:            types.StringValue(""),
				AvailableOptions: types.ListNull(types.StringType),
			},
		},
		{
			VizConfig: &vizConfigModel{
				ByValue: &vizByValueModel{
					lensByValueChartBlocks: lensByValueChartBlocks{
						TagcloudConfig: &tagcloudConfigModel{
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

	assert.Equal(t, planPanels[0].VizConfig.ByValue.MosaicConfig.Title, statePanels[0].VizConfig.ByValue.MosaicConfig.Title)
	assert.Equal(t, planPanels[0].VizConfig.ByValue.MosaicConfig.Description, statePanels[0].VizConfig.ByValue.MosaicConfig.Description)
	assert.Equal(t, planPanels[1].EsqlControlConfig.EsqlQuery, statePanels[1].EsqlControlConfig.EsqlQuery)
	assert.Equal(t, planPanels[1].EsqlControlConfig.Title, statePanels[1].EsqlControlConfig.Title)
	assert.Equal(t, planPanels[1].EsqlControlConfig.AvailableOptions, statePanels[1].EsqlControlConfig.AvailableOptions)
	assert.Equal(t, planPanels[2].VizConfig.ByValue.TagcloudConfig.Title, statePanels[2].VizConfig.ByValue.TagcloudConfig.Title)
	assert.Equal(t, planPanels[2].VizConfig.ByValue.TagcloudConfig.Description, statePanels[2].VizConfig.ByValue.TagcloudConfig.Description)
	assert.Equal(t, planPanels[2].VizConfig.ByValue.TagcloudConfig.TagByJSON.ValueString(), statePanels[2].VizConfig.ByValue.TagcloudConfig.TagByJSON.ValueString())
}

func Test_alignPanelStateFromPlan_preservesMosaicTreemapPartitionSnapshots(t *testing.T) {
	plan := panelModel{
		VizConfig: &vizConfigModel{
			ByValue: &vizByValueModel{
				lensByValueChartBlocks: lensByValueChartBlocks{
					MosaicConfig: &mosaicConfigModel{
						Title:               types.StringValue("M"),
						Description:         types.StringValue("d"),
						IgnoreGlobalFilters: types.BoolValue(true),
						Sampling:            types.Float64Value(0.5),
					},
					TreemapConfig: &treemapConfigModel{
						Title:               types.StringValue("T"),
						Description:         types.StringValue("d"),
						IgnoreGlobalFilters: types.BoolValue(true),
						Sampling:            types.Float64Value(0.5),
					},
				},
			},
		},
	}
	state := panelModel{
		VizConfig: &vizConfigModel{
			ByValue: &vizByValueModel{
				lensByValueChartBlocks: lensByValueChartBlocks{
					MosaicConfig: &mosaicConfigModel{
						Title:               types.StringValue("M"),
						Description:         types.StringValue("d"),
						IgnoreGlobalFilters: types.BoolNull(),
						Sampling:            types.Float64Null(),
					},
					TreemapConfig: &treemapConfigModel{
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

	assert.True(t, state.VizConfig.ByValue.MosaicConfig.IgnoreGlobalFilters.ValueBool())
	assert.InEpsilon(t, 0.5, state.VizConfig.ByValue.MosaicConfig.Sampling.ValueFloat64(), 1e-9)
	assert.True(t, state.VizConfig.ByValue.TreemapConfig.IgnoreGlobalFilters.ValueBool())
	assert.InEpsilon(t, 0.5, state.VizConfig.ByValue.TreemapConfig.Sampling.ValueFloat64(), 1e-9)
}

func mustTagcloudJSON(v string) customtypes.JSONWithDefaultsValue[map[string]any] {
	return customtypes.NewJSONWithDefaultsValue(v, populateTagcloudTagByDefaults)
}
