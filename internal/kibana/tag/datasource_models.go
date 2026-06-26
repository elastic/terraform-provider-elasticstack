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

package tag

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tagsDataSourceModel struct {
	entitycore.KibanaConnectionField
	Query   types.String `tfsdk:"query"`
	SpaceID types.String `tfsdk:"space_id"`
	Tags    types.List   `tfsdk:"tags"`
}

type tagItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
	Managed     types.Bool   `tfsdk:"managed"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

var _ entitycore.WithVersionRequirements = (*tagsDataSourceModel)(nil)

func (tagsDataSourceModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *tagMinVersion,
			ErrorMessage: fmt.Sprintf("Kibana tags require Elastic Stack v%s or later (introduced in Kibana 9.5).", tagMinVersion),
		},
	}, nil
}

func tagItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":            types.StringType,
		attrName:        types.StringType,
		attrColor:       types.StringType,
		attrDescription: types.StringType,
		attrManaged:     types.BoolType,
		attrCreatedAt:   types.StringType,
		attrUpdatedAt:   types.StringType,
	}
}

func tagItemElemType() attr.Type {
	return types.ObjectType{AttrTypes: tagItemAttrTypes()}
}

func tagItemFromAPI(detail kibanaoapi.TagDetail) tagItemModel {
	item := tagItemModel{
		ID:    types.StringValue(detail.ID),
		Name:  types.StringValue(detail.Name),
		Color: types.StringValue(detail.Color),
	}
	item.Description = optionalStringPointerValue(detail.Description)
	item.Managed = types.BoolPointerValue(detail.Managed)
	item.CreatedAt = types.StringPointerValue(detail.CreatedAt)
	item.UpdatedAt = types.StringPointerValue(detail.UpdatedAt)
	return item
}

func (m *tagsDataSourceModel) setTags(ctx context.Context, tags []kibanaoapi.TagDetail) diag.Diagnostics {
	if len(tags) == 0 {
		m.Tags = types.ListValueMust(tagItemElemType(), []attr.Value{})
		return nil
	}

	elems := make([]attr.Value, 0, len(tags))
	for _, tag := range tags {
		obj, diags := types.ObjectValueFrom(ctx, tagItemAttrTypes(), tagItemFromAPI(tag))
		if diags.HasError() {
			return diags
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(tagItemElemType(), elems)
	if diags.HasError() {
		return diags
	}
	m.Tags = list
	return nil
}
