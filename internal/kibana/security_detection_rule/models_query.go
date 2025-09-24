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

func (d SecurityDetectionRuleData) toQueryRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
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
		Meta:                              &queryRule.Meta,
		Filters:                           &queryRule.Filters,
	}, &diags)

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

func (d SecurityDetectionRuleData) toQueryRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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
		Meta:                              &queryRule.Meta,
		Filters:                           &queryRule.Filters,
	}, &diags)

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
func (d *SecurityDetectionRuleData) updateFromQueryRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIQueryRule) diag.Diagnostics {
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
	if rule.BuildingBlockType != nil {
		d.BuildingBlockType = types.StringValue(string(*rule.BuildingBlockType))
	} else {
		d.BuildingBlockType = types.StringNull()
	}

	// Update read-only fields
	d.CreatedAt = utils.TimeToStringValue(rule.CreatedAt)
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = utils.TimeToStringValue(rule.UpdatedAt)
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// Update index patterns
	if rule.Index != nil && len(*rule.Index) > 0 {
		d.Index = utils.ListValueFrom(ctx, *rule.Index, types.StringType, path.Root("index"), &diags)
	} else {
		d.Index = types.ListValueMust(types.StringType, []attr.Value{})
	}

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

	// Update meta field
	metaDiags := d.updateMetaFromApi(ctx, rule.Meta)
	diags.Append(metaDiags...)

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
