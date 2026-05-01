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

package ingest

import (
	"maps"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorSortModel struct {
	CommonProcessorModel
	ID          types.String `tfsdk:"id"`
	JSON        types.String `tfsdk:"json"`
	Field       types.String `tfsdk:"field"`
	Order       types.String `tfsdk:"order"`
	TargetField types.String `tfsdk:"target_field"`
}

func (m *processorSortModel) TypeName() string    { return "sort" }
func (m *processorSortModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorSortModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorSortModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorSortBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if m.Order.IsNull() || m.Order.IsUnknown() {
		m.Order = types.StringValue("asc")
		body.Order = "asc"
	} else {
		body.Order = m.Order.ValueString()
	}
	if IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorSortDataSource returns a PF data source for the sort processor.
func NewProcessorSortDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource",
			Computed:    true,
		},
		"json": schema.StringAttribute{
			Description: "JSON representation of this data source.",
			Computed:    true,
		},
		"field": schema.StringAttribute{
			Description: "The field to be sorted",
			Required:    true,
		},
		"order": schema.StringAttribute{
			Description: "The sort order to use. Accepts `asc` or `desc`.",
			Optional:    true,
			Computed:    true,
			Validators:  []validator.String{stringvalidator.OneOf("asc", "desc")},
		},
		"target_field": schema.StringAttribute{
			Description: "The field to assign the sorted value to, by default `field` is updated in-place",
			Optional:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorSortModel{}, schema.Schema{
		Description: processorSortDataSourceDescription,
		Attributes:  attrs,
	})
}
