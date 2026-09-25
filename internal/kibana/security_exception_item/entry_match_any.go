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

// convertMatchAnyEntryToAPI converts a match_any entry to API format
func convertMatchAnyEntryToAPI(
	ctx context.Context,
	entry EntryModel,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	// Validate required field
	if !typeutils.IsKnown(entry.Values) {
		diags.AddError("Invalid Configuration", "Attribute 'values' is required when type is 'match_any'")
		return result, diags
	}

	values := typeutils.ListTypeAs[string](ctx, entry.Values, path.Empty(), &diags)
	if diags.HasError() {
		return result, diags
	}

	if len(values) == 0 {
		diags.AddError("Invalid Configuration", "Attribute 'values' must contain at least one value when type is 'match_any'")
		return result, diags
	}

	apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchAny{
		Type:     entryTypeMatchAny,
		Field:    field,
		Operator: operator,
		Value:    values,
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
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem

	// Validate required field
	if !typeutils.IsKnown(entry.Values) {
		diags.AddError("Invalid Configuration", "Attribute 'values' is required for nested entry when type is 'match_any'")
		return result, diags
	}

	values := typeutils.ListTypeAs[string](ctx, entry.Values, path.Empty(), &diags)
	if diags.HasError() {
		return result, diags
	}

	if len(values) == 0 {
		diags.AddError("Invalid Configuration", "Attribute 'values' must contain at least one value for nested entry when type is 'match_any'")
		return result, diags
	}

	apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchAny{
		Type:     entryTypeMatchAny,
		Field:    field,
		Operator: operator,
		Value:    values,
	}
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatchAny(apiEntry); err != nil {
		diags.AddError("Failed to create nested match_any entry", err.Error())
	}

	return result, diags
}

// convertMatchAnyEntryFromAPI converts match_any entries from API format
func convertMatchAnyEntryFromAPI(ctx context.Context, entryMap map[string]any, entry *EntryModel) diag.Diagnostics {
	var diags diag.Diagnostics
	resetEntryModelFields(entry, attrValues)

	if values, ok := entryMap["value"].([]any); ok {
		strValues := make([]string, 0, len(values))
		for _, v := range values {
			if str, ok := v.(string); ok {
				strValues = append(strValues, str)
			}
		}
		list, d := types.ListValueFrom(ctx, types.StringType, strValues)
		diags.Append(d...)
		entry.Values = list
	} else {
		entry.Values = types.ListNull(types.StringType)
	}
	return diags
}

// convertNestedMatchAnyFromMap converts nested match_any entries from map format
func convertNestedMatchAnyFromMap(ctx context.Context, entryMap map[string]any, entry *NestedEntryModel) diag.Diagnostics {
	var diags diag.Diagnostics
	resetNestedEntryModelFields(entry, attrValues)

	if values, ok := entryMap["value"].([]any); ok {
		strValues := make([]string, 0, len(values))
		for _, v := range values {
			if str, ok := v.(string); ok {
				strValues = append(strValues, str)
			}
		}
		list, d := types.ListValueFrom(ctx, types.StringType, strValues)
		diags.Append(d...)
		entry.Values = list
	} else {
		entry.Values = types.ListNull(types.StringType)
	}
	return diags
}
