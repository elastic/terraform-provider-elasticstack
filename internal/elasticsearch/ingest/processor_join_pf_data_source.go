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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorJoinModel struct {
	CommonProcessorModel
	ID          types.String `tfsdk:"id"`
	JSON        types.String `tfsdk:"json"`
	Field       types.String `tfsdk:"field"`
	Separator   types.String `tfsdk:"separator"`
	TargetField types.String `tfsdk:"target_field"`
}

func (m *processorJoinModel) TypeName() string    { return "join" }
func (m *processorJoinModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorJoinModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorJoinModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorJoinBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.Separator) {
		body.Separator = m.Separator.ValueString()
	}
	if IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorJoinDataSource returns a PF data source for the join processor.
func NewProcessorJoinDataSource() datasource.DataSource {
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
			Description: "Field containing array values to join.",
			Required:    true,
		},
		"separator": schema.StringAttribute{
			Description: "The separator character.",
			Required:    true,
		},
		"target_field": schema.StringAttribute{
			Description: "The field to assign the converted value to, by default `field` is updated in-place.",
			Optional:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorJoinModel{}, schema.Schema{
		Description: processorJoinDataSourceDescription,
		Attributes:  attrs,
	})
}
