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
	"fmt"
	"strings"

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
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
	if !m.ComposedOf.IsNull() && !m.ComposedOf.IsUnknown() {
		diags.Append(m.ComposedOf.ElementsAs(ctx, &comps, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}
	out.ComposedOf = comps

	if !m.IgnoreMissingComponentTemplates.IsNull() && !m.IgnoreMissingComponentTemplates.IsUnknown() {
		var ignore []string
		diags.Append(m.IgnoreMissingComponentTemplates.ElementsAs(ctx, &ignore, false)...)
		if diags.HasError() {
			return nil, diags
		}
		out.IgnoreMissingComponentTemplates = ignore
	}

	if !m.DataStream.IsNull() && !m.DataStream.IsUnknown() {
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

	if !m.Metadata.IsNull() && !m.Metadata.IsUnknown() {
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

	if !m.Priority.IsNull() && !m.Priority.IsUnknown() {
		p := m.Priority.ValueInt64()
		out.Priority = &p
	}

	if !m.Template.IsNull() && !m.Template.IsUnknown() {
		tpl, d := expandTemplateBlock(ctx, m.Template)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		out.Template = tpl
	}

	if !m.Version.IsNull() && !m.Version.IsUnknown() {
		v := m.Version.ValueInt64()
		out.Version = &v
	}

	return out, diags
}

func expandDataStreamBlock(obj types.Object) *models.DataStreamSettings {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	attrs := obj.Attributes()
	dSettings := &models.DataStreamSettings{}
	if hidden, ok := attrs["hidden"]; ok && !hidden.IsNull() && !hidden.IsUnknown() {
		if hv, ok := hidden.(types.Bool); ok {
			h := hv.ValueBool()
			dSettings.Hidden = &h
		}
	}
	if acr, ok := attrs["allow_custom_routing"]; ok && !acr.IsNull() && !acr.IsUnknown() {
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
	attrs := obj.Attributes()
	t := &models.Template{}

	if v, ok := attrs["alias"]; ok && !v.IsNull() && !v.IsUnknown() {
		setV, ok := v.(types.Set)
		if !ok {
			diags.AddError("Internal error", fmt.Sprintf("expected Set for template.alias, got %T", v))
			return nil, diags
		}
		t.Aliases = make(map[string]models.IndexAlias, len(setV.Elements()))
		for _, el := range setV.Elements() {
			aliasVal, ok := el.(AliasObjectValue)
			if !ok {
				diags.AddError("Internal error", fmt.Sprintf("expected AliasObjectValue, got %T", el))
				return nil, diags
			}
			var am AliasElementModel
			diags.Append(aliasVal.As(ctx, &am, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
			if diags.HasError() {
				return nil, diags
			}
			name := am.Name.ValueString()
			ia, d := expandAliasElement(am)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			t.Aliases[name] = ia
		}
	}

	if v, ok := attrs["mappings"]; ok && !v.IsNull() && !v.IsUnknown() {
		norm, ok := v.(esindex.MappingsValue)
		if !ok {
			diags.AddError("Internal error", fmt.Sprintf("expected index.MappingsValue for mappings, got %T", v))
			return nil, diags
		}
		s := strings.TrimSpace(norm.ValueString())
		if s != "" {
			maps := make(map[string]any)
			if err := json.Unmarshal([]byte(s), &maps); err != nil {
				diags.AddError("Invalid template.mappings JSON", err.Error())
				return nil, diags
			}
			t.Mappings = maps
		}
	}

	if v, ok := attrs["settings"]; ok && !v.IsNull() && !v.IsUnknown() {
		is, ok := v.(customtypes.IndexSettingsValue)
		if !ok {
			diags.AddError("Internal error", fmt.Sprintf("expected IndexSettingsValue for settings, got %T", v))
			return nil, diags
		}
		s := strings.TrimSpace(is.ValueString())
		if s != "" {
			sets := make(map[string]any)
			if err := json.Unmarshal([]byte(s), &sets); err != nil {
				diags.AddError("Invalid template.settings JSON", err.Error())
				return nil, diags
			}
			t.Settings = sets
		}
	}

	if v, ok := attrs["lifecycle"]; ok && !v.IsNull() && !v.IsUnknown() {
		lcObj, ok := v.(types.Object)
		if !ok {
			diags.AddError("Internal error", fmt.Sprintf("expected Object for lifecycle, got %T", v))
			return nil, diags
		}
		t.Lifecycle = expandTemplateLifecycle(lcObj)
	}

	if v, ok := attrs["data_stream_options"]; ok && !v.IsNull() && !v.IsUnknown() {
		dsoObj, ok := v.(types.Object)
		if !ok {
			diags.AddError("Internal error", fmt.Sprintf("expected Object for data_stream_options, got %T", v))
			return nil, diags
		}
		dso, d := expandDataStreamOptions(dsoObj)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		t.DataStreamOptions = dso
	}

	return t, diags
}

func expandAliasElement(e AliasElementModel) (models.IndexAlias, diag.Diagnostics) {
	var diags diag.Diagnostics
	ia := models.IndexAlias{Name: e.Name.ValueString()}

	if !e.Filter.IsNull() && !e.Filter.IsUnknown() {
		fs := strings.TrimSpace(e.Filter.ValueString())
		if fs != "" {
			filterMap := make(map[string]any)
			if err := json.Unmarshal([]byte(fs), &filterMap); err != nil {
				diags.AddError("Invalid alias filter JSON", err.Error())
				return ia, diags
			}
			ia.Filter = filterMap
		}
	}

	ia.IndexRouting = tfStringValue(e.IndexRouting)
	ia.SearchRouting = tfStringValue(e.SearchRouting)
	ia.Routing = tfStringValue(e.Routing)

	if !e.IsHidden.IsNull() && !e.IsHidden.IsUnknown() {
		ia.IsHidden = e.IsHidden.ValueBool()
	}
	if !e.IsWriteIndex.IsNull() && !e.IsWriteIndex.IsUnknown() {
		ia.IsWriteIndex = e.IsWriteIndex.ValueBool()
	}
	return ia, diags
}

func tfStringValue(s types.String) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

func expandTemplateLifecycle(obj types.Object) *models.LifecycleSettings {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	attrs := obj.Attributes()
	drAttr, ok := attrs["data_retention"]
	if !ok || drAttr.IsNull() || drAttr.IsUnknown() {
		return nil
	}
	drStr, ok := drAttr.(types.String)
	if !ok {
		return nil
	}
	return &models.LifecycleSettings{DataRetention: drStr.ValueString()}
}

func expandDataStreamOptions(obj types.Object) (*models.DataStreamOptions, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}
	attrs := obj.Attributes()
	fsVal, ok := attrs["failure_store"]
	if !ok || fsVal.IsNull() || fsVal.IsUnknown() {
		return nil, diags
	}
	fsObj, ok := fsVal.(types.Object)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected Object for failure_store, got %T", fsVal))
		return nil, diags
	}
	fsAttrs := fsObj.Attributes()
	out := &models.DataStreamOptions{
		FailureStore: &models.FailureStoreOptions{},
	}
	if en, ok := fsAttrs["enabled"]; ok && !en.IsNull() && !en.IsUnknown() {
		if b, ok := en.(types.Bool); ok {
			out.FailureStore.Enabled = b.ValueBool()
		}
	}
	if lcVal, ok := fsAttrs["lifecycle"]; ok && !lcVal.IsNull() && !lcVal.IsUnknown() {
		lcObj, ok := lcVal.(types.Object)
		if !ok {
			diags.AddError("Internal error", fmt.Sprintf("expected Object for failure_store.lifecycle, got %T", lcVal))
			return nil, diags
		}
		lcAttrs := lcObj.Attributes()
		if drAttr, ok := lcAttrs["data_retention"]; ok && !drAttr.IsNull() && !drAttr.IsUnknown() {
			if drStr, ok := drAttr.(types.String); ok {
				out.FailureStore.Lifecycle = &models.FailureStoreLifecycle{
					DataRetention: drStr.ValueString(),
				}
			}
		}
	}
	return out, diags
}
