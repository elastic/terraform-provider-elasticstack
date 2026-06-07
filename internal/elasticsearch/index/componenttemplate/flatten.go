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

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/templateutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// flattenToData maps a component template API response into a Data value.
// The prior Data is used to carry over the ID, ElasticsearchConnection, and
// alias routing values that the API omits on round-trip.
//
// The input is the locally-defined models.ComponentTemplateResponse populated
// by the raw GET response decoder, not the typed go-elasticsearch struct: see
// GetComponentTemplate and issue #3124 for why typed decoding is unsafe for
// free-form settings.
func flattenToData(ctx context.Context, tpl *models.ComponentTemplateResponse, prior Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	out := Data{
		ID:                      prior.ID,
		ElasticsearchConnection: prior.ElasticsearchConnection,
		Name:                    types.StringValue(tpl.Name),
	}

	if tpl.ComponentTemplate.Meta != nil {
		b, err := json.Marshal(tpl.ComponentTemplate.Meta)
		if err != nil {
			diags.AddError("Failed to marshal metadata", err.Error())
			return out, diags
		}
		out.Metadata = jsontypes.NewNormalizedValue(string(b))
	} else {
		out.Metadata = jsontypes.NewNormalizedNull()
	}

	if tpl.ComponentTemplate.Version != nil {
		out.Version = types.Int64Value(*tpl.ComponentTemplate.Version)
	} else {
		out.Version = types.Int64Null()
	}

	preservedRouting := extractAliasRoutingFromData(prior)
	priorMappings, priorSettings := extractEmptyObjectOverridesFromData(prior)
	templateObj, d := flattenTemplateBlock(ctx, tpl.ComponentTemplate.Template, preservedRouting, priorMappings, priorSettings)
	diags.Append(d...)
	if diags.HasError() {
		return out, diags
	}
	out.Template = templateObj

	return out, diags
}

// flattenTemplateBlock maps the component template's template block into a types.Object.
//
// priorMappings and priorSettings carry the prior Terraform values for
// `template.mappings` and `template.settings`. They are only consulted when the
// API response contains no mappings/settings (nil or empty map): if the prior
// value was a known, semantically-empty JSON object (for example because the
// practitioner explicitly wrote `mappings = jsonencode({})` in HCL), the prior
// value is preserved in state. This avoids the post-apply consistency error
// the Plugin Framework raises when the planned value `"{}"` collides with a
// flattened state value of `null`, since the framework short-circuits its
// semantic-equality check when either side of the comparison is null.
func flattenTemplateBlock(
	ctx context.Context,
	t *models.Template,
	preservedRouting map[string]string,
	priorMappings esindex.MappingsValue,
	priorSettings customtypes.IndexSettingsValue,
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if t == nil {
		return types.ObjectNull(templateAttrTypes()), diags
	}

	mappings, mappingsDiags := flattenMappings(t.Mappings, priorMappings)
	diags.Append(mappingsDiags...)
	if diags.HasError() {
		return types.ObjectNull(templateAttrTypes()), diags
	}

	settings, settingsDiags := flattenSettings(t.Settings, priorSettings)
	diags.Append(settingsDiags...)
	if diags.HasError() {
		return types.ObjectNull(templateAttrTypes()), diags
	}

	aliasSet, d := flattenAliasSet(ctx, t.Aliases, preservedRouting)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(templateAttrTypes()), diags
	}

	var dsoObj types.Object
	if t.DataStreamOptions != nil && t.DataStreamOptions.FailureStore != nil {
		var dsoDiags diag.Diagnostics
		dsoObj, dsoDiags = datastreamoptions.FlattenLocal(t.DataStreamOptions)
		diags.Append(dsoDiags...)
		if diags.HasError() {
			return types.ObjectNull(templateAttrTypes()), diags
		}
	} else {
		dsoObj = types.ObjectNull(datastreamoptions.AttrTypes())
	}

	tplAttrs := map[string]attr.Value{
		attrAlias:             aliasSet,
		attrMappings:          mappings,
		attrSettings:          settings,
		attrDataStreamOptions: dsoObj,
	}

	obj, d := types.ObjectValue(templateAttrTypes(), tplAttrs)
	diags.Append(d...)
	return obj, diags
}

// flattenAliasSet maps Elasticsearch alias API responses into a types.Set.
// preservedRouting carries user-configured routing values to restore when the API omits them.
func flattenAliasSet(ctx context.Context, aliases map[string]models.IndexAlias, preservedRouting map[string]string) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	aliasElemType := types.ObjectType{AttrTypes: aliasAttrTypes()}

	if len(aliases) == 0 {
		return types.SetNull(aliasElemType), diags
	}

	vals := make([]attr.Value, 0, len(aliases))
	for name, alias := range aliases {
		av, d := aliasutil.FlattenAliasElement(name, alias, preservedRouting, aliasAttrTypes())
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

// flattenMappings maps the API mappings response onto a MappingsValue, applying
// the prior-preservation rule when the API returns no mappings.
func flattenMappings(apiMappings map[string]any, prior esindex.MappingsValue) (esindex.MappingsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(apiMappings) > 0 {
		b, err := json.Marshal(apiMappings)
		if err != nil {
			diags.AddError("Failed to marshal template.mappings", err.Error())
			return esindex.NewMappingsNull(), diags
		}
		return esindex.NewMappingsValue(string(b)), diags
	}
	if templateutil.IsKnownSemanticallyEmptyMappings(prior) {
		return prior, diags
	}
	return esindex.NewMappingsNull(), diags
}

// flattenSettings maps the API settings response onto an IndexSettingsValue,
// applying the prior-preservation rule when the API returns no settings.
func flattenSettings(apiSettings map[string]any, prior customtypes.IndexSettingsValue) (customtypes.IndexSettingsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(apiSettings) > 0 {
		b, err := json.Marshal(apiSettings)
		if err != nil {
			diags.AddError("Failed to marshal template.settings", err.Error())
			return customtypes.NewIndexSettingsNull(), diags
		}
		return customtypes.NewIndexSettingsValue(string(b)), diags
	}
	if templateutil.IsKnownSemanticallyEmptySettings(prior) {
		return prior, diags
	}
	return customtypes.NewIndexSettingsNull(), diags
}

// extractEmptyObjectOverridesFromData pulls the prior mappings and settings
// values out of a prior Data so flattenTemplateBlock can preserve them when
// the API omits the corresponding field. Returns null values when the prior
// template block is absent or carries a non-string attribute shape.
func extractEmptyObjectOverridesFromData(prior Data) (esindex.MappingsValue, customtypes.IndexSettingsValue) {
	if prior.Template.IsNull() || prior.Template.IsUnknown() {
		return esindex.NewMappingsNull(), customtypes.NewIndexSettingsNull()
	}

	tplAttrs := prior.Template.Attributes()

	mappings := esindex.NewMappingsNull()
	if mv, ok := tplAttrs[attrMappings].(esindex.MappingsValue); ok {
		mappings = mv
	}

	settings := customtypes.NewIndexSettingsNull()
	if sv, ok := tplAttrs[attrSettings].(customtypes.IndexSettingsValue); ok {
		settings = sv
	}

	return mappings, settings
}

// extractAliasRoutingFromData extracts user-configured alias routing values from
// prior Data state so they can be preserved when the API omits them on read.
func extractAliasRoutingFromData(prior Data) map[string]string {
	result := make(map[string]string)
	if prior.Template.IsNull() || prior.Template.IsUnknown() {
		return result
	}

	tplAttrs := prior.Template.Attributes()
	aliasVal, ok := tplAttrs[attrAlias]
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
		nameAttr, ok := attrs[attrName].(types.String)
		if !ok || nameAttr.IsNull() || nameAttr.IsUnknown() {
			continue
		}
		routingAttr, ok := attrs[attrRouting].(types.String)
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
