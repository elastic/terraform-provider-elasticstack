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

package cloudconnector

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// reconcileDualRepresentationPlan copies read-populated sibling attributes from
// state into plan so Optional (non-Computed) aws/vars parents do not show
// spurious removal diffs after Read dual-populates both representations.
func reconcileDualRepresentationPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	config cloudConnectorModel,
) {
	var state cloudConnectorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan cloudConnectorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch {
	case isNestedBlockConfigured(config.AWS):
		copyStateSiblingToPlan(ctx, resp, state.Vars, path.Root(attrVarsMap), config.Vars)
	case isNestedBlockConfigured(config.Azure):
		copyStateSiblingToPlan(ctx, resp, state.Vars, path.Root(attrVarsMap), config.Vars)
	case isVarsMapConfigured(config.Vars):
		if isNestedBlockConfigured(config.AWS) || isNestedBlockConfigured(config.Azure) {
			return
		}
		if !plan.Vars.Equal(state.Vars) {
			return
		}
		switch {
		case isNestedBlockConfigured(state.AWS):
			copyStateSiblingToPlan(ctx, resp, state.AWS, path.Root(attrAWSBlock), config.AWS)
		case isNestedBlockConfigured(state.Azure):
			copyStateSiblingToPlan(ctx, resp, state.Azure, path.Root(attrAzureBlock), config.Azure)
		}
	}
}

func isNestedBlockConfigured(block types.Object) bool {
	return typeutils.IsKnown(block)
}

func isVarsMapConfigured(vars types.Map) bool {
	return typeutils.IsKnown(vars)
}

func copyStateSiblingToPlan(
	ctx context.Context,
	resp *resource.ModifyPlanResponse,
	stateValue attr.Value,
	planPath path.Path,
	configValue attr.Value,
) {
	if typeutils.IsKnown(configValue) {
		return
	}
	if stateValue.IsNull() {
		return
	}
	if stateValue.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Skipped dual representation plan reconciliation",
			fmt.Sprintf("Could not copy %s from state into plan because the state value is unknown.", planPath.String()),
		)
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, planPath, stateValue)...)
}
