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

type SavedQueryRuleProcessor struct{}

func (s SavedQueryRuleProcessor) HandlesRuleType(t string) bool {
	return t == "saved_query"
}

func (s SavedQueryRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toSavedQueryRuleCreateProps(ctx, client)
}

func (s SavedQueryRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toSavedQueryRuleUpdateProps(ctx, client)
}

func (s SavedQueryRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPISavedQueryRule)
	return ok
}

func (s SavedQueryRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *SecurityDetectionRuleData) diag.Diagnostics {
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

func (s SavedQueryRuleProcessor) ExtractId(response any) (string, diag.Diagnostics) {
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

func (d SecurityDetectionRuleData) toSavedQueryRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	savedQueryRule := kbapi.SecurityDetectionsAPISavedQueryRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPISavedQueryRuleCreatePropsType("saved_query"),
		SavedId:     kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &savedQueryRule.Actions,
		ResponseActions:                   &savedQueryRule.ResponseActions,
		RuleId:                            &savedQueryRule.RuleId,
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
		DataViewId:                        &savedQueryRule.DataViewId,
		Namespace:                         &savedQueryRule.Namespace,
		RuleNameOverride:                  &savedQueryRule.RuleNameOverride,
		TimestampOverride:                 &savedQueryRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &savedQueryRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &savedQueryRule.InvestigationFields,
		Filters:                           &savedQueryRule.Filters,
	}, &diags, client)

	// Set optional query for saved query rules
	if utils.IsKnown(d.Query) {
		query := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())
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
func (d SecurityDetectionRuleData) toSavedQueryRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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
	var id = kbapi.SecurityDetectionsAPIRuleObjectId(uid)

	savedQueryRule := kbapi.SecurityDetectionsAPISavedQueryRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPISavedQueryRuleUpdatePropsType("saved_query"),
		SavedId:     kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		savedQueryRule.RuleId = &ruleId
		savedQueryRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &savedQueryRule.Actions,
		ResponseActions:                   &savedQueryRule.ResponseActions,
		RuleId:                            &savedQueryRule.RuleId,
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
		DataViewId:                        &savedQueryRule.DataViewId,
		Namespace:                         &savedQueryRule.Namespace,
		RuleNameOverride:                  &savedQueryRule.RuleNameOverride,
		TimestampOverride:                 &savedQueryRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &savedQueryRule.TimestampOverrideFallbackDisabled,
		Filters:                           &savedQueryRule.Filters,
	}, &diags, client)

	// Set optional query for saved query rules
	if utils.IsKnown(d.Query) {
		query := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())
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

func (d *SecurityDetectionRuleData) updateFromSavedQueryRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPISavedQueryRule) diag.Diagnostics {
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
	diags.Append(d.updateDataViewIdFromApi(ctx, rule.DataViewId)...)
	diags.Append(d.updateNamespaceFromApi(ctx, rule.Namespace)...)
	diags.Append(d.updateRuleNameOverrideFromApi(ctx, rule.RuleNameOverride)...)
	diags.Append(d.updateTimestampOverrideFromApi(ctx, rule.TimestampOverride)...)
	diags.Append(d.updateTimestampOverrideFallbackDisabledFromApi(ctx, rule.TimestampOverrideFallbackDisabled)...)

	d.SavedId = types.StringValue(string(rule.SavedId))
	d.Enabled = types.BoolValue(bool(rule.Enabled))
	d.From = types.StringValue(string(rule.From))

	// Update building block type
	diags.Append(d.updateBuildingBlockTypeFromApi(ctx, rule.BuildingBlockType)...)
	d.To = types.StringValue(string(rule.To))
	d.Interval = types.StringValue(string(rule.Interval))
	d.Description = types.StringValue(string(rule.Description))
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

	// Update index patterns
	diags.Append(d.updateIndexFromApi(ctx, rule.Index)...)

	// Optional query for saved query rules
	d.Query = types.StringPointerValue(rule.Query)

	// Language for saved query rules (not a pointer)
	d.Language = types.StringValue(string(rule.Language))

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
