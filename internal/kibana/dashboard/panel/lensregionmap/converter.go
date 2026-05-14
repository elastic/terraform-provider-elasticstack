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
	attrs := lenscommon.LensChartBaseAttributes()
	attrs["data_source_json"] = lenscommon.DataSourceJSONAttribute(
		"Dataset configuration as JSON. For ES|QL, this specifies the ES|QL query. For standard layers, this specifies the data view and query.",
	)
	attrs["query"] = lenscommon.QueryAttribute(
		"Query configuration for filtering data. Required for non-ES|QL region map configurations.",
	)
	attrs["metric_json"] = lenscommon.MetricJSONAttribute(
		"Metric configuration as JSON. For ES|QL, this defines the metric column and format. For standard mode, this defines the metric operation or formula.",
		lenscommon.PopulateRegionMapMetricDefaults, true, "",
	)
	attrs["region_json"] = schema.StringAttribute{
		MarkdownDescription: "Region configuration as JSON. For ES|QL, this defines the region column and EMS join. " +
			"For standard mode, this defines the bucket operation (terms, histogram, range, filters) and optional EMS settings.",
		CustomType: jsontypes.NormalizedType{},
		Required:   true,
	}
	maps.Copy(attrs, lenscommon.LensChartPresentationAttributes())
	return lenscommon.ByValueChartNestedAttribute("region_map_config", attrs)
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
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics
	if blocks == nil {
		return attrs, diags
	}
	return regionMapConfigToAPI(blocks.RegionMapConfig, resolver)
}

func (converter) AlignStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	alignRegionMapStateFromPlan(ctx, plan, state)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populateRegionMapLensAttributes(attrs)
}
