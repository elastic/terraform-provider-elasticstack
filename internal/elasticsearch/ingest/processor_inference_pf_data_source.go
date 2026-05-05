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

type processorInferenceModel struct {
	CommonProcessorModel
	ModelID     types.String `tfsdk:"model_id"`
	InputOutput types.Object `tfsdk:"input_output"`
	FieldMap    types.Map    `tfsdk:"field_map"`
	TargetField types.String `tfsdk:"target_field"`
}

func (m *processorInferenceModel) TypeName() string { return "inference" }

func (m *processorInferenceModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorInferenceBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if IsKnown(m.ModelID) {
		body.ModelID = m.ModelID.ValueString()
	}

	if IsKnown(m.InputOutput) {
		io := &processorInferenceInputOutputBody{}
		if v, ok := m.InputOutput.Attributes()["input_field"]; ok {
			if s, ok := v.(types.String); ok && IsKnown(s) {
				io.InputField = s.ValueString()
			}
		}
		if v, ok := m.InputOutput.Attributes()["output_field"]; ok {
			if s, ok := v.(types.String); ok && IsKnown(s) {
				io.OutputField = s.ValueString()
			}
		}
		body.InputOutput = io
	}

	if IsKnown(m.FieldMap) {
		fm := make(map[string]string, len(m.FieldMap.Elements()))
		for k, v := range m.FieldMap.Elements() {
			str, ok := v.(types.String)
			if !ok || !IsKnown(str) {
				if !ok {
					diags.AddError("Invalid field_map element type", "expected types.String")
				} else {
					diags.AddError("Unknown field_map element", "field_map elements cannot be unknown")
				}
				continue
			}
			fm[k] = str.ValueString()
		}
		body.FieldMap = fm
	}

	if IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorInferenceDataSource returns a PF data source for the inference processor.
func NewProcessorInferenceDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource",
			Computed:    true,
		},
		"json": schema.StringAttribute{
			Description: "JSON representation of this data source.",
			Computed:    true,
		},
		"model_id": schema.StringAttribute{
			Description: "The ID or alias for the trained model, or the ID of the deployment.",
			Required:    true,
		},

		"field_map": schema.MapAttribute{
			Description: "Maps the document field names to the known field names of the model. Maps the document fields to the model's expected input fields.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"target_field": schema.StringAttribute{
			Description: "Field added to incoming documents to contain results objects.",
			Optional:    true,
		},
	}

	blocks := map[string]schema.Block{
		"input_output": schema.SingleNestedBlock{
			Description: "Input and output field mappings for the inference processor.",
			Attributes: map[string]schema.Attribute{
				"input_field": schema.StringAttribute{
					Description: "The field name from which the inference processor reads its input value.",
					Required:    true,
				},
				"output_field": schema.StringAttribute{
					Description: "The field name to which the inference processor writes its output.",
					Optional:    true,
				},
			},
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorInferenceModel{}, schema.Schema{
		Description: processorInferenceDataSourceDescription,
		Attributes:  attrs,
		Blocks:      blocks,
	})
}
