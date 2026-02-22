package connectors

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan tfModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, plan.KibanaConnection, r.client)
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

	version, sdkDiags := client.ServerVersion(ctx)
	response.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if response.Diagnostics.HasError() {
		return
	}

	if apiModel.ConnectorID != "" && version.LessThan(MinVersionSupportingPreconfiguredIDs) {
		response.Diagnostics.AddError(
			"Unsupported Elastic Stack version",
			"Preconfigured connector IDs are only supported for Elastic Stack v"+MinVersionSupportingPreconfiguredIDs.String()+" and above."+
				" Either remove the `connector_id` attribute or upgrade your target cluster to supported version",
		)
		return
	}

	connectorID, diags := kibanaoapi.CreateConnector(ctx, oapiClient, apiModel)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	compositeID := &clients.CompositeID{ClusterID: apiModel.SpaceID, ResourceID: connectorID}
	plan.ID = types.StringValue(compositeID.String())

	// Read the connector back to populate all computed fields
	client, diags = clients.MaybeNewAPIClientFromFrameworkResource(ctx, plan.KibanaConnection, r.client)
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
		response.Diagnostics.AddError("Connector not found after creation", "The connector was created but could not be found afterward")
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}
