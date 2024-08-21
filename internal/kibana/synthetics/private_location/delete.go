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

	resourceId := plan.ID.ValueString()
	namespace := plan.SpaceID.ValueString()

	compositeId, dg := tryReadCompositeId(resourceId)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeId != nil {
		resourceId = compositeId.ResourceId
		namespace = compositeId.ClusterId
	}

	err := kibanaClient.KibanaSynthetics.PrivateLocation.Delete(ctx, resourceId, namespace)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to delete private location `%s`, namespace %s", resourceId, namespace), err.Error())
		return
	}

}
