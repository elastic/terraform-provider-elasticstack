package security

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
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
		Description: "Get the information about the user in the ES cluster. See the [security API get user documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-user.html) for more details.",

		ReadContext: dataSourceSecurityUserRead,

		Schema: userSchema,
	}
}

func dataSourceSecurityUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	usernameId := d.Get("username").(string)
	id, diags := client.ID(ctx, usernameId)
	if diags.HasError() {
		return diags
	}
	d.SetId(id.String())

	user, diags := elasticsearch.GetUser(ctx, client, usernameId)
	if user == nil && diags == nil {
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	metadata, err := json.Marshal(user.Metadata)
	if err != nil {
		return diag.FromErr(err)
	}

	// set the fields
	if err := d.Set("username", usernameId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("full_name", user.FullName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("roles", user.Roles); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("metadata", string(metadata)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", user.Enabled); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
