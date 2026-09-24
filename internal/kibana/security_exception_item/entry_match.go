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

package securityexceptionitem

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// buildMatchAPIEntry validates the entry value and builds the shared API match entry
// used by both the top-level and nested match conversions.
func buildMatchAPIEntry(
	value types.String,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
	missingValueMessage string,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntryMatch, diag.Diagnostics) {
	var diags diag.Diagnostics
	var apiEntry kbapi.SecurityExceptionsAPIExceptionListItemEntryMatch

	if !typeutils.IsKnown(value) || value.ValueString() == "" {
		diags.AddError("Invalid Configuration", missingValueMessage)
		return apiEntry, diags
	}

	apiEntry = kbapi.SecurityExceptionsAPIExceptionListItemEntryMatch{
		Type:     entryTypeMatch,
		Field:    field,
		Operator: operator,
		Value:    value.ValueString(),
	}
	return apiEntry, diags
}

// convertMatchEntryToAPI converts a match entry to API format
func convertMatchEntryToAPI(
	entry EntryModel,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	apiEntry, diags := buildMatchAPIEntry(entry.Value, field, operator, "Attribute 'value' is required when type is 'match'")
	if diags.HasError() {
		return result, diags
	}

	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatch(apiEntry); err != nil {
		diags.AddError("Failed to create match entry", err.Error())
	}

	return result, diags
}

// convertWildcardEntryToAPI converts a wildcard entry to API format
func convertWildcardEntryToAPI(
	entry EntryModel,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	// Validate required field
	if !typeutils.IsKnown(entry.Value) || entry.Value.ValueString() == "" {
		diags.AddError("Invalid Configuration", "Attribute 'value' is required when type is 'wildcard'")
		return result, diags
	}

	apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchWildcard{
		Type:     "wildcard",
		Field:    field,
		Operator: operator,
		Value:    entry.Value.ValueString(),
	}
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatchWildcard(apiEntry); err != nil {
		diags.AddError("Failed to create wildcard entry", err.Error())
	}

	return result, diags
}

// convertNestedMatchEntryToAPI converts a nested match entry to API format
func convertNestedMatchEntryToAPI(
	entry NestedEntryModel,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, diag.Diagnostics) {
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem

	apiEntry, diags := buildMatchAPIEntry(entry.Value, field, operator, "Attribute 'value' is required for nested entry when type is 'match'")
	if diags.HasError() {
		return result, diags
	}

	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatch(apiEntry); err != nil {
		diags.AddError("Failed to create nested match entry", err.Error())
	}

	return result, diags
}

// extractValueFromMap reads the "value" string field shared by match/wildcard entries.
func extractValueFromMap(entryMap map[string]any) types.String {
	if value, ok := entryMap["value"].(string); ok {
		return types.StringValue(value)
	}
	return types.StringNull()
}

// convertMatchOrWildcardEntryFromAPI converts match or wildcard entries from API format
func convertMatchOrWildcardEntryFromAPI(entryMap map[string]any, entry *EntryModel) {
	entry.Value = extractValueFromMap(entryMap)
	entry.Values = types.ListNull(types.StringType)
	entry.List = types.ObjectNull(getListAttrTypes())
	entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})
}

// convertNestedMatchFromMap converts nested match entries from map format
func convertNestedMatchFromMap(entryMap map[string]any, entry *NestedEntryModel) {
	entry.Value = extractValueFromMap(entryMap)
	entry.Values = types.ListNull(types.StringType)
}
