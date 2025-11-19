package exception_item

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *exceptionItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state exceptionItemModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract composite ID
	compId, diags := clients.CompositeIdFromStr(state.ID.ValueString())
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

	// Read the exception item
	diags = read(ctx, kibanaClient, &state, compId.ClusterId, compId.ResourceId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// read is an internal function to read exception item data
func read(ctx context.Context, client *kbapi.ClientWithResponses, state *exceptionItemModel, spaceID, itemID string) (diags diag.Diagnostics) {
	// Make API call to read exception item
	params := kbapi.ReadExceptionListItemParams{
		ItemId:        &itemID,
		NamespaceType: utils.Pointer("single"),
	}

	// If namespace type is known from state, use it
	if utils.IsKnown(state.NamespaceType) {
		nsType := state.NamespaceType.ValueString()
		params.NamespaceType = &nsType
	}

	apiResp, err := client.ReadExceptionListItemWithResponse(
		clients.WithKibanaSpaceContext(ctx, spaceID),
		&params,
	)
	if err != nil {
		diags.AddError(
			"Failed to read exception item",
			fmt.Sprintf("Failed to read exception item: %s", err),
		)
		return diags
	}

	if apiResp.StatusCode() == 404 {
		// Resource no longer exists
		state.ID = types.StringNull()
		return diags
	}

	if apiResp.StatusCode() != 200 {
		diags.AddError(
			"Failed to read exception item",
			fmt.Sprintf("API returned status %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return diags
	}

	if apiResp.JSON200 == nil {
		diags.AddError(
			"Failed to read exception item",
			"API response body is empty",
		)
		return diags
	}

	// Populate state from response
	diags.Append(state.fromAPIResponse(ctx, apiResp.JSON200, spaceID)...)
	return diags
}
