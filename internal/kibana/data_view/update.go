package data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *DataViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel dataViewModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibana2Client()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	body, diags := planModel.toAPIUpdateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	viewID, spaceID := planModel.getViewIDAndSpaceID()
	dataView, diags := kibana2.UpdateDataView(ctx, client, spaceID, viewID, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = planModel.populateFromAPI(ctx, dataView)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
