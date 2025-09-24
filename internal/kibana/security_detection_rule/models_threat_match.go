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

func (d SecurityDetectionRuleData) toThreatMatchRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	threatMatchRule := kbapi.SecurityDetectionsAPIThreatMatchRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIThreatMatchRuleCreatePropsType("threat_match"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set threat index
	if utils.IsKnown(d.ThreatIndex) {
		threatIndex := utils.ListTypeAs[string](ctx, d.ThreatIndex, path.Root("threat_index"), &diags)
		if !diags.HasError() {
			threatMatchRule.ThreatIndex = threatIndex
		}
	}

	if utils.IsKnown(d.ThreatMapping) && len(d.ThreatMapping.Elements()) > 0 {
		apiThreatMapping, threatMappingDiags := d.threatMappingToApi(ctx)
		if !threatMappingDiags.HasError() {
			threatMatchRule.ThreatMapping = apiThreatMapping
		}
		diags.Append(threatMappingDiags...)
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &threatMatchRule.Actions,
		ResponseActions:                   &threatMatchRule.ResponseActions,
		RuleId:                            &threatMatchRule.RuleId,
		Enabled:                           &threatMatchRule.Enabled,
		From:                              &threatMatchRule.From,
		To:                                &threatMatchRule.To,
		Interval:                          &threatMatchRule.Interval,
		Index:                             &threatMatchRule.Index,
		Author:                            &threatMatchRule.Author,
		Tags:                              &threatMatchRule.Tags,
		FalsePositives:                    &threatMatchRule.FalsePositives,
		References:                        &threatMatchRule.References,
		License:                           &threatMatchRule.License,
		Note:                              &threatMatchRule.Note,
		Setup:                             &threatMatchRule.Setup,
		MaxSignals:                        &threatMatchRule.MaxSignals,
		Version:                           &threatMatchRule.Version,
		ExceptionsList:                    &threatMatchRule.ExceptionsList,
		AlertSuppression:                  &threatMatchRule.AlertSuppression,
		RiskScoreMapping:                  &threatMatchRule.RiskScoreMapping,
		SeverityMapping:                   &threatMatchRule.SeverityMapping,
		RelatedIntegrations:               &threatMatchRule.RelatedIntegrations,
		RequiredFields:                    &threatMatchRule.RequiredFields,
		BuildingBlockType:                 &threatMatchRule.BuildingBlockType,
		DataViewId:                        &threatMatchRule.DataViewId,
		Namespace:                         &threatMatchRule.Namespace,
		RuleNameOverride:                  &threatMatchRule.RuleNameOverride,
		TimestampOverride:                 &threatMatchRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &threatMatchRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &threatMatchRule.InvestigationFields,
		Meta:                              &threatMatchRule.Meta,
		Filters:                           &threatMatchRule.Filters,
	}, &diags)

	// Set threat-specific fields
	if utils.IsKnown(d.ThreatQuery) {
		threatMatchRule.ThreatQuery = kbapi.SecurityDetectionsAPIThreatQuery(d.ThreatQuery.ValueString())
	}

	if utils.IsKnown(d.ThreatIndicatorPath) {
		threatIndicatorPath := kbapi.SecurityDetectionsAPIThreatIndicatorPath(d.ThreatIndicatorPath.ValueString())
		threatMatchRule.ThreatIndicatorPath = &threatIndicatorPath
	}

	if utils.IsKnown(d.ConcurrentSearches) {
		concurrentSearches := kbapi.SecurityDetectionsAPIConcurrentSearches(d.ConcurrentSearches.ValueInt64())
		threatMatchRule.ConcurrentSearches = &concurrentSearches
	}

	if utils.IsKnown(d.ItemsPerSearch) {
		itemsPerSearch := kbapi.SecurityDetectionsAPIItemsPerSearch(d.ItemsPerSearch.ValueInt64())
		threatMatchRule.ItemsPerSearch = &itemsPerSearch
	}

	// Set query language
	threatMatchRule.Language = d.getKQLQueryLanguage()

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		threatMatchRule.SavedId = &savedId
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIThreatMatchRuleCreateProps(threatMatchRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert threat match rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}
func (d SecurityDetectionRuleData) toThreatMatchRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	threatMatchRule := kbapi.SecurityDetectionsAPIThreatMatchRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIThreatMatchRuleUpdatePropsType("threat_match"),
		Query:       kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString()),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		threatMatchRule.RuleId = &ruleId
		threatMatchRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set threat index
	if utils.IsKnown(d.ThreatIndex) {
		threatIndex := utils.ListTypeAs[string](ctx, d.ThreatIndex, path.Root("threat_index"), &diags)
		if !diags.HasError() {
			threatMatchRule.ThreatIndex = threatIndex
		}
	}

	if utils.IsKnown(d.ThreatMapping) && len(d.ThreatMapping.Elements()) > 0 {
		apiThreatMapping, threatMappingDiags := d.threatMappingToApi(ctx)
		if !threatMappingDiags.HasError() {
			threatMatchRule.ThreatMapping = apiThreatMapping
		}
		diags.Append(threatMappingDiags...)
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &threatMatchRule.Actions,
		ResponseActions:                   &threatMatchRule.ResponseActions,
		RuleId:                            &threatMatchRule.RuleId,
		Enabled:                           &threatMatchRule.Enabled,
		From:                              &threatMatchRule.From,
		To:                                &threatMatchRule.To,
		Interval:                          &threatMatchRule.Interval,
		Index:                             &threatMatchRule.Index,
		Author:                            &threatMatchRule.Author,
		Tags:                              &threatMatchRule.Tags,
		FalsePositives:                    &threatMatchRule.FalsePositives,
		References:                        &threatMatchRule.References,
		License:                           &threatMatchRule.License,
		Note:                              &threatMatchRule.Note,
		InvestigationFields:               &threatMatchRule.InvestigationFields,
		Meta:                              &threatMatchRule.Meta,
		Setup:                             &threatMatchRule.Setup,
		MaxSignals:                        &threatMatchRule.MaxSignals,
		Version:                           &threatMatchRule.Version,
		ExceptionsList:                    &threatMatchRule.ExceptionsList,
		AlertSuppression:                  &threatMatchRule.AlertSuppression,
		RiskScoreMapping:                  &threatMatchRule.RiskScoreMapping,
		SeverityMapping:                   &threatMatchRule.SeverityMapping,
		RelatedIntegrations:               &threatMatchRule.RelatedIntegrations,
		RequiredFields:                    &threatMatchRule.RequiredFields,
		BuildingBlockType:                 &threatMatchRule.BuildingBlockType,
		DataViewId:                        &threatMatchRule.DataViewId,
		Namespace:                         &threatMatchRule.Namespace,
		RuleNameOverride:                  &threatMatchRule.RuleNameOverride,
		TimestampOverride:                 &threatMatchRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &threatMatchRule.TimestampOverrideFallbackDisabled,
		Filters:                           &threatMatchRule.Filters,
	}, &diags)

	// Set threat-specific fields
	if utils.IsKnown(d.ThreatQuery) {
		threatMatchRule.ThreatQuery = kbapi.SecurityDetectionsAPIThreatQuery(d.ThreatQuery.ValueString())
	}

	if utils.IsKnown(d.ThreatIndicatorPath) {
		threatIndicatorPath := kbapi.SecurityDetectionsAPIThreatIndicatorPath(d.ThreatIndicatorPath.ValueString())
		threatMatchRule.ThreatIndicatorPath = &threatIndicatorPath
	}

	if utils.IsKnown(d.ConcurrentSearches) {
		concurrentSearches := kbapi.SecurityDetectionsAPIConcurrentSearches(d.ConcurrentSearches.ValueInt64())
		threatMatchRule.ConcurrentSearches = &concurrentSearches
	}

	if utils.IsKnown(d.ItemsPerSearch) {
		itemsPerSearch := kbapi.SecurityDetectionsAPIItemsPerSearch(d.ItemsPerSearch.ValueInt64())
		threatMatchRule.ItemsPerSearch = &itemsPerSearch
	}

	// Set query language
	threatMatchRule.Language = d.getKQLQueryLanguage()

	if utils.IsKnown(d.SavedId) {
		savedId := kbapi.SecurityDetectionsAPISavedQueryId(d.SavedId.ValueString())
		threatMatchRule.SavedId = &savedId
	}

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIThreatMatchRuleUpdateProps(threatMatchRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert threat match rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d *SecurityDetectionRuleData) updateFromThreatMatchRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIThreatMatchRule) diag.Diagnostics {
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

	// Update building block type
	if rule.BuildingBlockType != nil {
		d.BuildingBlockType = types.StringValue(string(*rule.BuildingBlockType))
	} else {
		d.BuildingBlockType = types.StringNull()
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

	// Threat Match-specific fields
	d.ThreatQuery = types.StringValue(string(rule.ThreatQuery))
	if len(rule.ThreatIndex) > 0 {
		d.ThreatIndex = utils.ListValueFrom(ctx, rule.ThreatIndex, types.StringType, path.Root("threat_index"), &diags)
	} else {
		d.ThreatIndex = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if rule.ThreatIndicatorPath != nil {
		d.ThreatIndicatorPath = types.StringValue(string(*rule.ThreatIndicatorPath))
	} else {
		d.ThreatIndicatorPath = types.StringNull()
	}

	if rule.ConcurrentSearches != nil {
		d.ConcurrentSearches = types.Int64Value(int64(*rule.ConcurrentSearches))
	} else {
		d.ConcurrentSearches = types.Int64Null()
	}

	if rule.ItemsPerSearch != nil {
		d.ItemsPerSearch = types.Int64Value(int64(*rule.ItemsPerSearch))
	} else {
		d.ItemsPerSearch = types.Int64Null()
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

	// Convert threat mapping
	if len(rule.ThreatMapping) > 0 {
		listValue, threatMappingDiags := convertThreatMappingToModel(ctx, rule.ThreatMapping)
		diags.Append(threatMappingDiags...)
		if !threatMappingDiags.HasError() {
			d.ThreatMapping = listValue
		}
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
