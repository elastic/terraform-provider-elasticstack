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

type ApmServiceMapConfigModel struct {
	Title                    types.String    `tfsdk:"title"`
	Description              types.String    `tfsdk:"description"`
	HideTitle                types.Bool      `tfsdk:"hide_title"`
	HideBorder               types.Bool      `tfsdk:"hide_border"`
	Environment              types.String    `tfsdk:"environment"`
	ServiceName              types.String    `tfsdk:"service_name"`
	ServiceGroupID           types.String    `tfsdk:"service_group_id"`
	Kuery                    types.String    `tfsdk:"kuery"`
	MapOrientation           types.String    `tfsdk:"map_orientation"`
	SyncWithDashboardFilters types.Bool      `tfsdk:"sync_with_dashboard_filters"`
	AlertStatusFilter        types.Set       `tfsdk:"alert_status_filter"`
	AnomalySeverityFilter    types.Set       `tfsdk:"anomaly_severity_filter"`
	ConnectionFilter         types.Set       `tfsdk:"connection_filter"`
	SloStatusFilter          types.Set       `tfsdk:"slo_status_filter"`
	TimeRange                *TimeRangeModel `tfsdk:"time_range"`
}
