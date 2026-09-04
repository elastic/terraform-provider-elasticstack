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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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

	if typeutils.IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	body.TargetFields = typeutils.StringElements(m.TargetFields, &diags)
	body.IgnoreMissing = typeutils.BoolDefault(&m.IgnoreMissing, false)
	body.Separator = typeutils.StringDefault(&m.Separator, ",")
	body.Quote = typeutils.StringDefault(&m.Quote, "\"")
	body.Trim = typeutils.BoolDefault(&m.Trim, false)
	if typeutils.IsKnown(m.EmptyValue) {
		body.EmptyValue = m.EmptyValue.ValueString()
	}

	typeutils.BoolDefault(&m.IgnoreFailure, false)

	return body, diags
}

// NewProcessorCSVDataSource returns a PF data source for the csv processor.
func NewProcessorCSVDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		attrField: schema.StringAttribute{
			Description: "The field to extract data from.",
			Required:    true,
		},
		"target_fields": schema.ListAttribute{
			Description: "The array of fields to assign extracted values to.",
			Required:    true,
			ElementType: types.StringType,
			Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
		},
		attrIgnoreMissing: schema.BoolAttribute{
			Description: descIgnoreMissingDocStop,
			Optional:    true,
			Computed:    true,
		},
		attrSeparator: schema.StringAttribute{
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
