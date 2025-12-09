package security_list_data_streams

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Update is a no-op for this resource because the only configurable attribute (space_id)
// has RequiresReplace plan modifier. This method exists to satisfy the resource.Resource interface.
// If this method is called, it means the framework has determined no replacement is needed,
// so we simply pass the plan through to state.
func (r *securityListDataStreamsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecurityListDataStreamsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
