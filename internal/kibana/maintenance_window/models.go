package maintenancewindow

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	ID             types.String `tfsdk:"id"`
	SpaceID        types.String `tfsdk:"space_id"`
	Title          types.String `tfsdk:"title"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	CustomSchedule Schedule     `tfsdk:"custom_schedule"`
	Scope          *Scope       `tfsdk:"scope"`
}

type Scope struct {
	Alerting AlertingScope `tfsdk:"alerting"`
}

type AlertingScope struct {
	Kql types.String `tfsdk:"kql"`
}

type Schedule struct {
	Start     types.String       `tfsdk:"start"`
	Duration  types.String       `tfsdk:"duration"`
	Timezone  types.String       `tfsdk:"timezone"`
	Recurring *ScheduleRecurring `tfsdk:"recurring"`
}

type ScheduleRecurring struct {
	End         types.String `tfsdk:"end"`
	Every       types.String `tfsdk:"every"`
	Occurrences types.Int32  `tfsdk:"occurrences"`
	OnWeekDay   types.List   `tfsdk:"on_week_day"`
	OnMonthDay  types.List   `tfsdk:"on_month_day"`
	OnMonth     types.List   `tfsdk:"on_month"`
}

/* CREATE */

func (model Model) toAPICreateRequest(ctx context.Context) (kbapi.PostMaintenanceWindowJSONRequestBody, diag.Diagnostics) {
	body := kbapi.PostMaintenanceWindowJSONRequestBody{
		Enabled: model.Enabled.ValueBoolPointer(),
		Title:   model.Title.ValueString(),
	}

	body.Schedule.Custom.Duration = model.CustomSchedule.Duration.ValueString()
	body.Schedule.Custom.Start = model.CustomSchedule.Start.ValueString()

	if !model.CustomSchedule.Timezone.IsNull() && !model.CustomSchedule.Timezone.IsUnknown() {
		body.Schedule.Custom.Timezone = model.CustomSchedule.Timezone.ValueStringPointer()
	}

	customRecurring, diags := model.CustomSchedule.Recurring.toAPIRequest(ctx)
	body.Schedule.Custom.Recurring = customRecurring
	body.Scope = model.Scope.toAPIRequest()

	return body, diags
}

/* READ */

func (model *Model) fromAPIReadResponse(ctx context.Context, data *kbapi.GetMaintenanceWindowIdResponse) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags = diag.Diagnostics{}
	var response = &ResponseJSON{}

	if err := json.Unmarshal(data.Body, response); err != nil {
		diags.AddError(err.Error(), "cannot unmarshal GetMaintenanceWindowIdResponse")
		return diags
	}

	return model._fromAPIResponse(ctx, *response)
}

/* UPDATE */

func (model Model) toAPIUpdateRequest(ctx context.Context) (kbapi.PatchMaintenanceWindowIdJSONRequestBody, diag.Diagnostics) {
	body := kbapi.PatchMaintenanceWindowIdJSONRequestBody{
		Enabled: model.Enabled.ValueBoolPointer(),
		Title:   model.Title.ValueStringPointer(),
	}

	body.Schedule = &struct {
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

	if typeutils.IsKnown(model.CustomSchedule.Timezone) {
		body.Schedule.Custom.Timezone = model.CustomSchedule.Timezone.ValueStringPointer()
	}

	customRecurring, diags := model.CustomSchedule.Recurring.toAPIRequest(ctx)
	body.Schedule.Custom.Recurring = customRecurring
	body.Scope = model.Scope.toAPIRequest()

	return body, diags
}

/* DELETE */

func (model Model) getMaintenanceWindowIDAndSpaceID() (maintenanceWindowID string, spaceID string) {
	maintenanceWindowID = model.ID.ValueString()
	spaceID = model.SpaceID.ValueString()

	resourceID := model.ID.ValueString()
	maybeCompositeID, _ := clients.CompositeIDFromStr(resourceID)
	if maybeCompositeID != nil {
		maintenanceWindowID = maybeCompositeID.ResourceID
		spaceID = maybeCompositeID.ClusterID
	}

	return
}

/* RESPONSE HANDLER */

func (model *Model) _fromAPIResponse(ctx context.Context, response ResponseJSON) diag.Diagnostics {
	var diags = diag.Diagnostics{}

	model.Title = types.StringValue(response.Title)
	model.Enabled = types.BoolValue(response.Enabled)

	model.CustomSchedule = Schedule{
		Start:    types.StringValue(response.Schedule.Custom.Start),
		Duration: types.StringValue(response.Schedule.Custom.Duration),
		Timezone: types.StringPointerValue(response.Schedule.Custom.Timezone),
		Recurring: &ScheduleRecurring{
			End:        types.StringNull(),
			Every:      types.StringNull(),
			OnWeekDay:  types.ListNull(types.StringType),
			OnMonth:    types.ListNull(types.Int32Type),
			OnMonthDay: types.ListNull(types.Int32Type),
		},
	}

	if response.Schedule.Custom.Recurring != nil {
		model.CustomSchedule.Recurring.End = types.StringPointerValue(response.Schedule.Custom.Recurring.End)
		model.CustomSchedule.Recurring.Every = types.StringPointerValue(response.Schedule.Custom.Recurring.Every)

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
		model.Scope = &Scope{
			Alerting: AlertingScope{
				Kql: types.StringValue(response.Scope.Alerting.Query.Kql),
			},
		}
	}

	return diags
}

/* HELPERS */

func (model *Scope) toAPIRequest() *struct {
	Alerting struct {
		Query struct {
			Kql string `json:"kql"`
		} `json:"query"`
	} `json:"alerting"`
} {
	if model == nil {
		return nil
	}

	return &struct {
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
				Kql: model.Alerting.Kql.ValueString(),
			},
		},
	}
}

func (model *ScheduleRecurring) toAPIRequest(ctx context.Context) (*struct {
	End         *string    `json:"end,omitempty"`
	Every       *string    `json:"every,omitempty"`
	Occurrences *float32   `json:"occurrences,omitempty"`
	OnMonth     *[]float32 `json:"onMonth,omitempty"`
	OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
	OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
}, diag.Diagnostics) {
	if model == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	result := &struct {
		End         *string    `json:"end,omitempty"`
		Every       *string    `json:"every,omitempty"`
		Occurrences *float32   `json:"occurrences,omitempty"`
		OnMonth     *[]float32 `json:"onMonth,omitempty"`
		OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
		OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
	}{}

	if typeutils.IsKnown(model.End) {
		result.End = model.End.ValueStringPointer()
	}

	if typeutils.IsKnown(model.Every) {
		result.Every = model.Every.ValueStringPointer()
	}

	if typeutils.IsKnown(model.Occurrences) {
		occurrences := float32(model.Occurrences.ValueInt32())
		result.Occurrences = &occurrences
	}

	if typeutils.IsKnown(model.OnWeekDay) {
		var onWeekDay []string
		diags.Append(model.OnWeekDay.ElementsAs(ctx, &onWeekDay, true)...)
		result.OnWeekDay = &onWeekDay
	}

	if typeutils.IsKnown(model.OnMonth) {
		var onMonth []float32
		diags.Append(model.OnMonth.ElementsAs(ctx, &onMonth, true)...)
		result.OnMonth = &onMonth
	}

	if typeutils.IsKnown(model.OnMonthDay) {
		var onMonthDay []float32
		diags.Append(model.OnMonthDay.ElementsAs(ctx, &onMonthDay, true)...)
		result.OnMonthDay = &onMonthDay
	}

	return result, diags
}
