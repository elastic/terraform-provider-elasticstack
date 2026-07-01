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

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/templateutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fromAPIModel maps an API index template into this model.
// It does not set id or elasticsearch_connection; the caller merges those as needed.
// Alias routing echo shapes vs practitioner config are aligned in aliasutil.ApplyTemplateAliasReconciliationFromReference
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
		lv, d := types.ListValueFrom(ctx, types.StringType, composedOf)
		diags.Append(d...)
		m.ComposedOf = lv
	}

	ignoreMissing := in.IgnoreMissingComponentTemplates
	if ignoreMissing == nil {
		ignoreMissing = []string{}
	}
	{
		lv, d := types.ListValueFrom(ctx, types.StringType, ignoreMissing)
		diags.Append(d...)
		m.IgnoreMissingComponentTemplates = lv
	}

	indexPatterns := in.IndexPatterns
	if indexPatterns == nil {
		indexPatterns = []string{}
	}
	{
		sv, d := types.SetValueFrom(ctx, types.StringType, indexPatterns)
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

	m.Priority = types.Int64PointerValue(in.Priority)
	m.Version = types.Int64PointerValue(in.Version)
	m.AllowAutoCreate = types.BoolPointerValue(in.AllowAutoCreate)

	var d diag.Diagnostics
	m.DataStream, d = flattenDataStream(in.DataStream)
	diags.Append(d...)

	m.Template, d = flattenTemplateBody(ctx, in.Template)
	diags.Append(d...)

	return diags
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

	core, d := templateutil.FlattenTemplateCore(
		ctx,
		t,
		index.NewMappingsNull(),
		customtypes.NewIndexSettingsNull(),
		nil,
		aliasutil.NewAliasObjectType(),
		aliasutil.AliasAttributeTypes(),
	)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectUnknown(TemplateAttrTypes()), diags
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
		var lc diag.Diagnostics
		lcObj, lc = types.ObjectValue(LifecycleAttrTypes(), lcAttrs)
		diags.Append(lc...)
		if diags.HasError() {
			return types.ObjectUnknown(TemplateAttrTypes()), diags
		}
	} else {
		lcObj = types.ObjectNull(LifecycleAttrTypes())
	}

	tplAttrs := map[string]attr.Value{
		attrAlias:             core.AliasSet,
		attrMappings:          core.Mappings,
		attrSettings:          core.Settings,
		attrLifecycle:         lcObj,
		attrDataStreamOptions: core.DsoObj,
	}
	obj, d := types.ObjectValue(TemplateAttrTypes(), tplAttrs)
	diags.Append(d...)
	return obj, diags
}
