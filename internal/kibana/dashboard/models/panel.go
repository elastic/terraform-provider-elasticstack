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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PanelModel struct {
	Type                          types.String                                      `tfsdk:"type"`
	Grid                          PanelGridModel                                    `tfsdk:"grid"`
	ID                            types.String                                      `tfsdk:"id"`
	MarkdownConfig                *MarkdownConfigModel                              `tfsdk:"markdown_config"`
	TimeSliderControlConfig       *TimeSliderControlConfigModel                     `tfsdk:"time_slider_control_config"`
	SloBurnRateConfig             *SloBurnRateConfigModel                           `tfsdk:"slo_burn_rate_config"`
	SloOverviewConfig             *SloOverviewConfigModel                           `tfsdk:"slo_overview_config"`
	SloErrorBudgetConfig          *SloErrorBudgetConfigModel                        `tfsdk:"slo_error_budget_config"`
	EsqlControlConfig             *EsqlControlConfigModel                           `tfsdk:"esql_control_config"`
	OptionsListControlConfig      *OptionsListControlConfigModel                    `tfsdk:"options_list_control_config"`
	RangeSliderControlConfig      *RangeSliderControlConfigModel                    `tfsdk:"range_slider_control_config"`
	SyntheticsStatsOverviewConfig *SyntheticsStatsOverviewConfigModel               `tfsdk:"synthetics_stats_overview_config"`
	SyntheticsMonitorsConfig      *SyntheticsMonitorsConfigModel                    `tfsdk:"synthetics_monitors_config"`
	LensDashboardAppConfig        *LensDashboardAppConfigModel                      `tfsdk:"lens_dashboard_app_config"`
	VisConfig                     *VisConfigModel                                   `tfsdk:"vis_config"`
	ImageConfig                   *ImagePanelConfigModel                            `tfsdk:"image_config"`
	SloAlertsConfig               *SloAlertsPanelConfigModel                        `tfsdk:"slo_alerts_config"`
	DiscoverSessionConfig         *DiscoverSessionPanelConfigModel                  `tfsdk:"discover_session_config"`
	ConfigJSON                    customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config_json"`
}

type PinnedPanelModel struct {
	Type                     types.String                   `tfsdk:"type"`
	TimeSliderControlConfig  *TimeSliderControlConfigModel  `tfsdk:"time_slider_control_config"`
	EsqlControlConfig        *EsqlControlConfigModel        `tfsdk:"esql_control_config"`
	OptionsListControlConfig *OptionsListControlConfigModel `tfsdk:"options_list_control_config"`
	RangeSliderControlConfig *RangeSliderControlConfigModel `tfsdk:"range_slider_control_config"`
}

type PanelGridModel struct {
	X types.Int64 `tfsdk:"x"`
	Y types.Int64 `tfsdk:"y"`
	W types.Int64 `tfsdk:"w"`
	H types.Int64 `tfsdk:"h"`
}

type SectionModel struct {
	Title     types.String     `tfsdk:"title"`
	ID        types.String     `tfsdk:"id"`
	Collapsed types.Bool       `tfsdk:"collapsed"`
	Grid      SectionGridModel `tfsdk:"grid"`
	Panels    []PanelModel     `tfsdk:"panels"`
}

type SectionGridModel struct {
	Y types.Int64 `tfsdk:"y"`
}
