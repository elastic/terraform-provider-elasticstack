package synthetics

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {

	tflog.Info(ctx, "### Create monitor")

	kibanaClient := GetKibanaClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	var plan *tfModelV0 = new(tfModelV0)
	diags := request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	input, diags := plan.toKibanaAPIRequest()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	namespace := plan.SpaceID.ValueString()
	result, err := kibanaClient.KibanaSynthetics.Monitor.Add(input.config, input.fields, namespace)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to create Kibana monitor `%s`, namespace %s", input.config.Name, namespace), err.Error())
		return
	}

	plan, err = toModelV0(result)
	if err != nil {
		response.Diagnostics.AddError("Failed to convert Kibana monitor API to TF state", err.Error())
		return
	}

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
