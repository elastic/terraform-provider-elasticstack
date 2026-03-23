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

package calendar_event

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CalendarEventTFModel struct {
	ID                      types.String      `tfsdk:"id"`
	ElasticsearchConnection types.List        `tfsdk:"elasticsearch_connection"`
	CalendarID              types.String      `tfsdk:"calendar_id"`
	Description             types.String      `tfsdk:"description"`
	StartTime               timetypes.RFC3339 `tfsdk:"start_time"`
	EndTime                 timetypes.RFC3339 `tfsdk:"end_time"`
	EventID                 types.String      `tfsdk:"event_id"`
}

type CalendarEventAPIModel struct {
	Description string `json:"description"`
	StartTime   any    `json:"start_time"`
	EndTime     any    `json:"end_time"`
	CalendarID  string `json:"calendar_id,omitempty"`
	EventID     string `json:"event_id,omitempty"`
}

type PostCalendarEventsRequest struct {
	Events []CalendarEventAPIModel `json:"events"`
}

type PostCalendarEventsResponse struct {
	Events []CalendarEventAPIModel `json:"events"`
}

type GetCalendarEventsResponse struct {
	Count  int                     `json:"count"`
	Events []CalendarEventAPIModel `json:"events"`
}

func (m *CalendarEventTFModel) toAPIModel(_ context.Context) (*CalendarEventAPIModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	startTime, d := m.StartTime.ValueRFC3339Time()
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	endTime, d := m.EndTime.ValueRFC3339Time()
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &CalendarEventAPIModel{
		Description: m.Description.ValueString(),
		StartTime:   startTime.UnixMilli(),
		EndTime:     endTime.UnixMilli(),
	}, diags
}

func (m *CalendarEventTFModel) fromAPIModel(_ context.Context, apiModel *CalendarEventAPIModel) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	m.Description = types.StringValue(apiModel.Description)
	m.CalendarID = types.StringValue(apiModel.CalendarID)
	m.EventID = types.StringValue(apiModel.EventID)

	// The API returns epoch milliseconds as float64 (JSON number decoded via any).
	startMillis, ok := apiModel.StartTime.(float64)
	if !ok {
		diags.AddError("Invalid start_time format", "Expected epoch milliseconds as a number from the API")
		return diags
	}
	endMillis, ok := apiModel.EndTime.(float64)
	if !ok {
		diags.AddError("Invalid end_time format", "Expected epoch milliseconds as a number from the API")
		return diags
	}

	startLoc := time.UTC
	if t, d := m.StartTime.ValueRFC3339Time(); !d.HasError() && !m.StartTime.IsNull() && !m.StartTime.IsUnknown() {
		startLoc = t.Location()
	}
	endLoc := time.UTC
	if t, d := m.EndTime.ValueRFC3339Time(); !d.HasError() && !m.EndTime.IsNull() && !m.EndTime.IsUnknown() {
		endLoc = t.Location()
	}

	startTime := time.UnixMilli(int64(startMillis)).In(startLoc)
	endTime := time.UnixMilli(int64(endMillis)).In(endLoc)

	m.StartTime = timetypes.NewRFC3339TimeValue(startTime)
	m.EndTime = timetypes.NewRFC3339TimeValue(endTime)

	return diags
}
