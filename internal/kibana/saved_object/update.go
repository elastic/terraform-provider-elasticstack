package saved_object

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model ksoModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibanaClient, err := r.client.GetKibanaClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get kibana client", err.Error())
		return
	}

	result, err := kibanaClient.KibanaSavedObject.Import([]byte(model.Imported.ValueString()), true, model.SpaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to import saved object", err.Error())
		return
	}

	var success any
	var ok bool
	if success, ok = result["success"]; !ok {
		resp.Diagnostics.AddError("failed to import saved object", "success key not found in response")
		return
	}
	if success != true {
		resp.Diagnostics.AddError("failed to import saved object", fmt.Sprintf("%v\n", result["errors"]))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
