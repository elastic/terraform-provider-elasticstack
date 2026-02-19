package security_enable_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *EnableRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model enableRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	minVersion := version.Must(version.NewVersion("8.11.0"))
	if serverVersion.LessThan(minVersion) {
		resp.Diagnostics.AddError("Unsupported server version", "Security detection rules bulk actions are not supported until Elastic Stack v8.11.0. Upgrade the target server to use this resource")
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Failed to get Kibana client")
		return
	}

	spaceID := model.SpaceID.ValueString()
	key := model.Key.ValueString()
	value := model.Value.ValueString()

	allEnabled, checkDiags := kibana_oapi.CheckRulesEnabledByTag(ctx, client, spaceID, key, value)
	resp.Diagnostics.Append(checkDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !allEnabled {
		tflog.Info(ctx, "Drift detected: some rules matching the tag are disabled, re-enabling them", map[string]interface{}{
			"space_id": spaceID,
			"key":      key,
			"value":    value,
		})

		resp.Diagnostics.Append(kibana_oapi.EnableRulesByTag(ctx, client, spaceID, key, value)...)
		if resp.Diagnostics.HasError() {
			return
		}

		model.AllRulesEnabled = types.BoolValue(true)
	} else {
		tflog.Debug(ctx, "All rules matching the tag are enabled", map[string]interface{}{
			"space_id": spaceID,
			"key":      key,
			"value":    value,
		})
		model.AllRulesEnabled = types.BoolValue(true)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
