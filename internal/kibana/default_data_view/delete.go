package default_data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *DefaultDataViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state defaultDataViewModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If skip_delete is true, leave the default data view unchanged
	if state.SkipDelete.ValueBool() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get kibana client", err.Error())
		return
	}

	// Unset the default data view by setting it to null
	var nullDataViewID *string = nil
	force := true

	setReq := kbapi.SetDefaultDatailViewDefaultJSONRequestBody{
		DataViewId: nullDataViewID,
		Force:      &force,
	}

	diags = kibana_oapi.SetDefaultDataView(ctx, client, setReq)
	resp.Diagnostics.Append(diags...)
}
