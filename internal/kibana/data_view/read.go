package data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *DataViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel dataViewModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibana2Client()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	viewID, spaceID := stateModel.getViewIDAndSpaceID()
	dataView, diags := kibana2.GetDataView(ctx, client, spaceID, viewID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if dataView == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = stateModel.populateFromAPI(ctx, dataView)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
