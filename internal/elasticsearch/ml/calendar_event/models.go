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
	"encoding/json"
	"time"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
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

// GetID implements entitycore.ElasticsearchResourceModel.
func (m CalendarEventTFModel) GetID() types.String { return m.ID }

// GetResourceID implements entitycore.ElasticsearchResourceModel.
func (m CalendarEventTFModel) GetResourceID() types.String { return m.CalendarID }

// GetElasticsearchConnection implements entitycore.ElasticsearchResourceModel.
func (m CalendarEventTFModel) GetElasticsearchConnection() types.List {
	return m.ElasticsearchConnection
}

type CalendarEventAPIModel struct {
	Description string `json:"description"`
	StartTime   any    `json:"start_time"`
	EndTime     any    `json:"end_time"`
	CalendarID  string `json:"calendar_id,omitempty"`
	EventID     string `json:"event_id,omitempty"`
}

func calendarEventAnyTimeToUnixMilli(v any) (int64, bool) {
	if v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int64:
		return x, true
	case int:
		return int64(x), true
	case uint64:
		return int64(x), true
	case json.Number:
		i, err := x.Int64()
		if err == nil {
			return i, true
		}
		f, err := x.Float64()
		if err != nil {
			return 0, false
		}
		return int64(f), true
	case string:
		t, err := time.Parse(time.RFC3339, x)
		if err != nil {
			return 0, false
		}
		return t.UnixMilli(), true
	case estypes.DateTime:
		return calendarEventDateTimeToUnixMilli(x)
	default:
		return 0, false
	}
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

	startMillis, ok := calendarEventAnyTimeToUnixMilli(apiModel.StartTime)
	if !ok {
		diags.AddError(
			"Invalid start_time format",
			"Expected a supported time representation from the API (epoch millis, RFC3339 string, or typed DateTime)",
		)
		return diags
	}
	endMillis, ok := calendarEventAnyTimeToUnixMilli(apiModel.EndTime)
	if !ok {
		diags.AddError(
			"Invalid end_time format",
			"Expected a supported time representation from the API (epoch millis, RFC3339 string, or typed DateTime)",
		)
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

	startTime := time.UnixMilli(startMillis).In(startLoc)
	endTime := time.UnixMilli(endMillis).In(endLoc)

	m.StartTime = timetypes.NewRFC3339TimeValue(startTime)
	m.EndTime = timetypes.NewRFC3339TimeValue(endTime)

	return diags
}
