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

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CalendarEventTFModel struct {
	entitycore.ResourceTimeoutsField
	ID                      types.String      `tfsdk:"id"`
	ElasticsearchConnection types.List        `tfsdk:"elasticsearch_connection"`
	CalendarID              types.String      `tfsdk:"calendar_id"`
	Description             types.String      `tfsdk:"description"`
	StartTime               timetypes.RFC3339 `tfsdk:"start_time"`
	EndTime                 timetypes.RFC3339 `tfsdk:"end_time"`
	EventID                 types.String      `tfsdk:"event_id"`
	SkipResult              types.Bool        `tfsdk:"skip_result"`
	SkipModelUpdate         types.Bool        `tfsdk:"skip_model_update"`
	ForceTimeShift          types.String      `tfsdk:"force_time_shift"`
}

// GetID implements entitycore.ElasticsearchResourceModel.
func (m CalendarEventTFModel) GetID() types.String { return m.ID }

// GetResourceID implements entitycore.ElasticsearchResourceModel.
// It returns "<calendar_id>/<event_id>" for the composite Elasticsearch resource ID segment
// (after the cluster UUID), matching delete/read and the envelope write path.
func (m CalendarEventTFModel) GetResourceID() types.String {
	if !typeutils.IsKnown(m.CalendarID) || !typeutils.IsKnown(m.EventID) {
		return types.StringUnknown()
	}
	if m.CalendarID.IsNull() || m.EventID.IsNull() {
		return types.StringNull()
	}
	c := m.CalendarID.ValueString()
	e := m.EventID.ValueString()
	if c == "" || e == "" {
		return types.StringUnknown()
	}
	return types.StringValue(c + "/" + e)
}

// GetElasticsearchConnection implements entitycore.ElasticsearchResourceModel.
func (m CalendarEventTFModel) GetElasticsearchConnection() types.List {
	return m.ElasticsearchConnection
}

type CalendarEventAPIModel struct {
	Description     string  `json:"description"`
	StartTime       any     `json:"start_time"`
	EndTime         any     `json:"end_time"`
	CalendarID      string  `json:"calendar_id,omitempty"`
	EventID         string  `json:"event_id,omitempty"`
	SkipResult      *bool   `json:"skip_result,omitempty"`
	SkipModelUpdate *bool   `json:"skip_model_update,omitempty"`
	ForceTimeShift  *string `json:"force_time_shift,omitempty"`
}

func (m *CalendarEventTFModel) fromAPIModel(_ context.Context, apiModel *CalendarEventAPIModel) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	m.Description = types.StringValue(apiModel.Description)
	m.CalendarID = types.StringValue(apiModel.CalendarID)
	m.EventID = types.StringValue(apiModel.EventID)

	startMillis, ok := typeutils.ElasticDateTimeToMillis(apiModel.StartTime)
	if !ok {
		diags.AddError(
			"Invalid start_time format",
			"Expected a supported time representation from the API (epoch millis, RFC3339 string, or typed DateTime)",
		)
		return diags
	}
	endMillis, ok := typeutils.ElasticDateTimeToMillis(apiModel.EndTime)
	if !ok {
		diags.AddError(
			"Invalid end_time format",
			"Expected a supported time representation from the API (epoch millis, RFC3339 string, or typed DateTime)",
		)
		return diags
	}

	startLoc := time.UTC
	if typeutils.IsKnown(m.StartTime) {
		if t, d := m.StartTime.ValueRFC3339Time(); !d.HasError() {
			startLoc = t.Location()
		}
	}
	endLoc := time.UTC
	if typeutils.IsKnown(m.EndTime) {
		if t, d := m.EndTime.ValueRFC3339Time(); !d.HasError() {
			endLoc = t.Location()
		}
	}

	startTime := time.UnixMilli(startMillis).In(startLoc)
	endTime := time.UnixMilli(endMillis).In(endLoc)

	m.StartTime = timetypes.NewRFC3339TimeValue(startTime)
	m.EndTime = timetypes.NewRFC3339TimeValue(endTime)

	m.SkipResult = typeutils.BoolPointerValue(apiModel.SkipResult)
	m.SkipModelUpdate = typeutils.BoolPointerValue(apiModel.SkipModelUpdate)
	m.ForceTimeShift = typeutils.StringishPointerValue(apiModel.ForceTimeShift)

	return diags
}
