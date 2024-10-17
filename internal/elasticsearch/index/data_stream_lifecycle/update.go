package data_stream_lifecycle

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.create(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}
