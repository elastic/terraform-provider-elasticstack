package import_saved_objects

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	r.importObjects(ctx, request.Plan, &response.State, &response.Diagnostics)
}
