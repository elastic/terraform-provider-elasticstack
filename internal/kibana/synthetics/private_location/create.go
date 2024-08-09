package private_location

import (
	"context"
	"fmt"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {

	kibanaClient := synthetics.GetKibanaClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	var plan tfModelV0
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	input := plan.toPrivateLocation()

	namespace := plan.SpaceID.ValueString()
	result, err := kibanaClient.KibanaSynthetics.PrivateLocation.Create(input.PrivateLocationConfig, namespace)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to create private location `%s`, namespace %s", input.Label, namespace), err.Error())
		return
	}

	plan = toModelV0(*result)

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
