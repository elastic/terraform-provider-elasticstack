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
)

type ThresholdRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPIThresholdRule]
}

func newThresholdRuleProcessor() ThresholdRuleProcessor {
	return ThresholdRuleProcessor{
		updateFn: func(ctx context.Context, v *kbapi.SecurityDetectionsAPIThresholdRule, d *Data) diag.Diagnostics {
			return d.updateFromThresholdRule(ctx, v)
		},
		idFn: func(v kbapi.SecurityDetectionsAPIThresholdRule) string {
			return v.Id.String()
		},
	}
}

func (th ThresholdRuleProcessor) HandlesRuleType(t string) bool {
	return t == ruleTypeThreshold
}

func (th ThresholdRuleProcessor) ToCreateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toThresholdRuleCreateProps(ctx)
}

func (th ThresholdRuleProcessor) ToUpdateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toThresholdRuleUpdateProps(ctx)
}

func (d Data) toThresholdRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	thresholdRule := kbapi.SecurityDetectionsAPIThresholdRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIThresholdRuleCreatePropsType(ruleTypeThreshold),
		Query:       d.Query.ValueString(),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set threshold - this is required for threshold rules
	threshold := d.thresholdToAPI(ctx, &diags)
	if threshold != nil {
		thresholdRule.Threshold = *threshold
	}

	// Threshold's AlertSuppression is a distinct API type; excluded here and handled specially below.
	d.setCommonCreateProps(ctx, buildCommonRuleProps(&thresholdRule, "AlertSuppression"), &diags)

	// Handle threshold-specific alert suppression
	if typeutils.IsKnown(d.AlertSuppression) {
		alertSuppression := d.alertSuppressionToThresholdAPI(ctx, &diags)
		if alertSuppression != nil {
			thresholdRule.AlertSuppression = alertSuppression
		}
	}

	// Set query language
	thresholdRule.Language = d.getKQLQueryLanguage()

	if typeutils.IsKnown(d.SavedID) {
		savedID := d.SavedID.ValueString()
		thresholdRule.SavedId = &savedID
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIThresholdRuleCreateProps(thresholdRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert threshold rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}
func (d Data) toThresholdRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	uid, ok := d.parseResourceUUID(&diags)
	if !ok {
		return updateProps, diags
	}

	thresholdRule := kbapi.SecurityDetectionsAPIThresholdRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIThresholdRuleUpdatePropsType(ruleTypeThreshold),
		Query:       d.Query.ValueString(),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		thresholdRule.RuleId = &ruleID
		thresholdRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set threshold - this is required for threshold rules
	threshold := d.thresholdToAPI(ctx, &diags)
	if threshold != nil {
		thresholdRule.Threshold = *threshold
	}

	// Threshold's AlertSuppression is a distinct API type; excluded here and handled specially below.
	d.setCommonUpdateProps(ctx, buildCommonRuleProps(&thresholdRule, "AlertSuppression"), &diags)

	// Handle threshold-specific alert suppression
	if typeutils.IsKnown(d.AlertSuppression) {
		alertSuppression := d.alertSuppressionToThresholdAPI(ctx, &diags)
		if alertSuppression != nil {
			thresholdRule.AlertSuppression = alertSuppression
		}
	}

	// Set query language
	thresholdRule.Language = d.getKQLQueryLanguage()

	if typeutils.IsKnown(d.SavedID) {
		savedID := d.SavedID.ValueString()
		thresholdRule.SavedId = &savedID
	}

	// Convert to union type
	err := updateProps.FromSecurityDetectionsAPIThresholdRuleUpdateProps(thresholdRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert threshold rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d *Data) updateFromThresholdRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIThresholdRule) diag.Diagnostics {
	var diags diag.Diagnostics

	// Threshold rules use a different AlertSuppression type, so we exclude it here and handle it
	// separately below via updateThresholdAlertSuppressionFromAPI.
	diags.Append(d.updateCommonRuleFieldsFromAPI(ctx, buildCommonAPIRuleFields(rule.Id.String(), rule, "AlertSuppression"))...)

	d.Query = typeutils.StringishValue(rule.Query)
	d.Language = typeutils.StringishValue(rule.Language)

	diags.Append(d.updateFiltersFromAPI(ctx, rule.Filters)...)

	// Threshold-specific fields
	thresholdObj, thresholdDiags := convertThresholdToModel(ctx, rule.Threshold)
	diags.Append(thresholdDiags...)
	if !thresholdDiags.HasError() {
		d.Threshold = thresholdObj
	}

	d.SavedID = typeutils.StringishPointerValue(rule.SavedId)

	// Threshold uses a distinct alert suppression type that overwrites the null set by the common helper.
	diags.Append(d.updateThresholdAlertSuppressionFromAPI(ctx, rule.AlertSuppression)...)

	return diags
}
