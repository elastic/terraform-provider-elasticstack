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

type processorGsubModel struct {
	CommonProcessorModel
	WithIgnorableTargetField
	Pattern     types.String `tfsdk:"pattern"`
	Replacement types.String `tfsdk:"replacement"`
}

func (m *processorGsubModel) TypeName() string { return "gsub" }

func (m *processorGsubModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorGsubBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}
	body.WithIgnorableTargetFieldBody = m.toIgnorableTargetFieldBody(false)

	if IsKnown(m.Pattern) {
		body.Pattern = m.Pattern.ValueString()
	}
	if IsKnown(m.Replacement) {
		body.Replacement = m.Replacement.ValueString()
	}

	return body, diags
}

// NewProcessorGsubDataSource returns a PF data source for the gsub processor.
func NewProcessorGsubDataSource() datasource.DataSource {
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
			Description: "The field to apply the replacement to.",
			Required:    true,
		},
		"pattern": schema.StringAttribute{
			Description: "The pattern to be replaced.",
			Required:    true,
		},
		"replacement": schema.StringAttribute{
			Description: "The string to replace the matching patterns with.",
			Required:    true,
		},
		attrTargetField: schema.StringAttribute{
			Description: descTargetFieldInPlace,
			Optional:    true,
		},
		attrIgnoreMissing: schema.BoolAttribute{
			Description: descIgnoreMissingDocStop,
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorGsubModel{}, schema.Schema{
		Description: processorGsubDataSourceDescription,
		Attributes:  attrs,
	})
}
