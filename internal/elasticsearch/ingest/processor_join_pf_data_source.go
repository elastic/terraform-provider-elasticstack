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

type processorJoinModel struct {
	CommonProcessorModel
	Field       types.String `tfsdk:"field"`
	Separator   types.String `tfsdk:"separator"`
	TargetField types.String `tfsdk:"target_field"`
}

func (m *processorJoinModel) TypeName() string { return "join" }

func (m *processorJoinModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorJoinBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if typeutils.IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if typeutils.IsKnown(m.Separator) {
		body.Separator = m.Separator.ValueString()
	}
	if typeutils.IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}

	return body, diags
}

// NewProcessorJoinDataSource returns a PF data source for the join processor.
func NewProcessorJoinDataSource() datasource.DataSource {
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
			Description: "Field containing array values to join.",
			Required:    true,
		},
		attrSeparator: schema.StringAttribute{
			Description: "The separator character.",
			Required:    true,
		},
		attrTargetField: schema.StringAttribute{
			Description: descTargetFieldInPlace,
			Optional:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorJoinModel{}, schema.Schema{
		Description: processorJoinDataSourceDescription,
		Attributes:  attrs,
	})
}
