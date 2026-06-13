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

package maintenancewindow

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kibanacustomtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

var modelWithAllFields = Model{
	Title:   types.StringValue("test response"),
	Enabled: types.BoolValue(true),

	CustomSchedule: Schedule{
		Start:    types.StringValue("1993-01-01T05:00:00.200Z"),
		Duration: kibanacustomtypes.NewAlertingDurationValue("13d"),
		Timezone: types.StringValue("America/Martinique"),

		Recurring: &ScheduleRecurring{
			Every:       kibanacustomtypes.NewAlertingDurationValue("21d"),
			End:         types.StringValue("2029-05-17T05:05:00.000Z"),
			Occurrences: types.Int32Null(),
			OnWeekDay:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("MO"), types.StringValue("-2FR"), types.StringValue("+4SA")}),
			OnMonth:     types.ListValueMust(types.Int32Type, []attr.Value{types.Int32Value(6)}),
			OnMonthDay:  types.ListValueMust(types.Int32Type, []attr.Value{types.Int32Value(1), types.Int32Value(2), types.Int32Value(3)}),
		},
	},

	Scope: &Scope{
		Alerting: AlertingScope{
			Kql: types.StringValue("_id: '1234'"),
		},
	},
}

var modelOccurrencesNoScope = Model{
	Title:   types.StringValue("test response"),
	Enabled: types.BoolValue(true),

	CustomSchedule: Schedule{
		Start:    types.StringValue("1993-01-01T05:00:00.200Z"),
		Duration: kibanacustomtypes.NewAlertingDurationValue("13d"),
		Timezone: types.StringNull(),

		Recurring: &ScheduleRecurring{
			Every:       kibanacustomtypes.NewAlertingDurationValue("21d"),
			End:         types.StringNull(),
			Occurrences: types.Int32Value(42),
			OnWeekDay:   types.ListNull(types.StringType),
			OnMonth:     types.ListNull(types.Int32Type),
			OnMonthDay:  types.ListNull(types.Int32Type),
		},
	},

	Scope: nil,
}

func TestMaintenanceWindowFromAPI(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tz := "America/Martinique"
	every := "21d"
	end := "2029-05-17T05:05:00.000Z"
	onWeekDay := []string{"MO", "-2FR", "+4SA"}
	onMonth := []float32{6}
	onMonthDay := []float32{1, 2, 3}
	duration := "13d"
	start := "1993-01-01T05:00:00.200Z"
	kql := "_id: '1234'"
	var occurrences float32 = 42

	tests := []struct {
		name          string
		response      kbapi.KibanaHTTPAPIsMaintenanceWindowResponse
		existingModel Model
		expectedModel Model
	}{
		{
			name:          "all fields",
			existingModel: Model{},
			response: kbapi.KibanaHTTPAPIsMaintenanceWindowResponse{
				Id:        "existing-space-id/id",
				CreatedAt: "created_at",
				Enabled:   true,
				Title:     "test response",
				Schedule: struct {
					Custom kbapi.KibanaHTTPAPIsMaintenanceWindowScheduleResponse `json:"custom"`
				}{
					Custom: kbapi.KibanaHTTPAPIsMaintenanceWindowScheduleResponse{
						Start:    start,
						Duration: duration,
						Timezone: &tz,
						Recurring: &kbapi.KibanaHTTPAPIsMaintenanceWindowScheduleRecurringResponse{
							Every:      &every,
							End:        &end,
							OnWeekDay:  &onWeekDay,
							OnMonth:    &onMonth,
							OnMonthDay: &onMonthDay,
						},
					},
				},
				Scope: &kbapi.KibanaHTTPAPIsMaintenanceWindowScope{
					Alerting: struct {
						Query struct {
							Kql string `json:"kql"`
						} `json:"query"`
					}{
						Query: struct {
							Kql string `json:"kql"`
						}{
							Kql: kql,
						},
					},
				},
			},
			expectedModel: modelWithAllFields,
		},
		{
			name:          "occurrences and no scope",
			existingModel: Model{},
			response: kbapi.KibanaHTTPAPIsMaintenanceWindowResponse{
				Id:        "existing-space-id/id",
				CreatedAt: "created_at",
				Enabled:   true,
				Title:     "test response",
				Schedule: struct {
					Custom kbapi.KibanaHTTPAPIsMaintenanceWindowScheduleResponse `json:"custom"`
				}{
					Custom: kbapi.KibanaHTTPAPIsMaintenanceWindowScheduleResponse{
						Start:    start,
						Duration: duration,
						Recurring: &kbapi.KibanaHTTPAPIsMaintenanceWindowScheduleRecurringResponse{
							Every:       &every,
							Occurrences: &occurrences,
						},
					},
				},
			},
			expectedModel: modelOccurrencesNoScope,
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := tt.existingModel._fromAPIResponse(ctx, tt.response)

			require.Equal(t, tt.expectedModel, tt.existingModel)
			require.Empty(t, diags)
		})
	}
}
