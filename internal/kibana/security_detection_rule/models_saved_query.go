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

package securitydetectionrule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SavedQueryRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPISavedQueryRule]
}

func newSavedQueryRuleProcessor() SavedQueryRuleProcessor {
	return SavedQueryRuleProcessor{
		updateFn: func(ctx context.Context, v *kbapi.SecurityDetectionsAPISavedQueryRule, d *Data) diag.Diagnostics {
			return d.updateFromSavedQueryRule(ctx, v)
		},
		idFn: func(v kbapi.SecurityDetectionsAPISavedQueryRule) string {
			return v.Id.String()
		},
	}
}

func (s SavedQueryRuleProcessor) HandlesRuleType(t string) bool {
	return t == ruleTypeSavedQuery
}

func (s SavedQueryRuleProcessor) ToCreateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toSavedQueryRuleCreateProps(ctx)
}

func (s SavedQueryRuleProcessor) ToUpdateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toSavedQueryRuleUpdateProps(ctx)
}

func (d Data) toSavedQueryRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	savedQueryRule := kbapi.SecurityDetectionsAPISavedQueryRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPISavedQueryRuleCreatePropsType(ruleTypeSavedQuery),
		SavedId:     d.SavedID.ValueString(),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, buildCommonRuleProps(&savedQueryRule), &diags)

	// Set optional query for saved query rules
	if typeutils.IsKnown(d.Query) {
		query := d.Query.ValueString()
		savedQueryRule.Query = &query
	}

	// Set query language
	savedQueryRule.Language = d.getKQLQueryLanguage()

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPISavedQueryRuleCreateProps(savedQueryRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert saved query rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}
func (d Data) toSavedQueryRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	uid, ok := d.parseResourceUUID(&diags)
	if !ok {
		return updateProps, diags
	}

	savedQueryRule := kbapi.SecurityDetectionsAPISavedQueryRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPISavedQueryRuleUpdatePropsType(ruleTypeSavedQuery),
		SavedId:     d.SavedID.ValueString(),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		savedQueryRule.RuleId = &ruleID
		savedQueryRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, buildCommonRuleProps(&savedQueryRule), &diags)

	// Set optional query for saved query rules
	if typeutils.IsKnown(d.Query) {
		query := d.Query.ValueString()
		savedQueryRule.Query = &query
	}

	// Set query language
	savedQueryRule.Language = d.getKQLQueryLanguage()

	// Convert to union type
	err := updateProps.FromSecurityDetectionsAPISavedQueryRuleUpdateProps(savedQueryRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert saved query rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d *Data) updateFromSavedQueryRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPISavedQueryRule) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(d.updateCommonRuleFieldsFromAPI(ctx, buildCommonAPIRuleFields(rule.Id.String(), rule))...)

	d.SavedID = types.StringValue(rule.SavedId)
	d.Query = types.StringPointerValue(rule.Query)
	d.Language = typeutils.StringishValue(rule.Language)

	diags.Append(d.updateFiltersFromAPI(ctx, rule.Filters)...)

	return diags
}
