package index

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceTemplate() *schema.Resource {
	templateSchema := map[string]*schema.Schema{
		"name": {
			Description: "Name of the index template to create.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"composed_of": {
			Description: "An ordered list of component template names.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"data_stream": {
			Description: "If this object is included, the template is used to create data streams and their backing indices. Supports an empty object.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"hidden": {
						Description: "If true, the data stream is hidden.",
						Type:        schema.TypeBool,
						Default:     false,
						Optional:    true,
					},
				},
			},
		},
		"index_patterns": {
			Description: "Array of wildcard (*) expressions used to match the names of data streams and indices during creation.",
			Type:        schema.TypeSet,
			Required:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"metadata": {
			Description:      "Optional user metadata about the index template.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
		"priority": {
			Description:  "Priority to determine index template precedence when a new data stream or index is created.",
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Optional:     true,
		},
		"template": {
			Description: "Template to be applied. It may optionally include an aliases, mappings, or settings configuration.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"aliases": {
						Description: "Aliases to add.",
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
						DiffSuppressFunc: utils.DiffIndexTemplateSettingSuppress,
						ValidateFunc:     validation.StringIsJSON,
					},
				},
			},
		},
		"version": {
			Description: "Version number used to manage index templates externally.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
	}

	utils.AddConnectionSchema(templateSchema)

	return &schema.Resource{
		Description: "Creates or updates an index template. Index templates define settings, mappings, and aliases that can be applied automatically to new indices. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-put-template.html",

		CreateContext: resourceIndexTemplatePut,
		UpdateContext: resourceIndexTemplatePut,
		ReadContext:   resourceIndexTemplateRead,
		DeleteContext: resourceIndexTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: templateSchema,
	}
}

func resourceIndexTemplatePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	templateId := d.Get("name").(string)
	id, diags := client.ID(templateId)
	if diags.HasError() {
		return diags
	}
	var indexTemplate models.IndexTemplate

	compsOf := make([]string, 0)
	if v, ok := d.GetOk("composed_of"); ok {
		for _, c := range v.([]interface{}) {
			compsOf = append(compsOf, c.(string))
		}
	}
	indexTemplate.ComposedOf = compsOf

	if v, ok := d.GetOk("data_stream"); ok {
		// only one definition of stream allowed
		stream := v.([]interface{})[0].(map[string]interface{})
		indexTemplate.DataStream = stream
	}

	if v, ok := d.GetOk("index_patterns"); ok {
		definedIndPats := v.(*schema.Set)
		indPats := make([]string, definedIndPats.Len())
		for i, p := range definedIndPats.List() {
			indPats[i] = p.(string)
		}
		indexTemplate.IndexPatterns = indPats
	}

	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		indexTemplate.Meta = metadata
	}

	if v, ok := d.GetOk("priority"); ok {
		definedPr := v.(int)
		indexTemplate.Priority = &definedPr
	}

	if v, ok := d.GetOk("template"); ok {
		// only one template block allowed to be declared
		definedTempl := v.([]interface{})[0].(map[string]interface{})
		definedAliases := definedTempl["aliases"].(*schema.Set)
		templ := models.Template{}

		aliases := make(map[string]models.TemplateAlias, definedAliases.Len())
		for _, a := range definedAliases.List() {
			alias := a.(map[string]interface{})
			templAlias := models.TemplateAlias{}

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

		indexTemplate.Template = &templ
	}

	if v, ok := d.GetOk("version"); ok {
		definedVer := v.(int)
		indexTemplate.Version = &definedVer
	}

	templateBytes, err := json.Marshal(indexTemplate)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending request to ES: %s to create template '%s' ", templateBytes, templateId)

	res, err := client.Indices.PutIndexTemplate(templateId, bytes.NewReader(templateBytes))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create index template"); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceIndexTemplateRead(ctx, d, meta)
}

func resourceIndexTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	templateId := compId.ResourceId

	req := client.Indices.GetIndexTemplate.WithName(templateId)
	res, err := client.Indices.GetIndexTemplate(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to request index template."); diags.HasError() {
		return diags
	}

	var indexTemplates models.IndexTemplatesResponse
	if err := json.NewDecoder(res.Body).Decode(&indexTemplates); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] read data from API: %+v", indexTemplates)

	// we requested only 1 template
	if len(indexTemplates.IndexTemplates) != 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong number of templates returned",
			Detail:   fmt.Sprintf("Elasticsearch API returned %d when requsted '%s' template.", len(indexTemplates.IndexTemplates), templateId),
		})
		return diags
	}
	tpl := indexTemplates.IndexTemplates[0]

	// set the fields
	if err := d.Set("name", tpl.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("composed_of", tpl.IndexTemplate.ComposedOf); err != nil {
		return diag.FromErr(err)
	}
	if tpl.IndexTemplate.DataStream != nil {
		ds := make([]interface{}, 0)
		ds = append(ds, tpl.IndexTemplate.DataStream)
		if err := d.Set("data_stream", ds); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("index_patterns", tpl.IndexTemplate.IndexPatterns); err != nil {
		return diag.FromErr(err)
	}
	if tpl.IndexTemplate.Meta != nil {
		metadata, err := json.Marshal(tpl.IndexTemplate.Meta)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("metadata", string(metadata)); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("priority", tpl.IndexTemplate.Priority); err != nil {
		return diag.FromErr(err)
	}

	if tpl.IndexTemplate.Template != nil {
		template, diags := flattenTemplateData(tpl.IndexTemplate.Template)
		if diags.HasError() {
			return diags
		}
		if err := d.Set("template", template); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("version", tpl.IndexTemplate.Version); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenTemplateData(template *models.Template) ([]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	tmpl := make(map[string]interface{})
	if template.Mappings != nil {
		m, err := json.Marshal(template.Mappings)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		tmpl["mappings"] = string(m)
	}
	if template.Settings != nil {
		s, err := json.Marshal(template.Settings)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		tmpl["settings"] = string(s)
	}

	if template.Aliases != nil {
		aliases := make([]interface{}, 0)
		for k, v := range template.Aliases {
			alias := make(map[string]interface{})
			alias["name"] = k

			if v.Filter != nil {
				f, err := json.Marshal(v.Filter)
				if err != nil {
					return nil, diag.FromErr(err)
				}
				alias["filter"] = string(f)
			}

			alias["index_routing"] = v.IndexRouting
			alias["is_hidden"] = v.IsHidden
			alias["is_write_index"] = v.IsWriteIndex
			alias["routing"] = v.Routing
			alias["search_routing"] = v.SearchRouting

			aliases = append(aliases, alias)
		}
		tmpl["aliases"] = aliases
	}

	return []interface{}{tmpl}, diags
}

func resourceIndexTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}

	res, err := client.Indices.DeleteIndexTemplate(compId.ResourceId)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete index template"); diags.HasError() {
		return diags
	}
	d.SetId("")

	return diags
}
