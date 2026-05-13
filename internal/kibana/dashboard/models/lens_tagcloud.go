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

package models

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TagcloudConfigModel struct {
	LensChartPresentationTFModel
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized                              `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *FilterSimpleModel                                `tfsdk:"query"`
	Filters             []ChartFilterJSONModel                            `tfsdk:"filters"`
	Orientation         types.String                                      `tfsdk:"orientation"`
	FontSize            *FontSizeModel                                    `tfsdk:"font_size"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
	TagByJSON           customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"tag_by_json"`
	EsqlMetric          *TagcloudEsqlMetric                               `tfsdk:"esql_metric"`
	EsqlTagBy           *TagcloudEsqlTagBy                                `tfsdk:"esql_tag_by"`
}

type TagcloudEsqlMetric struct {
	Column     types.String         `tfsdk:"column"`
	FormatJSON jsontypes.Normalized `tfsdk:"format_json"`
	Label      types.String         `tfsdk:"label"`
}

type TagcloudEsqlTagBy struct {
	Column     types.String         `tfsdk:"column"`
	FormatJSON jsontypes.Normalized `tfsdk:"format_json"`
	ColorJSON  jsontypes.Normalized `tfsdk:"color_json"`
	Label      types.String         `tfsdk:"label"`
}

type FontSizeModel struct {
	Min types.Float64 `tfsdk:"min"`
	Max types.Float64 `tfsdk:"max"`
}
