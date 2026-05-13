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
	"encoding/json"
	"fmt"

	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// calendarEventWire is the JSON shape for ML calendar events (POST body and GET list responses).
// The go-elasticsearch typed CalendarEvent struct does not yet include skip_result, skip_model_update,
// or force_time_shift, so we decode list/post payloads with this type.
//
// https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-ml-post-calendar-events
type calendarEventWire struct {
	Description string          `json:"description"`
	StartTime   json.RawMessage `json:"start_time"`
	EndTime     json.RawMessage `json:"end_time"`
	EventID     *string         `json:"event_id,omitempty"`
	CalendarID  *string         `json:"calendar_id,omitempty"`
	SkipResult  *bool           `json:"skip_result,omitempty"`
	// SkipResultsLegacy is accepted on decode only; Elasticsearch may still return `skip_results` in some versions.
	SkipResultsLegacy *bool   `json:"skip_results,omitempty"`
	SkipModelUpdate   *bool   `json:"skip_model_update,omitempty"`
	ForceTimeShift    *string `json:"force_time_shift,omitempty"`
}

func rawJSONToAny(raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty JSON value")
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func effectiveSkipResultPtr(w *calendarEventWire) *bool {
	if w == nil {
		return nil
	}
	if w.SkipResult != nil {
		return w.SkipResult
	}
	return w.SkipResultsLegacy
}

func wireEventToAPIModel(w *calendarEventWire) (*CalendarEventAPIModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	startAny, err := rawJSONToAny(w.StartTime)
	if err != nil {
		diags.AddError("Invalid start_time in API response", err.Error())
		return nil, diags
	}
	endAny, err := rawJSONToAny(w.EndTime)
	if err != nil {
		diags.AddError("Invalid end_time in API response", err.Error())
		return nil, diags
	}
	m := &CalendarEventAPIModel{
		Description:     w.Description,
		StartTime:       startAny,
		EndTime:         endAny,
		SkipResult:      effectiveSkipResultPtr(w),
		SkipModelUpdate: w.SkipModelUpdate,
		ForceTimeShift:  w.ForceTimeShift,
	}
	if w.CalendarID != nil {
		m.CalendarID = *w.CalendarID
	}
	if w.EventID != nil {
		m.EventID = *w.EventID
	}
	return m, diags
}

func millisJSONRaw(ms int64) json.RawMessage {
	b, err := json.Marshal(ms)
	if err != nil {
		return json.RawMessage("0")
	}
	return json.RawMessage(b)
}

func calendarEventWireFromTFModel(m *CalendarEventTFModel) (calendarEventWire, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	startTime, d := m.StartTime.ValueRFC3339Time()
	diags.Append(d...)
	if diags.HasError() {
		return calendarEventWire{}, diags
	}
	endTime, d := m.EndTime.ValueRFC3339Time()
	diags.Append(d...)
	if diags.HasError() {
		return calendarEventWire{}, diags
	}
	w := calendarEventWire{
		Description: m.Description.ValueString(),
		StartTime:   millisJSONRaw(startTime.UnixMilli()),
		EndTime:     millisJSONRaw(endTime.UnixMilli()),
	}
	if !m.SkipResult.IsNull() && !m.SkipResult.IsUnknown() {
		v := m.SkipResult.ValueBool()
		w.SkipResult = &v
	}
	if !m.SkipModelUpdate.IsNull() && !m.SkipModelUpdate.IsUnknown() {
		v := m.SkipModelUpdate.ValueBool()
		w.SkipModelUpdate = &v
	}
	if !m.ForceTimeShift.IsNull() && !m.ForceTimeShift.IsUnknown() {
		s := m.ForceTimeShift.ValueString()
		w.ForceTimeShift = &s
	}
	return w, diags
}

func optionalBoolPtrEqual(a, b *bool) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}

func optionalStringPtrEqual(a, b *string) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}

func calendarEventWireTimesMillis(w *calendarEventWire) (startMs, endMs int64, ok bool) {
	startAny, err := rawJSONToAny(w.StartTime)
	if err != nil {
		return 0, 0, false
	}
	endAny, err := rawJSONToAny(w.EndTime)
	if err != nil {
		return 0, 0, false
	}
	sm, ok1 := calendarEventAnyTimeToUnixMilli(startAny)
	em, ok2 := calendarEventAnyTimeToUnixMilli(endAny)
	return sm, em, ok1 && ok2
}

// calendarEventMatchesPlanWire returns true when ev matches the planned event (description, times, optional flags).
func calendarEventMatchesPlanWire(ev, plan calendarEventWire) bool {
	if ev.Description != plan.Description {
		return false
	}
	evSm, evEm, ok1 := calendarEventWireTimesMillis(&ev)
	plSm, plEm, ok2 := calendarEventWireTimesMillis(&plan)
	if !ok1 || !ok2 || evSm != plSm || evEm != plEm {
		return false
	}
	if !optionalBoolPtrEqual(effectiveSkipResultPtr(&ev), effectiveSkipResultPtr(&plan)) {
		return false
	}
	if !optionalBoolPtrEqual(ev.SkipModelUpdate, plan.SkipModelUpdate) {
		return false
	}
	if !optionalStringPtrEqual(ev.ForceTimeShift, plan.ForceTimeShift) {
		return false
	}
	return true
}

func calendarEventWireEventID(w *calendarEventWire) string {
	if w.EventID != nil {
		return *w.EventID
	}
	return ""
}
