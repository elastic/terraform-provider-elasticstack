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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// applyTemplateAliasReconciliationFromReference replaces template.alias elements read from the API
// with the reference encoding when ObjectSemanticEquals matches. Use configuration on Create and
// Update (req.Plan can leave Optional+Computed nested attributes unknown); use prior state on Read.
func applyTemplateAliasReconciliationFromReference(ctx context.Context, out *Model, ref *Model) diag.Diagnostics {
	var diags diag.Diagnostics
	if out.Template.IsNull() || out.Template.IsUnknown() {
		return diags
	}
	if ref.Template.IsNull() || ref.Template.IsUnknown() {
		return diags
	}

	outAttrs := out.Template.Attributes()
	refAttrs := ref.Template.Attributes()
	apiAliasVal := outAttrs["alias"]
	refAliasVal := refAttrs["alias"]
	if apiAliasVal.IsNull() || apiAliasVal.IsUnknown() {
		return diags
	}
	if refAliasVal.IsNull() || refAliasVal.IsUnknown() {
		return diags
	}

	merged, changed, d := mergeAliasSetPreferReferenceEncoding(ctx, apiAliasVal, refAliasVal)
	diags.Append(d...)
	if diags.HasError() || !changed {
		return diags
	}

	outAttrs["alias"] = merged
	newTpl, d := types.ObjectValue(TemplateAttrTypes(), outAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	out.Template = newTpl
	return diags
}

// mergePlanAliasSetWithPriorState walks planned alias elements; when prior state has the same name and
// planElt.ObjectSemanticEquals(stateElt), replace the plan element with the state's encoding so planned
// values match stored state under Optional+Computed+Default nested sets.
func mergePlanAliasSetWithPriorState(ctx context.Context, planAliases, stateAliases attr.Value) (attr.Value, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	pSet, ok := planAliases.(basetypes.SetValue)
	if !ok {
		return planAliases, false, diags
	}
	sSet, ok := stateAliases.(basetypes.SetValue)
	if !ok {
		return planAliases, false, diags
	}
	if pSet.IsNull() || pSet.IsUnknown() || sSet.IsNull() || sSet.IsUnknown() {
		return planAliases, false, diags
	}

	stateByName := make(map[string]AliasObjectValue)
	for _, se := range sSet.Elements() {
		sAlias, sOK, d := aliasObjectFromAttr(ctx, se)
		diags.Append(d...)
		if diags.HasError() {
			return planAliases, false, diags
		}
		if !sOK || sAlias.IsNull() || sAlias.IsUnknown() {
			continue
		}
		var sm AliasElementModel
		diags.Append(sAlias.As(ctx, &sm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return planAliases, false, diags
		}
		stateByName[sm.Name.ValueString()] = sAlias
	}

	planElems := pSet.Elements()
	newElems := make([]attr.Value, len(planElems))
	changed := false
	for i, pe := range planElems {
		pAlias, pOK, d := aliasObjectFromAttr(ctx, pe)
		diags.Append(d...)
		if diags.HasError() {
			return planAliases, false, diags
		}
		if !pOK || pAlias.IsNull() || pAlias.IsUnknown() {
			newElems[i] = pe
			continue
		}
		var pm AliasElementModel
		diags.Append(pAlias.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return planAliases, false, diags
		}
		sAlias, ok := stateByName[pm.Name.ValueString()]
		if !ok {
			newElems[i] = pe
			continue
		}
		eq, d := pAlias.ObjectSemanticEquals(ctx, sAlias)
		diags.Append(d...)
		if diags.HasError() {
			return planAliases, false, diags
		}
		if eq && !pAlias.Equal(sAlias) {
			newElems[i] = sAlias
			changed = true
			continue
		}
		newElems[i] = pe
	}

	if !changed {
		return planAliases, false, diags
	}

	newSet, d := types.SetValue(NewAliasObjectType(), newElems)
	diags.Append(d...)
	if diags.HasError() {
		return planAliases, false, diags
	}
	return newSet, true, diags
}

func mergeAliasSetPreferReferenceEncoding(ctx context.Context, apiSet, refSet attr.Value) (attr.Value, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiS, ok := apiSet.(basetypes.SetValue)
	if !ok {
		return apiSet, false, diags
	}
	refS, ok := refSet.(basetypes.SetValue)
	if !ok {
		return apiSet, false, diags
	}
	if apiS.IsNull() || apiS.IsUnknown() || refS.IsNull() || refS.IsUnknown() {
		return apiSet, false, diags
	}

	refByName := make(map[string]AliasObjectValue)
	for _, re := range refS.Elements() {
		refAlias, refOK, d := aliasObjectFromAttr(ctx, re)
		diags.Append(d...)
		if diags.HasError() {
			return apiSet, false, diags
		}
		if !refOK || refAlias.IsNull() || refAlias.IsUnknown() {
			continue
		}
		var m AliasElementModel
		diags.Append(refAlias.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return apiSet, false, diags
		}
		refByName[m.Name.ValueString()] = refAlias
	}

	apiElems := apiS.Elements()
	newElems := make([]attr.Value, 0, len(apiElems))
	changed := false
	for _, ae := range apiElems {
		apiAlias, apiOK, d := aliasObjectFromAttr(ctx, ae)
		diags.Append(d...)
		if diags.HasError() {
			return apiSet, false, diags
		}
		if !apiOK || apiAlias.IsNull() || apiAlias.IsUnknown() {
			newElems = append(newElems, ae)
			continue
		}
		var am AliasElementModel
		diags.Append(apiAlias.As(ctx, &am, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return apiSet, false, diags
		}
		refAlias, ok := refByName[am.Name.ValueString()]
		if !ok {
			newElems = append(newElems, ae)
			continue
		}
		eq, d := refAlias.ObjectSemanticEquals(ctx, apiAlias)
		diags.Append(d...)
		if diags.HasError() {
			return apiSet, false, diags
		}
		if eq && !refAlias.Equal(apiAlias) {
			newElems = append(newElems, refAlias)
			changed = true
			continue
		}
		newElems = append(newElems, ae)
	}

	if !changed {
		return apiSet, false, diags
	}

	newSet, d := types.SetValue(NewAliasObjectType(), newElems)
	diags.Append(d...)
	if diags.HasError() {
		return apiSet, false, diags
	}
	return newSet, true, diags
}

// projectConfigAliasMatchesOntoPlan walks the plan's alias elements and, for each one, looks up
// the same-named config alias and the same-named state alias. When the config encoding semantically
// equals the state encoding, the plan element is replaced with the state encoding (so the planned
// value matches stored state under Optional+Computed+Default nested sets, even when the plan
// element carries unknowns). Plan elements with no config counterpart are preserved unchanged.
func projectConfigAliasMatchesOntoPlan(ctx context.Context, planAliases, configAliases, stateAliases attr.Value) (attr.Value, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	pSet, ok := planAliases.(basetypes.SetValue)
	if !ok {
		return planAliases, false, diags
	}
	cSet, ok := configAliases.(basetypes.SetValue)
	if !ok {
		return planAliases, false, diags
	}
	sSet, ok := stateAliases.(basetypes.SetValue)
	if !ok {
		return planAliases, false, diags
	}
	if pSet.IsNull() || pSet.IsUnknown() || cSet.IsNull() || cSet.IsUnknown() || sSet.IsNull() || sSet.IsUnknown() {
		return planAliases, false, diags
	}

	configByName, d := aliasObjectsByName(ctx, cSet)
	diags.Append(d...)
	if diags.HasError() {
		return planAliases, false, diags
	}
	stateByName, d := aliasObjectsByName(ctx, sSet)
	diags.Append(d...)
	if diags.HasError() {
		return planAliases, false, diags
	}

	planElems := pSet.Elements()
	newElems := make([]attr.Value, len(planElems))
	changed := false
	for i, pe := range planElems {
		pAlias, pOK, d := aliasObjectFromAttr(ctx, pe)
		diags.Append(d...)
		if diags.HasError() {
			return planAliases, false, diags
		}
		if !pOK || pAlias.IsNull() || pAlias.IsUnknown() {
			newElems[i] = pe
			continue
		}
		var pm AliasElementModel
		diags.Append(pAlias.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return planAliases, false, diags
		}
		name := pm.Name.ValueString()
		cAlias, cFound := configByName[name]
		sAlias, sFound := stateByName[name]
		if !cFound || !sFound {
			newElems[i] = pe
			continue
		}
		eq, d := cAlias.ObjectSemanticEquals(ctx, sAlias)
		diags.Append(d...)
		if diags.HasError() {
			return planAliases, false, diags
		}
		if eq && !pAlias.Equal(sAlias) {
			newElems[i] = sAlias
			changed = true
			continue
		}
		newElems[i] = pe
	}

	if !changed {
		return planAliases, false, diags
	}

	newSet, d := types.SetValue(NewAliasObjectType(), newElems)
	diags.Append(d...)
	if diags.HasError() {
		return planAliases, false, diags
	}
	return newSet, true, diags
}

func aliasObjectsByName(ctx context.Context, set basetypes.SetValue) (map[string]AliasObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := make(map[string]AliasObjectValue)
	for _, e := range set.Elements() {
		av, ok, d := aliasObjectFromAttr(ctx, e)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		if !ok || av.IsNull() || av.IsUnknown() {
			continue
		}
		var m AliasElementModel
		diags.Append(av.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return nil, diags
		}
		out[m.Name.ValueString()] = av
	}
	return out, diags
}

func aliasObjectFromAttr(ctx context.Context, v attr.Value) (AliasObjectValue, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if v == nil || v.IsNull() || v.IsUnknown() {
		return AliasObjectValue{}, false, diags
	}
	switch t := v.(type) {
	case AliasObjectValue:
		return t, true, diags
	case basetypes.ObjectValue:
		valuable, d := NewAliasObjectType().ValueFromObject(ctx, t)
		diags.Append(d...)
		if diags.HasError() {
			return AliasObjectValue{}, false, diags
		}
		av, ok := valuable.(AliasObjectValue)
		if !ok {
			diags.AddError(
				"Internal error",
				fmt.Sprintf("expected AliasObjectValue from alias ValueFromObject, got %T", valuable),
			)
			return AliasObjectValue{}, false, diags
		}
		return av, true, diags
	default:
		return AliasObjectValue{}, false, diags
	}
}
