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

func TestCalendarEventMatchesPlanWire(t *testing.T) {
	startMs := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC).UnixMilli()
	endMs := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC).UnixMilli()
	tr := true
	fsRaw, err := json.Marshal(int64(86400))
	assert.NoError(t, err)

	plan := calendarEventWire{
		Description: "outage",
		StartTime:   millisJSONRaw(startMs),
		EndTime:     millisJSONRaw(endMs),
		SkipResult:  &tr,
	}
	ev := plan
	ev.ForceTimeShift = json.RawMessage(fsRaw)

	assert.False(t, calendarEventMatchesPlanWire(ev, plan))

	evMatch := plan
	assert.True(t, calendarEventMatchesPlanWire(evMatch, plan))

	ev2 := plan
	ev2.Description = "other"
	assert.False(t, calendarEventMatchesPlanWire(ev2, plan))
}
