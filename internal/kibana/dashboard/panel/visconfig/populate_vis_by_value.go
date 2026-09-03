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

package visconfig

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// lensByValueChartBlocksFromPanel returns vis_config.by_value chart blocks when populated.
func lensByValueChartBlocksFromPanel(pm *models.PanelModel) *models.LensByValueChartBlocks {
	if pm == nil || pm.VisConfig == nil || pm.VisConfig.ByValue == nil {
		return nil
	}
	return &pm.VisConfig.ByValue.LensByValueChartBlocks
}

func seedWaffleLensByValueChartFromPriorPanel(dest *models.LensByValueChartBlocks, prior *models.PanelModel) {
	if dest == nil || prior == nil || prior.VisConfig == nil || prior.VisConfig.ByValue == nil {
		return
	}
	src := &prior.VisConfig.ByValue.LensByValueChartBlocks
	if src.WaffleConfig != nil {
		dest.WaffleConfig = src.WaffleConfig
	}
}

// seedLensChartPriorIntoBlocks copies the chart block matching vizType from prior panel state into dest
// before PopulateFromAttributes merges API omissions with practitioner state.
func seedLensChartPriorIntoBlocks(tfPanel *models.PanelModel, dest *models.LensByValueChartBlocks, vizType string) {
	if dest == nil {
		return
	}
	prior := lensByValueChartBlocksFromPanel(tfPanel)
	switch vizType {
	case string(kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanelTypeXy):
		if prior != nil && prior.XYChartConfig != nil {
			cpy := *prior.XYChartConfig
			dest.XYChartConfig = &cpy
		} else {
			dest.XYChartConfig = nil
		}
	case string(kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanelTypeTreemap):
		if prior != nil && prior.TreemapConfig != nil {
			cpy := *prior.TreemapConfig
			dest.TreemapConfig = &cpy
		} else {
			dest.TreemapConfig = nil
		}
	case string(kbapi.KibanaHTTPAPIsMosaicNoESQLByValuePanelTypeMosaic):
		if prior != nil && prior.MosaicConfig != nil {
			cpy := *prior.MosaicConfig
			dest.MosaicConfig = &cpy
		} else {
			dest.MosaicConfig = nil
		}
	case string(kbapi.KibanaHTTPAPIsDatatableNoESQLByValuePanelTypeDataTable):
		if prior != nil && prior.DatatableConfig != nil {
			cpy := *prior.DatatableConfig
			dest.DatatableConfig = &cpy
		} else {
			dest.DatatableConfig = nil
		}
	case string(kbapi.KibanaHTTPAPIsTagcloudNoESQLByValuePanelTypeTagCloud):
		if prior != nil && prior.TagcloudConfig != nil {
			cpy := *prior.TagcloudConfig
			dest.TagcloudConfig = &cpy
		} else {
			dest.TagcloudConfig = nil
		}
	case string(kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanelTypeHeatmap):
		if prior != nil && prior.HeatmapConfig != nil {
			cpy := *prior.HeatmapConfig
			dest.HeatmapConfig = &cpy
		} else {
			dest.HeatmapConfig = nil
		}
	case string(kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanelTypeRegionMap):
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
	case string(kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric):
		if prior != nil && prior.MetricChartConfig != nil {
			cpy := *prior.MetricChartConfig
			dest.MetricChartConfig = &cpy
		} else {
			dest.MetricChartConfig = nil
		}
	case string(kbapi.KibanaHTTPAPIsPieNoESQLByValuePanelTypePie):
		if prior != nil && prior.PieChartConfig != nil {
			cpy := *prior.PieChartConfig
			dest.PieChartConfig = &cpy
		} else {
			dest.PieChartConfig = nil
		}
	case string(kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanelTypeGauge):
		if prior != nil && prior.GaugeConfig != nil {
			cpy := *prior.GaugeConfig
			dest.GaugeConfig = &cpy
		} else {
			dest.GaugeConfig = nil
		}
	case string(kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanelTypeWaffle):
	default:
	}
}

func populateLensVisByValueFromTypedChartAPI(
	ctx context.Context,
	tfPanel *models.PanelModel,
	blocks *models.LensByValueChartBlocks,
	config0 lenscommon.VisByValueConfig0,
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
	diags.Append(conv.PopulateFromAttributes(ctx, blocks, config0)...)
	return diags
}
