package exception_item

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *exceptionItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan exceptionItemModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Kibana client
	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get Kibana client",
			fmt.Sprintf("Unable to get Kibana client: %s", err),
		)
		return
	}

	// Determine space ID
	spaceID := "default"
	if utils.IsKnown(plan.SpaceID) {
		spaceID = plan.SpaceID.ValueString()
	}

	// Build create request
	createReq, diags := plan.toCreateRequest(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make API call
	apiResp, err := kibanaClient.CreateExceptionListItemWithResponse(
		clients.WithKibanaSpaceContext(ctx, spaceID),
		createReq,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create exception item",
			fmt.Sprintf("Failed to create exception item: %s", err),
		)
		return
	}

	if apiResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Failed to create exception item",
			fmt.Sprintf("API returned status %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Failed to create exception item",
			"API response body is empty",
		)
		return
	}

	// Populate state from response
	diags = plan.fromAPIResponse(ctx, apiResp.JSON200, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ID for import
	plan.ID = types.StringValue(clients.CompositeId{ClusterId: spaceID, ResourceId: apiResp.JSON200.Id}.String())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
