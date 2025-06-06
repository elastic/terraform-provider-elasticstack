package parameter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	kibanaClient := synthetics.GetKibanaOAPIClient(r, response.Diagnostics)
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
	// We shouldn't have to do this json marshalling ourselves,
	// https://github.com/oapi-codegen/oapi-codegen/issues/1620 means the generated code doesn't handle the oneOf
	// request body properly.
	inputJson, err := json.Marshal(input)
	createResult, err := kibanaClient.API.PostParametersWithBodyWithResponse(ctx, "application/json", bytes.NewReader(inputJson))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to create parameter `%s`", input.Key), err.Error())
		return
	}

	createResponse, err := createResult.JSON200.AsSyntheticsPostParameterResponse()
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to parse parameter response `%s`", input.Key), err.Error())
		return
	}

	if createResponse.Id == nil {
		response.Diagnostics.AddError(fmt.Sprintf("Unexpected nil id in create parameter response `%s`", input.Key), err.Error())
		return
	}

	resourceId := *createResponse.Id

	// We can't trust the response from the POST request. At least with Kibana
	// 9.0.0, it responds without the `value` field set.
	getResult, err := kibanaClient.API.GetParameterWithResponse(ctx, resourceId)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to get parameter after creation `%s`", resourceId), err.Error())
		return
	}

	plan = modelV0FromOAPI(*getResult.JSON200)

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
