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

type EqlRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPIEqlRule]
}

func newEqlRuleProcessor() EqlRuleProcessor {
	return EqlRuleProcessor{
		updateFn: updateFromEqlRule,
		idFn: func(v kbapi.SecurityDetectionsAPIEqlRule) string {
			return v.Id.String()
		},
	}
}

func (e EqlRuleProcessor) HandlesRuleType(t string) bool {
	return t == ruleTypeEQL
}

func (e EqlRuleProcessor) ToCreateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return toEqlRuleCreateProps(ctx, d)
}

func (e EqlRuleProcessor) ToUpdateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return toEqlRuleUpdateProps(ctx, d)
}

func toEqlRuleCreateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	eqlRule := kbapi.SecurityDetectionsAPIEqlRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIEqlRuleCreatePropsType(ruleTypeEQL),
		Query:       d.Query.ValueString(),
		Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguage(ruleTypeEQL),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, buildCommonRuleProps(&eqlRule), &diags)

	// Set EQL-specific fields
	if typeutils.IsKnown(d.TiebreakerField) {
		tiebreakerField := d.TiebreakerField.ValueString()
		eqlRule.TiebreakerField = &tiebreakerField
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIEqlRuleCreateProps(eqlRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert EQL rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}
func toEqlRuleUpdateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	uid, ok := d.parseResourceUUID(&diags)
	if !ok {
		return updateProps, diags
	}

	eqlRule := kbapi.SecurityDetectionsAPIEqlRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIEqlRuleUpdatePropsType(ruleTypeEQL),
		Query:       d.Query.ValueString(),
		Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguage(ruleTypeEQL),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		eqlRule.RuleId = &ruleID
		eqlRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, buildCommonRuleProps(&eqlRule), &diags)

	// Set EQL-specific fields
	if typeutils.IsKnown(d.TiebreakerField) {
		tiebreakerField := d.TiebreakerField.ValueString()
		eqlRule.TiebreakerField = &tiebreakerField
	}

	// Convert to union type
	err := updateProps.FromSecurityDetectionsAPIEqlRuleUpdateProps(eqlRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert EQL rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}
func updateFromEqlRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIEqlRule, d *Data) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(d.updateCommonRuleFieldsFromAPI(ctx, buildCommonAPIRuleFields(rule.Id.String(), rule))...)

	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))

	diags.Append(d.updateFiltersFromAPI(ctx, rule.Filters)...)

	d.TiebreakerField = typeutils.StringishPointerValue(rule.TiebreakerField)

	return diags
}
