package default_data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *DefaultDataViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan defaultDataViewModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get kibana client", err.Error())
		return
	}

	dataViewID := plan.DataViewID.ValueString()
	force := plan.Force.ValueBool()

	setReq := kbapi.SetDefaultDatailViewDefaultJSONRequestBody{
		DataViewId: &dataViewID,
		Force:      &force,
	}

	diags = kibana_oapi.SetDefaultDataView(ctx, client, setReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep the ID as "default" - it should remain static
	plan.ID = types.StringValue("default")

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
