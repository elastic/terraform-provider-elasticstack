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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ThresholdRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPIThresholdRule]
}

func newThresholdRuleProcessor() ThresholdRuleProcessor {
	return ThresholdRuleProcessor{
		baseRuleProcessor: baseRuleProcessor[kbapi.SecurityDetectionsAPIThresholdRule]{
			updateFn: func(ctx context.Context, v *kbapi.SecurityDetectionsAPIThresholdRule, d *Data) diag.Diagnostics {
				return d.updateFromThresholdRule(ctx, v)
			},
			idFn: func(v kbapi.SecurityDetectionsAPIThresholdRule) string {
				return v.Id.String()
			},
		},
	}
}

func (th ThresholdRuleProcessor) HandlesRuleType(t string) bool {
	return t == "threshold"
}

func (th ThresholdRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toThresholdRuleCreateProps(ctx, client)
}

func (th ThresholdRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toThresholdRuleUpdateProps(ctx, client)
}

func (d Data) toThresholdRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	thresholdRule := kbapi.SecurityDetectionsAPIThresholdRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIThresholdRuleCreatePropsType("threshold"),
		Query:       d.Query.ValueString(),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set threshold - this is required for threshold rules
	threshold := d.thresholdToAPI(ctx, &diags)
	if threshold != nil {
		thresholdRule.Threshold = *threshold
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &thresholdRule.Actions,
		ResponseActions:                   &thresholdRule.ResponseActions,
		RuleID:                            &thresholdRule.RuleId,
		Enabled:                           &thresholdRule.Enabled,
		From:                              &thresholdRule.From,
		To:                                &thresholdRule.To,
		Interval:                          &thresholdRule.Interval,
		Index:                             &thresholdRule.Index,
		Author:                            &thresholdRule.Author,
		Tags:                              &thresholdRule.Tags,
		FalsePositives:                    &thresholdRule.FalsePositives,
		References:                        &thresholdRule.References,
		License:                           &thresholdRule.License,
		Note:                              &thresholdRule.Note,
		Setup:                             &thresholdRule.Setup,
		MaxSignals:                        &thresholdRule.MaxSignals,
		Version:                           &thresholdRule.Version,
		ExceptionsList:                    &thresholdRule.ExceptionsList,
		RiskScoreMapping:                  &thresholdRule.RiskScoreMapping,
		SeverityMapping:                   &thresholdRule.SeverityMapping,
		RelatedIntegrations:               &thresholdRule.RelatedIntegrations,
		RequiredFields:                    &thresholdRule.RequiredFields,
		BuildingBlockType:                 &thresholdRule.BuildingBlockType,
		DataViewID:                        &thresholdRule.DataViewId,
		Namespace:                         &thresholdRule.Namespace,
		RuleNameOverride:                  &thresholdRule.RuleNameOverride,
		TimestampOverride:                 &thresholdRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &thresholdRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &thresholdRule.InvestigationFields,
		Filters:                           &thresholdRule.Filters,
		Threat:                            &thresholdRule.Threat,
		AlertSuppression:                  nil, // Handle specially for threshold rule
		TimelineID:                        &thresholdRule.TimelineId,
		TimelineTitle:                     &thresholdRule.TimelineTitle,
	}, &diags, client)

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
func (d Data) toThresholdRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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
		Type:        kbapi.SecurityDetectionsAPIThresholdRuleUpdatePropsType("threshold"),
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

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &thresholdRule.Actions,
		ResponseActions:                   &thresholdRule.ResponseActions,
		RuleID:                            &thresholdRule.RuleId,
		Enabled:                           &thresholdRule.Enabled,
		From:                              &thresholdRule.From,
		To:                                &thresholdRule.To,
		Interval:                          &thresholdRule.Interval,
		Index:                             &thresholdRule.Index,
		Author:                            &thresholdRule.Author,
		Tags:                              &thresholdRule.Tags,
		FalsePositives:                    &thresholdRule.FalsePositives,
		References:                        &thresholdRule.References,
		License:                           &thresholdRule.License,
		Note:                              &thresholdRule.Note,
		InvestigationFields:               &thresholdRule.InvestigationFields,
		Setup:                             &thresholdRule.Setup,
		MaxSignals:                        &thresholdRule.MaxSignals,
		Version:                           &thresholdRule.Version,
		ExceptionsList:                    &thresholdRule.ExceptionsList,
		RiskScoreMapping:                  &thresholdRule.RiskScoreMapping,
		SeverityMapping:                   &thresholdRule.SeverityMapping,
		RelatedIntegrations:               &thresholdRule.RelatedIntegrations,
		RequiredFields:                    &thresholdRule.RequiredFields,
		BuildingBlockType:                 &thresholdRule.BuildingBlockType,
		DataViewID:                        &thresholdRule.DataViewId,
		Namespace:                         &thresholdRule.Namespace,
		RuleNameOverride:                  &thresholdRule.RuleNameOverride,
		TimestampOverride:                 &thresholdRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &thresholdRule.TimestampOverrideFallbackDisabled,
		Filters:                           &thresholdRule.Filters,
		Threat:                            &thresholdRule.Threat,
		AlertSuppression:                  nil, // Handle specially for threshold rule
		TimelineID:                        &thresholdRule.TimelineId,
		TimelineTitle:                     &thresholdRule.TimelineTitle,
	}, &diags, client)

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

	// Threshold rules use a different AlertSuppression type, so we pass nil and handle it separately below.
	diags.Append(d.updateCommonRuleFieldsFromAPI(ctx, commonAPIRuleFields{
		ResourceID:                        rule.Id.String(),
		RuleID:                            rule.RuleId,
		Name:                              rule.Name,
		Type:                              string(rule.Type),
		Enabled:                           rule.Enabled,
		From:                              rule.From,
		To:                                rule.To,
		Interval:                          rule.Interval,
		Description:                       rule.Description,
		RiskScore:                         int64(rule.RiskScore),
		Severity:                          string(rule.Severity),
		MaxSignals:                        int64(rule.MaxSignals),
		Version:                           int64(rule.Version),
		Revision:                          int64(rule.Revision),
		CreatedAt:                         rule.CreatedAt,
		CreatedBy:                         rule.CreatedBy,
		UpdatedAt:                         rule.UpdatedAt,
		UpdatedBy:                         rule.UpdatedBy,
		TimelineID:                        rule.TimelineId,
		TimelineTitle:                     rule.TimelineTitle,
		DataViewID:                        rule.DataViewId,
		Namespace:                         rule.Namespace,
		RuleNameOverride:                  rule.RuleNameOverride,
		TimestampOverride:                 rule.TimestampOverride,
		TimestampOverrideFallbackDisabled: rule.TimestampOverrideFallbackDisabled,
		BuildingBlockType:                 rule.BuildingBlockType,
		License:                           rule.License,
		Note:                              rule.Note,
		Setup:                             rule.Setup,
		Index:                             rule.Index,
		Author:                            rule.Author,
		Tags:                              rule.Tags,
		FalsePositives:                    rule.FalsePositives,
		References:                        rule.References,
		Actions:                           rule.Actions,
		ExceptionsList:                    rule.ExceptionsList,
		RiskScoreMapping:                  rule.RiskScoreMapping,
		InvestigationFields:               rule.InvestigationFields,
		Threat:                            rule.Threat,
		SeverityMapping:                   rule.SeverityMapping,
		RelatedIntegrations:               rule.RelatedIntegrations,
		RequiredFields:                    rule.RequiredFields,
		AlertSuppression:                  nil, // handled below via updateThresholdAlertSuppressionFromAPI
		ResponseActions:                   rule.ResponseActions,
	})...)

	d.Query = typeutils.StringishValue(rule.Query)
	d.Language = typeutils.StringishValue(rule.Language)

	diags.Append(d.updateFiltersFromAPI(ctx, rule.Filters)...)

	// Threshold-specific fields
	thresholdObj, thresholdDiags := convertThresholdToModel(ctx, rule.Threshold)
	diags.Append(thresholdDiags...)
	if !thresholdDiags.HasError() {
		d.Threshold = thresholdObj
	}

	if rule.SavedId != nil {
		d.SavedID = types.StringValue(*rule.SavedId)
	} else {
		d.SavedID = types.StringNull()
	}

	// Threshold uses a distinct alert suppression type that overwrites the null set by the common helper.
	diags.Append(d.updateThresholdAlertSuppressionFromAPI(ctx, rule.AlertSuppression)...)

	return diags
}
