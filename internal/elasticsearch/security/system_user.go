package security

import (
	"context"
	"fmt"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceSystemUser() *schema.Resource {
	userSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"username": {
			Description: "An identifier for the system user (see https://www.elastic.co/guide/en/elasticsearch/reference/current/built-in-users.html).",
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
		"enabled": {
			Description: "Specifies whether the user is enabled. The default value is true.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
	}

	utils.AddConnectionSchema(userSchema)

	return &schema.Resource{
		Description: "Updates system user's password and enablement. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/built-in-users.html",

		CreateContext: resourceSecuritySystemUserPut,
		UpdateContext: resourceSecuritySystemUserPut,
		ReadContext:   resourceSecuritySystemUserRead,
		DeleteContext: resourceSecuritySystemUserDelete,

		Schema: userSchema,
	}
}

func resourceSecuritySystemUserPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	usernameId := d.Get("username").(string)
	id, diags := client.ID(ctx, usernameId)
	if diags.HasError() {
		return diags
	}

	user, diags := elasticsearch.GetUser(ctx, client, usernameId)
	if diags.HasError() {
		return diags
	}
	if user == nil || !user.IsSystemUser() {
		return diag.Errorf(`System user "%s" not found`, usernameId)
	}

	var userPassword models.UserPassword
	if v, ok := d.GetOk("password"); ok && d.HasChange("password") {
		password := v.(string)
		userPassword.Password = &password
	}
	if v, ok := d.GetOk("password_hash"); ok && d.HasChange("password_hash") {
		pass_hash := v.(string)
		userPassword.PasswordHash = &pass_hash
	}
	if userPassword.Password != nil || userPassword.PasswordHash != nil {
		if diags := elasticsearch.ChangeUserPassword(ctx, client, usernameId, &userPassword); diags.HasError() {
			return diags
		}
	}

	if d.HasChange("enabled") {
		if d.Get("enabled").(bool) {
			if diags := elasticsearch.EnableUser(ctx, client, usernameId); diags.HasError() {
				return diags
			}
		} else {
			if diags := elasticsearch.DisableUser(ctx, client, usernameId); diags.HasError() {
				return diags
			}
		}
	}

	d.SetId(id.String())
	return resourceSecuritySystemUserRead(ctx, d, meta)
}

func resourceSecuritySystemUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	usernameId := compId.ResourceId

	user, diags := elasticsearch.GetUser(ctx, client, usernameId)
	if diags == nil && (user == nil || !user.IsSystemUser()) {
		tflog.Warn(ctx, fmt.Sprintf(`System user "%s" not found, removing from state`, compId.ResourceId))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	if err := d.Set("username", usernameId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", user.Enabled); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSecuritySystemUserDelete(ctx context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	tflog.Warn(ctx, fmt.Sprintf(`System user '%s' is not deletable, just removing from state`, compId.ResourceId))
	return nil
}
