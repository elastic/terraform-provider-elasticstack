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
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// finalizeTemplateAliasAfterRead rewrites template.alias after read: when cfg/plan is non-null, any alias
// that is semantically equivalent to the planned alias is replaced with the merged plan reference object
// (cfg entries overwritten by plan so planned defaults match Terraform's post-apply correlation).
// cfg is the practitioner configuration; plan holds provider-planned values (defaults applied).
func finalizeTemplateAliasAfterRead(ctx context.Context, read *Model, cfg, plan Model) diag.Diagnostics {
	var diags diag.Diagnostics
	if read.Template.IsNull() || read.Template.IsUnknown() {
		return diags
	}
	readAttrs := read.Template.Attributes()
	readAliasVal, ok := readAttrs["alias"]
	if !ok || readAliasVal.IsNull() || readAliasVal.IsUnknown() {
		return diags
	}
	readSet, ok := readAliasVal.(types.Set)
	if !ok {
		diags.AddError("Internal error", "expected Set for template.alias")
		return diags
	}

	planByName := aliasObjectValuesByNameFromTemplate(ctx, cfg, &diags)
	if diags.HasError() {
		return diags
	}
	maps.Copy(planByName, aliasObjectValuesByNameFromTemplate(ctx, plan, &diags))
	if diags.HasError() {
		return diags
	}

	elems := readSet.Elements()
	if len(elems) == 0 {
		return diags
	}
	out := make([]attr.Value, 0, len(elems))
	for _, el := range elems {
		readAv, ok, d := coerceAliasObjectValue(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if !ok {
			out = append(out, el)
			continue
		}
		var readM AliasElementModel
		diags.Append(readAv.As(ctx, &readM, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return diags
		}
		chosen := readAv
		if planAv, found := planByName[readM.Name.ValueString()]; found {
			var planM AliasElementModel
			diags.Append(planAv.As(ctx, &planM, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
			if diags.HasError() {
				return diags
			}
			fwd, d := aliasElementModelsSemanticallyEqual(ctx, planM, readM)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			rev, d := aliasElementModelsSemanticallyEqual(ctx, readM, planM)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			if fwd || rev {
				chosen = planAv
			}
		}
		norm, d := canonicalizeAliasObjectEncoding(ctx, chosen)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		out = append(out, norm)
	}
	newSet, d := types.SetValueFrom(ctx, NewAliasObjectType(), out)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	readAttrs = read.Template.Attributes()
	readAttrs["alias"] = newSet
	tplObj, d := types.ObjectValue(TemplateAttrTypes(), readAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	read.Template = tplObj
	return diags
}

func aliasObjectValuesByNameFromTemplate(ctx context.Context, m Model, diags *diag.Diagnostics) map[string]AliasObjectValue {
	out := map[string]AliasObjectValue{}
	if m.Template.IsNull() || m.Template.IsUnknown() {
		return out
	}
	pAttrs := m.Template.Attributes()
	pAlias, ok := pAttrs["alias"]
	if !ok || pAlias.IsNull() || pAlias.IsUnknown() {
		return out
	}
	pSet, ok := pAlias.(types.Set)
	if !ok {
		diags.AddError("Internal error", "expected Set for template.alias")
		return out
	}
	for _, el := range pSet.Elements() {
		pAv, ok, d := coerceAliasObjectValue(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return out
		}
		if !ok {
			continue
		}
		var pm AliasElementModel
		diags.Append(pAv.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return out
		}
		out[pm.Name.ValueString()] = pAv
	}
	return out
}

func canonicalizeAliasObjectEncoding(_ context.Context, av AliasObjectValue) (AliasObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	return av, diags
}
