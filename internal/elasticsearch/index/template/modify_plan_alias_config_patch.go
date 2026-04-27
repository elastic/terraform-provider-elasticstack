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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// patchPlanAliasFromConfig restores template.alias fields on the plan from practitioner config when
// Terraform's PlanResourceChange set correlation re-applies schema defaults (e.g. routing "", or
// is_write_index / is_hidden false) and drops explicit configuration.
func patchPlanAliasFromConfig(ctx context.Context, plan, config Model) (Model, bool, diag.Diagnostics) {
	plan, ch1, d1 := patchPlanAliasRoutingFromConfig(ctx, plan, config)
	if d1.HasError() {
		return plan, ch1, d1
	}
	plan, ch2, d2 := patchPlanAliasOptionalBoolsFromConfig(ctx, plan, config)
	d1.Append(d2...)
	return plan, ch1 || ch2, d1
}

// patchPlanAliasRoutingFromConfig restores template.alias[].routing on the plan from practitioner
// config when Terraform's PlanResourceChange correlation re-applied the schema default ("") and
// cleared an explicitly configured non-empty routing (observed in TF_LOG: "setting ... routing to default value").
func patchPlanAliasRoutingFromConfig(ctx context.Context, plan, config Model) (Model, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	changed := false
	if plan.Template.IsNull() || plan.Template.IsUnknown() || config.Template.IsNull() || config.Template.IsUnknown() {
		return plan, changed, diags
	}
	pAttrs := plan.Template.Attributes()
	cAttrs := config.Template.Attributes()
	pAlias, okP := pAttrs["alias"]
	cAlias, okC := cAttrs["alias"]
	if !okP || !okC || pAlias.IsNull() || pAlias.IsUnknown() || cAlias.IsNull() || cAlias.IsUnknown() {
		return plan, changed, diags
	}
	pSet, ok1 := pAlias.(types.Set)
	cSet, ok2 := cAlias.(types.Set)
	if !ok1 || !ok2 {
		return plan, changed, diags
	}

	cfgRoutingByName := map[string]attr.Value{}
	for _, el := range cSet.Elements() {
		cAv, ok, d := coerceAliasObjectValue(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return plan, false, diags
		}
		if !ok {
			continue
		}
		var cm AliasElementModel
		diags.Append(cAv.As(ctx, &cm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return plan, false, diags
		}
		name := cm.Name.ValueString()
		if cm.Routing.IsNull() || cm.Routing.IsUnknown() || cm.Routing.ValueString() == "" {
			continue
		}
		cfgRoutingByName[name] = cAv.Attributes()["routing"]
	}
	if len(cfgRoutingByName) == 0 {
		return plan, changed, diags
	}

	newElems := make([]attr.Value, 0, len(pSet.Elements()))
	for _, el := range pSet.Elements() {
		pAv, ok, d := coerceAliasObjectValue(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return plan, false, diags
		}
		if !ok {
			newElems = append(newElems, el)
			continue
		}
		var pm AliasElementModel
		diags.Append(pAv.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return plan, false, diags
		}
		wantR, ok := cfgRoutingByName[pm.Name.ValueString()]
		if !ok {
			newElems = append(newElems, el)
			continue
		}
		if !pm.Routing.IsNull() && !pm.Routing.IsUnknown() && pm.Routing.ValueString() != "" {
			newElems = append(newElems, el)
			continue
		}
		attrs := pAv.Attributes()
		attrs["routing"] = wantR
		fixed, d := NewAliasObjectValue(attrs)
		diags.Append(d...)
		if diags.HasError() {
			return plan, false, diags
		}
		newElems = append(newElems, fixed)
		changed = true
	}
	if !changed {
		return plan, changed, diags
	}
	newSet, d := types.SetValueFrom(ctx, NewAliasObjectType(), newElems)
	diags.Append(d...)
	if diags.HasError() {
		return plan, false, diags
	}
	pAttrs = plan.Template.Attributes()
	pAttrs["alias"] = newSet
	tplObj, d := types.ObjectValue(TemplateAttrTypes(), pAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return plan, false, diags
	}
	plan.Template = tplObj
	return plan, true, diags
}

func patchPlanAliasOptionalBoolsFromConfig(ctx context.Context, plan, config Model) (Model, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	changed := false
	if plan.Template.IsNull() || plan.Template.IsUnknown() || config.Template.IsNull() || config.Template.IsUnknown() {
		return plan, changed, diags
	}
	pAttrs := plan.Template.Attributes()
	cAttrs := config.Template.Attributes()
	pAlias, okP := pAttrs["alias"]
	cAlias, okC := cAttrs["alias"]
	if !okP || !okC || pAlias.IsNull() || pAlias.IsUnknown() || cAlias.IsNull() || cAlias.IsUnknown() {
		return plan, changed, diags
	}
	pSet, ok1 := pAlias.(types.Set)
	cSet, ok2 := cAlias.(types.Set)
	if !ok1 || !ok2 {
		return plan, changed, diags
	}

	cfgByName := make(map[string]AliasObjectValue)
	for _, el := range cSet.Elements() {
		cAv, ok, d := coerceAliasObjectValue(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return plan, false, diags
		}
		if !ok {
			continue
		}
		var cm AliasElementModel
		diags.Append(cAv.As(ctx, &cm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return plan, false, diags
		}
		cfgByName[cm.Name.ValueString()] = cAv
	}

	newElems := make([]attr.Value, 0, len(pSet.Elements()))
	for _, el := range pSet.Elements() {
		pAv, ok, d := coerceAliasObjectValue(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return plan, false, diags
		}
		if !ok {
			newElems = append(newElems, el)
			continue
		}
		var pm AliasElementModel
		diags.Append(pAv.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return plan, false, diags
		}
		cAv, ok := cfgByName[pm.Name.ValueString()]
		if !ok {
			newElems = append(newElems, el)
			continue
		}
		var cm AliasElementModel
		diags.Append(cAv.As(ctx, &cm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return plan, false, diags
		}

		attrs := pAv.Attributes()
		elemChanged := false
		if !cm.IsHidden.IsNull() && !cm.IsHidden.IsUnknown() && !pm.IsHidden.Equal(cm.IsHidden) {
			attrs["is_hidden"] = cm.IsHidden
			elemChanged = true
		}
		if !cm.IsWriteIndex.IsNull() && !cm.IsWriteIndex.IsUnknown() && !pm.IsWriteIndex.Equal(cm.IsWriteIndex) {
			attrs["is_write_index"] = cm.IsWriteIndex
			elemChanged = true
		}
		if !elemChanged {
			newElems = append(newElems, el)
			continue
		}
		fixed, d := NewAliasObjectValue(attrs)
		diags.Append(d...)
		if diags.HasError() {
			return plan, false, diags
		}
		newElems = append(newElems, fixed)
		changed = true
	}
	if !changed {
		return plan, changed, diags
	}
	newSet, d := types.SetValueFrom(ctx, NewAliasObjectType(), newElems)
	diags.Append(d...)
	if diags.HasError() {
		return plan, false, diags
	}
	pAttrs = plan.Template.Attributes()
	pAttrs["alias"] = newSet
	tplObj, d := types.ObjectValue(TemplateAttrTypes(), pAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return plan, false, diags
	}
	plan.Template = tplObj
	return plan, true, diags
}
