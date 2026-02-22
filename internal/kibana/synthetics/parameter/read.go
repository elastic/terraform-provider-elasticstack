package parameter

import (
	"context"
	"errors"
	"fmt"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r *Resource) readState(ctx context.Context, kibanaClient *kibanaoapi.Client, resourceID string, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	getResult, err := kibanaClient.API.GetParameterWithResponse(ctx, resourceID)
	if err != nil {
		var apiError *kbapi.APIError
		if errors.As(err, &apiError) && apiError.Code == 404 {
			state.RemoveResource(ctx)
			return
		}

		diagnostics.AddError(fmt.Sprintf("Failed to get parameter `%s`", resourceID), err.Error())
		return
	}

	model := modelV0FromOAPI(*getResult.JSON200)

	// Set refreshed state
	diags := state.Set(ctx, &model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	kibanaClient := synthetics.GetKibanaOAPIClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	var state tfModelV0
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceID := state.ID.ValueString()

	compositeID, dg := tryReadCompositeID(resourceID)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeID != nil {
		resourceID = compositeID.ResourceID
	}

	r.readState(ctx, kibanaClient, resourceID, &response.State, &response.Diagnostics)
}
