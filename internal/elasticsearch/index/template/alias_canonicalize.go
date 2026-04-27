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

// canonicalizeAliasSetElements applies [canonicalizeAliasObjectForState] to each element of a template.alias set.
func canonicalizeAliasSetElements(ctx context.Context, set attr.Value) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	sSet, ok := set.(basetypes.SetValue)
	if !ok {
		return set, diags
	}
	if sSet.IsNull() || sSet.IsUnknown() {
		return set, diags
	}
	elems := sSet.Elements()
	outElems := make([]attr.Value, len(elems))
	for i, el := range elems {
		av, ok, d := aliasObjectFromAttr(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return set, diags
		}
		if !ok {
			outElems[i] = el
			continue
		}
		canon, d := canonicalizeAliasObjectForState(ctx, av)
		diags.Append(d...)
		if diags.HasError() {
			return set, diags
		}
		outElems[i] = canon
	}
	newSet, d := types.SetValue(NewAliasObjectType(), outElems)
	diags.Append(d...)
	if diags.HasError() {
		return set, diags
	}
	return newSet, diags
}

// canonicalizeTemplateAliasSetInModel rewrites template.alias so Optional+Computed attributes use the
// same known representations as schema defaults (empty string / false). Config and plan use those
// defaults during non-refresh planning; prior state from config/plan reconciliation can still hold
// null unknowns, which makes set elements hash differently and triggers perpetual diffs.
func canonicalizeTemplateAliasSetInModel(ctx context.Context, m *Model) diag.Diagnostics {
	var diags diag.Diagnostics
	if m.Template.IsNull() || m.Template.IsUnknown() {
		return diags
	}
	attrs := m.Template.Attributes()
	aliasVal, ok := attrs["alias"]
	if !ok || aliasVal.IsNull() || aliasVal.IsUnknown() {
		return diags
	}
	setV, ok := aliasVal.(basetypes.SetValue)
	if !ok {
		return diags
	}
	elems := setV.Elements()
	outElems := make([]attr.Value, len(elems))
	for i, el := range elems {
		av, ok, d := aliasObjectFromAttr(ctx, el)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if !ok {
			outElems[i] = el
			continue
		}
		canon, d := canonicalizeAliasObjectForState(ctx, av)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		outElems[i] = canon
	}
	newSet, d := types.SetValue(NewAliasObjectType(), outElems)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	attrs["alias"] = newSet
	newTpl, d := types.ObjectValue(TemplateAttrTypes(), attrs)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.Template = newTpl
	return diags
}

func canonicalizeAliasObjectForState(ctx context.Context, v AliasObjectValue) (AliasObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	var m AliasElementModel
	diags.Append(v.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
	if diags.HasError() {
		return v, diags
	}
	if m.IndexRouting.IsNull() || m.IndexRouting.IsUnknown() {
		m.IndexRouting = types.StringValue("")
	}
	if m.SearchRouting.IsNull() || m.SearchRouting.IsUnknown() {
		m.SearchRouting = types.StringValue("")
	}
	if m.Routing.IsNull() || m.Routing.IsUnknown() {
		m.Routing = types.StringValue("")
	}
	if m.IsHidden.IsNull() || m.IsHidden.IsUnknown() {
		m.IsHidden = types.BoolValue(false)
	}
	if m.IsWriteIndex.IsNull() || m.IsWriteIndex.IsUnknown() {
		m.IsWriteIndex = types.BoolValue(false)
	}
	attrs := map[string]attr.Value{
		"name":           m.Name,
		"index_routing":  m.IndexRouting,
		"routing":        m.Routing,
		"search_routing": m.SearchRouting,
		"filter":         m.Filter,
		"is_hidden":      m.IsHidden,
		"is_write_index": m.IsWriteIndex,
	}
	out, d := NewAliasObjectValue(attrs)
	diags.Append(d...)
	return out, diags
}
