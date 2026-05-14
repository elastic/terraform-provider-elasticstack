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

package panelkit

// TypedSiblingPanelConfigBlockNames lists panel-level typed config attribute names (`*_config`).
// Mirrors internal/kibana/dashboard/schema.go `panelConfigNames`; kept centralized for
// SiblingTypedPanelConfigConflictPathsExcept on migrated panel schemas (OpenSpec dashboard-panel-contract).
var TypedSiblingPanelConfigBlockNames = []string{
	"config_json",
	"markdown_config",
	"vis_config",
	"lens_dashboard_app_config",
	"esql_control_config",
	"options_list_control_config",
	"range_slider_control_config",
	"time_slider_control_config",
	"slo_alerts_config",
	"slo_burn_rate_config",
	"slo_overview_config",
	"slo_error_budget_config",
	"synthetics_monitors_config",
	"synthetics_stats_overview_config",
	"image_config",
	"discover_session_config",
}
