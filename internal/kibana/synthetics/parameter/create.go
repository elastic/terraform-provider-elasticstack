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

	input := plan.toParameterRequest(false)

	// We shouldn't have to do this json marshalling ourselves,
	// https://github.com/oapi-codegen/oapi-codegen/issues/1620 means the generated code doesn't handle the oneOf
	// request body properly.
	inputJson, err := json.Marshal(input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to marshal JSON for parameter `%s`", input.Key), err.Error())
		return
	}

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
		response.Diagnostics.AddError(fmt.Sprintf("Unexpected nil id in create parameter response `%s`", input.Key), "")
		return
	}

	// We can't trust the response from the POST request, so read the parameter
	// again. At least with Kibana 9.0.0, the POST request responds without the
	// `value` field set.
	r.readState(ctx, kibanaClient, *createResponse.Id, &response.State, &response.Diagnostics)
}
