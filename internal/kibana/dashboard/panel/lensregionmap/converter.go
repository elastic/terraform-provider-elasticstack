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
	"encoding/json"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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
	return string(kbapi.KibanaHTTPAPIsRegionMapNoESQLTypeRegionMap)
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

func (converter) PopulateFromAttributes(ctx context.Context, blocks *models.LensByValueChartBlocks, attrs lenscommon.LensByValueConfig) diag.Diagnostics {
	if diags := lenscommon.ValidateLensBlocks(blocks, "region_map_config"); diags.HasError() {
		return diags
	}
	prior := lenscommon.SnapshotAndResetBlock(&blocks.RegionMapConfig)
	diags := lenscommon.PopulateFromNoESQLOrESQL(
		ctx, blocks.RegionMapConfig, prior,
		func() (kbapi.KibanaHTTPAPIsRegionMapNoESQL, error) {
			var chart kbapi.KibanaHTTPAPIsRegionMapNoESQL
			return chart, regionMapChartFromLensAPI(attrs.Chart, &chart)
		},
		func() (kbapi.KibanaHTTPAPIsRegionMapESQL, error) {
			var chart kbapi.KibanaHTTPAPIsRegionMapESQL
			return chart, regionMapChartFromLensAPI(attrs.Chart, &chart)
		},
		func(v kbapi.KibanaHTTPAPIsRegionMapNoESQL) bool {
			return !lenscommon.IsNoESQLCandidateActuallyESQL(v.DataSource)
		},
		regionMapConfigFromAPINoESQL,
		regionMapConfigFromAPIESQL,
	)
	if diags.HasError() {
		return diags
	}
	if !lenscommon.PopulateLensChartPresentation(
		ctx, &blocks.RegionMapConfig.LensChartPresentationTFModel, prior,
		attrs.Presentation.TimeRange, attrs.Presentation.HideTitle, attrs.Presentation.HideBorder,
		attrs.Presentation.References, attrs.Presentation.Drilldowns, &diags,
	) {
		return diags
	}
	diags.Append(populateRegionMapRawDimensions(blocks.RegionMapConfig, attrs.Chart)...)
	return diags
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks) (lenscommon.LensByValueConfig, diag.Diagnostics) {
	var attrs lenscommon.LensByValueConfig
	var diags diag.Diagnostics
	if blocks == nil {
		return attrs, diags
	}
	return regionMapConfigToAPI(blocks.RegionMapConfig)
}

func (converter) AlignStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	alignRegionMapStateFromPlan(ctx, plan, state)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populateRegionMapLensAttributes(attrs)
}

func populateRegionMapRawDimensions(m *models.RegionMapConfigModel, chart kbapi.KibanaHTTPAPIsLensApiConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	raw, err := json.Marshal(chart)
	if err != nil {
		diags.AddError("Failed to marshal region map chart", err.Error())
		return diags
	}
	var payload struct {
		Metric json.RawMessage `json:"metric"`
		Region json.RawMessage `json:"region"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		diags.AddError("Failed to decode region map dimensions", err.Error())
		return diags
	}
	if payload.Metric != nil {
		m.MetricJSON = customtypes.NewJSONWithDefaultsValue(string(payload.Metric), lenscommon.PopulateRegionMapMetricDefaults)
	}
	if payload.Region != nil {
		m.RegionJSON = jsontypes.NewNormalizedValue(string(payload.Region))
	}
	return diags
}
