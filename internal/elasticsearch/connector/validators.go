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

package connector

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const configurationValueBranchErrorSummary = "Invalid configuration value"

type configurationValueBranchValidator struct{}

func (configurationValueBranchValidator) Description(_ context.Context) string {
	return "Exactly one of string, number, bool, json, or secret_value must be set"
}

func (v configurationValueBranchValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v configurationValueBranchValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	if configurationValueAllBranchesUnknown(attrs) {
		return
	}

	setCount := 0
	var setBranches []string

	for _, name := range configurationValueBranchAttrNames {
		val, ok := attrs[name]
		if !ok {
			continue
		}
		if configurationValueBranchIsSet(val) {
			setCount++
			setBranches = append(setBranches, name)
		}
	}

	switch {
	case setCount == 0:
		resp.Diagnostics.AddAttributeError(
			req.Path,
			configurationValueBranchErrorSummary,
			"Exactly one of string, number, bool, json, or secret_value must be set.",
		)
	case setCount > 1:
		resp.Diagnostics.AddAttributeError(
			req.Path,
			configurationValueBranchErrorSummary,
			fmt.Sprintf("Exactly one of string, number, bool, json, or secret_value must be set; found %d set (%v).", setCount, setBranches),
		)
	}
}

func configurationValueBranchIsSet(val attr.Value) bool {
	if val == nil || val.IsNull() || val.IsUnknown() {
		return false
	}

	switch v := val.(type) {
	case types.String:
		return true
	case types.Number:
		return true
	case types.Bool:
		return true
	case jsontypes.Normalized:
		return !v.IsNull() && !v.IsUnknown()
	default:
		// Unexpected attribute types are treated as unset to avoid hiding schema drift.
		return false
	}
}

func configurationValueAllBranchesUnknown(attrs map[string]attr.Value) bool {
	sawBranch := false
	for _, name := range configurationValueBranchAttrNames {
		val, ok := attrs[name]
		if !ok {
			continue
		}
		sawBranch = true
		if !val.IsUnknown() {
			return false
		}
	}
	return sawBranch
}

// Ensure validators satisfy interfaces at compile time.
var _ validator.Object = configurationValueBranchValidator{}
