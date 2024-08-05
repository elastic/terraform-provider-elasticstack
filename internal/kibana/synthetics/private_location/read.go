package private_location

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {

	tflog.Info(ctx, "Read private location")

	// TODO: dry
	if !r.resourceReady(&response.Diagnostics) {
		return
	}

	kibanaClient, err := r.client.GetKibanaClient()
	if err != nil {
		response.Diagnostics.AddError("unable to get kibana client", err.Error())
		return
	}

	//TODO: handle 404 error / remove resource

	var state tfModelV0
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	label := state.Label.ValueString()
	namespace := state.SpaceID.ValueString()
	result, err := kibanaClient.KibanaSynthetics.PrivateLocation.Get(label, namespace)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to get private location `%s`, namespace %s", label, namespace), err.Error())
		return
	}

	state = toModelV0(*result)

	// Set refreshed state
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
