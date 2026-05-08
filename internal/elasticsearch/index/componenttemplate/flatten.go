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

package componenttemplate

import (
	"context"
	"encoding/json"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// flattenToData maps an API ClusterComponentTemplate response into a Data value.
// The prior Data is used to carry over the ID, ElasticsearchConnection, and
// alias routing values that the API omits on round-trip.
func flattenToData(ctx context.Context, tpl *estypes.ClusterComponentTemplate, prior Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	out := Data{
		ID:                      prior.ID,
		ElasticsearchConnection: prior.ElasticsearchConnection,
		Name:                    types.StringValue(tpl.Name),
	}

	// Metadata
	if tpl.ComponentTemplate.Meta_ != nil {
		b, err := json.Marshal(tpl.ComponentTemplate.Meta_)
		if err != nil {
			diags.AddError("Failed to marshal metadata", err.Error())
			return out, diags
		}
		out.Metadata = jsontypes.NewNormalizedValue(string(b))
	} else {
		out.Metadata = jsontypes.NewNormalizedNull()
	}

	// Version (ComponentTemplateNode.Version is *int64)
	if tpl.ComponentTemplate.Version != nil {
		out.Version = types.Int64Value(*tpl.ComponentTemplate.Version)
	} else {
		out.Version = types.Int64Null()
	}

	// Template block
	preservedRouting := extractAliasRoutingFromData(prior)
	templateObj, d := flattenTemplateBlock(ctx, tpl, preservedRouting)
	diags.Append(d...)
	if diags.HasError() {
		return out, diags
	}
	out.Template = templateObj

	return out, diags
}

// flattenTemplateBlock maps the component template's template block into a types.Object.
func flattenTemplateBlock(ctx context.Context, tpl *estypes.ClusterComponentTemplate, preservedRouting map[string]string) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	t := tpl.ComponentTemplate.Template

	// Mappings
	var mappings jsontypes.Normalized
	if t.Mappings != nil {
		b, err := json.Marshal(t.Mappings)
		if err != nil {
			diags.AddError("Failed to marshal template.mappings", err.Error())
			return types.ObjectNull(templateAttrTypes()), diags
		}
		mappings = jsontypes.NewNormalizedValue(string(b))
	} else {
		mappings = jsontypes.NewNormalizedNull()
	}

	// Settings
	var settings customtypes.IndexSettingsValue
	if t.Settings != nil {
		b, err := json.Marshal(t.Settings)
		if err != nil {
			diags.AddError("Failed to marshal template.settings", err.Error())
			return types.ObjectNull(templateAttrTypes()), diags
		}
		settings = customtypes.NewIndexSettingsValue(string(b))
	} else {
		settings = customtypes.NewIndexSettingsNull()
	}

	// Aliases
	aliasSet, d := flattenAliasSet(ctx, t.Aliases, preservedRouting)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(templateAttrTypes()), diags
	}

	tplAttrs := map[string]attr.Value{
		"alias":    aliasSet,
		"mappings": mappings,
		"settings": settings,
	}

	obj, d := types.ObjectValue(templateAttrTypes(), tplAttrs)
	diags.Append(d...)
	return obj, diags
}

// flattenAliasSet maps Elasticsearch alias API responses into a types.Set.
// preservedRouting carries user-configured routing values to restore when the API omits them.
func flattenAliasSet(ctx context.Context, aliases map[string]estypes.AliasDefinition, preservedRouting map[string]string) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	aliasElemType := types.ObjectType{AttrTypes: aliasAttrTypes()}

	if len(aliases) == 0 {
		return types.SetNull(aliasElemType), diags
	}

	vals := make([]attr.Value, 0, len(aliases))
	for name, alias := range aliases {
		av, d := flattenAliasElement(name, alias, preservedRouting)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetUnknown(aliasElemType), diags
		}
		vals = append(vals, av)
	}

	sv, d := types.SetValueFrom(ctx, aliasElemType, vals)
	diags.Append(d...)
	return sv, diags
}

// flattenAliasElement maps a single Elasticsearch alias API response to a types.Object.
func flattenAliasElement(name string, a estypes.AliasDefinition, preservedRouting map[string]string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	routing := typeutils.Deref(a.Routing)
	// Preserve routing from prior state when the API omits it
	if routing == "" {
		if pr, ok := preservedRouting[name]; ok {
			routing = pr
		}
	}

	attrs := map[string]attr.Value{
		"name":           types.StringValue(name),
		"index_routing":  types.StringValue(typeutils.Deref(a.IndexRouting)),
		"routing":        types.StringValue(routing),
		"search_routing": types.StringValue(typeutils.Deref(a.SearchRouting)),
		"is_hidden":      types.BoolValue(typeutils.Deref(a.IsHidden)),
		"is_write_index": types.BoolValue(typeutils.Deref(a.IsWriteIndex)),
	}

	if a.Filter != nil {
		b, err := json.Marshal(a.Filter)
		if err != nil {
			diags.AddError("Failed to marshal alias filter", err.Error())
			return nil, diags
		}
		attrs["filter"] = jsontypes.NewNormalizedValue(string(b))
	} else {
		attrs["filter"] = jsontypes.NewNormalizedNull()
	}

	obj, d := types.ObjectValue(aliasAttrTypes(), attrs)
	diags.Append(d...)
	return obj, diags
}

// extractAliasRoutingFromData extracts user-configured alias routing values from
// prior Data state so they can be preserved when the API omits them on read.
func extractAliasRoutingFromData(prior Data) map[string]string {
	result := make(map[string]string)
	if prior.Template.IsNull() || prior.Template.IsUnknown() {
		return result
	}

	tplAttrs := prior.Template.Attributes()
	aliasVal, ok := tplAttrs["alias"]
	if !ok {
		return result
	}
	aliasSet, ok := aliasVal.(types.Set)
	if !ok || aliasSet.IsNull() || aliasSet.IsUnknown() {
		return result
	}

	for _, el := range aliasSet.Elements() {
		obj, ok := el.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		nameAttr, ok := attrs["name"].(types.String)
		if !ok || nameAttr.IsNull() || nameAttr.IsUnknown() {
			continue
		}
		routingAttr, ok := attrs["routing"].(types.String)
		if !ok || routingAttr.IsNull() || routingAttr.IsUnknown() {
			continue
		}
		name := nameAttr.ValueString()
		routing := routingAttr.ValueString()
		if name != "" && routing != "" {
			result[name] = routing
		}
	}

	return result
}
