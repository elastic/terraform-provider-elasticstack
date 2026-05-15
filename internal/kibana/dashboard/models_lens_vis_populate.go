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
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// seedLensChartPriorIntoBlocks copies the chart block matching vizType from LensByValueChartBlocksFromPanel(tfPanel)
// into dest so lenscommon.VizConverter.PopulateFromAttributes can read prior-dependent merge state.
// vizType must be the kbapi discriminator string returned by lenscommon.DetectVizType.
// Waffle uses seedWaffleLensByValueChartFromPriorPanel instead of copying here.
func seedLensChartPriorIntoBlocks(tfPanel *models.PanelModel, dest *models.LensByValueChartBlocks, vizType string) {
	if dest == nil {
		return
	}
	prior := LensByValueChartBlocksFromPanel(tfPanel)
	switch vizType {
	case string(kbapi.XyChartNoESQLTypeXy):
		if prior != nil && prior.XYChartConfig != nil {
			cpy := *prior.XYChartConfig
			dest.XYChartConfig = &cpy
		} else {
			dest.XYChartConfig = nil
		}
	case string(kbapi.TreemapNoESQLTypeTreemap):
		if prior != nil && prior.TreemapConfig != nil {
			cpy := *prior.TreemapConfig
			dest.TreemapConfig = &cpy
		} else {
			dest.TreemapConfig = nil
		}
	case string(kbapi.MosaicNoESQLTypeMosaic):
		if prior != nil && prior.MosaicConfig != nil {
			cpy := *prior.MosaicConfig
			dest.MosaicConfig = &cpy
		} else {
			dest.MosaicConfig = nil
		}
	case string(kbapi.DatatableNoESQLTypeDataTable):
		if prior != nil && prior.DatatableConfig != nil {
			cpy := *prior.DatatableConfig
			dest.DatatableConfig = &cpy
		} else {
			dest.DatatableConfig = nil
		}
	case string(kbapi.TagcloudNoESQLTypeTagCloud):
		if prior != nil && prior.TagcloudConfig != nil {
			cpy := *prior.TagcloudConfig
			dest.TagcloudConfig = &cpy
		} else {
			dest.TagcloudConfig = nil
		}
	case string(kbapi.HeatmapNoESQLTypeHeatmap):
		if prior != nil && prior.HeatmapConfig != nil {
			cpy := *prior.HeatmapConfig
			dest.HeatmapConfig = &cpy
		} else {
			dest.HeatmapConfig = nil
		}
	case string(kbapi.RegionMapNoESQLTypeRegionMap):
		if prior != nil && prior.RegionMapConfig != nil {
			cpy := *prior.RegionMapConfig
			dest.RegionMapConfig = &cpy
		} else {
			dest.RegionMapConfig = nil
		}
	case string(kbapi.LegacyMetric):
		if prior != nil && prior.LegacyMetricConfig != nil {
			cpy := *prior.LegacyMetricConfig
			dest.LegacyMetricConfig = &cpy
		} else {
			dest.LegacyMetricConfig = nil
		}
	case string(kbapi.MetricNoESQLTypeMetric):
		if prior != nil && prior.MetricChartConfig != nil {
			cpy := *prior.MetricChartConfig
			dest.MetricChartConfig = &cpy
		} else {
			dest.MetricChartConfig = nil
		}
	case string(kbapi.PieNoESQLTypePie):
		if prior != nil && prior.PieChartConfig != nil {
			cpy := *prior.PieChartConfig
			dest.PieChartConfig = &cpy
		} else {
			dest.PieChartConfig = nil
		}
	case string(kbapi.GaugeNoESQLTypeGauge):
		if prior != nil && prior.GaugeConfig != nil {
			cpy := *prior.GaugeConfig
			dest.GaugeConfig = &cpy
		} else {
			dest.GaugeConfig = nil
		}
	case string(kbapi.WaffleNoESQLTypeWaffle):
		// Pointer semantics come from seedWaffleLensByValueChartFromPriorPanel.
	default:
	}
}

func populateLensVisByValueFromTypedChartAPI(
	ctx context.Context,
	dashboard *models.DashboardModel,
	tfPanel *models.PanelModel,
	blocks *models.LensByValueChartBlocks,
	config0 kbapi.KbnDashboardPanelTypeVisConfig0,
	unknownTypeAddsError bool,
) diag.Diagnostics {
	var diags diag.Diagnostics
	visType := lenscommon.DetectVizType(config0)
	if visType == "" {
		if unknownTypeAddsError {
			diags.AddError(
				"Unsupported visualization chart type",
				"The `vis` panel config has a top-level chart discriminator but could not resolve a Lens chart kind from the union; use panel-level `config_json` until this shape is modeled.",
			)
		}
		return diags
	}
	conv := lenscommon.ForType(visType)
	if conv == nil {
		diags.AddError(
			"Unsupported visualization chart type",
			fmt.Sprintf(
				"The dashboard returned Lens visualization discriminator %q which this provider does not support as typed `vis_config.by_value`. "+
					"Use panel-level `config_json` as the escape hatch to manage this panel until support is added.",
				visType,
			),
		)
		return diags
	}
	seedWaffleLensByValueChartFromPriorPanel(blocks, tfPanel)
	seedLensChartPriorIntoBlocks(tfPanel, blocks, visType)
	diags.Append(conv.PopulateFromAttributes(ctx, lensChartResolver(dashboard), blocks, config0)...)
	return diags
}
