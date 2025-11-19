package exception_item

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *exceptionItemResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan ExceptionItemData
	var state ExceptionItemData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, plan.KibanaConnection, r.client)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		response.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	compositeID, diags := plan.GetID()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Read current state to get version
	exists, diags := r.readExceptionItemFromAPI(ctx, client, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !exists {
		response.Diagnostics.AddError("Exception item not found", "The exception item could not be found for update")
		return
	}

	// Get version from state for optimistic locking
	var version string
	if state.ID.ValueString() != "" {
		itemID := compositeID.ResourceId
		id := kbapi.SecurityExceptionsAPIExceptionListItemId(itemID)
		params := &kbapi.ReadExceptionListItemParams{
			Id: &id,
		}
		currentResp, err := oapiClient.API.ReadExceptionListItemWithResponse(ctx, params)
		if err != nil {
			response.Diagnostics.Append(diagutil.FrameworkDiagFromError(err)...)
			return
		}
		if currentResp.StatusCode() == http.StatusOK && currentResp.JSON200 != nil && currentResp.JSON200.UnderscoreVersion != nil {
			version = *currentResp.JSON200.UnderscoreVersion
		}
	}

	apiModel, diags := plan.toAPIUpdateRequest(ctx, version)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set the ID in the request body
	itemID := compositeID.ResourceId
	id := kbapi.SecurityExceptionsAPIExceptionListItemId(itemID)
	apiModel.Id = &id

	resp, err := oapiClient.API.UpdateExceptionListItemWithResponse(ctx, apiModel)
	if err != nil {
		response.Diagnostics.Append(diagutil.FrameworkDiagFromError(err)...)
		return
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		// Read the exception item back to populate all computed fields
		client, diags = clients.MaybeNewApiClientFromFrameworkResource(ctx, plan.KibanaConnection, r.client)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		exists, diags = r.readExceptionItemFromAPI(ctx, client, &plan)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		if !exists {
			response.Diagnostics.AddError("Exception item not found after update", "The exception item was updated but could not be found afterward")
			return
		}

		plan.ID = types.StringValue(compositeID.String())
	default:
		response.Diagnostics.AddError(
			"Unexpected status code from server",
			fmt.Sprintf("API returned status %d: %s", resp.StatusCode(), string(resp.Body)),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}
