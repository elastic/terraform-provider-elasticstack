package data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	if !r.resourceReady(&response.Diagnostics) {
		return
	}

	dataviewClient, err := r.client.GetDataViewsClient()
	if err != nil {
		response.Diagnostics.AddError("unable to get data view client", err.Error())
		return
	}

	var model tfModelV0
	response.Diagnostics.Append(request.Plan.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	apiModel, diags := model.ToUpdateRequest(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, spaceID := model.getIDAndSpaceID()
	authCtx := r.client.SetDataviewAuthContext(ctx)
	_, res, err := dataviewClient.UpdateDataView(authCtx, id, spaceID).UpdateDataViewRequestObject(apiModel).KbnXsrf("true").Execute()
	if err != nil && res == nil {
		response.Diagnostics.AddError("Failed to update data view", err.Error())
		return
	}

	defer res.Body.Close()
	response.Diagnostics.Append(utils.CheckHttpErrorFromFW(res, "Unable to update data view")...)
	if response.Diagnostics.HasError() {
		return
	}

	readModel, diags := r.read(ctx, model)
	response.Diagnostics = append(response.Diagnostics, diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, readModel)...)
}
