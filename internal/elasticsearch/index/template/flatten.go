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
	"encoding/json"
	"sort"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fromAPIModel maps an API index template into this model.
// It does not set id or elasticsearch_connection; the caller merges those as needed.
// Alias routing echo shapes vs practitioner config are aligned in applyTemplateAliasReconciliationFromReference
// after read (managed resource only); the data source preserves the API shape.
//
// The input is the locally-defined models.IndexTemplate populated by the raw GET
// response decoder, not the typed go-elasticsearch struct: see GetIndexTemplate
// and issue #3124 for why typed decoding is unsafe for free-form settings.
func (m *Model) fromAPIModel(ctx context.Context, name string, in *models.IndexTemplate) diag.Diagnostics {
	var diags diag.Diagnostics
	*m = Model{
		Name: types.StringValue(name),
	}
	if in == nil {
		return diags
	}

	composedOf := in.ComposedOf
	if composedOf == nil {
		composedOf = []string{}
	}
	{
		vals := stringSliceToAttrValues(composedOf)
		lv, d := types.ListValueFrom(ctx, types.StringType, vals)
		diags.Append(d...)
		m.ComposedOf = lv
	}

	ignoreMissing := in.IgnoreMissingComponentTemplates
	if ignoreMissing == nil {
		ignoreMissing = []string{}
	}
	{
		vals := stringSliceToAttrValues(ignoreMissing)
		lv, d := types.ListValueFrom(ctx, types.StringType, vals)
		diags.Append(d...)
		m.IgnoreMissingComponentTemplates = lv
	}

	indexPatterns := in.IndexPatterns
	if indexPatterns == nil {
		indexPatterns = []string{}
	}
	{
		vals := stringSliceToAttrValues(indexPatterns)
		sv, d := types.SetValueFrom(ctx, types.StringType, vals)
		diags.Append(d...)
		m.IndexPatterns = sv
	}

	if in.Meta != nil {
		b, err := json.Marshal(in.Meta)
		if err != nil {
			diags.AddError("Failed to marshal metadata", err.Error())
			return diags
		}
		m.Metadata = jsontypes.NewNormalizedValue(string(b))
	} else {
		m.Metadata = jsontypes.NewNormalizedNull()
	}

	m.Priority = int64FromInt64Ptr(in.Priority)
	m.Version = int64FromInt64Ptr(in.Version)
	m.AllowAutoCreate = boolFromBoolPtr(in.AllowAutoCreate)

	var d diag.Diagnostics
	m.DataStream, d = flattenDataStream(in.DataStream)
	diags.Append(d...)

	m.Template, d = flattenTemplateBody(ctx, in.Template)
	diags.Append(d...)

	return diags
}

func stringSliceToAttrValues(elems []string) []attr.Value {
	vals := make([]attr.Value, len(elems))
	for i, s := range elems {
		vals[i] = types.StringValue(s)
	}
	return vals
}

func int64FromInt64Ptr(p *int64) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*p)
}

func boolFromBoolPtr(p *bool) types.Bool {
	if p == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*p)
}

func flattenDataStream(ds *models.DataStreamSettings) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ds == nil {
		return types.ObjectNull(DataStreamAttrTypes()), diags
	}
	// Both attributes carry a schema-level default of false, so when ES does not
	// echo them back (e.g. older versions or omitted fields), state must mirror
	// the planned default rather than null to satisfy the Computed contract.
	attrs := map[string]attr.Value{
		attrHidden:             types.BoolValue(false),
		attrAllowCustomRouting: types.BoolValue(false),
	}
	if ds.Hidden != nil {
		attrs[attrHidden] = types.BoolValue(*ds.Hidden)
	}
	if ds.AllowCustomRouting != nil {
		attrs[attrAllowCustomRouting] = types.BoolValue(*ds.AllowCustomRouting)
	}
	obj, d := types.ObjectValue(DataStreamAttrTypes(), attrs)
	diags.Append(d...)
	return obj, diags
}

func flattenTemplateBody(ctx context.Context, t *models.Template) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if t == nil {
		return types.ObjectNull(TemplateAttrTypes()), diags
	}

	var aliasSet types.Set
	if t.Aliases == nil {
		aliasSet = types.SetNull(NewAliasObjectType())
	} else {
		names := make([]string, 0, len(t.Aliases))
		for k := range t.Aliases {
			names = append(names, k)
		}
		sort.Strings(names)
		vals := make([]attr.Value, 0, len(names))
		for _, name := range names {
			alias := t.Aliases[name]
			av, d := flattenAliasElement(name, alias)
			diags.Append(d...)
			if diags.HasError() {
				return types.ObjectUnknown(TemplateAttrTypes()), diags
			}
			vals = append(vals, av)
		}
		sv, d := types.SetValueFrom(ctx, NewAliasObjectType(), vals)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectUnknown(TemplateAttrTypes()), diags
		}
		aliasSet = sv
	}

	var mappings index.MappingsValue
	if t.Mappings != nil {
		b, err := json.Marshal(t.Mappings)
		if err != nil {
			diags.AddError("Failed to marshal template.mappings", err.Error())
			return types.ObjectUnknown(TemplateAttrTypes()), diags
		}
		mappings = index.NewMappingsValue(string(b))
	} else {
		mappings = index.NewMappingsNull()
	}

	var settings customtypes.IndexSettingsValue
	if t.Settings != nil {
		b, err := json.Marshal(t.Settings)
		if err != nil {
			diags.AddError("Failed to marshal template.settings", err.Error())
			return types.ObjectUnknown(TemplateAttrTypes()), diags
		}
		settings = customtypes.NewIndexSettingsValue(string(b))
	} else {
		settings = customtypes.NewIndexSettingsNull()
	}

	var lcObj types.Object
	if t.Lifecycle != nil {
		dataRetention := types.StringNull()
		if t.Lifecycle.DataRetention != "" {
			dataRetention = types.StringValue(t.Lifecycle.DataRetention)
		}
		lcAttrs := map[string]attr.Value{
			attrDataRetention: dataRetention,
		}
		var d diag.Diagnostics
		lcObj, d = types.ObjectValue(LifecycleAttrTypes(), lcAttrs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectUnknown(TemplateAttrTypes()), diags
		}
	} else {
		lcObj = types.ObjectNull(LifecycleAttrTypes())
	}

	var dsoObj types.Object
	if t.DataStreamOptions != nil && t.DataStreamOptions.FailureStore != nil {
		var dsoDiags diag.Diagnostics
		dsoObj, dsoDiags = datastreamoptions.FlattenLocal(t.DataStreamOptions)
		diags.Append(dsoDiags...)
		if diags.HasError() {
			return types.ObjectUnknown(TemplateAttrTypes()), diags
		}
	} else {
		dsoObj = types.ObjectNull(datastreamoptions.AttrTypes())
	}

	tplAttrs := map[string]attr.Value{
		attrAlias:             aliasSet,
		attrMappings:          mappings,
		attrSettings:          settings,
		attrLifecycle:         lcObj,
		attrDataStreamOptions: dsoObj,
	}
	obj, d := types.ObjectValue(TemplateAttrTypes(), tplAttrs)
	diags.Append(d...)
	return obj, diags
}

func flattenAliasElement(name string, a models.IndexAlias) (attr.Value, diag.Diagnostics) {
	attrs, diags := aliasutil.AliasAttrsFromModel(name, a)
	if diags.HasError() {
		return nil, diags
	}
	aliasObj, d := NewAliasObjectValue(attrs)
	diags.Append(d...)
	return aliasObj, diags
}
