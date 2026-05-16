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

type EsqlRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPIEsqlRule]
}

func newEsqlRuleProcessor() EsqlRuleProcessor {
	return EsqlRuleProcessor{
		baseRuleProcessor: baseRuleProcessor[kbapi.SecurityDetectionsAPIEsqlRule]{
			updateFn: func(ctx context.Context, v *kbapi.SecurityDetectionsAPIEsqlRule, d *Data) diag.Diagnostics {
				return d.updateFromEsqlRule(ctx, v)
			},
			idFn: func(v kbapi.SecurityDetectionsAPIEsqlRule) string {
				return v.Id.String()
			},
		},
	}
}

func (e EsqlRuleProcessor) HandlesRuleType(t string) bool {
	return t == "esql"
}

func (e EsqlRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toEsqlRuleCreateProps(ctx, client)
}

func (e EsqlRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toEsqlRuleUpdateProps(ctx, client)
}

func (d Data) toEsqlRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	esqlRule := kbapi.SecurityDetectionsAPIEsqlRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIEsqlRuleCreatePropsType("esql"),
		Query:       d.Query.ValueString(),
		Language:    kbapi.SecurityDetectionsAPIEsqlQueryLanguage("esql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &esqlRule.Actions,
		ResponseActions:                   &esqlRule.ResponseActions,
		RuleID:                            &esqlRule.RuleId,
		Enabled:                           &esqlRule.Enabled,
		From:                              &esqlRule.From,
		To:                                &esqlRule.To,
		Interval:                          &esqlRule.Interval,
		Index:                             nil, // ESQL rules don't use index patterns
		Author:                            &esqlRule.Author,
		Tags:                              &esqlRule.Tags,
		FalsePositives:                    &esqlRule.FalsePositives,
		References:                        &esqlRule.References,
		License:                           &esqlRule.License,
		Note:                              &esqlRule.Note,
		Setup:                             &esqlRule.Setup,
		MaxSignals:                        &esqlRule.MaxSignals,
		Version:                           &esqlRule.Version,
		ExceptionsList:                    &esqlRule.ExceptionsList,
		AlertSuppression:                  &esqlRule.AlertSuppression,
		RiskScoreMapping:                  &esqlRule.RiskScoreMapping,
		SeverityMapping:                   &esqlRule.SeverityMapping,
		RelatedIntegrations:               &esqlRule.RelatedIntegrations,
		RequiredFields:                    &esqlRule.RequiredFields,
		BuildingBlockType:                 &esqlRule.BuildingBlockType,
		DataViewID:                        nil, // ESQL rules don't have DataViewID
		Namespace:                         &esqlRule.Namespace,
		RuleNameOverride:                  &esqlRule.RuleNameOverride,
		TimestampOverride:                 &esqlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &esqlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &esqlRule.InvestigationFields,
		Filters:                           nil, // ESQL rules don't support this field
		Threat:                            &esqlRule.Threat,
		TimelineID:                        &esqlRule.TimelineId,
		TimelineTitle:                     &esqlRule.TimelineTitle,
	}, &diags, client)

	// ESQL rules don't use index patterns as they use FROM clause in the query

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

func (d Data) toEsqlRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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
		Type:        kbapi.SecurityDetectionsAPIEsqlRuleUpdatePropsType("esql"),
		Query:       d.Query.ValueString(),
		Language:    kbapi.SecurityDetectionsAPIEsqlQueryLanguage("esql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		esqlRule.RuleId = &ruleID
		esqlRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &esqlRule.Actions,
		ResponseActions:                   &esqlRule.ResponseActions,
		RuleID:                            &esqlRule.RuleId,
		Enabled:                           &esqlRule.Enabled,
		From:                              &esqlRule.From,
		To:                                &esqlRule.To,
		Interval:                          &esqlRule.Interval,
		Index:                             nil, // ESQL rules don't use index patterns
		Author:                            &esqlRule.Author,
		Tags:                              &esqlRule.Tags,
		FalsePositives:                    &esqlRule.FalsePositives,
		References:                        &esqlRule.References,
		License:                           &esqlRule.License,
		Note:                              &esqlRule.Note,
		Setup:                             &esqlRule.Setup,
		MaxSignals:                        &esqlRule.MaxSignals,
		Version:                           &esqlRule.Version,
		ExceptionsList:                    &esqlRule.ExceptionsList,
		AlertSuppression:                  &esqlRule.AlertSuppression,
		RiskScoreMapping:                  &esqlRule.RiskScoreMapping,
		SeverityMapping:                   &esqlRule.SeverityMapping,
		RelatedIntegrations:               &esqlRule.RelatedIntegrations,
		RequiredFields:                    &esqlRule.RequiredFields,
		BuildingBlockType:                 &esqlRule.BuildingBlockType,
		DataViewID:                        nil, // ESQL rules don't have DataViewID
		Namespace:                         &esqlRule.Namespace,
		RuleNameOverride:                  &esqlRule.RuleNameOverride,
		TimestampOverride:                 &esqlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &esqlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &esqlRule.InvestigationFields,
		Filters:                           nil, // ESQL rules don't have Filters
		Threat:                            &esqlRule.Threat,
		TimelineID:                        &esqlRule.TimelineId,
		TimelineTitle:                     &esqlRule.TimelineTitle,
	}, &diags, client)

	// ESQL rules don't use index patterns as they use FROM clause in the query

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

	// ESQL rules don't support DataViewId or Filters; pass nil so the common helper sets them to their zero values.
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
		DataViewID:                        nil, // ESQL rules don't have DataViewId
		Namespace:                         rule.Namespace,
		RuleNameOverride:                  rule.RuleNameOverride,
		TimestampOverride:                 rule.TimestampOverride,
		TimestampOverrideFallbackDisabled: rule.TimestampOverrideFallbackDisabled,
		BuildingBlockType:                 rule.BuildingBlockType,
		License:                           rule.License,
		Note:                              rule.Note,
		Setup:                             rule.Setup,
		Index:                             nil, // ESQL rules don't use index patterns
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
		AlertSuppression:                  rule.AlertSuppression,
		ResponseActions:                   rule.ResponseActions,
	})...)

	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))

	return diags
}
