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

package lensregionmap

import (
	"context"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func init() {
	lenscommon.Register(converter{})
}

type converter struct{}

func (converter) VizType() string {
	return string(kbapi.RegionMapNoESQLTypeRegionMap)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.RegionMapConfig != nil
}

func (converter) SchemaAttribute() schema.Attribute {
	attrs := maps.Clone(lenscommon.LensChartBaseAttributes())
	attrs["data_source_json"] = schema.StringAttribute{
		MarkdownDescription: "Dataset configuration as JSON. For ES|QL, this specifies the ES|QL query. For standard layers, this specifies the data view and query.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	attrs["query"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Query configuration for filtering data. Required for non-ES|QL region map configurations.",
		Optional:            true,
		Attributes:          lenscommon.LensChartFilterSimpleAttributes(),
	}
	attrs["metric_json"] = schema.StringAttribute{
		MarkdownDescription: "Metric configuration as JSON. For ES|QL, this defines the metric column and format. For standard mode, this defines the metric operation or formula.",
		CustomType:          customtypes.NewJSONWithDefaultsType(lenscommon.PopulateRegionMapMetricDefaults),
		Required:            true,
	}
	attrs["region_json"] = schema.StringAttribute{
		MarkdownDescription: "Region configuration as JSON. For ES|QL, this defines the region column and EMS join. " +
			"For standard mode, this defines the bucket operation (terms, histogram, range, filters) and optional EMS settings.",
		CustomType:          jsontypes.NormalizedType{},
		Required:            true,
	}
	maps.Copy(attrs, lenscommon.LensChartPresentationAttributes())
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Typed Lens visualization inside `vis_config.by_value`. " +
			"Mutually exclusive with the other chart blocks in the same `by_value` block. " +
			"Shares the attribute shape with `lens_dashboard_app_config.by_value.region_map_config`.",
		Optional:   true,
		Attributes: attrs,
	}
}

func (converter) PopulateFromAttributes(ctx context.Context, resolver lenscommon.Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	var prior *models.RegionMapConfigModel
	if blocks != nil && blocks.RegionMapConfig != nil {
		cpy := *blocks.RegionMapConfig
		prior = &cpy
	}
	if blocks == nil {
		var d diag.Diagnostics
		d.AddError("Lens chart blocks missing", "cannot populate region_map_config without chart blocks")
		return d
	}
	blocks.RegionMapConfig = &models.RegionMapConfigModel{}

	if noESQL, err := attrs.AsRegionMapNoESQL(); err == nil && !isRegionMapNoESQLCandidateActuallyESQL(noESQL) {
		return regionMapConfigFromAPINoESQL(ctx, blocks.RegionMapConfig, resolver, prior, noESQL)
	}

	regionMapESQL, err := attrs.AsRegionMapESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return regionMapConfigFromAPIESQL(ctx, blocks.RegionMapConfig, resolver, prior, regionMapESQL)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks, resolver lenscommon.Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	if blocks == nil || blocks.RegionMapConfig == nil {
		diags.AddError("Region map config missing", "region_map_config block is required")
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}
	configModel := *blocks.RegionMapConfig
	return regionMapConfigToAPI(&configModel, resolver)
}

func (converter) AlignStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	alignRegionMapStateFromPlan(ctx, plan, state)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populateRegionMapLensAttributes(attrs)
}
