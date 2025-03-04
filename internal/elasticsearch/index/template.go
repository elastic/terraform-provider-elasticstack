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

func ResourceTemplate() *schema.Resource {
	templateSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
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
						Description: "If true, the data stream is hidden. Defaults to `false`. Available only in **8.x**",
						Type:        schema.TypeBool,
						Default:     false,
						Optional:    true,
					},
					"allow_custom_routing": {
						Description: "If `true`, the data stream supports custom routing. Defaults to `false`. Available only in **8.x**",
						Type:        schema.TypeBool,
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
			Description: "Template to be applied. It may optionally include an aliases, mappings, lifecycle, or settings configuration.",
			Type:        schema.TypeList,
			Optional:    true,
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
									Description: "The alias name.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"filter": {
									Description:      "Query used to limit documents the alias can access.",
									Type:             schema.TypeString,
									Optional:         true,
									Default:          "",
									DiffSuppressFunc: utils.DiffJsonSuppress,
									ValidateFunc:     validation.StringIsJSON,
								},
								"index_routing": {
									Description: "Value used to route indexing operations to a specific shard. If specified, this overwrites the `routing` value for indexing operations.",
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "",
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
									Default:     "",
								},
								"search_routing": {
									Description: "Value used to route search operations to a specific shard. If specified, this overwrites the routing value for search operations.",
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "",
								},
							},
						},
					},
					"mappings": {
						Description:      "Mapping for fields in the index. Should be specified as a JSON object of field mappings. See the documentation (https://www.elastic.co/guide/en/elasticsearch/reference/current/explicit-mapping.html) for more details",
						Type:             schema.TypeString,
						Optional:         true,
						DiffSuppressFunc: utils.DiffJsonSuppress,
						ValidateFunc: validation.All(
							validation.StringIsJSON, stringIsJSONObject,
						),
					},
					"settings": {
						Description:      "Configuration options for the index. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings",
						Type:             schema.TypeString,
						Optional:         true,
						DiffSuppressFunc: utils.DiffIndexSettingSuppress,
						ValidateFunc: validation.All(
							validation.StringIsJSON, stringIsJSONObject,
						),
					},
					"lifecycle": {
						Description: "Lifecycle of data stream. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/data-stream-lifecycle.html",
						Type:        schema.TypeSet,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"data_retention": {
									Description: "The retention period of the data indexed in this data stream.",
									Type:        schema.TypeString,
									Required:    true,
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
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	templateId := d.Get("name").(string)
	id, diags := client.ID(ctx, templateId)
	if diags.HasError() {
		return diags
	}
	var indexTemplate models.IndexTemplate
	indexTemplate.Name = templateId

	compsOf := make([]string, 0)
	if v, ok := d.GetOk("composed_of"); ok {
		for _, c := range v.([]interface{}) {
			compsOf = append(compsOf, c.(string))
		}
	}
	indexTemplate.ComposedOf = compsOf

	if v, ok := d.GetOk("data_stream"); ok {
		// 8.x workaround
		hasAllowCustomRouting := false
		// 7.11.x workaround
		hasHidden := false
		if d.HasChange("data_stream") {
			old, _ := d.GetChange("data_stream")

			if old != nil && len(old.([]interface{})) == 1 {
				if old.([]interface{})[0] != nil {
					setting := old.([]interface{})[0].(map[string]interface{})
					if acr, ok := setting["allow_custom_routing"]; ok && acr.(bool) {
						hasAllowCustomRouting = true
					}
					if h, ok := setting["hidden"]; ok && h.(bool) {
						hasHidden = true
					}
				}
			}
		}

		// only one definition of stream allowed
		if v.([]interface{})[0] != nil {
			stream := v.([]interface{})[0].(map[string]interface{})
			dSettings := &models.DataStreamSettings{}
			if s, ok := stream["hidden"]; ok && (hasHidden || s.(bool)) {
				hidden := s.(bool)
				dSettings.Hidden = &hidden
			}
			if s, ok := stream["allow_custom_routing"]; ok && (hasAllowCustomRouting || s.(bool)) {
				allow := s.(bool)
				dSettings.AllowCustomRouting = &allow
			}
			indexTemplate.DataStream = dSettings
		}
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
		templ, ok, diags := expandTemplate(v)
		if diags != nil {
			return diags
		}

		if ok {
			indexTemplate.Template = &templ
		}
	}

	if v, ok := d.GetOk("version"); ok {
		definedVer := v.(int)
		indexTemplate.Version = &definedVer
	}

	if diags := elasticsearch.PutIndexTemplate(ctx, client, &indexTemplate); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceIndexTemplateRead(ctx, d, meta)
}

func expandTemplate(config interface{}) (models.Template, bool, diag.Diagnostics) {
	templ := models.Template{}
	// only one template block allowed to be declared
	definedTempl, ok := config.([]interface{})[0].(map[string]interface{})
	if !ok {
		return templ, false, nil
	}

	aliases, diags := ExpandIndexAliases(definedTempl["alias"].(*schema.Set))
	if diags.HasError() {
		return templ, false, diags
	}
	templ.Aliases = aliases

	if lc, ok := definedTempl["lifecycle"]; ok {
		lifecycle := ExpandLifecycle(lc.(*schema.Set))
		if lifecycle != nil {
			templ.Lifecycle = lifecycle
		}
	}

	if mappings, ok := definedTempl["mappings"]; ok {
		if mappings.(string) != "" {
			maps := make(map[string]interface{})
			if err := json.Unmarshal([]byte(mappings.(string)), &maps); err != nil {
				return templ, false, diag.FromErr(err)
			}
			templ.Mappings = maps
		}
	}

	if settings, ok := definedTempl["settings"]; ok {
		if settings.(string) != "" {
			sets := make(map[string]interface{})
			if err := json.Unmarshal([]byte(settings.(string)), &sets); err != nil {
				return templ, false, diag.FromErr(err)
			}
			templ.Settings = sets
		}
	}

	return templ, true, nil
}

func resourceIndexTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	templateId := compId.ResourceId

	tpl, diags := elasticsearch.GetIndexTemplate(ctx, client, templateId)
	if tpl == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Index template "%s" not found, removing from state`, compId.ResourceId))
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
	if err := d.Set("composed_of", tpl.IndexTemplate.ComposedOf); err != nil {
		return diag.FromErr(err)
	}
	if stream := tpl.IndexTemplate.DataStream; stream != nil {
		ds := make([]interface{}, 1)
		dSettings := make(map[string]interface{})
		if v := stream.Hidden; v != nil {
			dSettings["hidden"] = *v
		}
		if v := stream.AllowCustomRouting; v != nil {
			dSettings["allow_custom_routing"] = *v
		}
		ds[0] = dSettings
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
		aliases, diags := FlattenIndexAliases(template.Aliases)
		if diags.HasError() {
			return nil, diags
		}
		tmpl["alias"] = aliases
	}

	if template.Lifecycle != nil {
		lifecycle := FlattenLifecycle(template.Lifecycle)
		tmpl["lifecycle"] = lifecycle
	}

	return []interface{}{tmpl}, diags
}

func resourceIndexTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}
	if diags := elasticsearch.DeleteIndexTemplate(ctx, client, compId.ResourceId); diags.HasError() {
		return diags
	}
	return diags
}
