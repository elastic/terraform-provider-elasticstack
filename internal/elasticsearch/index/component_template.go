package index

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
			DiffSuppressFunc: utils.DiffJsonSuppress,
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
									Description: "The alias name. Index alias names support date math. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/date-math-index-names.html",
									Type:        schema.TypeString,
									Required:    true,
								},
								"filter": {
									Description:      "Query used to limit documents the alias can access.",
									Type:             schema.TypeString,
									Optional:         true,
									DiffSuppressFunc: utils.DiffJsonSuppress,
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
						Description:      "Mapping for fields in the index.",
						Type:             schema.TypeString,
						Optional:         true,
						DiffSuppressFunc: utils.DiffJsonSuppress,
						ValidateFunc:     validation.StringIsJSON,
					},
					"settings": {
						Description:      "Configuration options for the index. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings",
						Type:             schema.TypeString,
						Optional:         true,
						DiffSuppressFunc: utils.DiffIndexSettingSuppress,
						ValidateFunc:     validation.StringIsJSON,
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

	utils.AddConnectionSchema(componentTemplateSchema)

	return &schema.Resource{
		Description: "Creates or updates a component template. Component templates are building blocks for constructing index templates that specify index mappings, settings, and aliases. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-component-template.html",

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

func resourceComponentTemplatePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	componentId := d.Get("name").(string)
	id, diags := client.ID(ctx, componentId)
	if diags.HasError() {
		return diags
	}
	var componentTemplate models.ComponentTemplate
	componentTemplate.Name = componentId

	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		componentTemplate.Meta = metadata
	}

	if v, ok := d.GetOk("template"); ok {
		// only one template block allowed to be declared
		definedTempl := v.([]interface{})[0].(map[string]interface{})
		definedAliases := definedTempl["alias"].(*schema.Set)
		templ := models.Template{}

		aliases := make(map[string]models.IndexAlias, definedAliases.Len())
		for _, a := range definedAliases.List() {
			alias := a.(map[string]interface{})
			templAlias := models.IndexAlias{}

			if f, ok := alias["filter"]; ok {
				if f.(string) != "" {
					filterMap := make(map[string]interface{})
					if err := json.Unmarshal([]byte(f.(string)), &filterMap); err != nil {
						return diag.FromErr(err)
					}
					templAlias.Filter = filterMap
				}
			}
			if ir, ok := alias["index_routing"]; ok {
				templAlias.IndexRouting = ir.(string)
			}
			templAlias.IsHidden = alias["is_hidden"].(bool)
			templAlias.IsWriteIndex = alias["is_write_index"].(bool)
			if r, ok := alias["routing"]; ok {
				templAlias.Routing = r.(string)
			}
			if sr, ok := alias["search_routing"]; ok {
				templAlias.SearchRouting = sr.(string)
			}

			aliases[alias["name"].(string)] = templAlias
		}
		templ.Aliases = aliases

		if mappings, ok := definedTempl["mappings"]; ok {
			if mappings.(string) != "" {
				maps := make(map[string]interface{})
				if err := json.Unmarshal([]byte(mappings.(string)), &maps); err != nil {
					return diag.FromErr(err)
				}
				templ.Mappings = maps
			}
		}

		if settings, ok := definedTempl["settings"]; ok {
			if settings.(string) != "" {
				sets := make(map[string]interface{})
				if err := json.Unmarshal([]byte(settings.(string)), &sets); err != nil {
					return diag.FromErr(err)
				}
				templ.Settings = sets
			}
		}

		componentTemplate.Template = &templ
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

func resourceComponentTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
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

func resourceComponentTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}
	if diags := elasticsearch.DeleteComponentTemplate(ctx, client, compId.ResourceId); diags.HasError() {
		return diags
	}
	return diags
}
