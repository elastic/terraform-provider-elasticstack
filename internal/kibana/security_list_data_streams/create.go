package security_list_data_streams

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityListDataStreamsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecurityListDataStreamsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Kibana client
	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Create the list data streams
	spaceID := plan.SpaceID.ValueString()
	_, diags := kibana_oapi.CreateListIndex(ctx, client, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data streams to get the actual state
	exists, diags := kibana_oapi.ReadListIndex(ctx, client, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !exists {
		resp.Diagnostics.AddError(
			"Failed to verify list data streams",
			"List data streams were created but could not be verified",
		)
		return
	}

	// Set the ID to the space_id since this is a singleton resource per space
	plan.ID = types.StringValue(spaceID)
	plan.Acknowledged = types.BoolValue(exists)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
