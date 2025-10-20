package default_data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *DefaultDataViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.setDefaultDataView(ctx, req.Plan, &resp.State)...)
}

// setDefaultDataView is a helper method that contains the core logic for setting the default data view.
func (r *DefaultDataViewResource) setDefaultDataView(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var model defaultDataViewModel
	diags := plan.Get(ctx, &model)
	if diags.HasError() {
		return diags
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("unable to get kibana client", err.Error())
		return diags
	}

	dataViewID := model.DataViewID.ValueString()
	force := model.Force.ValueBool()
	spaceID := model.SpaceID.ValueString()
	setReq := kbapi.SetDefaultDatailViewDefaultJSONRequestBody{
		DataViewId: &dataViewID,
		Force:      &force,
	}

	apiDiags := kibana_oapi.SetDefaultDataView(ctx, client, spaceID, setReq)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return diags
	}

	// Use the space_id as the resource ID
	model.ID = types.StringValue(spaceID)

	diags = state.Set(ctx, model)
	return diags
}
