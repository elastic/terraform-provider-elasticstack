package parameter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	kibanaClient := synthetics.GetKibanaOAPIClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	var state tfModelV0
	diags := request.Plan.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceID := state.ID.ValueString()

	compositeID, dg := tryReadCompositeID(resourceID)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeID != nil {
		resourceID = compositeID.ResourceID
	}

	input := state.toParameterRequest(true)

	// We shouldn't have to do this json marshalling ourselves,
	// https://github.com/oapi-codegen/oapi-codegen/issues/1620 means the generated code doesn't handle the oneOf
	// request body properly.
	inputJSON, err := json.Marshal(input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to marshal JSON for parameter `%s`", input.Key), err.Error())
		return
	}

	_, err = kibanaClient.API.PutParameterWithBodyWithResponse(ctx, resourceID, "application/json", bytes.NewReader(inputJSON))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to update parameter `%s`", resourceID), err.Error())
		return
	}

	// We can't trust the response from the PUT request, so read the parameter
	// again. At least with Kibana 9.0.0, the PUT request responds with the new
	// values for every field, except `value`, which contains the old value.
	r.readState(ctx, kibanaClient, resourceID, &response.State, &response.Diagnostics)
}
