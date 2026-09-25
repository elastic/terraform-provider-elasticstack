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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// buildExistsAPIEntry builds the shared API exists entry used by both the
// top-level and nested exists conversions.
func buildExistsAPIEntry(
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) kbapi.SecurityExceptionsAPIExceptionListItemEntryExists {
	return kbapi.SecurityExceptionsAPIExceptionListItemEntryExists{
		Type:     entryTypeExists,
		Field:    field,
		Operator: operator,
	}
}

// convertExistsEntryToAPI converts an exists entry to API format
func convertExistsEntryToAPI(
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	apiEntry := buildExistsAPIEntry(field, operator)
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryExists(apiEntry); err != nil {
		diags.AddError("Failed to create exists entry", err.Error())
	}

	return result, diags
}

// convertNestedExistsEntryToAPI converts a nested exists entry to API format
func convertNestedExistsEntryToAPI(
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem

	apiEntry := buildExistsAPIEntry(field, operator)
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryExists(apiEntry); err != nil {
		diags.AddError("Failed to create nested exists entry", err.Error())
	}

	return result, diags
}

// convertExistsEntryFromAPI converts exists entries from API format
func convertExistsEntryFromAPI(entry *EntryModel) {
	entry.Value = types.StringNull()
	entry.Values = types.ListNull(types.StringType)
	entry.List = types.ObjectNull(getListAttrTypes())
	entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})
}

// convertNestedExistsFromMap converts nested exists entries from map format
func convertNestedExistsFromMap(entry *NestedEntryModel) {
	entry.Value = types.StringNull()
	entry.Values = types.ListNull(types.StringType)
}
