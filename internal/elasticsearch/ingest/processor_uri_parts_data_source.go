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

func DataSourceProcessorURIParts() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "Field containing the URI string.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"target_field": {
			Description: "Output field for the URI object.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"keep_original": {
			Description: "If true, the processor copies the unparsed URI to `<target_field>.original.`",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"remove_if_successful": {
			Description: "If `true`, the processor removes the `field` after parsing the URI string. If parsing fails, the processor does not remove the `field`.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"description": {
			Description: "Description of the processor. ",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"if": {
			Description: "Conditionally execute the processor",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"ignore_failure": {
			Description: "Ignore failures for the processor. ",
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
		Description: processorURIPartsDataSourceDescription,

		ReadContext: dataSourceProcessorURIPartsRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorURIPartsRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorURIParts{}

	processor.Field = d.Get("field").(string)
	processor.IgnoreFailure = d.Get("ignore_failure").(bool)
	processor.KeepOriginal = d.Get("keep_original").(bool)
	processor.RemoveIfSuccessful = d.Get("remove_if_successful").(bool)

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

	processorJSON, err := json.MarshalIndent(map[string]*models.ProcessorURIParts{"uri_parts": processor}, "", " ")
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
