package security

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceUser() *schema.Resource {
	userSchema := map[string]*schema.Schema{
		"username": {
			Description: "An identifier for the user (see https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-user.html#security-api-put-user-path-params).",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 1024),
				validation.StringMatch(regexp.MustCompile(`^[[:graph:]]+$`), "must contain alphanumeric characters (a-z, A-Z, 0-9), spaces, punctuation, and printable symbols in the Basic Latin (ASCII) block. Leading or trailing whitespace is not allowed"),
			),
		},
		"password": {
			Description:   "The user’s password. Passwords must be at least 6 characters long.",
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			ValidateFunc:  validation.StringLenBetween(6, 128),
			ConflictsWith: []string{"password_hash"},
		},
		"password_hash": {
			Description:   "A hash of the user’s password. This must be produced using the same hashing algorithm as has been configured for password storage (see https://www.elastic.co/guide/en/elasticsearch/reference/current/security-settings.html#hashing-settings).",
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			ValidateFunc:  validation.StringLenBetween(6, 128),
			ConflictsWith: []string{"password"},
		},
		"full_name": {
			Description: "The full name of the user.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
		},
		"email": {
			Description: "The email of the user.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
		},
		"roles": {
			Description: "A set of roles the user has. The roles determine the user’s access permissions. Default is [].",
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"metadata": {
			Description:      "Arbitrary metadata that you want to associate with the user.",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
		"enabled": {
			Description: "Specifies whether the user is enabled. The default value is true.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
	}

	utils.AddConnectionSchema(userSchema)

	return &schema.Resource{
		Description: "Adds and updates users in the native realm. These users are commonly referred to as native users. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-user.html",

		CreateContext: resourceSecurityUserPut,
		UpdateContext: resourceSecurityUserPut,
		ReadContext:   resourceSecurityUserRead,
		DeleteContext: resourceSecurityUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: userSchema,
	}
}

func resourceSecurityUserPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	usernameId := d.Get("username").(string)
	id, diags := client.ID(usernameId)
	if diags.HasError() {
		return diags
	}

	var user models.User
	if pass, ok := d.GetOk("password"); ok {
		password := pass.(string)
		user.Password = &password
	} else {
		pass_hash := d.Get("password_hash").(string)
		user.PasswordHash = &pass_hash
	}

	user.Email = d.Get("email").(string)
	user.FullName = d.Get("full_name").(string)
	user.Enabled = d.Get("enabled").(bool)

	roles := make([]string, 0)
	if v, ok := d.GetOk("roles"); ok {
		for _, role := range v.(*schema.Set).List() {
			roles = append(roles, role.(string))
		}
	}
	user.Roles = roles

	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		user.Metadata = metadata
	}

	// prepare request
	userBytes, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(err)
	}
	// create / update the user
	res, err := client.Security.PutUser(usernameId, bytes.NewReader(userBytes))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update a user"); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceSecurityUserRead(ctx, d, meta)
}

func resourceSecurityUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	usernameId := compId.ResourceId

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
	if err := d.Set("username", users[usernameId].Username); err != nil {
		return diag.FromErr(err)
	}
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

	return diags
}

func resourceSecurityUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := clients.NewApiClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	res, err := client.Security.DeleteUser(compId.ResourceId)
	if err != nil && res.IsError() {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete a user"); diags.HasError() {
		return diags
	}

	d.SetId("")
	return diags
}
