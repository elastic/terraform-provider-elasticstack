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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// lensDashboardAppConfigModel is the TF model for a lens-dashboard-app panel.
type lensDashboardAppConfigModel struct {
	ByValue     *lensDashboardAppByValueModel     `tfsdk:"by_value"`
	ByReference *lensDashboardAppByReferenceModel `tfsdk:"by_reference"`
}

type lensDashboardAppByValueModel struct {
	ConfigJSON         jsontypes.Normalized           `tfsdk:"config_json"`
	XYChartConfig      *xyChartConfigModel            `tfsdk:"xy_chart_config"`
	TreemapConfig      *treemapConfigModel            `tfsdk:"treemap_config"`
	MosaicConfig       *mosaicConfigModel             `tfsdk:"mosaic_config"`
	DatatableConfig    *datatableConfigModel          `tfsdk:"datatable_config"`
	TagcloudConfig     *tagcloudConfigModel           `tfsdk:"tagcloud_config"`
	HeatmapConfig      *heatmapConfigModel            `tfsdk:"heatmap_config"`
	WaffleConfig       *waffleConfigModel             `tfsdk:"waffle_config"`
	RegionMapConfig    *regionMapConfigModel          `tfsdk:"region_map_config"`
	GaugeConfig        *gaugeConfigModel              `tfsdk:"gauge_config"`
	MetricChartConfig  *metricChartLensByValueTFModel `tfsdk:"metric_chart_config"`
	PieChartConfig     *pieChartConfigModel           `tfsdk:"pie_chart_config"`
	LegacyMetricConfig *legacyMetricConfigModel       `tfsdk:"legacy_metric_config"`
}

type lensDashboardAppByReferenceModel struct {
	RefID          types.String                   `tfsdk:"ref_id"`
	ReferencesJSON jsontypes.Normalized           `tfsdk:"references_json"`
	Title          types.String                   `tfsdk:"title"`
	Description    types.String                   `tfsdk:"description"`
	HideTitle      types.Bool                     `tfsdk:"hide_title"`
	HideBorder     types.Bool                     `tfsdk:"hide_border"`
	Drilldowns     drilldownsModel                `tfsdk:"drilldowns"`
	TimeRange      lensDashboardAppTimeRangeModel `tfsdk:"time_range"`
}

type lensDashboardAppTimeRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
	Mode types.String `tfsdk:"mode"`
}
