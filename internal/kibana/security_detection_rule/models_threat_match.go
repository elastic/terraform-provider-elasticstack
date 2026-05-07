// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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

type ThreatMatchRuleProcessor struct{}

func (t ThreatMatchRuleProcessor) HandlesRuleType(ruleType string) bool {
	return ruleType == "threat_match"
}

func (t ThreatMatchRuleProcessor) ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toThreatMatchRuleCreateProps(ctx, client)
}

func (t ThreatMatchRuleProcessor) ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toThreatMatchRuleUpdateProps(ctx, client)
}

func (t ThreatMatchRuleProcessor) HandlesAPIRuleResponse(rule any) bool {
	return handlesAPIRuleResponse[kbapi.SecurityDetectionsAPIThreatMatchRule](rule)
}

func (t ThreatMatchRuleProcessor) UpdateFromResponse(ctx context.Context, rule any, d *Data) diag.Diagnostics {
	return updateFromRuleResponse[kbapi.SecurityDetectionsAPIThreatMatchRule](rule, func(v *kbapi.SecurityDetectionsAPIThreatMatchRule) diag.Diagnostics {
		return d.updateFromThreatMatchRule(ctx, v)
	})
}

func (t ThreatMatchRuleProcessor) ExtractID(response any) (string, diag.Diagnostics) {
	return extractRuleID[kbapi.SecurityDetectionsAPIThreatMatchRule](response, func(v kbapi.SecurityDetectionsAPIThreatMatchRule) string {
		return v.Id.String()
	})
}

func (d Data) toThreatMatchRuleCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	threatMatchRule := kbapi.SecurityDetectionsAPIThreatMatchRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIThreatMatchRuleCreatePropsType("threat_match"),
		Query:       d.Query.ValueString(),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set threat index
	if typeutils.IsKnown(d.ThreatIndex) {
		threatIndex := typeutils.ListTypeAs[string](ctx, d.ThreatIndex, path.Root("threat_index"), &diags)
		if !diags.HasError() {
			threatMatchRule.ThreatIndex = threatIndex
		}
	}

	if typeutils.IsKnown(d.ThreatMapping) && len(d.ThreatMapping.Elements()) > 0 {
		apiThreatMapping, threatMappingDiags := d.threatMappingToAPI(ctx)
		if !threatMappingDiags.HasError() {
			threatMatchRule.ThreatMapping = apiThreatMapping
		}
		diags.Append(threatMappingDiags...)
	}

	apiThreatFilters, threatFiltersDiags := d.threatFiltersToAPI(ctx)
	diags.Append(threatFiltersDiags...)
	if !threatFiltersDiags.HasError() && apiThreatFilters != nil {
		threatMatchRule.ThreatFilters = apiThreatFilters
	}

	d.setCommonCreateProps(ctx, &CommonCreateProps{
		Actions:                           &threatMatchRule.Actions,
		ResponseActions:                   &threatMatchRule.ResponseActions,
		RuleID:                            &threatMatchRule.RuleId,
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
		DataViewID:                        &threatMatchRule.DataViewId,
		Namespace:                         &threatMatchRule.Namespace,
		RuleNameOverride:                  &threatMatchRule.RuleNameOverride,
		TimestampOverride:                 &threatMatchRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &threatMatchRule.TimestampOverrideFallbackDisabled,
		InvestigationFields:               &threatMatchRule.InvestigationFields,
		Filters:                           &threatMatchRule.Filters,
		Threat:                            &threatMatchRule.Threat,
		TimelineID:                        &threatMatchRule.TimelineId,
		TimelineTitle:                     &threatMatchRule.TimelineTitle,
	}, &diags, client)

	// Set threat-specific fields
	if typeutils.IsKnown(d.ThreatQuery) {
		threatMatchRule.ThreatQuery = d.ThreatQuery.ValueString()
	}

	if typeutils.IsKnown(d.ThreatIndicatorPath) {
		threatIndicatorPath := d.ThreatIndicatorPath.ValueString()
		threatMatchRule.ThreatIndicatorPath = &threatIndicatorPath
	}

	if typeutils.IsKnown(d.ConcurrentSearches) {
		concurrentSearches := kbapi.SecurityDetectionsAPIConcurrentSearches(d.ConcurrentSearches.ValueInt64())
		threatMatchRule.ConcurrentSearches = &concurrentSearches
	}

	if typeutils.IsKnown(d.ItemsPerSearch) {
		itemsPerSearch := kbapi.SecurityDetectionsAPIItemsPerSearch(d.ItemsPerSearch.ValueInt64())
		threatMatchRule.ItemsPerSearch = &itemsPerSearch
	}

	// Set query language
	threatMatchRule.Language = d.getKQLQueryLanguage()

	if typeutils.IsKnown(d.SavedID) {
		savedID := d.SavedID.ValueString()
		threatMatchRule.SavedId = &savedID
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
func (d Data) toThreatMatchRuleUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
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

	threatMatchRule := kbapi.SecurityDetectionsAPIThreatMatchRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIThreatMatchRuleUpdatePropsType("threat_match"),
		Query:       d.Query.ValueString(),
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		threatMatchRule.RuleId = &ruleID
		threatMatchRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set threat index
	if typeutils.IsKnown(d.ThreatIndex) {
		threatIndex := typeutils.ListTypeAs[string](ctx, d.ThreatIndex, path.Root("threat_index"), &diags)
		if !diags.HasError() {
			threatMatchRule.ThreatIndex = threatIndex
		}
	}

	if typeutils.IsKnown(d.ThreatMapping) && len(d.ThreatMapping.Elements()) > 0 {
		apiThreatMapping, threatMappingDiags := d.threatMappingToAPI(ctx)
		if !threatMappingDiags.HasError() {
			threatMatchRule.ThreatMapping = apiThreatMapping
		}
		diags.Append(threatMappingDiags...)
	}

	apiThreatFilters, threatFiltersDiags := d.threatFiltersToAPI(ctx)
	diags.Append(threatFiltersDiags...)
	if !threatFiltersDiags.HasError() && apiThreatFilters != nil {
		threatMatchRule.ThreatFilters = apiThreatFilters
	}

	d.setCommonUpdateProps(ctx, &CommonUpdateProps{
		Actions:                           &threatMatchRule.Actions,
		ResponseActions:                   &threatMatchRule.ResponseActions,
		RuleID:                            &threatMatchRule.RuleId,
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
		DataViewID:                        &threatMatchRule.DataViewId,
		Namespace:                         &threatMatchRule.Namespace,
		RuleNameOverride:                  &threatMatchRule.RuleNameOverride,
		TimestampOverride:                 &threatMatchRule.TimestampOverride,
		TimestampOverrideFallbackDisabled: &threatMatchRule.TimestampOverrideFallbackDisabled,
		Filters:                           &threatMatchRule.Filters,
		Threat:                            &threatMatchRule.Threat,
		TimelineID:                        &threatMatchRule.TimelineId,
		TimelineTitle:                     &threatMatchRule.TimelineTitle,
	}, &diags, client)

	// Set threat-specific fields
	if typeutils.IsKnown(d.ThreatQuery) {
		threatMatchRule.ThreatQuery = d.ThreatQuery.ValueString()
	}

	if typeutils.IsKnown(d.ThreatIndicatorPath) {
		threatIndicatorPath := d.ThreatIndicatorPath.ValueString()
		threatMatchRule.ThreatIndicatorPath = &threatIndicatorPath
	}

	if typeutils.IsKnown(d.ConcurrentSearches) {
		concurrentSearches := kbapi.SecurityDetectionsAPIConcurrentSearches(d.ConcurrentSearches.ValueInt64())
		threatMatchRule.ConcurrentSearches = &concurrentSearches
	}

	if typeutils.IsKnown(d.ItemsPerSearch) {
		itemsPerSearch := kbapi.SecurityDetectionsAPIItemsPerSearch(d.ItemsPerSearch.ValueInt64())
		threatMatchRule.ItemsPerSearch = &itemsPerSearch
	}

	// Set query language
	threatMatchRule.Language = d.getKQLQueryLanguage()

	if typeutils.IsKnown(d.SavedID) {
		savedID := d.SavedID.ValueString()
		threatMatchRule.SavedId = &savedID
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

func (d *Data) updateFromThreatMatchRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIThreatMatchRule) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(d.updateCommonRuleFieldsFromAPI(ctx, commonAPIRuleFields{
		ResourceID:                        rule.Id.String(),
		RuleID:                            rule.RuleId,
		Name:                              rule.Name,
		Type:                              string(rule.Type),
		Enabled:                           rule.Enabled,
		From:                              rule.From,
		To:                                rule.To,
		Interval:                          rule.Interval,
		Description:                       rule.Description,
		RiskScore:                         int64(rule.RiskScore),
		Severity:                          string(rule.Severity),
		MaxSignals:                        int64(rule.MaxSignals),
		Version:                           int64(rule.Version),
		Revision:                          int64(rule.Revision),
		CreatedAt:                         rule.CreatedAt,
		CreatedBy:                         rule.CreatedBy,
		UpdatedAt:                         rule.UpdatedAt,
		UpdatedBy:                         rule.UpdatedBy,
		TimelineID:                        rule.TimelineId,
		TimelineTitle:                     rule.TimelineTitle,
		DataViewID:                        rule.DataViewId,
		Namespace:                         rule.Namespace,
		RuleNameOverride:                  rule.RuleNameOverride,
		TimestampOverride:                 rule.TimestampOverride,
		TimestampOverrideFallbackDisabled: rule.TimestampOverrideFallbackDisabled,
		BuildingBlockType:                 rule.BuildingBlockType,
		License:                           rule.License,
		Note:                              rule.Note,
		Setup:                             rule.Setup,
		Index:                             rule.Index,
		Author:                            rule.Author,
		Tags:                              rule.Tags,
		FalsePositives:                    rule.FalsePositives,
		References:                        rule.References,
		Actions:                           rule.Actions,
		ExceptionsList:                    rule.ExceptionsList,
		RiskScoreMapping:                  rule.RiskScoreMapping,
		InvestigationFields:               rule.InvestigationFields,
		Threat:                            rule.Threat,
		SeverityMapping:                   rule.SeverityMapping,
		RelatedIntegrations:               rule.RelatedIntegrations,
		RequiredFields:                    rule.RequiredFields,
		AlertSuppression:                  rule.AlertSuppression,
		ResponseActions:                   rule.ResponseActions,
	})...)

	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))

	diags.Append(d.updateFiltersFromAPI(ctx, rule.Filters)...)

	// Threat Match-specific fields
	d.ThreatQuery = types.StringValue(rule.ThreatQuery)
	if len(rule.ThreatIndex) > 0 {
		d.ThreatIndex = typeutils.ListValueFrom(ctx, rule.ThreatIndex, types.StringType, path.Root("threat_index"), &diags)
	} else {
		d.ThreatIndex = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if rule.ThreatIndicatorPath != nil {
		d.ThreatIndicatorPath = types.StringValue(*rule.ThreatIndicatorPath)
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

	diags.Append(d.updateThreatFiltersFromAPI(ctx, rule.ThreatFilters)...)

	if rule.SavedId != nil {
		d.SavedID = types.StringValue(*rule.SavedId)
	} else {
		d.SavedID = types.StringNull()
	}

	if len(rule.ThreatMapping) > 0 {
		listValue, threatMappingDiags := convertThreatMappingToModel(ctx, rule.ThreatMapping)
		diags.Append(threatMappingDiags...)
		if !threatMappingDiags.HasError() {
			d.ThreatMapping = listValue
		}
	}

	return diags
}
