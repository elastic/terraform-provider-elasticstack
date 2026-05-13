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

package filter

import (
	"context"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TFModel struct {
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	FilterID                types.String `tfsdk:"filter_id"`
	Description             types.String `tfsdk:"description"`
	Items                   types.Set    `tfsdk:"items"`
}

func (m TFModel) GetID() types.String { return m.ID }

func (m TFModel) GetResourceID() types.String { return m.FilterID }

func (m TFModel) GetElasticsearchConnection() types.List { return m.ElasticsearchConnection }

type UpdateAPIModel struct {
	Description *string  `json:"description,omitempty"`
	AddItems    []string `json:"add_items,omitempty"`
	RemoveItems []string `json:"remove_items,omitempty"`
}

func descriptionFromMLFilter(f *estypes.MLFilter) string {
	if f == nil || f.Description == nil {
		return ""
	}
	return *f.Description
}

// fromMLFilter maps an Elasticsearch MLFilter into Terraform state.
func (m *TFModel) fromMLFilter(ctx context.Context, f *estypes.MLFilter) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	if f == nil {
		return diags
	}

	m.FilterID = types.StringValue(f.FilterId)

	desc := descriptionFromMLFilter(f)
	if desc != "" {
		m.Description = types.StringValue(desc)
	} else {
		m.Description = types.StringNull()
	}

	if len(f.Items) == 0 && m.Items.IsNull() {
		return diags
	}

	if len(f.Items) == 0 {
		emptySet, d := types.SetValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		m.Items = emptySet
	} else {
		itemsSet, d := types.SetValueFrom(ctx, types.StringType, f.Items)
		diags.Append(d...)
		m.Items = itemsSet
	}

	return diags
}
