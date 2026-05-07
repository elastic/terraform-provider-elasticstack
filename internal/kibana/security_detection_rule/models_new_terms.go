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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NewTermsRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPINewTermsRule]
}

func newNewTermsRuleProcessor() NewTermsRuleProcessor {
	return NewTermsRuleProcessor{
		baseRuleProcessor: baseRuleProcessor[kbapi.SecurityDetectionsAPINewTermsRule]{
			updateFn: func(ctx context.Context, v *kbapi.SecurityDetectionsAPINewTermsRule, d *Data) diag.Diagnostics {
				return d.updateFromNewTermsRule(ctx, v)
			},
			idFn: func(v kbapi.SecurityDetectionsAPINewTermsRule) string {
				return v.Id.String()
			},
		},
	}
}

func (n NewTermsRuleProcessor) HandlesRuleType(t string) bool {
	return t == "new_terms"
}

func (n NewTermsRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toNewTermsRuleCreateProps(ctx, client)
}

func (n NewTermsRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toNewTermsRuleUpdateProps(ctx, client)
}

func (d Data) toNewTermsRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	newTermsRule := kbapi.SecurityDetectionsAPINewTermsRuleCreateProps{
		Name:               d.Name.ValueString(),
		Description:        d.Description.ValueString(),
		Type:               kbapi.SecurityDetectionsAPINewTermsRuleCreatePropsType("new_terms"),
		Query:              d.Query.ValueString(),
		HistoryWindowStart: d.HistoryWindowStart.ValueString(),
		RiskScore:          kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:           kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set new terms fields
	if typeutils.IsKnown(d.NewTermsFields) {
		newTermsFields := typeutils.ListTypeAs[string](ctx, d.NewTermsFields, path.Root("new_terms_fields"), &diags)
		if !diags.HasError() {
			newTermsRule.NewTermsFields = newTermsFields
		}
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &newTermsRule.Actions,
		ResponseActions:                   &newTermsRule.ResponseActions,
		RuleID:                            &newTermsRule.RuleId,
		Enabled:                           &newTermsRule.Enabled,
		From:                              &newTermsRule.From,
		To:                                &newTermsRule.To,
		Interval:                          &newTermsRule.Interval,
		Index:                             &newTermsRule.Index,
		Author:                            &newTermsRule.Author,
		Tags:                              &newTermsRule.Tags,
		FalsePositives:                    &newTermsRule.FalsePositives,
		References:                        &newTermsRule.References,
		License:                           &newTermsRule.License,
		Note:                              &newTermsRule.Note,
		Setup:                             &newTermsRule.Setup,
		MaxSignals:                        &newTermsRule.MaxSignals,
		Version:                           &newTermsRule.Version,
		ExceptionsList:                    &newTermsRule.ExceptionsList,
		AlertSuppression:                  &newTermsRule.AlertSuppression,
		RiskScoreMapping:                  &newTermsRule.RiskScoreMapping,
		SeverityMapping:                   &newTermsRule.SeverityMapping,
		RelatedIntegrations:               &newTermsRule.RelatedIntegrations,
		RequiredFields:                    &newTermsRule.RequiredFields,
		BuildingBlockType:                 &newTermsRule.BuildingBlockType,
		DataViewID:                        &newTermsRule.DataViewId,
		Namespace:                         &newTermsRule.Namespace,
		RuleNameOverride:                  &newTermsRule.RuleNameOverride,
		TimestampOverride:                 &newTermsRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &newTermsRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &newTermsRule.InvestigationFields,
		Filters:                           &newTermsRule.Filters,
		Threat:                            &newTermsRule.Threat,
		TimelineID:                        &newTermsRule.TimelineId,
		TimelineTitle:                     &newTermsRule.TimelineTitle,
	}, &diags, client)

	// Set query language
	newTermsRule.Language = d.getKQLQueryLanguage()

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPINewTermsRuleCreateProps(newTermsRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert new terms rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}
func (d Data) toNewTermsRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	newTermsRule := kbapi.SecurityDetectionsAPINewTermsRuleUpdateProps{
		Id:                 &uid,
		Name:               d.Name.ValueString(),
		Description:        d.Description.ValueString(),
		Type:               kbapi.SecurityDetectionsAPINewTermsRuleUpdatePropsType("new_terms"),
		Query:              d.Query.ValueString(),
		HistoryWindowStart: d.HistoryWindowStart.ValueString(),
		RiskScore:          kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:           kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		newTermsRule.RuleId = &ruleID
		newTermsRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set new terms fields
	if typeutils.IsKnown(d.NewTermsFields) {
		newTermsFields := typeutils.ListTypeAs[string](ctx, d.NewTermsFields, path.Root("new_terms_fields"), &diags)
		if !diags.HasError() {
			newTermsRule.NewTermsFields = newTermsFields
		}
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &newTermsRule.Actions,
		ResponseActions:                   &newTermsRule.ResponseActions,
		RuleID:                            &newTermsRule.RuleId,
		Enabled:                           &newTermsRule.Enabled,
		From:                              &newTermsRule.From,
		To:                                &newTermsRule.To,
		Interval:                          &newTermsRule.Interval,
		Index:                             &newTermsRule.Index,
		Author:                            &newTermsRule.Author,
		Tags:                              &newTermsRule.Tags,
		FalsePositives:                    &newTermsRule.FalsePositives,
		References:                        &newTermsRule.References,
		License:                           &newTermsRule.License,
		Note:                              &newTermsRule.Note,
		InvestigationFields:               &newTermsRule.InvestigationFields,
		Setup:                             &newTermsRule.Setup,
		MaxSignals:                        &newTermsRule.MaxSignals,
		Version:                           &newTermsRule.Version,
		ExceptionsList:                    &newTermsRule.ExceptionsList,
		AlertSuppression:                  &newTermsRule.AlertSuppression,
		RiskScoreMapping:                  &newTermsRule.RiskScoreMapping,
		SeverityMapping:                   &newTermsRule.SeverityMapping,
		RelatedIntegrations:               &newTermsRule.RelatedIntegrations,
		RequiredFields:                    &newTermsRule.RequiredFields,
		BuildingBlockType:                 &newTermsRule.BuildingBlockType,
		DataViewID:                        &newTermsRule.DataViewId,
		Namespace:                         &newTermsRule.Namespace,
		RuleNameOverride:                  &newTermsRule.RuleNameOverride,
		TimestampOverride:                 &newTermsRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &newTermsRule.TimestampOverrideFallbackDisabled,
		Filters:                           &newTermsRule.Filters,
		Threat:                            &newTermsRule.Threat,
		TimelineID:                        &newTermsRule.TimelineId,
		TimelineTitle:                     &newTermsRule.TimelineTitle,
	}, &diags, client)

	// Set query language
	newTermsRule.Language = d.getKQLQueryLanguage()

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPINewTermsRuleUpdateProps(newTermsRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert new terms rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}
func (d *Data) updateFromNewTermsRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPINewTermsRule) diag.Diagnostics {
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

	// New Terms-specific fields
	d.HistoryWindowStart = types.StringValue(rule.HistoryWindowStart)
	if len(rule.NewTermsFields) > 0 {
		d.NewTermsFields = typeutils.ListValueFrom(ctx, rule.NewTermsFields, types.StringType, path.Root("new_terms_fields"), &diags)
	} else {
		d.NewTermsFields = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}
