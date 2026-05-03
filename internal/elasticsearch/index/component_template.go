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

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
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
						Set:         hashAliasByName,
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
	factory, diags := clients.ConvertMetaToFactory(meta)
	if diags.HasError() {
		return diags
	}
	client, diags := factory.GetElasticsearchClientFromSDK(d)
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
	factory, diags := clients.ConvertMetaToFactory(meta)
	if diags.HasError() {
		return diags
	}
	client, diags := factory.GetElasticsearchClientFromSDK(d)
	if diags.HasError() {
		return diags
	}
	compID, diags := clients.CompositeIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	templateID := compID.ResourceID

	tpl, diags := elasticsearch.GetComponentTemplate(ctx, client, templateID)
	if tpl == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Component template "%s" not found, removing from state`, compID.ResourceID))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	modelTpl := toModelComponentTemplateResponse(tpl)
	if modelTpl == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Component template "%s" not found, removing from state`, compID.ResourceID))
		d.SetId("")
		return diags
	}

	// set the fields
	if err := d.Set("name", modelTpl.Name); err != nil {
		return diag.FromErr(err)
	}

	if modelTpl.ComponentTemplate.Meta != nil {
		metadata, err := json.Marshal(modelTpl.ComponentTemplate.Meta)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("metadata", string(metadata)); err != nil {
			return diag.FromErr(err)
		}
	}

	if modelTpl.ComponentTemplate.Template != nil {
		template, diags := flattenTemplateData(
			modelTpl.ComponentTemplate.Template,
			extractAliasRoutingFromTemplateState(d.Get("template")),
		)
		if diags.HasError() {
			return diags
		}

		if err := d.Set("template", template); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("version", modelTpl.ComponentTemplate.Version); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceComponentTemplateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	factory, diags := clients.ConvertMetaToFactory(meta)
	if diags.HasError() {
		return diags
	}
	client, diags := factory.GetElasticsearchClientFromSDK(d)
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

func toModelComponentTemplateResponse(tpl *estypes.ClusterComponentTemplate) *models.ComponentTemplateResponse {
	if tpl == nil {
		return nil
	}

	resp := &models.ComponentTemplateResponse{
		Name: tpl.Name,
		ComponentTemplate: models.ComponentTemplate{
			Name: tpl.Name,
		},
	}

	if tpl.ComponentTemplate.Meta_ != nil {
		metaBytes, _ := json.Marshal(tpl.ComponentTemplate.Meta_)
		var metaMap map[string]any
		_ = json.Unmarshal(metaBytes, &metaMap)
		resp.ComponentTemplate.Meta = metaMap
	}

	{
		t := &models.Template{}

		if tpl.ComponentTemplate.Template.Settings != nil {
			settingsBytes, _ := json.Marshal(tpl.ComponentTemplate.Template.Settings)
			var settingsMap map[string]any
			_ = json.Unmarshal(settingsBytes, &settingsMap)
			t.Settings = settingsMap
		}

		if tpl.ComponentTemplate.Template.Mappings != nil {
			mappingsBytes, _ := json.Marshal(tpl.ComponentTemplate.Template.Mappings)
			var mappingsMap map[string]any
			_ = json.Unmarshal(mappingsBytes, &mappingsMap)
			t.Mappings = mappingsMap
		}

		if len(tpl.ComponentTemplate.Template.Aliases) > 0 {
			t.Aliases = make(map[string]models.IndexAlias, len(tpl.ComponentTemplate.Template.Aliases))
			for name, alias := range tpl.ComponentTemplate.Template.Aliases {
				ia := models.IndexAlias{Name: name}
				if alias.Filter != nil {
					filterBytes, _ := json.Marshal(alias.Filter)
					var filterMap map[string]any
					_ = json.Unmarshal(filterBytes, &filterMap)
					ia.Filter = filterMap
				}
				if alias.IndexRouting != nil {
					ia.IndexRouting = *alias.IndexRouting
				}
				if alias.IsHidden != nil {
					ia.IsHidden = *alias.IsHidden
				}
				if alias.IsWriteIndex != nil {
					ia.IsWriteIndex = *alias.IsWriteIndex
				}
				if alias.Routing != nil {
					ia.Routing = *alias.Routing
				}
				if alias.SearchRouting != nil {
					ia.SearchRouting = *alias.SearchRouting
				}
				t.Aliases[name] = ia
			}
		}

		resp.ComponentTemplate.Template = t
	}

	return resp
}
