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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kibanacustomtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	entitycore.ResourceTimeoutsField
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
	Start     types.String                       `tfsdk:"start"`
	Duration  kibanacustomtypes.AlertingDuration `tfsdk:"duration"`
	Timezone  types.String                       `tfsdk:"timezone"`
	Recurring *ScheduleRecurring                 `tfsdk:"recurring"`
}

type ScheduleRecurring struct {
	End         types.String                       `tfsdk:"end"`
	Every       kibanacustomtypes.AlertingDuration `tfsdk:"every"`
	Occurrences types.Int32                        `tfsdk:"occurrences"`
	OnWeekDay   types.List                         `tfsdk:"on_week_day"`
	OnMonthDay  types.List                         `tfsdk:"on_month_day"`
	OnMonth     types.List                         `tfsdk:"on_month"`
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
func (m Model) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
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

func (m *Model) populateFromAPI(ctx context.Context, data *kbapi.GetMaintenanceWindowIdResponse) diag.Diagnostics {
	if data == nil || data.JSON200 == nil {
		return nil
	}

	return m._fromAPIResponse(ctx, *data.JSON200)
}

/* UPDATE */

func (m Model) toAPIUpdateRequest(ctx context.Context) (kbapi.PatchMaintenanceWindowIdJSONRequestBody, diag.Diagnostics) {
	body := kbapi.PatchMaintenanceWindowIdJSONRequestBody{
		Enabled: m.Enabled.ValueBoolPointer(),
		Title:   m.Title.ValueStringPointer(),
	}

	body.Schedule = &struct {
		Custom kbapi.KibanaHTTPAPIsMaintenanceWindowScheduleRequest `json:"custom"`
	}{
		Custom: kbapi.KibanaHTTPAPIsMaintenanceWindowScheduleRequest{
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

func (m *Model) _fromAPIResponse(ctx context.Context, response kbapi.KibanaHTTPAPIsMaintenanceWindowResponse) diag.Diagnostics {
	var diags = diag.Diagnostics{}

	m.Title = types.StringValue(response.Title)
	m.Enabled = types.BoolValue(response.Enabled)

	m.CustomSchedule = Schedule{
		Start:    types.StringValue(response.Schedule.Custom.Start),
		Duration: kibanacustomtypes.NewAlertingDurationValue(response.Schedule.Custom.Duration),
		Timezone: types.StringPointerValue(response.Schedule.Custom.Timezone),
		Recurring: &ScheduleRecurring{
			End:        types.StringNull(),
			Every:      kibanacustomtypes.NewAlertingDurationNull(),
			OnWeekDay:  types.ListNull(types.StringType),
			OnMonth:    types.ListNull(types.Int32Type),
			OnMonthDay: types.ListNull(types.Int32Type),
		},
	}

	if response.Schedule.Custom.Recurring != nil {
		m.CustomSchedule.Recurring.End = types.StringPointerValue(response.Schedule.Custom.Recurring.End)
		m.CustomSchedule.Recurring.Every = kibanacustomtypes.NewAlertingDurationPointerValue(response.Schedule.Custom.Recurring.Every)

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

func (model *Scope) toAPIRequest() *kbapi.KibanaHTTPAPIsMaintenanceWindowScope {
	if model == nil {
		return nil
	}

	return &kbapi.KibanaHTTPAPIsMaintenanceWindowScope{
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

func (model *ScheduleRecurring) toAPIRequest(ctx context.Context) (*kbapi.KibanaHTTPAPIsMaintenanceWindowScheduleRecurringRequest, diag.Diagnostics) {
	if model == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	result := &kbapi.KibanaHTTPAPIsMaintenanceWindowScheduleRecurringRequest{}

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
