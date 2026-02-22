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

type MachineLearningRuleProcessor struct{}

func (m MachineLearningRuleProcessor) HandlesRuleType(t string) bool {
	return t == "machine_learning"
}

func (m MachineLearningRuleProcessor) ToCreateProps(
	ctx context.Context,
	client clients.MinVersionEnforceable,
	d Data,
) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toMachineLearningRuleCreateProps(ctx, client)
}

func (m MachineLearningRuleProcessor) ToUpdateProps(
	ctx context.Context,
	client clients.MinVersionEnforceable,
	d Data,
) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toMachineLearningRuleUpdateProps(ctx, client)
}

func (m MachineLearningRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPIMachineLearningRule)
	return ok
}

func (m MachineLearningRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *Data) diag.Diagnostics {
	var diags diag.Diagnostics
	value, ok := rule.(kbapi.SecurityDetectionsAPIMachineLearningRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return diags
	}

	return d.updateFromMachineLearningRule(ctx, &value)
}

func (m MachineLearningRuleProcessor) ExtractID(response any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, ok := response.(kbapi.SecurityDetectionsAPIMachineLearningRule)
	if !ok {
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return "", diags
	}
	return value.Id.String(), diags
}

func (d Data) toMachineLearningRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	mlRule := kbapi.SecurityDetectionsAPIMachineLearningRuleCreateProps{
		Name:             d.Name.ValueString(),
		Description:      d.Description.ValueString(),
		Type:             kbapi.SecurityDetectionsAPIMachineLearningRuleCreatePropsType("machine_learning"),
		AnomalyThreshold: kbapi.SecurityDetectionsAPIAnomalyThreshold(d.AnomalyThreshold.ValueInt64()),
		RiskScore:        kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:         kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set ML job ID(s) - can be single string or array
	if typeutils.IsKnown(d.MachineLearningJobID) {
		jobIDs := typeutils.ListTypeAs[string](ctx, d.MachineLearningJobID, path.Root("machine_learning_job_id"), &diags)
		if !diags.HasError() {
			var mlJobID kbapi.SecurityDetectionsAPIMachineLearningJobId
			err := mlJobID.FromSecurityDetectionsAPIMachineLearningJobId1(jobIDs)
			if err != nil {
				diags.AddError("Error setting ML job IDs", err.Error())
			} else {
				mlRule.MachineLearningJobId = mlJobID
			}
		}
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &mlRule.Actions,
		ResponseActions:                   &mlRule.ResponseActions,
		RuleID:                            &mlRule.RuleId,
		Enabled:                           &mlRule.Enabled,
		From:                              &mlRule.From,
		To:                                &mlRule.To,
		Interval:                          &mlRule.Interval,
		Index:                             nil, // ML rules don't use index patterns
		Author:                            &mlRule.Author,
		Tags:                              &mlRule.Tags,
		FalsePositives:                    &mlRule.FalsePositives,
		References:                        &mlRule.References,
		License:                           &mlRule.License,
		Note:                              &mlRule.Note,
		Setup:                             &mlRule.Setup,
		MaxSignals:                        &mlRule.MaxSignals,
		Version:                           &mlRule.Version,
		ExceptionsList:                    &mlRule.ExceptionsList,
		AlertSuppression:                  &mlRule.AlertSuppression,
		RiskScoreMapping:                  &mlRule.RiskScoreMapping,
		SeverityMapping:                   &mlRule.SeverityMapping,
		RelatedIntegrations:               &mlRule.RelatedIntegrations,
		RequiredFields:                    &mlRule.RequiredFields,
		BuildingBlockType:                 &mlRule.BuildingBlockType,
		DataViewID:                        nil, // ML rules don't have DataViewID
		Namespace:                         &mlRule.Namespace,
		RuleNameOverride:                  &mlRule.RuleNameOverride,
		TimestampOverride:                 &mlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &mlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &mlRule.InvestigationFields,
		Filters:                           nil, // ML rules don't have Filters
		Threat:                            &mlRule.Threat,
		TimelineID:                        &mlRule.TimelineId,
		TimelineTitle:                     &mlRule.TimelineTitle,
	}, &diags, client)

	// ML rules don't use index patterns or query

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIMachineLearningRuleCreateProps(mlRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert ML rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}
func (d Data) toMachineLearningRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	mlRule := kbapi.SecurityDetectionsAPIMachineLearningRuleUpdateProps{
		Id:               &uid,
		Name:             d.Name.ValueString(),
		Description:      d.Description.ValueString(),
		Type:             kbapi.SecurityDetectionsAPIMachineLearningRuleUpdatePropsType("machine_learning"),
		AnomalyThreshold: kbapi.SecurityDetectionsAPIAnomalyThreshold(d.AnomalyThreshold.ValueInt64()),
		RiskScore:        kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:         kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		mlRule.RuleId = &ruleID
		mlRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set ML job ID(s) - can be single string or array
	if typeutils.IsKnown(d.MachineLearningJobID) {
		jobIDs := typeutils.ListTypeAs[string](ctx, d.MachineLearningJobID, path.Root("machine_learning_job_id"), &diags)
		if !diags.HasError() {
			var mlJobID kbapi.SecurityDetectionsAPIMachineLearningJobId
			err := mlJobID.FromSecurityDetectionsAPIMachineLearningJobId1(jobIDs)
			if err != nil {
				diags.AddError("Error setting ML job IDs", err.Error())
			} else {
				mlRule.MachineLearningJobId = mlJobID
			}
		}
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &mlRule.Actions,
		ResponseActions:                   &mlRule.ResponseActions,
		RuleID:                            &mlRule.RuleId,
		Enabled:                           &mlRule.Enabled,
		From:                              &mlRule.From,
		To:                                &mlRule.To,
		Interval:                          &mlRule.Interval,
		Index:                             nil, // ML rules don't use index patterns
		Author:                            &mlRule.Author,
		Tags:                              &mlRule.Tags,
		FalsePositives:                    &mlRule.FalsePositives,
		References:                        &mlRule.References,
		License:                           &mlRule.License,
		Note:                              &mlRule.Note,
		Setup:                             &mlRule.Setup,
		MaxSignals:                        &mlRule.MaxSignals,
		Version:                           &mlRule.Version,
		ExceptionsList:                    &mlRule.ExceptionsList,
		AlertSuppression:                  &mlRule.AlertSuppression,
		RiskScoreMapping:                  &mlRule.RiskScoreMapping,
		SeverityMapping:                   &mlRule.SeverityMapping,
		RelatedIntegrations:               &mlRule.RelatedIntegrations,
		RequiredFields:                    &mlRule.RequiredFields,
		BuildingBlockType:                 &mlRule.BuildingBlockType,
		DataViewID:                        nil, // ML rules don't have DataViewID
		Namespace:                         &mlRule.Namespace,
		RuleNameOverride:                  &mlRule.RuleNameOverride,
		TimestampOverride:                 &mlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &mlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &mlRule.InvestigationFields,
		Filters:                           nil, // ML rules don't have Filters
		Threat:                            &mlRule.Threat,
		TimelineID:                        &mlRule.TimelineId,
		TimelineTitle:                     &mlRule.TimelineTitle,
	}, &diags, client)

	// ML rules don't use index patterns or query

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIMachineLearningRuleUpdateProps(mlRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert ML rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d *Data) updateFromMachineLearningRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIMachineLearningRule) diag.Diagnostics {
	var diags diag.Diagnostics

	compID := clients.CompositeID{
		ClusterID:  d.SpaceID.ValueString(),
		ResourceID: rule.Id.String(),
	}
	d.ID = types.StringValue(compID.String())

	d.RuleID = types.StringValue(rule.RuleId)
	d.Name = types.StringValue(rule.Name)
	d.Type = types.StringValue(string(rule.Type))

	// Update common fields (ML doesn't support DataViewID)
	d.DataViewID = types.StringNull()
	diags.Append(d.updateTimelineIDFromAPI(ctx, rule.TimelineId)...)
	diags.Append(d.updateTimelineTitleFromAPI(ctx, rule.TimelineTitle)...)
	diags.Append(d.updateNamespaceFromAPI(ctx, rule.Namespace)...)
	diags.Append(d.updateRuleNameOverrideFromAPI(ctx, rule.RuleNameOverride)...)
	diags.Append(d.updateTimestampOverrideFromAPI(ctx, rule.TimestampOverride)...)
	diags.Append(d.updateTimestampOverrideFallbackDisabledFromAPI(ctx, rule.TimestampOverrideFallbackDisabled)...)

	d.Enabled = types.BoolValue(rule.Enabled)
	d.From = types.StringValue(rule.From)
	d.To = types.StringValue(rule.To)
	d.Interval = types.StringValue(rule.Interval)
	d.Description = types.StringValue(rule.Description)
	d.RiskScore = types.Int64Value(int64(rule.RiskScore))
	d.Severity = types.StringValue(string(rule.Severity))

	// Update building block type
	diags.Append(d.updateBuildingBlockTypeFromAPI(ctx, rule.BuildingBlockType)...)
	d.MaxSignals = types.Int64Value(int64(rule.MaxSignals))
	d.Version = types.Int64Value(int64(rule.Version))

	// Update read-only fields
	d.CreatedAt = types.StringValue(rule.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = types.StringValue(rule.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// ML rules don't use index patterns or query
	d.Index = types.ListValueMust(types.StringType, []attr.Value{})
	d.Query = types.StringNull()
	d.Language = types.StringNull()

	// ML-specific fields
	d.AnomalyThreshold = types.Int64Value(int64(rule.AnomalyThreshold))

	// Handle ML job ID(s) - can be single string or array
	// Try to extract as single job ID first, then as array
	if singleJobID, err := rule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId0(); err == nil {
		// Single job ID
		d.MachineLearningJobID = typeutils.ListValueFrom(ctx, []string{singleJobID}, types.StringType, path.Root("machine_learning_job_id"), &diags)
	} else if multipleJobIDs, err := rule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId1(); err == nil {
		// Multiple job IDs
		jobIDStrings := make([]string, len(multipleJobIDs))
		copy(jobIDStrings, multipleJobIDs)
		d.MachineLearningJobID = typeutils.ListValueFrom(ctx, jobIDStrings, types.StringType, path.Root("machine_learning_job_id"), &diags)
	} else {
		d.MachineLearningJobID = types.ListValueMust(types.StringType, []attr.Value{})
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
