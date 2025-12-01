package security_exception_item

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *ExceptionItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ExceptionItemModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse composite ID to get space_id and resource_id
	compId, compIdDiags := clients.CompositeIdFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(compIdDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SpaceID = types.StringValue(compId.ClusterId)

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Read by ID
	id := kbapi.SecurityExceptionsAPIExceptionListItemId(compId.ResourceId)
	params := &kbapi.ReadExceptionListItemParams{
		Id: &id,
	}

	readResp, diags := kibana_oapi.GetExceptionListItem(ctx, client, state.SpaceID.ValueString(), params)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readResp == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with response using model method
	diags = state.fromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
