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
// to the Terraform attribute name used inside `vis_config.by_value`.
// Returns "" if vizType is not one of the supported Lens chart kinds.
func TerraformChartBlockKey(vizType string) string {
	switch vizType {
	case string(kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanelTypeXy):
		return "xy_chart_config"
	case string(kbapi.KibanaHTTPAPIsDatatableNoESQLByValuePanelTypeDataTable):
		return "datatable_config"
	case string(kbapi.KibanaHTTPAPIsTagcloudNoESQLByValuePanelTypeTagCloud):
		return "tagcloud_config"
	case string(kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanelTypeRegionMap):
		return "region_map_config"
	case string(kbapi.KibanaHTTPAPIsPieNoESQLByValuePanelTypePie):
		return "pie_chart_config"
	case string(kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric):
		return "metric_chart_config"
	case string(kbapi.LegacyMetric):
		return "legacy_metric_config"
	case string(kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanelTypeGauge):
		return "gauge_config"
	case string(kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanelTypeHeatmap):
		return "heatmap_config"
	case string(kbapi.KibanaHTTPAPIsMosaicNoESQLByValuePanelTypeMosaic):
		return "mosaic_config"
	case string(kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanelTypeTreemap):
		return "treemap_config"
	case string(kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanelTypeWaffle):
		return "waffle_config"
	default:
		return ""
	}
}
