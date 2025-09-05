package connectors

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// readConnectorFromAPI fetches a connector from the API and populates the given model
// Returns true if the connector was found, false if it doesn't exist
func (r *Resource) readConnectorFromAPI(ctx context.Context, client *clients.ApiClient, model *tfModel) (bool, diag.Diagnostics) {
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

	connector, diagsTemp := kibana_oapi.GetConnector(ctx, oapiClient, compositeID.ResourceId, compositeID.ClusterId)
	if connector == nil && diagsTemp == nil {
		// Resource not found
		return false, diags
	}
	diags.Append(diagsTemp...)
	if diags.HasError() {
		return false, diags
	}

	diags.Append(model.populateFromAPI(connector, compositeID)...)
	return true, diags
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tfModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, state.KibanaConnection, r.client)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	exists, diags := r.readConnectorFromAPI(ctx, client, &state)
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
