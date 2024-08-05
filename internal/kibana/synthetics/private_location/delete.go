package private_location

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {

	tflog.Info(ctx, "Delete private location")

	if !r.resourceReady(&response.Diagnostics) {
		return
	}

	kibanaClient, err := r.client.GetKibanaClient()
	if err != nil {
		response.Diagnostics.AddError("unable to get kibana client", err.Error())
		return
	}

	var plan tfModelV0
	diags := request.State.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()
	namespace := plan.SpaceID.ValueString()
	err = kibanaClient.KibanaSynthetics.PrivateLocation.Delete(id, namespace)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to delete private location `%s`, namespace %s", id, namespace), err.Error())
		return
	}

}
