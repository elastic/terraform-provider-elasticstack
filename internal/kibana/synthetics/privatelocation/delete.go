package privatelocation

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

	resourceID := plan.ID.ValueString()

	compositeID, dg := tryReadCompositeID(resourceID)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeID != nil {
		resourceID = compositeID.ResourceID
	}

	err := kibanaClient.KibanaSynthetics.PrivateLocation.Delete(ctx, resourceID)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to delete private location `%s`", resourceID), err.Error())
		return
	}

}
