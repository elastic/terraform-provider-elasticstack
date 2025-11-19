package exception_container

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

func (r *exceptionContainerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan ExceptionContainerData
	var state ExceptionContainerData

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
	exists, diags := r.readExceptionListFromAPI(ctx, client, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !exists {
		response.Diagnostics.AddError("Exception list not found", "The exception list could not be found for update")
		return
	}

	// Get version from state for optimistic locking
	var version string
	if state.ID.ValueString() != "" {
		listID := compositeID.ResourceId
		id := kbapi.SecurityExceptionsAPIExceptionListId(listID)
		params := &kbapi.ReadExceptionListParams{
			Id: &id,
		}
		currentResp, err := oapiClient.API.ReadExceptionListWithResponse(ctx, params)
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
	listID := compositeID.ResourceId
	id := kbapi.SecurityExceptionsAPIExceptionListId(listID)
	apiModel.Id = &id

	resp, err := oapiClient.API.UpdateExceptionListWithResponse(ctx, apiModel)
	if err != nil {
		response.Diagnostics.Append(diagutil.FrameworkDiagFromError(err)...)
		return
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		// Read the exception list back to populate all computed fields
		client, diags = clients.MaybeNewApiClientFromFrameworkResource(ctx, plan.KibanaConnection, r.client)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		// Store plan values for optional fields that might not be returned by API
		planOsTypes := plan.OsTypes
		planTags := plan.Tags

		// Read the updated resource from API
		oapiClient, err = client.GetKibanaOapiClient()
		if err != nil {
			response.Diagnostics.AddError("Failed to get Kibana client", err.Error())
			return
		}

		readParams := &kbapi.ReadExceptionListParams{
			Id: &id,
		}
		readResp, err := oapiClient.API.ReadExceptionListWithResponse(ctx, readParams)
		if err != nil {
			response.Diagnostics.Append(diagutil.FrameworkDiagFromError(err)...)
			return
		}

		if readResp.StatusCode() != http.StatusOK || readResp.JSON200 == nil {
			response.Diagnostics.AddError("Exception list not found after update", "The exception list was updated but could not be found afterward")
			return
		}

		// Populate from API response
		diags = plan.populateFromAPI(ctx, readResp.JSON200, compositeID)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		// Preserve plan values for optional fields if API didn't return them
		// Check if the populated value is null but the plan had a value
		if plan.OsTypes.IsNull() && !planOsTypes.IsNull() && !planOsTypes.IsUnknown() {
			plan.OsTypes = planOsTypes
		}
		if plan.Tags.IsNull() && !planTags.IsNull() && !planTags.IsUnknown() {
			plan.Tags = planTags
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
