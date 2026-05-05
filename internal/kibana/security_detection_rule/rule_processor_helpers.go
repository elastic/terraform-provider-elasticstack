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

import "github.com/hashicorp/terraform-plugin-framework/diag"

// handlesAPIRuleResponse is a generic helper for the HandlesAPIRuleResponse method
// shared by all rule processors. It returns true when rule can be type-asserted to T.
func handlesAPIRuleResponse[T any](rule any) bool {
	_, ok := rule.(T)
	return ok
}

// castRuleResponse type-asserts response to T, adding a diagnostic error on failure.
func castRuleResponse[T any](response any) (T, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, ok := response.(T)
	if !ok {
		var zero T
		diags.AddError(
			"Error extracting rule ID",
			"Could not extract rule ID from response",
		)
		return zero, diags
	}
	return value, diags
}

// updateFromRuleResponse is a generic helper for the UpdateFromResponse method shared
// by all rule processors. It type-asserts response to T and delegates to updateFn.
func updateFromRuleResponse[T any](response any, updateFn func(*T) diag.Diagnostics) diag.Diagnostics {
	value, diags := castRuleResponse[T](response)
	if diags.HasError() {
		return diags
	}
	return updateFn(&value)
}

// extractRuleID is a generic helper for the ExtractID method shared by all rule
// processors. It type-asserts response to T and calls idFn to obtain the string ID.
func extractRuleID[T any](response any, idFn func(T) string) (string, diag.Diagnostics) {
	value, diags := castRuleResponse[T](response)
	if diags.HasError() {
		return "", diags
	}
	return idFn(value), diags
}
