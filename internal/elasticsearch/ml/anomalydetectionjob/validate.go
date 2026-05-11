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

package anomalydetectionjob

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// validateConfigCustomRules ensures each configured custom rule has a non-empty scope map or at least
// one condition when both attributes are known (Elasticsearch rejects rules with neither).
func validateConfigCustomRules(ctx context.Context, config *TFModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if config == nil || config.AnalysisConfig == nil {
		return diags
	}

	ac := config.AnalysisConfig
	for i := range ac.Detectors {
		det := &ac.Detectors[i]
		cr := det.CustomRules
		if cr.IsUnknown() || cr.IsNull() {
			continue
		}

		var rules []CustomRuleTFModel
		diags.Append(cr.ElementsAs(ctx, &rules, false)...)
		if diags.HasError() {
			return diags
		}

		for j, rule := range rules {
			if !ruleViolatesScopeOrConditionsRequirement(rule.Conditions, rule.Scope) {
				continue
			}
			diags.AddAttributeError(
				path.Root("analysis_config").AtName("detectors").AtListIndex(i).AtName("custom_rules").AtListIndex(j),
				`Invalid detector "custom_rules" entry`,
				`Each rule must set either a non-empty "scope" map or at least one "conditions" block (Elasticsearch requirement).`,
			)
		}
	}

	return diags
}

func ruleViolatesScopeOrConditionsRequirement(conditions types.List, scope types.Map) bool {
	condEmpty, condKnown := listKnownEmpty(conditions)
	scopeEmpty, scopeKnown := mapKnownEmpty(scope)
	return condKnown && scopeKnown && condEmpty && scopeEmpty
}

func listKnownEmpty(l types.List) (empty, known bool) {
	if l.IsUnknown() {
		return false, false
	}
	if l.IsNull() {
		return true, true
	}
	return len(l.Elements()) == 0, true
}

func mapKnownEmpty(m types.Map) (empty, known bool) {
	if m.IsUnknown() {
		return false, false
	}
	if m.IsNull() {
		return true, true
	}
	return len(m.Elements()) == 0, true
}
