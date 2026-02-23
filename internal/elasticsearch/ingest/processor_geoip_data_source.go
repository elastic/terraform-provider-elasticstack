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
	_ "embed"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//go:embed processor_geoip_data_source.md
var geoipDataSourceDescription string

func DataSourceProcessorGeoip() *schema.Resource {
	processorSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"field": {
			Description: "The field to get the ip address from for the geographical lookup.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"target_field": {
			Description: "The field that will hold the geographical information looked up from the MaxMind database.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "geoip",
		},
		"database_file": {
			Description: processorGeoIPDatabaseFileDescription,
			Type:        schema.TypeString,
			Optional:    true,
		},
		"properties": {
			Description: "Controls what properties are added to the `target_field` based on the geoip lookup.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ignore_missing": {
			Description: "If `true` and `field` does not exist, the processor quietly exits without modifying the document.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"first_only": {
			Description: "If `true` only first found geoip data will be returned, even if field contains array.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"json": {
			Description: "JSON representation of this data source.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	return &schema.Resource{
		Description: geoipDataSourceDescription,

		ReadContext: dataSourceProcessorGeoipRead,

		Schema: processorSchema,
	}
}

func dataSourceProcessorGeoipRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	var diags diag.Diagnostics

	processor := &models.ProcessorGeoip{}

	processor.IgnoreMissing = d.Get("ignore_missing").(bool)
	processor.FirstOnly = d.Get("first_only").(bool)
	processor.Field = d.Get("field").(string)
	processor.TargetField = d.Get("target_field").(string)

	if v, ok := d.GetOk("properties"); ok {
		props := v.(*schema.Set)
		properties := make([]string, props.Len())
		for i, p := range props.List() {
			properties[i] = p.(string)
		}
		processor.Properties = properties
	}

	if v, ok := d.GetOk("database_file"); ok {
		processor.DatabaseFile = v.(string)
	}

	processorJSON, err := json.MarshalIndent(map[string]*models.ProcessorGeoip{"geoip": processor}, "", " ")
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
