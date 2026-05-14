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

// validateConfigCustomRules ensures each configured custom rule satisfies Elasticsearch rules: a rule
// must either have a non-empty scope or at least one condition (when both are known at plan time).
func validateConfigCustomRules(ctx context.Context, config *TFModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if config == nil || config.AnalysisConfig == nil {
		return diags
	}

	ac := config.AnalysisConfig
	for i := range ac.Detectors {
		cr := ac.Detectors[i].CustomRules
		if cr.IsUnknown() || cr.IsNull() {
			continue
		}

		var rules []CustomRuleTFModel
		diags.Append(cr.ElementsAs(ctx, &rules, false)...)
		if diags.HasError() {
			return diags
		}

		for j, rule := range rules {
			if !customRuleMissingScopeAndConditions(rule.Conditions, rule.Scope) {
				continue
			}
			diags.AddAttributeError(
				path.Root("analysis_config").AtName("detectors").AtListIndex(i).AtName("custom_rules").AtListIndex(j),
				`Invalid detector "custom_rules" entry`,
				`A rule must either have a non-empty "scope" or at least one condition. Multiple conditions are combined together with a logical AND.`,
			)
		}
	}

	return diags
}

func customRuleMissingScopeAndConditions(conditions types.List, scope types.Map) bool {
	if conditions.IsUnknown() || scope.IsUnknown() {
		return false
	}
	condEmpty := conditions.IsNull() || len(conditions.Elements()) == 0
	scopeEmpty := scope.IsNull() || len(scope.Elements()) == 0
	return condEmpty && scopeEmpty
}
