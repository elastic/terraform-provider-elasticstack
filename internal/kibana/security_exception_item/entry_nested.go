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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// convertNestedEntryArrayToAPI converts nested entries to API format
func convertNestedEntryArrayToAPI(ctx context.Context, entry EntryModel, field kbapi.SecurityExceptionsAPINonEmptyString) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	// Validate required field
	if !typeutils.IsKnown(entry.Entries) {
		diags.AddError("Invalid Configuration", "Attribute 'entries' is required when type is 'nested'")
		return result, diags
	}

	nestedEntries := typeutils.ListTypeAs[NestedEntryModel](ctx, entry.Entries, path.Empty(), &diags)
	if diags.HasError() {
		return result, diags
	}

	if len(nestedEntries) == 0 {
		diags.AddError("Invalid Configuration", "Attribute 'entries' must contain at least one entry when type is 'nested'")
		return result, diags
	}

	apiNestedEntries := make([]kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, 0, len(nestedEntries))
	for _, ne := range nestedEntries {
		nestedAPIEntry, d := convertNestedEntryToAPI(ctx, ne)
		diags.Append(d...)
		if d.HasError() {
			continue
		}
		apiNestedEntries = append(apiNestedEntries, nestedAPIEntry)
	}

	apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryNested{
		Type:    "nested",
		Field:   field,
		Entries: apiNestedEntries,
	}
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryNested(apiEntry); err != nil {
		diags.AddError("Failed to create nested entry", err.Error())
	}

	return result, diags
}

// convertNestedEntryToAPI converts a nested entry model to an API nested entry model
func convertNestedEntryToAPI(ctx context.Context, entry NestedEntryModel) (kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem

	entryType := entry.Type.ValueString()
	operator := kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator(entry.Operator.ValueString())
	field := entry.Field.ValueString()

	switch entryType {
	case entryTypeMatch:
		return convertNestedMatchEntryToAPI(entry, field, operator)
	case entryTypeMatchAny:
		return convertNestedMatchAnyEntryToAPI(ctx, entry, field, operator)
	case entryTypeExists:
		return convertNestedExistsEntryToAPI(field, operator)
	default:
		diags.AddError("Invalid nested entry type", fmt.Sprintf("Unknown nested entry type: %s. Only 'match', 'match_any', and 'exists' are allowed.", entryType))
		return result, diags
	}
}

// convertNestedEntryFromAPI converts nested entries from API format
func convertNestedEntryFromAPI(ctx context.Context, entryMap map[string]any, entry *EntryModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Nested entries don't have an operator field in the API
	entry.Operator = types.StringNull()
	resetEntryModelFields(entry, attrEntries)
	if entriesData, ok := entryMap["entries"].([]any); ok {
		nestedEntries := make([]NestedEntryModel, 0, len(entriesData))
		for _, neData := range entriesData {
			if neMap, ok := neData.(map[string]any); ok {
				ne, d := convertNestedEntryFromMap(ctx, neMap)
				diags.Append(d...)
				if !d.HasError() {
					nestedEntries = append(nestedEntries, ne)
				}
			}
		}
		list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getNestedEntryAttrTypes()}, nestedEntries)
		diags.Append(d...)
		entry.Entries = list
	} else {
		entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})
	}
	return diags
}

// convertNestedEntryFromMap converts a map representation of nested entry to a model
func convertNestedEntryFromMap(ctx context.Context, entryMap map[string]any) (NestedEntryModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var entry NestedEntryModel

	if entryType, ok := entryMap["type"].(string); ok {
		entry.Type = types.StringValue(entryType)
	}
	if field, ok := entryMap["field"].(string); ok {
		entry.Field = types.StringValue(field)
	}
	if operator, ok := entryMap["operator"].(string); ok {
		entry.Operator = types.StringValue(operator)
	}

	entryType := entry.Type.ValueString()
	switch entryType {
	case entryTypeMatch:
		convertNestedMatchFromMap(entryMap, &entry)
	case entryTypeMatchAny:
		d := convertNestedMatchAnyFromMap(ctx, entryMap, &entry)
		diags.Append(d...)
	case entryTypeExists:
		convertNestedExistsFromMap(&entry)
	}

	return entry, diags
}
