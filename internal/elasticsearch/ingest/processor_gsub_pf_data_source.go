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
	ID            types.String `tfsdk:"id"`
	JSON          types.String `tfsdk:"json"`
	Field         types.String `tfsdk:"field"`
	Pattern       types.String `tfsdk:"pattern"`
	Replacement   types.String `tfsdk:"replacement"`
	TargetField   types.String `tfsdk:"target_field"`
	IgnoreMissing types.Bool   `tfsdk:"ignore_missing"`
}

func (m *processorGsubModel) TypeName() string    { return "gsub" }
func (m *processorGsubModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorGsubModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorGsubModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorGsubBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.Pattern) {
		body.Pattern = m.Pattern.ValueString()
	}
	if IsKnown(m.Replacement) {
		body.Replacement = m.Replacement.ValueString()
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

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorGsubDataSource returns a PF data source for the gsub processor.
func NewProcessorGsubDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource.",
			Computed:    true,
		},
		"json": schema.StringAttribute{
			Description: "JSON representation of this data source.",
			Computed:    true,
		},
		"field": schema.StringAttribute{
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
		"target_field": schema.StringAttribute{
			Description: "The field to assign the converted value to, by default `field` is updated in-place.",
			Optional:    true,
		},
		"ignore_missing": schema.BoolAttribute{
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
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
