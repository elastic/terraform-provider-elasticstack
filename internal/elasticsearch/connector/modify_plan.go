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

package connector

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *contentConnectorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var config ContentConnectorData
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ContentConnectorData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configMap := configurationValuesFromModel(ctx, config.ConfigurationValues, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	stateMap := configurationValuesFromModel(ctx, state.ConfigurationValues, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	outcome, evalDiags := evaluateSecretPlanChanges(configMap, stateMap, func(key string) ([]byte, diag.Diagnostics) {
		var diags diag.Diagnostics
		raw, getDiags := req.Private.GetKey(ctx, secretHashKey(key))
		diags.Append(getDiags...)
		if diags.HasError() {
			return nil, diags
		}
		decoded, err := decodeSecretHashFromPrivateState(raw)
		if err != nil {
			diags.AddError("Failed to decode write-only secret hash", err.Error())
			return nil, diags
		}
		return decoded, diags
	})
	resp.Diagnostics.Append(evalDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, warning := range outcome.Warnings {
		resp.Diagnostics.AddWarning("Write-only attribute changed", warning)
	}

	for _, key := range outcome.KeysToClear {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, secretHashKey(key), nil)...)
	}

	if outcome.NeedsUpdate {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), fwtypes.StringUnknown())...)
	}
}
