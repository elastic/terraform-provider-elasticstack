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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AiopsLogRateAnalysisConfigModel models the aiops_log_rate_analysis_config block.
type AiopsLogRateAnalysisConfigModel struct {
	DataViewID  types.String    `tfsdk:"data_view_id"`
	Title       types.String    `tfsdk:"title"`
	Description types.String    `tfsdk:"description"`
	HideTitle   types.Bool      `tfsdk:"hide_title"`
	HideBorder  types.Bool      `tfsdk:"hide_border"`
	TimeRange   *TimeRangeModel `tfsdk:"time_range"`
}

// AiopsPatternAnalysisConfigModel models the aiops_pattern_analysis_config block.
type AiopsPatternAnalysisConfigModel struct {
	DataViewID               types.String    `tfsdk:"data_view_id"`
	FieldName                types.String    `tfsdk:"field_name"`
	MinimumTimeRange         types.String    `tfsdk:"minimum_time_range"`
	RandomSamplerMode        types.String    `tfsdk:"random_sampler_mode"`
	RandomSamplerProbability types.Float64   `tfsdk:"random_sampler_probability"`
	Title                    types.String    `tfsdk:"title"`
	Description              types.String    `tfsdk:"description"`
	HideTitle                types.Bool      `tfsdk:"hide_title"`
	HideBorder               types.Bool      `tfsdk:"hide_border"`
	TimeRange                *TimeRangeModel `tfsdk:"time_range"`
}

// AiopsChangePointChartConfigModel models the aiops_change_point_chart_config block.
type AiopsChangePointChartConfigModel struct {
	DataViewID          types.String    `tfsdk:"data_view_id"`
	MetricField         types.String    `tfsdk:"metric_field"`
	AggregationFunction types.String    `tfsdk:"aggregation_function"`
	SplitField          types.String    `tfsdk:"split_field"`
	Partitions          types.Set       `tfsdk:"partitions"`
	MaxSeriesToPlot     types.Float64   `tfsdk:"max_series_to_plot"`
	ViewType            types.String    `tfsdk:"view_type"`
	Title               types.String    `tfsdk:"title"`
	Description         types.String    `tfsdk:"description"`
	HideTitle           types.Bool      `tfsdk:"hide_title"`
	HideBorder          types.Bool      `tfsdk:"hide_border"`
	TimeRange           *TimeRangeModel `tfsdk:"time_range"`
}
