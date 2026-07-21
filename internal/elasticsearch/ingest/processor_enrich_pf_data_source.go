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

type processorEnrichModel struct {
	CommonProcessorModel
	WithIgnorableTargetField
	PolicyName    types.String `tfsdk:"policy_name"`
	Override      types.Bool   `tfsdk:"override"`
	MaxMatches    types.Int64  `tfsdk:"max_matches"`
	ShapeRelation types.String `tfsdk:"shape_relation"`
}

func (m *processorEnrichModel) TypeName() string { return "enrich" }

func (m *processorEnrichModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorEnrichBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}
	body.WithIgnorableTargetFieldBody = m.toIgnorableTargetFieldBody(false)

	if typeutils.IsKnown(m.PolicyName) {
		body.PolicyName = m.PolicyName.ValueString()
	}
	if m.Override.IsNull() || m.Override.IsUnknown() {
		m.Override = types.BoolValue(true)
		body.Override = true
	} else {
		body.Override = m.Override.ValueBool()
	}
	if m.MaxMatches.IsNull() || m.MaxMatches.IsUnknown() {
		m.MaxMatches = types.Int64Value(1)
		body.MaxMatches = 1
	} else {
		body.MaxMatches = int(m.MaxMatches.ValueInt64())
	}
	if typeutils.IsKnown(m.ShapeRelation) {
		body.ShapeRelation = m.ShapeRelation.ValueString()
	}

	return body, diags
}

// NewProcessorEnrichDataSource returns a PF data source for the enrich processor.
func NewProcessorEnrichDataSource() datasource.DataSource {
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
			Description: "The field in the input document that matches the policies match_field used to retrieve the enrichment data.",
			Required:    true,
		},
		attrTargetField: schema.StringAttribute{
			Description: "Field added to incoming documents to contain enrich data.",
			Required:    true,
		},
		attrIgnoreMissing: schema.BoolAttribute{
			Description: descIgnoreMissingDocStop,
			Optional:    true,
			Computed:    true,
		},
		"policy_name": schema.StringAttribute{
			Description: "The name of the enrich policy to use.",
			Required:    true,
		},
		attrOverride: schema.BoolAttribute{
			Description: "If processor will update fields with pre-existing non-null-valued field.",
			Optional:    true,
			Computed:    true,
		},
		"max_matches": schema.Int64Attribute{
			Description: "The maximum number of matched documents to include under the configured target field.",
			Optional:    true,
			Computed:    true,
		},
		"shape_relation": schema.StringAttribute{
			Description: "A spatial relation operator used to match the geoshape of incoming documents to documents in the enrich index.",
			Optional:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorEnrichModel{}, schema.Schema{
		Description: processorEnrichDataSourceDescription,
		Attributes:  attrs,
	})
}
