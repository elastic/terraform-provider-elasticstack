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
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
//
// It operates on the artifacts list as framework values rather than decoding
// the whole Model: a wholly-unknown nested object or artifact element (for
// example one derived from another resource's not-yet-known value) cannot
// decode into a Go struct, so a whole-plan Get would fail with a
// value-conversion error. Unknown elements are left untouched.
func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plannedArtifacts, stateArtifacts types.List
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, artifactsPath, &plannedArtifacts)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, artifactsPath, &stateArtifacts)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !typeutils.IsKnown(plannedArtifacts) || !typeutils.IsKnown(stateArtifacts) {
		return
	}

	planElems := plannedArtifacts.Elements()
	stateElems := stateArtifacts.Elements()
	if len(planElems) != len(stateElems) {
		return
	}

	restored := false
	newElems := make([]attr.Value, len(planElems))
	copy(newElems, planElems)
	for i := range planElems {
		planObj, planOK := planElems[i].(types.Object)
		stateObj, stateOK := stateElems[i].(types.Object)
		if !planOK || !stateOK || planObj.IsUnknown() || stateObj.IsUnknown() {
			continue
		}
		planData, planOK := planObj.Attributes()["data"].(jsontypes.Normalized)
		stateData, stateOK := stateObj.Attributes()["data"].(jsontypes.Normalized)
		if !planOK || !stateOK || planData.IsUnknown() || stateData.IsNull() {
			continue
		}
		equal, diags := stateData.StringSemanticEquals(ctx, planData)
		resp.Diagnostics.Append(diags...)
		if equal && !planData.Equal(stateData) {
			attrs := maps.Clone(planObj.Attributes())
			attrs["data"] = stateData
			restoredObj, objDiags := types.ObjectValue(planObj.AttributeTypes(ctx), attrs)
			resp.Diagnostics.Append(objDiags...)
			newElems[i] = restoredObj
			restored = true
		}
	}
	if !restored || resp.Diagnostics.HasError() {
		return
	}

	restoredArtifacts, listDiags := types.ListValue(plannedArtifacts.ElementType(ctx), newElems)
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, artifactsPath, restoredArtifacts)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// When the reformat was the only change, the server-owned version and
	// updated_at keep their prior values too, so the plan is a true no-op.
	var stateVersion types.Int64
	var stateUpdatedAt types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("version"), &stateVersion)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("updated_at"), &stateUpdatedAt)...)
	if resp.Diagnostics.HasError() {
		return
	}
	candidate := tfsdk.Plan{Schema: resp.Plan.Schema, Raw: resp.Plan.Raw}
	resp.Diagnostics.Append(candidate.SetAttribute(ctx, path.Root("version"), stateVersion)...)
	resp.Diagnostics.Append(candidate.SetAttribute(ctx, path.Root("updated_at"), stateUpdatedAt)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if candidate.Raw.Equal(req.State.Raw) {
		resp.Plan = candidate
	}
}
