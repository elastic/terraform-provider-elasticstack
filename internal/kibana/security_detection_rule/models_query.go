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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type QueryRuleProcessor struct{}

func (q QueryRuleProcessor) HandlesRuleType(t string) bool {
	return t == "query"
}

func (q QueryRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return toQueryRuleCreateProps(ctx, client, d)
}

func (q QueryRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return toQueryRuleUpdateProps(ctx, client, d)
}

func (q QueryRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPIQueryRule)
	return ok
}

func (q QueryRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *Data) diag.Diagnostics {
	var diags diag.Diagnostics
	value, ok := rule.(kbapi.SecurityDetectionsAPIQueryRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return diags
	}

	return updateFromQueryRule(ctx, &value, d)
}

func (q QueryRuleProcessor) ExtractID(response any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, ok := response.(kbapi.SecurityDetectionsAPIQueryRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return "", diags
	}
	return value.Id.String(), diags
}

func toQueryRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	queryRuleQuery := d.Query.ValueString()
	queryRule := kbapi.SecurityDetectionsAPIQueryRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleCreatePropsType("query"),
		Query:       &queryRuleQuery,
		RiskScore:   int(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &queryRule.Actions,
		ResponseActions:                   &queryRule.ResponseActions,
		RuleID:                            &queryRule.RuleId,
		Enabled:                           &queryRule.Enabled,
		From:                              &queryRule.From,
		To:                                &queryRule.To,
		Interval:                          &queryRule.Interval,
		Index:                             &queryRule.Index,
		Author:                            &queryRule.Author,
		Tags:                              &queryRule.Tags,
		FalsePositives:                    &queryRule.FalsePositives,
		References:                        &queryRule.References,
		License:                           &queryRule.License,
		Note:                              &queryRule.Note,
		Setup:                             &queryRule.Setup,
		MaxSignals:                        &queryRule.MaxSignals,
		Version:                           &queryRule.Version,
		ExceptionsList:                    &queryRule.ExceptionsList,
		AlertSuppression:                  &queryRule.AlertSuppression,
		RiskScoreMapping:                  &queryRule.RiskScoreMapping,
		SeverityMapping:                   &queryRule.SeverityMapping,
		RelatedIntegrations:               &queryRule.RelatedIntegrations,
		RequiredFields:                    &queryRule.RequiredFields,
		BuildingBlockType:                 &queryRule.BuildingBlockType,
		DataViewID:                        &queryRule.DataViewId,
		Namespace:                         &queryRule.Namespace,
		RuleNameOverride:                  &queryRule.RuleNameOverride,
		TimestampOverride:                 &queryRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &queryRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &queryRule.InvestigationFields,
		Filters:                           &queryRule.Filters,
		Threat:                            &queryRule.Threat,
		TimelineID:                        &queryRule.TimelineId,
		TimelineTitle:                     &queryRule.TimelineTitle,
	}, &diags, client)

	// Set query-specific fields
	queryRule.Language = d.getKQLQueryLanguage()

	if typeutils.IsKnown(d.SavedID) {
		savedID := d.SavedID.ValueString()
		queryRule.SavedId = &savedID
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIQueryRuleCreateProps(queryRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert query rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func toQueryRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	queryRuleQuery := d.Query.ValueString()

	// Parse ID to get space_id and rule_id
	compID, resourceIDDiags := clients.CompositeIDFromStrFw(d.ID.ValueString())
	diags.Append(resourceIDDiags...)

	uid, err := uuid.Parse(compID.ResourceID)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return updateProps, diags
	}

	queryRule := kbapi.SecurityDetectionsAPIQueryRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleUpdatePropsType("query"),
		Query:       &queryRuleQuery,
		RiskScore:   int(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		queryRule.RuleId = &ruleID
		queryRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &queryRule.Actions,
		ResponseActions:                   &queryRule.ResponseActions,
		RuleID:                            &queryRule.RuleId,
		Enabled:                           &queryRule.Enabled,
		From:                              &queryRule.From,
		To:                                &queryRule.To,
		Interval:                          &queryRule.Interval,
		Index:                             &queryRule.Index,
		Author:                            &queryRule.Author,
		Tags:                              &queryRule.Tags,
		FalsePositives:                    &queryRule.FalsePositives,
		References:                        &queryRule.References,
		License:                           &queryRule.License,
		Note:                              &queryRule.Note,
		Setup:                             &queryRule.Setup,
		MaxSignals:                        &queryRule.MaxSignals,
		Version:                           &queryRule.Version,
		ExceptionsList:                    &queryRule.ExceptionsList,
		AlertSuppression:                  &queryRule.AlertSuppression,
		RiskScoreMapping:                  &queryRule.RiskScoreMapping,
		SeverityMapping:                   &queryRule.SeverityMapping,
		RelatedIntegrations:               &queryRule.RelatedIntegrations,
		RequiredFields:                    &queryRule.RequiredFields,
		BuildingBlockType:                 &queryRule.BuildingBlockType,
		DataViewID:                        &queryRule.DataViewId,
		Namespace:                         &queryRule.Namespace,
		RuleNameOverride:                  &queryRule.RuleNameOverride,
		TimestampOverride:                 &queryRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &queryRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &queryRule.InvestigationFields,
		Filters:                           &queryRule.Filters,
		Threat:                            &queryRule.Threat,
		TimelineID:                        &queryRule.TimelineId,
		TimelineTitle:                     &queryRule.TimelineTitle,
	}, &diags, client)

	// Set query-specific fields
	queryRule.Language = d.getKQLQueryLanguage()

	if typeutils.IsKnown(d.SavedID) {
		savedID := d.SavedID.ValueString()
		queryRule.SavedId = &savedID
	}

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIQueryRuleUpdateProps(queryRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert query rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}
func updateFromQueryRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIQueryRule, d *Data) diag.Diagnostics {
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
	dataViewIDDiags := d.updateDataViewIDFromAPI(ctx, rule.DataViewId)
	diags.Append(dataViewIDDiags...)

	namespaceDiags := d.updateNamespaceFromAPI(ctx, rule.Namespace)
	diags.Append(namespaceDiags...)

	ruleNameOverrideDiags := d.updateRuleNameOverrideFromAPI(ctx, rule.RuleNameOverride)
	diags.Append(ruleNameOverrideDiags...)

	timestampOverrideDiags := d.updateTimestampOverrideFromAPI(ctx, rule.TimestampOverride)
	diags.Append(timestampOverrideDiags...)

	timestampOverrideFallbackDisabledDiags := d.updateTimestampOverrideFallbackDisabledFromAPI(ctx, rule.TimestampOverrideFallbackDisabled)
	diags.Append(timestampOverrideFallbackDisabledDiags...)

	d.Query = types.StringValue(rule.Query)
	d.Language = typeutils.StringishValue(rule.Language)
	d.Enabled = types.BoolValue(rule.Enabled)
	d.From = types.StringValue(rule.From)
	d.To = types.StringValue(rule.To)
	d.Interval = types.StringValue(rule.Interval)
	d.Description = types.StringValue(rule.Description)
	d.RiskScore = types.Int64Value(int64(rule.RiskScore))
	d.Severity = typeutils.StringishValue(rule.Severity)
	d.MaxSignals = types.Int64Value(int64(rule.MaxSignals))
	d.Version = types.Int64Value(int64(rule.Version))

	// Update building block type
	buildingBlockTypeDiags := d.updateBuildingBlockTypeFromAPI(ctx, rule.BuildingBlockType)
	diags.Append(buildingBlockTypeDiags...)

	// Update read-only fields
	d.CreatedAt = schemautil.TimeToStringValue(rule.CreatedAt)
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = schemautil.TimeToStringValue(rule.UpdatedAt)
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// Update threat
	threatDiags := d.updateThreatFromAPI(ctx, &rule.Threat)
	diags.Append(threatDiags...)

	// Update index patterns
	indexDiags := d.updateIndexFromAPI(ctx, rule.Index)
	diags.Append(indexDiags...)

	// Update author
	authorDiags := d.updateAuthorFromAPI(ctx, rule.Author)
	diags.Append(authorDiags...)

	// Update tags
	tagsDiags := d.updateTagsFromAPI(ctx, rule.Tags)
	diags.Append(tagsDiags...)

	// Update false positives
	falsePositivesDiags := d.updateFalsePositivesFromAPI(ctx, rule.FalsePositives)
	diags.Append(falsePositivesDiags...)

	// Update references
	referencesDiags := d.updateReferencesFromAPI(ctx, rule.References)
	diags.Append(referencesDiags...)

	// Update optional string fields
	licenseDiags := d.updateLicenseFromAPI(ctx, rule.License)
	diags.Append(licenseDiags...)

	noteDiags := d.updateNoteFromAPI(ctx, rule.Note)
	diags.Append(noteDiags...)

	setupDiags := d.updateSetupFromAPI(ctx, rule.Setup)
	diags.Append(setupDiags...)

	// Update actions
	actionDiags := d.updateActionsFromAPI(ctx, rule.Actions)
	diags.Append(actionDiags...)

	// Update exceptions list
	exceptionsListDiags := d.updateExceptionsListFromAPI(ctx, rule.ExceptionsList)
	diags.Append(exceptionsListDiags...)

	// Update risk score mapping
	riskScoreMappingDiags := d.updateRiskScoreMappingFromAPI(ctx, rule.RiskScoreMapping)
	diags.Append(riskScoreMappingDiags...)

	// Update severity mapping
	severityMappingDiags := d.updateSeverityMappingFromAPI(ctx, &rule.SeverityMapping)
	diags.Append(severityMappingDiags...)

	// Update related integrations
	relatedIntegrationsDiags := d.updateRelatedIntegrationsFromAPI(ctx, &rule.RelatedIntegrations)
	diags.Append(relatedIntegrationsDiags...)

	// Update required fields
	requiredFieldsDiags := d.updateRequiredFieldsFromAPI(ctx, &rule.RequiredFields)
	diags.Append(requiredFieldsDiags...)

	// Update investigation fields
	investigationFieldsDiags := d.updateInvestigationFieldsFromAPI(ctx, rule.InvestigationFields)
	diags.Append(investigationFieldsDiags...)

	// Update filters field
	filtersDiags := d.updateFiltersFromAPI(ctx, rule.Filters)
	diags.Append(filtersDiags...)

	// Update alert suppression
	alertSuppressionDiags := d.updateAlertSuppressionFromAPI(ctx, rule.AlertSuppression)
	diags.Append(alertSuppressionDiags...)

	// Update response actions
	responseActionsDiags := d.updateResponseActionsFromAPI(ctx, rule.ResponseActions)
	diags.Append(responseActionsDiags...)

	return diags
}
