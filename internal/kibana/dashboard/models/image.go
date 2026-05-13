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

type ImagePanelConfigModel struct {
	Src             ImagePanelSrcModel         `tfsdk:"src"`
	AltText         types.String               `tfsdk:"alt_text"`
	ObjectFit       types.String               `tfsdk:"object_fit"`
	BackgroundColor types.String               `tfsdk:"background_color"`
	Title           types.String               `tfsdk:"title"`
	Description     types.String               `tfsdk:"description"`
	HideTitle       types.Bool                 `tfsdk:"hide_title"`
	HideBorder      types.Bool                 `tfsdk:"hide_border"`
	Drilldowns      []ImagePanelDrilldownModel `tfsdk:"drilldowns"`
}

type ImagePanelSrcModel struct {
	File *ImagePanelSrcFileModel `tfsdk:"file"`
	URL  *ImagePanelSrcURLModel  `tfsdk:"url"`
}

type ImagePanelSrcFileModel struct {
	FileID types.String `tfsdk:"file_id"`
}

type ImagePanelSrcURLModel struct {
	URL types.String `tfsdk:"url"`
}

type ImagePanelDrilldownModel struct {
	DashboardDrilldown *ImagePanelDashboardDrilldownModel `tfsdk:"dashboard_drilldown"`
	URLDrilldown       *ImagePanelURLDrilldownModel       `tfsdk:"url_drilldown"`
}

type ImagePanelDashboardDrilldownModel struct {
	DashboardID  types.String `tfsdk:"dashboard_id"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	UseFilters   types.Bool   `tfsdk:"use_filters"`
	UseTimeRange types.Bool   `tfsdk:"use_time_range"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type ImagePanelURLDrilldownModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}
