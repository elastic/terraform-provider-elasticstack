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

package index

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceTemplate() *schema.Resource {
	templateSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the index template.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"composed_of": {
			Description: "An ordered list of component template names.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ignore_missing_component_templates": {
			Description: "A list of component template names that are ignored if missing.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"data_stream": {
			Description: "If this object is included, the template is used to create data streams and their backing indices. Supports an empty object.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"hidden": {
						Description: "If true, the data stream is hidden.",
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"allow_custom_routing": {
						Description: "If `true`, the data stream supports custom routing. Defaults to `false`. Available only in **8.x**",
						Type:        schema.TypeBool,
						Computed:    true,
					},
				},
			},
		},
		"index_patterns": {
			Description: "Array of wildcard (*) expressions used to match the names of data streams and indices during creation.",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"metadata": {
			Description: "Optional user metadata about the index template.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"priority": {
			Description: "Priority to determine index template precedence when a new data stream or index is created.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"template": {
			Description: "Template to be applied. It may optionally include an aliases, mappings, lifecycle, or settings configuration.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"alias": {
						Description: "Alias to add.",
						Type:        schema.TypeSet,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Description: "The alias name.",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"filter": {
									Description: "Query used to limit documents the alias can access.",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"index_routing": {
									Description: "Value used to route indexing operations to a specific shard. If specified, this overwrites the `routing` value for indexing operations.",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"is_hidden": {
									Description: "If true, the alias is hidden.",
									Type:        schema.TypeBool,
									Computed:    true,
								},
								"is_write_index": {
									Description: "If true, the index is the write index for the alias.",
									Type:        schema.TypeBool,
									Computed:    true,
								},
								"routing": {
									Description: "Value used to route indexing and search operations to a specific shard.",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"search_routing": {
									Description: "Value used to route search operations to a specific shard. If specified, this overwrites the routing value for search operations.",
									Type:        schema.TypeString,
									Computed:    true,
								},
							},
						},
					},
					"mappings": {
						Description: indexTemplateMappingsDescription,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"settings": {
						Description: "Configuration options for the index. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"lifecycle": {
						Description: "Lifecycle of data stream. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/data-stream-lifecycle.html",
						Type:        schema.TypeSet,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"data_retention": {
									Description: "The retention period of the data indexed in this data stream.",
									Type:        schema.TypeString,
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
		"version": {
			Description: "Version number used to manage index templates externally.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
	}

	schemautil.AddConnectionSchema(templateSchema)

	return &schema.Resource{
		Description: "Retrieves information about an existing index template definition. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-template.html",
		ReadContext: dataSourceIndexTemplateRead,
		Schema:      templateSchema,
	}
}

func dataSourceIndexTemplateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diags := clients.NewAPIClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	templateName := d.Get("name").(string)
	id, diags := client.ID(ctx, templateName)
	if diags.HasError() {
		return diags
	}
	d.SetId(id.String())

	return resourceIndexTemplateRead(ctx, d, meta)
}
