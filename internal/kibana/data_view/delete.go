package data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *DataViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	diags = kibana2.DeleteDataView(ctx, client, spaceID, viewID)
	resp.Diagnostics.Append(diags...)
}
