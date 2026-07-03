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

// Structural panel attribute names used by the dashboard schema, panel
// marshalling helpers, and the panel-config mutual-exclusion validator.
// `attrPanelType`, `attrPanelGrid`, and `attrPanelID` correspond to the
// Kibana dashboard panel envelope fields (`type`, `grid`, `id`) and are not
// considered "panel config" branches by panelConfigNames().
const (
	attrPanelType = "type"
	attrPanelGrid = "grid"
	attrPanelID   = "id"

	attrPanels       = "panels"
	attrPinnedPanels = "pinned_panels"
	attrSections     = "sections"

	attrDashboardID = "dashboard_id"
)

// Pinned panel control block names (Kibana dashboard control bar). These
// strings appear both in the pinned-panel mutual exclusion list and as
// SingleNestedAttribute keys / panelkit PanelConfigDescription block names.
const (
	controlBlockTimeSlider  = "time_slider_control_config"
	controlBlockESQL        = "esql_control_config"
	controlBlockOptionsList = "options_list_control_config"
	controlBlockRangeSlider = "range_slider_control_config"
)

// Common attribute keys used inside one or more control-config blocks.
// Centralising them avoids repeating the same string literal across schema
// definitions for the ES|QL, options-list, range-slider, and time-slider
// controls.
const (
	attrEndPercentageOfTimeRange = "end_percentage_of_time_range"
	attrIsAnchored               = "is_anchored"
	attrESQLQuery                = "esql_query"
	attrControlType              = "control_type"
	attrAvailableOptions         = "available_options"
	attrDisplaySettings          = "display_settings"
	attrPlaceholder              = "placeholder"
	attrHideActionBar            = "hide_action_bar"
	attrHideExclude              = "hide_exclude"
	attrHideExists               = "hide_exists"
	attrHideSort                 = "hide_sort"

	attrTitle                      = "title"
	attrValue                      = "value"
	attrStartPercentageOfTimeRange = "start_percentage_of_time_range"
	attrSelectedOptions            = "selected_options"
	attrVariableName               = "variable_name"
	attrVariableType               = "variable_type"
	attrSingleSelect               = "single_select"
)
