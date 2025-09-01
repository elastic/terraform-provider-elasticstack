package maintenance_window

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MaintenanceWindowModel struct {
	ID             types.String              `tfsdk:"id"`
	SpaceID        types.String              `tfsdk:"space_id"`
	Title          types.String              `tfsdk:"title"`
	Enabled        types.Bool                `tfsdk:"enabled"`
	CustomSchedule MaintenanceWindowSchedule `tfsdk:"custom_schedule"`
	Scope          *MaintenanceWindowScope   `tfsdk:"scope"`
}

type MaintenanceWindowScope struct {
	Alerting MaintenanceWindowAlertingScope `tfsdk:"alerting"`
}

type MaintenanceWindowAlertingScope struct {
	Kql types.String `tfsdk:"kql"`
}

type MaintenanceWindowSchedule struct {
	Start     types.String                        `tfsdk:"start"`
	Duration  types.String                        `tfsdk:"duration"`
	Timezone  types.String                        `tfsdk:"timezone"`
	Recurring *MaintenanceWindowScheduleRecurring `tfsdk:"recurring"`
}

type MaintenanceWindowScheduleRecurring struct {
	End         types.String `tfsdk:"end"`
	Every       types.String `tfsdk:"every"`
	Occurrences types.Int32  `tfsdk:"occurrences"`
	OnWeekDay   types.List   `tfsdk:"on_week_day"`
	OnMonthDay  types.List   `tfsdk:"on_month_day"`
	OnMonth     types.List   `tfsdk:"on_month"`
}

/* CREATE */

func (model MaintenanceWindowModel) toAPICreateRequest(ctx context.Context) (kbapi.PostMaintenanceWindowJSONRequestBody, diag.Diagnostics) {
	var diags = diag.Diagnostics{}

	body := kbapi.PostMaintenanceWindowJSONRequestBody{
		Enabled: model.Enabled.ValueBoolPointer(),
		Title:   *model.Title.ValueStringPointer(),
	}

	body.Schedule.Custom.Duration = model.CustomSchedule.Duration.ValueString()
	body.Schedule.Custom.Start = model.CustomSchedule.Start.ValueString()

	if !model.CustomSchedule.Timezone.IsNull() && !model.CustomSchedule.Timezone.IsUnknown() {
		body.Schedule.Custom.Timezone = model.CustomSchedule.Timezone.ValueStringPointer()
	}

	if model.CustomSchedule.Recurring != nil {

		body.Schedule.Custom.Recurring = &struct {
			End         *string    `json:"end,omitempty"`
			Every       *string    `json:"every,omitempty"`
			Occurrences *float32   `json:"occurrences,omitempty"`
			OnMonth     *[]float32 `json:"onMonth,omitempty"`
			OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
			OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
		}{
			End:   model.CustomSchedule.Recurring.End.ValueStringPointer(),
			Every: model.CustomSchedule.Recurring.Every.ValueStringPointer(),
		}

		if !model.CustomSchedule.Recurring.Occurrences.IsNull() && !model.CustomSchedule.Recurring.Occurrences.IsUnknown() && model.CustomSchedule.Recurring.Occurrences.ValueInt32() > 0 {
			occurrences := float32(model.CustomSchedule.Recurring.Occurrences.ValueInt32())
			body.Schedule.Custom.Recurring.Occurrences = &occurrences
		}

		if utils.IsKnown(model.CustomSchedule.Recurring.OnWeekDay) {
			var onWeekDay []string
			diags.Append(model.CustomSchedule.Recurring.OnWeekDay.ElementsAs(ctx, &onWeekDay, true)...)
			body.Schedule.Custom.Recurring.OnWeekDay = &onWeekDay
		}

		if utils.IsKnown(model.CustomSchedule.Recurring.OnMonth) {
			var onMonth []float32
			diags.Append(model.CustomSchedule.Recurring.OnMonth.ElementsAs(ctx, &onMonth, true)...)
			body.Schedule.Custom.Recurring.OnMonth = &onMonth
		}

		if utils.IsKnown(model.CustomSchedule.Recurring.OnMonthDay) {
			var onMonthDay []float32
			diags.Append(model.CustomSchedule.Recurring.OnMonthDay.ElementsAs(ctx, &onMonthDay, true)...)
			body.Schedule.Custom.Recurring.OnMonthDay = &onMonthDay
		}
	}

	if model.Scope != nil {
		body.Scope = &struct {
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
					Kql: model.Scope.Alerting.Kql.ValueString(),
				},
			},
		}
	}

	return body, diags
}

func (model *MaintenanceWindowModel) fromAPICreateResponse(ctx context.Context, data *kbapi.PostMaintenanceWindowResponse) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags = diag.Diagnostics{}
	var response = &ResponseJson{}

	if err := json.Unmarshal(data.Body, response); err != nil {
		diags.AddError(err.Error(), "cannot unmarshal PostMaintenanceWindowResponse")
		return diags
	}

	return model._fromAPIResponse(ctx, *response)
}

/* READ */

func (model *MaintenanceWindowModel) fromAPIReadResponse(ctx context.Context, data *kbapi.GetMaintenanceWindowIdResponse) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags = diag.Diagnostics{}
	var response = &ResponseJson{}

	if err := json.Unmarshal(data.Body, response); err != nil {
		diags.AddError(err.Error(), "cannot unmarshal GetMaintenanceWindowIdResponse")
		return diags
	}

	return model._fromAPIResponse(ctx, *response)
}

/* UPDATE */

func (model MaintenanceWindowModel) toAPIUpdateRequest(ctx context.Context) (kbapi.PatchMaintenanceWindowIdJSONRequestBody, diag.Diagnostics) {
	var diags = diag.Diagnostics{}

	body := kbapi.PatchMaintenanceWindowIdJSONRequestBody{
		Enabled: model.Enabled.ValueBoolPointer(),
		Title:   model.Title.ValueStringPointer(),
	}

	schedule := struct {
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
			Duration: model.CustomSchedule.Duration.ValueString(),
			Start:    model.CustomSchedule.Start.ValueString(),
		},
	}

	body.Schedule = &schedule

	if utils.IsKnown(model.CustomSchedule.Timezone) {
		body.Schedule.Custom.Timezone = model.CustomSchedule.Timezone.ValueStringPointer()
	}

	if model.CustomSchedule.Recurring != nil {

		body.Schedule.Custom.Recurring = &struct {
			End         *string    `json:"end,omitempty"`
			Every       *string    `json:"every,omitempty"`
			Occurrences *float32   `json:"occurrences,omitempty"`
			OnMonth     *[]float32 `json:"onMonth,omitempty"`
			OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
			OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
		}{}

		if !model.CustomSchedule.Recurring.End.IsNull() {
			body.Schedule.Custom.Recurring.End = model.CustomSchedule.Recurring.End.ValueStringPointer()
		}

		if !model.CustomSchedule.Recurring.Every.IsNull() {
			body.Schedule.Custom.Recurring.Every = model.CustomSchedule.Recurring.Every.ValueStringPointer()
		}

		if !model.CustomSchedule.Recurring.Occurrences.IsNull() && !model.CustomSchedule.Recurring.Occurrences.IsUnknown() && model.CustomSchedule.Recurring.Occurrences.ValueInt32() > 0 {
			occurrences := float32(model.CustomSchedule.Recurring.Occurrences.ValueInt32())
			body.Schedule.Custom.Recurring.Occurrences = &occurrences
		}

		if !model.CustomSchedule.Recurring.OnWeekDay.IsNull() && !model.CustomSchedule.Recurring.OnWeekDay.IsUnknown() {
			var onWeekDay []string
			diags.Append(model.CustomSchedule.Recurring.OnWeekDay.ElementsAs(ctx, &onWeekDay, true)...)
			body.Schedule.Custom.Recurring.OnWeekDay = &onWeekDay
		}

		if !model.CustomSchedule.Recurring.OnMonth.IsNull() && !model.CustomSchedule.Recurring.OnMonth.IsUnknown() {
			var onMonth []float32
			diags.Append(model.CustomSchedule.Recurring.OnMonth.ElementsAs(ctx, &onMonth, true)...)
			body.Schedule.Custom.Recurring.OnMonth = &onMonth
		}

		if !model.CustomSchedule.Recurring.OnMonthDay.IsNull() && !model.CustomSchedule.Recurring.OnMonthDay.IsUnknown() {
			var onMonthDay []float32
			diags.Append(model.CustomSchedule.Recurring.OnMonthDay.ElementsAs(ctx, &onMonthDay, true)...)
			body.Schedule.Custom.Recurring.OnMonthDay = &onMonthDay
		}
	}

	if model.Scope != nil {
		body.Scope = &struct {
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
					Kql: model.Scope.Alerting.Kql.ValueString(),
				},
			},
		}
	}

	return body, diags
}

func (model *MaintenanceWindowModel) fromAPIUpdateResponse(ctx context.Context, data *kbapi.PatchMaintenanceWindowIdResponse) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags = diag.Diagnostics{}
	var response = &ResponseJson{}

	if err := json.Unmarshal(data.Body, response); err != nil {
		diags.AddError(err.Error(), "cannot unmarshal PatchMaintenanceWindowIdResponse")
		return diags
	}

	return model._fromAPIResponse(ctx, *response)
}

/* DELETE */

func (model MaintenanceWindowModel) getMaintenanceWindowIDAndSpaceID() (maintenanceWindowID string, spaceID string) {
	maintenanceWindowID = model.ID.ValueString()
	spaceID = model.SpaceID.ValueString()

	resourceID := model.ID.ValueString()
	maybeCompositeID, _ := clients.CompositeIdFromStr(resourceID)
	if maybeCompositeID != nil {
		maintenanceWindowID = maybeCompositeID.ResourceId
		spaceID = maybeCompositeID.ClusterId
	}

	return
}

/* RESPONSE HANDLER */

func (model *MaintenanceWindowModel) _fromAPIResponse(ctx context.Context, response ResponseJson) diag.Diagnostics {
	var diags = diag.Diagnostics{}

	resourceID := clients.CompositeId{
		ClusterId:  model.SpaceID.ValueString(),
		ResourceId: response.Id,
	}

	model.ID = types.StringValue(resourceID.String())
	model.Title = types.StringValue(response.Title)
	model.Enabled = types.BoolValue(response.Enabled)

	model.CustomSchedule = MaintenanceWindowSchedule{
		Start:    types.StringValue(response.Schedule.Custom.Start),
		Duration: types.StringValue(response.Schedule.Custom.Duration),
		Timezone: types.StringPointerValue(response.Schedule.Custom.Timezone),
	}

	if response.Schedule.Custom.Recurring != nil {
		model.CustomSchedule.Recurring = &MaintenanceWindowScheduleRecurring{
			End:        types.StringPointerValue(response.Schedule.Custom.Recurring.End),
			Every:      types.StringPointerValue(response.Schedule.Custom.Recurring.Every),
			OnWeekDay:  types.ListNull(types.StringType),
			OnMonth:    types.ListNull(types.Int32Type),
			OnMonthDay: types.ListNull(types.Int32Type),
		}

		if response.Schedule.Custom.Recurring.Occurrences != nil {
			occurrences := int32(*response.Schedule.Custom.Recurring.Occurrences)
			model.CustomSchedule.Recurring.Occurrences = types.Int32PointerValue(&occurrences)
		}

		if response.Schedule.Custom.Recurring.OnWeekDay != nil {
			onWeekDay, d := types.ListValueFrom(ctx, types.StringType, response.Schedule.Custom.Recurring.OnWeekDay)

			if d.HasError() {
				diags.Append(d...)
			} else {
				model.CustomSchedule.Recurring.OnWeekDay = onWeekDay
			}
		}

		if response.Schedule.Custom.Recurring.OnMonth != nil {
			onMonth, d := types.ListValueFrom(ctx, types.Int32Type, response.Schedule.Custom.Recurring.OnMonth)

			if d.HasError() {
				diags.Append(d...)
			} else {
				model.CustomSchedule.Recurring.OnMonth = onMonth
			}
		}

		if response.Schedule.Custom.Recurring.OnMonthDay != nil {
			onMonthDay, d := types.ListValueFrom(ctx, types.Int32Type, response.Schedule.Custom.Recurring.OnMonthDay)

			if d.HasError() {
				diags.Append(d...)
			} else {
				model.CustomSchedule.Recurring.OnMonthDay = onMonthDay
			}
		}
	}

	if response.Scope != nil {
		model.Scope = &MaintenanceWindowScope{
			Alerting: MaintenanceWindowAlertingScope{
				Kql: types.StringValue(response.Scope.Alerting.Query.Kql),
			},
		}
	}

	return diags
}
