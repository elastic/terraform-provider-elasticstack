package prebuilt_rules

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *PrebuiltRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model prebuiltRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
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

	// Install/update prebuilt rules and timelines
	resp.Diagnostics.Append(installPrebuiltRules(ctx, client, spaceID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Enable/disable rules based on tags if specified
	tags, tagDiags := model.getTags(ctx)
	resp.Diagnostics.Append(tagDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(tags) > 0 {
		resp.Diagnostics.Append(manageRulesByTags(ctx, client, spaceID, tags)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Set the resource ID to the space ID
	model.ID = model.SpaceID

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
