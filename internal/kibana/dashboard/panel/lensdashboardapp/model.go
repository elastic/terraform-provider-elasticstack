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

package lensdashboardapp

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
)

// lensConfigClass identifies how a lens-dashboard-app panel `config` JSON should be
// interpreted before trusting generated union helpers alone.
type lensConfigClass int

const (
	// lensConfigClassByValueChart means the payload has a non-empty string at top-level "type".
	lensConfigClassByValueChart lensConfigClass = iota
	// lensConfigClassByReference means ref_id plus time_range.from/to (by-reference shape).
	lensConfigClassByReference
	// lensConfigClassAmbiguous means neither a chart payload nor a complete by-reference shape.
	lensConfigClassAmbiguous
)

func classifyLensDashboardAppConfigFromRoot(root map[string]any) lensConfigClass {
	switch {
	case hasLensByValueChartTypeAtRoot(root):
		return lensConfigClassByValueChart
	case hasLensByReferenceShapeAtRoot(root):
		return lensConfigClassByReference
	default:
		return lensConfigClassAmbiguous
	}
}

func hasLensByValueChartTypeAtRoot(m map[string]any) bool {
	if m == nil {
		return false
	}
	v, ok := m["type"]
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s != ""
}

func hasLensByReferenceShapeAtRoot(m map[string]any) bool {
	if m == nil {
		return false
	}
	ref, ok := m["ref_id"]
	if !ok {
		return false
	}
	refS, ok := ref.(string)
	if !ok || refS == "" {
		return false
	}
	trAny, ok := m["time_range"]
	if !ok {
		return false
	}
	tr, ok := trAny.(map[string]any)
	if !ok {
		return false
	}
	from, fOK := tr["from"].(string)
	to, tOK := tr["to"].(string)
	return fOK && tOK && from != "" && to != ""
}

// configPriorForLensRead returns the last known lens_dashboard_app_config from plan/state.
func configPriorForLensRead(tfPanel, pm *models.PanelModel) *models.LensDashboardAppConfigModel {
	if tfPanel != nil && tfPanel.LensDashboardAppConfig != nil {
		return tfPanel.LensDashboardAppConfig
	}
	if pm != nil {
		return pm.LensDashboardAppConfig
	}
	return nil
}

// terraformChartBlockKey maps kbapi chart discriminators to Terraform attribute names under by_value.
func terraformChartBlockKey(vizType string) string {
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
