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

	core, d := templateutil.FlattenTemplateCore(
		ctx,
		t,
		priorMappings,
		priorSettings,
		preservedRouting,
		aliasutil.NewAliasObjectType(),
		aliasutil.AliasAttributeTypes(),
	)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(templateAttrTypes()), diags
	}

	tplAttrs := map[string]attr.Value{
		attrAlias:             core.AliasSet,
		attrMappings:          core.Mappings,
		attrSettings:          core.Settings,
		attrDataStreamOptions: core.DsoObj,
	}

	obj, d := types.ObjectValue(templateAttrTypes(), tplAttrs)
	diags.Append(d...)
	return obj, diags
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
		var attrs map[string]attr.Value
		switch obj := el.(type) {
		case aliasutil.AliasObjectValue:
			attrs = obj.Attributes()
		case types.Object:
			attrs = obj.Attributes()
		default:
			continue
		}
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
