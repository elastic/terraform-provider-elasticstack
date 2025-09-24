package security_detection_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d SecurityDetectionRuleData) toSavedQueryRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
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
		Meta:                              &savedQueryRule.Meta,
		Filters:                           &savedQueryRule.Filters,
	}, &diags)

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
func (d SecurityDetectionRuleData) toSavedQueryRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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
		Meta:                              &savedQueryRule.Meta,
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
	}, &diags)

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
	if rule.DataViewId != nil {
		d.DataViewId = types.StringValue(string(*rule.DataViewId))
	} else {
		d.DataViewId = types.StringNull()
	}

	if rule.Namespace != nil {
		d.Namespace = types.StringValue(string(*rule.Namespace))
	} else {
		d.Namespace = types.StringNull()
	}

	if rule.RuleNameOverride != nil {
		d.RuleNameOverride = types.StringValue(string(*rule.RuleNameOverride))
	} else {
		d.RuleNameOverride = types.StringNull()
	}

	if rule.TimestampOverride != nil {
		d.TimestampOverride = types.StringValue(string(*rule.TimestampOverride))
	} else {
		d.TimestampOverride = types.StringNull()
	}

	if rule.TimestampOverrideFallbackDisabled != nil {
		d.TimestampOverrideFallbackDisabled = types.BoolValue(bool(*rule.TimestampOverrideFallbackDisabled))
	} else {
		d.TimestampOverrideFallbackDisabled = types.BoolNull()
	}

	d.SavedId = types.StringValue(string(rule.SavedId))
	d.Enabled = types.BoolValue(bool(rule.Enabled))
	d.From = types.StringValue(string(rule.From))

	// Update building block type
	if rule.BuildingBlockType != nil {
		d.BuildingBlockType = types.StringValue(string(*rule.BuildingBlockType))
	} else {
		d.BuildingBlockType = types.StringNull()
	}
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
	if rule.Index != nil && len(*rule.Index) > 0 {
		d.Index = utils.ListValueFrom(ctx, *rule.Index, types.StringType, path.Root("index"), &diags)
	} else {
		d.Index = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Optional query for saved query rules
	if rule.Query != nil {
		d.Query = types.StringValue(*rule.Query)
	} else {
		d.Query = types.StringNull()
	}

	// Language for saved query rules (not a pointer)
	d.Language = types.StringValue(string(rule.Language))

	// Update author
	if len(rule.Author) > 0 {
		d.Author = utils.ListValueFrom(ctx, rule.Author, types.StringType, path.Root("author"), &diags)
	} else {
		d.Author = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update tags
	if len(rule.Tags) > 0 {
		d.Tags = utils.ListValueFrom(ctx, rule.Tags, types.StringType, path.Root("tags"), &diags)
	} else {
		d.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update false positives
	if len(rule.FalsePositives) > 0 {
		d.FalsePositives = utils.ListValueFrom(ctx, rule.FalsePositives, types.StringType, path.Root("false_positives"), &diags)
	} else {
		d.FalsePositives = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update references
	if len(rule.References) > 0 {
		d.References = utils.ListValueFrom(ctx, rule.References, types.StringType, path.Root("references"), &diags)
	} else {
		d.References = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update optional string fields
	if rule.License != nil {
		d.License = types.StringValue(string(*rule.License))
	} else {
		d.License = types.StringNull()
	}

	if rule.Note != nil {
		d.Note = types.StringValue(string(*rule.Note))
	} else {
		d.Note = types.StringNull()
	}

	// Handle setup field - if empty, set to null to maintain consistency with optional schema
	if string(rule.Setup) != "" {
		d.Setup = types.StringValue(string(rule.Setup))
	} else {
		d.Setup = types.StringNull()
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

	// Update meta field
	metaDiags := d.updateMetaFromApi(ctx, rule.Meta)
	diags.Append(metaDiags...)

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
