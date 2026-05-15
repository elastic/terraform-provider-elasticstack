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

package lenstreemap

import (
	"context"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func init() {
	lenscommon.Register(converter{})
}

type converter struct{}

func (converter) VizType() string {
	return string(kbapi.TreemapNoESQLTypeTreemap)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.TreemapConfig != nil
}

func (converter) SchemaAttribute() schema.Attribute {
	attrs := lenscommon.PartitionChartBaseAttributes(true)
	treemapSpecific := map[string]schema.Attribute{
		"group_by_json": schema.StringAttribute{
			MarkdownDescription: "Array of breakdown dimensions as JSON (minimum 1). " +
				"For non-ES|QL, each item can be date histogram, terms, histogram, range, or filters operations; " +
				"for ES|QL, each item is the column/operation/color configuration.",
			CustomType: customtypes.NewJSONWithDefaultsType(lenscommon.PopulatePartitionGroupByDefaults),
			Optional:   true,
			Validators: lenscommon.MutuallyExclusiveStringValidator("esql_group_by"),
		},
		"metrics_json": schema.StringAttribute{
			MarkdownDescription: "Array of metric configurations as JSON (minimum 1). " +
				"For non-ES|QL, each item can be a field metric, pipeline metric, or formula; " +
				"for ES|QL, each item is the column/operation/color/format configuration.",
			CustomType: customtypes.NewJSONWithDefaultsType(lenscommon.PopulatePartitionMetricsDefaults),
			Optional:   true,
			Validators: lenscommon.MutuallyExclusiveStringValidator("esql_metrics"),
		},
		"legend": schema.SingleNestedAttribute{
			MarkdownDescription: "Legend configuration for the treemap chart.",
			Required:            true,
			Attributes:          lenscommon.PartitionLegendSchemaAttributes(),
		},
		"value_display": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for displaying values in chart cells.",
			Optional:            true,
			Attributes:          lenscommon.PartitionValueDisplaySchemaAttributes(),
		},
		"esql_metrics": schema.ListNestedAttribute{
			MarkdownDescription: "Metric columns for ES|QL treemaps. Mutually exclusive with `metrics_json`.",
			Optional:            true,
			NestedObject:        lenscommon.PartitionESQLMetricNestedObject(),
			Validators:          lenscommon.MutuallyExclusiveListValidator("metrics_json"),
		},
		"esql_group_by": schema.ListNestedAttribute{
			MarkdownDescription: "Breakdown columns for ES|QL treemaps. Mutually exclusive with `group_by_json`.",
			Optional:            true,
			NestedObject:        lenscommon.PartitionESQLGroupByNestedObject(),
			Validators:          lenscommon.MutuallyExclusiveListValidator("group_by_json"),
		},
	}
	maps.Copy(attrs, treemapSpecific)
	return lenscommon.ByValueChartNestedAttribute("treemap_config", attrs)
}

func (converter) PopulateFromAttributes(ctx context.Context, resolver lenscommon.Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	var prior *models.TreemapConfigModel
	if blocks != nil && blocks.TreemapConfig != nil {
		cpy := *blocks.TreemapConfig
		prior = &cpy
	}
	if blocks == nil {
		var d diag.Diagnostics
		d.AddError("Lens chart blocks missing", "cannot populate treemap_config without chart blocks")
		return d
	}
	blocks.TreemapConfig = &models.TreemapConfigModel{}

	if noESQL, err := attrs.AsTreemapNoESQL(); err == nil && !isTreemapNoESQLCandidateActuallyESQL(noESQL) {
		return treemapConfigFromAPINoESQL(ctx, blocks.TreemapConfig, resolver, prior, noESQL)
	}

	esql, err := attrs.AsTreemapESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return treemapConfigFromAPIESQL(ctx, blocks.TreemapConfig, resolver, prior, esql)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks, resolver lenscommon.Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics
	if blocks == nil {
		return attrs, diags
	}
	return treemapConfigToAPI(blocks.TreemapConfig, resolver)
}

func (converter) AlignStateFromPlan(_ context.Context, plan, state *models.LensByValueChartBlocks) {
	alignTreemapStateFromPlan(plan, state)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populateTreemapLensAttributes(attrs)
}
