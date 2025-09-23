package system_user

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *systemUserResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var data SystemUserData
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &data)...)
	if diags.HasError() {
		return diags
	}

	usernameId := data.Username.ValueString()
	id, sdkDiags := r.client.ID(ctx, usernameId)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(diags...)
	if diags.HasError() {
		return diags
	}

	user, sdkDiags := elasticsearch.GetUser(ctx, client, usernameId)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
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
		diags.Append(elasticsearch.ChangeUserPassword(ctx, r.client, usernameId, &userPassword)...)
		if diags.HasError() {
			return diags
		}
	}

	if utils.IsKnown(data.Enabled) && !data.Enabled.IsNull() && data.Enabled.ValueBool() != user.Enabled {
		if data.Enabled.ValueBool() {
			diags.Append(elasticsearch.EnableUser(ctx, r.client, usernameId)...)
		} else {
			diags.Append(elasticsearch.DisableUser(ctx, r.client, usernameId)...)
		}
		if diags.HasError() {
			return diags
		}
	}

	data.Id = types.StringValue(id.String())
	diags.Append(state.Set(ctx, &data)...)
	return diags
}

func (r *systemUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := r.update(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
