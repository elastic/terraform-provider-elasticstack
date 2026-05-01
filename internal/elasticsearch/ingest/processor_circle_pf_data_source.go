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

type processorCircleModel struct {
	CommonProcessorModel
	ID            types.String  `tfsdk:"id"`
	JSON          types.String  `tfsdk:"json"`
	Field         types.String  `tfsdk:"field"`
	TargetField   types.String  `tfsdk:"target_field"`
	IgnoreMissing types.Bool    `tfsdk:"ignore_missing"`
	ErrorDistance types.Float64 `tfsdk:"error_distance"`
	ShapeType     types.String  `tfsdk:"shape_type"`
}

func (m *processorCircleModel) TypeName() string    { return "circle" }
func (m *processorCircleModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorCircleModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorCircleModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorCircleBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}
	if m.IgnoreMissing.IsNull() || m.IgnoreMissing.IsUnknown() {
		m.IgnoreMissing = types.BoolValue(false)
		body.IgnoreMissing = false
	} else {
		body.IgnoreMissing = m.IgnoreMissing.ValueBool()
	}
	if IsKnown(m.ErrorDistance) {
		body.ErrorDistance = m.ErrorDistance.ValueFloat64()
	}
	if IsKnown(m.ShapeType) {
		body.ShapeType = m.ShapeType.ValueString()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorCircleDataSource returns a PF data source for the circle processor.
func NewProcessorCircleDataSource() datasource.DataSource {
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
			Description: "The field containing the circle geometry to convert.",
			Required:    true,
		},
		"target_field": schema.StringAttribute{
			Description: "The field to assign the converted value to, by default `field` is updated in-place",
			Optional:    true,
		},
		"ignore_missing": schema.BoolAttribute{
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
			Optional:    true,
			Computed:    true,
		},
		"error_distance": schema.Float64Attribute{
			Description: "The difference between the resulting inscribed distance from center to side and the circle's radius (measured in meters for `geo_shape`, unit-less for `shape`)",
			Required:    true,
		},
		"shape_type": schema.StringAttribute{
			Description: "Which field mapping type is to be used when processing the circle.",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf("geo_shape", "shape")},
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorCircleModel{}, schema.Schema{
		Description: processorCircleDataSourceDescription,
		Attributes:  attrs,
	})
}
