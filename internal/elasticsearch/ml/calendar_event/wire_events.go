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
	"strconv"
	"strings"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// calendarEventWire is the JSON shape for ML calendar events (POST body and GET list responses).
// The go-elasticsearch typed types.CalendarEvent struct omits skip_result, skip_model_update, and
// force_time_shift, so create uses postcalendarevents.Request only when those fields are absent;
// otherwise it posts this wire JSON via Raw. List responses are still decoded with this type.
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
	SkipResultsLegacy *bool `json:"skip_results,omitempty"`
	SkipModelUpdate   *bool `json:"skip_model_update,omitempty"`
	// ForceTimeShift is JSON number or string in API responses; RawMessage accepts both.
	ForceTimeShift json.RawMessage `json:"force_time_shift,omitempty"`
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

// forceTimeShiftWireToStringPtr decodes API JSON (number or string) into a decimal string for Terraform state.
// json.Unmarshal into any yields float64 for JSON numbers (not json.Number) unless a decoder with UseNumber is used.
func forceTimeShiftWireToStringPtr(raw json.RawMessage) (*string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var u any
	if err := json.Unmarshal(raw, &u); err != nil {
		return nil, err
	}
	switch v := u.(type) {
	case float64:
		if v == float64(int64(v)) {
			s := strconv.FormatInt(int64(v), 10)
			return &s, nil
		}
		s := strconv.FormatFloat(v, 'f', -1, 64)
		return &s, nil
	case string:
		return &v, nil
	default:
		return nil, fmt.Errorf("unsupported force_time_shift type %T", u)
	}
}

func forceTimeShiftStringToJSONRaw(s string) (json.RawMessage, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty force_time_shift")
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return json.Marshal(n)
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("force_time_shift must be numeric: %w", err)
	}
	return json.Marshal(f)
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
	fts, err := forceTimeShiftWireToStringPtr(w.ForceTimeShift)
	if err != nil {
		diags.AddError("Invalid force_time_shift in API response", err.Error())
		return nil, diags
	}
	m := &CalendarEventAPIModel{
		Description:     w.Description,
		StartTime:       startAny,
		EndTime:         endAny,
		SkipResult:      effectiveSkipResultPtr(w),
		SkipModelUpdate: w.SkipModelUpdate,
		ForceTimeShift:  fts,
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
	if typeutils.IsKnown(m.SkipResult) {
		v := m.SkipResult.ValueBool()
		w.SkipResult = &v
	}
	if typeutils.IsKnown(m.SkipModelUpdate) {
		v := m.SkipModelUpdate.ValueBool()
		w.SkipModelUpdate = &v
	}
	if typeutils.IsKnown(m.ForceTimeShift) {
		raw, err := forceTimeShiftStringToJSONRaw(m.ForceTimeShift.ValueString())
		if err != nil {
			diags.AddError("Invalid force_time_shift", err.Error())
			return calendarEventWire{}, diags
		}
		w.ForceTimeShift = raw
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

// optionalBoolPtrEqualLenientMLCalendarBool compares optional ML calendar booleans for POST/list identity.
// Elasticsearch may omit JSON false values; POST echoes may include server defaults (typically true) while
// Terraform leaves unconfigured attributes unset (nil in the plan wire).
func optionalBoolPtrEqualLenientMLCalendarBool(a, b *bool) bool {
	if optionalBoolPtrEqual(a, b) {
		return true
	}
	if a != nil && !*a && b == nil {
		return true
	}
	if b != nil && !*b && a == nil {
		return true
	}
	if a != nil && *a && b == nil {
		return true
	}
	if b != nil && *b && a == nil {
		return true
	}
	return false
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
	sm, ok1 := typeutils.ElasticDateTimeToMillis(startAny)
	em, ok2 := typeutils.ElasticDateTimeToMillis(endAny)
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
	if !optionalBoolPtrEqualLenientMLCalendarBool(effectiveSkipResultPtr(&ev), effectiveSkipResultPtr(&plan)) {
		return false
	}
	if !optionalBoolPtrEqualLenientMLCalendarBool(ev.SkipModelUpdate, plan.SkipModelUpdate) {
		return false
	}
	evFT, errEv := forceTimeShiftWireToStringPtr(ev.ForceTimeShift)
	if errEv != nil {
		return false
	}
	plFT, errPl := forceTimeShiftWireToStringPtr(plan.ForceTimeShift)
	if errPl != nil {
		return false
	}
	if !optionalStringPtrEqual(evFT, plFT) {
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

// postCalendarEventWireNeedsRawPOSTBody reports whether the POST body must be sent as raw JSON.
// go-elasticsearch types.CalendarEvent (and thus postcalendarevents.Request) omits skip_result,
// skip_model_update, and force_time_shift, so those fields require the wire JSON path.
func postCalendarEventWireNeedsRawPOSTBody(w calendarEventWire) bool {
	if w.SkipResult != nil || w.SkipModelUpdate != nil {
		return true
	}
	return len(w.ForceTimeShift) > 0
}

func calendarEventTypesFromWireForPOST(w calendarEventWire) (estypes.CalendarEvent, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	startMs, endMs, ok := calendarEventWireTimesMillis(&w)
	if !ok {
		diags.AddError("Invalid event times", "Could not read start_time and end_time as epoch millis for the ML calendar event request")
		return estypes.CalendarEvent{}, diags
	}
	return estypes.CalendarEvent{
		Description: w.Description,
		StartTime:   estypes.DateTime(startMs),
		EndTime:     estypes.DateTime(endMs),
	}, diags
}

func calendarEventWireFromTypesCalendarEvent(ev estypes.CalendarEvent) (calendarEventWire, error) {
	b, err := json.Marshal(ev)
	if err != nil {
		return calendarEventWire{}, err
	}
	var w calendarEventWire
	if err := json.Unmarshal(b, &w); err != nil {
		return calendarEventWire{}, err
	}
	return w, nil
}
