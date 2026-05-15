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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_seedLensChartPriorIntoBlocks_perChart(t *testing.T) {
	t.Parallel()

	type row struct {
		name    string
		vizType string
		tfPanel *models.PanelModel
		dest    *models.LensByValueChartBlocks
		seedWaf bool
		assert  func(t *testing.T, dest *models.LensByValueChartBlocks)
	}

	for _, tc := range []row{
		{
			name:    "xy",
			vizType: string(kbapi.XyChartNoESQLTypeXy),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				XYChartConfig: &models.XYChartConfigModel{Title: types.StringValue("xy_prior")},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.XYChartConfig)
				require.Equal(t, "xy_prior", dest.XYChartConfig.Title.ValueString())
			},
		},
		{
			name:    "treemap",
			vizType: string(kbapi.TreemapNoESQLTypeTreemap),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				TreemapConfig: &models.TreemapConfigModel{Title: types.StringValue("treemap_prior")},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.TreemapConfig)
				require.Equal(t, "treemap_prior", dest.TreemapConfig.Title.ValueString())
			},
		},
		{
			name:    "mosaic",
			vizType: string(kbapi.MosaicNoESQLTypeMosaic),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				MosaicConfig: &models.MosaicConfigModel{Title: types.StringValue("mosaic_prior")},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.MosaicConfig)
				require.Equal(t, "mosaic_prior", dest.MosaicConfig.Title.ValueString())
			},
		},
		{
			name:    "datatable",
			vizType: string(kbapi.DatatableNoESQLTypeDataTable),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				DatatableConfig: &models.DatatableConfigModel{
					NoESQL: &models.DatatableNoESQLConfigModel{Title: types.StringValue("datatable_prior")},
				},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.DatatableConfig)
				require.NotNil(t, dest.DatatableConfig.NoESQL)
				require.Equal(t, "datatable_prior", dest.DatatableConfig.NoESQL.Title.ValueString())
			},
		},
		{
			name:    "tagcloud",
			vizType: string(kbapi.TagcloudNoESQLTypeTagCloud),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				TagcloudConfig: &models.TagcloudConfigModel{Title: types.StringValue("tagcloud_prior")},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.TagcloudConfig)
				require.Equal(t, "tagcloud_prior", dest.TagcloudConfig.Title.ValueString())
			},
		},
		{
			name:    "heatmap",
			vizType: string(kbapi.HeatmapNoESQLTypeHeatmap),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				HeatmapConfig: &models.HeatmapConfigModel{Title: types.StringValue("heatmap_prior")},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.HeatmapConfig)
				require.Equal(t, "heatmap_prior", dest.HeatmapConfig.Title.ValueString())
			},
		},
		{
			name:    "region_map",
			vizType: string(kbapi.RegionMapNoESQLTypeRegionMap),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				RegionMapConfig: &models.RegionMapConfigModel{Title: types.StringValue("region_prior")},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.RegionMapConfig)
				require.Equal(t, "region_prior", dest.RegionMapConfig.Title.ValueString())
			},
		},
		{
			name:    "legacy_metric",
			vizType: string(kbapi.LegacyMetric),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				LegacyMetricConfig: &models.LegacyMetricConfigModel{Title: types.StringValue("legacy_prior")},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.LegacyMetricConfig)
				require.Equal(t, "legacy_prior", dest.LegacyMetricConfig.Title.ValueString())
			},
		},
		{
			name:    "metric",
			vizType: string(kbapi.MetricNoESQLTypeMetric),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				MetricChartConfig: &models.MetricChartConfigModel{
					MetricChartCoreTFModel: models.MetricChartCoreTFModel{
						Title: types.StringValue("metric_prior"),
					},
				},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.MetricChartConfig)
				require.Equal(t, "metric_prior", dest.MetricChartConfig.Title.ValueString())
			},
		},
		{
			name:    "pie",
			vizType: string(kbapi.PieNoESQLTypePie),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				PieChartConfig: &models.PieChartConfigModel{Title: types.StringValue("pie_prior")},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.PieChartConfig)
				require.Equal(t, "pie_prior", dest.PieChartConfig.Title.ValueString())
			},
		},
		{
			name:    "gauge",
			vizType: string(kbapi.GaugeNoESQLTypeGauge),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				GaugeConfig: &models.GaugeConfigModel{Title: types.StringValue("gauge_prior")},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.GaugeConfig)
				require.Equal(t, "gauge_prior", dest.GaugeConfig.Title.ValueString())
			},
		},
		{
			name:    "nil_tf_panel_xy_stays_nil",
			vizType: string(kbapi.XyChartNoESQLTypeXy),
			tfPanel: nil,
			dest:    &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.Nil(t, dest.XYChartConfig)
			},
		},
		{
			name:    "prior_panel_missing_xy_block",
			vizType: string(kbapi.XyChartNoESQLTypeXy),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				MosaicConfig: &models.MosaicConfigModel{Title: types.StringValue("only_mosaic")},
			}),
			dest: &models.LensByValueChartBlocks{},
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.Nil(t, dest.XYChartConfig)
			},
		},
		{
			name:    "waffle_pointer_semantics_via_seedWaffle",
			vizType: string(kbapi.WaffleNoESQLTypeWaffle),
			tfPanel: panelWithLensBlocks(models.LensByValueChartBlocks{
				WaffleConfig: &models.WaffleConfigModel{Title: types.StringValue("waffle_seed")},
			}),
			dest:    &models.LensByValueChartBlocks{},
			seedWaf: true,
			assert: func(t *testing.T, dest *models.LensByValueChartBlocks) {
				t.Helper()
				require.NotNil(t, dest.WaffleConfig)
				require.Equal(t, "waffle_seed", dest.WaffleConfig.Title.ValueString())
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.seedWaf {
				seedWaffleLensByValueChartFromPriorPanel(tc.dest, tc.tfPanel)
			}
			seedLensChartPriorIntoBlocks(tc.tfPanel, tc.dest, tc.vizType)
			tc.assert(t, tc.dest)
		})
	}
}

func panelWithLensBlocks(blocks models.LensByValueChartBlocks) *models.PanelModel {
	return &models.PanelModel{
		VisConfig: &models.VisConfigModel{
			ByValue: &models.VisByValueModel{
				LensByValueChartBlocks: blocks,
			},
		},
	}
}
