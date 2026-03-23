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

type SavedQueryRuleProcessor struct{}

func (s SavedQueryRuleProcessor) HandlesRuleType(t string) bool {
	return t == "saved_query"
}

func (s SavedQueryRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toSavedQueryRuleCreateProps(ctx, client)
}

func (s SavedQueryRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toSavedQueryRuleUpdateProps(ctx, client)
}

func (s SavedQueryRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPISavedQueryRule)
	return ok
}

func (s SavedQueryRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *Data) diag.Diagnostics {
	var diags diag.Diagnostics
	value, ok := rule.(kbapi.SecurityDetectionsAPISavedQueryRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return diags
	}

	return d.updateFromSavedQueryRule(ctx, &value)
}

func (s SavedQueryRuleProcessor) ExtractID(response any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, ok := response.(kbapi.SecurityDetectionsAPISavedQueryRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return "", diags
	}
	return value.Id.String(), diags
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

	compID := clients.CompositeID{
		ClusterID:  d.SpaceID.ValueString(),
		ResourceID: rule.Id.String(),
	}
	d.ID = types.StringValue(compID.String())

	d.RuleID = types.StringValue(rule.RuleId)
	d.Name = types.StringValue(rule.Name)
	d.Type = typeutils.StringishValue(rule.Type)

	// Update common fields
	diags.Append(d.updateTimelineIDFromAPI(ctx, rule.TimelineId)...)
	diags.Append(d.updateTimelineTitleFromAPI(ctx, rule.TimelineTitle)...)
	diags.Append(d.updateDataViewIDFromAPI(ctx, rule.DataViewId)...)
	diags.Append(d.updateNamespaceFromAPI(ctx, rule.Namespace)...)
	diags.Append(d.updateRuleNameOverrideFromAPI(ctx, rule.RuleNameOverride)...)
	diags.Append(d.updateTimestampOverrideFromAPI(ctx, rule.TimestampOverride)...)
	diags.Append(d.updateTimestampOverrideFallbackDisabledFromAPI(ctx, rule.TimestampOverrideFallbackDisabled)...)

	d.SavedID = types.StringValue(rule.SavedId)
	d.Enabled = types.BoolValue(rule.Enabled)
	d.From = types.StringValue(rule.From)

	// Update building block type
	diags.Append(d.updateBuildingBlockTypeFromAPI(ctx, rule.BuildingBlockType)...)
	d.To = types.StringValue(rule.To)
	d.Interval = types.StringValue(rule.Interval)
	d.Description = types.StringValue(rule.Description)
	d.RiskScore = types.Int64Value(int64(rule.RiskScore))
	d.Severity = typeutils.StringishValue(rule.Severity)
	d.MaxSignals = types.Int64Value(int64(rule.MaxSignals))
	d.Version = types.Int64Value(int64(rule.Version))

	// Update read-only fields
	d.CreatedAt = types.StringValue(rule.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = types.StringValue(rule.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// Update index patterns
	diags.Append(d.updateIndexFromAPI(ctx, rule.Index)...)

	// Optional query for saved query rules
	d.Query = types.StringPointerValue(rule.Query)

	// Language for saved query rules (not a pointer)
	d.Language = typeutils.StringishValue(rule.Language)

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

	// Update filters field
	filtersDiags := d.updateFiltersFromAPI(ctx, rule.Filters)
	diags.Append(filtersDiags...)

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
