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

	// We can't trust the response from the POST request, so read the parameter
	// again. At least with Kibana 9.0.0, the POST request responds without the
	// `value` field set.
	r.readState(ctx, kibanaClient, toModelV0(*result), &response.State, &response.Diagnostics)
}
