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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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

	needsUpdate := false

	for key, elem := range configMap {
		if !typeutils.IsKnown(elem.SecretValue) {
			continue
		}
		value := elem.SecretValue.ValueString()
		storedHash, diags := req.Private.GetKey(ctx, secretHashKey(key))
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(storedHash) == 0 {
			continue
		}
		if !secretHasher.Matches(value, storedHash) {
			needsUpdate = true
			resp.Diagnostics.AddWarning(
				"Write-only attribute changed",
				fmt.Sprintf(
					`Detected a change to write-only attribute configuration_values["%s"].secret_value; the resource will be updated.`,
					key,
				),
			)
		}
	}

	for key, priorElem := range stateMap {
		if !typeutils.IsKnown(priorElem.SecretValue) {
			continue
		}
		if _, inConfig := configMap[key]; inConfig && typeutils.IsKnown(configMap[key].SecretValue) {
			continue
		}
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, secretHashKey(key), nil)...)
	}

	if needsUpdate {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), fwtypes.StringUnknown())...)
	}
}
