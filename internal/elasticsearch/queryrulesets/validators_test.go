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

package queryrulesets

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCriteriaValuesJSONValidator_acceptsNonEmptyArray(t *testing.T) {
	t.Parallel()

	var resp validator.StringResponse
	criteriaValuesJSONValidator{}.ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("values"),
		ConfigValue: types.StringValue(`["laptop",42]`),
	}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error, got %v", resp.Diagnostics)
	}
}

func TestCriteriaValuesJSONValidator_rejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"null", "42", "{}", `"hello"`, "[]", "not-json"} {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			var resp validator.StringResponse
			criteriaValuesJSONValidator{}.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("values"),
				ConfigValue: types.StringValue(input),
			}, &resp)

			if !resp.Diagnostics.HasError() {
				t.Fatalf("expected validation error for %q", input)
			}
		})
	}
}

func TestQueryRuleActionsValidator_skipsWhenNestedValuesUnknown(t *testing.T) {
	t.Parallel()

	actionsObj, diags := types.ObjectValue(
		queryRuleActionsModelAttrTypes(),
		map[string]attr.Value{
			"ids":  types.ListUnknown(types.StringType),
			"docs": types.ListNull(types.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}),
		},
	)
	if diags.HasError() {
		t.Fatalf("building actions object: %v", diags)
	}

	var resp validator.ObjectResponse
	queryRuleActionsValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        path.Root("actions"),
		ConfigValue: actionsObj,
	}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error when nested values are unknown, got %v", resp.Diagnostics)
	}
}

func TestQueryRuleActionsValidator_requiresExactlyOne(t *testing.T) {
	t.Parallel()

	actionsObj, diags := types.ObjectValue(
		queryRuleActionsModelAttrTypes(),
		map[string]attr.Value{
			"ids":  types.ListNull(types.StringType),
			"docs": types.ListNull(types.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}),
		},
	)
	if diags.HasError() {
		t.Fatalf("building actions object: %v", diags)
	}

	var resp validator.ObjectResponse
	queryRuleActionsValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        path.Root("actions"),
		ConfigValue: actionsObj,
	}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected validation error when neither ids nor docs is set")
	}
}
