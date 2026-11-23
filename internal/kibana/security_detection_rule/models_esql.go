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

type EsqlRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPIEsqlRule]
}

func newEsqlRuleProcessor() EsqlRuleProcessor {
	return EsqlRuleProcessor{
		updateFn: func(ctx context.Context, v *kbapi.SecurityDetectionsAPIEsqlRule, d *Data) diag.Diagnostics {
			return d.updateFromEsqlRule(ctx, v)
		},
		idFn: func(v kbapi.SecurityDetectionsAPIEsqlRule) string {
			return v.Id.String()
		},
	}
}

func (e EsqlRuleProcessor) HandlesRuleType(t string) bool {
	return t == ruleTypeESQL
}

func (e EsqlRuleProcessor) ToCreateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toEsqlRuleCreateProps(ctx)
}

func (e EsqlRuleProcessor) ToUpdateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toEsqlRuleUpdateProps(ctx)
}

func (d Data) toEsqlRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	esqlRule := kbapi.SecurityDetectionsAPIEsqlRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIEsqlRuleCreatePropsType(ruleTypeESQL),
		Query:       d.Query.ValueString(),
		Language:    kbapi.SecurityDetectionsAPIEsqlQueryLanguage(ruleTypeESQL),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// ESQL rules don't use index patterns (they use the FROM clause in the query) and don't
	// support DataViewID or Filters; buildCommonRuleProps leaves those nil automatically since
	// esqlRule has no such fields.
	d.setCommonCreateProps(ctx, buildCommonRuleProps(&esqlRule), &diags)

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIEsqlRuleCreateProps(esqlRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert ESQL rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func (d Data) toEsqlRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	uid, ok := d.parseResourceUUID(&diags)
	if !ok {
		return updateProps, diags
	}

	esqlRule := kbapi.SecurityDetectionsAPIEsqlRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIEsqlRuleUpdatePropsType(ruleTypeESQL),
		Query:       d.Query.ValueString(),
		Language:    kbapi.SecurityDetectionsAPIEsqlQueryLanguage(ruleTypeESQL),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		esqlRule.RuleId = &ruleID
		esqlRule.Id = nil // if rule_id is set, we cant send id
	}

	// ESQL rules don't use index patterns (they use the FROM clause in the query) and don't
	// support DataViewID or Filters; buildCommonRuleProps leaves those nil automatically since
	// esqlRule has no such fields.
	d.setCommonUpdateProps(ctx, buildCommonRuleProps(&esqlRule), &diags)

	// Convert to union type
	err := updateProps.FromSecurityDetectionsAPIEsqlRuleUpdateProps(esqlRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert ESQL rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}
func (d *Data) updateFromEsqlRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIEsqlRule) diag.Diagnostics {
	var diags diag.Diagnostics

	// ESQL rules don't support DataViewId or Filters; buildCommonAPIRuleFields leaves those
	// zero-valued automatically since rule has no such fields.
	diags.Append(d.updateCommonRuleFieldsFromAPI(ctx, buildCommonAPIRuleFields(rule.Id.String(), rule))...)

	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))

	return diags
}
