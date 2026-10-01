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

type QueryRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPIQueryRule]
}

func newQueryRuleProcessor() QueryRuleProcessor {
	return QueryRuleProcessor{
		updateFn: updateFromQueryRule,
		idFn: func(v kbapi.SecurityDetectionsAPIQueryRule) string {
			return v.Id.String()
		},
	}
}

func (q QueryRuleProcessor) HandlesRuleType(t string) bool {
	return t == ruleTypeQuery
}

func (q QueryRuleProcessor) ToCreateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return toQueryRuleCreateProps(ctx, d)
}

func (q QueryRuleProcessor) ToUpdateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return toQueryRuleUpdateProps(ctx, d)
}

func toQueryRuleCreateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	queryRuleQuery := d.Query.ValueString()
	queryRule := kbapi.SecurityDetectionsAPIQueryRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleCreatePropsType(ruleTypeQuery),
		Query:       &queryRuleQuery,
		RiskScore:   int(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, buildCommonRuleProps(&queryRule), &diags)

	// Set query-specific fields
	queryRule.Language = d.getKQLQueryLanguage()

	if typeutils.IsKnown(d.SavedID) {
		savedID := d.SavedID.ValueString()
		queryRule.SavedId = &savedID
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIQueryRuleCreateProps(queryRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert query rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func toQueryRuleUpdateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	queryRuleQuery := d.Query.ValueString()

	uid, ok := d.parseResourceUUID(&diags)
	if !ok {
		return updateProps, diags
	}

	queryRule := kbapi.SecurityDetectionsAPIQueryRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleUpdatePropsType(ruleTypeQuery),
		Query:       &queryRuleQuery,
		RiskScore:   int(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		queryRule.RuleId = &ruleID
		queryRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, buildCommonRuleProps(&queryRule), &diags)

	// Set query-specific fields
	queryRule.Language = d.getKQLQueryLanguage()

	if typeutils.IsKnown(d.SavedID) {
		savedID := d.SavedID.ValueString()
		queryRule.SavedId = &savedID
	}

	// Convert to union type
	err := updateProps.FromSecurityDetectionsAPIQueryRuleUpdateProps(queryRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert query rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}
func updateFromQueryRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIQueryRule, d *Data) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(d.updateCommonRuleFieldsFromAPI(ctx, buildCommonAPIRuleFields(rule.Id.String(), rule))...)

	d.Query = types.StringValue(rule.Query)
	d.Language = typeutils.StringishValue(rule.Language)

	diags.Append(d.updateFiltersFromAPI(ctx, rule.Filters)...)

	return diags
}
