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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ThreatMatchRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPIThreatMatchRule]
}

func newThreatMatchRuleProcessor() ThreatMatchRuleProcessor {
	return ThreatMatchRuleProcessor{
		updateFn: func(ctx context.Context, v *kbapi.SecurityDetectionsAPIThreatMatchRule, d *Data) diag.Diagnostics {
			return d.updateFromThreatMatchRule(ctx, v)
		},
		idFn: func(v kbapi.SecurityDetectionsAPIThreatMatchRule) string {
			return v.Id.String()
		},
	}
}

func (t ThreatMatchRuleProcessor) HandlesRuleType(ruleType string) bool {
	return ruleType == ruleTypeThreatMatch
}

func (t ThreatMatchRuleProcessor) ToCreateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toThreatMatchRuleCreateProps(ctx)
}

func (t ThreatMatchRuleProcessor) ToUpdateProps(ctx context.Context, d Data) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toThreatMatchRuleUpdateProps(ctx)
}

func (d Data) toThreatMatchRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	threatMatchRule := kbapi.SecurityDetectionsAPIThreatMatchRuleCreateProps{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIThreatMatchRuleCreatePropsType(ruleTypeThreatMatch),
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

	d.setCommonCreateProps(ctx, buildCommonRuleProps(&threatMatchRule), &diags)

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
func (d Data) toThreatMatchRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	uid, ok := d.parseResourceUUID(&diags)
	if !ok {
		return updateProps, diags
	}

	threatMatchRule := kbapi.SecurityDetectionsAPIThreatMatchRuleUpdateProps{
		Id:          &uid,
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Type:        kbapi.SecurityDetectionsAPIThreatMatchRuleUpdatePropsType(ruleTypeThreatMatch),
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

	d.setCommonUpdateProps(ctx, buildCommonRuleProps(&threatMatchRule), &diags)

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
	err := updateProps.FromSecurityDetectionsAPIThreatMatchRuleUpdateProps(threatMatchRule)
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

	diags.Append(d.updateCommonRuleFieldsFromAPI(ctx, buildCommonAPIRuleFields(rule.Id.String(), rule))...)

	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))

	diags.Append(d.updateFiltersFromAPI(ctx, rule.Filters)...)

	// Threat Match-specific fields
	d.ThreatQuery = types.StringValue(rule.ThreatQuery)
	if len(rule.ThreatIndex) > 0 {
		d.ThreatIndex = typeutils.ListValueFrom(ctx, rule.ThreatIndex, types.StringType, path.Root("threat_index"), &diags)
	} else {
		d.ThreatIndex = typeutils.StringsToListMust(nil)
	}

	d.ThreatIndicatorPath = typeutils.StringishPointerValue(rule.ThreatIndicatorPath)

	d.ConcurrentSearches = typeutils.IntPointerToInt64Value(rule.ConcurrentSearches)
	d.ItemsPerSearch = typeutils.IntPointerToInt64Value(rule.ItemsPerSearch)

	diags.Append(d.updateThreatFiltersFromAPI(ctx, rule.ThreatFilters)...)

	d.SavedID = typeutils.StringishPointerValue(rule.SavedId)

	if len(rule.ThreatMapping) > 0 {
		listValue, threatMappingDiags := convertThreatMappingToModel(ctx, rule.ThreatMapping)
		diags.Append(threatMappingDiags...)
		if !threatMappingDiags.HasError() {
			d.ThreatMapping = listValue
		}
	}

	return diags
}
