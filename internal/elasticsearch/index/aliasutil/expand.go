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
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AliasFields holds the fields shared by template.AliasElementModel and
// componenttemplate.AliasModel. Both structs expose the same seven fields with the
// same types, so callers can populate this struct directly before delegating to
// ExpandAliasFields.
type AliasFields struct {
	Name          types.String
	Filter        jsontypes.Normalized
	IndexRouting  types.String
	SearchRouting types.String
	Routing       types.String
	IsHidden      types.Bool
	IsWriteIndex  types.Bool
}

// ExpandAliasFields converts an AliasFields to a models.IndexAlias.
// Routing string fields (IndexRouting, SearchRouting, Routing) are only set when
// non-null and non-unknown, so the caller's zero value is not silently overwritten
// for optional fields that were not configured.
func ExpandAliasFields(f AliasFields) (models.IndexAlias, diag.Diagnostics) {
	var diags diag.Diagnostics
	ia := models.IndexAlias{Name: f.Name.ValueString()}

	if !f.Filter.IsNull() && !f.Filter.IsUnknown() {
		fs := strings.TrimSpace(f.Filter.ValueString())
		if fs != "" {
			filterMap := make(map[string]any)
			if err := json.Unmarshal([]byte(fs), &filterMap); err != nil {
				diags.AddError("Invalid alias filter JSON", err.Error())
				return ia, diags
			}
			ia.Filter = filterMap
		}
	}

	if !f.IndexRouting.IsNull() && !f.IndexRouting.IsUnknown() {
		ia.IndexRouting = f.IndexRouting.ValueString()
	}
	if !f.SearchRouting.IsNull() && !f.SearchRouting.IsUnknown() {
		ia.SearchRouting = f.SearchRouting.ValueString()
	}
	if !f.Routing.IsNull() && !f.Routing.IsUnknown() {
		ia.Routing = f.Routing.ValueString()
	}
	if !f.IsHidden.IsNull() && !f.IsHidden.IsUnknown() {
		ia.IsHidden = f.IsHidden.ValueBool()
	}
	if !f.IsWriteIndex.IsNull() && !f.IsWriteIndex.IsUnknown() {
		ia.IsWriteIndex = f.IsWriteIndex.ValueBool()
	}

	return ia, diags
}
