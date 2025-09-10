package prebuilt_rules

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *PrebuiltRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model prebuiltRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
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

	spaceID := model.ID.ValueString()

	// Get current status
	status, statusDiags := getPrebuiltRulesStatus(ctx, client, spaceID)
	resp.Diagnostics.Append(statusDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update computed values from status
	resp.Diagnostics.Append(model.populateFromStatus(ctx, status)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
