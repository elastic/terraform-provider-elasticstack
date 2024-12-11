package saved_object

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model ksoModelV0

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := model.UpdateModelWithObject()
	if err != nil {
		resp.Diagnostics.AddError("failed to update model from object", err.Error())
		return
	}

	kibanaClient, err := r.client.GetKibanaClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get kibana client", err.Error())
		return
	}

	result, err := kibanaClient.KibanaSavedObject.Import([]byte(model.Imported.ValueString()), false, model.SpaceID.ValueString())
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
