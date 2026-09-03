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
	"context"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

	if typeutils.IsKnown(m.ModelID) {
		body.ModelID = m.ModelID.ValueString()
	}

	if typeutils.IsKnown(m.InputOutput) {
		io := &processorInferenceInputOutputBody{}
		if v, ok := m.InputOutput.Attributes()["input_field"]; ok {
			if s, ok := v.(types.String); ok && typeutils.IsKnown(s) {
				io.InputField = s.ValueString()
			}
		}
		if v, ok := m.InputOutput.Attributes()["output_field"]; ok {
			if s, ok := v.(types.String); ok && typeutils.IsKnown(s) {
				io.OutputField = s.ValueString()
			}
		}
		body.InputOutput = io
	}

	if typeutils.IsKnown(m.FieldMap) {
		body.FieldMap = typeutils.MapTypeAs[string](context.Background(), m.FieldMap, path.Root("field_map"), &diags)
	}

	if typeutils.IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}

	typeutils.BoolDefault(&m.IgnoreFailure, false)

	return body, diags
}

// NewProcessorInferenceDataSource returns a PF data source for the inference processor.
func NewProcessorInferenceDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"model_id": schema.StringAttribute{
			Description: "The ID or alias for the trained model, or the ID of the deployment.",
			Required:    true,
		},

		"field_map": schema.MapAttribute{
			Description: "Maps the document field names to the known field names of the model. Maps the document fields to the model's expected input fields.",
			Optional:    true,
			ElementType: types.StringType,
		},
		attrTargetField: schema.StringAttribute{
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
