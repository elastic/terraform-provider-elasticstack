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
	"fmt"
	"testing"
	"time"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarEventAnyTimeToUnixMilli(t *testing.T) {
	ms := time.Date(2026, 7, 1, 15, 30, 0, 0, time.FixedZone("EDT", -4*3600)).UnixMilli()

	t.Run("json.Number int", func(t *testing.T) {
		n := json.Number(fmt.Sprintf("%d", ms))
		got, ok := calendarEventAnyTimeToUnixMilli(n)
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("RFC3339 string with offset", func(t *testing.T) {
		got, ok := calendarEventAnyTimeToUnixMilli("2026-07-01T15:30:00-04:00")
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("DateTime typed", func(t *testing.T) {
		got, ok := calendarEventAnyTimeToUnixMilli(estypes.DateTime(float64(ms)))
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("unsupported bool", func(t *testing.T) {
		_, ok := calendarEventAnyTimeToUnixMilli(true)
		assert.False(t, ok)
	})

	t.Run("nil", func(t *testing.T) {
		_, ok := calendarEventAnyTimeToUnixMilli(nil)
		assert.False(t, ok)
	})
}

func TestCalendarEventTFModel_fromAPIModel_viaAnyTypes(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	startMs := base.UnixMilli()
	endMs := base.Add(time.Hour).UnixMilli()

	t.Run("json.Number millis", func(t *testing.T) {
		m := &CalendarEventTFModel{
			StartTime: timetypes.NewRFC3339TimeValue(base),
			EndTime:   timetypes.NewRFC3339TimeValue(base.Add(time.Hour)),
		}
		diags := m.fromAPIModel(ctx, &CalendarEventAPIModel{
			Description: "x",
			StartTime:   json.Number(fmt.Sprintf("%d", startMs)),
			EndTime:     json.Number(fmt.Sprintf("%d", endMs)),
			CalendarID:  "c",
			EventID:     "e",
		})
		require.False(t, diags.HasError())
		st, d := m.StartTime.ValueRFC3339Time()
		require.False(t, d.HasError())
		assert.Equal(t, startMs, st.UnixMilli())
	})

	t.Run("unsupported map for start_time", func(t *testing.T) {
		m := &CalendarEventTFModel{
			StartTime: timetypes.NewRFC3339Unknown(),
			EndTime:   timetypes.NewRFC3339Unknown(),
		}
		diags := m.fromAPIModel(ctx, &CalendarEventAPIModel{
			Description: "x",
			StartTime:   map[string]any{"k": "v"},
			EndTime:     float64(endMs),
			CalendarID:  "c",
			EventID:     "e",
		})
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Summary(), "Invalid start_time format")
	})
}
