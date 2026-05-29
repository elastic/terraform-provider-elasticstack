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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ModifyPlan detects drift on write-only secret attributes by comparing config
// values against bcrypt hashes stored in resource private state.
func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var config cloudConnectorModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Dual-populated siblings are optional+computed with UseStateForUnknown; the
	// create plan leaves unconfigured siblings unknown so apply can populate them.
	if req.State.Raw.IsNull() {
		if isVarsMapConfigured(config.Vars) {
			varsMap, varsDiags := planVarsForCreate(ctx, config.Vars)
			resp.Diagnostics.Append(varsDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrVarsMap), varsMap)...)
		}
		return
	}

	var priv privateData
	if r.testModifyPlanPrivate != nil {
		priv = r.testModifyPlanPrivate
	} else if req.Private != nil {
		priv = req.Private
	}
	hasher := cloudConnectorHasher()
	results, driftDiags := evaluateWriteOnlyDrift(ctx, hasher, config, priv)
	resp.Diagnostics.Append(driftDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(results) > 0 {
		if r.pendingWriteOnlyResubmit == nil {
			r.pendingWriteOnlyResubmit = make(map[string]struct{})
		}
		for _, result := range results {
			resp.Diagnostics.Append(driftWarningDiagnostic(result))
			if !result.IsImportBaseline {
				r.pendingWriteOnlyResubmit[result.AttributePath] = struct{}{}
			}
		}
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrUpdatedAt), types.StringUnknown())...)
	}

	reconcileDualRepresentationPlan(ctx, req, resp, config)
	markComputedRefreshFieldsUnknown(ctx, req, resp, config, len(results) > 0)
}

func markComputedRefreshFieldsUnknown(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	config cloudConnectorModel,
	writeOnlyDriftDetected bool,
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

	if planMutatesAPIResource(state, plan, config) || writeOnlyDriftDetected {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrUpdatedAt), types.StringUnknown())...)
	}

	if planMutatesAPIResource(state, plan, config) && !writeOnlyDriftDetected && !typedBlockOrVarsChanged(state, plan, config) {
		refreshDualRepresentationFromState(ctx, resp, state, config)
	}
}

// refreshDualRepresentationFromState copies read-populated dual representation
// fields from state into plan for metadata-only updates (e.g. name changes).
func refreshDualRepresentationFromState(
	ctx context.Context,
	resp *resource.ModifyPlanResponse,
	state cloudConnectorModel,
	config cloudConnectorModel,
) {
	if typeutils.IsKnown(state.Vars) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrVarsMap), state.Vars)...)
	}

	switch {
	case isNestedBlockConfigured(config.AWS) && typeutils.IsKnown(state.AWS):
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrAWSBlock), state.AWS)...)
	case isNestedBlockConfigured(config.Azure) && typeutils.IsKnown(state.Azure):
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrAzureBlock), state.Azure)...)
	case isVarsMapConfigured(config.Vars):
		switch config.CloudProvider.ValueString() {
		case cloudProviderAWS:
			if typeutils.IsKnown(state.AWS) {
				resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrAWSBlock), state.AWS)...)
			}
		case cloudProviderAzure:
			if typeutils.IsKnown(state.Azure) {
				resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrAzureBlock), state.Azure)...)
			}
		}
	}
}

// typedBlockOrVarsChanged reports whether the configured representation changed
// in ways that require recompiling wire vars rather than a metadata-only refresh.
func typedBlockOrVarsChanged(state, plan, config cloudConnectorModel) bool {
	if isNestedBlockConfigured(config.AWS) && !awsBlockEqualIgnoringWriteOnly(plan.AWS, state.AWS) {
		return true
	}
	if isNestedBlockConfigured(config.Azure) && !azureBlockEqualIgnoringWriteOnly(plan.Azure, state.Azure) {
		return true
	}
	if isVarsMapConfigured(config.Vars) && !plan.Vars.Equal(state.Vars) {
		return true
	}
	return false
}

func awsBlockEqualIgnoringWriteOnly(planBlock, stateBlock types.Object) bool {
	if planBlock.Equal(stateBlock) {
		return true
	}
	planAWS, planDiags := awsBlockFromObject(planBlock)
	stateAWS, stateDiags := awsBlockFromObject(stateBlock)
	if planDiags.HasError() || stateDiags.HasError() {
		return false
	}
	return planAWS.RoleArn.Equal(stateAWS.RoleArn) &&
		planAWS.ExternalIDSecretRef.Equal(stateAWS.ExternalIDSecretRef)
}

func azureBlockEqualIgnoringWriteOnly(planBlock, stateBlock types.Object) bool {
	if planBlock.Equal(stateBlock) {
		return true
	}
	planAzure, planDiags := azureBlockFromObject(planBlock)
	stateAzure, stateDiags := azureBlockFromObject(stateBlock)
	if planDiags.HasError() || stateDiags.HasError() {
		return false
	}
	return planAzure.CloudConnectorID.Equal(stateAzure.CloudConnectorID) &&
		planAzure.TenantIDSecretRef.Equal(stateAzure.TenantIDSecretRef) &&
		planAzure.ClientIDSecretRef.Equal(stateAzure.ClientIDSecretRef)
}

// planMutatesAPIResource reports whether the current plan would issue an update
// that refreshes computed API fields such as updated_at.
func planMutatesAPIResource(state, plan, config cloudConnectorModel) bool {
	if !plan.Name.Equal(state.Name) {
		return true
	}
	if !plan.AccountType.Equal(state.AccountType) {
		return true
	}
	if isNestedBlockConfigured(config.AWS) && !plan.AWS.Equal(state.AWS) {
		return true
	}
	if isNestedBlockConfigured(config.Azure) && !plan.Azure.Equal(state.Azure) {
		return true
	}
	if isVarsMapConfigured(config.Vars) && !plan.Vars.Equal(state.Vars) {
		return true
	}
	return false
}
