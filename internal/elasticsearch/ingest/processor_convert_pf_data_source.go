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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorConvertModel struct {
	CommonProcessorModel
	WithIgnorableTargetField
	Type types.String `tfsdk:"type"`
}

func (m *processorConvertModel) TypeName() string { return "convert" }

func (m *processorConvertModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorConvertBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}
	body.WithIgnorableTargetFieldBody = m.toIgnorableTargetFieldBody(false)

	if typeutils.IsKnown(m.Type) {
		body.Type = m.Type.ValueString()
	}

	return body, diags
}

// NewProcessorConvertDataSource returns a PF data source for the convert processor.
func NewProcessorConvertDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: descIdentifierWithPeriod,
			Computed:    true,
		},
		attrJSON: schema.StringAttribute{
			Description: descJSONDataSource,
			Computed:    true,
		},
		attrField: schema.StringAttribute{
			Description: "The field whose value is to be converted.",
			Required:    true,
		},
		attrTargetField: schema.StringAttribute{
			Description: "The field to assign the converted value to.",
			Optional:    true,
		},
		attrIgnoreMissing: schema.BoolAttribute{
			Description: descIgnoreMissingDocStop,
			Optional:    true,
			Computed:    true,
		},
		"type": schema.StringAttribute{
			Description: "The type to convert the existing value to",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf("integer", "long", "float", "double", "string", "boolean", "ip", "auto")},
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorConvertModel{}, schema.Schema{
		Description: processorConvertDataSourceDescription,
		Attributes:  attrs,
	})
}
