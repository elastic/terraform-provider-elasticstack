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

func (d SecurityDetectionRuleData) toThresholdRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	thresholdRule := kbapi.SecurityDetectionsAPIThresholdRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIThresholdRuleCreatePropsType("threshold"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set threshold - this is required for threshold rules
	threshold := d.thresholdToApi(ctx, &diags)
	if threshold != nil {
		thresholdRule.Threshold = *threshold
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &thresholdRule.Actions,
		ResponseActions:                   &thresholdRule.ResponseActions,
		RuleId:                            &thresholdRule.RuleId,
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
		DataViewId:                        &thresholdRule.DataViewId,
		Namespace:                         &thresholdRule.Namespace,
		RuleNameOverride:                  &thresholdRule.RuleNameOverride,
		TimestampOverride:                 &thresholdRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &thresholdRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &thresholdRule.InvestigationFields,
		Meta:                              &thresholdRule.Meta,
		Filters:                           &thresholdRule.Filters,
		AlertSuppression:                  nil, // Handle specially for threshold rule
	}, &diags)

	// Handle threshold-specific alert suppression
	if utils.IsKnown(d.AlertSuppression) {
		alertSuppression := d.alertSuppressionToThresholdApi(ctx, &diags)
		if alertSuppression != nil {
			thresholdRule.AlertSuppression = alertSuppression
		}
	}

	// Set query language
	thresholdRule.Language = d.getKQLQueryLanguage()

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		thresholdRule.SavedId = &savedId
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
func (d SecurityDetectionRuleData) toThresholdRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	thresholdRule := kbapi.SecurityDetectionsAPIThresholdRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIThresholdRuleUpdatePropsType("threshold"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		thresholdRule.RuleId = &ruleId
		thresholdRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set threshold - this is required for threshold rules
	threshold := d.thresholdToApi(ctx, &diags)
	if threshold != nil {
		thresholdRule.Threshold = *threshold
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &thresholdRule.Actions,
		ResponseActions:                   &thresholdRule.ResponseActions,
		RuleId:                            &thresholdRule.RuleId,
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
		Meta:                              &thresholdRule.Meta,
		Setup:                             &thresholdRule.Setup,
		MaxSignals:                        &thresholdRule.MaxSignals,
		Version:                           &thresholdRule.Version,
		ExceptionsList:                    &thresholdRule.ExceptionsList,
		RiskScoreMapping:                  &thresholdRule.RiskScoreMapping,
		SeverityMapping:                   &thresholdRule.SeverityMapping,
		RelatedIntegrations:               &thresholdRule.RelatedIntegrations,
		RequiredFields:                    &thresholdRule.RequiredFields,
		BuildingBlockType:                 &thresholdRule.BuildingBlockType,
		DataViewId:                        &thresholdRule.DataViewId,
		Namespace:                         &thresholdRule.Namespace,
		RuleNameOverride:                  &thresholdRule.RuleNameOverride,
		TimestampOverride:                 &thresholdRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &thresholdRule.TimestampOverrideFallbackDisabled,
		Filters:                           &thresholdRule.Filters,
		AlertSuppression:                  nil, // Handle specially for threshold rule
	}, &diags)

	// Handle threshold-specific alert suppression
	if utils.IsKnown(d.AlertSuppression) {
		alertSuppression := d.alertSuppressionToThresholdApi(ctx, &diags)
		if alertSuppression != nil {
			thresholdRule.AlertSuppression = alertSuppression
		}
	}

	// Set query language
	thresholdRule.Language = d.getKQLQueryLanguage()

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		thresholdRule.SavedId = &savedId
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

func (d *SecurityDetectionRuleData) updateFromThresholdRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIThresholdRule) diag.Diagnostics {
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

	// Update building block type
	if rule.BuildingBlockType != nil {
		d.BuildingBlockType = types.StringValue(string(*rule.BuildingBlockType))
	} else {
		d.BuildingBlockType = types.StringNull()
	}
	d.From = types.StringValue(string(rule.From))
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

	// Threshold-specific fields
	thresholdObj, thresholdDiags := convertThresholdToModel(ctx, rule.Threshold)
	diags.Append(thresholdDiags...)
	if !thresholdDiags.HasError() {
		d.Threshold = thresholdObj
	}

	// Optional saved query ID
	if rule.SavedId != nil {
		d.SavedId = types.StringValue(string(*rule.SavedId))
	} else {
		d.SavedId = types.StringNull()
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
	thresholdAlertSuppressionDiags := d.updateThresholdAlertSuppressionFromApi(ctx, rule.AlertSuppression)
	diags.Append(thresholdAlertSuppressionDiags...)

	// Update response actions
	responseActionsDiags := d.updateResponseActionsFromApi(ctx, rule.ResponseActions)
	diags.Append(responseActionsDiags...)

	return diags
}
