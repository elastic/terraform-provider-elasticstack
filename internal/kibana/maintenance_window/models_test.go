package maintenance_window

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

var modelWithAllFields = MaintenanceWindowModel{
	Title:   types.StringValue("test response"),
	Enabled: types.BoolValue(true),

	CustomSchedule: MaintenanceWindowSchedule{
		Start:    types.StringValue("1993-01-01T05:00:00.200Z"),
		Duration: types.StringValue("13d"),
		Timezone: types.StringValue("America/Martinique"),

		Recurring: &MaintenanceWindowScheduleRecurring{
			Every:       types.StringValue("21d"),
			End:         types.StringValue("2029-05-17T05:05:00.000Z"),
			Occurrences: types.Int32Null(),
			OnWeekDay:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("MO"), types.StringValue("-2FR"), types.StringValue("+4SA")}),
			OnMonth:     types.ListValueMust(types.Int32Type, []attr.Value{types.Int32Value(6)}),
			OnMonthDay:  types.ListValueMust(types.Int32Type, []attr.Value{types.Int32Value(1), types.Int32Value(2), types.Int32Value(3)}),
		},
	},

	Scope: &MaintenanceWindowScope{
		Alerting: MaintenanceWindowAlertingScope{
			Kql: types.StringValue("_id: '1234'"),
		},
	},
}

var modelOccurrencesNoScope = MaintenanceWindowModel{
	Title:   types.StringValue("test response"),
	Enabled: types.BoolValue(true),

	CustomSchedule: MaintenanceWindowSchedule{
		Start:    types.StringValue("1993-01-01T05:00:00.200Z"),
		Duration: types.StringValue("13d"),
		Timezone: types.StringNull(),

		Recurring: &MaintenanceWindowScheduleRecurring{
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
		response      ResponseJson
		existingModel MaintenanceWindowModel
		expectedModel MaintenanceWindowModel
	}{
		{
			name:          "all fields",
			existingModel: MaintenanceWindowModel{},
			response: ResponseJson{
				Id:        "existing-space-id/id",
				CreatedAt: "created_at",
				Enabled:   true,
				Title:     "test response",
				Schedule: ResponseJsonSchedule{
					Custom: ResponseJsonCustomSchedule{
						Start:    "1993-01-01T05:00:00.200Z",
						Duration: "13d",
						Timezone: utils.Pointer("America/Martinique"),
						Recurring: &ResponseJsonRecurring{
							Every:      utils.Pointer("21d"),
							End:        utils.Pointer("2029-05-17T05:05:00.000Z"),
							OnWeekDay:  utils.Pointer([]string{"MO", "-2FR", "+4SA"}),
							OnMonth:    utils.Pointer([]float32{6}),
							OnMonthDay: utils.Pointer([]float32{1, 2, 3}),
						},
					},
				},
				Scope: &ResponseJsonScope{
					Alerting: ResponseJsonAlerting{
						Query: ResponseJsonAlertingQuery{
							Kql: "_id: '1234'",
						},
					},
				},
			},
			expectedModel: modelWithAllFields,
		},
		{
			name:          "occurrences and no scope",
			existingModel: MaintenanceWindowModel{},
			response: ResponseJson{
				Id:        "existing-space-id/id",
				CreatedAt: "created_at",
				Enabled:   true,
				Title:     "test response",
				Schedule: ResponseJsonSchedule{
					Custom: ResponseJsonCustomSchedule{
						Start:    "1993-01-01T05:00:00.200Z",
						Duration: "13d",
						Recurring: &ResponseJsonRecurring{
							Every:       utils.Pointer("21d"),
							Occurrences: utils.Pointer(float32(42)),
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
		model           MaintenanceWindowModel
		expectedRequest kbapi.PostMaintenanceWindowJSONRequestBody
	}{
		{
			name:  "all fields",
			model: modelWithAllFields,
			expectedRequest: kbapi.PostMaintenanceWindowJSONRequestBody{
				Enabled: utils.Pointer(true),
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
						Timezone: utils.Pointer("America/Martinique"),
						Recurring: utils.Pointer(struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						}{
							Every:      utils.Pointer("21d"),
							End:        utils.Pointer("2029-05-17T05:05:00.000Z"),
							OnWeekDay:  utils.Pointer([]string{"MO", "-2FR", "+4SA"}),
							OnMonth:    utils.Pointer([]float32{6}),
							OnMonthDay: utils.Pointer([]float32{1, 2, 3}),
						}),
					},
				},
				Scope: utils.Pointer(struct {
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
				Enabled: utils.Pointer(true),
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

						Recurring: utils.Pointer(struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						}{
							Every:       utils.Pointer("21d"),
							Occurrences: utils.Pointer(float32(42)),
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
			require.Equal(t, request, tt.expectedRequest)
			require.Empty(t, diags)
		})
	}
}

func TestMaintenanceWindowToAPIUpdateRequest(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name            string
		model           MaintenanceWindowModel
		expectedRequest kbapi.PatchMaintenanceWindowIdJSONRequestBody
	}{
		{
			name:  "all fields",
			model: modelWithAllFields,
			expectedRequest: kbapi.PatchMaintenanceWindowIdJSONRequestBody{
				Enabled: utils.Pointer(true),
				Title:   utils.Pointer("test response"),
				Schedule: utils.Pointer(struct {
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
						Timezone: utils.Pointer("America/Martinique"),
						Recurring: utils.Pointer(struct {
							End         *string    `json:"end,omitempty"`
							Every       *string    `json:"every,omitempty"`
							Occurrences *float32   `json:"occurrences,omitempty"`
							OnMonth     *[]float32 `json:"onMonth,omitempty"`
							OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
							OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
						}{
							Every:      utils.Pointer("21d"),
							End:        utils.Pointer("2029-05-17T05:05:00.000Z"),
							OnWeekDay:  utils.Pointer([]string{"MO", "-2FR", "+4SA"}),
							OnMonth:    utils.Pointer([]float32{6}),
							OnMonthDay: utils.Pointer([]float32{1, 2, 3}),
						}),
					},
				}),
				Scope: utils.Pointer(struct {
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
			model: MaintenanceWindowModel{
				ID:      types.StringValue("/existing-space-id/id"),
				Title:   types.StringValue("test response"),
				Enabled: types.BoolValue(true),
				CustomSchedule: MaintenanceWindowSchedule{
					Start:    types.StringValue("1993-01-01T05:00:00.200Z"),
					Duration: types.StringValue("13d"),
				},
			},
			expectedRequest: kbapi.PatchMaintenanceWindowIdJSONRequestBody{
				Enabled: utils.Pointer(true),
				Title:   utils.Pointer("test response"),
				Schedule: utils.Pointer(struct {
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
			model: MaintenanceWindowModel{
				ID: types.StringValue("/existing-space-id/id"),

				CustomSchedule: MaintenanceWindowSchedule{
					Start:    types.StringValue("1993-01-01T05:00:00.200Z"),
					Duration: types.StringValue("13d"),
				},

				Scope: &MaintenanceWindowScope{
					Alerting: MaintenanceWindowAlertingScope{
						Kql: types.StringValue("_id: '1234'"),
					},
				},
			},
			expectedRequest: kbapi.PatchMaintenanceWindowIdJSONRequestBody{
				Schedule: utils.Pointer(struct {
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

				Scope: utils.Pointer(struct {
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
			require.Equal(t, request, tt.expectedRequest)
			require.Empty(t, diags)
		})
	}
}
