package defaultdataview

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

	spaceID := state.SpaceID.ValueString()

	// Unset the default data view by setting it to null
	setReq := kbapi.SetDefaultDatailViewDefaultJSONRequestBody{
		Force: schemautil.Pointer(true),
	}

	diags = kibanaoapi.SetDefaultDataView(ctx, client, spaceID, setReq)
	resp.Diagnostics.Append(diags...)
}
