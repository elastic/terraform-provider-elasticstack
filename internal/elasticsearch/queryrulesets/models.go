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
	"reflect"
	"strconv"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/queryrulecriteriatype"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/queryruletype"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var attrTypesString = fwtypes.StringType

const (
	queryRuleActionDocIndexAttrName   = "_index"
	queryRuleActionDocIDAttrName      = "_id"
	queryRuleActionsIDsAttrName       = "ids"
	queryRuleActionsDocsAttrName      = "docs"
	queryRuleCriteriaTypeAttrName     = "type"
	queryRuleCriteriaMetadataAttrName = "metadata"
	queryRuleCriteriaValuesAttrName   = "values"
	queryRuleRuleIDAttrName           = "rule_id"
	queryRulePriorityAttrName         = "priority"
	queryRuleCriteriaAttrName         = "criteria"
	queryRuleActionsAttrName          = "actions"
)

func queryRuleActionDocModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		queryRuleActionDocIndexAttrName: fwtypes.StringType,
		queryRuleActionDocIDAttrName:    fwtypes.StringType,
	}
}

// QueryRuleActionsModel represents the actions block for a query rule.
type QueryRuleActionsModel struct {
	IDs  fwtypes.List `tfsdk:"ids"`
	Docs fwtypes.List `tfsdk:"docs"`
}

func queryRuleActionsModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		queryRuleActionsIDsAttrName:  fwtypes.ListType{ElemType: fwtypes.StringType},
		queryRuleActionsDocsAttrName: fwtypes.ListType{ElemType: fwtypes.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}},
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
		queryRuleCriteriaTypeAttrName:     fwtypes.StringType,
		queryRuleCriteriaMetadataAttrName: fwtypes.StringType,
		queryRuleCriteriaValuesAttrName:   jsontypes.NormalizedType{},
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
		queryRuleRuleIDAttrName:       fwtypes.StringType,
		queryRuleCriteriaTypeAttrName: fwtypes.StringType,
		queryRulePriorityAttrName:     fwtypes.Int64Type,
		queryRuleCriteriaAttrName:     fwtypes.ListType{ElemType: fwtypes.ObjectType{AttrTypes: queryRuleCriteriaModelAttrTypes()}},
		queryRuleActionsAttrName:      fwtypes.ObjectType{AttrTypes: queryRuleActionsModelAttrTypes()},
	}
}

// QueryRulesetData is the Terraform state model for the query ruleset resource.
// The data source uses queryRulesetDataSourceModel and maps through this type
// for its shared read/populate logic.
type QueryRulesetData struct {
	entitycore.ElasticsearchConnectionField
	entitycore.ResourceTimeoutsField
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
func (data QueryRulesetData) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *MinSupportedVersion,
		ErrorMessage: "Elasticsearch query rulesets require Elasticsearch v8.16.0 or above (Query Rules API with `priority` and `exclude` rule type support).",
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

	actionsObj, d := queryRuleActionsFromAPI(ctx, rule.Actions)
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

	model.Priority = typeutils.IntPointerToInt64Value(rule.Priority)

	return model
}

func queryRuleCriteriaFromAPI(criterion types.QueryRuleCriteria, diagnostics *diag.Diagnostics) QueryRuleCriteriaModel {
	model := QueryRuleCriteriaModel{
		Type: fwtypes.StringValue(criterion.Type.String()),
	}

	// Elasticsearch returns metadata as an empty string for criteria types that do
	// not use it (notably `always`, since 8.19). Normalize empty strings to null so
	// state stays consistent with configurations that omit `metadata`.
	model.Metadata = typeutils.NonEmptyStringOrNull(criterion.Metadata)

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

func queryRuleActionsFromAPI(ctx context.Context, actions types.QueryRuleActions) (fwtypes.Object, diag.Diagnostics) {
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
		docValues := make([]attr.Value, len(actions.Docs))
		for i, doc := range actions.Docs {
			obj, d := fwtypes.ObjectValue(queryRuleActionDocModelAttrTypes(), map[string]attr.Value{
				queryRuleActionDocIndexAttrName: fwtypes.StringValue(doc.Index_),
				queryRuleActionDocIDAttrName:    fwtypes.StringValue(doc.Id_),
			})
			diags.Append(d...)
			if diags.HasError() {
				return fwtypes.ObjectNull(queryRuleActionsModelAttrTypes()), diags
			}
			docValues[i] = obj
		}
		docs, d := fwtypes.ListValue(fwtypes.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}, docValues)
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
		rule.Priority = typeutils.OptionalInt(model.Priority)
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
		actions.Docs = pinnedDocsFromList(model.Docs, diagnostics)
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

func pinnedDocsFromList(docs fwtypes.List, diagnostics *diag.Diagnostics) []types.PinnedDoc {
	elems := docs.Elements()
	result := make([]types.PinnedDoc, len(elems))
	for i, elem := range elems {
		obj, ok := elem.(fwtypes.Object)
		if !ok {
			diagnostics.AddError("Invalid actions docs", "Expected an object for each docs entry.")
			return nil
		}

		attrs := obj.Attributes()
		indexAttr, ok := attrs[queryRuleActionDocIndexAttrName].(fwtypes.String)
		if !ok || indexAttr.IsNull() || indexAttr.IsUnknown() {
			diagnostics.AddError("Invalid actions docs", "Each docs entry must include `_index`.")
			return nil
		}
		idAttr, ok := attrs[queryRuleActionDocIDAttrName].(fwtypes.String)
		if !ok || idAttr.IsNull() || idAttr.IsUnknown() {
			diagnostics.AddError("Invalid actions docs", "Each docs entry must include `_id`.")
			return nil
		}

		result[i] = types.PinnedDoc{
			Index_: indexAttr.ValueString(),
			Id_:    idAttr.ValueString(),
		}
	}
	return result
}

func preserveCriteriaValuesFromPrior(ctx context.Context, data *QueryRulesetData, priorRules fwtypes.List, diagnostics *diag.Diagnostics) {
	if priorRules.IsNull() || priorRules.IsUnknown() || data.Rules.IsNull() || data.Rules.IsUnknown() {
		return
	}

	var priorModels, currentModels []QueryRuleModel
	diagnostics.Append(priorRules.ElementsAs(ctx, &priorModels, false)...)
	diagnostics.Append(data.Rules.ElementsAs(ctx, &currentModels, false)...)
	if diagnostics.HasError() {
		return
	}

	priorByRuleID := make(map[string]QueryRuleModel, len(priorModels))
	for _, rule := range priorModels {
		priorByRuleID[rule.RuleID.ValueString()] = rule
	}

	for i := range currentModels {
		priorRule, ok := priorByRuleID[currentModels[i].RuleID.ValueString()]
		if !ok {
			continue
		}
		preserveRuleCriteriaValuesFromPrior(ctx, &currentModels[i], priorRule, diagnostics)
		if diagnostics.HasError() {
			return
		}
	}

	list, d := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: queryRuleModelAttrTypes()}, currentModels)
	diagnostics.Append(d...)
	if diagnostics.HasError() {
		return
	}
	data.Rules = list
}

func preserveRuleCriteriaValuesFromPrior(ctx context.Context, current *QueryRuleModel, prior QueryRuleModel, diagnostics *diag.Diagnostics) {
	if current.Criteria.IsNull() || current.Criteria.IsUnknown() || prior.Criteria.IsNull() || prior.Criteria.IsUnknown() {
		return
	}

	var currentCriteria, priorCriteria []QueryRuleCriteriaModel
	diagnostics.Append(current.Criteria.ElementsAs(ctx, &currentCriteria, false)...)
	diagnostics.Append(prior.Criteria.ElementsAs(ctx, &priorCriteria, false)...)
	if diagnostics.HasError() {
		return
	}

	for i := range currentCriteria {
		priorMatch, ok := matchingPriorCriterion(priorCriteria, currentCriteria, i)
		if !ok {
			continue
		}
		if criteriaValuesSemanticallyEqual(priorMatch.Values, currentCriteria[i].Values) {
			currentCriteria[i].Values = priorMatch.Values
		}
	}

	list, d := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: queryRuleCriteriaModelAttrTypes()}, currentCriteria)
	diagnostics.Append(d...)
	if diagnostics.HasError() {
		return
	}
	current.Criteria = list
}

type criteriaIdentityKey struct {
	typ      string
	metadata string
}

func criteriaIdentityKeyFromModel(criterion QueryRuleCriteriaModel) criteriaIdentityKey {
	metadata := ""
	if !criterion.Metadata.IsNull() && !criterion.Metadata.IsUnknown() {
		metadata = criterion.Metadata.ValueString()
	}
	return criteriaIdentityKey{
		typ:      criterion.Type.ValueString(),
		metadata: metadata,
	}
}

func matchingPriorCriterion(priorCriteria, currentCriteria []QueryRuleCriteriaModel, currentIndex int) (QueryRuleCriteriaModel, bool) {
	key := criteriaIdentityKeyFromModel(currentCriteria[currentIndex])

	matches := make([]QueryRuleCriteriaModel, 0, len(priorCriteria))
	for _, prior := range priorCriteria {
		if criteriaIdentityKeyFromModel(prior) == key {
			matches = append(matches, prior)
		}
	}
	if len(matches) == 0 {
		return QueryRuleCriteriaModel{}, false
	}

	occurrence := 0
	for i := range currentIndex {
		if criteriaIdentityKeyFromModel(currentCriteria[i]) == key {
			occurrence++
		}
	}
	if occurrence >= len(matches) {
		return QueryRuleCriteriaModel{}, false
	}
	return matches[occurrence], true
}

func criteriaValuesSemanticallyEqual(prior, current jsontypes.Normalized) bool {
	if prior.IsNull() || prior.IsUnknown() || current.IsNull() || current.IsUnknown() {
		return prior.IsNull() == current.IsNull()
	}

	var priorVals, currentVals []json.RawMessage
	if err := json.Unmarshal([]byte(prior.ValueString()), &priorVals); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(current.ValueString()), &currentVals); err != nil {
		return false
	}
	if len(priorVals) != len(currentVals) {
		return false
	}

	for i := range priorVals {
		if !jsonRawMessageSemanticallyEqual(priorVals[i], currentVals[i]) {
			return false
		}
	}
	return true
}

func jsonRawMessageSemanticallyEqual(a, b json.RawMessage) bool {
	normalize := func(raw json.RawMessage) (any, bool) {
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, false
		}
		switch typed := value.(type) {
		case string:
			if number, err := strconv.ParseFloat(typed, 64); err == nil {
				return number, true
			}
			return typed, true
		default:
			return typed, true
		}
	}

	left, okLeft := normalize(a)
	right, okRight := normalize(b)
	if !okLeft || !okRight {
		return string(a) == string(b)
	}
	return reflect.DeepEqual(left, right)
}

func queryRuleCriteriaTypeFromString(value string) (queryrulecriteriatype.QueryRuleCriteriaType, diag.Diagnostics) {
	var diags diag.Diagnostics
	var criteriaType queryrulecriteriatype.QueryRuleCriteriaType
	if err := criteriaType.UnmarshalText([]byte(value)); err != nil {
		diags.AddError("Invalid criteria type", fmt.Sprintf("Unable to parse criteria type %q: %s", value, err))
	}
	return criteriaType, diags
}
