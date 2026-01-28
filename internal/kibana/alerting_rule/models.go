package alerting_rule

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// alertingRuleModel is the Terraform model for the alerting rule resource.
type alertingRuleModel struct {
	ID                  types.String  `tfsdk:"id"`
	RuleID              types.String  `tfsdk:"rule_id"`
	SpaceID             types.String  `tfsdk:"space_id"`
	Name                types.String  `tfsdk:"name"`
	Consumer            types.String  `tfsdk:"consumer"`
	NotifyWhen          types.String  `tfsdk:"notify_when"`
	Params              types.String  `tfsdk:"params"`
	RuleTypeID          types.String  `tfsdk:"rule_type_id"`
	Interval            types.String  `tfsdk:"interval"`
	Enabled             types.Bool    `tfsdk:"enabled"`
	Tags                types.List    `tfsdk:"tags"`
	Throttle            types.String  `tfsdk:"throttle"`
	ScheduledTaskID     types.String  `tfsdk:"scheduled_task_id"`
	LastExecutionStatus types.String  `tfsdk:"last_execution_status"`
	LastExecutionDate   types.String  `tfsdk:"last_execution_date"`
	AlertDelay          types.Float64 `tfsdk:"alert_delay"`
	Actions             []actionModel `tfsdk:"actions"`
}

type actionModel struct {
	Group        types.String       `tfsdk:"group"`
	ID           types.String       `tfsdk:"id"`
	Params       types.String       `tfsdk:"params"`
	Frequency    *frequencyModel    `tfsdk:"frequency"`
	AlertsFilter *alertsFilterModel `tfsdk:"alerts_filter"`
}

type frequencyModel struct {
	Summary    types.Bool   `tfsdk:"summary"`
	NotifyWhen types.String `tfsdk:"notify_when"`
	Throttle   types.String `tfsdk:"throttle"`
}

type alertsFilterModel struct {
	Kql       types.String    `tfsdk:"kql"`
	Timeframe *timeframeModel `tfsdk:"timeframe"`
}

type timeframeModel struct {
	Days       types.List   `tfsdk:"days"`
	Timezone   types.String `tfsdk:"timezone"`
	HoursStart types.String `tfsdk:"hours_start"`
	HoursEnd   types.String `tfsdk:"hours_end"`
}

// JSON helper types for action marshaling
type actionJSON struct {
	Group        *string                 `json:"group,omitempty"`
	ID           string                  `json:"id"`
	Params       *map[string]interface{} `json:"params,omitempty"`
	Frequency    *frequencyJSON          `json:"frequency,omitempty"`
	AlertsFilter *alertsFilterJSON       `json:"alerts_filter,omitempty"`
}

type frequencyJSON struct {
	NotifyWhen string  `json:"notify_when"`
	Summary    bool    `json:"summary"`
	Throttle   *string `json:"throttle,omitempty"`
}

type alertsFilterJSON struct {
	Query     *queryJSON     `json:"query,omitempty"`
	Timeframe *timeframeJSON `json:"timeframe,omitempty"`
}

type queryJSON struct {
	Kql string `json:"kql"`
}

type timeframeJSON struct {
	Days     []int     `json:"days"`
	Hours    hoursJSON `json:"hours"`
	Timezone string    `json:"timezone"`
}

type hoursJSON struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// toAPICreateRequest converts the Terraform model to a kbapi create request.
func (m *alertingRuleModel) toAPICreateRequest(ctx context.Context) (kbapi.AlertingRuleCreateRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Parse params JSON
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(m.Params.ValueString()), &params); err != nil {
		diags.AddError("Invalid params JSON", err.Error())
		return kbapi.AlertingRuleCreateRequest{}, diags
	}

	req := kbapi.AlertingRuleCreateRequest{
		Name:       m.Name.ValueString(),
		Consumer:   m.Consumer.ValueString(),
		RuleTypeId: m.RuleTypeID.ValueString(),
		Schedule: struct {
			Interval string `json:"interval"`
		}{
			Interval: m.Interval.ValueString(),
		},
	}

	// Set params using the union approach - marshal params to JSON and set as raw
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		diags.AddError("Failed to marshal params", err.Error())
		return kbapi.AlertingRuleCreateRequest{}, diags
	}
	paramsWrapper := &kbapi.AlertingRuleCreateRequest_Params{}
	if err := paramsWrapper.UnmarshalJSON(paramsJSON); err != nil {
		diags.AddError("Failed to set params", err.Error())
		return kbapi.AlertingRuleCreateRequest{}, diags
	}
	req.Params = paramsWrapper

	// Optional fields
	if !m.NotifyWhen.IsNull() && !m.NotifyWhen.IsUnknown() {
		notifyWhen := kbapi.AlertingRuleCreateRequestNotifyWhen(m.NotifyWhen.ValueString())
		req.NotifyWhen = &notifyWhen
	}

	if !m.Enabled.IsNull() && !m.Enabled.IsUnknown() {
		enabled := m.Enabled.ValueBool()
		req.Enabled = &enabled
	}

	if !m.Throttle.IsNull() && !m.Throttle.IsUnknown() {
		throttle := m.Throttle.ValueString()
		req.Throttle = &throttle
	}

	if !m.AlertDelay.IsNull() && !m.AlertDelay.IsUnknown() {
		req.AlertDelay = &struct {
			Active float32 `json:"active"`
		}{
			Active: float32(m.AlertDelay.ValueFloat64()),
		}
	}

	// Tags
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		var tags []string
		diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)
		if diags.HasError() {
			return kbapi.AlertingRuleCreateRequest{}, diags
		}
		req.Tags = &tags
	}

	// Actions - build using JSON marshaling for complex nested types
	if len(m.Actions) > 0 {
		actionsJSON, actionDiags := m.actionsToJSON(ctx)
		diags.Append(actionDiags...)
		if diags.HasError() {
			return kbapi.AlertingRuleCreateRequest{}, diags
		}

		// Unmarshal into the correct type
		var actions []struct {
			AlertsFilter *struct {
				Query *struct {
					Dsl     *string `json:"dsl,omitempty"`
					Filters []struct {
						State *struct {
							Store kbapi.AlertingRuleCreateRequestActionsAlertsFilterQueryFiltersStateStore `json:"store"`
						} `json:"$state,omitempty"`
						Meta  map[string]interface{}  `json:"meta"`
						Query *map[string]interface{} `json:"query,omitempty"`
					} `json:"filters"`
					Kql string `json:"kql"`
				} `json:"query,omitempty"`
				Timeframe *struct {
					Days  []kbapi.AlertingRuleCreateRequestActionsAlertsFilterTimeframeDays `json:"days"`
					Hours struct {
						End   string `json:"end"`
						Start string `json:"start"`
					} `json:"hours"`
					Timezone string `json:"timezone"`
				} `json:"timeframe,omitempty"`
			} `json:"alerts_filter,omitempty"`
			Frequency *struct {
				NotifyWhen kbapi.AlertingRuleCreateRequestActionsFrequencyNotifyWhen `json:"notify_when"`
				Summary    bool                                                      `json:"summary"`
				Throttle   *string                                                   `json:"throttle"`
			} `json:"frequency,omitempty"`
			Group                   *string                 `json:"group,omitempty"`
			Id                      string                  `json:"id"`
			Params                  *map[string]interface{} `json:"params,omitempty"`
			UseAlertDataForTemplate *bool                   `json:"use_alert_data_for_template,omitempty"`
			Uuid                    *string                 `json:"uuid,omitempty"`
		}

		if err := json.Unmarshal(actionsJSON, &actions); err != nil {
			diags.AddError("Failed to parse actions", err.Error())
			return kbapi.AlertingRuleCreateRequest{}, diags
		}
		req.Actions = &actions
	}

	return req, diags
}

// toAPIUpdateRequest converts the Terraform model to a kbapi update request.
func (m *alertingRuleModel) toAPIUpdateRequest(ctx context.Context) (kbapi.AlertingRuleUpdateRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Parse params JSON
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(m.Params.ValueString()), &params); err != nil {
		diags.AddError("Invalid params JSON", err.Error())
		return kbapi.AlertingRuleUpdateRequest{}, diags
	}

	req := kbapi.AlertingRuleUpdateRequest{
		Name: m.Name.ValueString(),
		Schedule: struct {
			Interval string `json:"interval"`
		}{
			Interval: m.Interval.ValueString(),
		},
		Params: &params,
	}

	// Optional fields
	if !m.NotifyWhen.IsNull() && !m.NotifyWhen.IsUnknown() {
		notifyWhen := kbapi.AlertingRuleUpdateRequestNotifyWhen(m.NotifyWhen.ValueString())
		req.NotifyWhen = &notifyWhen
	}

	if !m.Throttle.IsNull() && !m.Throttle.IsUnknown() {
		throttle := m.Throttle.ValueString()
		req.Throttle = &throttle
	}

	if !m.AlertDelay.IsNull() && !m.AlertDelay.IsUnknown() {
		req.AlertDelay = &struct {
			Active float32 `json:"active"`
		}{
			Active: float32(m.AlertDelay.ValueFloat64()),
		}
	}

	// Tags
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		var tags []string
		diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)
		if diags.HasError() {
			return kbapi.AlertingRuleUpdateRequest{}, diags
		}
		req.Tags = &tags
	}

	// Actions
	if len(m.Actions) > 0 {
		actionsJSON, actionDiags := m.actionsToJSON(ctx)
		diags.Append(actionDiags...)
		if diags.HasError() {
			return kbapi.AlertingRuleUpdateRequest{}, diags
		}

		var actions []struct {
			AlertsFilter *struct {
				Query *struct {
					Dsl     *string `json:"dsl,omitempty"`
					Filters []struct {
						State *struct {
							Store kbapi.AlertingRuleUpdateRequestActionsAlertsFilterQueryFiltersStateStore `json:"store"`
						} `json:"$state,omitempty"`
						Meta  map[string]interface{}  `json:"meta"`
						Query *map[string]interface{} `json:"query,omitempty"`
					} `json:"filters"`
					Kql string `json:"kql"`
				} `json:"query,omitempty"`
				Timeframe *struct {
					Days  []kbapi.AlertingRuleUpdateRequestActionsAlertsFilterTimeframeDays `json:"days"`
					Hours struct {
						End   string `json:"end"`
						Start string `json:"start"`
					} `json:"hours"`
					Timezone string `json:"timezone"`
				} `json:"timeframe,omitempty"`
			} `json:"alerts_filter,omitempty"`
			Frequency *struct {
				NotifyWhen kbapi.AlertingRuleUpdateRequestActionsFrequencyNotifyWhen `json:"notify_when"`
				Summary    bool                                                      `json:"summary"`
				Throttle   *string                                                   `json:"throttle"`
			} `json:"frequency,omitempty"`
			Group                   *string                 `json:"group,omitempty"`
			Id                      string                  `json:"id"`
			Params                  *map[string]interface{} `json:"params,omitempty"`
			UseAlertDataForTemplate *bool                   `json:"use_alert_data_for_template,omitempty"`
			Uuid                    *string                 `json:"uuid,omitempty"`
		}

		if err := json.Unmarshal(actionsJSON, &actions); err != nil {
			diags.AddError("Failed to parse actions", err.Error())
			return kbapi.AlertingRuleUpdateRequest{}, diags
		}
		req.Actions = &actions
	}

	return req, diags
}

// actionsToJSON converts actions to JSON bytes for marshaling into the complex inline types.
func (m *alertingRuleModel) actionsToJSON(ctx context.Context) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	actions := make([]actionJSON, 0, len(m.Actions))

	for _, a := range m.Actions {
		action := actionJSON{
			ID: a.ID.ValueString(),
		}

		if !a.Group.IsNull() && !a.Group.IsUnknown() {
			group := a.Group.ValueString()
			action.Group = &group
		}

		// Parse action params
		if !a.Params.IsNull() && !a.Params.IsUnknown() {
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(a.Params.ValueString()), &params); err != nil {
				diags.AddError("Invalid action params JSON", err.Error())
				return nil, diags
			}
			action.Params = &params
		}

		// Frequency
		if a.Frequency != nil {
			action.Frequency = &frequencyJSON{
				NotifyWhen: a.Frequency.NotifyWhen.ValueString(),
				Summary:    a.Frequency.Summary.ValueBool(),
			}
			if !a.Frequency.Throttle.IsNull() && !a.Frequency.Throttle.IsUnknown() {
				throttle := a.Frequency.Throttle.ValueString()
				action.Frequency.Throttle = &throttle
			}
		}

		// Alerts filter
		if a.AlertsFilter != nil {
			action.AlertsFilter = &alertsFilterJSON{}

			if !a.AlertsFilter.Kql.IsNull() && !a.AlertsFilter.Kql.IsUnknown() {
				action.AlertsFilter.Query = &queryJSON{
					Kql: a.AlertsFilter.Kql.ValueString(),
				}
			}

			if a.AlertsFilter.Timeframe != nil {
				var days []int64
				diags.Append(a.AlertsFilter.Timeframe.Days.ElementsAs(ctx, &days, false)...)
				if diags.HasError() {
					return nil, diags
				}

				intDays := make([]int, 0, len(days))
				for _, d := range days {
					intDays = append(intDays, int(d))
				}

				action.AlertsFilter.Timeframe = &timeframeJSON{
					Days:     intDays,
					Timezone: a.AlertsFilter.Timeframe.Timezone.ValueString(),
					Hours: hoursJSON{
						Start: a.AlertsFilter.Timeframe.HoursStart.ValueString(),
						End:   a.AlertsFilter.Timeframe.HoursEnd.ValueString(),
					},
				}
			}
		}

		actions = append(actions, action)
	}

	result, err := json.Marshal(actions)
	if err != nil {
		diags.AddError("Failed to marshal actions", err.Error())
		return nil, diags
	}
	return result, diags
}

// populateFromAPI populates the Terraform model from a kbapi response.
func (m *alertingRuleModel) populateFromAPI(ctx context.Context, resp *kbapi.AlertingRuleResponse, spaceID string) diag.Diagnostics {
	var diags diag.Diagnostics

	m.RuleID = types.StringValue(resp.Id)
	m.SpaceID = types.StringValue(spaceID)
	m.ID = types.StringValue((&clients.CompositeId{ClusterId: spaceID, ResourceId: resp.Id}).String())
	m.Name = types.StringValue(resp.Name)
	m.Consumer = types.StringValue(resp.Consumer)
	m.RuleTypeID = types.StringValue(resp.RuleTypeId)
	m.Interval = types.StringValue(resp.Schedule.Interval)
	m.Enabled = types.BoolValue(resp.Enabled)

	if resp.NotifyWhen != nil {
		m.NotifyWhen = types.StringValue(string(*resp.NotifyWhen))
	} else {
		m.NotifyWhen = types.StringNull()
	}

	if resp.Throttle != nil {
		m.Throttle = types.StringValue(*resp.Throttle)
	} else {
		m.Throttle = types.StringNull()
	}

	if resp.ScheduledTaskId != nil {
		m.ScheduledTaskID = types.StringValue(*resp.ScheduledTaskId)
	} else {
		m.ScheduledTaskID = types.StringNull()
	}

	m.LastExecutionStatus = types.StringValue(string(resp.ExecutionStatus.Status))

	if resp.ExecutionStatus.LastExecutionDate != "" {
		m.LastExecutionDate = types.StringValue(resp.ExecutionStatus.LastExecutionDate)
	} else {
		m.LastExecutionDate = types.StringNull()
	}

	if resp.AlertDelay != nil {
		m.AlertDelay = types.Float64Value(float64(resp.AlertDelay.Active))
	} else {
		m.AlertDelay = types.Float64Null()
	}

	// Params - convert to JSON string
	paramsJSON, err := json.Marshal(resp.Params)
	if err != nil {
		diags.AddError("Failed to marshal params", err.Error())
		return diags
	}
	m.Params = types.StringValue(string(paramsJSON))

	// Tags
	if len(resp.Tags) > 0 {
		tags, tagDiags := types.ListValueFrom(ctx, types.StringType, resp.Tags)
		diags.Append(tagDiags...)
		m.Tags = tags
	} else {
		m.Tags = types.ListNull(types.StringType)
	}

	// Actions
	m.Actions = make([]actionModel, 0, len(resp.Actions))
	for _, action := range resp.Actions {
		actionM, actionDiags := actionFromAPI(ctx, action)
		diags.Append(actionDiags...)
		m.Actions = append(m.Actions, actionM)
	}

	return diags
}

func actionFromAPI(ctx context.Context, action kbapi.AlertingRuleAction) (actionModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	am := actionModel{
		ID: types.StringValue(action.Id),
	}

	// Group is a pointer
	if action.Group != nil {
		am.Group = types.StringValue(*action.Group)
	} else {
		am.Group = types.StringValue("default")
	}

	// Params - always present as map[string]interface{}
	if len(action.Params) > 0 {
		paramsJSON, err := json.Marshal(action.Params)
		if err != nil {
			diags.AddError("Failed to marshal action params", err.Error())
			return am, diags
		}
		am.Params = types.StringValue(string(paramsJSON))
	} else {
		am.Params = types.StringValue("{}")
	}

	// Frequency
	if action.Frequency != nil {
		am.Frequency = &frequencyModel{
			Summary:    types.BoolValue(action.Frequency.Summary),
			NotifyWhen: types.StringValue(string(action.Frequency.NotifyWhen)),
		}
		if action.Frequency.Throttle != nil {
			am.Frequency.Throttle = types.StringValue(*action.Frequency.Throttle)
		} else {
			am.Frequency.Throttle = types.StringNull()
		}
	}

	// Alerts filter
	if action.AlertsFilter != nil {
		am.AlertsFilter = &alertsFilterModel{}

		if action.AlertsFilter.Query != nil && action.AlertsFilter.Query.Kql != "" {
			am.AlertsFilter.Kql = types.StringValue(action.AlertsFilter.Query.Kql)
		} else {
			am.AlertsFilter.Kql = types.StringNull()
		}

		if action.AlertsFilter.Timeframe != nil {
			tf := action.AlertsFilter.Timeframe
			days := make([]int64, 0, len(tf.Days))
			for _, d := range tf.Days {
				days = append(days, int64(d))
			}
			daysList, daysDiags := types.ListValueFrom(ctx, types.Int64Type, days)
			diags.Append(daysDiags...)

			am.AlertsFilter.Timeframe = &timeframeModel{
				Days:       daysList,
				Timezone:   types.StringValue(tf.Timezone),
				HoursStart: types.StringValue(tf.Hours.Start),
				HoursEnd:   types.StringValue(tf.Hours.End),
			}
		}
	}

	return am, diags
}
