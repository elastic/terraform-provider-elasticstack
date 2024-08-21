package synthetics

import (
	"context"
	"fmt"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {

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

	compositeId, dg := GetCompositeId(plan.ID.ValueString())
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	namespace := plan.SpaceID.ValueString()
	_, err := kibanaClient.KibanaSynthetics.Monitor.Delete(ctx, namespace, kbapi.MonitorID(compositeId.ResourceId))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to delete private location `%s`, namespace %s", compositeId, namespace), err.Error())
		return
	}
}
