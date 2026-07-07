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

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/queryruletype"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func TestOrderAPIRules_importSortsByRuleID(t *testing.T) {
	t.Parallel()

	apiRules := []types.QueryRule{
		{RuleId: "b", Type: queryruletype.Pinned},
		{RuleId: "a", Type: queryruletype.Pinned},
		{RuleId: "c", Type: queryruletype.Pinned},
	}

	ordered := orderAPIRules(apiRules, nil)
	if len(ordered) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(ordered))
	}
	if ordered[0].RuleId != "a" || ordered[1].RuleId != "b" || ordered[2].RuleId != "c" {
		t.Fatalf("unexpected order: %q, %q, %q", ordered[0].RuleId, ordered[1].RuleId, ordered[2].RuleId)
	}
}

func TestOrderAPIRules_matchesPriorOrderAndAppendsExtras(t *testing.T) {
	t.Parallel()

	apiRules := []types.QueryRule{
		{RuleId: "a", Type: queryruletype.Pinned},
		{RuleId: "c", Type: queryruletype.Pinned},
		{RuleId: "b", Type: queryruletype.Pinned},
	}

	ordered := orderAPIRules(apiRules, []string{"c", "a"})
	if len(ordered) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(ordered))
	}
	if ordered[0].RuleId != "c" || ordered[1].RuleId != "a" || ordered[2].RuleId != "b" {
		t.Fatalf("unexpected order: %q, %q, %q", ordered[0].RuleId, ordered[1].RuleId, ordered[2].RuleId)
	}
}

func TestOrderRulesFromRead_usesPriorStateOrder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ids, idsDiags := fwtypes.ListValueFrom(ctx, fwtypes.StringType, []string{"doc-1"})
	if idsDiags.HasError() {
		t.Fatalf("building ids list: %v", idsDiags)
	}

	actionsObj, objDiags := fwtypes.ObjectValueFrom(ctx, queryRuleActionsModelAttrTypes(), QueryRuleActionsModel{
		IDs:  ids,
		Docs: fwtypes.ListNull(fwtypes.ObjectType{AttrTypes: queryRuleActionDocModelAttrTypes()}),
	})
	if objDiags.HasError() {
		t.Fatalf("building actions object: %v", objDiags)
	}

	criteriaList, listDiags := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: queryRuleCriteriaModelAttrTypes()}, []QueryRuleCriteriaModel{
		{
			Type:     fwtypes.StringValue("always"),
			Metadata: fwtypes.StringNull(),
			Values:   jsontypes.Normalized{StringValue: fwtypes.StringNull()},
		},
	})
	if listDiags.HasError() {
		t.Fatalf("building criteria list: %v", listDiags)
	}

	priorRules, diags := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: queryRuleModelAttrTypes()}, []QueryRuleModel{
		{
			RuleID:   fwtypes.StringValue("second"),
			Type:     fwtypes.StringValue("pinned"),
			Priority: fwtypes.Int64Null(),
			Criteria: criteriaList,
			Actions:  actionsObj,
		},
		{
			RuleID:   fwtypes.StringValue("first"),
			Type:     fwtypes.StringValue("pinned"),
			Priority: fwtypes.Int64Null(),
			Criteria: criteriaList,
			Actions:  actionsObj,
		},
	})
	if diags.HasError() {
		t.Fatalf("building prior rules: %v", diags)
	}

	apiRules := []types.QueryRule{
		{RuleId: "first", Type: queryruletype.Pinned},
		{RuleId: "third", Type: queryruletype.Pinned},
		{RuleId: "second", Type: queryruletype.Pinned},
	}

	ordered, orderDiags := orderRulesFromRead(ctx, apiRules, priorRules)
	if orderDiags.HasError() {
		t.Fatalf("orderRulesFromRead: %v", orderDiags)
	}
	if ordered[0].RuleId != "second" || ordered[1].RuleId != "first" || ordered[2].RuleId != "third" {
		t.Fatalf("unexpected order: %q, %q, %q", ordered[0].RuleId, ordered[1].RuleId, ordered[2].RuleId)
	}
}
