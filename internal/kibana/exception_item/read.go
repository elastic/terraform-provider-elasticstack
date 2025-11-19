package exception_item

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// readExceptionItemFromAPI fetches an exception item from the API and populates the given model
// Returns true if the exception item was found, false if it doesn't exist
func (r *exceptionItemResource) readExceptionItemFromAPI(ctx context.Context, client *clients.ApiClient, model *ExceptionItemData) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return false, diags
	}

	compositeID, diagsTemp := model.GetID()
	diags.Append(diagsTemp...)
	if diags.HasError() {
		return false, diags
	}

	itemID := compositeID.ResourceId
	id := kbapi.SecurityExceptionsAPIExceptionListItemId(itemID)
	params := &kbapi.ReadExceptionListItemParams{
		Id: &id,
	}

	resp, err := oapiClient.API.ReadExceptionListItemWithResponse(ctx, params)
	if err != nil {
		diags.Append(diagutil.FrameworkDiagFromError(err)...)
		return false, diags
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			diags.AddError("Invalid response", "Response body is nil")
			return false, diags
		}
		diags.Append(model.populateFromAPI(ctx, resp.JSON200, compositeID)...)
		return true, diags
	case http.StatusNotFound:
		return false, diags
	default:
		diags.AddError(
			"Unexpected status code from server",
			fmt.Sprintf("API returned status %d: %s", resp.StatusCode(), string(resp.Body)),
		)
		return false, diags
	}
}

func (r *exceptionItemResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
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

	exists, diags := r.readExceptionItemFromAPI(ctx, client, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !exists {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
