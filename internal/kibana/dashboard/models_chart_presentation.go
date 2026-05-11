// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

// lensChartPresentationTFModel mirrors optional chart-root presentation fields on typed Lens configs.
// Populate on read and serialize on write in expose-lens-chart-presentation-fields tasks 2–3.
type lensChartPresentationTFModel struct {
	TimeRange      *timeRangeModel            `tfsdk:"time_range"`
	HideTitle      types.Bool                 `tfsdk:"hide_title"`
	HideBorder     types.Bool                 `tfsdk:"hide_border"`
	ReferencesJSON jsontypes.Normalized       `tfsdk:"references_json"`
	Drilldowns     []lensDrilldownItemTFModel `tfsdk:"drilldowns"`
}

// lensDrilldownItemTFModel is one drilldown entry; exactly one nested variant is set after validation.
type lensDrilldownItemTFModel struct {
	DashboardDrilldown *lensDashboardDrilldownTFModel `tfsdk:"dashboard_drilldown"`
	DiscoverDrilldown  *lensDiscoverDrilldownTFModel  `tfsdk:"discover_drilldown"`
	URLDrilldown       *lensURLDrilldownTFModel       `tfsdk:"url_drilldown"`
}

type lensDashboardDrilldownTFModel struct {
	DashboardID  types.String `tfsdk:"dashboard_id"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	UseFilters   types.Bool   `tfsdk:"use_filters"`
	UseTimeRange types.Bool   `tfsdk:"use_time_range"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type lensDiscoverDrilldownTFModel struct {
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type lensURLDrilldownTFModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}
