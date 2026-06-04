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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/templateutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// toAPIModel converts the Terraform model into an API index template body.
// allow_custom_routing is sent only when true so older Elasticsearch versions (pre-8.0) that do not
// know about the field do not reject the request. The 8.x update workaround in update.go handles
// the case where a previously-true value must be explicitly reset to false on the existing template.
func (m Model) toAPIModel(ctx context.Context) (*models.IndexTemplate, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := &models.IndexTemplate{
		Name: m.Name.ValueString(),
	}

	comps := make([]string, 0)
	if typeutils.IsKnown(m.ComposedOf) {
		diags.Append(m.ComposedOf.ElementsAs(ctx, &comps, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}
	out.ComposedOf = comps

	if typeutils.IsKnown(m.IgnoreMissingComponentTemplates) {
		var ignore []string
		diags.Append(m.IgnoreMissingComponentTemplates.ElementsAs(ctx, &ignore, false)...)
		if diags.HasError() {
			return nil, diags
		}
		out.IgnoreMissingComponentTemplates = ignore
	}

	if typeutils.IsKnown(m.DataStream) {
		out.DataStream = expandDataStreamBlock(m.DataStream)
	}

	if m.IndexPatterns.IsNull() || m.IndexPatterns.IsUnknown() {
		diags.AddError("Configuration error", "index_patterns must be set")
		return nil, diags
	}
	var patterns []string
	diags.Append(m.IndexPatterns.ElementsAs(ctx, &patterns, false)...)
	if diags.HasError() {
		return nil, diags
	}
	out.IndexPatterns = patterns

	if typeutils.IsKnown(m.Metadata) {
		s := strings.TrimSpace(m.Metadata.ValueString())
		if s != "" {
			meta := make(map[string]any)
			if err := json.Unmarshal([]byte(s), &meta); err != nil {
				diags.AddError("Invalid metadata JSON", err.Error())
				return nil, diags
			}
			out.Meta = meta
		}
	}

	if typeutils.IsKnown(m.Priority) {
		p := m.Priority.ValueInt64()
		out.Priority = &p
	}

	if typeutils.IsKnown(m.Template) {
		tpl, d := expandTemplateBlock(ctx, m.Template)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		out.Template = tpl
	}

	if typeutils.IsKnown(m.Version) {
		v := m.Version.ValueInt64()
		out.Version = &v
	}

	if typeutils.IsKnown(m.AllowAutoCreate) {
		out.AllowAutoCreate = m.AllowAutoCreate.ValueBoolPointer()
	}

	return out, diags
}

func expandDataStreamBlock(obj types.Object) *models.DataStreamSettings {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	attrs := obj.Attributes()
	dSettings := &models.DataStreamSettings{}
	if hidden, ok := attrs[attrHidden]; ok && !hidden.IsNull() && !hidden.IsUnknown() {
		if hv, ok := hidden.(types.Bool); ok {
			h := hv.ValueBool()
			dSettings.Hidden = &h
		}
	}
	if acr, ok := attrs[attrAllowCustomRouting]; ok && !acr.IsNull() && !acr.IsUnknown() {
		if av, ok := acr.(types.Bool); ok && av.ValueBool() {
			t := true
			dSettings.AllowCustomRouting = &t
		}
	}
	return dSettings
}

func expandTemplateBlock(ctx context.Context, obj types.Object) (*models.Template, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	var tm TemplateBlockModel
	diags.Append(obj.As(ctx, &tm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
	if diags.HasError() {
		return nil, diags
	}

	t, d := templateutil.ExpandTemplateCore(ctx, tm.Alias, tm.Mappings, tm.Settings, tm.DataStreamOptions)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	if !tm.Lifecycle.IsNull() && !tm.Lifecycle.IsUnknown() {
		t.Lifecycle = expandTemplateLifecycle(tm.Lifecycle)
	}

	return t, diags
}

func expandTemplateLifecycle(obj types.Object) *models.LifecycleSettings {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	attrs := obj.Attributes()
	drAttr, ok := attrs[attrDataRetention]
	if !ok || drAttr.IsNull() || drAttr.IsUnknown() {
		return nil
	}
	drStr, ok := drAttr.(types.String)
	if !ok {
		return nil
	}
	return &models.LifecycleSettings{DataRetention: drStr.ValueString()}
}
