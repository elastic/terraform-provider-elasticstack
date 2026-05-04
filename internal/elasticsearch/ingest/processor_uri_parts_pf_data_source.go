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

type processorURIPartsModel struct {
	CommonProcessorModel
	ID                 types.String `tfsdk:"id"`
	JSON               types.String `tfsdk:"json"`
	Field              types.String `tfsdk:"field"`
	TargetField        types.String `tfsdk:"target_field"`
	KeepOriginal       types.Bool   `tfsdk:"keep_original"`
	RemoveIfSuccessful types.Bool   `tfsdk:"remove_if_successful"`
}

func (m *processorURIPartsModel) TypeName() string    { return "uri_parts" }
func (m *processorURIPartsModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorURIPartsModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorURIPartsModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorURIPartsBody{}

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
	if m.KeepOriginal.IsNull() || m.KeepOriginal.IsUnknown() {
		m.KeepOriginal = types.BoolValue(true)
		body.KeepOriginal = true
	} else {
		body.KeepOriginal = m.KeepOriginal.ValueBool()
	}
	if m.RemoveIfSuccessful.IsNull() || m.RemoveIfSuccessful.IsUnknown() {
		m.RemoveIfSuccessful = types.BoolValue(false)
		body.RemoveIfSuccessful = false
	} else {
		body.RemoveIfSuccessful = m.RemoveIfSuccessful.ValueBool()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorURIPartsDataSource returns a PF data source for the uri_parts processor.
func NewProcessorURIPartsDataSource() datasource.DataSource {
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
			Description: "Field containing the URI string.",
			Required:    true,
		},
		"target_field": schema.StringAttribute{
			Description: "Output field for the URI object.",
			Optional:    true,
		},
		"keep_original": schema.BoolAttribute{
			Description: "If true, the processor copies the unparsed URI to `<target_field>.original.`",
			Optional:    true,
			Computed:    true,
		},
		"remove_if_successful": schema.BoolAttribute{
			Description: "If `true`, the processor removes the `field` after parsing the URI string. If parsing fails, the processor does not remove the `field`.",
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorURIPartsModel{}, schema.Schema{
		Description: processorURIPartsDataSourceDescription,
		Attributes:  attrs,
	})
}
