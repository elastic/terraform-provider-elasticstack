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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DiscoverSessionPanelConfigModel struct {
	Title       types.String                      `tfsdk:"title"`
	Description types.String                      `tfsdk:"description"`
	HideTitle   types.Bool                        `tfsdk:"hide_title"`
	HideBorder  types.Bool                        `tfsdk:"hide_border"`
	Drilldowns  []DiscoverSessionPanelDrilldown   `tfsdk:"drilldowns"`
	ByValue     *DiscoverSessionPanelByValueModel `tfsdk:"by_value"`
	ByReference *DiscoverSessionPanelByRefModel   `tfsdk:"by_reference"`
}

type DiscoverSessionPanelDrilldown struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type DiscoverSessionPanelByValueModel struct {
	TimeRange *TimeRangeModel         `tfsdk:"time_range"`
	Tab       DiscoverSessionTabModel `tfsdk:"tab"`
}

type DiscoverSessionTabModel struct {
	DSL  *DiscoverSessionDSLTabModel  `tfsdk:"dsl"`
	ESQL *DiscoverSessionESQLTabModel `tfsdk:"esql"`
}

type DiscoverSessionDSLTabModel struct {
	ColumnOrder     types.List                 `tfsdk:"column_order"`
	ColumnSettings  types.Map                  `tfsdk:"column_settings"`
	Sort            []DiscoverSessionSortModel `tfsdk:"sort"`
	Density         types.String               `tfsdk:"density"`
	HeaderRowHeight types.String               `tfsdk:"header_row_height"`
	RowHeight       types.String               `tfsdk:"row_height"`
	RowsPerPage     types.Int64                `tfsdk:"rows_per_page"`
	SampleSize      types.Int64                `tfsdk:"sample_size"`
	ViewMode        types.String               `tfsdk:"view_mode"`
	Query           *FilterSimpleModel         `tfsdk:"query"`
	DataSourceJSON  jsontypes.Normalized       `tfsdk:"data_source_json"`
	Filters         []ChartFilterJSONModel     `tfsdk:"filters"`
}

type DiscoverSessionESQLTabModel struct {
	ColumnOrder     types.List                 `tfsdk:"column_order"`
	ColumnSettings  types.Map                  `tfsdk:"column_settings"`
	Sort            []DiscoverSessionSortModel `tfsdk:"sort"`
	Density         types.String               `tfsdk:"density"`
	HeaderRowHeight types.String               `tfsdk:"header_row_height"`
	RowHeight       types.String               `tfsdk:"row_height"`
	DataSourceJSON  jsontypes.Normalized       `tfsdk:"data_source_json"`
}

type DiscoverSessionSortModel struct {
	Name      types.String `tfsdk:"name"`
	Direction types.String `tfsdk:"direction"`
}

type DiscoverSessionColumnSettingModel struct {
	Width types.Float64 `tfsdk:"width"`
}

type DiscoverSessionPanelByRefModel struct {
	TimeRange     *TimeRangeModel                `tfsdk:"time_range"`
	RefID         types.String                   `tfsdk:"ref_id"`
	SelectedTabID types.String                   `tfsdk:"selected_tab_id"`
	Overrides     *DiscoverSessionOverridesModel `tfsdk:"overrides"`
}

type DiscoverSessionOverridesModel struct {
	ColumnOrder     types.List                 `tfsdk:"column_order"`
	ColumnSettings  types.Map                  `tfsdk:"column_settings"`
	Sort            []DiscoverSessionSortModel `tfsdk:"sort"`
	Density         types.String               `tfsdk:"density"`
	HeaderRowHeight types.String               `tfsdk:"header_row_height"`
	RowHeight       types.String               `tfsdk:"row_height"`
	RowsPerPage     types.Int64                `tfsdk:"rows_per_page"`
	SampleSize      types.Int64                `tfsdk:"sample_size"`
}
