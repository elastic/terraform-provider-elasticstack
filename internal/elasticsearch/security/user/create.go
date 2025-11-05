package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.update(ctx, req.Plan, &resp.State)...)
}
