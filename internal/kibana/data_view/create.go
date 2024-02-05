package data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
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

	apiModel, diags := model.ToCreateRequest(ctx)
	response.Diagnostics.Append(diags...)
	authCtx := r.client.SetDataviewAuthContext(ctx)
	respModel, res, err := dataviewClient.CreateDataView(authCtx, model.SpaceID.ValueString()).CreateDataViewRequestObject(apiModel).KbnXsrf("true").Execute()
	if err != nil && res == nil {
		response.Diagnostics.AddError("Failed to create data view", err.Error())
		return
	}

	defer res.Body.Close()
	response.Diagnostics.Append(utils.CheckHttpErrorFromFW(res, "Unable to create data view")...)
	if response.Diagnostics.HasError() {
		return
	}

	model.ID = types.StringPointerValue(respModel.DataView.Id)
	readModel, diags := r.read(ctx, model)
	response.Diagnostics = append(response.Diagnostics, diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, readModel)...)
}
