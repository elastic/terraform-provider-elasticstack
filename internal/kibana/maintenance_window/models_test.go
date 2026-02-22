package maintenancewindow

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
		Duration: types.StringValue("13d"),
		Timezone: types.StringValue("America/Martinique"),

		Recurring: &ScheduleRecurring{
			Every:       types.StringValue("21d"),
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
		Duration: types.StringValue("13d"),
		Timezone: types.StringNull(),

		Recurring: &ScheduleRecurring{
			Every:       types.StringValue("21d"),
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

	tests := []struct {
		name          string
		response      ResponseJSON
		existingModel Model
		expectedModel Model
	}{
		{
			name:          "all fields",
			existingModel: Model{},
			response: ResponseJSON{
				ID:        "existing-space-id/id",
				CreatedAt: "created_at",
				Enabled:   true,
				Title:     "test response",
				Schedule: ResponseJSONSchedule{
					Custom: ResponseJSONCustomSchedule{
						Start:    "1993-01-01T05:00:00.200Z",
						Duration: "13d",
						Timezone: schemautil.Pointer("America/Martinique"),
						Recurring: &ResponseJSONRecurring{
							Every:      schemautil.Pointer("21d"),
							End:        schemautil.Pointer("2029-05-17T05:05:00.000Z"),
							OnWeekDay:  schemautil.Pointer([]string{"MO", "-2FR", "+4SA"}),
							OnMonth:    schemautil.Pointer([]float32{6}),
							OnMonthDay: schemautil.Pointer([]float32{1, 2, 3}),
						},
					},
				},
				Scope: &ResponseJSONScope{
					Alerting: ResponseJSONAlerting{
						Query: ResponseJSONAlertingQuery{
							Kql: "_id: '1234'",
						},
					},
				},
			},
			expectedModel: modelWithAllFields,
		},
		{
			name:          "occurrences and no scope",
			existingModel: Model{},
			response: ResponseJSON{
				ID:        "existing-space-id/id",
				CreatedAt: "created_at",
				Enabled:   true,
				Title:     "test response",
				Schedule: ResponseJSONSchedule{
					Custom: ResponseJSONCustomSchedule{
						Start:    "1993-01-01T05:00:00.200Z",
						Duration: "13d",
						Recurring: &ResponseJSONRecurring{
							Every:       schemautil.Pointer("21d"),
							Occurrences: schemautil.Pointer(float32(42)),
						},
					},
				},
				Scope: nil,
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

func TestMaintenanceWindowToAPICreateRequest(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name            string
		model           Model
		expectedRequest kbapi.PostMaintenanceWindowJSONRequestBody
	}{
		{
			name:  "all fields",
			model: modelWithAllFields,
			expectedRequest: kbapi.PostMaintenanceWindowJSONRequestBody{
				Enabled: schemautil.Pointer(true),
				Title:   "test response",
				Schedule: struct {
					Custom struct {
						Duration  string `json:"duration"`
						Recurring *struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						} `json:"recurring,omitempty"`
						Start    string  `json:"start"`
						Timezone *string `json:"timezone,omitempty"`
					} `json:"custom"`
				}{
					Custom: struct {
						Duration  string `json:"duration"`
						Recurring *struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						} `json:"recurring,omitempty"`
						Start    string  `json:"start"`
						Timezone *string `json:"timezone,omitempty"`
					}{
						Start:    "1993-01-01T05:00:00.200Z",
						Duration: "13d",
						Timezone: schemautil.Pointer("America/Martinique"),
						Recurring: schemautil.Pointer(struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						}{
							Every:      schemautil.Pointer("21d"),
							End:        schemautil.Pointer("2029-05-17T05:05:00.000Z"),
							OnWeekDay:  schemautil.Pointer([]string{"MO", "-2FR", "+4SA"}),
							OnMonth:    schemautil.Pointer([]float32{6}),
							OnMonthDay: schemautil.Pointer([]float32{1, 2, 3}),
						}),
					},
				},
				Scope: schemautil.Pointer(struct {
					Alerting struct {
						Query struct {
							Kql string `json:"kql"`
						} `json:"query"`
					} `json:"alerting"`
				}{
					Alerting: struct {
						Query struct {
							Kql string `json:"kql"`
						} `json:"query"`
					}{
						Query: struct {
							Kql string `json:"kql"`
						}{
							Kql: "_id: '1234'",
						},
					},
				},
				),
			},
		},
		{
			name:  "occurrences and no scope",
			model: modelOccurrencesNoScope,
			expectedRequest: kbapi.PostMaintenanceWindowJSONRequestBody{
				Enabled: schemautil.Pointer(true),
				Title:   "test response",
				Schedule: struct {
					Custom struct {
						Duration  string `json:"duration"`
						Recurring *struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						} `json:"recurring,omitempty"`
						Start    string  `json:"start"`
						Timezone *string `json:"timezone,omitempty"`
					} `json:"custom"`
				}{
					Custom: struct {
						Duration  string `json:"duration"`
						Recurring *struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						} `json:"recurring,omitempty"`
						Start    string  `json:"start"`
						Timezone *string `json:"timezone,omitempty"`
					}{
						Start:    "1993-01-01T05:00:00.200Z",
						Duration: "13d",

						Recurring: schemautil.Pointer(struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						}{
							Every:       schemautil.Pointer("21d"),
							Occurrences: schemautil.Pointer(float32(42)),
						}),
					},
				},
			},
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, diags := tt.model.toAPICreateRequest(ctx)
			require.Equal(t, tt.expectedRequest, request)
			require.Empty(t, diags)
		})
	}
}

func TestMaintenanceWindowToAPIUpdateRequest(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name            string
		model           Model
		expectedRequest kbapi.PatchMaintenanceWindowIdJSONRequestBody
	}{
		{
			name:  "all fields",
			model: modelWithAllFields,
			expectedRequest: kbapi.PatchMaintenanceWindowIdJSONRequestBody{
				Enabled: schemautil.Pointer(true),
				Title:   schemautil.Pointer("test response"),
				Schedule: schemautil.Pointer(struct {
					Custom struct {
						Duration  string `json:"duration"`
						Recurring *struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						} `json:"recurring,omitempty"`
						Start    string  `json:"start"`
						Timezone *string `json:"timezone,omitempty"`
					} `json:"custom"`
				}{
					Custom: struct {
						Duration  string `json:"duration"`
						Recurring *struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						} `json:"recurring,omitempty"`
						Start    string  `json:"start"`
						Timezone *string `json:"timezone,omitempty"`
					}{
						Start:    "1993-01-01T05:00:00.200Z",
						Duration: "13d",
						Timezone: schemautil.Pointer("America/Martinique"),
						Recurring: schemautil.Pointer(struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						}{
							Every:      schemautil.Pointer("21d"),
							End:        schemautil.Pointer("2029-05-17T05:05:00.000Z"),
							OnWeekDay:  schemautil.Pointer([]string{"MO", "-2FR", "+4SA"}),
							OnMonth:    schemautil.Pointer([]float32{6}),
							OnMonthDay: schemautil.Pointer([]float32{1, 2, 3}),
						}),
					},
				}),
				Scope: schemautil.Pointer(struct {
					Alerting struct {
						Query struct {
							Kql string `json:"kql"`
						} `json:"query"`
					} `json:"alerting"`
				}{
					Alerting: struct {
						Query struct {
							Kql string `json:"kql"`
						} `json:"query"`
					}{
						Query: struct {
							Kql string `json:"kql"`
						}{
							Kql: "_id: '1234'",
						},
					},
				},
				),
			},
		},
		{
			name: "just title, enabled and schedule",
			model: Model{
				ID:      types.StringValue("/existing-space-id/id"),
				Title:   types.StringValue("test response"),
				Enabled: types.BoolValue(true),
				CustomSchedule: Schedule{
					Start:    types.StringValue("1993-01-01T05:00:00.200Z"),
					Duration: types.StringValue("13d"),
				},
			},
			expectedRequest: kbapi.PatchMaintenanceWindowIdJSONRequestBody{
				Enabled: schemautil.Pointer(true),
				Title:   schemautil.Pointer("test response"),
				Schedule: schemautil.Pointer(struct {
					Custom struct {
						Duration  string `json:"duration"`
						Recurring *struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						} `json:"recurring,omitempty"`
						Start    string  `json:"start"`
						Timezone *string `json:"timezone,omitempty"`
					} `json:"custom"`
				}{
					Custom: struct {
						Duration  string `json:"duration"`
						Recurring *struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						} `json:"recurring,omitempty"`
						Start    string  `json:"start"`
						Timezone *string `json:"timezone,omitempty"`
					}{
						Start:    "1993-01-01T05:00:00.200Z",
						Duration: "13d",
					},
				}),
			},
		},
		{
			name: "just the scope and schedule",
			model: Model{
				ID: types.StringValue("/existing-space-id/id"),

				CustomSchedule: Schedule{
					Start:    types.StringValue("1993-01-01T05:00:00.200Z"),
					Duration: types.StringValue("13d"),
				},

				Scope: &Scope{
					Alerting: AlertingScope{
						Kql: types.StringValue("_id: '1234'"),
					},
				},
			},
			expectedRequest: kbapi.PatchMaintenanceWindowIdJSONRequestBody{
				Schedule: schemautil.Pointer(struct {
					Custom struct {
						Duration  string `json:"duration"`
						Recurring *struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						} `json:"recurring,omitempty"`
						Start    string  `json:"start"`
						Timezone *string `json:"timezone,omitempty"`
					} `json:"custom"`
				}{
					Custom: struct {
						Duration  string `json:"duration"`
						Recurring *struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						} `json:"recurring,omitempty"`
						Start    string  `json:"start"`
						Timezone *string `json:"timezone,omitempty"`
					}{
						Start:    "1993-01-01T05:00:00.200Z",
						Duration: "13d",
					},
				}),

				Scope: schemautil.Pointer(struct {
					Alerting struct {
						Query struct {
							Kql string `json:"kql"`
						} `json:"query"`
					} `json:"alerting"`
				}{
					Alerting: struct {
						Query struct {
							Kql string `json:"kql"`
						} `json:"query"`
					}{
						Query: struct {
							Kql string `json:"kql"`
						}{
							Kql: "_id: '1234'",
						},
					},
				},
				),
			},
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, diags := tt.model.toAPIUpdateRequest(ctx)
			require.Equal(t, tt.expectedRequest, request)
			require.Empty(t, diags)
		})
	}
}
