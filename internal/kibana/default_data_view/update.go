package default_data_view

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *DefaultDataViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(r.setDefaultDataView(ctx, req.Plan, &resp.State)...)
}
