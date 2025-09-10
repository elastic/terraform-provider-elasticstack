package prebuilt_rules

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *PrebuiltRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model prebuiltRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isSupported, sdkDiags := r.client.EnforceMinVersion(ctx, version.Must(version.NewVersion("8.0.0")))
	resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !isSupported {
		resp.Diagnostics.AddError("Unsupported server version", "Prebuilt rules are not supported until Elastic Stack v8.0.0. Upgrade the target server to use this resource")
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
