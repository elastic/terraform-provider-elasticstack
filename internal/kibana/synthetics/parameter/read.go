package parameter

import (
	"context"
	"errors"
	"fmt"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r *Resource) readState(ctx context.Context, kibanaClient *kibana_oapi.Client, resourceId string, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	getResult, err := kibanaClient.API.GetParameterWithResponse(ctx, resourceId)
	if err != nil {
		var apiError *kbapi.APIError
		if errors.As(err, &apiError) && apiError.Code == 404 {
			state.RemoveResource(ctx)
			return
		}

		diagnostics.AddError(fmt.Sprintf("Failed to get parameter `%s`", resourceId), err.Error())
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

	resourceId := state.ID.ValueString()

	compositeId, dg := tryReadCompositeId(resourceId)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeId != nil {
		resourceId = compositeId.ResourceId
	}

	r.readState(ctx, kibanaClient, resourceId, &response.State, &response.Diagnostics)
}
