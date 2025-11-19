package exception_item

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *exceptionItemResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state ExceptionItemData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, state.KibanaConnection, r.client)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		response.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	compositeID, diags := state.GetID()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	itemID := compositeID.ResourceId
	id := kbapi.SecurityExceptionsAPIExceptionListItemId(itemID)
	params := &kbapi.DeleteExceptionListItemParams{
		Id: &id,
	}

	resp, err := oapiClient.API.DeleteExceptionListItemWithResponse(ctx, params)
	if err != nil {
		response.Diagnostics.Append(diagutil.FrameworkDiagFromError(err)...)
		return
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusNotFound {
		response.Diagnostics.AddError(
			"Unexpected status code from server",
			fmt.Sprintf("API returned status %d: %s", resp.StatusCode(), string(resp.Body)),
		)
		return
	}
}
