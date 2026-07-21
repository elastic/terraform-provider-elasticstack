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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorDissectModel struct {
	CommonProcessorModel
	Field           types.String `tfsdk:"field"`
	Pattern         types.String `tfsdk:"pattern"`
	AppendSeparator types.String `tfsdk:"append_separator"`
	IgnoreMissing   types.Bool   `tfsdk:"ignore_missing"`
}

func (m *processorDissectModel) TypeName() string { return "dissect" }

func (m *processorDissectModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorDissectBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if typeutils.IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if typeutils.IsKnown(m.Pattern) {
		body.Pattern = m.Pattern.ValueString()
	}
	if typeutils.IsKnown(m.AppendSeparator) {
		body.AppendSeparator = m.AppendSeparator.ValueString()
	}
	if m.IgnoreMissing.IsNull() || m.IgnoreMissing.IsUnknown() {
		m.IgnoreMissing = types.BoolValue(false)
		body.IgnoreMissing = false
	} else {
		body.IgnoreMissing = m.IgnoreMissing.ValueBool()
	}

	return body, diags
}

// NewProcessorDissectDataSource returns a PF data source for the dissect processor.
func NewProcessorDissectDataSource() datasource.DataSource {
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
			Description: "The field to dissect.",
			Required:    true,
		},
		"pattern": schema.StringAttribute{
			Description: "The pattern to apply to the field.",
			Required:    true,
		},
		"append_separator": schema.StringAttribute{
			Description: "The character(s) that separate the appended fields.",
			Optional:    true,
			Computed:    true,
		},
		attrIgnoreMissing: schema.BoolAttribute{
			Description: descIgnoreMissingDocStop,
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorDissectModel{}, schema.Schema{
		Description: processorDissectDataSourceDescription,
		Attributes:  attrs,
	})
}
