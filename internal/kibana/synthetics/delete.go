package synthetics

import (
	"context"
	"fmt"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {

	tflog.Info(ctx, "### Delete monitor")

	kibanaClient := GetKibanaClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	var plan tfModelV0
	diags := request.State.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	monitorId, dg := plan.getCompositeId()
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	namespace := plan.SpaceID.ValueString()
	_, err := kibanaClient.KibanaSynthetics.Monitor.Delete(ctx, namespace, kbapi.MonitorID(monitorId.ResourceId))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to delete private location `%s`, namespace %s", monitorId, namespace), err.Error())
		return
	}
}
