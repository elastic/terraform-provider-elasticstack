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
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/tfsdkutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceProcessorInference() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"model_id": {
			Description: "The ID or alias for the trained model, or the ID of the deployment.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"input_output": {
			Description: "Input and output field mappings for the inference processor.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"input_field": {
						Description: "The field name from which the inference processor reads its input value.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"output_field": {
						Description: "The field name to which the inference processor writes its output.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
		},
		"field_map": {
			Description: "Maps the document field names to the known field names of the model. Maps the document fields to the model's expected input fields.",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"target_field": {
			Description: "Field added to incoming documents to contain results objects.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"description": {
			Description: "Description of the processor.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"if": {
			Description: "Conditionally execute the processor.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"ignore_failure": {
			Description: "Ignore failures for the processor.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"on_failure": {
			Description: "Handle failures for the processor.",
			Type:        schema.TypeList,
			Optional:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: tfsdkutils.DiffJSONSuppress,
			},
		},
		"tag": {
			Description: "Identifier for the processor.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"json": {
			Description: "JSON representation of this data source.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	return &schema.Resource{
		Description: processorInferenceDataSourceDescription,
		ReadContext: dataSourceProcessorInferenceRead,
		Schema:      processorSchema,
	}
}

func dataSourceProcessorInferenceRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorInference{}
	processor.ModelID = d.Get("model_id").(string)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)

	if v, ok := d.GetOk("input_output"); ok {
		list := v.([]any)
		if len(list) > 0 {
			raw := list[0].(map[string]any)
			io := &models.ProcessorInferenceInputOutput{
				InputField: raw["input_field"].(string),
			}
			if out, ok := raw["output_field"].(string); ok && out != "" {
				io.OutputField = out
			}
			processor.InputOutput = io
		}
	}

	if v, ok := d.GetOk("field_map"); ok {
		fm := v.(map[string]any)
		fieldMap := make(map[string]string, len(fm))
		for k, val := range fm {
			fieldMap[k] = val.(string)
		}
		processor.FieldMap = fieldMap
	}

	if v, ok := d.GetOk("target_field"); ok {
		processor.TargetField = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		processor.Description = v.(string)
	}
	if v, ok := d.GetOk("if"); ok {
		processor.If = v.(string)
	}
	if v, ok := d.GetOk("tag"); ok {
		processor.Tag = v.(string)
	}
	if v, ok := d.GetOk("on_failure"); ok {
		onFailure := make([]map[string]any, len(v.([]any)))
		for i, f := range v.([]any) {
			item := make(map[string]any)
			if err := json.NewDecoder(strings.NewReader(f.(string))).Decode(&item); err != nil {
				return diag.FromErr(err)
			}
			onFailure[i] = item
		}
		processor.OnFailure = onFailure
	}

	processorJSON, err := json.MarshalIndent(map[string]*models.ProcessorInference{"inference": processor}, "", " ")
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("json", string(processorJSON)); err != nil {
		return diag.FromErr(err)
	}

	hash, err := schemautil.StringToHash(string(processorJSON))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*hash)

	return diags
}
