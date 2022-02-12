package security

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
)

func ResourceRoleMapping() *schema.Resource {
	roleMappingSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the role mapping.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"enabled": {
			Description: "Specifies whether the role mapping is enabled. The default value is true.",
			Type:        schema.TypeBool,
			Required:    true,
		},
		"roles": {
			Description: " A list of role names that are granted to the users that match the role mapping rule. Default is [].",
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"role_templates": {
			Description: "A list of mustache templates that will be evaluated to determine the roles names that should granted to the users that match the role mapping rules. Each record must be a valid JSON document.",
			Type:        schema.TypeList,
			Optional:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsJSON,
			},
		},
		"rules": {
			Description:  "The rules that determine which users should be matched by the mapping.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsJSON,
		},
		"metadata": {
			Description:      "Optional meta-data.",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
	}

	utils.AddConnectionSchema(roleMappingSchema)

	return &schema.Resource{
		Description: "Manages the adding and updating of role mappings. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api.html#security-role-mapping-apis",

		CreateContext: resourceSecurityRoleMappingPut,
		UpdateContext: resourceSecurityRoleMappingPut,
		ReadContext:   resourceSecurityRoleMappingRead,
		DeleteContext: resourceSecurityRoleMappingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: roleMappingSchema,
	}
}

func resourceSecurityRoleMappingPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	roleMappingID := d.Get("name").(string)
	id, diags := client.ID(roleMappingID)
	if diags.HasError() {
		return diags
	}
	var roleMapping models.RoleMapping
	roleMapping.Name = roleMappingID

	if v, ok := d.GetOk("enabled"); ok {
		r := v.(bool)
		roleMapping.Enabled = r
	}

	roles := make([]string, 0)
	if v, ok := d.GetOk("roles"); ok {
		p := v.(*schema.Set)
		for _, e := range p.List() {
			roles = append(roles, e.(string))
		}
	}
	roleMapping.Roles = roles

	if v, ok := d.GetOk("role_templates"); ok {
		roleTemplates := make([]map[string]interface{}, len(v.([]interface{})))
		for i, f := range v.([]interface{}) {
			item := make(map[string]interface{})
			if err := json.NewDecoder(strings.NewReader(f.(string))).Decode(&item); err != nil {
				return diag.FromErr(err)
			}
			roleTemplates[i] = item
		}
		roleMapping.RoleTemplates = roleTemplates
	}

	if v, ok := d.GetOk("rules"); ok {
		rules := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&rules); err != nil {
			return diag.FromErr(err)
		}
		roleMapping.Rules = rules
	}

	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		roleMapping.Metadata = metadata
	}

	debugMapping, _ := json.Marshal(&roleMapping)
	fmt.Println(string(debugMapping))

	if diags := client.PutElasticsearchRoleMapping(&roleMapping); diags.HasError() {
		return diags
	}
	d.SetId(id.String())
	return resourceSecurityRoleMappingRead(ctx, d, meta)
}

func resourceSecurityRoleMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	roleMapping, diags := client.GetElasticsearchRoleMapping(compId.ResourceId)
	if roleMapping == nil && diags == nil {
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	if err := d.Set("name", roleMapping.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("enabled", roleMapping.Enabled); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("roles", roleMapping.Roles); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("roles", roleMapping.Roles); err != nil {
		return diag.FromErr(err)
	}

	if roleTemplates := roleMapping.RoleTemplates; roleTemplates != nil {
		fProcs := make([]string, len(roleTemplates))
		for i, v := range roleTemplates {
			res, err := json.Marshal(v)
			if err != nil {
				return diag.FromErr(err)
			}
			fProcs[i] = string(res)
		}

		if err := d.Set("role_templates", fProcs); err != nil {
			return diag.FromErr(err)
		}
	}

	if roleMapping.Rules != nil {
		rules, err := json.Marshal(roleMapping.Rules)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("rules", string(rules)); err != nil {
			return diag.FromErr(err)
		}
	}

	if meta := roleMapping.Metadata; meta != nil {
		meta, err := json.Marshal(meta)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("metadata", string(meta)); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceSecurityRoleMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	if diags := client.DeleteElasticsearchRoleMapping(compId.ResourceId); diags.HasError() {
		return diags
	}

	d.SetId("")
	return diags
}
