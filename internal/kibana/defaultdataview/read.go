package defaultdataview

import (
	"context"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	state, diags = r.read(ctx, client, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) read(ctx context.Context, client *kibanaoapi.Client, state defaultDataViewModel) (defaultDataViewModel, diag.Diagnostics) {
	spaceID := state.SpaceID.ValueString()
	defaultDataViewID, diags := kibanaoapi.GetDefaultDataView(ctx, client, spaceID)
	if diags.HasError() {
		return state, diags
	}

	// Update state with current default data view
	state.DataViewID = types.StringPointerValue(defaultDataViewID)

	// Use the space_id as the resource ID
	state.ID = types.StringValue(spaceID)

	return state, nil
}
