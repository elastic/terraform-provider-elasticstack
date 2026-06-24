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
	"io"
	"os"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
	resp.Diagnostics.Append(plan.Artifacts.As(ctx, &am, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !typeutils.IsKnown(am.InvestigationGuide) || am.InvestigationGuide.IsNull() {
		return
	}

	var ig investigationGuideModel
	resp.Diagnostics.Append(am.InvestigationGuide.As(ctx, &ig, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if ig.ContentPath.IsUnknown() {
		return
	}
	if ig.ContentPath.IsNull() {
		return
	}

	filePath := ig.ContentPath.ValueString()
	f, err := os.Open(filePath)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("artifacts").AtName("investigation_guide").AtName(attrInvestigationGuideContentPath),
			"Cannot read investigation guide file",
			"Failed to open content_path for checksum computation: "+err.Error(),
		)
		return
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("artifacts").AtName("investigation_guide").AtName(attrInvestigationGuideContentPath),
			"Cannot read investigation guide file",
			"Failed to compute SHA256 of content_path: "+err.Error(),
		)
		return
	}
	newChecksum := hex.EncodeToString(h.Sum(nil))

	var state alertingRuleModel
	if req.State.Raw.IsNull() {
		ig.Checksum = types.StringUnknown()
		plan.ID = types.StringUnknown()
		igObj, d := types.ObjectValueFrom(ctx, investigationGuideAttrTypes(), ig)
		resp.Diagnostics.Append(d...)
		am.InvestigationGuide = igObj
		obj, d := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), am)
		resp.Diagnostics.Append(d...)
		plan.Artifacts = obj
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	priorChecksum := ""
	if typeutils.IsKnown(state.Artifacts) && !state.Artifacts.IsNull() {
		var stateAM artifactsModel
		resp.Diagnostics.Append(state.Artifacts.As(ctx, &stateAM, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		if typeutils.IsKnown(stateAM.InvestigationGuide) && !stateAM.InvestigationGuide.IsNull() {
			var stateIG investigationGuideModel
			resp.Diagnostics.Append(stateAM.InvestigationGuide.As(ctx, &stateIG, basetypes.ObjectAsOptions{})...)
			if resp.Diagnostics.HasError() {
				return
			}
			if typeutils.IsKnown(stateIG.Checksum) && !stateIG.Checksum.IsNull() {
				priorChecksum = stateIG.Checksum.ValueString()
			}
		}
	}

	if priorChecksum == "" || newChecksum != priorChecksum {
		ig.Checksum = types.StringUnknown()
		plan.ID = types.StringUnknown()
		igObj, d := types.ObjectValueFrom(ctx, investigationGuideAttrTypes(), ig)
		resp.Diagnostics.Append(d...)
		am.InvestigationGuide = igObj
		obj, d := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), am)
		resp.Diagnostics.Append(d...)
		plan.Artifacts = obj
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}
