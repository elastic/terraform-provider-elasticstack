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

type ThresholdRuleProcessor struct{}

func (th ThresholdRuleProcessor) HandlesRuleType(t string) bool {
	return t == "threshold"
}

func (th ThresholdRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toThresholdRuleCreateProps(ctx, client)
}

func (th ThresholdRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toThresholdRuleUpdateProps(ctx, client)
}

func (th ThresholdRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPIThresholdRule)
	return ok
}

func (th ThresholdRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *Data) diag.Diagnostics {
	var diags diag.Diagnostics
	value, ok := rule.(kbapi.SecurityDetectionsAPIThresholdRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return diags
	}

	return d.updateFromThresholdRule(ctx, &value)
}

func (th ThresholdRuleProcessor) ExtractID(response any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, ok := response.(kbapi.SecurityDetectionsAPIThresholdRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return "", diags
	}
	return value.Id.String(), diags
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

	// Parse ID to get space_id and rule_id
	compID, resourceIDDiags := clients.CompositeIDFromStrFw(d.ID.ValueString())
	diags.Append(resourceIDDiags...)

	uid, err := uuid.Parse(compID.ResourceID)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
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
	err = updateProps.FromSecurityDetectionsAPIThresholdRuleUpdateProps(thresholdRule)
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

	d.Query = typeutils.StringishValue(rule.Query)
	d.Language = typeutils.StringishValue(rule.Language)
	d.Enabled = types.BoolValue(rule.Enabled)

	// Update building block type
	diags.Append(d.updateBuildingBlockTypeFromAPI(ctx, rule.BuildingBlockType)...)
	d.From = types.StringValue(rule.From)
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

	// Update threat
	threatDiags := d.updateThreatFromAPI(ctx, &rule.Threat)
	diags.Append(threatDiags...)

	// Update index patterns
	diags.Append(d.updateIndexFromAPI(ctx, rule.Index)...)

	// Threshold-specific fields
	thresholdObj, thresholdDiags := convertThresholdToModel(ctx, rule.Threshold)
	diags.Append(thresholdDiags...)
	if !thresholdDiags.HasError() {
		d.Threshold = thresholdObj
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
	thresholdAlertSuppressionDiags := d.updateThresholdAlertSuppressionFromAPI(ctx, rule.AlertSuppression)
	diags.Append(thresholdAlertSuppressionDiags...)

	// Update response actions
	responseActionsDiags := d.updateResponseActionsFromAPI(ctx, rule.ResponseActions)
	diags.Append(responseActionsDiags...)

	return diags
}
