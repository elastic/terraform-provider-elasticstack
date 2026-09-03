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

package alertingrule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/fileutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ModifyPlan detects external changes to a file referenced by
// artifacts.investigation_guide.content_path. When the file's SHA-256 differs
// from the checksum recorded in state (or the resource is being created), the
// computed checksum is marked unknown so Terraform produces a non-empty plan.
func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to do on destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan alertingRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planIG, d := plan.investigationGuideFrom(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	var priorChecksum string
	hasPriorState := !req.State.Raw.IsNull()
	if hasPriorState {
		var state alertingRuleModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		stateIG, d := state.investigationGuideFrom(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if stateIG != nil && typeutils.IsKnown(stateIG.Checksum) && !stateIG.Checksum.IsNull() {
			priorChecksum = stateIG.Checksum.ValueString()
		}
	}

	changed, d := contentPathChecksumChanged(planIG, priorChecksum, hasPriorState)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// On create (no prior state) or when the file changed, mark checksum unknown
	// so the plan shows a change and Create/Update recomputes the concrete value.
	if changed {
		resp.Diagnostics.Append(setInvestigationGuideChecksumUnknown(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}

// contentPathChecksumChanged reports whether the file-based investigation guide
// checksum must be invalidated (marked unknown) so Terraform schedules a write.
//
// It returns false (no-op) when there is no investigation guide, when the guide
// uses inline content, or when content_path is unknown (e.g. derived from
// another resource and not yet resolvable). Otherwise it reads the file at
// content_path, and returns true when the resource is being created
// (hasPriorState == false) or the freshly computed SHA-256 differs from the
// checksum recorded in state.
func contentPathChecksumChanged(planIG *investigationGuideModel, priorChecksum string, hasPriorState bool) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Only the file-based investigation guide participates in checksum drift
	// detection. Inline content is handled by normal string diffing, and an
	// unknown content_path cannot be read yet.
	if planIG == nil || !typeutils.IsKnown(planIG.ContentPath) || planIG.ContentPath.IsNull() {
		return false, diags
	}

	_, changed, err := fileutil.FileChecksumDrifted(planIG.ContentPath.ValueString(), priorChecksum, hasPriorState)
	if err != nil {
		diags.AddAttributeError(
			path.Root("artifacts").AtName("investigation_guide").AtName("content_path"),
			"Cannot read investigation guide file",
			"Failed to compute SHA-256 of content_path: "+err.Error(),
		)
		return false, diags
	}

	return changed, diags
}

// setInvestigationGuideChecksumUnknown rebuilds plan.Artifacts with the
// investigation guide checksum marked unknown, preserving the other fields.
func setInvestigationGuideChecksumUnknown(ctx context.Context, plan *alertingRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	var am artifactsModel
	diags.Append(plan.Artifacts.As(ctx, &am, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return diags
	}
	var ig investigationGuideModel
	diags.Append(am.InvestigationGuide.As(ctx, &ig, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return diags
	}
	ig.Checksum = types.StringUnknown()

	igObj, d := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), ig)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	artObj, d := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{InvestigationGuide: igObj})
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	plan.Artifacts = artObj
	return diags
}
