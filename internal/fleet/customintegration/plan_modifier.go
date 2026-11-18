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

package customintegration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/fileutil"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to do on destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan customIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If package_path is unknown (e.g. computed from another resource), we
	// cannot read the file yet; leave the plan as-is.
	if plan.PackagePath.IsUnknown() {
		return
	}

	filePath := plan.PackagePath.ValueString()

	// Retrieve prior checksum from state, if any.
	hasPriorState := !req.State.Raw.IsNull()
	var priorChecksum string
	if hasPriorState {
		var state customIntegrationModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		priorChecksum = state.Checksum.ValueString()
	}

	_, changed, err := fileutil.FileChecksumDrifted(filePath, priorChecksum, hasPriorState)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("package_path"),
			"Cannot read package file",
			"Failed to compute SHA256 of package_path: "+err.Error(),
		)
		return
	}

	// If the checksum has changed (or there is no prior state), invalidate
	// the computed fields so Terraform knows a real write will happen.
	if changed {
		plan.Checksum = types.StringUnknown()
		plan.PackageName = types.StringUnknown()
		plan.PackageVersion = types.StringUnknown()
		plan.ID = types.StringUnknown()
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}
