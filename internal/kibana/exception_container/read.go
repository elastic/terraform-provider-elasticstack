package exception_container

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

// readExceptionListFromAPI fetches an exception list from the API and populates the given model
// Returns true if the exception list was found, false if it doesn't exist
func (r *exceptionContainerResource) readExceptionListFromAPI(ctx context.Context, client *clients.ApiClient, model *ExceptionContainerData) (bool, diag.Diagnostics) {
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

	listID := compositeID.ResourceId
	id := kbapi.SecurityExceptionsAPIExceptionListId(listID)
	params := &kbapi.ReadExceptionListParams{
		Id: &id,
	}

	resp, err := oapiClient.API.ReadExceptionListWithResponse(ctx, params)
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

func (r *exceptionContainerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ExceptionContainerData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Store state values for optional fields that might not be returned by API
	stateOsTypes := state.OsTypes
	stateTags := state.Tags

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, state.KibanaConnection, r.client)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	exists, diags := r.readExceptionListFromAPI(ctx, client, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !exists {
		response.State.RemoveResource(ctx)
		return
	}

	// Preserve state values for optional fields if API didn't return them
	// This handles cases where the API doesn't return these fields in the response
	if state.OsTypes.IsNull() && !stateOsTypes.IsNull() && !stateOsTypes.IsUnknown() {
		state.OsTypes = stateOsTypes
	}
	if state.Tags.IsNull() && !stateTags.IsNull() && !stateTags.IsUnknown() {
		state.Tags = stateTags
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
