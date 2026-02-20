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

type EqlRuleProcessor struct{}

func (e EqlRuleProcessor) HandlesRuleType(t string) bool {
	return t == "eql"
}

func (e EqlRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return toEqlRuleCreateProps(ctx, client, d)
}

func (e EqlRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return toEqlRuleUpdateProps(ctx, client, d)
}

func (e EqlRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPIEqlRule)
	return ok
}

func (e EqlRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *Data) diag.Diagnostics {
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

func (e EqlRuleProcessor) ExtractID(response any) (string, diag.Diagnostics) {
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

func toEqlRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	eqlRule := kbapi.SecurityDetectionsAPIEqlRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIEqlRuleCreatePropsType("eql"),
		Query:       d.Query.ValueString(),
		Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguage("eql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &eqlRule.Actions,
		ResponseActions:                   &eqlRule.ResponseActions,
		RuleID:                            &eqlRule.RuleId,
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
		DataViewID:                        &eqlRule.DataViewId,
		Namespace:                         &eqlRule.Namespace,
		RuleNameOverride:                  &eqlRule.RuleNameOverride,
		TimestampOverride:                 &eqlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &eqlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &eqlRule.InvestigationFields,
		Filters:                           &eqlRule.Filters,
		Threat:                            &eqlRule.Threat,
		TimelineID:                        &eqlRule.TimelineId,
		TimelineTitle:                     &eqlRule.TimelineTitle,
	}, &diags, client)

	// Set EQL-specific fields
	if typeutils.IsKnown(d.TiebreakerField) {
		tiebreakerField := d.TiebreakerField.ValueString()
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
func toEqlRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	eqlRule := kbapi.SecurityDetectionsAPIEqlRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIEqlRuleUpdatePropsType("eql"),
		Query:       d.Query.ValueString(),
		Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguage("eql"),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		eqlRule.RuleId = &ruleID
		eqlRule.Id = nil // if rule_id is set, we cant send id
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &eqlRule.Actions,
		ResponseActions:                   &eqlRule.ResponseActions,
		RuleID:                            &eqlRule.RuleId,
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
		DataViewID:                        &eqlRule.DataViewId,
		Namespace:                         &eqlRule.Namespace,
		RuleNameOverride:                  &eqlRule.RuleNameOverride,
		TimestampOverride:                 &eqlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &eqlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &eqlRule.InvestigationFields,
		Filters:                           &eqlRule.Filters,
		Threat:                            &eqlRule.Threat,
		TimelineID:                        &eqlRule.TimelineId,
		TimelineTitle:                     &eqlRule.TimelineTitle,
	}, &diags, client)

	// Set EQL-specific fields
	if typeutils.IsKnown(d.TiebreakerField) {
		tiebreakerField := d.TiebreakerField.ValueString()
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
func updateFromEqlRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIEqlRule, d *Data) diag.Diagnostics {
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

	// Update building block type
	diags.Append(d.updateBuildingBlockTypeFromAPI(ctx, rule.BuildingBlockType)...)

	// Update read-only fields
	d.CreatedAt = schemautil.TimeToStringValue(rule.CreatedAt)
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = schemautil.TimeToStringValue(rule.UpdatedAt)
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// Update index patterns
	diags.Append(d.updateIndexFromAPI(ctx, rule.Index)...)

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

	// EQL-specific fields
	if rule.TiebreakerField != nil {
		d.TiebreakerField = types.StringValue(*rule.TiebreakerField)
	} else {
		d.TiebreakerField = types.StringNull()
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

	// Update threat
	threatDiags := d.updateThreatFromAPI(ctx, &rule.Threat)
	diags.Append(threatDiags...)

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
