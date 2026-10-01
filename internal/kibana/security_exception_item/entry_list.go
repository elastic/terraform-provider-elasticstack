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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// convertListEntryToAPI converts a list entry to API format
func convertListEntryToAPI(
	ctx context.Context,
	entry EntryModel,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	// Validate required field
	if !typeutils.IsKnown(entry.List) {
		diags.AddError("Invalid Configuration", "Attribute 'list' is required when type is 'list'")
		return result, diags
	}

	var listModel EntryListModel
	diags.Append(entry.List.As(ctx, &listModel, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return result, diags
	}
	apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryList{
		Type:     entryTypeList,
		Field:    field,
		Operator: operator,
	}
	apiEntry.List.Id = listModel.ID.ValueString()
	apiEntry.List.Type = kbapi.SecurityExceptionsAPIListType(listModel.Type.ValueString())
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryList(apiEntry); err != nil {
		diags.AddError("Failed to create list entry", err.Error())
	}

	return result, diags
}

// convertListEntryFromAPI converts list entries from API format
func convertListEntryFromAPI(ctx context.Context, entryMap map[string]any, entry *EntryModel) diag.Diagnostics {
	var diags diag.Diagnostics
	resetEntryModelFields(entry, entryTypeList)

	if listData, ok := entryMap["list"].(map[string]any); ok {
		listModel := EntryListModel{
			ID:   types.StringValue(listData["id"].(string)),
			Type: types.StringValue(listData["type"].(string)),
		}
		obj, d := types.ObjectValueFrom(ctx, getListAttrTypes(), listModel)
		diags.Append(d...)
		entry.List = obj
	} else {
		entry.List = types.ObjectNull(getListAttrTypes())
	}
	return diags
}
