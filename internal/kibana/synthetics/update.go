package synthetics

import (
	"context"
	"fmt"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {

	tflog.Info(ctx, "### Update monitor")

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
	monitorId := kbapi.MonitorID(plan.ID.ValueString())
	result, err := kibanaClient.KibanaSynthetics.Monitor.Update(ctx, monitorId, input.config, input.fields, namespace)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to update Kibana monitor `%s`, namespace %s", input.config.Name, namespace), err.Error())
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
