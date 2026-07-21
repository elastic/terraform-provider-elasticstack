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

type processorSplitModel struct {
	CommonProcessorModel
	WithIgnorableTargetField
	Separator        types.String `tfsdk:"separator"`
	PreserveTrailing types.Bool   `tfsdk:"preserve_trailing"`
}

func (m *processorSplitModel) TypeName() string { return "split" }

func (m *processorSplitModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorSplitBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}
	body.WithIgnorableTargetFieldBody = m.toIgnorableTargetFieldBody(false)

	if typeutils.IsKnown(m.Separator) {
		body.Separator = m.Separator.ValueString()
	}
	if m.PreserveTrailing.IsNull() || m.PreserveTrailing.IsUnknown() {
		m.PreserveTrailing = types.BoolValue(false)
		body.PreserveTrailing = false
	} else {
		body.PreserveTrailing = m.PreserveTrailing.ValueBool()
	}

	return body, diags
}

// NewProcessorSplitDataSource returns a PF data source for the split processor.
func NewProcessorSplitDataSource() datasource.DataSource {
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
			Description: "The field to split",
			Required:    true,
		},
		attrSeparator: schema.StringAttribute{
			Description: "A regex which matches the separator, eg `,` or `\\s+`",
			Required:    true,
		},
		attrTargetField: schema.StringAttribute{
			Description: descTargetFieldInPlace,
			Optional:    true,
		},
		"preserve_trailing": schema.BoolAttribute{
			Description: "Preserves empty trailing fields, if any.",
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

	return NewProcessorDataSource(&processorSplitModel{}, schema.Schema{
		Description: processorSplitDataSourceDescription,
		Attributes:  attrs,
	})
}
