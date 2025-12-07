package security_detection_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EqlRuleProcessor struct{}

func (e EqlRuleProcessor) HandlesRuleType(t string) bool {
	return t == "eql"
}

func (e EqlRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return toEqlRuleCreateProps(ctx, client, d)
}

func (e EqlRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return toEqlRuleUpdateProps(ctx, client, d)
}

func (e EqlRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPIEqlRule)
	return ok
}

func (e EqlRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *SecurityDetectionRuleData) diag.Diagnostics {
	var diags diag.Diagnostics
	value, ok := rule.(kbapi.SecurityDetectionsAPIEqlRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return diags
	}

	return updateFromEqlRule(ctx, &value, d)
}

func (e EqlRuleProcessor) ExtractId(response any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, ok := response.(kbapi.SecurityDetectionsAPIEqlRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return "", diags
	}
	return value.Id.String(), diags
}

func toEqlRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	eqlRule := kbapi.SecurityDetectionsAPIEqlRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIEqlRuleCreatePropsType("eql"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguage("eql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &eqlRule.Actions,
		ResponseActions:                   &eqlRule.ResponseActions,
		RuleId:                            &eqlRule.RuleId,
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
		DataViewId:                        &eqlRule.DataViewId,
		Namespace:                         &eqlRule.Namespace,
		RuleNameOverride:                  &eqlRule.RuleNameOverride,
		TimestampOverride:                 &eqlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &eqlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &eqlRule.InvestigationFields,
		Filters:                           &eqlRule.Filters,
		Threat:                            &eqlRule.Threat,
		TimelineId:                        &eqlRule.TimelineId,
		TimelineTitle:                     &eqlRule.TimelineTitle,
	}, &diags, client)

	// Set EQL-specific fields
	if utils.IsKnown(d.TiebreakerField) {
		tiebreakerField := kbapi.SecurityDetectionsAPITiebreakerField(d.TiebreakerField.ValueString())
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
func toEqlRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	// Parse ID to get space_id and rule_id
	compId, resourceIdDiags := clients.CompositeIdFromStrFw(d.Id.ValueString())
	diags.Append(resourceIdDiags...)

	uid, err := uuid.Parse(compId.ResourceId)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return updateProps, diags
	}

	eqlRule := kbapi.SecurityDetectionsAPIEqlRuleUpdateProps{
		Id:          &uid,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIEqlRuleUpdatePropsType("eql"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguage("eql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		eqlRule.RuleId = &ruleId
		eqlRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &eqlRule.Actions,
		ResponseActions:                   &eqlRule.ResponseActions,
		RuleId:                            &eqlRule.RuleId,
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
		DataViewId:                        &eqlRule.DataViewId,
		Namespace:                         &eqlRule.Namespace,
		RuleNameOverride:                  &eqlRule.RuleNameOverride,
		TimestampOverride:                 &eqlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &eqlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &eqlRule.InvestigationFields,
		Filters:                           &eqlRule.Filters,
		Threat:                            &eqlRule.Threat,
		TimelineId:                        &eqlRule.TimelineId,
		TimelineTitle:                     &eqlRule.TimelineTitle,
	}, &diags, client)

	// Set EQL-specific fields
	if utils.IsKnown(d.TiebreakerField) {
		tiebreakerField := kbapi.SecurityDetectionsAPITiebreakerField(d.TiebreakerField.ValueString())
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
func updateFromEqlRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIEqlRule, d *SecurityDetectionRuleData) diag.Diagnostics {
	var diags diag.Diagnostics

	compId := clients.CompositeId{
		ClusterId:  d.SpaceId.ValueString(),
		ResourceId: rule.Id.String(),
	}
	d.Id = types.StringValue(compId.String())

	d.RuleId = types.StringValue(string(rule.RuleId))
	d.Name = types.StringValue(string(rule.Name))
	d.Type = types.StringValue(string(rule.Type))

	// Update common fields
	diags.Append(d.updateTimelineIdFromApi(ctx, rule.TimelineId)...)
	diags.Append(d.updateTimelineTitleFromApi(ctx, rule.TimelineTitle)...)
	diags.Append(d.updateDataViewIdFromApi(ctx, rule.DataViewId)...)
	diags.Append(d.updateNamespaceFromApi(ctx, rule.Namespace)...)
	diags.Append(d.updateRuleNameOverrideFromApi(ctx, rule.RuleNameOverride)...)
	diags.Append(d.updateTimestampOverrideFromApi(ctx, rule.TimestampOverride)...)
	diags.Append(d.updateTimestampOverrideFallbackDisabledFromApi(ctx, rule.TimestampOverrideFallbackDisabled)...)

	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))
	d.Enabled = types.BoolValue(bool(rule.Enabled))
	d.From = types.StringValue(string(rule.From))
	d.To = types.StringValue(string(rule.To))
	d.Interval = types.StringValue(string(rule.Interval))
	d.Description = types.StringValue(string(rule.Description))
	d.RiskScore = types.Int64Value(int64(rule.RiskScore))
	d.Severity = types.StringValue(string(rule.Severity))
	d.MaxSignals = types.Int64Value(int64(rule.MaxSignals))
	d.Version = types.Int64Value(int64(rule.Version))

	// Update building block type
	diags.Append(d.updateBuildingBlockTypeFromApi(ctx, rule.BuildingBlockType)...)

	// Update read-only fields
	d.CreatedAt = utils.TimeToStringValue(rule.CreatedAt)
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = utils.TimeToStringValue(rule.UpdatedAt)
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// Update index patterns
	diags.Append(d.updateIndexFromApi(ctx, rule.Index)...)

	// Update author
	diags.Append(d.updateAuthorFromApi(ctx, rule.Author)...)

	// Update tags
	diags.Append(d.updateTagsFromApi(ctx, rule.Tags)...)

	// Update false positives
	diags.Append(d.updateFalsePositivesFromApi(ctx, rule.FalsePositives)...)

	// Update references
	diags.Append(d.updateReferencesFromApi(ctx, rule.References)...)

	// Update optional string fields
	diags.Append(d.updateLicenseFromApi(ctx, rule.License)...)
	diags.Append(d.updateNoteFromApi(ctx, rule.Note)...)
	diags.Append(d.updateSetupFromApi(ctx, rule.Setup)...)

	// EQL-specific fields
	if rule.TiebreakerField != nil {
		d.TiebreakerField = types.StringValue(string(*rule.TiebreakerField))
	} else {
		d.TiebreakerField = types.StringNull()
	}

	// Update actions
	actionDiags := d.updateActionsFromApi(ctx, rule.Actions)
	diags.Append(actionDiags...)

	// Update exceptions list
	exceptionsListDiags := d.updateExceptionsListFromApi(ctx, rule.ExceptionsList)
	diags.Append(exceptionsListDiags...)

	// Update risk score mapping
	riskScoreMappingDiags := d.updateRiskScoreMappingFromApi(ctx, rule.RiskScoreMapping)
	diags.Append(riskScoreMappingDiags...)

	// Update investigation fields
	investigationFieldsDiags := d.updateInvestigationFieldsFromApi(ctx, rule.InvestigationFields)
	diags.Append(investigationFieldsDiags...)

	// Update filters field
	filtersDiags := d.updateFiltersFromApi(ctx, rule.Filters)
	diags.Append(filtersDiags...)

	// Update threat
	threatDiags := d.updateThreatFromApi(ctx, &rule.Threat)
	diags.Append(threatDiags...)

	// Update severity mapping
	severityMappingDiags := d.updateSeverityMappingFromApi(ctx, &rule.SeverityMapping)
	diags.Append(severityMappingDiags...)

	// Update related integrations
	relatedIntegrationsDiags := d.updateRelatedIntegrationsFromApi(ctx, &rule.RelatedIntegrations)
	diags.Append(relatedIntegrationsDiags...)

	// Update required fields
	requiredFieldsDiags := d.updateRequiredFieldsFromApi(ctx, &rule.RequiredFields)
	diags.Append(requiredFieldsDiags...)

	// Update alert suppression
	alertSuppressionDiags := d.updateAlertSuppressionFromApi(ctx, rule.AlertSuppression)
	diags.Append(alertSuppressionDiags...)

	// Update response actions
	responseActionsDiags := d.updateResponseActionsFromApi(ctx, rule.ResponseActions)
	diags.Append(responseActionsDiags...)

	return diags
}
