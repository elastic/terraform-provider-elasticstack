package exception_item

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *exceptionItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan exceptionItemModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract composite ID
	compId, diags := clients.CompositeIdFromStrFw(plan.ID.ValueString())
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

	// Build update request
	updateReq, diags := plan.toUpdateRequest(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make API call
	apiResp, diags := kibana_oapi.UpdateExceptionListItem(ctx, kibanaClient, updateReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state from response
	diags = plan.fromAPIResponse(ctx, apiResp, compId.ClusterId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
