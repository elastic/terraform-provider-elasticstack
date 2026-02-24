// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
