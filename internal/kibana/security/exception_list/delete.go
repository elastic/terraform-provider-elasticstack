package exception_list

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *exceptionListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

	// Build delete parameters
	listID := compId.ResourceId
	nsType := "single"
	if utils.IsKnown(state.NamespaceType) {
		nsType = state.NamespaceType.ValueString()
	}
	nsTypeAPI := kbapi.SecurityExceptionsAPIExceptionNamespaceType(nsType)

	params := kbapi.DeleteExceptionListParams{
		ListId:        &listID,
		NamespaceType: &nsTypeAPI,
	}

	// Make API call
	apiResp, err := kibanaClient.API.DeleteExceptionListWithResponse(ctx, compId.ClusterId, &params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete exception list",
			fmt.Sprintf("Failed to delete exception list: %s", err),
		)
		return
	}

	// 404 means resource already deleted
	if apiResp.StatusCode() != http.StatusOK && apiResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Failed to delete exception list",
			fmt.Sprintf("API returned status %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}
}
