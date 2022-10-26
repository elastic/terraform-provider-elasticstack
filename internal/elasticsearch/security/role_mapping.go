package security

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRoleMapping() *schema.Resource {
	roleMappingSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The distinct name that identifies the role mapping, used solely as an identifier.",
		},
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Mappings that have `enabled` set to `false` are ignored when role mapping is performed.",
		},
		"rules": {
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Description:      "The rules that determine which users should be matched by the mapping. A rule is a logical condition that is expressed by using a JSON DSL.",
		},
		"roles": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description:   "A list of role names that are granted to the users that match the role mapping rules.",
			Optional:      true,
			ConflictsWith: []string{"role_templates"},
			ExactlyOneOf:  []string{"roles", "role_templates"},
		},
		"role_templates": {
			Type:             schema.TypeString,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Description:      "A list of mustache templates that will be evaluated to determine the roles names that should granted to the users that match the role mapping rules.",
			Optional:         true,
			ConflictsWith:    []string{"roles"},
			ExactlyOneOf:     []string{"roles", "role_templates"},
		},
		"metadata": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "{}",
			DiffSuppressFunc: utils.DiffJsonSuppress,
			Description:      "Additional metadata that helps define which roles are assigned to each user. Keys beginning with `_` are reserved for system usage.",
		},
	}

	utils.AddConnectionSchema(roleMappingSchema)

	return &schema.Resource{
		Description: "Manage role mappings. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html",

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
	roleMappingName := d.Get("name").(string)
	id, diags := client.ID(ctx, roleMappingName)
	if diags.HasError() {
		return diags
	}

	var rules map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("rules").(string)), &rules); err != nil {
		return diag.FromErr(err)
	}

	var roleTemplates []map[string]interface{}
	if t, ok := d.GetOk("role_templates"); ok && t.(string) != "" {
		if err := json.Unmarshal([]byte(t.(string)), &roleTemplates); err != nil {
			return diag.FromErr(err)
		}
	}

	roleMapping := models.RoleMapping{
		Name:          roleMappingName,
		Enabled:       d.Get("enabled").(bool),
		Roles:         utils.ExpandStringSet(d.Get("roles").(*schema.Set)),
		RoleTemplates: roleTemplates,
		Rules:         rules,
		Metadata:      json.RawMessage(d.Get("metadata").(string)),
	}
	if diags := client.PutElasticsearchRoleMapping(ctx, &roleMapping); diags.HasError() {
		return diags
	}
	d.SetId(id.String())

	return resourceSecurityRoleMappingRead(ctx, d, meta)
}

func resourceSecurityRoleMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceID, diags := clients.ResourceIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	roleMapping, diags := client.GetElasticsearchRoleMapping(ctx, resourceID)
	if roleMapping == nil && diags == nil {
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	rules, err := json.Marshal(roleMapping.Rules)
	if err != nil {
		diag.FromErr(err)
	}

	metadata, err := json.Marshal(roleMapping.Metadata)
	if err != nil {
		diag.FromErr(err)
	}

	if err := d.Set("name", roleMapping.Name); err != nil {
		return diag.FromErr(err)
	}
	if len(roleMapping.Roles) > 0 {
		if err := d.Set("roles", roleMapping.Roles); err != nil {
			return diag.FromErr(err)
		}
	}
	if len(roleMapping.RoleTemplates) > 0 {
		roleTemplates, err := json.Marshal(roleMapping.RoleTemplates)
		if err != nil {
			diag.FromErr(err)
		}

		if err := d.Set("role_templates", string(roleTemplates)); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("enabled", roleMapping.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("rules", string(rules)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("metadata", string(metadata)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceSecurityRoleMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceID, diags := clients.ResourceIDFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	if diags := client.DeleteElasticsearchRoleMapping(ctx, resourceID); diags.HasError() {
		return diags
	}
	return nil
}
