package default_data_view

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *DefaultDataViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state defaultDataViewModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get kibana client", err.Error())
		return
	}

	defaultDataViewID, diags := kibana_oapi.GetDefaultDataView(ctx, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If no default data view is set, remove from state
	if defaultDataViewID == nil || *defaultDataViewID == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with current default data view
	state.DataViewID = types.StringValue(*defaultDataViewID)
	state.ID = types.StringValue(fmt.Sprintf("default-data-view:%s", *defaultDataViewID))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
