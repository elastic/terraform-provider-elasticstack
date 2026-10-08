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

package alertingrules

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const lastExecutionDateLayout = "2006-01-02T15:04:05.000Z"

// alertingRulesSearchResourceID is the envelope identity used when the data
// source searches rather than looking up one rule. The read path must never
// send this value to the single-rule fetch.
const alertingRulesSearchResourceID = "alerting_rules"

type alertingRulesDataSourceModel struct {
	entitycore.KibanaConnectionField
	ID      types.String `tfsdk:"id"`
	SpaceID types.String `tfsdk:"space_id"`
	RuleID  types.String `tfsdk:"rule_id"`
	Filter  types.String `tfsdk:"filter"`
	Rules   types.List   `tfsdk:"rules"`
}

func (m alertingRulesDataSourceModel) GetID() types.String { return m.ID }

func (m alertingRulesDataSourceModel) GetSpaceID() types.String { return m.SpaceID }

func (m alertingRulesDataSourceModel) GetResourceID() types.String {
	if ruleIDConfigured(m.RuleID) {
		return m.RuleID
	}
	return types.StringValue(alertingRulesSearchResourceID)
}

func ruleIDConfigured(ruleID types.String) bool {
	return typeutils.IsKnown(ruleID) && ruleID.ValueString() != ""
}

type ruleElementModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	RuleTypeID          types.String `tfsdk:"rule_type_id"`
	Consumer            types.String `tfsdk:"consumer"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Tags                types.Set    `tfsdk:"tags"`
	ScheduledTaskID     types.String `tfsdk:"scheduled_task_id"`
	LastExecutionStatus types.String `tfsdk:"last_execution_status"`
	LastExecutionDate   types.String `tfsdk:"last_execution_date"`
}

func ruleElementFromAPI(rule models.AlertingRule) ruleElementModel {
	tags := make([]attr.Value, 0, len(rule.Tags))
	for _, tag := range rule.Tags {
		tags = append(tags, types.StringValue(tag))
	}
	element := ruleElementModel{
		ID:                  types.StringValue(rule.RuleID),
		Name:                types.StringValue(rule.Name),
		RuleTypeID:          types.StringValue(rule.RuleTypeID),
		Consumer:            types.StringValue(rule.Consumer),
		Enabled:             types.BoolPointerValue(rule.Enabled),
		Tags:                types.SetValueMust(types.StringType, tags),
		ScheduledTaskID:     types.StringPointerValue(rule.ScheduledTaskID),
		LastExecutionStatus: types.StringPointerValue(rule.ExecutionStatus.Status),
		LastExecutionDate:   types.StringNull(),
	}
	if rule.ExecutionStatus.LastExecutionDate != nil {
		element.LastExecutionDate = types.StringValue(rule.ExecutionStatus.LastExecutionDate.UTC().Format(lastExecutionDateLayout))
	}
	return element
}

func ruleElementAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                    types.StringType,
		"name":                  types.StringType,
		"rule_type_id":          types.StringType,
		"consumer":              types.StringType,
		"enabled":               types.BoolType,
		"tags":                  types.SetType{ElemType: types.StringType},
		"scheduled_task_id":     types.StringType,
		"last_execution_status": types.StringType,
		"last_execution_date":   types.StringType,
	}
}

func (m *alertingRulesDataSourceModel) setRules(ctx context.Context, rules []models.AlertingRule) diag.Diagnostics {
	elems := make([]attr.Value, 0, len(rules))
	for _, rule := range rules {
		obj, diags := types.ObjectValueFrom(ctx, ruleElementAttrTypes(), ruleElementFromAPI(rule))
		if diags.HasError() {
			return diags
		}
		elems = append(elems, obj)
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: ruleElementAttrTypes()}, elems)
	if diags.HasError() {
		return diags
	}
	m.Rules = list
	return nil
}
