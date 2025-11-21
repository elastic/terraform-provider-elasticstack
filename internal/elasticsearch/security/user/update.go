package user

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *userResource) update(ctx context.Context, plan tfsdk.Plan, config tfsdk.Config, state *tfsdk.State) diag.Diagnostics {
	var planData UserData
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &planData)...)
	if diags.HasError() {
		return diags
	}

	// Check if we have existing state (this is an update, not a create)
	hasState := false
	var stateData UserData
	if state != nil && !state.Raw.IsNull() {
		hasState = true
		diags.Append(state.Get(ctx, &stateData)...)
		if diags.HasError() {
			return diags
		}
	}

	usernameId := planData.Username.ValueString()
	id, sdkDiags := r.client.ID(ctx, usernameId)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	client, connDiags := clients.MaybeNewApiClientFromFrameworkResource(ctx, planData.ElasticsearchConnection, r.client)
	diags.Append(connDiags...)
	if diags.HasError() {
		return diags
	}

	var user models.User
	user.Username = usernameId

	// Handle password fields - only set password if it's in the plan AND (it's a create OR it has changed from state)
	// Priority: password_wo > password > password_hash
	// Read password_wo from config as per Terraform write-only attribute guidelines
	var passwordWoFromConfig types.String
	diags.Append(config.GetAttribute(ctx, path.Root("password_wo"), &passwordWoFromConfig)...)
	if diags.HasError() {
		return diags
	}

	if utils.IsKnown(passwordWoFromConfig) && (!hasState || !planData.PasswordWoVersion.Equal(stateData.PasswordWoVersion)) {
		// Use write-only password - changes triggered by version change
		password := passwordWoFromConfig.ValueString()
		user.Password = &password
	} else if utils.IsKnown(planData.Password) && (!hasState || !planData.Password.Equal(stateData.Password)) {
		password := planData.Password.ValueString()
		user.Password = &password
	} else if utils.IsKnown(planData.PasswordHash) && (!hasState || !planData.PasswordHash.Equal(stateData.PasswordHash)) {
		passwordHash := planData.PasswordHash.ValueString()
		user.PasswordHash = &passwordHash
	}

	if utils.IsKnown(planData.Email) {
		user.Email = planData.Email.ValueString()
	}
	if utils.IsKnown(planData.FullName) {
		user.FullName = planData.FullName.ValueString()
	}
	user.Enabled = planData.Enabled.ValueBool()

	roles := make([]string, 0, len(planData.Roles.Elements()))
	diags.Append(planData.Roles.ElementsAs(ctx, &roles, false)...)
	if diags.HasError() {
		return diags
	}
	user.Roles = roles

	if !planData.Metadata.IsNull() && !planData.Metadata.IsUnknown() {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(planData.Metadata.ValueString())).Decode(&metadata); err != nil {
			diags.AddError("Failed to decode metadata", err.Error())
			return diags
		}
		user.Metadata = metadata
	}

	diags.Append(elasticsearch.PutUser(ctx, client, &user)...)
	if diags.HasError() {
		return diags
	}

	// Read the user back to get computed fields like metadata
	readUser, sdkDiags := elasticsearch.GetUser(ctx, client, usernameId)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	planData.Id = types.StringValue(id.String())

	// Set computed fields from the API response
	if len(readUser.Metadata) > 0 {
		metadata, err := json.Marshal(readUser.Metadata)
		if err != nil {
			diags.AddError("Failed to marshal metadata", err.Error())
			return diags
		}
		planData.Metadata = jsontypes.NewNormalizedValue(string(metadata))
	} else {
		planData.Metadata = jsontypes.NewNormalizedNull()
	}

	diags.Append(state.Set(ctx, &planData)...)
	return diags
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(r.update(ctx, req.Plan, req.Config, &resp.State)...)
}
