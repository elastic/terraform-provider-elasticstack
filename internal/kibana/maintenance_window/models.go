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
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	entitycore.KibanaConnectionField
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

/* INTERFACE METHODS */

func (m Model) GetID() types.String         { return m.ID }
func (m Model) GetResourceID() types.String { return m.ID }
func (m Model) GetSpaceID() types.String    { return m.SpaceID }

var maintenanceWindowMinVersion = version.Must(version.NewVersion("9.1.0"))

// GetVersionRequirements returns the minimum Kibana version required for
// maintenance windows. This satisfies the optional
// entitycore.WithVersionRequirements interface, allowing the
// generic Kibana resource envelope to enforce the requirement before invoking
// lifecycle callbacks.
func (m Model) GetVersionRequirements() ([]entitycore.DataSourceVersionRequirement, diag.Diagnostics) {
	return []entitycore.DataSourceVersionRequirement{
		{
			MinVersion:   *maintenanceWindowMinVersion,
			ErrorMessage: fmt.Sprintf("Maintenance windows require Elastic Stack v%s or later.", maintenanceWindowMinVersion),
		},
	}, nil
}

/* CREATE */

func (m Model) toAPICreateRequest(ctx context.Context) (kbapi.PostMaintenanceWindowJSONRequestBody, diag.Diagnostics) {
	body := kbapi.PostMaintenanceWindowJSONRequestBody{
		Enabled: m.Enabled.ValueBoolPointer(),
		Title:   m.Title.ValueString(),
	}

	body.Schedule.Custom.Duration = m.CustomSchedule.Duration.ValueString()
	body.Schedule.Custom.Start = m.CustomSchedule.Start.ValueString()

	if !m.CustomSchedule.Timezone.IsNull() && !m.CustomSchedule.Timezone.IsUnknown() {
		body.Schedule.Custom.Timezone = m.CustomSchedule.Timezone.ValueStringPointer()
	}

	customRecurring, diags := m.CustomSchedule.Recurring.toAPIRequest(ctx)
	body.Schedule.Custom.Recurring = customRecurring
	body.Scope = m.Scope.toAPIRequest()

	return body, diags
}

/* READ */

func (m *Model) fromAPIReadResponse(ctx context.Context, data *kbapi.GetMaintenanceWindowIdResponse) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags = diag.Diagnostics{}
	var response = &ResponseJSON{}

	if err := json.Unmarshal(data.Body, response); err != nil {
		diags.AddError(err.Error(), "cannot unmarshal GetMaintenanceWindowIdResponse")
		return diags
	}

	return m._fromAPIResponse(ctx, *response)
}

/* UPDATE */

func (m Model) toAPIUpdateRequest(ctx context.Context) (kbapi.PatchMaintenanceWindowIdJSONRequestBody, diag.Diagnostics) {
	body := kbapi.PatchMaintenanceWindowIdJSONRequestBody{
		Enabled: m.Enabled.ValueBoolPointer(),
		Title:   m.Title.ValueStringPointer(),
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
			Duration: m.CustomSchedule.Duration.ValueString(),
			Start:    m.CustomSchedule.Start.ValueString(),
		},
	}

	if typeutils.IsKnown(m.CustomSchedule.Timezone) {
		body.Schedule.Custom.Timezone = m.CustomSchedule.Timezone.ValueStringPointer()
	}

	customRecurring, diags := m.CustomSchedule.Recurring.toAPIRequest(ctx)
	body.Schedule.Custom.Recurring = customRecurring
	body.Scope = m.Scope.toAPIRequest()

	return body, diags
}

/* RESPONSE HANDLER */

func (m *Model) _fromAPIResponse(ctx context.Context, response ResponseJSON) diag.Diagnostics {
	var diags = diag.Diagnostics{}

	m.Title = types.StringValue(response.Title)
	m.Enabled = types.BoolValue(response.Enabled)

	m.CustomSchedule = Schedule{
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
		m.CustomSchedule.Recurring.End = types.StringPointerValue(response.Schedule.Custom.Recurring.End)
		m.CustomSchedule.Recurring.Every = types.StringPointerValue(response.Schedule.Custom.Recurring.Every)

		if response.Schedule.Custom.Recurring.Occurrences != nil {
			occurrences := int32(*response.Schedule.Custom.Recurring.Occurrences)
			m.CustomSchedule.Recurring.Occurrences = types.Int32PointerValue(&occurrences)
		}

		if response.Schedule.Custom.Recurring.OnWeekDay != nil {
			onWeekDay, d := types.ListValueFrom(ctx, types.StringType, response.Schedule.Custom.Recurring.OnWeekDay)

			if d.HasError() {
				diags.Append(d...)
			} else {
				m.CustomSchedule.Recurring.OnWeekDay = onWeekDay
			}
		}

		if response.Schedule.Custom.Recurring.OnMonth != nil {
			onMonth, d := types.ListValueFrom(ctx, types.Int32Type, response.Schedule.Custom.Recurring.OnMonth)

			if d.HasError() {
				diags.Append(d...)
			} else {
				m.CustomSchedule.Recurring.OnMonth = onMonth
			}
		}

		if response.Schedule.Custom.Recurring.OnMonthDay != nil {
			onMonthDay, d := types.ListValueFrom(ctx, types.Int32Type, response.Schedule.Custom.Recurring.OnMonthDay)

			if d.HasError() {
				diags.Append(d...)
			} else {
				m.CustomSchedule.Recurring.OnMonthDay = onMonthDay
			}
		}
	}

	if response.Scope != nil {
		m.Scope = &Scope{
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
