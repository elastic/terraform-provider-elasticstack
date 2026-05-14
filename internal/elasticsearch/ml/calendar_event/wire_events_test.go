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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarEventMatchesPlanWire_lenientFalseOmit(t *testing.T) {
	startMs := time.Date(2017, 12, 19, 0, 0, 0, 0, time.UTC).UnixMilli()
	endMs := time.Date(2017, 12, 20, 0, 0, 0, 0, time.UTC).UnixMilli()
	sf := false

	plan := calendarEventWire{
		Description: "event 1",
		StartTime:   millisJSONRaw(startMs),
		EndTime:     millisJSONRaw(endMs),
		SkipResult:  &sf,
	}
	// API echoed event without skip_result when false
	ev := calendarEventWire{
		Description: "event 1",
		StartTime:   millisJSONRaw(startMs),
		EndTime:     millisJSONRaw(endMs),
	}
	assert.True(t, calendarEventMatchesPlanWire(ev, plan))
}

func TestCalendarEventMatchesPlanWire_event2PostEcho(t *testing.T) {
	startMs := time.Date(2017, 12, 21, 0, 0, 0, 0, time.UTC).UnixMilli()
	endMs := time.Date(2017, 12, 22, 0, 0, 0, 0, time.UTC).UnixMilli()
	sf := false
	tr := true
	plan := calendarEventWire{
		Description:     "event 2",
		StartTime:       millisJSONRaw(startMs),
		EndTime:         millisJSONRaw(endMs),
		SkipModelUpdate: &sf,
	}
	ev := calendarEventWire{
		Description:     "event 2",
		StartTime:       millisJSONRaw(startMs),
		EndTime:         millisJSONRaw(endMs),
		SkipResult:      &tr,
		SkipModelUpdate: &sf,
	}
	assert.True(t, calendarEventMatchesPlanWire(ev, plan))
}

func TestCalendarEventMatchesPlanWire_event3PostEcho(t *testing.T) {
	startMs := time.Date(2017, 12, 25, 0, 0, 0, 0, time.UTC).UnixMilli()
	endMs := time.Date(2017, 12, 26, 0, 0, 0, 0, time.UTC).UnixMilli()
	tr := true
	ftRaw, err := json.Marshal(int64(3600))
	require.NoError(t, err)
	plan := calendarEventWire{
		Description:    "event 3",
		StartTime:      millisJSONRaw(startMs),
		EndTime:        millisJSONRaw(endMs),
		ForceTimeShift: json.RawMessage(ftRaw),
	}
	ev := calendarEventWire{
		Description:     "event 3",
		StartTime:       millisJSONRaw(startMs),
		EndTime:         millisJSONRaw(endMs),
		SkipResult:      &tr,
		SkipModelUpdate: &tr,
		ForceTimeShift:  json.RawMessage(ftRaw),
	}
	assert.True(t, calendarEventMatchesPlanWire(ev, plan))
}

func TestForceTimeShiftWireToStringPtr(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got, err := forceTimeShiftWireToStringPtr(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
	t.Run("json number", func(t *testing.T) {
		got, err := forceTimeShiftWireToStringPtr(json.RawMessage(`3600`))
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "3600", *got)
	})
	t.Run("json string", func(t *testing.T) {
		got, err := forceTimeShiftWireToStringPtr(json.RawMessage(`"7200"`))
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "7200", *got)
	})
}

func TestOptionalBoolPtrEqualLenientMLCalendarBool(t *testing.T) {
	f, tr := false, true
	assert.True(t, optionalBoolPtrEqualLenientMLCalendarBool(&f, nil))
	assert.True(t, optionalBoolPtrEqualLenientMLCalendarBool(nil, &f))
	assert.True(t, optionalBoolPtrEqualLenientMLCalendarBool(nil, nil))
	assert.True(t, optionalBoolPtrEqualLenientMLCalendarBool(&tr, &tr))
	assert.False(t, optionalBoolPtrEqualLenientMLCalendarBool(&tr, &f))
	assert.False(t, optionalBoolPtrEqualLenientMLCalendarBool(&f, &tr))
	// Terraform omitted; API echoed default true
	assert.True(t, optionalBoolPtrEqualLenientMLCalendarBool(nil, &tr))
	assert.True(t, optionalBoolPtrEqualLenientMLCalendarBool(&tr, nil))
}

func TestCalendarEventMatchesPlanWire_apiFillsDefaults(t *testing.T) {
	startMs := time.Date(2017, 12, 19, 0, 0, 0, 0, time.UTC).UnixMilli()
	endMs := time.Date(2017, 12, 20, 0, 0, 0, 0, time.UTC).UnixMilli()
	sf, tr := false, true
	plan := calendarEventWire{
		Description: "event 1",
		StartTime:   millisJSONRaw(startMs),
		EndTime:     millisJSONRaw(endMs),
		SkipResult:  &sf,
	}
	ev := calendarEventWire{
		Description:     "event 1",
		StartTime:       millisJSONRaw(startMs),
		EndTime:         millisJSONRaw(endMs),
		SkipResult:      &sf,
		SkipModelUpdate: &tr,
	}
	assert.True(t, calendarEventMatchesPlanWire(ev, plan))
}

func TestForceTimeShiftStringToJSONRaw(t *testing.T) {
	raw, err := forceTimeShiftStringToJSONRaw(" 3600 ")
	require.NoError(t, err)
	assert.Equal(t, "3600", string(raw))
}

func TestPostCalendarEventWireNeedsRawPOSTBody(t *testing.T) {
	startMs := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	endMs := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()
	base := calendarEventWire{
		Description: "x",
		StartTime:   millisJSONRaw(startMs),
		EndTime:     millisJSONRaw(endMs),
	}
	assert.False(t, postCalendarEventWireNeedsRawPOSTBody(base))

	withSkip := base
	tr := true
	withSkip.SkipResult = &tr
	assert.True(t, postCalendarEventWireNeedsRawPOSTBody(withSkip))
}
