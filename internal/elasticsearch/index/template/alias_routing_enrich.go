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

func coerceAliasObjectValue(ctx context.Context, v attr.Value) (AliasObjectValue, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if av, ok := v.(AliasObjectValue); ok {
		return av, true, diags
	}
	ov, ok := v.(basetypes.ObjectValue)
	if !ok {
		return AliasObjectValue{}, false, diags
	}
	valuable, d := NewAliasObjectType().ValueFromObject(ctx, ov)
	diags.Append(d...)
	if diags.HasError() {
		return AliasObjectValue{}, false, diags
	}
	aov, ok := valuable.(AliasObjectValue)
	if !ok {
		return AliasObjectValue{}, false, diags
	}
	return aov, ok, diags
}

// enrichTemplateAliasesRoutingFromReference copies configured `routing` from ref (plan or prior
// state) onto the read model when Elasticsearch omitted `routing` on GET but the alias is still
// semantically equivalent. This mirrors the legacy SDK preserveAliasRoutingInFlattenedAliases
// behavior and keeps post-apply state aligned with configuration.
func enrichTemplateAliasesRoutingFromReference(ctx context.Context, read *Model, ref Model) diag.Diagnostics {
	var diags diag.Diagnostics
	if read.Template.IsNull() || read.Template.IsUnknown() || ref.Template.IsNull() || ref.Template.IsUnknown() {
		return diags
	}

	readAttrs := read.Template.Attributes()
	refAttrs := ref.Template.Attributes()
	readAliasVal, ok := readAttrs["alias"]
	if !ok || readAliasVal.IsNull() || readAliasVal.IsUnknown() {
		return diags
	}
	refAliasVal, ok := refAttrs["alias"]
	if !ok || refAliasVal.IsNull() || refAliasVal.IsUnknown() {
		return diags
	}

	readSet, ok := readAliasVal.(types.Set)
	if !ok {
		diags.AddError("Internal error", "expected Set for template.alias on read model")
		return diags
	}
	refSet, ok := refAliasVal.(types.Set)
	if !ok {
		diags.AddError("Internal error", "expected Set for template.alias on reference model")
		return diags
	}

	refByName := make(map[string]AliasObjectValue)
	for _, el := range refSet.Elements() {
		refAv, ok, d := coerceAliasObjectValue(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if !ok {
			continue
		}
		var refM AliasElementModel
		diags.Append(refAv.As(ctx, &refM, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return diags
		}
		refByName[refM.Name.ValueString()] = refAv
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
		var readM AliasElementModel
		diags.Append(readAv.As(ctx, &readM, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return diags
		}
		refAv, ok := refByName[readM.Name.ValueString()]
		if !ok {
			newElems = append(newElems, readAv)
			continue
		}
		var refM AliasElementModel
		diags.Append(refAv.As(ctx, &refM, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		if diags.HasError() {
			return diags
		}

		eq, d := aliasElementModelsSemanticallyEqual(ctx, refM, readM)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if !eq {
			newElems = append(newElems, readAv)
			continue
		}

		refRUnset := refM.Routing.IsNull() || refM.Routing.ValueString() == ""
		readRUnset := readM.Routing.IsNull() || readM.Routing.ValueString() == ""
		if refRUnset || !readRUnset {
			newElems = append(newElems, readAv)
			continue
		}
		// Do not inject routing when ES echoed generic routing into index_routing and omitted routing;
		// copying routing here would break AliasObjectValue.ObjectSemanticEquals (see aliasEsIndexRoutingEchoesPriorMainRouting).
		if aliasEsIndexRoutingEchoesPriorMainRouting(refM, readM) {
			newElems = append(newElems, readAv)
			continue
		}

		attrs := readAv.Attributes()
		attrs["routing"] = refM.Routing
		patched, d := NewAliasObjectValue(attrs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		newElems = append(newElems, patched)
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
