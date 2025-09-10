package prebuilt_rules

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *PrebuiltRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model prebuiltRuleModel
	var priorModel prebuiltRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &priorModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverVersion, diags := r.client.ServerVersion(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	serverFlavor, diags := r.client.ServerFlavor(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = validatePrebuiltRulesServer(serverVersion, serverFlavor)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	spaceID := model.SpaceID.ValueString()

	// Check if we need to install/update rules
	if needsRuleUpdate(ctx, client, spaceID) {
		resp.Diagnostics.Append(installPrebuiltRules(ctx, client, spaceID)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Handle tag transitions for declarative behavior
	newTags, newTagDiags := model.getTags(ctx)
	resp.Diagnostics.Append(newTagDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldTags, oldTagDiags := priorModel.getTags(ctx)
	resp.Diagnostics.Append(oldTagDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use transition logic to handle tag changes declaratively
	resp.Diagnostics.Append(manageRulesTagTransition(ctx, client, spaceID, oldTags, newTags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the current status to populate computed attributes
	status, statusDiags := getPrebuiltRulesStatus(ctx, client, spaceID)
	resp.Diagnostics.Append(statusDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(model.populateFromStatus(ctx, status)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
