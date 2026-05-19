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

package spaces

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ModifyPlan runs after all per-attribute plan modifiers and coordinates the
// interaction between `solution` and `disabled_features`. Kibana derives the
// effective `disabled_features` from `solution` whenever a solution is set,
// so any user-driven change to `solution` must invalidate the cached
// `disabled_features` value carried forward by `UseStateForUnknown`.
//
// Running this at the resource level (rather than as a sibling attribute
// plan modifier) guarantees that `solution`'s own `UseStateForUnknown`
// modifier has already resolved its plan value, so we compare against the
// final planned `solution` rather than a transiently-unknown intermediate.
func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to coordinate on create (no prior solution) or destroy.
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var state, plan, config resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Respect an explicit configuration value for disabled_features. The
	// ConflictsWith validator already prevents this from coexisting with
	// solution in config.
	if typeutils.IsKnown(config.DisabledFeatures) {
		return
	}

	// If the effective solution hasn't changed, the state-carried
	// disabled_features value remains accurate.
	if state.Solution.Equal(plan.Solution) {
		return
	}

	resp.Diagnostics.Append(
		resp.Plan.SetAttribute(ctx, path.Root("disabled_features"), types.SetUnknown(types.StringType))...,
	)
}
