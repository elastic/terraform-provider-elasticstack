package exception_list

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *exceptionListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state exceptionListModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract composite ID
	compId, diags := clients.CompositeIdFromStrFw(state.ID.ValueString())
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

	// Read the exception list
	diags = read(ctx, kibanaClient, &state, compId.ClusterId, compId.ResourceId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// read is an internal function to read exception list data
func read(ctx context.Context, client *kibana_oapi.Client, state *exceptionListModel, spaceID, listID string) (diags diag.Diagnostics) {
	// Make API call to read exception list
	nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType("single")
	params := kbapi.ReadExceptionListParams{
		ListId:        &listID,
		NamespaceType: &nsType,
	}

	// If namespace type is known from state, use it
	if utils.IsKnown(state.NamespaceType) {
		nsType = kbapi.SecurityExceptionsAPIExceptionNamespaceType(state.NamespaceType.ValueString())
		params.NamespaceType = &nsType
	}

	apiResp, err := client.API.ReadExceptionListWithResponse(ctx, spaceID, &params)
	if err != nil {
		diags.AddError(
			"Failed to read exception list",
			fmt.Sprintf("Failed to read exception list: %s", err),
		)
		return diags
	}

	if apiResp.StatusCode() == http.StatusNotFound {
		// Resource no longer exists
		state.ID = types.StringNull()
		return diags
	}

	if apiResp.StatusCode() != http.StatusOK {
		diags.AddError(
			"Failed to read exception list",
			fmt.Sprintf("API returned status %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return diags
	}

	if apiResp.JSON200 == nil {
		diags.AddError(
			"Failed to read exception list",
			"API response body is empty",
		)
		return diags
	}

	// Populate state from response
	diags.Append(state.fromAPIResponse(ctx, apiResp.JSON200, spaceID)...)
	return diags
}
