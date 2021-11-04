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

func DataSourceUser() *schema.Resource {
	userSchema := map[string]*schema.Schema{
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
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	usernameId := d.Get("username").(string)
	id, diags := client.ID(usernameId)
	if diags.HasError() {
		return diags
	}

	// create request and run it
	req := client.Security.GetUser.WithUsername(usernameId)
	res, err := client.Security.GetUser(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to get a user."); diags.HasError() {
		return diags
	}

	// unmarshal our response to proper type
	users := make(map[string]models.User)
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		return diag.FromErr(err)
	}
	metadata, err := json.Marshal(users[usernameId].Metadata)
	if err != nil {
		return diag.FromErr(err)
	}

	// set the fields
	if err := d.Set("email", users[usernameId].Email); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("full_name", users[usernameId].FullName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("roles", users[usernameId].Roles); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("metadata", string(metadata)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", users[usernameId].Enabled); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id.String())
	return diags
}
