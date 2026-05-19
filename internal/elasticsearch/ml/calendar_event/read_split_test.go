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

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitCalendarEventResourcePath(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, _, diags := splitCalendarEventResourcePath("")
		require.True(t, diags.HasError())
	})

	t.Run("missing slash", func(t *testing.T) {
		_, _, diags := splitCalendarEventResourcePath("only-calendar")
		require.True(t, diags.HasError())
	})

	t.Run("empty calendar_id", func(t *testing.T) {
		_, _, diags := splitCalendarEventResourcePath("/event-1")
		require.True(t, diags.HasError())
	})

	t.Run("empty event_id", func(t *testing.T) {
		_, _, diags := splitCalendarEventResourcePath("cal-1/")
		require.True(t, diags.HasError())
	})

	t.Run("valid", func(t *testing.T) {
		cal, evt, diags := splitCalendarEventResourcePath("my-cal/evt-1")
		require.False(t, diags.HasError())
		assert.Equal(t, "my-cal", cal)
		assert.Equal(t, "evt-1", evt)
	})

	t.Run("event_id with slashes", func(t *testing.T) {
		cal, evt, diags := splitCalendarEventResourcePath("my-cal/evt/with/slashes")
		require.False(t, diags.HasError())
		assert.Equal(t, "my-cal", cal)
		assert.Equal(t, "evt/with/slashes", evt)
	})
}

func TestParseCalendarEventFullCompositeID(t *testing.T) {
	t.Run("missing outer slash", func(t *testing.T) {
		_, _, diags := parseCalendarEventFullCompositeID("cluster-uuid-only")
		require.True(t, diags.HasError())
	})

	t.Run("empty resource segment after cluster", func(t *testing.T) {
		_, _, diags := parseCalendarEventFullCompositeID("cluster-uuid/")
		require.True(t, diags.HasError())
	})

	t.Run("empty resource segment", func(t *testing.T) {
		_, _, diags := parseCalendarEventFullCompositeID("cluster/")
		require.True(t, diags.HasError())
	})

	t.Run("valid nested resource id", func(t *testing.T) {
		cal, evt, diags := parseCalendarEventFullCompositeID("cluster-uuid/my-cal/evt/with/slashes")
		require.False(t, diags.HasError())
		assert.Equal(t, "my-cal", cal)
		assert.Equal(t, "evt/with/slashes", evt)
	})
}

func TestCalendarEventReadWindowRFC3339(t *testing.T) {
	t.Run("unknown times", func(t *testing.T) {
		var m CalendarEventTFModel
		m.StartTime = timetypes.NewRFC3339Unknown()
		m.EndTime = timetypes.NewRFC3339ValueMust(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).Format(time.RFC3339))
		_, _, ok := calendarEventReadWindowRFC3339(m)
		assert.False(t, ok)
	})

	t.Run("known window", func(t *testing.T) {
		start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.FixedZone("ACST", 9*3600))
		end := time.Date(2026, 6, 1, 6, 0, 0, 0, time.FixedZone("ACST", 9*3600))
		m := CalendarEventTFModel{
			StartTime: timetypes.NewRFC3339TimeValue(start),
			EndTime:   timetypes.NewRFC3339TimeValue(end),
		}
		ws, we, ok := calendarEventReadWindowRFC3339(m)
		require.True(t, ok)
		assert.Equal(t, start.UTC().Format(time.RFC3339), ws)
		assert.Equal(t, end.UTC().Format(time.RFC3339), we)
	})
}
