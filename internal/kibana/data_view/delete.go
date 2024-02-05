package data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	if !r.resourceReady(&response.Diagnostics) {
		return
	}

	dataviewClient, err := r.client.GetDataViewsClient()
	if err != nil {
		response.Diagnostics.AddError("unable to get data view client", err.Error())
		return
	}

	var model tfModelV0
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	authCtx := r.client.SetDataviewAuthContext(ctx)
	res, err := dataviewClient.DeleteDataView(authCtx, model.ID.ValueString(), model.SpaceID.ValueString()).KbnXsrf("true").Execute()
	if err != nil && res == nil {
		response.Diagnostics.AddError("Failed to delete data view", err.Error())
	}

	defer res.Body.Close()
	response.Diagnostics.Append(utils.CheckHttpErrorFromFW(res, "Unable to delete data view")...)
}
