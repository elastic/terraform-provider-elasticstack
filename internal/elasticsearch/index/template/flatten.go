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

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fromAPIModel maps an API index template into this model.
// It does not set id or elasticsearch_connection; the caller merges those as needed.
// Alias routing echo shapes vs practitioner config are aligned in applyTemplateAliasReconciliationFromReference
// after read (managed resource only); the data source preserves the API shape.
func (m *Model) fromAPIModel(ctx context.Context, name string, in *estypes.IndexTemplate) diag.Diagnostics {
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

	if in.Meta_ != nil {
		b, err := json.Marshal(in.Meta_)
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

func flattenDataStream(ds *estypes.IndexTemplateDataStreamConfiguration) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ds == nil {
		return types.ObjectNull(DataStreamAttrTypes()), diags
	}
	// Both attributes carry a schema-level default of false, so when ES does not
	// echo them back (e.g. older versions or omitted fields), state must mirror
	// the planned default rather than null to satisfy the Computed contract.
	attrs := map[string]attr.Value{
		"hidden":               types.BoolValue(false),
		"allow_custom_routing": types.BoolValue(false),
	}
	if ds.Hidden != nil {
		attrs["hidden"] = types.BoolValue(*ds.Hidden)
	}
	if ds.AllowCustomRouting != nil {
		attrs["allow_custom_routing"] = types.BoolValue(*ds.AllowCustomRouting)
	}
	obj, d := types.ObjectValue(DataStreamAttrTypes(), attrs)
	diags.Append(d...)
	return obj, diags
}

func flattenTemplateBody(ctx context.Context, t *estypes.IndexTemplateSummary) (types.Object, diag.Diagnostics) {
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
		if t.Lifecycle.DataRetention != nil {
			if dr, ok := t.Lifecycle.DataRetention.(string); ok && dr != "" {
				dataRetention = types.StringValue(dr)
			}
		}
		lcAttrs := map[string]attr.Value{
			"data_retention": dataRetention,
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
		dsoObj, dsoDiags = flattenDataStreamOptions(t.DataStreamOptions)
		diags.Append(dsoDiags...)
		if diags.HasError() {
			return types.ObjectUnknown(TemplateAttrTypes()), diags
		}
	} else {
		dsoObj = types.ObjectNull(DataStreamOptionsAttrTypes())
	}

	tplAttrs := map[string]attr.Value{
		"alias":               aliasSet,
		"mappings":            mappings,
		"settings":            settings,
		"lifecycle":           lcObj,
		"data_stream_options": dsoObj,
	}
	obj, d := types.ObjectValue(TemplateAttrTypes(), tplAttrs)
	diags.Append(d...)
	return obj, diags
}

func flattenDataStreamOptions(dso *estypes.DataStreamOptions) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	fs := dso.FailureStore
	fsAttrs := map[string]attr.Value{
		"enabled":   types.BoolValue(typeutils.Deref(fs.Enabled)),
		"lifecycle": types.ObjectNull(FailureStoreLifecycleAttrTypes()),
	}
	if fs.Lifecycle != nil {
		dataRetention := types.StringNull()
		if fs.Lifecycle.DataRetention != nil {
			if dr, ok := fs.Lifecycle.DataRetention.(string); ok && dr != "" {
				dataRetention = types.StringValue(dr)
			}
		}
		lcAttrs := map[string]attr.Value{
			"data_retention": dataRetention,
		}
		lcObj, d := types.ObjectValue(FailureStoreLifecycleAttrTypes(), lcAttrs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectUnknown(DataStreamOptionsAttrTypes()), diags
		}
		fsAttrs["lifecycle"] = lcObj
	}
	fsObj, d := types.ObjectValue(FailureStoreAttrTypes(), fsAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectUnknown(DataStreamOptionsAttrTypes()), diags
	}
	outer := map[string]attr.Value{
		"failure_store": fsObj,
	}
	obj, d := types.ObjectValue(DataStreamOptionsAttrTypes(), outer)
	diags.Append(d...)
	return obj, diags
}

func flattenAliasElement(name string, a estypes.Alias) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := map[string]attr.Value{
		"name":           types.StringValue(name),
		"index_routing":  types.StringValue(typeutils.Deref(a.IndexRouting)),
		"routing":        types.StringValue(typeutils.Deref(a.Routing)),
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
		var filterMap map[string]any
		if err := json.Unmarshal(b, &filterMap); err != nil {
			diags.AddError("Failed to unmarshal alias filter", err.Error())
			return nil, diags
		}
		normalized := elasticsearch.NormalizeQueryFilter(filterMap)
		if nm, ok := normalized.(map[string]any); ok {
			filterMap = nm
		}
		normalizedBytes, _ := json.Marshal(filterMap)
		attrs["filter"] = jsontypes.NewNormalizedValue(string(normalizedBytes))
	} else {
		attrs["filter"] = jsontypes.NewNormalizedNull()
	}
	aliasObj, d := NewAliasObjectValue(attrs)
	diags.Append(d...)
	return aliasObj, diags
}
