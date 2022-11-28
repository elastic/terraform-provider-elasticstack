package security

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceRoleMapping() *schema.Resource {
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
			Computed:    true,
			Description: "Mappings that have `enabled` set to `false` are ignored when role mapping is performed.",
		},
		"rules": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The rules that determine which users should be matched by the mapping. A rule is a logical condition that is expressed by using a JSON DSL.",
		},
		"roles": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Computed:    true,
			Description: "A list of role names that are granted to the users that match the role mapping rules.",
		},
		"role_templates": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "A list of mustache templates that will be evaluated to determine the roles names that should granted to the users that match the role mapping rules.",
		},
		"metadata": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Additional metadata that helps define which roles are assigned to each user. Keys beginning with `_` are reserved for system usage.",
		},
	}

	utils.AddConnectionSchema(roleMappingSchema)

	return &schema.Resource{
		Description: "Retrieves role mappings. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role-mapping.html",
		ReadContext: dataSourceSecurityRoleMappingRead,
		Schema:      roleMappingSchema,
	}
}

func dataSourceSecurityRoleMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	roleId := d.Get("name").(string)
	id, diags := client.ID(ctx, roleId)
	if diags.HasError() {
		return diags
	}
	d.SetId(id.String())

	return resourceSecurityRoleMappingRead(ctx, d, meta)
}
