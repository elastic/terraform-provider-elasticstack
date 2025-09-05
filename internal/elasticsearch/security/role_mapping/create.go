package role_mapping

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *roleMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	diags := r.update(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}