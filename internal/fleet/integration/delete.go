package integration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *integrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var stateModel integrationModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	name := stateModel.Name.ValueString()
	version := stateModel.Version.ValueString()
	force := stateModel.Force.ValueBool()
	skipDestroy := stateModel.SkipDestroy.ValueBool()
	if skipDestroy {
		tflog.Debug(ctx, "Skipping uninstall of integration package", map[string]any{"name": name, "version": version})
		return
	}

	diags = fleet.Uninstall(ctx, client, name, version, force)
	resp.Diagnostics.Append(diags...)
}
