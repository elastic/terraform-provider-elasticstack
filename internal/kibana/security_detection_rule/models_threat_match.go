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

type ThreatMatchRuleProcessor struct{}

func (t ThreatMatchRuleProcessor) HandlesRuleType(ruleType string) bool {
	return ruleType == "threat_match"
}

func (t ThreatMatchRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toThreatMatchRuleCreateProps(ctx, client)
}

func (t ThreatMatchRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toThreatMatchRuleUpdateProps(ctx, client)
}

func (t ThreatMatchRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPIThreatMatchRule)
	return ok
}

func (t ThreatMatchRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *Data) diag.Diagnostics {
	var diags diag.Diagnostics
	value, ok := rule.(kbapi.SecurityDetectionsAPIThreatMatchRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return diags
	}

	return d.updateFromThreatMatchRule(ctx, &value)
}

func (t ThreatMatchRuleProcessor) ExtractID(response any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, ok := response.(kbapi.SecurityDetectionsAPIThreatMatchRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return "", diags
	}
	return value.Id.String(), diags
}

func (d Data) toThreatMatchRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	threatMatchRule := kbapi.SecurityDetectionsAPIThreatMatchRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIThreatMatchRuleCreatePropsType("threat_match"),
		Query:       d.Query.ValueString(),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set threat index
	if typeutils.IsKnown(d.ThreatIndex) {
		threatIndex := typeutils.ListTypeAs[string](ctx, d.ThreatIndex, path.Root("threat_index"), &diags)
		if !diags.HasError() {
			threatMatchRule.ThreatIndex = threatIndex
		}
	}

	if typeutils.IsKnown(d.ThreatMapping) && len(d.ThreatMapping.Elements()) > 0 {
		apiThreatMapping, threatMappingDiags := d.threatMappingToAPI(ctx)
		if !threatMappingDiags.HasError() {
			threatMatchRule.ThreatMapping = apiThreatMapping
		}
		diags.Append(threatMappingDiags...)
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &threatMatchRule.Actions,
		ResponseActions:                   &threatMatchRule.ResponseActions,
		RuleID:                            &threatMatchRule.RuleId,
		Enabled:                           &threatMatchRule.Enabled,
		From:                              &threatMatchRule.From,
		To:                                &threatMatchRule.To,
		Interval:                          &threatMatchRule.Interval,
		Index:                             &threatMatchRule.Index,
		Author:                            &threatMatchRule.Author,
		Tags:                              &threatMatchRule.Tags,
		FalsePositives:                    &threatMatchRule.FalsePositives,
		References:                        &threatMatchRule.References,
		License:                           &threatMatchRule.License,
		Note:                              &threatMatchRule.Note,
		Setup:                             &threatMatchRule.Setup,
		MaxSignals:                        &threatMatchRule.MaxSignals,
		Version:                           &threatMatchRule.Version,
		ExceptionsList:                    &threatMatchRule.ExceptionsList,
		AlertSuppression:                  &threatMatchRule.AlertSuppression,
		RiskScoreMapping:                  &threatMatchRule.RiskScoreMapping,
		SeverityMapping:                   &threatMatchRule.SeverityMapping,
		RelatedIntegrations:               &threatMatchRule.RelatedIntegrations,
		RequiredFields:                    &threatMatchRule.RequiredFields,
		BuildingBlockType:                 &threatMatchRule.BuildingBlockType,
		DataViewID:                        &threatMatchRule.DataViewId,
		Namespace:                         &threatMatchRule.Namespace,
		RuleNameOverride:                  &threatMatchRule.RuleNameOverride,
		TimestampOverride:                 &threatMatchRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &threatMatchRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &threatMatchRule.InvestigationFields,
		Filters:                           &threatMatchRule.Filters,
		Threat:                            &threatMatchRule.Threat,
		TimelineID:                        &threatMatchRule.TimelineId,
		TimelineTitle:                     &threatMatchRule.TimelineTitle,
	}, &diags, client)

	// Set threat-specific fields
	if typeutils.IsKnown(d.ThreatQuery) {
		threatMatchRule.ThreatQuery = d.ThreatQuery.ValueString()
	}

	if typeutils.IsKnown(d.ThreatIndicatorPath) {
		threatIndicatorPath := d.ThreatIndicatorPath.ValueString()
		threatMatchRule.ThreatIndicatorPath = &threatIndicatorPath
	}

	if typeutils.IsKnown(d.ConcurrentSearches) {
		concurrentSearches := kbapi.SecurityDetectionsAPIConcurrentSearches(d.ConcurrentSearches.ValueInt64())
		threatMatchRule.ConcurrentSearches = &concurrentSearches
	}

	if typeutils.IsKnown(d.ItemsPerSearch) {
		itemsPerSearch := kbapi.SecurityDetectionsAPIItemsPerSearch(d.ItemsPerSearch.ValueInt64())
		threatMatchRule.ItemsPerSearch = &itemsPerSearch
	}

	// Set query language
	threatMatchRule.Language = d.getKQLQueryLanguage()

	if typeutils.IsKnown(d.SavedID) {
		savedID := d.SavedID.ValueString()
		threatMatchRule.SavedId = &savedID
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIThreatMatchRuleCreateProps(threatMatchRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert threat match rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}
func (d Data) toThreatMatchRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	threatMatchRule := kbapi.SecurityDetectionsAPIThreatMatchRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIThreatMatchRuleUpdatePropsType("threat_match"),
		Query:       d.Query.ValueString(),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		threatMatchRule.RuleId = &ruleID
		threatMatchRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set threat index
	if typeutils.IsKnown(d.ThreatIndex) {
		threatIndex := typeutils.ListTypeAs[string](ctx, d.ThreatIndex, path.Root("threat_index"), &diags)
		if !diags.HasError() {
			threatMatchRule.ThreatIndex = threatIndex
		}
	}

	if typeutils.IsKnown(d.ThreatMapping) && len(d.ThreatMapping.Elements()) > 0 {
		apiThreatMapping, threatMappingDiags := d.threatMappingToAPI(ctx)
		if !threatMappingDiags.HasError() {
			threatMatchRule.ThreatMapping = apiThreatMapping
		}
		diags.Append(threatMappingDiags...)
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &threatMatchRule.Actions,
		ResponseActions:                   &threatMatchRule.ResponseActions,
		RuleID:                            &threatMatchRule.RuleId,
		Enabled:                           &threatMatchRule.Enabled,
		From:                              &threatMatchRule.From,
		To:                                &threatMatchRule.To,
		Interval:                          &threatMatchRule.Interval,
		Index:                             &threatMatchRule.Index,
		Author:                            &threatMatchRule.Author,
		Tags:                              &threatMatchRule.Tags,
		FalsePositives:                    &threatMatchRule.FalsePositives,
		References:                        &threatMatchRule.References,
		License:                           &threatMatchRule.License,
		Note:                              &threatMatchRule.Note,
		InvestigationFields:               &threatMatchRule.InvestigationFields,
		Setup:                             &threatMatchRule.Setup,
		MaxSignals:                        &threatMatchRule.MaxSignals,
		Version:                           &threatMatchRule.Version,
		ExceptionsList:                    &threatMatchRule.ExceptionsList,
		AlertSuppression:                  &threatMatchRule.AlertSuppression,
		RiskScoreMapping:                  &threatMatchRule.RiskScoreMapping,
		SeverityMapping:                   &threatMatchRule.SeverityMapping,
		RelatedIntegrations:               &threatMatchRule.RelatedIntegrations,
		RequiredFields:                    &threatMatchRule.RequiredFields,
		BuildingBlockType:                 &threatMatchRule.BuildingBlockType,
		DataViewID:                        &threatMatchRule.DataViewId,
		Namespace:                         &threatMatchRule.Namespace,
		RuleNameOverride:                  &threatMatchRule.RuleNameOverride,
		TimestampOverride:                 &threatMatchRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &threatMatchRule.TimestampOverrideFallbackDisabled,
		Filters:                           &threatMatchRule.Filters,
		Threat:                            &threatMatchRule.Threat,
		TimelineID:                        &threatMatchRule.TimelineId,
		TimelineTitle:                     &threatMatchRule.TimelineTitle,
	}, &diags, client)

	// Set threat-specific fields
	if typeutils.IsKnown(d.ThreatQuery) {
		threatMatchRule.ThreatQuery = d.ThreatQuery.ValueString()
	}

	if typeutils.IsKnown(d.ThreatIndicatorPath) {
		threatIndicatorPath := d.ThreatIndicatorPath.ValueString()
		threatMatchRule.ThreatIndicatorPath = &threatIndicatorPath
	}

	if typeutils.IsKnown(d.ConcurrentSearches) {
		concurrentSearches := kbapi.SecurityDetectionsAPIConcurrentSearches(d.ConcurrentSearches.ValueInt64())
		threatMatchRule.ConcurrentSearches = &concurrentSearches
	}

	if typeutils.IsKnown(d.ItemsPerSearch) {
		itemsPerSearch := kbapi.SecurityDetectionsAPIItemsPerSearch(d.ItemsPerSearch.ValueInt64())
		threatMatchRule.ItemsPerSearch = &itemsPerSearch
	}

	// Set query language
	threatMatchRule.Language = d.getKQLQueryLanguage()

	if typeutils.IsKnown(d.SavedID) {
		savedID := d.SavedID.ValueString()
		threatMatchRule.SavedId = &savedID
	}

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIThreatMatchRuleUpdateProps(threatMatchRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert threat match rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d *Data) updateFromThreatMatchRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIThreatMatchRule) diag.Diagnostics {
	var diags diag.Diagnostics

	compID := clients.CompositeID{
		ClusterID:  d.SpaceID.ValueString(),
		ResourceID: rule.Id.String(),
	}
	d.ID = types.StringValue(compID.String())

	d.RuleID = types.StringValue(rule.RuleId)
	d.Name = types.StringValue(rule.Name)
	d.Type = types.StringValue(string(rule.Type))

	// Update common fields
	diags.Append(d.updateTimelineIDFromAPI(ctx, rule.TimelineId)...)
	diags.Append(d.updateTimelineTitleFromAPI(ctx, rule.TimelineTitle)...)
	diags.Append(d.updateDataViewIDFromAPI(ctx, rule.DataViewId)...)
	diags.Append(d.updateNamespaceFromAPI(ctx, rule.Namespace)...)
	diags.Append(d.updateRuleNameOverrideFromAPI(ctx, rule.RuleNameOverride)...)
	diags.Append(d.updateTimestampOverrideFromAPI(ctx, rule.TimestampOverride)...)
	diags.Append(d.updateTimestampOverrideFallbackDisabledFromAPI(ctx, rule.TimestampOverrideFallbackDisabled)...)

	// Update building block type
	diags.Append(d.updateBuildingBlockTypeFromAPI(ctx, rule.BuildingBlockType)...)
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

	// Update read-only fields
	d.CreatedAt = types.StringValue(rule.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = types.StringValue(rule.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// Update threat
	threatDiags := d.updateThreatFromAPI(ctx, &rule.Threat)
	diags.Append(threatDiags...)

	// Update index patterns
	diags.Append(d.updateIndexFromAPI(ctx, rule.Index)...)

	// Threat Match-specific fields
	d.ThreatQuery = types.StringValue(rule.ThreatQuery)
	if len(rule.ThreatIndex) > 0 {
		d.ThreatIndex = typeutils.ListValueFrom(ctx, rule.ThreatIndex, types.StringType, path.Root("threat_index"), &diags)
	} else {
		d.ThreatIndex = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if rule.ThreatIndicatorPath != nil {
		d.ThreatIndicatorPath = types.StringValue(*rule.ThreatIndicatorPath)
	} else {
		d.ThreatIndicatorPath = types.StringNull()
	}

	if rule.ConcurrentSearches != nil {
		d.ConcurrentSearches = types.Int64Value(int64(*rule.ConcurrentSearches))
	} else {
		d.ConcurrentSearches = types.Int64Null()
	}

	if rule.ItemsPerSearch != nil {
		d.ItemsPerSearch = types.Int64Value(int64(*rule.ItemsPerSearch))
	} else {
		d.ItemsPerSearch = types.Int64Null()
	}

	// Optional saved query ID
	if rule.SavedId != nil {
		d.SavedID = types.StringValue(*rule.SavedId)
	} else {
		d.SavedID = types.StringNull()
	}

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

	// Convert threat mapping
	if len(rule.ThreatMapping) > 0 {
		listValue, threatMappingDiags := convertThreatMappingToModel(ctx, rule.ThreatMapping)
		diags.Append(threatMappingDiags...)
		if !threatMappingDiags.HasError() {
			d.ThreatMapping = listValue
		}
	}

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
