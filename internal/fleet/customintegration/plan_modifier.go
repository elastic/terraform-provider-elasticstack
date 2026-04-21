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
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *customIntegrationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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

	// Read the file and compute its SHA256.
	filePath := plan.PackagePath.ValueString()
	f, err := os.Open(filePath)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("package_path"),
			"Cannot read package file",
			"Failed to open package_path for checksum computation: "+err.Error(),
		)
		return
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("package_path"),
			"Cannot read package file",
			"Failed to compute SHA256 of package_path: "+err.Error(),
		)
		return
	}
	newChecksum := hex.EncodeToString(h.Sum(nil))

	// Retrieve prior checksum from state.
	var state customIntegrationModel
	if req.State.Raw.IsNull() {
		// Resource is being created — no prior state. Let Create handle
		// everything; mark computed fields unknown so the plan is valid.
		plan.Checksum = types.StringUnknown()
		plan.PackageName = types.StringUnknown()
		plan.PackageVersion = types.StringUnknown()
		plan.ID = types.StringUnknown()
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the checksum has changed (or was never recorded), invalidate the
	// computed fields so Terraform knows a real update will happen.
	priorChecksum := state.Checksum.ValueString()
	if priorChecksum == "" || newChecksum != priorChecksum {
		plan.Checksum = types.StringUnknown()
		plan.PackageName = types.StringUnknown()
		plan.PackageVersion = types.StringUnknown()
		plan.ID = types.StringUnknown()
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}
