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
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func TestQueryRuleActionsValidator_acceptsOnlyIDs(t *testing.T) {
	t.Parallel()

	ids, diags := types.ListValueFrom(context.Background(), types.StringType, []string{"doc-1"})
	if diags.HasError() {
		t.Fatalf("building ids list: %v", diags)
	}

	actionsObj, objDiags := types.ObjectValue(
		queryRuleActionsModelAttrTypes(),
		map[string]attr.Value{
			"ids":  ids,
			"docs": types.ListNull(types.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}),
		},
	)
	if objDiags.HasError() {
		t.Fatalf("building actions object: %v", objDiags)
	}

	var resp validator.ObjectResponse
	queryRuleActionsValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        path.Root("actions"),
		ConfigValue: actionsObj,
	}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error when only ids is set, got %v", resp.Diagnostics)
	}
}

func TestQueryRuleActionsValidator_acceptsOnlyDocs(t *testing.T) {
	t.Parallel()

	docObj, docDiags := types.ObjectValue(
		queryRuleActionDocModelAttrTypes(),
		map[string]attr.Value{
			"_index": types.StringValue("my-index"),
			"_id":    types.StringValue("42"),
		},
	)
	if docDiags.HasError() {
		t.Fatalf("building doc object: %v", docDiags)
	}

	docs, listDiags := types.ListValue(types.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}, []attr.Value{docObj})
	if listDiags.HasError() {
		t.Fatalf("building docs list: %v", listDiags)
	}

	actionsObj, objDiags := types.ObjectValue(
		queryRuleActionsModelAttrTypes(),
		map[string]attr.Value{
			"ids":  types.ListNull(types.StringType),
			"docs": docs,
		},
	)
	if objDiags.HasError() {
		t.Fatalf("building actions object: %v", objDiags)
	}

	var resp validator.ObjectResponse
	queryRuleActionsValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        path.Root("actions"),
		ConfigValue: actionsObj,
	}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error when only docs is set, got %v", resp.Diagnostics)
	}
}

func TestQueryRuleActionsValidator_rejectsBothIDsAndDocs(t *testing.T) {
	t.Parallel()

	ids, diags := types.ListValueFrom(context.Background(), types.StringType, []string{"doc-1"})
	if diags.HasError() {
		t.Fatalf("building ids list: %v", diags)
	}

	docObj, docDiags := types.ObjectValue(
		queryRuleActionDocModelAttrTypes(),
		map[string]attr.Value{
			"_index": types.StringValue("my-index"),
			"_id":    types.StringValue("42"),
		},
	)
	if docDiags.HasError() {
		t.Fatalf("building doc object: %v", docDiags)
	}

	docs, listDiags := types.ListValue(types.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}, []attr.Value{docObj})
	if listDiags.HasError() {
		t.Fatalf("building docs list: %v", listDiags)
	}

	actionsObj, objDiags := types.ObjectValue(
		queryRuleActionsModelAttrTypes(),
		map[string]attr.Value{
			"ids":  ids,
			"docs": docs,
		},
	)
	if objDiags.HasError() {
		t.Fatalf("building actions object: %v", objDiags)
	}

	var resp validator.ObjectResponse
	queryRuleActionsValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        path.Root("actions"),
		ConfigValue: actionsObj,
	}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected validation error when both ids and docs are set")
	}
}

func criteriaObject(t *testing.T, attrs map[string]attr.Value) types.Object {
	t.Helper()

	obj, diags := types.ObjectValue(queryRuleCriteriaModelAttrTypes(), attrs)
	if diags.HasError() {
		t.Fatalf("building criteria object: %v", diags)
	}
	return obj
}

func criteriaValuesNull() attr.Value {
	return jsontypes.Normalized{StringValue: types.StringNull()}
}

func criteriaValuesUnknown() attr.Value {
	return jsontypes.Normalized{StringValue: types.StringUnknown()}
}

func criteriaValuesString(value string) attr.Value {
	return jsontypes.Normalized{StringValue: types.StringValue(value)}
}

func TestQueryRuleCriteriaValidator_acceptsAlwaysWithoutValues(t *testing.T) {
	t.Parallel()

	var resp validator.ObjectResponse
	queryRuleCriteriaValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path: path.Root("criteria"),
		ConfigValue: criteriaObject(t, map[string]attr.Value{
			"type":     types.StringValue("always"),
			"metadata": types.StringNull(),
			"values":   criteriaValuesNull(),
		}),
	}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error, got %v", resp.Diagnostics)
	}
}

func TestQueryRuleCriteriaValidator_rejectsAlwaysWithValues(t *testing.T) {
	t.Parallel()

	var resp validator.ObjectResponse
	queryRuleCriteriaValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path: path.Root("criteria"),
		ConfigValue: criteriaObject(t, map[string]attr.Value{
			"type":     types.StringValue("always"),
			"metadata": types.StringNull(),
			"values":   criteriaValuesString(`["x"]`),
		}),
	}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected validation error when always has values")
	}
	if !containsDiagnosticSummary(resp.Diagnostics, "always") {
		t.Fatalf("expected diagnostic mentioning always, got %v", resp.Diagnostics)
	}
}

func TestQueryRuleCriteriaValidator_requiresValuesForNonAlways(t *testing.T) {
	t.Parallel()

	var resp validator.ObjectResponse
	queryRuleCriteriaValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path: path.Root("criteria"),
		ConfigValue: criteriaObject(t, map[string]attr.Value{
			"type":     types.StringValue("exact"),
			"metadata": types.StringValue("query"),
			"values":   criteriaValuesNull(),
		}),
	}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected validation error when values is missing")
	}
	if !containsDiagnosticSummary(resp.Diagnostics, "values") {
		t.Fatalf("expected diagnostic mentioning values, got %v", resp.Diagnostics)
	}
}

func TestQueryRuleCriteriaValidator_acceptsNonAlwaysWithValues(t *testing.T) {
	t.Parallel()

	var resp validator.ObjectResponse
	queryRuleCriteriaValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path: path.Root("criteria"),
		ConfigValue: criteriaObject(t, map[string]attr.Value{
			"type":     types.StringValue("exact"),
			"metadata": types.StringValue("query"),
			"values":   criteriaValuesString(`["x"]`),
		}),
	}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error, got %v", resp.Diagnostics)
	}
}

func TestQueryRuleCriteriaValidator_skipsWhenTypeUnknown(t *testing.T) {
	t.Parallel()

	var resp validator.ObjectResponse
	queryRuleCriteriaValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path: path.Root("criteria"),
		ConfigValue: criteriaObject(t, map[string]attr.Value{
			"type":     types.StringUnknown(),
			"metadata": types.StringValue("query"),
			"values":   criteriaValuesNull(),
		}),
	}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error when type is unknown, got %v", resp.Diagnostics)
	}
}

func TestQueryRuleCriteriaValidator_skipsWhenValuesUnknown(t *testing.T) {
	t.Parallel()

	var resp validator.ObjectResponse
	queryRuleCriteriaValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path: path.Root("criteria"),
		ConfigValue: criteriaObject(t, map[string]attr.Value{
			"type":     types.StringValue("exact"),
			"metadata": types.StringValue("query"),
			"values":   criteriaValuesUnknown(),
		}),
	}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error when values is unknown, got %v", resp.Diagnostics)
	}
}

func containsDiagnosticSummary(diags diag.Diagnostics, substr string) bool {
	for _, d := range diags {
		if strings.Contains(d.Summary(), substr) || strings.Contains(d.Detail(), substr) {
			return true
		}
	}
	return false
}
