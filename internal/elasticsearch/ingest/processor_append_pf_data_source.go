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

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorAppendModel struct {
	CommonProcessorModel
	ID              types.String `tfsdk:"id"`
	JSON            types.String `tfsdk:"json"`
	Field           types.String `tfsdk:"field"`
	Value           types.List   `tfsdk:"value"`
	AllowDuplicates types.Bool   `tfsdk:"allow_duplicates"`
	MediaType       types.String `tfsdk:"media_type"`
}

func (m *processorAppendModel) TypeName() string    { return "append" }
func (m *processorAppendModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorAppendModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorAppendModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorAppendBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}

	if IsKnown(m.Value) {
		elems := make([]string, 0, len(m.Value.Elements()))
		for _, elem := range m.Value.Elements() {
			str, ok := elem.(types.String)
			if !ok || !IsKnown(str) {
				if !ok {
					diags.AddError("Invalid value element type", "expected types.String")
				} else {
					diags.AddError("Unknown value element", "value elements cannot be unknown")
				}
				continue
			}
			elems = append(elems, str.ValueString())
		}
		body.Value = elems
	}

	if m.AllowDuplicates.IsNull() || m.AllowDuplicates.IsUnknown() {
		m.AllowDuplicates = types.BoolValue(true)
		body.AllowDuplicates = true
	} else {
		body.AllowDuplicates = m.AllowDuplicates.ValueBool()
	}

	if IsKnown(m.MediaType) {
		body.MediaType = m.MediaType.ValueString()
	}

	// Ensure ignore_failure default is reflected in state.
	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorAppendDataSource returns a PF data source for the append processor.
func NewProcessorAppendDataSource() datasource.DataSource {
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
			Description: "The field to be appended to.",
			Required:    true,
		},
		"value": schema.ListAttribute{
			Description: "The value to be appended.",
			Required:    true,
			ElementType: types.StringType,
			Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
		},
		"allow_duplicates": schema.BoolAttribute{
			Description: "If `false`, the processor does not append values already present in the field.",
			Optional:    true,
			Computed:    true,
		},
		"media_type": schema.StringAttribute{
			Description: processorAppendMediaTypeDescription,
			Optional:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorAppendModel{}, schema.Schema{
		Description: processorAppendDataSourceDescription,
		Attributes:  attrs,
	})
}
