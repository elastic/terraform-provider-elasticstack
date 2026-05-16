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

type processorUserAgentModel struct {
	CommonProcessorModel
	WithIgnorableTargetField
	RegexFile         types.String `tfsdk:"regex_file"`
	Properties        types.Set    `tfsdk:"properties"`
	ExtractDeviceType types.Bool   `tfsdk:"extract_device_type"`
}

func (m *processorUserAgentModel) TypeName() string { return "user_agent" }

func (m *processorUserAgentModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorUserAgentBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}
	body.WithIgnorableTargetFieldBody = m.toIgnorableTargetFieldBody(false)

	if IsKnown(m.RegexFile) {
		body.RegexFile = m.RegexFile.ValueString()
	}
	if IsKnown(m.Properties) {
		elems := make([]string, 0, len(m.Properties.Elements()))
		for _, elem := range m.Properties.Elements() {
			str, ok := elem.(types.String)
			if !ok || !IsKnown(str) {
				if !ok {
					diags.AddError("Invalid properties element type", "expected types.String")
				} else {
					diags.AddError("Unknown properties element", "properties elements cannot be unknown")
				}
				continue
			}
			elems = append(elems, str.ValueString())
		}
		body.Properties = elems
	}
	if IsKnown(m.ExtractDeviceType) {
		v := m.ExtractDeviceType.ValueBool()
		body.ExtractDeviceType = &v
	}

	return body, diags
}

// NewProcessorUserAgentDataSource returns a PF data source for the user_agent processor.
func NewProcessorUserAgentDataSource() datasource.DataSource {
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
			Description: "The field containing the user agent string.",
			Required:    true,
		},
		"target_field": schema.StringAttribute{
			Description: "The field that will be filled with the user agent details.",
			Optional:    true,
		},
		"ignore_missing": schema.BoolAttribute{
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
			Optional:    true,
			Computed:    true,
		},
		"regex_file": schema.StringAttribute{
			Description: "The name of the file in the `config/ingest-user-agent` directory containing the regular expressions for parsing the user agent string.",
			Optional:    true,
		},
		"properties": schema.SetAttribute{
			Description: "Controls what properties are added to `target_field`.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"extract_device_type": schema.BoolAttribute{
			Description: "Extracts device type from the user agent string on a best-effort basis. Supported only starting from Elasticsearch version **8.0**",
			Optional:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorUserAgentModel{}, schema.Schema{
		Description: processorUserAgentDataSourceDescription,
		Attributes:  attrs,
	})
}
