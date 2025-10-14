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

type MachineLearningRuleProcessor struct{}

func (m MachineLearningRuleProcessor) HandlesRuleType(t string) bool {
	return t == "machine_learning"
}

func (m MachineLearningRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toMachineLearningRuleCreateProps(ctx, client)
}

func (m MachineLearningRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toMachineLearningRuleUpdateProps(ctx, client)
}

func (m MachineLearningRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	_, ok := rule.(kbapi.SecurityDetectionsAPIMachineLearningRule)
	return ok
}

func (m MachineLearningRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *SecurityDetectionRuleData) diag.Diagnostics {
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

func (m MachineLearningRuleProcessor) ExtractId(response any) (string, diag.Diagnostics) {
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

// applyMachineLearningValidations validates that Machine learning-specific constraints are met
func (d SecurityDetectionRuleData) applyMachineLearningValidations(diags *diag.Diagnostics) {
	if !utils.IsKnown(d.AnomalyThreshold) {
		diags.AddError(
			"Missing attribute 'anomaly_threshold'",
			"Machine learning rules require an 'anomaly_threshold' attribute.",
		)
	}
}

func (d SecurityDetectionRuleData) toMachineLearningRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	d.applyMachineLearningValidations(&diags)
	if diags.HasError() {
		return createProps, diags
	}

	mlRule := kbapi.SecurityDetectionsAPIMachineLearningRuleCreateProps{
		Name:             kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description:      kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:             kbapi.SecurityDetectionsAPIMachineLearningRuleCreatePropsType("machine_learning"),
		AnomalyThreshold: kbapi.SecurityDetectionsAPIAnomalyThreshold(d.AnomalyThreshold.ValueInt64()),
		RiskScore:        kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:         kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set ML job ID(s) - can be single string or array
	if utils.IsKnown(d.MachineLearningJobId) {
		jobIds := utils.ListTypeAs[string](ctx, d.MachineLearningJobId, path.Root("machine_learning_job_id"), &diags)
		if !diags.HasError() {
			var mlJobId kbapi.SecurityDetectionsAPIMachineLearningJobId
			err := mlJobId.FromSecurityDetectionsAPIMachineLearningJobId1(jobIds)
			if err != nil {
				diags.AddError("Error setting ML job IDs", err.Error())
			} else {
				mlRule.MachineLearningJobId = mlJobId
			}
		}
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &mlRule.Actions,
		ResponseActions:                   &mlRule.ResponseActions,
		RuleId:                            &mlRule.RuleId,
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
		DataViewId:                        nil, // ML rules don't have DataViewId
		Namespace:                         &mlRule.Namespace,
		RuleNameOverride:                  &mlRule.RuleNameOverride,
		TimestampOverride:                 &mlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &mlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &mlRule.InvestigationFields,
		Filters:                           nil, // ML rules don't have Filters
		Threat:                            &mlRule.Threat,
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
func (d SecurityDetectionRuleData) toMachineLearningRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	d.applyMachineLearningValidations(&diags)
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

	mlRule := kbapi.SecurityDetectionsAPIMachineLearningRuleUpdateProps{
		Id:               &id,
		Name:             kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description:      kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:             kbapi.SecurityDetectionsAPIMachineLearningRuleUpdatePropsType("machine_learning"),
		AnomalyThreshold: kbapi.SecurityDetectionsAPIAnomalyThreshold(d.AnomalyThreshold.ValueInt64()),
		RiskScore:        kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:         kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		mlRule.RuleId = &ruleId
		mlRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set ML job ID(s) - can be single string or array
	if utils.IsKnown(d.MachineLearningJobId) {
		jobIds := utils.ListTypeAs[string](ctx, d.MachineLearningJobId, path.Root("machine_learning_job_id"), &diags)
		if !diags.HasError() {
			var mlJobId kbapi.SecurityDetectionsAPIMachineLearningJobId
			err := mlJobId.FromSecurityDetectionsAPIMachineLearningJobId1(jobIds)
			if err != nil {
				diags.AddError("Error setting ML job IDs", err.Error())
			} else {
				mlRule.MachineLearningJobId = mlJobId
			}
		}
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &mlRule.Actions,
		ResponseActions:                   &mlRule.ResponseActions,
		RuleId:                            &mlRule.RuleId,
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
		DataViewId:                        nil, // ML rules don't have DataViewId
		Namespace:                         &mlRule.Namespace,
		RuleNameOverride:                  &mlRule.RuleNameOverride,
		TimestampOverride:                 &mlRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &mlRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &mlRule.InvestigationFields,
		Filters:                           nil, // ML rules don't have Filters
		Threat:                            &mlRule.Threat,
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

func (d *SecurityDetectionRuleData) updateFromMachineLearningRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIMachineLearningRule) diag.Diagnostics {
	var diags diag.Diagnostics

	compId := clients.CompositeId{
		ClusterId:  d.SpaceId.ValueString(),
		ResourceId: rule.Id.String(),
	}
	d.Id = types.StringValue(compId.String())

	d.RuleId = types.StringValue(string(rule.RuleId))
	d.Name = types.StringValue(string(rule.Name))
	d.Type = types.StringValue(string(rule.Type))

	// Update common fields (ML doesn't support DataViewId)
	d.DataViewId = types.StringNull()
	diags.Append(d.updateNamespaceFromApi(ctx, rule.Namespace)...)
	diags.Append(d.updateRuleNameOverrideFromApi(ctx, rule.RuleNameOverride)...)
	diags.Append(d.updateTimestampOverrideFromApi(ctx, rule.TimestampOverride)...)
	diags.Append(d.updateTimestampOverrideFallbackDisabledFromApi(ctx, rule.TimestampOverrideFallbackDisabled)...)

	d.Enabled = types.BoolValue(bool(rule.Enabled))
	d.From = types.StringValue(string(rule.From))
	d.To = types.StringValue(string(rule.To))
	d.Interval = types.StringValue(string(rule.Interval))
	d.Description = types.StringValue(string(rule.Description))
	d.RiskScore = types.Int64Value(int64(rule.RiskScore))
	d.Severity = types.StringValue(string(rule.Severity))

	// Update building block type
	diags.Append(d.updateBuildingBlockTypeFromApi(ctx, rule.BuildingBlockType)...)
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
	if singleJobId, err := rule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId0(); err == nil {
		// Single job ID
		d.MachineLearningJobId = utils.ListValueFrom(ctx, []string{string(singleJobId)}, types.StringType, path.Root("machine_learning_job_id"), &diags)
	} else if multipleJobIds, err := rule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId1(); err == nil {
		// Multiple job IDs
		jobIdStrings := make([]string, len(multipleJobIds))
		for i, jobId := range multipleJobIds {
			jobIdStrings[i] = string(jobId)
		}
		d.MachineLearningJobId = utils.ListValueFrom(ctx, jobIdStrings, types.StringType, path.Root("machine_learning_job_id"), &diags)
	} else {
		d.MachineLearningJobId = types.ListValueMust(types.StringType, []attr.Value{})
	}

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
