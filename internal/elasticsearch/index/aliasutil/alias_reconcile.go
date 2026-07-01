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

package aliasutil

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ApplyTemplateAliasReconciliationFromReference replaces template.alias elements read from the API
// with the reference encoding when ObjectSemanticEquals matches. Use configuration on Create and
// Update (req.Plan can leave Optional+Computed nested attributes unknown); use prior state on Read.
func ApplyTemplateAliasReconciliationFromReference(ctx context.Context, outTemplate *types.Object, refTemplate types.Object, templateAttrTypes map[string]attr.Type) diag.Diagnostics {
	var diags diag.Diagnostics
	if outTemplate.IsNull() || outTemplate.IsUnknown() {
		return diags
	}
	if refTemplate.IsNull() || refTemplate.IsUnknown() {
		return diags
	}

	outAttrs := outTemplate.Attributes()
	refAttrs := refTemplate.Attributes()
	apiAliasVal := outAttrs[templateAliasAttrKey]
	refAliasVal := refAttrs[templateAliasAttrKey]
	if apiAliasVal.IsNull() || apiAliasVal.IsUnknown() {
		return diags
	}
	if refAliasVal.IsNull() || refAliasVal.IsUnknown() {
		return diags
	}

	merged, changed, d := MergeAliasSetPreferReferenceEncoding(ctx, apiAliasVal, refAliasVal)
	diags.Append(d...)
	if diags.HasError() || !changed {
		return diags
	}

	outAttrs[templateAliasAttrKey] = merged
	newTpl, d := types.ObjectValue(templateAttrTypes, outAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	*outTemplate = newTpl
	return diags
}

// MergePlanAliasSetWithPriorState walks planned alias elements; when prior state has the same name and
// planElt.ObjectSemanticEquals(stateElt), replace the plan element with the state's encoding so planned
// values match stored state under Optional+Computed+Default nested sets.
func MergePlanAliasSetWithPriorState(ctx context.Context, planAliases, stateAliases attr.Value) (attr.Value, bool, diag.Diagnostics) {
	return mergeAliasSets(ctx, planAliases, stateAliases)
}

// MergeAliasSetPreferReferenceEncoding replaces API alias encodings with reference encodings when semantically equal.
func MergeAliasSetPreferReferenceEncoding(ctx context.Context, apiSet, refSet attr.Value) (attr.Value, bool, diag.Diagnostics) {
	return mergeAliasSets(ctx, apiSet, refSet)
}

// mergeAliasSets walks sourceSet elements; when lookupSet has a same-named element that is
// semantically equal but not byte-equal, replaces the source element with the lookup element.
// Returns (newSet, changed, diags).
func mergeAliasSets(ctx context.Context, sourceSet, lookupSet attr.Value) (attr.Value, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	srcS, ok := sourceSet.(basetypes.SetValue)
	if !ok {
		return sourceSet, false, diags
	}
	lkpS, ok := lookupSet.(basetypes.SetValue)
	if !ok {
		return sourceSet, false, diags
	}
	if srcS.IsNull() || srcS.IsUnknown() || lkpS.IsNull() || lkpS.IsUnknown() {
		return sourceSet, false, diags
	}

	lookupByName, d := aliasObjectsByName(ctx, lkpS)
	diags.Append(d...)
	if diags.HasError() {
		return sourceSet, false, diags
	}

	srcElems := srcS.Elements()
	newElems := make([]attr.Value, len(srcElems))
	changed := false
	for i, se := range srcElems {
		srcAlias, srcOK, d := aliasObjectFromAttr(ctx, se)
		diags.Append(d...)
		if diags.HasError() {
			return sourceSet, false, diags
		}
		if !srcOK || srcAlias.IsNull() || srcAlias.IsUnknown() {
			newElems[i] = se
			continue
		}
		var sm AliasModel
		diags.Append(srcAlias.As(ctx, &sm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return sourceSet, false, diags
		}
		lkpAlias, ok := lookupByName[sm.Name.ValueString()]
		if !ok {
			newElems[i] = se
			continue
		}
		eq, d := srcAlias.ObjectSemanticEquals(ctx, lkpAlias)
		diags.Append(d...)
		if diags.HasError() {
			return sourceSet, false, diags
		}
		if eq && !srcAlias.Equal(lkpAlias) {
			newElems[i] = lkpAlias
			changed = true
			continue
		}
		newElems[i] = se
	}

	if !changed {
		return sourceSet, false, diags
	}

	newSet, d := types.SetValue(NewAliasObjectType(), newElems)
	diags.Append(d...)
	if diags.HasError() {
		return sourceSet, false, diags
	}
	return newSet, true, diags
}

// ProjectConfigAliasMatchesOntoPlan walks the plan's alias elements and, for each one, looks up
// the same-named config alias and the same-named state alias. When the config encoding semantically
// equals the state encoding, the plan element is replaced with the state encoding (so the planned
// value matches stored state under Optional+Computed+Default nested sets, even when the plan
// element carries unknowns). Plan elements with no config counterpart are preserved unchanged.
func ProjectConfigAliasMatchesOntoPlan(ctx context.Context, planAliases, configAliases, stateAliases attr.Value) (attr.Value, bool, diag.Diagnostics) {
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
		var pm AliasModel
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
		var m AliasModel
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
