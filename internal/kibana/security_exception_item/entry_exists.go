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
)

// convertExistsEntryToAPI converts an exists entry to API format
func convertExistsEntryToAPI(
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry
	diags := fillExistsEntry(&result, field, operator, "Failed to create exists entry")
	return result, diags
}

// convertNestedExistsEntryToAPI converts a nested exists entry to API format
func convertNestedExistsEntryToAPI(
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
) (kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, diag.Diagnostics) {
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem
	diags := fillExistsEntry(&result, field, operator, "Failed to create nested exists entry")
	return result, diags
}
