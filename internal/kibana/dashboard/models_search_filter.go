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

package dashboard

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type searchFilterModel struct {
	Query    types.String         `tfsdk:"query"`
	MetaJSON jsontypes.Normalized `tfsdk:"meta_json"`
	Language types.String         `tfsdk:"language"`
}

func (m *searchFilterModel) fromAPI(apiFilter kbapi.SearchFilterSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try to extract from SearchFilterSchema0
	filterSchema, err := apiFilter.AsSearchFilterSchema0()
	if err != nil {
		diags.AddError("Failed to extract search filter", err.Error())
		return diags
	}

	// Extract string from union type
	queryStr, queryErr := filterSchema.Query.AsSearchFilterSchema0Query0()
	if queryErr != nil {
		diags.AddError("Failed to extract search filter query", queryErr.Error())
		return diags
	}

	m.Query = types.StringValue(queryStr)

	// Language defaults to "kuery" if the API doesn't return it
	// This is consistent with Kibana's default behavior
	if filterSchema.Language != nil {
		m.Language = types.StringValue(string(*filterSchema.Language))
	} else {
		m.Language = types.StringValue("kuery")
	}

	if filterSchema.Meta != nil {
		metaJSON, err := json.Marshal(filterSchema.Meta)
		if err == nil {
			m.MetaJSON = jsontypes.NewNormalizedValue(string(metaJSON))
		}
	}

	return diags
}

func (m *searchFilterModel) toAPI() (kbapi.SearchFilterSchema, diag.Diagnostics) {
	var diags diag.Diagnostics

	filter := kbapi.SearchFilterSchema0{}
	if typeutils.IsKnown(m.Query) {
		query := m.Query.ValueString()
		var queryUnion kbapi.SearchFilterSchema_0_Query
		if err := queryUnion.FromSearchFilterSchema0Query0(query); err != nil {
			diags.AddError("Failed to create search filter query", err.Error())
			return kbapi.SearchFilterSchema{}, diags
		}
		filter.Query = queryUnion
	}
	if typeutils.IsKnown(m.Language) {
		lang := kbapi.SearchFilterSchema0Language(m.Language.ValueString())
		filter.Language = &lang
	}
	if typeutils.IsKnown(m.MetaJSON) {
		var meta map[string]any
		diags.Append(m.MetaJSON.Unmarshal(&meta)...)
		if !diags.HasError() {
			filter.Meta = &meta
		}
	}

	var result kbapi.SearchFilterSchema
	if err := result.FromSearchFilterSchema0(filter); err != nil {
		diags.AddError("Failed to create search filter", err.Error())
	}
	return result, diags
}
