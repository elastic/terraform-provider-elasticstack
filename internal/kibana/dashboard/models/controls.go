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

type TimeSliderControlConfigModel struct {
	StartPercentageOfTimeRange types.Float32 `tfsdk:"start_percentage_of_time_range"`
	EndPercentageOfTimeRange   types.Float32 `tfsdk:"end_percentage_of_time_range"`
	IsAnchored                 types.Bool    `tfsdk:"is_anchored"`
}

type OptionsListControlDisplaySettingsModel struct {
	Placeholder   types.String `tfsdk:"placeholder"`
	HideActionBar types.Bool   `tfsdk:"hide_action_bar"`
	HideExclude   types.Bool   `tfsdk:"hide_exclude"`
	HideExists    types.Bool   `tfsdk:"hide_exists"`
	HideSort      types.Bool   `tfsdk:"hide_sort"`
}

type OptionsListControlSortModel struct {
	By        types.String `tfsdk:"by"`
	Direction types.String `tfsdk:"direction"`
}

type OptionsListControlConfigModel struct {
	DataViewID        types.String                            `tfsdk:"data_view_id"`
	FieldName         types.String                            `tfsdk:"field_name"`
	Title             types.String                            `tfsdk:"title"`
	UseGlobalFilters  types.Bool                              `tfsdk:"use_global_filters"`
	IgnoreValidations types.Bool                              `tfsdk:"ignore_validations"`
	SingleSelect      types.Bool                              `tfsdk:"single_select"`
	Exclude           types.Bool                              `tfsdk:"exclude"`
	ExistsSelected    types.Bool                              `tfsdk:"exists_selected"`
	RunPastTimeout    types.Bool                              `tfsdk:"run_past_timeout"`
	SearchTechnique   types.String                            `tfsdk:"search_technique"`
	SelectedOptions   types.List                              `tfsdk:"selected_options"`
	DisplaySettings   *OptionsListControlDisplaySettingsModel `tfsdk:"display_settings"`
	Sort              *OptionsListControlSortModel            `tfsdk:"sort"`
}

type RangeSliderControlConfigModel struct {
	Title             types.String  `tfsdk:"title"`
	DataViewID        types.String  `tfsdk:"data_view_id"`
	FieldName         types.String  `tfsdk:"field_name"`
	UseGlobalFilters  types.Bool    `tfsdk:"use_global_filters"`
	IgnoreValidations types.Bool    `tfsdk:"ignore_validations"`
	Value             types.List    `tfsdk:"value"`
	Step              types.Float32 `tfsdk:"step"`
}

type EsqlControlDisplaySettingsModel struct {
	Placeholder   types.String `tfsdk:"placeholder"`
	HideActionBar types.Bool   `tfsdk:"hide_action_bar"`
	HideExclude   types.Bool   `tfsdk:"hide_exclude"`
	HideExists    types.Bool   `tfsdk:"hide_exists"`
	HideSort      types.Bool   `tfsdk:"hide_sort"`
}

type EsqlControlConfigModel struct {
	SelectedOptions  types.List                       `tfsdk:"selected_options"`
	VariableName     types.String                     `tfsdk:"variable_name"`
	VariableType     types.String                     `tfsdk:"variable_type"`
	EsqlQuery        types.String                     `tfsdk:"esql_query"`
	ControlType      types.String                     `tfsdk:"control_type"`
	Title            types.String                     `tfsdk:"title"`
	SingleSelect     types.Bool                       `tfsdk:"single_select"`
	AvailableOptions types.List                       `tfsdk:"available_options"`
	DisplaySettings  *EsqlControlDisplaySettingsModel `tfsdk:"display_settings"`
}
