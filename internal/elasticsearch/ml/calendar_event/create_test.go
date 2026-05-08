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
	"testing"
	"time"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/stretchr/testify/assert"
)

func TestCalendarEventDateTimeToUnixMilli(t *testing.T) {
	ms := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC).UnixMilli()

	assert.Equal(t, int64(1000), mustMillis(t, estypes.DateTime(int64(1000))))
	assert.Equal(t, int64(1000), mustMillis(t, estypes.DateTime(float64(1000))))
	assert.Equal(t, ms, mustMillis(t, estypes.DateTime("2026-06-01T12:00:00Z")))
}

func mustMillis(t *testing.T, dt estypes.DateTime) int64 {
	t.Helper()
	v, ok := calendarEventDateTimeToUnixMilli(dt)
	assert.True(t, ok)
	return v
}

func TestCalendarEventMatchesPlan(t *testing.T) {
	startMs := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC).UnixMilli()
	endMs := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC).UnixMilli()

	ev := estypes.CalendarEvent{
		Description: "outage",
		StartTime:   estypes.DateTime(float64(startMs)),
		EndTime:     estypes.DateTime(float64(endMs)),
	}

	assert.True(t, calendarEventMatchesPlan(ev, "outage", startMs, endMs))
	assert.False(t, calendarEventMatchesPlan(ev, "other", startMs, endMs))
	assert.False(t, calendarEventMatchesPlan(ev, "outage", startMs+1, endMs))
}
