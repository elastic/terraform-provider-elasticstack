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

package datafeedstate

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustUnmarshalDatafeedStats(t *testing.T, raw string) *estypes.DatafeedStats {
	t.Helper()
	var stats estypes.DatafeedStats
	require.NoError(t, json.Unmarshal([]byte(raw), &stats))
	return &stats
}

func TestSetEffectiveSearchIntervalFromAPI_startedWithSearchInterval(t *testing.T) {
	stats := mustUnmarshalDatafeedStats(t, `{
		"datafeed_id": "df-1",
		"state": "started",
		"running_state": {
			"real_time_configured": false,
			"real_time_running": false,
			"search_interval": {
				"start_ms": 1640995200000,
				"end_ms": 1640998800000
			}
		}
	}`)

	userStart := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	userEnd := time.Date(2022, 1, 1, 3, 0, 0, 0, time.UTC)
	data := MLDatafeedStateData{
		Start: timetypes.NewRFC3339TimeValue(userStart),
		End:   timetypes.NewRFC3339TimeValue(userEnd),
	}

	diags := data.SetEffectiveSearchIntervalFromAPI(stats)
	require.False(t, diags.HasError())

	startTime, startDiags := data.Start.ValueRFC3339Time()
	require.False(t, startDiags.HasError())
	assert.Equal(t, userStart.UTC(), startTime.UTC())

	endTime, endDiags := data.End.ValueRFC3339Time()
	require.False(t, endDiags.HasError())
	assert.Equal(t, userEnd.UTC(), endTime.UTC())

	effectiveStart, effStartDiags := data.EffectiveSearchStart.ValueRFC3339Time()
	require.False(t, effStartDiags.HasError())
	assert.Equal(t, time.UnixMilli(1640995200000).UTC(), effectiveStart.UTC())

	effectiveEnd, effEndDiags := data.EffectiveSearchEnd.ValueRFC3339Time()
	require.False(t, effEndDiags.HasError())
	assert.Equal(t, time.UnixMilli(1640998800000).UTC(), effectiveEnd.UTC())
}

func TestSetEffectiveSearchIntervalFromAPI_preservesConfiguredTimezoneOnEffectiveFields(t *testing.T) {
	stats := mustUnmarshalDatafeedStats(t, `{
		"datafeed_id": "df-1",
		"state": "started",
		"running_state": {
			"real_time_configured": false,
			"real_time_running": false,
			"search_interval": {
				"start_ms": 1640995200000,
				"end_ms": 1640998800000
			}
		}
	}`)

	cet := time.FixedZone("CET", 3600)
	data := MLDatafeedStateData{
		Start: timetypes.NewRFC3339TimeValue(time.Date(2022, 1, 1, 0, 0, 0, 0, cet)),
	}

	diags := data.SetEffectiveSearchIntervalFromAPI(stats)
	require.False(t, diags.HasError())

	effectiveStart, effStartDiags := data.EffectiveSearchStart.ValueRFC3339Time()
	require.False(t, effStartDiags.HasError())
	_, offset := effectiveStart.Zone()
	assert.Equal(t, 3600, offset)
}

func TestSetEffectiveSearchIntervalFromAPI_startedRealTimeConfigured(t *testing.T) {
	stats := mustUnmarshalDatafeedStats(t, `{
		"datafeed_id": "df-1",
		"state": "started",
		"running_state": {
			"real_time_configured": true,
			"real_time_running": true,
			"search_interval": {
				"start_ms": 1640995200000,
				"end_ms": 1640998800000
			}
		}
	}`)

	data := MLDatafeedStateData{}
	diags := data.SetEffectiveSearchIntervalFromAPI(stats)
	require.False(t, diags.HasError())

	assert.False(t, data.EffectiveSearchStart.IsNull())
	assert.True(t, data.EffectiveSearchEnd.IsNull())
}

func TestSetEffectiveSearchIntervalFromAPI_stopped(t *testing.T) {
	stats := mustUnmarshalDatafeedStats(t, `{
		"datafeed_id": "df-1",
		"state": "stopped"
	}`)

	data := MLDatafeedStateData{
		Start: timetypes.NewRFC3339TimeValue(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		End:   timetypes.NewRFC3339TimeValue(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
	}

	diags := data.SetEffectiveSearchIntervalFromAPI(stats)
	require.False(t, diags.HasError())

	assert.True(t, data.EffectiveSearchStart.IsNull())
	assert.True(t, data.EffectiveSearchEnd.IsNull())

	startTime, startDiags := data.Start.ValueRFC3339Time()
	require.False(t, startDiags.HasError())
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), startTime)
}

func TestSetEffectiveSearchIntervalFromAPI_startedNilRunningState(t *testing.T) {
	stats := mustUnmarshalDatafeedStats(t, `{
		"datafeed_id": "df-1",
		"state": "started"
	}`)

	data := MLDatafeedStateData{}
	diags := data.SetEffectiveSearchIntervalFromAPI(stats)
	require.NotEmpty(t, diags.Warnings())
	assert.True(t, data.EffectiveSearchStart.IsNull())
	assert.True(t, data.EffectiveSearchEnd.IsNull())
}

func TestSetEffectiveSearchIntervalFromAPI_startedNilSearchInterval(t *testing.T) {
	stats := mustUnmarshalDatafeedStats(t, `{
		"datafeed_id": "df-1",
		"state": "started",
		"running_state": {
			"real_time_configured": false,
			"real_time_running": false
		}
	}`)

	data := MLDatafeedStateData{}
	diags := data.SetEffectiveSearchIntervalFromAPI(stats)
	require.False(t, diags.HasError())
	assert.True(t, data.EffectiveSearchStart.IsNull())
	assert.True(t, data.EffectiveSearchEnd.IsNull())
}

func TestGetSchema_effectiveSearchAttributes(t *testing.T) {
	s := GetSchema(context.Background())

	for _, name := range []string{"effective_search_start", "effective_search_end"} {
		attr, ok := s.Attributes[name]
		require.True(t, ok, "attribute %q should exist", name)

		stringAttr, ok := attr.(schema.StringAttribute)
		require.True(t, ok, "attribute %q should be StringAttribute", name)
		assert.True(t, stringAttr.Computed)
		assert.False(t, stringAttr.Optional)
		assert.IsType(t, timetypes.RFC3339Type{}, stringAttr.CustomType)
	}

	startAttr, ok := s.Attributes["start"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, startAttr.Optional)
	assert.False(t, startAttr.Computed)
	assert.Empty(t, startAttr.PlanModifiers)
}
