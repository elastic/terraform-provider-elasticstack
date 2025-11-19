package exception_item

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *exceptionItemResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan ExceptionItemData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
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

	apiModel, diags := plan.toAPICreateRequest(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := oapiClient.API.CreateExceptionListItemWithResponse(ctx, apiModel)
	if err != nil {
		response.Diagnostics.Append(diagutil.FrameworkDiagFromError(err)...)
		return
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			response.Diagnostics.AddError("Invalid response", "Response body is nil")
			return
		}
		compositeID := &clients.CompositeId{ClusterId: plan.SpaceID.ValueString(), ResourceId: resp.JSON200.Id}
		plan.ID = types.StringValue(compositeID.String())

		// Read the exception item back to populate all computed fields
		client, diags = clients.MaybeNewApiClientFromFrameworkResource(ctx, plan.KibanaConnection, r.client)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		exists, diags := r.readExceptionItemFromAPI(ctx, client, &plan)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		if !exists {
			response.Diagnostics.AddError("Exception item not found after creation", "The exception item was created but could not be found afterward")
			return
		}
	default:
		response.Diagnostics.AddError(
			"Unexpected status code from server",
			fmt.Sprintf("API returned status %d: %s", resp.StatusCode(), string(resp.Body)),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}
