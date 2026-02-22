package jobstate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *mlJobStateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// ML job state resource only manages the state, not the job itself.
	// When the resource is deleted, we simply remove it from Terraform state
	// without affecting the actual ML job state in Elasticsearch.
	// The job will remain in its current state (opened or closed).
	var jobID basetypes.StringValue
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("job_id"), &jobID)...)
	tflog.Info(ctx, fmt.Sprintf(`Dropping ML job state "%s", this does not close the job`, jobID.ValueString()))
}
