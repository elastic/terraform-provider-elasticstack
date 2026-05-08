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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// expandFromData builds a models.ComponentTemplate from a Data plan/state value.
func expandFromData(ctx context.Context, d Data) (models.ComponentTemplate, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := models.ComponentTemplate{
		Name: d.Name.ValueString(),
	}

	if !d.Metadata.IsNull() && !d.Metadata.IsUnknown() {
		s := strings.TrimSpace(d.Metadata.ValueString())
		if s != "" {
			meta := make(map[string]any)
			if err := json.Unmarshal([]byte(s), &meta); err != nil {
				diags.AddError("Invalid metadata JSON", err.Error())
				return out, diags
			}
			out.Meta = meta
		}
	}

	if !d.Template.IsNull() && !d.Template.IsUnknown() {
		tpl, d2 := expandTemplateBlock(ctx, d.Template)
		diags.Append(d2...)
		if diags.HasError() {
			return out, diags
		}
		out.Template = tpl
	}

	if !d.Version.IsNull() && !d.Version.IsUnknown() {
		v := d.Version.ValueInt64()
		out.Version = &v
	}

	return out, diags
}

// expandTemplateBlock expands the template block object to *models.Template.
func expandTemplateBlock(ctx context.Context, obj types.Object) (*models.Template, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	var tm TemplateModel
	diags.Append(obj.As(ctx, &tm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
	if diags.HasError() {
		return nil, diags
	}

	t := &models.Template{}

	if !tm.Mappings.IsNull() && !tm.Mappings.IsUnknown() {
		s := strings.TrimSpace(tm.Mappings.ValueString())
		if s != "" {
			maps := make(map[string]any)
			if err := json.Unmarshal([]byte(s), &maps); err != nil {
				diags.AddError("Invalid template.mappings JSON", err.Error())
				return nil, diags
			}
			t.Mappings = maps
		}
	}

	if !tm.Settings.IsNull() && !tm.Settings.IsUnknown() {
		s := strings.TrimSpace(tm.Settings.ValueString())
		if s != "" {
			sets := make(map[string]any)
			if err := json.Unmarshal([]byte(s), &sets); err != nil {
				diags.AddError("Invalid template.settings JSON", err.Error())
				return nil, diags
			}
			t.Settings = sets
		}
	}

	if !tm.Alias.IsNull() && !tm.Alias.IsUnknown() {
		aliases, d2 := expandAliasSet(ctx, tm.Alias)
		diags.Append(d2...)
		if diags.HasError() {
			return nil, diags
		}
		t.Aliases = aliases
	}

	return t, diags
}

// expandAliasSet expands a set of alias objects to map[string]models.IndexAlias.
func expandAliasSet(ctx context.Context, set types.Set) (map[string]models.IndexAlias, diag.Diagnostics) {
	var diags diag.Diagnostics
	if set.IsNull() || set.IsUnknown() {
		return nil, diags
	}

	var elems []AliasModel
	diags.Append(set.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return nil, diags
	}

	aliases := make(map[string]models.IndexAlias, len(elems))
	for _, am := range elems {
		ia, d := expandAliasElement(am)
		if d.HasError() {
			return nil, d
		}
		aliases[am.Name.ValueString()] = ia
	}
	return aliases, diags
}

// expandAliasElement converts a single AliasModel to a models.IndexAlias.
func expandAliasElement(am AliasModel) (models.IndexAlias, diag.Diagnostics) {
	var diags diag.Diagnostics
	ia := models.IndexAlias{Name: am.Name.ValueString()}

	if !am.Filter.IsNull() && !am.Filter.IsUnknown() {
		fs := strings.TrimSpace(am.Filter.ValueString())
		if fs != "" {
			filterMap := make(map[string]any)
			if err := json.Unmarshal([]byte(fs), &filterMap); err != nil {
				diags.AddError("Invalid alias filter JSON", err.Error())
				return ia, diags
			}
			ia.Filter = filterMap
		}
	}

	if !am.IndexRouting.IsNull() && !am.IndexRouting.IsUnknown() {
		ia.IndexRouting = am.IndexRouting.ValueString()
	}
	if !am.SearchRouting.IsNull() && !am.SearchRouting.IsUnknown() {
		ia.SearchRouting = am.SearchRouting.ValueString()
	}
	if !am.Routing.IsNull() && !am.Routing.IsUnknown() {
		ia.Routing = am.Routing.ValueString()
	}
	if !am.IsHidden.IsNull() && !am.IsHidden.IsUnknown() {
		ia.IsHidden = am.IsHidden.ValueBool()
	}
	if !am.IsWriteIndex.IsNull() && !am.IsWriteIndex.IsUnknown() {
		ia.IsWriteIndex = am.IsWriteIndex.ValueBool()
	}

	return ia, diags
}
