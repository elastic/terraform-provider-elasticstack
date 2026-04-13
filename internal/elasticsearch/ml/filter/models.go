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

type CreateAPIModel struct {
	Description string   `json:"description,omitempty"`
	Items       []string `json:"items,omitempty"`
}

type APIModel struct {
	FilterID    string   `json:"filter_id"`
	Description string   `json:"description,omitempty"`
	Items       []string `json:"items"`
}

type UpdateAPIModel struct {
	Description *string  `json:"description,omitempty"`
	AddItems    []string `json:"add_items,omitempty"`
	RemoveItems []string `json:"remove_items,omitempty"`
}

func (m *TFModel) toAPICreateModel(ctx context.Context) (*CreateAPIModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	apiModel := &CreateAPIModel{
		Description: m.Description.ValueString(),
	}

	if !m.Items.IsNull() && !m.Items.IsUnknown() {
		var items []string
		d := m.Items.ElementsAs(ctx, &items, false)
		diags.Append(d...)
		apiModel.Items = items
	}

	return apiModel, diags
}

func (m *TFModel) fromAPIModel(ctx context.Context, apiModel *APIModel) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	m.FilterID = types.StringValue(apiModel.FilterID)

	if apiModel.Description != "" {
		m.Description = types.StringValue(apiModel.Description)
	} else {
		m.Description = types.StringNull()
	}

	if len(apiModel.Items) == 0 && m.Items.IsNull() {
		return diags
	}

	if len(apiModel.Items) == 0 {
		emptySet, d := types.SetValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		m.Items = emptySet
	} else {
		itemsSet, d := types.SetValueFrom(ctx, types.StringType, apiModel.Items)
		diags.Append(d...)
		m.Items = itemsSet
	}

	return diags
}
