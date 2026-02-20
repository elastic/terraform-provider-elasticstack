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

type NewTermsRuleProcessor struct{}

func (n NewTermsRuleProcessor) HandlesRuleType(t string) bool {
	return t == "new_terms"
}

func (n NewTermsRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toNewTermsRuleCreateProps(ctx, client)
}

func (n NewTermsRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toNewTermsRuleUpdateProps(ctx, client)
}

func (n NewTermsRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPINewTermsRule)
	return ok
}

func (n NewTermsRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *Data) diag.Diagnostics {
	var diags diag.Diagnostics
	value, ok := rule.(kbapi.SecurityDetectionsAPINewTermsRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return diags
	}

	return d.updateFromNewTermsRule(ctx, &value)
}

func (n NewTermsRuleProcessor) ExtractID(response any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, ok := response.(kbapi.SecurityDetectionsAPINewTermsRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return "", diags
	}
	return value.Id.String(), diags
}

func (d Data) toNewTermsRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	newTermsRule := kbapi.SecurityDetectionsAPINewTermsRuleCreateProps{
		Name:               d.Name.ValueString(),
		Description:        d.Description.ValueString(),
		Type:               kbapi.SecurityDetectionsAPINewTermsRuleCreatePropsType("new_terms"),
		Query:              d.Query.ValueString(),
		HistoryWindowStart: d.HistoryWindowStart.ValueString(),
		RiskScore:          kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:           kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set new terms fields
	if typeutils.IsKnown(d.NewTermsFields) {
		newTermsFields := typeutils.ListTypeAs[string](ctx, d.NewTermsFields, path.Root("new_terms_fields"), &diags)
		if !diags.HasError() {
			newTermsRule.NewTermsFields = newTermsFields
		}
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &newTermsRule.Actions,
		ResponseActions:                   &newTermsRule.ResponseActions,
		RuleID:                            &newTermsRule.RuleId,
		Enabled:                           &newTermsRule.Enabled,
		From:                              &newTermsRule.From,
		To:                                &newTermsRule.To,
		Interval:                          &newTermsRule.Interval,
		Index:                             &newTermsRule.Index,
		Author:                            &newTermsRule.Author,
		Tags:                              &newTermsRule.Tags,
		FalsePositives:                    &newTermsRule.FalsePositives,
		References:                        &newTermsRule.References,
		License:                           &newTermsRule.License,
		Note:                              &newTermsRule.Note,
		Setup:                             &newTermsRule.Setup,
		MaxSignals:                        &newTermsRule.MaxSignals,
		Version:                           &newTermsRule.Version,
		ExceptionsList:                    &newTermsRule.ExceptionsList,
		AlertSuppression:                  &newTermsRule.AlertSuppression,
		RiskScoreMapping:                  &newTermsRule.RiskScoreMapping,
		SeverityMapping:                   &newTermsRule.SeverityMapping,
		RelatedIntegrations:               &newTermsRule.RelatedIntegrations,
		RequiredFields:                    &newTermsRule.RequiredFields,
		BuildingBlockType:                 &newTermsRule.BuildingBlockType,
		DataViewID:                        &newTermsRule.DataViewId,
		Namespace:                         &newTermsRule.Namespace,
		RuleNameOverride:                  &newTermsRule.RuleNameOverride,
		TimestampOverride:                 &newTermsRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &newTermsRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &newTermsRule.InvestigationFields,
		Filters:                           &newTermsRule.Filters,
		Threat:                            &newTermsRule.Threat,
		TimelineID:                        &newTermsRule.TimelineId,
		TimelineTitle:                     &newTermsRule.TimelineTitle,
	}, &diags, client)

	// Set query language
	newTermsRule.Language = d.getKQLQueryLanguage()

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPINewTermsRuleCreateProps(newTermsRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert new terms rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}
func (d Data) toNewTermsRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	newTermsRule := kbapi.SecurityDetectionsAPINewTermsRuleUpdateProps{
		Id:                 &uid,
		Name:               d.Name.ValueString(),
		Description:        d.Description.ValueString(),
		Type:               kbapi.SecurityDetectionsAPINewTermsRuleUpdatePropsType("new_terms"),
		Query:              d.Query.ValueString(),
		HistoryWindowStart: d.HistoryWindowStart.ValueString(),
		RiskScore:          kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:           kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		newTermsRule.RuleId = &ruleID
		newTermsRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set new terms fields
	if typeutils.IsKnown(d.NewTermsFields) {
		newTermsFields := typeutils.ListTypeAs[string](ctx, d.NewTermsFields, path.Root("new_terms_fields"), &diags)
		if !diags.HasError() {
			newTermsRule.NewTermsFields = newTermsFields
		}
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &newTermsRule.Actions,
		ResponseActions:                   &newTermsRule.ResponseActions,
		RuleID:                            &newTermsRule.RuleId,
		Enabled:                           &newTermsRule.Enabled,
		From:                              &newTermsRule.From,
		To:                                &newTermsRule.To,
		Interval:                          &newTermsRule.Interval,
		Index:                             &newTermsRule.Index,
		Author:                            &newTermsRule.Author,
		Tags:                              &newTermsRule.Tags,
		FalsePositives:                    &newTermsRule.FalsePositives,
		References:                        &newTermsRule.References,
		License:                           &newTermsRule.License,
		Note:                              &newTermsRule.Note,
		InvestigationFields:               &newTermsRule.InvestigationFields,
		Setup:                             &newTermsRule.Setup,
		MaxSignals:                        &newTermsRule.MaxSignals,
		Version:                           &newTermsRule.Version,
		ExceptionsList:                    &newTermsRule.ExceptionsList,
		AlertSuppression:                  &newTermsRule.AlertSuppression,
		RiskScoreMapping:                  &newTermsRule.RiskScoreMapping,
		SeverityMapping:                   &newTermsRule.SeverityMapping,
		RelatedIntegrations:               &newTermsRule.RelatedIntegrations,
		RequiredFields:                    &newTermsRule.RequiredFields,
		BuildingBlockType:                 &newTermsRule.BuildingBlockType,
		DataViewID:                        &newTermsRule.DataViewId,
		Namespace:                         &newTermsRule.Namespace,
		RuleNameOverride:                  &newTermsRule.RuleNameOverride,
		TimestampOverride:                 &newTermsRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &newTermsRule.TimestampOverrideFallbackDisabled,
		Filters:                           &newTermsRule.Filters,
		Threat:                            &newTermsRule.Threat,
		TimelineID:                        &newTermsRule.TimelineId,
		TimelineTitle:                     &newTermsRule.TimelineTitle,
	}, &diags, client)

	// Set query language
	newTermsRule.Language = d.getKQLQueryLanguage()

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPINewTermsRuleUpdateProps(newTermsRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert new terms rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}
func (d *Data) updateFromNewTermsRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPINewTermsRule) diag.Diagnostics {
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

	// Update building block type
	diags.Append(d.updateBuildingBlockTypeFromAPI(ctx, rule.BuildingBlockType)...)
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

	// Update index patterns
	diags.Append(d.updateIndexFromAPI(ctx, rule.Index)...)

	// New Terms-specific fields
	d.HistoryWindowStart = types.StringValue(rule.HistoryWindowStart)
	if len(rule.NewTermsFields) > 0 {
		d.NewTermsFields = typeutils.ListValueFrom(ctx, rule.NewTermsFields, types.StringType, path.Root("new_terms_fields"), &diags)
	} else {
		d.NewTermsFields = types.ListValueMust(types.StringType, []attr.Value{})
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
