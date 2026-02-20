package defaultdataview

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.setDefaultDataView(ctx, req.Plan, &resp.State)...)
}

// setDefaultDataView is a helper method that contains the core logic for setting the default data view.
func (r *Resource) setDefaultDataView(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
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

	dataViewID := model.DataViewID.ValueStringPointer()
	force := model.Force.ValueBool()
	spaceID := model.SpaceID.ValueString()
	setReq := kbapi.SetDefaultDatailViewDefaultJSONRequestBody{
		DataViewId: dataViewID,
		Force:      &force,
	}

	apiDiags := kibanaoapi.SetDefaultDataView(ctx, client, spaceID, setReq)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return diags
	}

	model, readDiags := r.read(ctx, client, model)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	diags = state.Set(ctx, model)
	return diags
}
