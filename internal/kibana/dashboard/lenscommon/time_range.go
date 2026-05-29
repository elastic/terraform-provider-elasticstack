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
func TimeRangeModelToAPI(tr *models.TimeRangeModel) *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema {
	if tr == nil {
		return nil
	}
	out := &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
		From: tr.From.ValueString(),
		To:   tr.To.ValueString(),
	}
	if typeutils.IsKnown(tr.Mode) {
		mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaMode(tr.Mode.ValueString())
		out.Mode = &mode
	}
	return out
}

// ResolveChartTimeRange returns the API time_range for a typed Lens chart root when chart-level is set;
// nil when chart-level time_range is unset (caller omits from API payload).
func ResolveChartTimeRange(_ *models.DashboardModel, chartLevel *models.TimeRangeModel) *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema {
	return TimeRangeModelToAPI(chartLevel)
}
