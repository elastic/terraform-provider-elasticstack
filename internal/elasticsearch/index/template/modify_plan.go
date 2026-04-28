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

package template

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state, config Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	merged, diags := reconcilePlanWithPriorStateForSemanticDrift(ctx, plan, state, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if merged == nil {
		return
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, merged)...)
}

// reconcilePlanWithPriorStateForSemanticDrift aligns planned template.settings with prior state when
// Terraform would show a spurious diff: strict inequality but semantic equality (index settings
// canonical form in state vs practitioner JSON in configuration).
func reconcilePlanWithPriorStateForSemanticDrift(ctx context.Context, plan, state, config Model) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics
	if plan.Template.IsNull() || plan.Template.IsUnknown() || state.Template.IsNull() || state.Template.IsUnknown() {
		return nil, diags
	}

	planAttrs := plan.Template.Attributes()
	stateAttrs := state.Template.Attributes()
	changed := false

	if ps, ok := planAttrs["settings"]; ok && !ps.IsNull() && !ps.IsUnknown() {
		if ss, ok := stateAttrs["settings"]; ok && !ss.IsNull() && !ss.IsUnknown() {
			if !ps.Equal(ss) {
				pSet, okP := ps.(customtypes.IndexSettingsValue)
				sSet, okS := ss.(customtypes.IndexSettingsValue)
				if okP && okS {
					eq, d := sSet.SemanticallyEqual(ctx, pSet)
					diags.Append(d...)
					if diags.HasError() {
						return nil, diags
					}
					if eq {
						planAttrs["settings"] = ss
						changed = true
					}
				}
			}
		}
	}

	if pa, ok := planAttrs["alias"]; ok && !pa.IsNull() && !pa.IsUnknown() {
		if sa, ok := stateAttrs["alias"]; ok && !sa.IsNull() && !sa.IsUnknown() {
			newAlias, aliasChanged, d := mergePlanAliasSetWithPriorState(ctx, pa, sa)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			if !aliasChanged && !config.Template.IsNull() && !config.Template.IsUnknown() {
				cfgAttrs := config.Template.Attributes()
				if ca, ok := cfgAttrs["alias"]; ok && !ca.IsNull() && !ca.IsUnknown() {
					// Use config encodings to match state (handles plan unknowns), but project
					// the result back onto the plan's element set so plan-only aliases are
					// preserved. mergePlanAliasSetWithPriorState alone would build the result
					// from its first argument and drop any aliases present in plan but not config.
					newAlias, aliasChanged, d = projectConfigAliasMatchesOntoPlan(ctx, pa, ca, sa)
					diags.Append(d...)
					if diags.HasError() {
						return nil, diags
					}
				}
			}
			if aliasChanged {
				canonAlias, d := canonicalizeAliasSetElements(ctx, newAlias)
				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}
				planAttrs["alias"] = canonAlias
				changed = true
			}
		}
	}

	if !changed {
		return nil, diags
	}

	newTpl, d := types.ObjectValue(TemplateAttrTypes(), planAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	out := plan
	out.Template = newTpl
	return &out, diags
}
