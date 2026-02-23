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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/tfsdkutils"
	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceComponentTemplate() *schema.Resource {
	// NOTE: component_template and index_template uses the same schema
	componentTemplateSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "Name of the component template to create.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"metadata": {
			Description:      "Optional user metadata about the component template.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: tfsdkutils.DiffJSONSuppress,
		},
		"template": {
			Description: "Template to be applied. It may optionally include an aliases, mappings, or settings configuration.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"alias": {
						Description: "Alias to add.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Description: componentTemplateAliasNameDescription,
									Type:        schema.TypeString,
									Required:    true,
								},
								"filter": {
									Description:      "Query used to limit documents the alias can access.",
									Type:             schema.TypeString,
									Optional:         true,
									DiffSuppressFunc: tfsdkutils.DiffJSONSuppress,
									ValidateFunc:     validation.StringIsJSON,
								},
								"index_routing": {
									Description: "Value used to route indexing operations to a specific shard. If specified, this overwrites the routing value for indexing operations.",
									Type:        schema.TypeString,
									Optional:    true,
								},
								"is_hidden": {
									Description: "If true, the alias is hidden.",
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     false,
								},
								"is_write_index": {
									Description: "If true, the index is the write index for the alias.",
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     false,
								},
								"routing": {
									Description: "Value used to route indexing and search operations to a specific shard.",
									Type:        schema.TypeString,
									Optional:    true,
								},
								"search_routing": {
									Description: "Value used to route search operations to a specific shard. If specified, this overwrites the routing value for search operations.",
									Type:        schema.TypeString,
									Optional:    true,
								},
							},
						},
					},
					"mappings": {
						Description:      indexTemplateMappingsDescription,
						Type:             schema.TypeString,
						Optional:         true,
						DiffSuppressFunc: tfsdkutils.DiffJSONSuppress,
						ValidateFunc: validation.All(
							validation.StringIsJSON, stringIsJSONObject,
						),
					},
					"settings": {
						Description:      componentTemplateSettingsDescription,
						Type:             schema.TypeString,
						Optional:         true,
						DiffSuppressFunc: tfsdkutils.DiffIndexSettingSuppress,
						ValidateFunc: validation.All(
							validation.StringIsJSON, stringIsJSONObject,
						),
					},
				},
			},
		},
		"version": {
			Description: "Version number used to manage component templates externally.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
	}

	schemautil.AddConnectionSchema(componentTemplateSchema)

	return &schema.Resource{
		Description: componentTemplateResourceDescription,

		CreateContext: resourceComponentTemplatePut,
		UpdateContext: resourceComponentTemplatePut,
		ReadContext:   resourceComponentTemplateRead,
		DeleteContext: resourceComponentTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: componentTemplateSchema,
	}
}

func resourceComponentTemplatePut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diags := clients.NewAPIClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	componentID := d.Get("name").(string)
	id, diags := client.ID(ctx, componentID)
	if diags.HasError() {
		return diags
	}
	var componentTemplate models.ComponentTemplate
	componentTemplate.Name = componentID

	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]any)
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		componentTemplate.Meta = metadata
	}

	if v, ok := d.GetOk("template"); ok {
		templ, ok, diags := expandTemplate(v)
		if diags != nil {
			return diags
		}

		if ok {
			componentTemplate.Template = &templ
		}
	}

	if v, ok := d.GetOk("version"); ok {
		definedVer := v.(int)
		componentTemplate.Version = &definedVer
	}

	if diags := elasticsearch.PutComponentTemplate(ctx, client, &componentTemplate); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceComponentTemplateRead(ctx, d, meta)
}

func resourceComponentTemplateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diags := clients.NewAPIClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	compID, diags := clients.CompositeIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	templateID := compID.ResourceID

	tpl, diags := elasticsearch.GetComponentTemplate(ctx, client, templateID, false)
	if tpl == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Component template "%s" not found, removing from state`, compID.ResourceID))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	// set the fields
	if err := d.Set("name", tpl.Name); err != nil {
		return diag.FromErr(err)
	}

	if tpl.ComponentTemplate.Meta != nil {
		metadata, err := json.Marshal(tpl.ComponentTemplate.Meta)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("metadata", string(metadata)); err != nil {
			return diag.FromErr(err)
		}
	}

	if tpl.ComponentTemplate.Template != nil {
		template, diags := flattenTemplateData(tpl.ComponentTemplate.Template)
		if diags.HasError() {
			return diags
		}

		if err := d.Set("template", template); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("version", tpl.ComponentTemplate.Version); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceComponentTemplateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diags := clients.NewAPIClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()
	compID, diags := clients.CompositeIDFromStr(id)
	if diags.HasError() {
		return diags
	}
	if diags := elasticsearch.DeleteComponentTemplate(ctx, client, compID.ResourceID); diags.HasError() {
		return diags
	}
	return diags
}
