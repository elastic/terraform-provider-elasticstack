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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPreserveRuleCriteriaValuesFromPrior_matchesByTypeAndMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := testRuleWithCriteria(ctx, t,
		criterion("exact", "query", `[100]`),
	)
	current := testRuleWithCriteria(ctx, t,
		criterion("exact", "query", `["100"]`),
	)

	preserveRuleCriteriaValuesFromPrior(ctx, &current, prior, &diag.Diagnostics{})

	assertCriterionValues(t, current, 0, `[100]`)
}

func TestPreserveRuleCriteriaValuesFromPrior_differentTypeAtSameIndex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := testRuleWithCriteria(ctx, t,
		criterion("exact", "query", `[100]`),
	)
	current := testRuleWithCriteria(ctx, t,
		criterion("contains", "query", `["100"]`),
	)

	preserveRuleCriteriaValuesFromPrior(ctx, &current, prior, &diag.Diagnostics{})

	assertCriterionValues(t, current, 0, `["100"]`)
}

func TestPreserveRuleCriteriaValuesFromPrior_reorderedCriteria(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := testRuleWithCriteria(ctx, t,
		criterion("exact", "query", `["alpha"]`),
		criterion("contains", "tags", `["beta"]`),
	)
	current := testRuleWithCriteria(ctx, t,
		criterion("contains", "tags", `["beta"]`),
		criterion("exact", "query", `["alpha"]`),
	)

	preserveRuleCriteriaValuesFromPrior(ctx, &current, prior, &diag.Diagnostics{})

	assertCriterionValues(t, current, 0, `["beta"]`)
	assertCriterionValues(t, current, 1, `["alpha"]`)
}

func TestPreserveRuleCriteriaValuesFromPrior_removedCriterion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := testRuleWithCriteria(ctx, t,
		criterion("exact", "query", `["alpha"]`),
		criterion("contains", "tags", `["beta"]`),
	)
	current := testRuleWithCriteria(ctx, t,
		criterion("exact", "query", `["alpha"]`),
	)

	preserveRuleCriteriaValuesFromPrior(ctx, &current, prior, &diag.Diagnostics{})

	assertCriterionValues(t, current, 0, `["alpha"]`)
}

func TestPreserveRuleCriteriaValuesFromPrior_newCriterionWithoutPriorMatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := testRuleWithCriteria(ctx, t,
		criterion("exact", "query", `["alpha"]`),
	)
	current := testRuleWithCriteria(ctx, t,
		criterion("exact", "query", `["alpha"]`),
		criterion("contains", "tags", `["beta"]`),
	)

	preserveRuleCriteriaValuesFromPrior(ctx, &current, prior, &diag.Diagnostics{})

	assertCriterionValues(t, current, 0, `["alpha"]`)
	assertCriterionValues(t, current, 1, `["beta"]`)
}

func TestPreserveRuleCriteriaValuesFromPrior_duplicateTypeMetadataKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := testRuleWithCriteria(ctx, t,
		criterion("exact", "query", `["first"]`),
		criterion("exact", "query", `["second"]`),
	)
	current := testRuleWithCriteria(ctx, t,
		criterion("exact", "query", `["first"]`),
		criterion("exact", "query", `["second"]`),
	)

	preserveRuleCriteriaValuesFromPrior(ctx, &current, prior, &diag.Diagnostics{})

	assertCriterionValues(t, current, 0, `["first"]`)
	assertCriterionValues(t, current, 1, `["second"]`)
}

func TestPreserveCriteriaValuesFromPrior_numericCreateRead(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	priorRules, diags := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: queryRuleModelAttrTypes()}, []QueryRuleModel{
		testRuleModel(ctx, t, criterion("gt", "popularity", `[100]`)),
	})
	if diags.HasError() {
		t.Fatalf("building prior rules: %v", diags)
	}

	data := QueryRulesetData{
		Rules: testRulesList(ctx, t, testRuleModel(ctx, t, criterion("gt", "popularity", `["100"]`))),
	}

	preserveCriteriaValuesFromPrior(ctx, &data, priorRules, &diag.Diagnostics{})

	assertCriterionValuesOnRule(t, data, 0, 0, `[100]`)
}

func TestPreserveCriteriaValuesFromPrior_emptyPriorRules(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := QueryRulesetData{
		Rules: testRulesList(ctx, t, testRuleModel(ctx, t, criterion("gt", "popularity", `["100"]`))),
	}

	preserveCriteriaValuesFromPrior(ctx, &data, fwtypes.ListNull(fwtypes.ObjectType{AttrTypes: queryRuleModelAttrTypes()}), &diag.Diagnostics{})

	assertCriterionValuesOnRule(t, data, 0, 0, `["100"]`)
}

func criterion(typ, metadata, values string) QueryRuleCriteriaModel {
	model := QueryRuleCriteriaModel{
		Type: fwtypes.StringValue(typ),
	}
	if metadata != "" {
		model.Metadata = fwtypes.StringValue(metadata)
	} else {
		model.Metadata = fwtypes.StringNull()
	}
	if values != "" {
		model.Values = jsontypes.Normalized{StringValue: fwtypes.StringValue(values)}
	} else {
		model.Values = jsontypes.Normalized{StringValue: fwtypes.StringNull()}
	}
	return model
}

func testRuleWithCriteria(ctx context.Context, t *testing.T, criteria ...QueryRuleCriteriaModel) QueryRuleModel {
	t.Helper()
	return testRuleModel(ctx, t, criteria...)
}

func testRuleModel(ctx context.Context, t *testing.T, criteria ...QueryRuleCriteriaModel) QueryRuleModel {
	t.Helper()

	criteriaList, diags := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: queryRuleCriteriaModelAttrTypes()}, criteria)
	if diags.HasError() {
		t.Fatalf("building criteria list: %v", diags)
	}

	ids, diags := fwtypes.ListValueFrom(ctx, fwtypes.StringType, []string{"doc-1"})
	if diags.HasError() {
		t.Fatalf("building ids list: %v", diags)
	}

	actionsObj, diags := fwtypes.ObjectValueFrom(ctx, queryRuleActionsModelAttrTypes(), QueryRuleActionsModel{
		IDs:  ids,
		Docs: fwtypes.ListNull(fwtypes.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}),
	})
	if diags.HasError() {
		t.Fatalf("building actions object: %v", diags)
	}

	return QueryRuleModel{
		RuleID:   fwtypes.StringValue("rule-1"),
		Type:     fwtypes.StringValue("pinned"),
		Priority: fwtypes.Int64Null(),
		Criteria: criteriaList,
		Actions:  actionsObj,
	}
}

func testRulesList(ctx context.Context, t *testing.T, rules ...QueryRuleModel) fwtypes.List {
	t.Helper()

	list, diags := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: queryRuleModelAttrTypes()}, rules)
	if diags.HasError() {
		t.Fatalf("building rules list: %v", diags)
	}
	return list
}

func assertCriterionValues(t *testing.T, rule QueryRuleModel, index int, want string) {
	t.Helper()

	ctx := context.Background()
	var criteria []QueryRuleCriteriaModel
	if err := rule.Criteria.ElementsAs(ctx, &criteria, false); err != nil {
		t.Fatalf("reading criteria: %v", err)
	}
	if index >= len(criteria) {
		t.Fatalf("criterion index %d out of range (%d)", index, len(criteria))
	}
	got := criteria[index].Values.ValueString()
	if got != want {
		t.Fatalf("criteria[%d].values = %q, want %q", index, got, want)
	}
}

func assertCriterionValuesOnRule(t *testing.T, data QueryRulesetData, ruleIndex, criterionIndex int, want string) {
	t.Helper()

	ctx := context.Background()
	var rules []QueryRuleModel
	if err := data.Rules.ElementsAs(ctx, &rules, false); err != nil {
		t.Fatalf("reading rules: %v", err)
	}
	if ruleIndex >= len(rules) {
		t.Fatalf("rule index %d out of range (%d)", ruleIndex, len(rules))
	}
	assertCriterionValues(t, rules[ruleIndex], criterionIndex, want)
}
