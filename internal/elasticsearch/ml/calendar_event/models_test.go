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
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarEventTFModel_toAPIModel(t *testing.T) {
	ctx := context.Background()

	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	model := &CalendarEventTFModel{
		Description: types.StringValue("maintenance window"),
		StartTime:   timetypes.NewRFC3339TimeValue(startTime),
		EndTime:     timetypes.NewRFC3339TimeValue(endTime),
	}

	apiModel, diags := model.toAPIModel(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.Equal(t, "maintenance window", apiModel.Description)
	assert.Equal(t, startTime.UnixMilli(), apiModel.StartTime)
	assert.Equal(t, endTime.UnixMilli(), apiModel.EndTime)
}

func TestCalendarEventTFModel_fromAPIModel(t *testing.T) {
	ctx := context.Background()

	startMillis := float64(time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC).UnixMilli())
	endMillis := float64(time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC).UnixMilli())

	tests := []struct {
		name            string
		initialModel    *CalendarEventTFModel
		apiModel        *CalendarEventAPIModel
		expectedEventID string
		expectedCalID   string
		expectedDesc    string
	}{
		{
			name: "populates all fields from API",
			initialModel: &CalendarEventTFModel{
				StartTime: timetypes.NewRFC3339TimeValue(time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)),
				EndTime:   timetypes.NewRFC3339TimeValue(time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC)),
			},
			apiModel: &CalendarEventAPIModel{
				Description: "planned outage",
				StartTime:   startMillis,
				EndTime:     endMillis,
				CalendarID:  "ops-calendar",
				EventID:     "evt-123",
			},
			expectedEventID: "evt-123",
			expectedCalID:   "ops-calendar",
			expectedDesc:    "planned outage",
		},
		{
			name: "handles unknown start/end times gracefully",
			initialModel: &CalendarEventTFModel{
				StartTime: timetypes.NewRFC3339Unknown(),
				EndTime:   timetypes.NewRFC3339Unknown(),
			},
			apiModel: &CalendarEventAPIModel{
				Description: "event",
				StartTime:   startMillis,
				EndTime:     endMillis,
				CalendarID:  "cal-1",
				EventID:     "evt-456",
			},
			expectedEventID: "evt-456",
			expectedCalID:   "cal-1",
			expectedDesc:    "event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := tt.initialModel.fromAPIModel(ctx, tt.apiModel)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			assert.Equal(t, tt.expectedEventID, tt.initialModel.EventID.ValueString())
			assert.Equal(t, tt.expectedCalID, tt.initialModel.CalendarID.ValueString())
			assert.Equal(t, tt.expectedDesc, tt.initialModel.Description.ValueString())

			resultStartTime, d := tt.initialModel.StartTime.ValueRFC3339Time()
			require.False(t, d.HasError())
			assert.Equal(t, int64(startMillis), resultStartTime.UnixMilli())

			resultEndTime, d := tt.initialModel.EndTime.ValueRFC3339Time()
			require.False(t, d.HasError())
			assert.Equal(t, int64(endMillis), resultEndTime.UnixMilli())
		})
	}
}

func TestCalendarEventTFModel_fromAPIModel_invalidTypes(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid start_time type", func(t *testing.T) {
		model := &CalendarEventTFModel{
			StartTime: timetypes.NewRFC3339Unknown(),
			EndTime:   timetypes.NewRFC3339Unknown(),
		}
		apiModel := &CalendarEventAPIModel{
			Description: "event",
			StartTime:   "not-a-number",
			EndTime:     float64(1000),
			CalendarID:  "cal",
			EventID:     "evt",
		}
		diags := model.fromAPIModel(ctx, apiModel)
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Summary(), "Invalid start_time format")
	})

	t.Run("invalid end_time type", func(t *testing.T) {
		model := &CalendarEventTFModel{
			StartTime: timetypes.NewRFC3339Unknown(),
			EndTime:   timetypes.NewRFC3339Unknown(),
		}
		apiModel := &CalendarEventAPIModel{
			Description: "event",
			StartTime:   float64(1000),
			EndTime:     "not-a-number",
			CalendarID:  "cal",
			EventID:     "evt",
		}
		diags := model.fromAPIModel(ctx, apiModel)
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Summary(), "Invalid end_time format")
	})
}

func TestParseCompositeID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		expectedCalID  string
		expectedEvtID  string
		expectError    bool
	}{
		{
			name:          "valid composite ID",
			id:            "cluster-uuid/my-calendar/event-123",
			expectedCalID: "my-calendar",
			expectedEvtID: "event-123",
		},
		{
			name:        "missing event_id",
			id:          "cluster-uuid/my-calendar",
			expectError: true,
		},
		{
			name:        "single segment",
			id:          "just-an-id",
			expectError: true,
		},
		{
			name:          "event_id with slashes",
			id:            "cluster-uuid/my-calendar/event/with/slashes",
			expectedCalID: "my-calendar",
			expectedEvtID: "event/with/slashes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calID, evtID, diags := parseCompositeID(tt.id)
			if tt.expectError {
				assert.True(t, diags.HasError())
			} else {
				require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
				assert.Equal(t, tt.expectedCalID, calID)
				assert.Equal(t, tt.expectedEvtID, evtID)
			}
		})
	}
}
