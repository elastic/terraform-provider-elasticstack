package prebuilt_rules

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *PrebuiltRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model prebuiltRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	spaceID := model.SpaceID.ValueString()

	// Disable rules that were managed by this resource
	tags, tagDiags := model.getTags(ctx)
	resp.Diagnostics.Append(tagDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(tags) > 0 {
		resp.Diagnostics.Append(performBulkActionByTags(ctx, client, spaceID, "disable", tags)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// The Terraform state will be removed automatically
}
