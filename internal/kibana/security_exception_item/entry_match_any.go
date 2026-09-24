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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// buildMatchAnyAPIEntry validates the entry values and builds the shared API match_any
// entry used by both the top-level and nested match_any conversions.
func buildMatchAnyAPIEntry(
	ctx context.Context,
	values types.List,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
	missingValuesMessage string,
	emptyValuesMessage string,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchAny, diag.Diagnostics) {
	var diags diag.Diagnostics
	var apiEntry kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchAny

	// Validate required field
	if !typeutils.IsKnown(values) {
		diags.AddError("Invalid Configuration", missingValuesMessage)
		return apiEntry, diags
	}

	stringValues := typeutils.ListTypeAs[string](ctx, values, path.Empty(), &diags)
	if diags.HasError() {
		return apiEntry, diags
	}

	if len(stringValues) == 0 {
		diags.AddError("Invalid Configuration", emptyValuesMessage)
		return apiEntry, diags
	}

	apiEntry = kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchAny{
		Type:     entryTypeMatchAny,
		Field:    field,
		Operator: operator,
		Value:    stringValues,
	}
	return apiEntry, diags
}

// convertMatchAnyEntryToAPI converts a match_any entry to API format
func convertMatchAnyEntryToAPI(
	ctx context.Context,
	entry EntryModel,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	apiEntry, diags := buildMatchAnyAPIEntry(
		ctx, entry.Values, field, operator,
		"Attribute 'values' is required when type is 'match_any'",
		"Attribute 'values' must contain at least one value when type is 'match_any'",
	)
	if diags.HasError() {
		return result, diags
	}

	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatchAny(apiEntry); err != nil {
		diags.AddError("Failed to create match_any entry", err.Error())
	}

	return result, diags
}

// convertNestedMatchAnyEntryToAPI converts a nested match_any entry to API format
func convertNestedMatchAnyEntryToAPI(
	ctx context.Context,
	entry NestedEntryModel,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, diag.Diagnostics) {
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem

	apiEntry, diags := buildMatchAnyAPIEntry(
		ctx, entry.Values, field, operator,
		"Attribute 'values' is required for nested entry when type is 'match_any'",
		"Attribute 'values' must contain at least one value for nested entry when type is 'match_any'",
	)
	if diags.HasError() {
		return result, diags
	}

	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatchAny(apiEntry); err != nil {
		diags.AddError("Failed to create nested match_any entry", err.Error())
	}

	return result, diags
}

// extractValuesFromMap reads the "value" string-array field shared by match_any entries.
func extractValuesFromMap(ctx context.Context, entryMap map[string]any) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	values, ok := entryMap["value"].([]any)
	if !ok {
		return types.ListNull(types.StringType), diags
	}

	strValues := make([]string, 0, len(values))
	for _, v := range values {
		if str, ok := v.(string); ok {
			strValues = append(strValues, str)
		}
	}
	list, d := types.ListValueFrom(ctx, types.StringType, strValues)
	diags.Append(d...)
	return list, diags
}

// convertMatchAnyEntryFromAPI converts match_any entries from API format
func convertMatchAnyEntryFromAPI(ctx context.Context, entryMap map[string]any, entry *EntryModel) diag.Diagnostics {
	values, diags := extractValuesFromMap(ctx, entryMap)
	entry.Values = values
	entry.Value = types.StringNull()
	entry.List = types.ObjectNull(getListAttrTypes())
	entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})
	return diags
}

// convertNestedMatchAnyFromMap converts nested match_any entries from map format
func convertNestedMatchAnyFromMap(ctx context.Context, entryMap map[string]any, entry *NestedEntryModel) diag.Diagnostics {
	values, diags := extractValuesFromMap(ctx, entryMap)
	entry.Values = values
	entry.Value = types.StringNull()
	return diags
}
