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
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ModifyPlan computes the SHA-256 checksum of investigation_guide.content_path
// at plan time. When the file content has changed (or on first create), the
// checksum and resource id are marked unknown so Terraform shows a non-empty
// plan and triggers an apply.
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

	if !typeutils.IsKnown(plan.Artifacts) || plan.Artifacts.IsNull() {
		return
	}

	var am artifactsModel
	resp.Diagnostics.Append(plan.Artifacts.As(ctx, &am, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !typeutils.IsKnown(am.InvestigationGuide) || am.InvestigationGuide.IsNull() {
		return
	}

	var igm investigationGuideModel
	resp.Diagnostics.Append(am.InvestigationGuide.As(ctx, &igm, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if igm.ContentPath.IsUnknown() {
		igm.Checksum = types.StringUnknown()
		newIGObj, d := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), igm)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		am.InvestigationGuide = newIGObj
		newArtifactsObj, d := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), am)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		plan.Artifacts = newArtifactsObj
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}

	if igm.ContentPath.IsNull() || igm.ContentPath.ValueString() == "" {
		return
	}

	// Read the file and compute its SHA-256.
	filePath := igm.ContentPath.ValueString()
	content, err := os.ReadFile(filePath)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("artifacts").AtName("investigation_guide").AtName("content_path"),
			"Cannot read investigation guide file",
			"Failed to open content_path for checksum computation: "+err.Error(),
		)
		return
	}
	sum := sha256.Sum256(content)
	newChecksum := hex.EncodeToString(sum[:])

	// On create there is no prior state to compare against.
	if req.State.Raw.IsNull() {
		igm.Checksum = types.StringUnknown()
		newIGObj, d := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), igm)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		am.InvestigationGuide = newIGObj
		newArtifactsObj, d := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), am)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		plan.Artifacts = newArtifactsObj
		plan.ID = types.StringUnknown()
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}

	var state alertingRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve prior checksum from state.
	priorChecksum := ""
	if typeutils.IsKnown(state.Artifacts) && !state.Artifacts.IsNull() {
		var priorAM artifactsModel
		d := state.Artifacts.As(ctx, &priorAM, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})
		resp.Diagnostics.Append(d...)
		if !resp.Diagnostics.HasError() && typeutils.IsKnown(priorAM.InvestigationGuide) && !priorAM.InvestigationGuide.IsNull() {
			var priorIGM investigationGuideModel
			d := priorAM.InvestigationGuide.As(ctx, &priorIGM, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})
			resp.Diagnostics.Append(d...)
			if !resp.Diagnostics.HasError() {
				priorChecksum = priorIGM.Checksum.ValueString()
			}
		}
	}

	// If the checksum has changed (or was never recorded), invalidate the
	// computed fields so Terraform knows a real update will happen.
	if priorChecksum == "" || newChecksum != priorChecksum {
		igm.Checksum = types.StringUnknown()
		newIGObj, d := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), igm)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		am.InvestigationGuide = newIGObj
		newArtifactsObj, d := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), am)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		plan.Artifacts = newArtifactsObj
		plan.ID = types.StringUnknown()
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}
