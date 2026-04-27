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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

	patchedPlan, patchCh, d := patchPlanAliasFromConfig(ctx, plan, config)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan = patchedPlan

	merged, diags := reconcilePlanWithPriorStateForSemanticDrift(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if merged == nil && !patchCh {
		return
	}
	out := plan
	if merged != nil {
		out = *merged
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &out)...)
}

// reconcilePlanWithPriorStateForSemanticDrift aligns the planned template block with prior state
// when Terraform would show a spurious diff: strict inequality but semantic equality (alias routing
// echoes, index settings canonical form in state vs user JSON in config).
func reconcilePlanWithPriorStateForSemanticDrift(ctx context.Context, plan, state Model) (*Model, diag.Diagnostics) {
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

	pa, okP := planAttrs["alias"]
	sa, okS := stateAttrs["alias"]
	if okP && okS && !pa.IsNull() && !pa.IsUnknown() && !sa.IsNull() && !sa.IsUnknown() {
		if !pa.Equal(sa) {
			pSet, ok1 := pa.(types.Set)
			sSet, ok2 := sa.(types.Set)
			if ok1 && ok2 {
				eq, d := aliasSetsSemanticallyEqual(ctx, pSet, sSet)
				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}
				if eq {
					planAttrs["alias"] = sa
					changed = true
				}
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

func aliasSetsSemanticallyEqual(ctx context.Context, planSet, stateSet basetypes.SetValue) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	pe := planSet.Elements()
	se := stateSet.Elements()
	if len(pe) != len(se) {
		return false, diags
	}

	stateByName := make(map[string]attr.Value, len(se))
	for _, el := range se {
		sAv, ok, d := coerceAliasObjectValue(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}
		if !ok {
			return false, diags
		}
		var sm AliasElementModel
		diags.Append(sAv.As(ctx, &sm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return false, diags
		}
		stateByName[sm.Name.ValueString()] = el
	}

	for _, el := range pe {
		pAv, ok, d := coerceAliasObjectValue(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}
		if !ok {
			return false, diags
		}
		var pm AliasElementModel
		diags.Append(pAv.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return false, diags
		}
		sEl, found := stateByName[pm.Name.ValueString()]
		if !found {
			return false, diags
		}
		sAv, ok2, d := coerceAliasObjectValue(ctx, sEl)
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}
		if !ok2 {
			return false, diags
		}
		eq1, d := aliasObjectValuesSemanticallyEqual(ctx, pAv, sAv)
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}
		eq2 := false
		if !eq1 {
			eq2, d = aliasObjectValuesSemanticallyEqual(ctx, sAv, pAv)
			diags.Append(d...)
			if diags.HasError() {
				return false, diags
			}
		}
		if !eq1 && !eq2 {
			return false, diags
		}
	}
	return true, diags
}
