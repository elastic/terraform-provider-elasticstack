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

package lenscommon

import "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"

// TerraformChartBlockKey maps a kbapi Lens chart discriminator (the value of VizConverter.VizType())
// to the Terraform attribute name used inside `vis_config.by_value` and `lens_dashboard_app_config.by_value`.
// Returns "" if vizType is not one of the supported Lens chart kinds.
func TerraformChartBlockKey(vizType string) string {
	switch vizType {
	case string(kbapi.XyChartNoESQLTypeXy):
		return "xy_chart_config"
	case string(kbapi.DatatableNoESQLTypeDataTable):
		return "datatable_config"
	case string(kbapi.TagcloudNoESQLTypeTagCloud):
		return "tagcloud_config"
	case string(kbapi.RegionMapNoESQLTypeRegionMap):
		return "region_map_config"
	case string(kbapi.PieNoESQLTypePie):
		return "pie_chart_config"
	case string(kbapi.MetricNoESQLTypeMetric):
		return "metric_chart_config"
	case string(kbapi.LegacyMetric):
		return "legacy_metric_config"
	case string(kbapi.GaugeNoESQLTypeGauge):
		return "gauge_config"
	case string(kbapi.HeatmapNoESQLTypeHeatmap):
		return "heatmap_config"
	case string(kbapi.MosaicNoESQLTypeMosaic):
		return "mosaic_config"
	case string(kbapi.TreemapNoESQLTypeTreemap):
		return "treemap_config"
	case string(kbapi.WaffleNoESQLTypeWaffle):
		return "waffle_config"
	default:
		return ""
	}
}
