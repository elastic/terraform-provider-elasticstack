package private_location

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {

	if !r.resourceReady(&response.Diagnostics) {
		return
	}

	kibanaClient, err := r.client.GetKibanaClient()
	if err != nil {
		response.Diagnostics.AddError("unable to get kibana client", err.Error())
		return
	}

	var plan tfModelV0
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	input := plan.toPrivateLocationConfig()

	namespace := plan.SpaceID.ValueString()
	result, err := kibanaClient.KibanaSynthetics.PrivateLocation.Create(input, namespace)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to create private location `%s`, namespace %s", input.Label, namespace), err.Error())
		return
	}

	plan = toModelV0(namespace, result.PrivateLocationConfig)

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
