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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type lensVisualizationConverter interface {
	vizType() string
	handlesTFConfigBlocks(blocks *lensByValueChartBlocks) bool
	populateFromAttributes(ctx context.Context, dashboard *dashboardModel, tfPanel *panelModel, blocks *lensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics
	buildAttributes(blocks *lensByValueChartBlocks, dashboard *dashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics)
}

type lensVisualizationBase struct {
	visualizationType string
	hasTFChartBlock   func(blocks *lensByValueChartBlocks) bool
}

func (c lensVisualizationBase) vizType() string {
	return c.visualizationType
}

func (c lensVisualizationBase) handlesTFConfigBlocks(blocks *lensByValueChartBlocks) bool {
	if blocks == nil || c.hasTFChartBlock == nil {
		return false
	}
	return c.hasTFChartBlock(blocks)
}

func detectLensVizType(attrs kbapi.KbnDashboardPanelTypeVisConfig0) string {
	if chart, err := attrs.AsXyChartNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsXyChartESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsTreemapNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsTreemapESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsMosaicNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsMosaicESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsDatatableNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsDatatableESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsTagcloudNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsTagcloudESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsHeatmapNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsHeatmapESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsRegionMapNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsRegionMapESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsLegacyMetricNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsMetricNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsMetricESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsPieNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsPieESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsGaugeNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsGaugeESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsWaffleNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsWaffleESQL(); err == nil {
		return string(chart.Type)
	}
	return ""
}

// lensVizConverterForType returns the typed Lens converter for viz_config.by_value whose discriminator
// matches strings produced by detectLensVizType, or nil when the provider does not model that chart kind.
func lensVizConverterForType(vizType string) lensVisualizationConverter {
	if vizType == "" {
		return nil
	}
	for _, c := range lensVizConverters {
		if c.vizType() == vizType {
			return c
		}
	}
	return nil
}
