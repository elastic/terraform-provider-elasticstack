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
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EqlRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPIEqlRule]
}

func newEqlRuleProcessor() EqlRuleProcessor {
	return EqlRuleProcessor{
		baseRuleProcessor: baseRuleProcessor[kbapi.SecurityDetectionsAPIEqlRule]{
			updateFn: updateFromEqlRule,
			idFn: func(v kbapi.SecurityDetectionsAPIEqlRule) string {
				return v.Id.String()
			},
		},
	}
}

func (e EqlRuleProcessor) HandlesRuleType(t string) bool {
	return t == "eql"
}

func (e EqlRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return toEqlRuleCreateProps(ctx, client, d)
}

func (e EqlRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return toEqlRuleUpdateProps(ctx, client, d)
}

func toEqlRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	eqlRule := kbapi.SecurityDetectionsAPIEqlRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIEqlRuleCreatePropsType("eql"),
		Query:       d.Query.ValueString(),
		Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguage("eql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &eqlRule.Actions,
		ResponseActions:                   &eqlRule.ResponseActions,
		RuleID:                            &eqlRule.RuleId,
		Enabled:                           &eqlRule.Enabled,
		From:                              &eqlRule.From,
		To:                                &eqlRule.To,
		Interval:                          &eqlRule.Interval,
		Index:                             &eqlRule.Index,
		Author:                            &eqlRule.Author,
		Tags:                              &eqlRule.Tags,
		FalsePositives:                    &eqlRule.FalsePositives,
		References:                        &eqlRule.References,
		License:                           &eqlRule.License,
		Note:                              &eqlRule.Note,
		Setup:                             &eqlRule.Setup,
		MaxSignals:                        &eqlRule.MaxSignals,
		Version:                           &eqlRule.Version,
		ExceptionsList:                    &eqlRule.ExceptionsList,
		AlertSuppression:                  &eqlRule.AlertSuppression,
		RiskScoreMapping:                  &eqlRule.RiskScoreMapping,
		SeverityMapping:                   &eqlRule.SeverityMapping,
		RelatedIntegrations:               &eqlRule.RelatedIntegrations,
		RequiredFields:                    &eqlRule.RequiredFields,
		BuildingBlockType:                 &eqlRule.BuildingBlockType,
		DataViewID:                        &eqlRule.DataViewId,
		Namespace:                         &eqlRule.Namespace,
		RuleNameOverride:                  &eqlRule.RuleNameOverride,
		TimestampOverride:                 &eqlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &eqlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &eqlRule.InvestigationFields,
		Filters:                           &eqlRule.Filters,
		Threat:                            &eqlRule.Threat,
		TimelineID:                        &eqlRule.TimelineId,
		TimelineTitle:                     &eqlRule.TimelineTitle,
	}, &diags, client)

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
func toEqlRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	// Parse ID to get space_id and rule_id
	compID, resourceIDDiags := clients.CompositeIDFromStrFw(d.ID.ValueString())
	diags.Append(resourceIDDiags...)

	uid, err := uuid.Parse(compID.ResourceID)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return updateProps, diags
	}

	eqlRule := kbapi.SecurityDetectionsAPIEqlRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIEqlRuleUpdatePropsType("eql"),
		Query:       d.Query.ValueString(),
		Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguage("eql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		eqlRule.RuleId = &ruleID
		eqlRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &eqlRule.Actions,
		ResponseActions:                   &eqlRule.ResponseActions,
		RuleID:                            &eqlRule.RuleId,
		Enabled:                           &eqlRule.Enabled,
		From:                              &eqlRule.From,
		To:                                &eqlRule.To,
		Interval:                          &eqlRule.Interval,
		Index:                             &eqlRule.Index,
		Author:                            &eqlRule.Author,
		Tags:                              &eqlRule.Tags,
		FalsePositives:                    &eqlRule.FalsePositives,
		References:                        &eqlRule.References,
		License:                           &eqlRule.License,
		Note:                              &eqlRule.Note,
		Setup:                             &eqlRule.Setup,
		MaxSignals:                        &eqlRule.MaxSignals,
		Version:                           &eqlRule.Version,
		ExceptionsList:                    &eqlRule.ExceptionsList,
		AlertSuppression:                  &eqlRule.AlertSuppression,
		RiskScoreMapping:                  &eqlRule.RiskScoreMapping,
		SeverityMapping:                   &eqlRule.SeverityMapping,
		RelatedIntegrations:               &eqlRule.RelatedIntegrations,
		RequiredFields:                    &eqlRule.RequiredFields,
		BuildingBlockType:                 &eqlRule.BuildingBlockType,
		DataViewID:                        &eqlRule.DataViewId,
		Namespace:                         &eqlRule.Namespace,
		RuleNameOverride:                  &eqlRule.RuleNameOverride,
		TimestampOverride:                 &eqlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &eqlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &eqlRule.InvestigationFields,
		Filters:                           &eqlRule.Filters,
		Threat:                            &eqlRule.Threat,
		TimelineID:                        &eqlRule.TimelineId,
		TimelineTitle:                     &eqlRule.TimelineTitle,
	}, &diags, client)

	// Set EQL-specific fields
	if typeutils.IsKnown(d.TiebreakerField) {
		tiebreakerField := d.TiebreakerField.ValueString()
		eqlRule.TiebreakerField = &tiebreakerField
	}

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIEqlRuleUpdateProps(eqlRule)
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
		AlertSuppression:                  rule.AlertSuppression,
		ResponseActions:                   rule.ResponseActions,
	})...)

	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))

	diags.Append(d.updateFiltersFromAPI(ctx, rule.Filters)...)

	if rule.TiebreakerField != nil {
		d.TiebreakerField = types.StringValue(*rule.TiebreakerField)
	} else {
		d.TiebreakerField = types.StringNull()
	}

	return diags
}
