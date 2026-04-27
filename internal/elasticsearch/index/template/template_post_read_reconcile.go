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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// postReadReconcileTemplateWithPlan aligns refreshed template.settings and template.alias with the
// planned or prior Terraform model when values are semantically equivalent to the API response.
// Terraform correlates nested set elements using strict attr.Value equality; alias routing echoes
// from Elasticsearch are normalized first by enrichTemplateAliasesRoutingFromReference, then
// substituted with the planned object here when AliasObjectValue.ObjectSemanticEquals agrees. Settings are rewritten
// to canonical Elasticsearch JSON when semantically equal to the plan, matching legacy SDK state.
func aliasIsRoutingOnlyHCLShape(planM AliasElementModel) bool {
	if planM.Routing.IsNull() || planM.Routing.ValueString() == "" {
		return false
	}
	ir := planM.IndexRouting.IsNull() || planM.IndexRouting.ValueString() == ""
	sr := planM.SearchRouting.IsNull() || planM.SearchRouting.ValueString() == ""
	return ir && sr
}

// mergeAliasObjectForStableState combines API-fresh alias fields with plan defaults so state matches
// Terraform's configured defaults (e.g. is_write_index) while preserving routing echoes from Elasticsearch
// when the practitioner omitted index_routing/search_routing.
func mergeAliasObjectForStableState(ctx context.Context, readAv, planAv AliasObjectValue) (AliasObjectValue, diag.Diagnostics) {
	var readM, planM AliasElementModel
	var diags diag.Diagnostics
	diags.Append(readAv.As(ctx, &readM, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
	diags.Append(planAv.As(ctx, &planM, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
	if diags.HasError() {
		return AliasObjectValue{}, diags
	}
	attrs := readAv.Attributes()
	attrs["name"] = planM.Name
	attrs["is_hidden"] = planM.IsHidden
	attrs["is_write_index"] = planM.IsWriteIndex
	if !planM.Filter.IsNull() {
		attrs["filter"] = planM.Filter
	} else {
		attrs["filter"] = readM.Filter
	}
	switch {
	case !planM.Routing.IsNull() && planM.Routing.ValueString() != "":
		attrs["routing"] = planM.Routing
	default:
		attrs["routing"] = readM.Routing
	}
	switch {
	case !planM.IndexRouting.IsNull() && planM.IndexRouting.ValueString() != "":
		attrs["index_routing"] = planM.IndexRouting
	default:
		attrs["index_routing"] = readM.IndexRouting
	}
	switch {
	case !planM.SearchRouting.IsNull() && planM.SearchRouting.ValueString() != "":
		attrs["search_routing"] = planM.SearchRouting
	default:
		attrs["search_routing"] = readM.SearchRouting
	}
	return NewAliasObjectValue(attrs)
}

func postReadReconcileTemplateWithPlan(ctx context.Context, read *Model, plan Model) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(reconcileTemplateAliasesStrictFromPlan(ctx, read, plan)...)
	diags.Append(reconcileTemplateSettingsCanonicalFromPlan(ctx, read, plan)...)
	return diags
}

func reconcileTemplateAliasesStrictFromPlan(ctx context.Context, read *Model, plan Model) diag.Diagnostics {
	var diags diag.Diagnostics
	if read.Template.IsNull() || read.Template.IsUnknown() || plan.Template.IsNull() || plan.Template.IsUnknown() {
		return diags
	}

	readAttrs := read.Template.Attributes()
	planAttrs := plan.Template.Attributes()
	readAliasVal, ok := readAttrs["alias"]
	if !ok || readAliasVal.IsNull() || readAliasVal.IsUnknown() {
		return diags
	}
	planAliasVal, ok := planAttrs["alias"]
	if !ok || planAliasVal.IsNull() || planAliasVal.IsUnknown() {
		return diags
	}

	if readAliasVal.Equal(planAliasVal) {
		return diags
	}

	readSet, ok := readAliasVal.(types.Set)
	if !ok {
		diags.AddError("Internal error", "expected Set for template.alias on read model")
		return diags
	}
	planSet, ok := planAliasVal.(types.Set)
	if !ok {
		diags.AddError("Internal error", "expected Set for template.alias on plan model")
		return diags
	}

	// Routing-only alias with a single element: use the plan's Set value verbatim so state matches
	// Terraform's planned tftypes for nested blocks (avoids SetValueFrom re-encoding churn).
	if len(planSet.Elements()) == 1 && len(readSet.Elements()) == 1 {
		planEl := planSet.Elements()[0]
		readEl := readSet.Elements()[0]
		planAv, planOK, d := coerceAliasObjectValue(ctx, planEl)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		readAv, readOK, d := coerceAliasObjectValue(ctx, readEl)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if planOK && readOK {
			var planM AliasElementModel
			diags.Append(planAv.As(ctx, &planM, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
			if diags.HasError() {
				return diags
			}
			if aliasIsRoutingOnlyHCLShape(planM) {
				eq, d := readAv.ObjectSemanticEquals(ctx, planAv)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				if eq {
					readAttrs = read.Template.Attributes()
					readAttrs["alias"] = planAliasVal
					tplObj, d := types.ObjectValue(TemplateAttrTypes(), readAttrs)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}
					read.Template = tplObj
					return diags
				}
			}
		}
	}

	planByName := make(map[string]AliasObjectValue)
	for _, planEl := range planSet.Elements() {
		planAv, ok, d := coerceAliasObjectValue(ctx, planEl)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if !ok {
			continue
		}
		var pm AliasElementModel
		diags.Append(planAv.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return diags
		}
		planByName[pm.Name.ValueString()] = planAv
	}

	newElems := make([]attr.Value, 0, len(readSet.Elements()))
	for _, el := range readSet.Elements() {
		readAv, ok, d := coerceAliasObjectValue(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if !ok {
			newElems = append(newElems, el)
			continue
		}
		var rm AliasElementModel
		diags.Append(readAv.As(ctx, &rm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return diags
		}
		planAv, ok := planByName[rm.Name.ValueString()]
		if !ok {
			newElems = append(newElems, readAv)
			continue
		}
		if readAv.Equal(planAv) {
			newElems = append(newElems, readAv)
			continue
		}
		eq, d := readAv.ObjectSemanticEquals(ctx, planAv)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if eq {
			var planM AliasElementModel
			diags.Append(planAv.As(ctx, &planM, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
			if diags.HasError() {
				return diags
			}
			// Routing-only aliases: use the plan object so the nested set element matches configuration
			// exactly (avoids TypeSet churn from API-echoed index_routing/search_routing).
			if aliasIsRoutingOnlyHCLShape(planM) {
				newElems = append(newElems, planAv)
				continue
			}
			merged, d := mergeAliasObjectForStableState(ctx, readAv, planAv)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			newElems = append(newElems, merged)
			continue
		}
		newElems = append(newElems, readAv)
	}

	newSet, d := types.SetValueFrom(ctx, NewAliasObjectType(), newElems)
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

func reconcileTemplateSettingsCanonicalFromPlan(ctx context.Context, read *Model, plan Model) diag.Diagnostics {
	var diags diag.Diagnostics
	if read.Template.IsNull() || read.Template.IsUnknown() || plan.Template.IsNull() || plan.Template.IsUnknown() {
		return diags
	}

	readAttrs := read.Template.Attributes()
	planAttrs := plan.Template.Attributes()
	rs, ok := readAttrs["settings"]
	if !ok || rs.IsNull() || rs.IsUnknown() {
		return diags
	}
	ps, ok := planAttrs["settings"]
	if !ok || ps.IsNull() || ps.IsUnknown() {
		return diags
	}

	readSettings, ok := rs.(customtypes.IndexSettingsValue)
	if !ok {
		diags.AddError("Internal error", "expected IndexSettingsValue for template.settings on read model")
		return diags
	}
	planSettings, ok := ps.(customtypes.IndexSettingsValue)
	if !ok {
		diags.AddError("Internal error", "expected IndexSettingsValue for template.settings on plan model")
		return diags
	}

	eq, d := readSettings.SemanticallyEqual(ctx, planSettings)
	diags.Append(d...)
	if diags.HasError() || !eq {
		return diags
	}

	canon, err := customtypes.CanonicalIndexSettingsJSON(readSettings.ValueString())
	if err != nil {
		diags.AddWarning("Index template settings normalization", err.Error())
		return diags
	}

	readAttrs = read.Template.Attributes()
	readAttrs["settings"] = customtypes.NewIndexSettingsValue(canon)
	tplObj, d := types.ObjectValue(TemplateAttrTypes(), readAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	read.Template = tplObj
	return diags
}
