package parameter

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

	compositeId, dg := tryReadCompositeId(resourceId)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeId != nil {
		resourceId = compositeId.ResourceId
	}

	_, err := kibanaClient.KibanaSynthetics.Parameter.Delete(ctx, resourceId)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to delete parameter `%s`", resourceId), err.Error())
		return
	}
}
