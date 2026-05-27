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

package queryrulesets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/queryrulecriteriatype"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/queryruletype"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var attrTypesString = fwtypes.StringType

// QueryRuleActionDocModel represents a document reference in rule actions.
type QueryRuleActionDocModel struct {
	Index fwtypes.String `tfsdk:"_index"`
	ID    fwtypes.String `tfsdk:"_id"`
}

func queryRuleActionDocModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"_index": fwtypes.StringType,
		"_id":    fwtypes.StringType,
	}
}

// QueryRuleActionsModel represents the actions block for a query rule.
type QueryRuleActionsModel struct {
	IDs  fwtypes.List `tfsdk:"ids"`
	Docs fwtypes.List `tfsdk:"docs"`
}

func queryRuleActionsModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ids":  fwtypes.ListType{ElemType: fwtypes.StringType},
		"docs": fwtypes.ListType{ElemType: fwtypes.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}},
	}
}

// QueryRuleCriteriaModel represents a single match criterion for a query rule.
type QueryRuleCriteriaModel struct {
	Type     fwtypes.String       `tfsdk:"type"`
	Metadata fwtypes.String       `tfsdk:"metadata"`
	Values   jsontypes.Normalized `tfsdk:"values"`
}

func queryRuleCriteriaModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":     fwtypes.StringType,
		"metadata": fwtypes.StringType,
		"values":   jsontypes.NormalizedType{},
	}
}

// QueryRuleModel represents a single query rule nested block.
type QueryRuleModel struct {
	RuleID   fwtypes.String `tfsdk:"rule_id"`
	Type     fwtypes.String `tfsdk:"type"`
	Priority fwtypes.Int64  `tfsdk:"priority"`
	Criteria fwtypes.List   `tfsdk:"criteria"`
	Actions  fwtypes.Object `tfsdk:"actions"`
}

func queryRuleModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rule_id":  fwtypes.StringType,
		"type":     fwtypes.StringType,
		"priority": fwtypes.Int64Type,
		"criteria": fwtypes.ListType{ElemType: fwtypes.ObjectType{AttrTypes: queryRuleCriteriaModelAttrTypes()}},
		"actions":  fwtypes.ObjectType{AttrTypes: queryRuleActionsModelAttrTypes()},
	}
}

// QueryRulesetData is the Terraform state model for the query ruleset resource and data source.
type QueryRulesetData struct {
	entitycore.ElasticsearchConnectionField
	ID        fwtypes.String `tfsdk:"id"`
	RulesetID fwtypes.String `tfsdk:"ruleset_id"`
	Rules     fwtypes.List   `tfsdk:"rules"`
}

func (data QueryRulesetData) GetID() fwtypes.String         { return data.ID }
func (data QueryRulesetData) GetResourceID() fwtypes.String { return data.RulesetID }
func (data QueryRulesetData) GetElasticsearchConnection() fwtypes.List {
	return data.ElasticsearchConnection
}

var (
	_ entitycore.ElasticsearchResourceModel = QueryRulesetData{}
	_ entitycore.WithVersionRequirements    = QueryRulesetData{}
)

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements].
func (data QueryRulesetData) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *MinSupportedVersion,
		ErrorMessage: "Elasticsearch query rulesets require Elasticsearch v8.12.0 or above (Query Rules API GA).",
	}}, nil
}

func (data *QueryRulesetData) populateFromAPI(ctx context.Context, rules []types.QueryRule, diagnostics *diag.Diagnostics) {
	models := make([]QueryRuleModel, len(rules))
	for i, rule := range rules {
		models[i] = queryRuleFromAPI(ctx, rule, diagnostics)
		if diagnostics.HasError() {
			return
		}
	}

	list, d := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: queryRuleModelAttrTypes()}, models)
	diagnostics.Append(d...)
	if diagnostics.HasError() {
		return
	}

	data.Rules = list
}

func queryRuleFromAPI(ctx context.Context, rule types.QueryRule, diagnostics *diag.Diagnostics) QueryRuleModel {
	criteriaModels := make([]QueryRuleCriteriaModel, len(rule.Criteria))
	for i, criterion := range rule.Criteria {
		criteriaModels[i] = queryRuleCriteriaFromAPI(criterion, diagnostics)
		if diagnostics.HasError() {
			return QueryRuleModel{}
		}
	}

	criteriaList, d := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: queryRuleCriteriaModelAttrTypes()}, criteriaModels)
	diagnostics.Append(d...)
	if diagnostics.HasError() {
		return QueryRuleModel{}
	}

	actionsObj, d := queryRuleActionsFromAPI(ctx, rule.Actions, diagnostics)
	diagnostics.Append(d...)
	if diagnostics.HasError() {
		return QueryRuleModel{}
	}

	model := QueryRuleModel{
		RuleID:   fwtypes.StringValue(rule.RuleId),
		Type:     fwtypes.StringValue(rule.Type.String()),
		Criteria: criteriaList,
		Actions:  actionsObj,
	}

	if rule.Priority != nil {
		model.Priority = fwtypes.Int64Value(int64(*rule.Priority))
	} else {
		model.Priority = fwtypes.Int64Null()
	}

	return model
}

func queryRuleCriteriaFromAPI(criterion types.QueryRuleCriteria, diagnostics *diag.Diagnostics) QueryRuleCriteriaModel {
	model := QueryRuleCriteriaModel{
		Type: fwtypes.StringValue(criterion.Type.String()),
	}

	if criterion.Metadata != nil {
		model.Metadata = fwtypes.StringValue(*criterion.Metadata)
	} else {
		model.Metadata = fwtypes.StringNull()
	}

	if len(criterion.Values) == 0 {
		model.Values = jsontypes.Normalized{StringValue: fwtypes.StringNull()}
	} else {
		encoded, err := json.Marshal(criterion.Values)
		if err != nil {
			diagnostics.AddError("Failed to encode criteria values", fmt.Sprintf("Unable to marshal criteria values to JSON: %s", err))
			return QueryRuleCriteriaModel{}
		}
		model.Values = jsontypes.Normalized{StringValue: fwtypes.StringValue(string(encoded))}
	}

	return model
}

func queryRuleActionsFromAPI(ctx context.Context, actions types.QueryRuleActions, diagnostics *diag.Diagnostics) (fwtypes.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := QueryRuleActionsModel{
		IDs:  fwtypes.ListNull(fwtypes.StringType),
		Docs: fwtypes.ListNull(fwtypes.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}),
	}

	if len(actions.Ids) > 0 {
		ids, d := fwtypes.ListValueFrom(ctx, fwtypes.StringType, actions.Ids)
		diags.Append(d...)
		if diags.HasError() {
			return fwtypes.ObjectNull(queryRuleActionsModelAttrTypes()), diags
		}
		model.IDs = ids
	}

	if len(actions.Docs) > 0 {
		docModels := make([]QueryRuleActionDocModel, len(actions.Docs))
		for i, doc := range actions.Docs {
			docModels[i] = QueryRuleActionDocModel{
				Index: fwtypes.StringValue(doc.Index_),
				ID:    fwtypes.StringValue(doc.Id_),
			}
		}
		docs, d := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}, docModels)
		diags.Append(d...)
		if diags.HasError() {
			return fwtypes.ObjectNull(queryRuleActionsModelAttrTypes()), diags
		}
		model.Docs = docs
	}

	obj, d := fwtypes.ObjectValueFrom(ctx, queryRuleActionsModelAttrTypes(), model)
	diags.Append(d...)
	return obj, diags
}

func (data QueryRulesetData) toAPIRules(ctx context.Context, diagnostics *diag.Diagnostics) []types.QueryRule {
	var models []QueryRuleModel
	diagnostics.Append(data.Rules.ElementsAs(ctx, &models, false)...)
	if diagnostics.HasError() {
		return nil
	}

	rules := make([]types.QueryRule, len(models))
	for i, model := range models {
		rules[i] = model.toAPIRule(ctx, diagnostics)
		if diagnostics.HasError() {
			return nil
		}
	}

	return rules
}

func (model QueryRuleModel) toAPIRule(ctx context.Context, diagnostics *diag.Diagnostics) types.QueryRule {
	var criteriaModels []QueryRuleCriteriaModel
	diagnostics.Append(model.Criteria.ElementsAs(ctx, &criteriaModels, false)...)
	if diagnostics.HasError() {
		return types.QueryRule{}
	}

	criteria := make([]types.QueryRuleCriteria, len(criteriaModels))
	for i, criterion := range criteriaModels {
		criteria[i] = criterion.toAPICriteria(diagnostics)
		if diagnostics.HasError() {
			return types.QueryRule{}
		}
	}

	var actionsModel QueryRuleActionsModel
	diagnostics.Append(model.Actions.As(ctx, &actionsModel, basetypes.ObjectAsOptions{})...)
	if diagnostics.HasError() {
		return types.QueryRule{}
	}

	ruleType, typeDiags := queryRuleTypeFromString(model.Type.ValueString())
	diagnostics.Append(typeDiags...)
	if diagnostics.HasError() {
		return types.QueryRule{}
	}

	rule := types.QueryRule{
		RuleId:   model.RuleID.ValueString(),
		Type:     ruleType,
		Criteria: criteria,
		Actions:  actionsModel.toAPIActions(ctx, diagnostics),
	}

	if !model.Priority.IsNull() && !model.Priority.IsUnknown() {
		priority := int(model.Priority.ValueInt64())
		rule.Priority = &priority
	}

	return rule
}

func (model QueryRuleCriteriaModel) toAPICriteria(diagnostics *diag.Diagnostics) types.QueryRuleCriteria {
	criteriaType, typeDiags := queryRuleCriteriaTypeFromString(model.Type.ValueString())
	diagnostics.Append(typeDiags...)
	if diagnostics.HasError() {
		return types.QueryRuleCriteria{}
	}

	criterion := types.QueryRuleCriteria{
		Type: criteriaType,
	}

	if !model.Metadata.IsNull() && !model.Metadata.IsUnknown() {
		metadata := model.Metadata.ValueString()
		criterion.Metadata = &metadata
	}

	if !model.Values.IsNull() && !model.Values.IsUnknown() {
		var values []json.RawMessage
		if err := json.Unmarshal([]byte(model.Values.ValueString()), &values); err != nil {
			diagnostics.AddError("Invalid criteria values", fmt.Sprintf("Unable to decode criteria values JSON: %s", err))
			return types.QueryRuleCriteria{}
		}
		criterion.Values = values
	}

	return criterion
}

func (model QueryRuleActionsModel) toAPIActions(ctx context.Context, diagnostics *diag.Diagnostics) types.QueryRuleActions {
	var actions types.QueryRuleActions

	if !model.IDs.IsNull() && !model.IDs.IsUnknown() {
		var ids []string
		diagnostics.Append(model.IDs.ElementsAs(ctx, &ids, false)...)
		if diagnostics.HasError() {
			return actions
		}
		actions.Ids = ids
	}

	if !model.Docs.IsNull() && !model.Docs.IsUnknown() {
		var docModels []QueryRuleActionDocModel
		diagnostics.Append(model.Docs.ElementsAs(ctx, &docModels, false)...)
		if diagnostics.HasError() {
			return actions
		}

		docs := make([]types.PinnedDoc, len(docModels))
		for i, doc := range docModels {
			docs[i] = types.PinnedDoc{
				Index_: doc.Index.ValueString(),
				Id_:    doc.ID.ValueString(),
			}
		}
		actions.Docs = docs
	}

	return actions
}

func queryRuleTypeFromString(value string) (queryruletype.QueryRuleType, diag.Diagnostics) {
	var diags diag.Diagnostics
	var ruleType queryruletype.QueryRuleType
	if err := ruleType.UnmarshalText([]byte(value)); err != nil {
		diags.AddError("Invalid rule type", fmt.Sprintf("Unable to parse rule type %q: %s", value, err))
	}
	return ruleType, diags
}

func queryRuleCriteriaTypeFromString(value string) (queryrulecriteriatype.QueryRuleCriteriaType, diag.Diagnostics) {
	var diags diag.Diagnostics
	var criteriaType queryrulecriteriatype.QueryRuleCriteriaType
	if err := criteriaType.UnmarshalText([]byte(value)); err != nil {
		diags.AddError("Invalid criteria type", fmt.Sprintf("Unable to parse criteria type %q: %s", value, err))
	}
	return criteriaType, diags
}
