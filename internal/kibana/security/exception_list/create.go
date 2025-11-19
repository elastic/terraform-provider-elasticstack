package exception_list

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *exceptionListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan exceptionListModel

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
	apiResp, diags := kibana_oapi.CreateExceptionList(ctx, kibanaClient, createReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state from response
	diags = plan.fromAPIResponse(ctx, apiResp, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ID for import
	plan.ID = types.StringValue(clients.CompositeId{ClusterId: spaceID, ResourceId: apiResp.Id}.String())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
