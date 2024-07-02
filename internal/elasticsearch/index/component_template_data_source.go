package index

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceComponentTemplate() *schema.Resource {
	componentTemplateSchema := map[string]*schema.Schema{
		"name": {
			Description: "Name of the component template to create.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"metadata": {
			Description: "Optional user metadata about the component template.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"template": {
			Description: "Template to be applied. It may optionally include an aliases, mappings, or settings configuration.",
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
									Description: "The alias name. Index alias names support date math. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/date-math-index-names.html",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"filter": {
									Description: "Query used to limit documents the alias can access.",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"index_routing": {
									Description: "Value used to route indexing operations to a specific shard. If specified, this overwrites the routing value for indexing operations.",
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
						Description: "Mapping for fields in the index.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"settings": {
						Description: "Configuration options for the index. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings",
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"version": {
			Description: "Version number used to manage component templates externally.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
	}

	utils.AddConnectionSchema(componentTemplateSchema)

	return &schema.Resource{
		Description: "Gets an existing component template. Component templates are building blocks for constructing index templates that specify index mappings, settings, and aliases. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-component-template.html",

		ReadContext: datasourceComponentTemplateRead,

		Schema: componentTemplateSchema,
	}
}

func datasourceComponentTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	componentId := d.Get("name").(string)
	compId, diags := client.ID(ctx, componentId)
	if diags.HasError() {
		return diags
	}
	templateId := compId.ResourceId

	tpl, diags := elasticsearch.GetComponentTemplate(ctx, client, templateId)
	if tpl == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Component template "%s" not found, removing from state`, compId.ResourceId))
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
	d.SetId(compId.String())

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
