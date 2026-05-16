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

type processorDateIndexNameModel struct {
	CommonProcessorModel
	Field           types.String `tfsdk:"field"`
	IndexNamePrefix types.String `tfsdk:"index_name_prefix"`
	DateRounding    types.String `tfsdk:"date_rounding"`
	DateFormats     types.List   `tfsdk:"date_formats"`
	Timezone        types.String `tfsdk:"timezone"`
	Locale          types.String `tfsdk:"locale"`
	IndexNameFormat types.String `tfsdk:"index_name_format"`
}

func (m *processorDateIndexNameModel) TypeName() string { return "date_index_name" }

func (m *processorDateIndexNameModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorDateIndexNameBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.IndexNamePrefix) {
		body.IndexNamePrefix = m.IndexNamePrefix.ValueString()
	}
	if IsKnown(m.DateRounding) {
		body.DateRounding = m.DateRounding.ValueString()
	}
	if IsKnown(m.DateFormats) {
		elems := make([]string, 0, len(m.DateFormats.Elements()))
		for _, elem := range m.DateFormats.Elements() {
			str, ok := elem.(types.String)
			if !ok || !IsKnown(str) {
				if !ok {
					diags.AddError("Invalid date_formats element type", "expected types.String")
				} else {
					diags.AddError("Unknown date_formats element", "date_formats elements cannot be unknown")
				}
				continue
			}
			elems = append(elems, str.ValueString())
		}
		body.DateFormats = elems
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
	if m.IndexNameFormat.IsNull() || m.IndexNameFormat.IsUnknown() {
		m.IndexNameFormat = types.StringValue("yyyy-MM-dd")
		body.IndexNameFormat = "yyyy-MM-dd"
	} else {
		body.IndexNameFormat = m.IndexNameFormat.ValueString()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorDateIndexNameDataSource returns a PF data source for the date_index_name processor.
func NewProcessorDateIndexNameDataSource() datasource.DataSource {
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
			Description: "The field to get the date or timestamp from.",
			Required:    true,
		},
		"index_name_prefix": schema.StringAttribute{
			Description: "A prefix of the index name to be prepended before the printed date.",
			Optional:    true,
		},
		"date_rounding": schema.StringAttribute{
			Description: "How to round the date when formatting the date into the index name.",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf("y", "M", "w", "d", "h", "m", "s")},
		},
		"date_formats": schema.ListAttribute{
			Description: "An array of the expected date formats for parsing dates / timestamps in the document being preprocessed.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"timezone": schema.StringAttribute{
			Description: "The timezone to use when parsing the date and when date math index supports resolves expressions into concrete index names.",
			Optional:    true,
			Computed:    true,
		},
		"locale": schema.StringAttribute{
			Description: "The locale to use when parsing the date from the document being preprocessed, relevant when parsing month names or week days.",
			Optional:    true,
			Computed:    true,
		},
		"index_name_format": schema.StringAttribute{
			Description: "The format to be used when printing the parsed date into the index name.",
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorDateIndexNameModel{}, schema.Schema{
		Description: processorDateIndexNameDataSourceDescription,
		Attributes:  attrs,
	})
}
