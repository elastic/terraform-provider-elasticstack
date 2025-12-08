package prebuilt_rules

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *PrebuiltRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model prebuiltRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	minVersion := version.Must(version.NewVersion("8.0.0"))
	if serverVersion.LessThan(minVersion) {
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
	status, statusDiags := kibana_oapi.GetPrebuiltRulesStatus(ctx, client, spaceID)
	resp.Diagnostics.Append(statusDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.populateFromStatus(status)

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
