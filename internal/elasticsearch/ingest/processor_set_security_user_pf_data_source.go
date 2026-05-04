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

type processorSetSecurityUserModel struct {
	CommonProcessorModel
	ID         types.String `tfsdk:"id"`
	JSON       types.String `tfsdk:"json"`
	Field      types.String `tfsdk:"field"`
	Properties types.Set    `tfsdk:"properties"`
}

func (m *processorSetSecurityUserModel) TypeName() string    { return "set_security_user" }
func (m *processorSetSecurityUserModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorSetSecurityUserModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorSetSecurityUserModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorSetSecurityUserBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
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

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorSetSecurityUserDataSource returns a PF data source for the set_security_user processor.
func NewProcessorSetSecurityUserDataSource() datasource.DataSource {
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
			Description: "The field to store the user information into.",
			Required:    true,
		},
		"properties": schema.SetAttribute{
			Description: "Controls what user related properties are added to the `field`.",
			Optional:    true,
			ElementType: types.StringType,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorSetSecurityUserModel{}, schema.Schema{
		Description: processorSetSecurityUserDataSourceDescription,
		Attributes:  attrs,
	})
}
