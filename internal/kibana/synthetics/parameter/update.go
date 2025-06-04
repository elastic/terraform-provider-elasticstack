package parameter

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	kibanaClient := synthetics.GetKibanaClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	var state tfModelV0
	diags := request.Plan.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceId := state.ID.ValueString()

	compositeId, dg := tryReadCompositeId(resourceId)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeId != nil {
		resourceId = compositeId.ResourceId
	}

	input := state.toParameterConfig(true)

	_, err := kibanaClient.KibanaSynthetics.Parameter.Update(ctx, resourceId, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to update parameter `%s`", resourceId), err.Error())
		return
	}

	// We can't trust the response from the PUT request. At least with Kibana
	// 9.0.0, it responds with the new values for every field, except `value`,
	// which contains the old value.
	result, err := kibanaClient.KibanaSynthetics.Parameter.Get(ctx, resourceId)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to get parameter after update `%s`", resourceId), err.Error())
		return
	}

	state = toModelV0(*result)

	// Set refreshed state
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
