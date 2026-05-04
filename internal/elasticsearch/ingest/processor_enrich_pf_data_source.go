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

type processorEnrichModel struct {
	CommonProcessorModel
	ID            types.String `tfsdk:"id"`
	JSON          types.String `tfsdk:"json"`
	Field         types.String `tfsdk:"field"`
	TargetField   types.String `tfsdk:"target_field"`
	IgnoreMissing types.Bool   `tfsdk:"ignore_missing"`
	PolicyName    types.String `tfsdk:"policy_name"`
	Override      types.Bool   `tfsdk:"override"`
	MaxMatches    types.Int64  `tfsdk:"max_matches"`
	ShapeRelation types.String `tfsdk:"shape_relation"`
}

func (m *processorEnrichModel) TypeName() string    { return "enrich" }
func (m *processorEnrichModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorEnrichModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorEnrichModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorEnrichBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}
	if m.IgnoreMissing.IsNull() || m.IgnoreMissing.IsUnknown() {
		m.IgnoreMissing = types.BoolValue(false)
		body.IgnoreMissing = false
	} else {
		body.IgnoreMissing = m.IgnoreMissing.ValueBool()
	}
	if IsKnown(m.PolicyName) {
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
	if IsKnown(m.ShapeRelation) {
		body.ShapeRelation = m.ShapeRelation.ValueString()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorEnrichDataSource returns a PF data source for the enrich processor.
func NewProcessorEnrichDataSource() datasource.DataSource {
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
			Description: "The field in the input document that matches the policies match_field used to retrieve the enrichment data.",
			Required:    true,
		},
		"target_field": schema.StringAttribute{
			Description: "Field added to incoming documents to contain enrich data.",
			Required:    true,
		},
		"ignore_missing": schema.BoolAttribute{
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
			Optional:    true,
			Computed:    true,
		},
		"policy_name": schema.StringAttribute{
			Description: "The name of the enrich policy to use.",
			Required:    true,
		},
		"override": schema.BoolAttribute{
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
