package private_location

import (
	"context"
	"fmt"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {

	kibanaClient := synthetics.GetKibanaClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	var plan tfModelV0
	diags := request.State.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	monitorId, dg := synthetics.GetCompositeId(plan.ID.ValueString())
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	namespace := plan.SpaceID.ValueString()
	err := kibanaClient.KibanaSynthetics.PrivateLocation.Delete(ctx, monitorId.ResourceId, namespace)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to delete private location `%s`, namespace %s", monitorId, namespace), err.Error())
		return
	}

}
