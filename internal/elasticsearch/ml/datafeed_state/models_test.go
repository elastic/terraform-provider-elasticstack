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
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetStartAndEndFromAPI_StoppedState_UnknownStart(t *testing.T) {
	data := &MLDatafeedStateData{
		Start: timetypes.NewRFC3339Unknown(),
		End:   timetypes.NewRFC3339Unknown(),
	}

	stats := &models.DatafeedStats{
		State: "stopped",
	}

	diags := data.SetStartAndEndFromAPI(stats)
	require.False(t, diags.HasError(), "unexpected errors: %v", diags)
	assert.True(t, data.Start.IsNull(), "start should be null for stopped datafeed with unknown start")
	assert.True(t, data.End.IsNull(), "end should be null for stopped datafeed with unknown end")
}

func TestSetStartAndEndFromAPI_StoppedState_NullStart(t *testing.T) {
	data := &MLDatafeedStateData{
		Start: timetypes.NewRFC3339Null(),
		End:   timetypes.NewRFC3339Null(),
	}

	stats := &models.DatafeedStats{
		State: "stopped",
	}

	diags := data.SetStartAndEndFromAPI(stats)
	require.False(t, diags.HasError(), "unexpected errors: %v", diags)
	assert.True(t, data.Start.IsNull(), "start should remain null for stopped datafeed")
	assert.True(t, data.End.IsNull(), "end should remain null for stopped datafeed")
}

func TestSetStartAndEndFromAPI_StartedState_WithRunningState(t *testing.T) {
	data := &MLDatafeedStateData{
		Start: timetypes.NewRFC3339Unknown(),
		End:   timetypes.NewRFC3339Unknown(),
	}

	startMS := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	endMS := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()

	stats := &models.DatafeedStats{
		State: "started",
		RunningState: &models.DatafeedRunning{
			RealTimeConfigured: false,
			SearchInterval: &models.DatafeedSearchInterval{
				StartMS: startMS,
				EndMS:   endMS,
			},
		},
	}

	diags := data.SetStartAndEndFromAPI(stats)
	require.False(t, diags.HasError(), "unexpected errors: %v", diags)
	assert.False(t, data.Start.IsNull(), "start should not be null for started datafeed")
	assert.False(t, data.Start.IsUnknown(), "start should not be unknown for started datafeed")
	assert.False(t, data.End.IsNull(), "end should not be null for non-realtime started datafeed")
	assert.False(t, data.End.IsUnknown(), "end should not be unknown for non-realtime started datafeed")
}

func TestSetStartAndEndFromAPI_StartedState_RealTime(t *testing.T) {
	data := &MLDatafeedStateData{
		Start: timetypes.NewRFC3339Unknown(),
		End:   timetypes.NewRFC3339Unknown(),
	}

	startMS := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	endMS := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()

	stats := &models.DatafeedStats{
		State: "started",
		RunningState: &models.DatafeedRunning{
			RealTimeConfigured: true,
			SearchInterval: &models.DatafeedSearchInterval{
				StartMS: startMS,
				EndMS:   endMS,
			},
		},
	}

	diags := data.SetStartAndEndFromAPI(stats)
	require.False(t, diags.HasError(), "unexpected errors: %v", diags)
	assert.False(t, data.Start.IsNull(), "start should not be null for started datafeed")
	assert.False(t, data.Start.IsUnknown(), "start should not be unknown for started datafeed")
	assert.True(t, data.End.IsNull(), "end should be null for real-time datafeed")
}

func TestSetStartAndEndFromAPI_StartedState_NilRunningState(t *testing.T) {
	data := &MLDatafeedStateData{
		Start: timetypes.NewRFC3339Unknown(),
		End:   timetypes.NewRFC3339Unknown(),
	}

	stats := &models.DatafeedStats{
		State:        "started",
		RunningState: nil,
	}

	diags := data.SetStartAndEndFromAPI(stats)
	assert.True(t, diags.HasError() || len(diags) > 0, "expected warning for nil running state")
	assert.True(t, data.Start.IsNull(), "start should be null when running state is nil")
	assert.True(t, data.End.IsNull(), "end should be null when running state is nil")
}

func TestResolveUnknowns(t *testing.T) {
	data := &MLDatafeedStateData{
		Start: timetypes.NewRFC3339Unknown(),
		End:   timetypes.NewRFC3339Unknown(),
	}

	data.resolveUnknowns()

	assert.True(t, data.Start.IsNull(), "unknown start should resolve to null")
	assert.True(t, data.End.IsNull(), "unknown end should resolve to null")

	// Concrete values should not be affected
	data2 := &MLDatafeedStateData{
		Start: timetypes.NewRFC3339TimeValue(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
		End:   timetypes.NewRFC3339Null(),
	}

	data2.resolveUnknowns()

	assert.False(t, data2.Start.IsNull(), "concrete start should not be changed")
	assert.True(t, data2.End.IsNull(), "null end should remain null")
}
