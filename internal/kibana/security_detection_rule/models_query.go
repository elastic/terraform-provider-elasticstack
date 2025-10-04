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

type QueryRuleProcessor struct{}

func (q QueryRuleProcessor) HandlesRuleType(t string) bool {
	return t == "query"
}

func (q QueryRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return toQueryRuleCreateProps(ctx, client, d)
}

func (q QueryRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return toQueryRuleUpdateProps(ctx, client, d)
}

func (q QueryRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPIQueryRule)
	return ok
}

func (q QueryRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *SecurityDetectionRuleData) diag.Diagnostics {
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

func (q QueryRuleProcessor) ExtractId(response any) (string, diag.Diagnostics) {
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

func toQueryRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	queryRuleQuery := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())
	queryRule := kbapi.SecurityDetectionsAPIQueryRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleCreatePropsType("query"),
		Query:       &queryRuleQuery,
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &queryRule.Actions,
		ResponseActions:                   &queryRule.ResponseActions,
		RuleId:                            &queryRule.RuleId,
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
		DataViewId:                        &queryRule.DataViewId,
		Namespace:                         &queryRule.Namespace,
		RuleNameOverride:                  &queryRule.RuleNameOverride,
		TimestampOverride:                 &queryRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &queryRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &queryRule.InvestigationFields,
		Filters:                           &queryRule.Filters,
	}, &diags, client)

	// Set query-specific fields
	queryRule.Language = d.getKQLQueryLanguage()

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		queryRule.SavedId = &savedId
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

func toQueryRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	queryRuleQuery := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())

	// Parse ID to get space_id and rule_id
	compId, resourceIdDiags := clients.CompositeIdFromStrFw(d.Id.ValueString())
	diags.Append(resourceIdDiags...)

	uid, err := uuid.Parse(compId.ResourceId)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return updateProps, diags
	}
	var id = kbapi.SecurityDetectionsAPIRuleObjectId(uid)

	queryRule := kbapi.SecurityDetectionsAPIQueryRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleUpdatePropsType("query"),
		Query:       &queryRuleQuery,
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		queryRule.RuleId = &ruleId
		queryRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &queryRule.Actions,
		ResponseActions:                   &queryRule.ResponseActions,
		RuleId:                            &queryRule.RuleId,
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
		DataViewId:                        &queryRule.DataViewId,
		Namespace:                         &queryRule.Namespace,
		RuleNameOverride:                  &queryRule.RuleNameOverride,
		TimestampOverride:                 &queryRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &queryRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &queryRule.InvestigationFields,
		Filters:                           &queryRule.Filters,
	}, &diags, client)

	// Set query-specific fields
	queryRule.Language = d.getKQLQueryLanguage()

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		queryRule.SavedId = &savedId
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
func updateFromQueryRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIQueryRule, d *SecurityDetectionRuleData) diag.Diagnostics {
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
	dataViewIdDiags := d.updateDataViewIdFromApi(ctx, rule.DataViewId)
	diags.Append(dataViewIdDiags...)

	namespaceDiags := d.updateNamespaceFromApi(ctx, rule.Namespace)
	diags.Append(namespaceDiags...)

	ruleNameOverrideDiags := d.updateRuleNameOverrideFromApi(ctx, rule.RuleNameOverride)
	diags.Append(ruleNameOverrideDiags...)

	timestampOverrideDiags := d.updateTimestampOverrideFromApi(ctx, rule.TimestampOverride)
	diags.Append(timestampOverrideDiags...)

	timestampOverrideFallbackDisabledDiags := d.updateTimestampOverrideFallbackDisabledFromApi(ctx, rule.TimestampOverrideFallbackDisabled)
	diags.Append(timestampOverrideFallbackDisabledDiags...)

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
	buildingBlockTypeDiags := d.updateBuildingBlockTypeFromApi(ctx, rule.BuildingBlockType)
	diags.Append(buildingBlockTypeDiags...)

	// Update read-only fields
	d.CreatedAt = utils.TimeToStringValue(rule.CreatedAt)
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = utils.TimeToStringValue(rule.UpdatedAt)
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// Update index patterns
	indexDiags := d.updateIndexFromApi(ctx, rule.Index)
	diags.Append(indexDiags...)

	// Update author
	authorDiags := d.updateAuthorFromApi(ctx, rule.Author)
	diags.Append(authorDiags...)

	// Update tags
	tagsDiags := d.updateTagsFromApi(ctx, rule.Tags)
	diags.Append(tagsDiags...)

	// Update false positives
	falsePositivesDiags := d.updateFalsePositivesFromApi(ctx, rule.FalsePositives)
	diags.Append(falsePositivesDiags...)

	// Update references
	referencesDiags := d.updateReferencesFromApi(ctx, rule.References)
	diags.Append(referencesDiags...)

	// Update optional string fields
	licenseDiags := d.updateLicenseFromApi(ctx, rule.License)
	diags.Append(licenseDiags...)

	noteDiags := d.updateNoteFromApi(ctx, rule.Note)
	diags.Append(noteDiags...)

	setupDiags := d.updateSetupFromApi(ctx, rule.Setup)
	diags.Append(setupDiags...)

	// Update actions
	actionDiags := d.updateActionsFromApi(ctx, rule.Actions)
	diags.Append(actionDiags...)

	// Update exceptions list
	exceptionsListDiags := d.updateExceptionsListFromApi(ctx, rule.ExceptionsList)
	diags.Append(exceptionsListDiags...)

	// Update risk score mapping
	riskScoreMappingDiags := d.updateRiskScoreMappingFromApi(ctx, rule.RiskScoreMapping)
	diags.Append(riskScoreMappingDiags...)

	// Update severity mapping
	severityMappingDiags := d.updateSeverityMappingFromApi(ctx, &rule.SeverityMapping)
	diags.Append(severityMappingDiags...)

	// Update related integrations
	relatedIntegrationsDiags := d.updateRelatedIntegrationsFromApi(ctx, &rule.RelatedIntegrations)
	diags.Append(relatedIntegrationsDiags...)

	// Update required fields
	requiredFieldsDiags := d.updateRequiredFieldsFromApi(ctx, &rule.RequiredFields)
	diags.Append(requiredFieldsDiags...)

	// Update investigation fields
	investigationFieldsDiags := d.updateInvestigationFieldsFromApi(ctx, rule.InvestigationFields)
	diags.Append(investigationFieldsDiags...)

	// Update filters field
	filtersDiags := d.updateFiltersFromApi(ctx, rule.Filters)
	diags.Append(filtersDiags...)

	// Update alert suppression
	alertSuppressionDiags := d.updateAlertSuppressionFromApi(ctx, rule.AlertSuppression)
	diags.Append(alertSuppressionDiags...)

	// Update response actions
	responseActionsDiags := d.updateResponseActionsFromApi(ctx, rule.ResponseActions)
	diags.Append(responseActionsDiags...)

	return diags
}
