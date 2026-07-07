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
	"sort"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func orderRulesFromRead(ctx context.Context, apiRules []types.QueryRule, priorRules fwtypes.List) ([]types.QueryRule, diag.Diagnostics) {
	priorOrder, diags := priorRuleOrder(ctx, priorRules)
	if diags.HasError() {
		return nil, diags
	}

	return orderAPIRules(apiRules, priorOrder), nil
}

func priorRuleOrder(ctx context.Context, priorRules fwtypes.List) ([]string, diag.Diagnostics) {
	if priorRules.IsNull() || priorRules.IsUnknown() || len(priorRules.Elements()) == 0 {
		return nil, nil
	}

	var models []QueryRuleModel
	var diags diag.Diagnostics
	diags.Append(priorRules.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}

	order := make([]string, len(models))
	for i, model := range models {
		order[i] = model.RuleID.ValueString()
	}

	return order, nil
}

func orderAPIRules(apiRules []types.QueryRule, priorOrder []string) []types.QueryRule {
	if len(priorOrder) == 0 {
		ordered := append([]types.QueryRule(nil), apiRules...)
		sort.SliceStable(ordered, func(i, j int) bool {
			return ordered[i].RuleId < ordered[j].RuleId
		})
		return ordered
	}

	byID := make(map[string]types.QueryRule, len(apiRules))
	for _, rule := range apiRules {
		byID[rule.RuleId] = rule
	}

	ordered := make([]types.QueryRule, 0, len(apiRules))
	seen := make(map[string]struct{}, len(apiRules))
	for _, ruleID := range priorOrder {
		rule, ok := byID[ruleID]
		if !ok {
			continue
		}
		ordered = append(ordered, rule)
		seen[ruleID] = struct{}{}
	}

	for _, rule := range apiRules {
		if _, ok := seen[rule.RuleId]; ok {
			continue
		}
		ordered = append(ordered, rule)
	}

	return ordered
}
