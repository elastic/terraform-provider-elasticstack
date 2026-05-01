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

type processorDateModel struct {
	CommonProcessorModel
	ID           types.String `tfsdk:"id"`
	JSON         types.String `tfsdk:"json"`
	Field        types.String `tfsdk:"field"`
	TargetField  types.String `tfsdk:"target_field"`
	Formats      types.List   `tfsdk:"formats"`
	Timezone     types.String `tfsdk:"timezone"`
	Locale       types.String `tfsdk:"locale"`
	OutputFormat types.String `tfsdk:"output_format"`
}

func (m *processorDateModel) TypeName() string    { return "date" }
func (m *processorDateModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorDateModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorDateModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorDateBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if m.TargetField.IsNull() || m.TargetField.IsUnknown() {
		m.TargetField = types.StringValue("@timestamp")
		body.TargetField = "@timestamp"
	} else {
		body.TargetField = m.TargetField.ValueString()
	}
	if IsKnown(m.Formats) {
		elems := make([]string, 0, len(m.Formats.Elements()))
		for _, elem := range m.Formats.Elements() {
			str, ok := elem.(types.String)
			if !ok || !IsKnown(str) {
				if !ok {
					diags.AddError("Invalid formats element type", "expected types.String")
				} else {
					diags.AddError("Unknown formats element", "formats elements cannot be unknown")
				}
				continue
			}
			elems = append(elems, str.ValueString())
		}
		body.Formats = elems
	}
	if m.Timezone.IsNull() || m.Timezone.IsUnknown() {
		m.Timezone = types.StringValue("UTC")
		body.Timezone = "UTC"
	} else {
		body.Timezone = m.Timezone.ValueString()
	}
	if m.Locale.IsNull() || m.Locale.IsUnknown() {
		m.Locale = types.StringValue("ENGLISH")
		body.Locale = "ENGLISH"
	} else {
		body.Locale = m.Locale.ValueString()
	}
	if m.OutputFormat.IsNull() || m.OutputFormat.IsUnknown() {
		m.OutputFormat = types.StringValue("yyyy-MM-dd'T'HH:mm:ss.SSSXXX")
		body.OutputFormat = "yyyy-MM-dd'T'HH:mm:ss.SSSXXX"
	} else {
		body.OutputFormat = m.OutputFormat.ValueString()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorDateDataSource returns a PF data source for the date processor.
func NewProcessorDateDataSource() datasource.DataSource {
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
			Description: "The field to get the date from.",
			Required:    true,
		},
		"target_field": schema.StringAttribute{
			Description: "The field that will hold the parsed date.",
			Optional:    true,
			Computed:    true,
		},
		"formats": schema.ListAttribute{
			Description: "An array of the expected date formats.",
			Required:    true,
			ElementType: types.StringType,
			Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
		},
		"timezone": schema.StringAttribute{
			Description: "The timezone to use when parsing the date.",
			Optional:    true,
			Computed:    true,
		},
		"locale": schema.StringAttribute{
			Description: "The locale to use when parsing the date, relevant when parsing month names or week days.",
			Optional:    true,
			Computed:    true,
		},
		"output_format": schema.StringAttribute{
			Description: "The format to use when writing the date to `target_field`.",
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorDateModel{}, schema.Schema{
		Description: processorDateDataSourceDescription,
		Attributes:  attrs,
	})
}
