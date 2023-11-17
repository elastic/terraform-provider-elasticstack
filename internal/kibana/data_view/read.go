package data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if !r.resourceReady(&response.Diagnostics) {
		return
	}

	var model tfModelV0
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	apiModel, diags := r.read(ctx, model)
	response.Diagnostics = append(response.Diagnostics, diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if apiModel == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, apiModel)...)
}

func (r *Resource) read(ctx context.Context, model tfModelV0) (*apiModelV0, diag.Diagnostics) {
	dataviewClient, err := r.client.GetDataViewsClient()
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("unable to get data view client", err.Error()),
		}
	}

	authCtx := r.client.SetDataviewAuthContext(ctx)
	respModel, res, err := dataviewClient.GetDataView(authCtx, model.ID.ValueString()).Execute()
	if err != nil && res == nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to read data view", err.Error()),
		}
	}

	defer res.Body.Close()
	if res.StatusCode == 404 {
		return nil, nil
	}

	if diags := utils.CheckHttpErrorFromFW(res, "Unable to read data view"); diags.HasError() {
		return nil, diags
	}

	apiModel, diags := model.FromResponse(ctx, respModel)
	if diags.HasError() {
		return nil, diags
	}

	return &apiModel, nil
}
