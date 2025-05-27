package security

import (
	"context"
	"fmt"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewSystemUserResource() resource.Resource {
	return &systemUserResource{}
}

type systemUserResource struct {
	client *clients.ApiClient
}

func (r *systemUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_user"
}

func (r *systemUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Updates system user's password and enablement. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/built-in-users.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "An identifier for the system user (see https://www.elastic.co/guide/en/elasticsearch/reference/current/built-in-users.html).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[[:graph:]]+$`), "must contain alphanumeric characters (a-z, A-Z, 0-9), spaces, punctuation, and printable symbols in the Basic Latin (ASCII) block. Leading or trailing whitespace is not allowed"),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The user's password. Passwords must be at least 6 characters long.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 128),
				},
			},
			"password_hash": schema.StringAttribute{
				MarkdownDescription: "A hash of the user's password. This must be produced using the same hashing algorithm as has been configured for password storage (see https://www.elastic.co/guide/en/elasticsearch/reference/current/security-settings.html#hashing-settings).",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 128),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the user is enabled. The default value is true.",
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

func (r *systemUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *systemUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	diags := r.update(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *systemUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SystemUserData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	usernameId := compId.ResourceId

	user, sdkDiags := elasticsearch.GetUser(ctx, r.client, usernameId)
	diags = utils.ConvertSDKDiagnosticsToFramework(sdkDiags)
	if diags == nil && (user == nil || !user.IsSystemUser()) {
		tflog.Warn(ctx, fmt.Sprintf(`System user "%s" not found, removing from state`, compId.ResourceId))
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Username = types.StringValue(usernameId)
	data.Enabled = types.BoolValue(user.Enabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *systemUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := r.update(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *systemUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SystemUserData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Warn(ctx, fmt.Sprintf(`System user '%s' is not deletable, just removing from state`, compId.ResourceId))
}

type SystemUserData struct {
	Id           types.String `tfsdk:"id"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	PasswordHash types.String `tfsdk:"password_hash"`
	Enabled      types.Bool   `tfsdk:"enabled"`
}

func (r *systemUserResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var data SystemUserData
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &data)...)
	if diags.HasError() {
		return diags
	}

	usernameId := data.Username.ValueString()
	id, sdkDiags := r.client.ID(ctx, usernameId)
	diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	user, sdkDiags := elasticsearch.GetUser(ctx, r.client, usernameId)
	diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}
	if user == nil || !user.IsSystemUser() {
		diags.AddError("", fmt.Sprintf(`System user "%s" not found`, usernameId))
		return diags
	}

	var userPassword models.UserPassword
	if utils.IsKnown(data.Password) && (user.Password == nil || data.Password.ValueString() != *user.Password) {
		userPassword.Password = data.Password.ValueStringPointer()
	}
	if utils.IsKnown(data.PasswordHash) && (user.PasswordHash == nil || data.PasswordHash.ValueString() != *user.PasswordHash) {
		userPassword.PasswordHash = data.PasswordHash.ValueStringPointer()
	}
	if userPassword.Password != nil || userPassword.PasswordHash != nil {
		sdkDiags := elasticsearch.ChangeUserPassword(ctx, r.client, usernameId, &userPassword)
		diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		if diags.HasError() {
			return diags
		}
	}

	if utils.IsKnown(data.Enabled) && !data.Enabled.IsNull() && data.Enabled.ValueBool() != user.Enabled {
		if data.Enabled.ValueBool() {
			sdkDiags := elasticsearch.EnableUser(ctx, r.client, usernameId)
			diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		} else {
			sdkDiags := elasticsearch.DisableUser(ctx, r.client, usernameId)
			diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		}
		if diags.HasError() {
			return diags
		}
	}

	data.Id = types.StringValue(id.String())
	diags.Append(state.Set(ctx, &data)...)
	return diags
}
