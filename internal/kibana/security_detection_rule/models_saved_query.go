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

type SavedQueryRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPISavedQueryRule]
}

func newSavedQueryRuleProcessor() SavedQueryRuleProcessor {
	return SavedQueryRuleProcessor{
		baseRuleProcessor: baseRuleProcessor[kbapi.SecurityDetectionsAPISavedQueryRule]{
			updateFn: func(ctx context.Context, v *kbapi.SecurityDetectionsAPISavedQueryRule, d *Data) diag.Diagnostics {
				return d.updateFromSavedQueryRule(ctx, v)
			},
			idFn: func(v kbapi.SecurityDetectionsAPISavedQueryRule) string {
				return v.Id.String()
			},
		},
	}
}

func (s SavedQueryRuleProcessor) HandlesRuleType(t string) bool {
	return t == "saved_query"
}

func (s SavedQueryRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toSavedQueryRuleCreateProps(ctx, client)
}

func (s SavedQueryRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toSavedQueryRuleUpdateProps(ctx, client)
}

func (d Data) toSavedQueryRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	savedQueryRule := kbapi.SecurityDetectionsAPISavedQueryRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPISavedQueryRuleCreatePropsType("saved_query"),
		SavedId:     d.SavedID.ValueString(),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &savedQueryRule.Actions,
		ResponseActions:                   &savedQueryRule.ResponseActions,
		RuleID:                            &savedQueryRule.RuleId,
		Enabled:                           &savedQueryRule.Enabled,
		From:                              &savedQueryRule.From,
		To:                                &savedQueryRule.To,
		Interval:                          &savedQueryRule.Interval,
		Index:                             &savedQueryRule.Index,
		Author:                            &savedQueryRule.Author,
		Tags:                              &savedQueryRule.Tags,
		FalsePositives:                    &savedQueryRule.FalsePositives,
		References:                        &savedQueryRule.References,
		License:                           &savedQueryRule.License,
		Note:                              &savedQueryRule.Note,
		Setup:                             &savedQueryRule.Setup,
		MaxSignals:                        &savedQueryRule.MaxSignals,
		Version:                           &savedQueryRule.Version,
		ExceptionsList:                    &savedQueryRule.ExceptionsList,
		AlertSuppression:                  &savedQueryRule.AlertSuppression,
		RiskScoreMapping:                  &savedQueryRule.RiskScoreMapping,
		SeverityMapping:                   &savedQueryRule.SeverityMapping,
		RelatedIntegrations:               &savedQueryRule.RelatedIntegrations,
		RequiredFields:                    &savedQueryRule.RequiredFields,
		BuildingBlockType:                 &savedQueryRule.BuildingBlockType,
		DataViewID:                        &savedQueryRule.DataViewId,
		Namespace:                         &savedQueryRule.Namespace,
		RuleNameOverride:                  &savedQueryRule.RuleNameOverride,
		TimestampOverride:                 &savedQueryRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &savedQueryRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &savedQueryRule.InvestigationFields,
		Filters:                           &savedQueryRule.Filters,
		Threat:                            &savedQueryRule.Threat,
		TimelineID:                        &savedQueryRule.TimelineId,
		TimelineTitle:                     &savedQueryRule.TimelineTitle,
	}, &diags, client)

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
func (d Data) toSavedQueryRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	savedQueryRule := kbapi.SecurityDetectionsAPISavedQueryRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPISavedQueryRuleUpdatePropsType("saved_query"),
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

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &savedQueryRule.Actions,
		ResponseActions:                   &savedQueryRule.ResponseActions,
		RuleID:                            &savedQueryRule.RuleId,
		Enabled:                           &savedQueryRule.Enabled,
		From:                              &savedQueryRule.From,
		To:                                &savedQueryRule.To,
		Interval:                          &savedQueryRule.Interval,
		Index:                             &savedQueryRule.Index,
		Author:                            &savedQueryRule.Author,
		Tags:                              &savedQueryRule.Tags,
		FalsePositives:                    &savedQueryRule.FalsePositives,
		References:                        &savedQueryRule.References,
		License:                           &savedQueryRule.License,
		Note:                              &savedQueryRule.Note,
		InvestigationFields:               &savedQueryRule.InvestigationFields,
		Setup:                             &savedQueryRule.Setup,
		MaxSignals:                        &savedQueryRule.MaxSignals,
		Version:                           &savedQueryRule.Version,
		ExceptionsList:                    &savedQueryRule.ExceptionsList,
		AlertSuppression:                  &savedQueryRule.AlertSuppression,
		RiskScoreMapping:                  &savedQueryRule.RiskScoreMapping,
		SeverityMapping:                   &savedQueryRule.SeverityMapping,
		RelatedIntegrations:               &savedQueryRule.RelatedIntegrations,
		RequiredFields:                    &savedQueryRule.RequiredFields,
		BuildingBlockType:                 &savedQueryRule.BuildingBlockType,
		DataViewID:                        &savedQueryRule.DataViewId,
		Namespace:                         &savedQueryRule.Namespace,
		RuleNameOverride:                  &savedQueryRule.RuleNameOverride,
		TimestampOverride:                 &savedQueryRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &savedQueryRule.TimestampOverrideFallbackDisabled,
		Filters:                           &savedQueryRule.Filters,
		Threat:                            &savedQueryRule.Threat,
		TimelineID:                        &savedQueryRule.TimelineId,
		TimelineTitle:                     &savedQueryRule.TimelineTitle,
	}, &diags, client)

	// Set optional query for saved query rules
	if typeutils.IsKnown(d.Query) {
		query := d.Query.ValueString()
		savedQueryRule.Query = &query
	}

	// Set query language
	savedQueryRule.Language = d.getKQLQueryLanguage()

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPISavedQueryRuleUpdateProps(savedQueryRule)
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

	d.SavedID = types.StringValue(rule.SavedId)
	d.Query = types.StringPointerValue(rule.Query)
	d.Language = typeutils.StringishValue(rule.Language)

	diags.Append(d.updateFiltersFromAPI(ctx, rule.Filters)...)

	return diags
}
