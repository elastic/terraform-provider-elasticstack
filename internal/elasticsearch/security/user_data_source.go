package security

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceUser() *schema.Resource {
	userSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"username": {
			Description: "An identifier for the user",
			Type:        schema.TypeString,
			Required:    true,
		},
		"full_name": {
			Description: "The full name of the user.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"email": {
			Description: "The email of the user.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"roles": {
			Description: "A set of roles the user has. The roles determine the user’s access permissions. Default is [].",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"metadata": {
			Description: "Arbitrary metadata that you want to associate with the user.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enabled": {
			Description: "Specifies whether the user is enabled. The default value is true.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}

	utils.AddConnectionSchema(userSchema)

	return &schema.Resource{
		Description: "Get the information about the user in the ES cluster. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-user.html",

		ReadContext: dataSourceSecurityUserRead,

		Schema: userSchema,
	}
}

func dataSourceSecurityUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	usernameId := d.Get("username").(string)
	id, diags := client.ID(ctx, usernameId)
	if diags.HasError() {
		return diags
	}
	d.SetId(id.String())

	return resourceSecurityUserRead(ctx, d, meta)
}
