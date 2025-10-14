package security_detection_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EsqlRuleProcessor struct{}

func (e EsqlRuleProcessor) HandlesRuleType(t string) bool {
	return t == "esql"
}

func (e EsqlRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toEsqlRuleCreateProps(ctx, client)
}

func (e EsqlRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toEsqlRuleUpdateProps(ctx, client)
}

func (e EsqlRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPIEsqlRule)
	return ok
}

func (e EsqlRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *SecurityDetectionRuleData) diag.Diagnostics {
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

func (e EsqlRuleProcessor) ExtractId(response any) (string, diag.Diagnostics) {
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

// applyEsqlValidations validates that ESQL-specific constraints are met
func (d SecurityDetectionRuleData) applyEsqlValidations(diags *diag.Diagnostics) {
	if utils.IsKnown(d.Index) {
		diags.AddError(
			"Invalid attribute 'index'",
			"ESQL rules do not use index patterns. Please remove the 'index' attribute.",
		)
	}

	if utils.IsKnown(d.Filters) {
		diags.AddError(
			"Invalid attribute 'filters'",
			"ESQL rules do not support filters. Please remove the 'filters' attribute.",
		)
	}
}

func (d SecurityDetectionRuleData) toEsqlRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	// Apply ESQL-specific validations
	d.applyEsqlValidations(&diags)
	if diags.HasError() {
		return createProps, diags
	}

	esqlRule := kbapi.SecurityDetectionsAPIEsqlRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIEsqlRuleCreatePropsType("esql"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		Language:    kbapi.SecurityDetectionsAPIEsqlQueryLanguage("esql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &esqlRule.Actions,
		ResponseActions:                   &esqlRule.ResponseActions,
		RuleId:                            &esqlRule.RuleId,
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
		DataViewId:                        nil, // ESQL rules don't have DataViewId
		Namespace:                         &esqlRule.Namespace,
		RuleNameOverride:                  &esqlRule.RuleNameOverride,
		TimestampOverride:                 &esqlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &esqlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &esqlRule.InvestigationFields,
		Filters:                           nil, // ESQL rules don't support this field
		Threat:                            &esqlRule.Threat,
		TimelineId:                        &esqlRule.TimelineId,
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

func (d SecurityDetectionRuleData) toEsqlRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	// Apply ESQL-specific validations
	d.applyEsqlValidations(&diags)
	if diags.HasError() {
		return updateProps, diags
	}

	// Parse ID to get space_id and rule_id
	compId, resourceIdDiags := clients.CompositeIdFromStrFw(d.Id.ValueString())
	diags.Append(resourceIdDiags...)

	uid, err := uuid.Parse(compId.ResourceId)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return updateProps, diags
	}
	var id = kbapi.SecurityDetectionsAPIRuleObjectId(uid)

	esqlRule := kbapi.SecurityDetectionsAPIEsqlRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIEsqlRuleUpdatePropsType("esql"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		Language:    kbapi.SecurityDetectionsAPIEsqlQueryLanguage("esql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		esqlRule.RuleId = &ruleId
		esqlRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &esqlRule.Actions,
		ResponseActions:                   &esqlRule.ResponseActions,
		RuleId:                            &esqlRule.RuleId,
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
		DataViewId:                        nil, // ESQL rules don't have DataViewId
		Namespace:                         &esqlRule.Namespace,
		RuleNameOverride:                  &esqlRule.RuleNameOverride,
		TimestampOverride:                 &esqlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &esqlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &esqlRule.InvestigationFields,
		Filters:                           nil, // ESQL rules don't have Filters
		Threat:                            &esqlRule.Threat,
		TimelineId:                        &esqlRule.TimelineId,
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
func (d *SecurityDetectionRuleData) updateFromEsqlRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIEsqlRule) diag.Diagnostics {
	var diags diag.Diagnostics

	compId := clients.CompositeId{
		ClusterId:  d.SpaceId.ValueString(),
		ResourceId: rule.Id.String(),
	}
	d.Id = types.StringValue(compId.String())

	d.RuleId = types.StringValue(string(rule.RuleId))
	d.Name = types.StringValue(string(rule.Name))
	d.Type = types.StringValue(string(rule.Type))

	// Update common fields (ESQL doesn't support DataViewId)
	d.DataViewId = types.StringNull()
	diags.Append(d.updateTimelineIdFromApi(ctx, rule.TimelineId)...)
	diags.Append(d.updateTimelineTitleFromApi(ctx, rule.TimelineTitle)...)
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
	d.CreatedAt = types.StringValue(rule.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = types.StringValue(rule.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// ESQL rules don't use index patterns
	d.Index = types.ListValueMust(types.StringType, []attr.Value{})

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
