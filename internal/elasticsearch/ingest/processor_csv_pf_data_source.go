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

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorCSVModel struct {
	CommonProcessorModel
	Field         types.String `tfsdk:"field"`
	TargetFields  types.List   `tfsdk:"target_fields"`
	IgnoreMissing types.Bool   `tfsdk:"ignore_missing"`
	Separator     types.String `tfsdk:"separator"`
	Quote         types.String `tfsdk:"quote"`
	Trim          types.Bool   `tfsdk:"trim"`
	EmptyValue    types.String `tfsdk:"empty_value"`
}

func (m *processorCSVModel) TypeName() string { return "csv" }

func (m *processorCSVModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorCSVBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.TargetFields) {
		elems := make([]string, 0, len(m.TargetFields.Elements()))
		for _, elem := range m.TargetFields.Elements() {
			str, ok := elem.(types.String)
			if !ok || !IsKnown(str) {
				if !ok {
					diags.AddError("Invalid target_fields element type", "expected types.String")
				} else {
					diags.AddError("Unknown target_fields element", "target_fields elements cannot be unknown")
				}
				continue
			}
			elems = append(elems, str.ValueString())
		}
		body.TargetFields = elems
	}
	if m.IgnoreMissing.IsNull() || m.IgnoreMissing.IsUnknown() {
		m.IgnoreMissing = types.BoolValue(false)
		body.IgnoreMissing = false
	} else {
		body.IgnoreMissing = m.IgnoreMissing.ValueBool()
	}
	if m.Separator.IsNull() || m.Separator.IsUnknown() {
		m.Separator = types.StringValue(",")
		body.Separator = ","
	} else {
		body.Separator = m.Separator.ValueString()
	}
	if m.Quote.IsNull() || m.Quote.IsUnknown() {
		m.Quote = types.StringValue("\"")
		body.Quote = "\""
	} else {
		body.Quote = m.Quote.ValueString()
	}
	if m.Trim.IsNull() || m.Trim.IsUnknown() {
		m.Trim = types.BoolValue(false)
		body.Trim = false
	} else {
		body.Trim = m.Trim.ValueBool()
	}
	if IsKnown(m.EmptyValue) {
		body.EmptyValue = m.EmptyValue.ValueString()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorCSVDataSource returns a PF data source for the csv processor.
func NewProcessorCSVDataSource() datasource.DataSource {
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
			Description: "The field to extract data from.",
			Required:    true,
		},
		"target_fields": schema.ListAttribute{
			Description: "The array of fields to assign extracted values to.",
			Required:    true,
			ElementType: types.StringType,
			Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
		},
		"ignore_missing": schema.BoolAttribute{
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
			Optional:    true,
			Computed:    true,
		},
		"separator": schema.StringAttribute{
			Description: "Separator used in CSV, has to be single character string.",
			Optional:    true,
			Computed:    true,
		},
		"quote": schema.StringAttribute{
			Description: "Quote used in CSV, has to be single character string",
			Optional:    true,
			Computed:    true,
		},
		"trim": schema.BoolAttribute{
			Description: "Trim whitespaces in unquoted fields.",
			Optional:    true,
			Computed:    true,
		},
		"empty_value": schema.StringAttribute{
			Description: "Value used to fill empty fields, empty fields will be skipped if this is not provided.",
			Optional:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorCSVModel{}, schema.Schema{
		Description: processorCSVDataSourceDescription,
		Attributes:  attrs,
	})
}
