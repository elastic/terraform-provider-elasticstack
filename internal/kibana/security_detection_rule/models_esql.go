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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EsqlRuleProcessor struct{}

func (e EsqlRuleProcessor) HandlesRuleType(t string) bool {
	return t == "esql"
}

func (e EsqlRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toEsqlRuleCreateProps(ctx, client)
}

func (e EsqlRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toEsqlRuleUpdateProps(ctx, client)
}

func (e EsqlRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPIEsqlRule)
	return ok
}

func (e EsqlRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *Data) diag.Diagnostics {
	var diags diag.Diagnostics
	value, ok := rule.(kbapi.SecurityDetectionsAPIEsqlRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return diags
	}

	return d.updateFromEsqlRule(ctx, &value)
}

func (e EsqlRuleProcessor) ExtractID(response any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, ok := response.(kbapi.SecurityDetectionsAPIEsqlRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return "", diags
	}
	return value.Id.String(), diags
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

	// Parse ID to get space_id and rule_id
	compID, resourceIDDiags := clients.CompositeIDFromStrFw(d.ID.ValueString())
	diags.Append(resourceIDDiags...)

	uid, err := uuid.Parse(compID.ResourceID)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
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
	err = updateProps.FromSecurityDetectionsAPIEsqlRuleUpdateProps(esqlRule)
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

	compID := clients.CompositeID{
		ClusterID:  d.SpaceID.ValueString(),
		ResourceID: rule.Id.String(),
	}
	d.ID = types.StringValue(compID.String())

	d.RuleID = types.StringValue(rule.RuleId)
	d.Name = types.StringValue(rule.Name)
	d.Type = types.StringValue(string(rule.Type))

	// Update common fields (ESQL doesn't support DataViewID)
	d.DataViewID = types.StringNull()
	diags.Append(d.updateTimelineIDFromAPI(ctx, rule.TimelineId)...)
	diags.Append(d.updateTimelineTitleFromAPI(ctx, rule.TimelineTitle)...)
	diags.Append(d.updateNamespaceFromAPI(ctx, rule.Namespace)...)
	diags.Append(d.updateRuleNameOverrideFromAPI(ctx, rule.RuleNameOverride)...)
	diags.Append(d.updateTimestampOverrideFromAPI(ctx, rule.TimestampOverride)...)
	diags.Append(d.updateTimestampOverrideFallbackDisabledFromAPI(ctx, rule.TimestampOverrideFallbackDisabled)...)

	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))
	d.Enabled = types.BoolValue(rule.Enabled)
	d.From = types.StringValue(rule.From)
	d.To = types.StringValue(rule.To)
	d.Interval = types.StringValue(rule.Interval)
	d.Description = types.StringValue(rule.Description)
	d.RiskScore = types.Int64Value(int64(rule.RiskScore))
	d.Severity = types.StringValue(string(rule.Severity))
	d.MaxSignals = types.Int64Value(int64(rule.MaxSignals))
	d.Version = types.Int64Value(int64(rule.Version))

	// Update building block type
	diags.Append(d.updateBuildingBlockTypeFromAPI(ctx, rule.BuildingBlockType)...)

	// Update read-only fields
	d.CreatedAt = types.StringValue(rule.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = types.StringValue(rule.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// ESQL rules don't use index patterns
	d.Index = types.ListValueMust(types.StringType, []attr.Value{})

	// Update author
	diags.Append(d.updateAuthorFromAPI(ctx, rule.Author)...)

	// Update tags
	diags.Append(d.updateTagsFromAPI(ctx, rule.Tags)...)

	// Update false positives
	diags.Append(d.updateFalsePositivesFromAPI(ctx, rule.FalsePositives)...)

	// Update references
	diags.Append(d.updateReferencesFromAPI(ctx, rule.References)...)

	// Update optional string fields
	diags.Append(d.updateLicenseFromAPI(ctx, rule.License)...)
	diags.Append(d.updateNoteFromAPI(ctx, rule.Note)...)
	diags.Append(d.updateSetupFromAPI(ctx, rule.Setup)...)

	// Update actions
	actionDiags := d.updateActionsFromAPI(ctx, rule.Actions)
	diags.Append(actionDiags...)

	// Update exceptions list
	exceptionsListDiags := d.updateExceptionsListFromAPI(ctx, rule.ExceptionsList)
	diags.Append(exceptionsListDiags...)

	// Update risk score mapping
	riskScoreMappingDiags := d.updateRiskScoreMappingFromAPI(ctx, rule.RiskScoreMapping)
	diags.Append(riskScoreMappingDiags...)

	// Update investigation fields
	investigationFieldsDiags := d.updateInvestigationFieldsFromAPI(ctx, rule.InvestigationFields)
	diags.Append(investigationFieldsDiags...)

	// Update threat
	threatDiags := d.updateThreatFromAPI(ctx, &rule.Threat)
	diags.Append(threatDiags...)

	// Update severity mapping
	severityMappingDiags := d.updateSeverityMappingFromAPI(ctx, &rule.SeverityMapping)
	diags.Append(severityMappingDiags...)

	// Update related integrations
	relatedIntegrationsDiags := d.updateRelatedIntegrationsFromAPI(ctx, &rule.RelatedIntegrations)
	diags.Append(relatedIntegrationsDiags...)

	// Update required fields
	requiredFieldsDiags := d.updateRequiredFieldsFromAPI(ctx, &rule.RequiredFields)
	diags.Append(requiredFieldsDiags...)

	// Update alert suppression
	alertSuppressionDiags := d.updateAlertSuppressionFromAPI(ctx, rule.AlertSuppression)
	diags.Append(alertSuppressionDiags...)

	// Update response actions
	responseActionsDiags := d.updateResponseActionsFromAPI(ctx, rule.ResponseActions)
	diags.Append(responseActionsDiags...)

	return diags
}
