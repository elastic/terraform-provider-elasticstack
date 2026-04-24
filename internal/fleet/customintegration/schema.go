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

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *customIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Uploads a locally-built Fleet custom integration package archive (`.zip` or `.tar.gz`) to " +
			"Kibana via the EPM binary upload API (`POST /api/fleet/epm/packages`) and manages its lifecycle. " +
			"Unlike `elasticstack_fleet_integration`, which installs packages from the Elastic package registry by " +
			"name and version, this resource accepts a local archive produced with `elastic-package build`.\n\n" +
			"Change detection is based on the SHA-256 hash of the file at `package_path`. When the file content " +
			"changes, the package is re-uploaded automatically on the next apply.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource. Derived from the uploaded package's name and version.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"package_path": schema.StringAttribute{
				Description: "Absolute or working-directory-relative path to the local custom integration package archive " +
					"(`.zip` or `.tar.gz`). The file must be readable at plan time.",
				Required: true,
			},
			"package_name": schema.StringAttribute{
				Description: "The name of the integration package, as declared in the package's `manifest.yml` and " +
					"returned by Fleet after upload. Computed.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"package_version": schema.StringAttribute{
				Description: "The installed version of the integration package, resolved from the upload response " +
					"(or from the Fleet packages list API if the response does not carry a version). Computed.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"checksum": schema.StringAttribute{
				Description: "SHA-256 hex digest of the file at `package_path` at the time of the last successful upload. " +
					"Used to detect content changes.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ignore_mapping_update_errors": schema.BoolAttribute{
				Description: "When set, forwarded as the `ignoreMappingUpdateErrors=true` query parameter on upload. " +
					"Requires Elastic Stack 8.11 or newer.",
				Optional: true,
			},
			"skip_data_stream_rollover": schema.BoolAttribute{
				Description: "When set, forwarded as the `skipDataStreamRollover=true` query parameter on upload. " +
					"Requires Elastic Stack 8.11 or newer.",
				Optional: true,
			},
			"skip_destroy": schema.BoolAttribute{
				Description: "Set to `true` to leave the package installed in Fleet when the resource is destroyed. " +
					"Defaults to `false` (the package is uninstalled).",
				Optional: true,
			},
			"space_id": schema.StringAttribute{
				Description: "The Kibana space in which to install the package. When set, Fleet API calls are routed " +
					"through `/s/<space_id>/api/fleet/epm/packages`. When omitted, the default space is used.",
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}
}

// ModifyPlan recomputes the SHA-256 of the file at `package_path` at plan
// time and, when it differs from the value stored in state, marks the
// computed trio (`checksum`, `package_name`, `package_version`) as unknown.
// That lets the plan diff communicate a pending re-upload even though
// `package_path` itself has not changed. If the file cannot be read, plan
// fails with a clear diagnostic rather than a generic apply-time error.
func (r *customIntegrationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip during destroy (no plan) and during create (no prior state).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan customIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state customIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// `package_path` is required, so it will always be known at plan time
	// except in rare cases where it is itself derived from another unknown.
	if plan.PackagePath.IsNull() || plan.PackagePath.IsUnknown() {
		return
	}

	newHash, err := sha256File(plan.PackagePath.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("package_path"),
			"Unable to read custom integration package",
			err.Error(),
		)
		return
	}

	if !state.Checksum.IsNull() && !state.Checksum.IsUnknown() && state.Checksum.ValueString() == newHash {
		// Content unchanged; leave the computed attributes taken from
		// state (via UseStateForUnknown) in place.
		return
	}

	// Content changed: tell Terraform the computed attributes are pending
	// (known after apply) so the plan diff is transparent.
	plan.Checksum = types.StringUnknown()
	plan.PackageName = types.StringUnknown()
	plan.PackageVersion = types.StringUnknown()
	plan.ID = types.StringUnknown()

	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}
