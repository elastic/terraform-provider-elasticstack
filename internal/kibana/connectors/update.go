package connectors

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan tfModel

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

	apiModel, diags := plan.toAPIModel()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	compositeID, diags := plan.GetID()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	apiModel.ConnectorID = compositeID.ResourceId

	connectorID, diags := kibana_oapi.UpdateConnector(ctx, oapiClient, apiModel)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	newCompositeID := &clients.CompositeId{ClusterId: apiModel.SpaceID, ResourceId: connectorID}
	plan.ID = types.StringValue(newCompositeID.String())

	// Read the connector back to populate all computed fields
	client, diags = clients.MaybeNewApiClientFromFrameworkResource(ctx, plan.KibanaConnection, r.client)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	exists, diags := r.readConnectorFromAPI(ctx, client, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !exists {
		response.Diagnostics.AddError("Connector not found after update", "The connector was updated but could not be found afterward")
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}
