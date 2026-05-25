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

package lenscommon

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
)

// TimeRangeModelToAPI converts a TimeRangeModel to the Kibana API time range schema.
func TimeRangeModelToAPI(tr *models.TimeRangeModel) kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema {
	if tr == nil {
		return kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{}
	}
	out := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
		From: tr.From.ValueString(),
		To:   tr.To.ValueString(),
	}
	if typeutils.IsKnown(tr.Mode) {
		mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaMode(tr.Mode.ValueString())
		out.Mode = &mode
	}
	return out
}

// ResolveChartTimeRange returns the API time_range for a typed Lens chart root: chart-level when set,
// otherwise copied from the dashboard-level time_range (both are required API inputs).
//
// Production dashboard writes always pass the enclosing models.DashboardModel, so null chart-level
// time_range inherits dashboard-level values (REQ-013).
//
// The `now-15m` / `now` fallback applies when there is no chart-level override and either no parent
// models.DashboardModel is in scope, or dashboard != nil but dashboard.TimeRange == nil.
func ResolveChartTimeRange(dashboard *models.DashboardModel, chartLevel *models.TimeRangeModel) kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema {
	if chartLevel != nil {
		return TimeRangeModelToAPI(chartLevel)
	}
	if dashboard != nil && dashboard.TimeRange != nil {
		return TimeRangeModelToAPI(dashboard.TimeRange)
	}
	return kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
		From: "now-15m",
		To:   "now",
	}
}

// DashboardLensComparableTimeRange returns the dashboard-level time range used when comparing
// chart-root API time_range for Terraform null-preservation. ok is false when no comparable range exists.
func DashboardLensComparableTimeRange(dashboard *models.DashboardModel) (kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema, bool) {
	if dashboard == nil || dashboard.TimeRange == nil {
		return kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{}, false
	}
	return TimeRangeModelToAPI(dashboard.TimeRange), true
}
