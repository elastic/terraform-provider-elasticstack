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

package rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ModifyPlan keeps a reformat-only edit of artifacts[].data from planning an
// update. The framework applies jsontypes.Normalized semantic equality after
// writes and reads but not while planning, and it marks every computed
// attribute unknown as soon as the proposed state differs from the prior
// state, so the prior data is restored here and, when nothing else changed,
// so are version and updated_at.
func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plannedArtifacts types.List
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, artifactsPath, &plannedArtifacts)...)
	if resp.Diagnostics.HasError() || !typeutils.IsKnown(plannedArtifacts) {
		return
	}

	var plan, state Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || len(plan.Artifacts) != len(state.Artifacts) {
		return
	}

	restored := false
	for i := range plan.Artifacts {
		if !plan.Artifacts[i].Data.IsUnknown() && !state.Artifacts[i].Data.IsNull() {
			equal, diags := state.Artifacts[i].Data.StringSemanticEquals(ctx, plan.Artifacts[i].Data)
			resp.Diagnostics.Append(diags...)
			if equal && !plan.Artifacts[i].Data.Equal(state.Artifacts[i].Data) {
				plan.Artifacts[i].Data = state.Artifacts[i].Data
				restored = true
			}
		}
	}
	if !restored || resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	candidate := tfsdk.Plan{Schema: resp.Plan.Schema, Raw: resp.Plan.Raw}
	resp.Diagnostics.Append(candidate.SetAttribute(ctx, path.Root("version"), state.Version)...)
	resp.Diagnostics.Append(candidate.SetAttribute(ctx, path.Root("updated_at"), state.UpdatedAt)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if candidate.Raw.Equal(req.State.Raw) {
		resp.Plan = candidate
	}
}
