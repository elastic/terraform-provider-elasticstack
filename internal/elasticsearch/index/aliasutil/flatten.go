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
	"encoding/json"
	"sort"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// NormalizeAliasFilterMap applies NormalizeQueryFilter to an already-decoded map
// and serializes the result to a jsontypes.Normalized.
// Returns jsontypes.NewNormalizedNull() when filterMap is empty or nil.
func NormalizeAliasFilterMap(filterMap map[string]any) (jsontypes.Normalized, diag.Diagnostics) {
	if len(filterMap) == 0 {
		return jsontypes.NewNormalizedNull(), nil
	}
	if nm, ok := normalizeQueryFilter(filterMap).(map[string]any); ok {
		filterMap = nm
	}
	normalizedBytes, err := json.Marshal(filterMap)
	if err != nil {
		return jsontypes.NewNormalizedNull(), diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
		}
	}
	return jsontypes.NewNormalizedValue(string(normalizedBytes)), nil
}

// NormalizeAliasFilterAnyToMap marshals v to JSON, unmarshals to map[string]any,
// then applies NormalizeQueryFilter. Returns (nil, nil) when v is nil.
func NormalizeAliasFilterAnyToMap(v any) (map[string]any, diag.Diagnostics) {
	if v == nil {
		return nil, nil
	}
	filterBytes, err := json.Marshal(v)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
		}
	}
	var filterMap map[string]any
	if err := json.Unmarshal(filterBytes, &filterMap); err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to unmarshal alias filter", err.Error()),
		}
	}
	if nm, ok := normalizeQueryFilter(filterMap).(map[string]any); ok {
		return nm, nil
	}
	return filterMap, nil
}

// NormalizeAliasFilterFromAny serializes v to a canonical jsontypes.Normalized alias filter.
// Returns jsontypes.NewNormalizedNull() when v is nil.
func NormalizeAliasFilterFromAny(v any) (jsontypes.Normalized, diag.Diagnostics) {
	if v == nil {
		return jsontypes.NewNormalizedNull(), nil
	}
	filterMap, diags := NormalizeAliasFilterAnyToMap(v)
	if diags.HasError() {
		return jsontypes.NewNormalizedNull(), diags
	}
	return NormalizeAliasFilterMap(filterMap)
}

// AliasAttrsFromModel builds the shared attribute map for a single alias from its models.IndexAlias
// representation. The returned map contains name, index_routing, routing, search_routing, is_hidden,
// is_write_index, and filter. Callers may override individual entries before constructing the final ObjectValue.
func AliasAttrsFromModel(name string, a models.IndexAlias) (map[string]attr.Value, diag.Diagnostics) {
	attrs := map[string]attr.Value{
		"name":           types.StringValue(name),
		"index_routing":  types.StringValue(a.IndexRouting),
		"routing":        types.StringValue(a.Routing),
		"search_routing": types.StringValue(a.SearchRouting),
		"is_hidden":      types.BoolValue(a.IsHidden),
		"is_write_index": types.BoolValue(a.IsWriteIndex),
	}
	filter, diags := NormalizeAliasFilterMap(a.Filter)
	if diags.HasError() {
		return nil, diags
	}
	attrs["filter"] = filter
	return attrs, nil
}

// AliasAttrsFromModelWithRouting is like AliasAttrsFromModel but also applies routing preservation:
// when the API omits the routing field (empty string) and preservedRouting contains a user-configured
// value for this alias name, that value is restored in the returned attribute map. This handles the
// Elasticsearch round-trip behavior where user-configured routing is not echoed back by the GET response.
func AliasAttrsFromModelWithRouting(name string, a models.IndexAlias, preservedRouting map[string]string) (map[string]attr.Value, diag.Diagnostics) {
	attrs, diags := AliasAttrsFromModel(name, a)
	if diags.HasError() {
		return nil, diags
	}
	if a.Routing == "" {
		if pr, ok := preservedRouting[name]; ok {
			attrs["routing"] = types.StringValue(pr)
		}
	}
	return attrs, diags
}

// FlattenAliasElement builds a types.Object for a single alias from its models.IndexAlias
// representation. preservedRouting carries user-configured routing values (may be nil) to restore
// when the API omits them. attrTypes specifies the attribute type map for the resulting object value.
func FlattenAliasElement(name string, a models.IndexAlias, preservedRouting map[string]string, attrTypes map[string]attr.Type) (attr.Value, diag.Diagnostics) {
	attrs, diags := AliasAttrsFromModelWithRouting(name, a, preservedRouting)
	if diags.HasError() {
		return nil, diags
	}
	obj, d := types.ObjectValue(attrTypes, attrs)
	diags.Append(d...)
	return obj, diags
}

// NewAliasModelFromAPI constructs an AliasModel from an API Alias response.
func NewAliasModelFromAPI(name string, apiModel estypes.Alias) (AliasModel, diag.Diagnostics) {
	tfAlias := AliasModel{
		Name:          types.StringValue(name),
		IndexRouting:  types.StringValue(typeutils.Deref(apiModel.IndexRouting)),
		IsHidden:      types.BoolValue(typeutils.Deref(apiModel.IsHidden)),
		IsWriteIndex:  types.BoolValue(typeutils.Deref(apiModel.IsWriteIndex)),
		Routing:       types.StringValue(typeutils.Deref(apiModel.Routing)),
		SearchRouting: types.StringValue(typeutils.Deref(apiModel.SearchRouting)),
	}

	filter, diags := NormalizeAliasFilterFromAny(apiModel.Filter)
	if diags.HasError() {
		return AliasModel{}, diags
	}
	tfAlias.Filter = filter

	return tfAlias, nil
}

// AliasesFromAPI converts an API aliases map into a Terraform types.Set value using the
// provided element type. The elemType parameter accommodates callers whose aliasElementType
// helper takes a context argument and those that do not.
func AliasesFromAPI(ctx context.Context, apiAliases map[string]estypes.Alias, elemType attr.Type) (basetypes.SetValue, diag.Diagnostics) {
	aliases := []AliasModel{}
	for name, alias := range apiAliases {
		tfAlias, diags := NewAliasModelFromAPI(name, alias)
		if diags.HasError() {
			return basetypes.SetValue{}, diags
		}
		aliases = append(aliases, tfAlias)
	}

	modelAliases, diags := types.SetValueFrom(ctx, elemType, aliases)
	if diags.HasError() {
		return basetypes.SetValue{}, diags
	}

	return modelAliases, nil
}

// FlattenAliasSet maps an Elasticsearch alias API response map into a types.Set.
// Aliases are iterated in sorted name order for stable plan output.
// preservedRouting carries user-configured routing values to restore when the API omits them (may be nil).
// elemType is the element type for the resulting set (callers may pass a custom object type).
// attrTypes is the attribute type map used when constructing each alias element object.
func FlattenAliasSet(ctx context.Context, aliases map[string]models.IndexAlias, preservedRouting map[string]string, elemType attr.Type, attrTypes map[string]attr.Type) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(aliases) == 0 {
		return types.SetNull(elemType), diags
	}

	names := make([]string, 0, len(aliases))
	for name := range aliases {
		names = append(names, name)
	}
	sort.Strings(names)

	vals := make([]attr.Value, 0, len(names))
	for _, name := range names {
		alias := aliases[name]
		av, d := FlattenAliasElement(name, alias, preservedRouting, attrTypes)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetUnknown(elemType), diags
		}
		vals = append(vals, av)
	}

	sv, d := types.SetValueFrom(ctx, elemType, vals)
	diags.Append(d...)
	return sv, diags
}

// normalizeQueryFilter recursively compacts expanded single-key query values
// produced by the typed client back to their shorthand form.
// For example: {"term":{"field":{"value":"x"}}} → {"term":{"field":"x"}}
func normalizeQueryFilter(v any) any {
	switch val := v.(type) {
	case map[string]any:
		if len(val) == 1 {
			if inner, ok := val["value"]; ok {
				switch inner.(type) {
				case string, float64, bool, int, int64:
					return inner
				}
			}
		}
		out := make(map[string]any, len(val))
		for k, vv := range val {
			out[k] = normalizeQueryFilter(vv)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, vv := range val {
			out[i] = normalizeQueryFilter(vv)
		}
		return out
	default:
		return v
	}
}
