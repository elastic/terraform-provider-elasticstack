package saved_object

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var configData ksoModelV0

	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := configData.UpdateModelWithObject()
	if err != nil {
		resp.Diagnostics.AddError("failed to update model from object", err.Error())
		return
	}

	resp.Plan.SetAttribute(ctx, path.Root("id"), configData.ID)
	resp.Plan.SetAttribute(ctx, path.Root("type"), configData.Type)
	resp.Plan.SetAttribute(ctx, path.Root("imported"), configData.Imported)
}
