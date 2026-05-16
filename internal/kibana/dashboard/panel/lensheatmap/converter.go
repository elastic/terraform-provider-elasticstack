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

package lensheatmap

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func init() {
	lenscommon.Register(converter{})
}

type converter struct{}

func (converter) VizType() string {
	return string(kbapi.HeatmapNoESQLTypeHeatmap)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.HeatmapConfig != nil
}

func (converter) SchemaAttribute() schema.Attribute {
	return lenscommon.ByValueChartNestedAttribute("heatmap_config", heatmapSchemaAttrs(true))
}

func (converter) PopulateFromAttributes(ctx context.Context, resolver lenscommon.Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	var prior *models.HeatmapConfigModel
	if blocks != nil && blocks.HeatmapConfig != nil {
		cpy := *blocks.HeatmapConfig
		prior = &cpy
	}
	if blocks == nil {
		var d diag.Diagnostics
		d.AddError("Lens chart blocks missing", "cannot populate heatmap_config without chart blocks")
		return d
	}
	blocks.HeatmapConfig = &models.HeatmapConfigModel{}

	if heatmapNoESQL, err := attrs.AsHeatmapNoESQL(); err == nil && !isHeatmapNoESQLCandidateActuallyESQL(heatmapNoESQL) {
		return heatmapConfigFromAPINoESQL(ctx, blocks.HeatmapConfig, resolver, prior, heatmapNoESQL)
	}
	heatmapESQL, err := attrs.AsHeatmapESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return heatmapConfigFromAPIESQL(ctx, blocks.HeatmapConfig, resolver, prior, heatmapESQL)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks, resolver lenscommon.Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics
	if blocks == nil {
		return attrs, diags
	}
	attrs, heatmapDiags := heatmapConfigToAPI(blocks.HeatmapConfig, resolver)
	diags.Append(heatmapDiags...)
	return attrs, diags
}

func (converter) AlignStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	if plan == nil || state == nil {
		return
	}
	if plan.HeatmapConfig == nil || state.HeatmapConfig == nil {
		return
	}
	alignHeatmapStateFromPlan(ctx, plan.HeatmapConfig, state.HeatmapConfig)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populateHeatmapLensAttributes(attrs)
}
