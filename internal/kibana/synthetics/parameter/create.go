package parameter

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

	input := plan.toParameterConfig(false)

	result, err := kibanaClient.KibanaSynthetics.Parameter.Add(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to create parameter `%s`", input.Key), err.Error())
		return
	}

	resourceId := result.Id

	// We can't trust the response from the POST request. At least with Kibana
	// 9.0.0, it responds without the `value` field set.
	result, err = kibanaClient.KibanaSynthetics.Parameter.Get(ctx, resourceId)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to get parameter after creation `%s`", resourceId), err.Error())
		return
	}

	plan = toModelV0(*result)

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
