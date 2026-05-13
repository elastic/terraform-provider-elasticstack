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

type AccessControlValue struct {
	AccessMode types.String `tfsdk:"access_mode"`
}

type TimeRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
	Mode types.String `tfsdk:"mode"`
}

type RefreshIntervalModel struct {
	Pause types.Bool  `tfsdk:"pause"`
	Value types.Int64 `tfsdk:"value"`
}

type DashboardQueryModel struct {
	Language types.String         `tfsdk:"language"`
	Text     types.String         `tfsdk:"text"`
	JSON     jsontypes.Normalized `tfsdk:"json"`
}

type OptionsModel struct {
	HidePanelTitles  types.Bool `tfsdk:"hide_panel_titles"`
	UseMargins       types.Bool `tfsdk:"use_margins"`
	SyncColors       types.Bool `tfsdk:"sync_colors"`
	SyncTooltips     types.Bool `tfsdk:"sync_tooltips"`
	SyncCursor       types.Bool `tfsdk:"sync_cursor"`
	AutoApplyFilters types.Bool `tfsdk:"auto_apply_filters"`
	HidePanelBorders types.Bool `tfsdk:"hide_panel_borders"`
}

type DashboardModel struct {
	ID               types.String          `tfsdk:"id"`
	KibanaConnection types.List            `tfsdk:"kibana_connection"`
	SpaceID          types.String          `tfsdk:"space_id"`
	DashboardID      types.String          `tfsdk:"dashboard_id"`
	Title            types.String          `tfsdk:"title"`
	Description      types.String          `tfsdk:"description"`
	TimeRange        *TimeRangeModel       `tfsdk:"time_range"`
	RefreshInterval  *RefreshIntervalModel `tfsdk:"refresh_interval"`
	Query            *DashboardQueryModel  `tfsdk:"query"`
	Filters          types.List            `tfsdk:"filters"`
	Tags             types.List            `tfsdk:"tags"`
	Options          *OptionsModel         `tfsdk:"options"`
	AccessControl    *AccessControlValue   `tfsdk:"access_control"`
	Panels           []PanelModel          `tfsdk:"panels"`
	PinnedPanels     []PinnedPanelModel    `tfsdk:"pinned_panels"`
	Sections         []SectionModel        `tfsdk:"sections"`
}
