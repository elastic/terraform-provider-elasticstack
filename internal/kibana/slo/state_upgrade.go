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

package slo

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: migrateV0ToV2,
		},
		1: {
			StateUpgrader: migrateV1ToV2,
		},
	}
}

func migrateV0ToV2(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	migrateV0ToV1(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare a new request with the upgraded state from v0 to v1.
	v1ToV2Req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: resp.DynamicValue.JSON},
	}

	migrateV1ToV2(ctx, v1ToV2Req, resp)
}

func migrateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	stateutil.SetDefaultState(req, resp)

	stateMap := stateutil.UnmarshalStateMap(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	groupBy, ok := stateMap["group_by"]
	if !ok || groupBy == nil {
		return
	}

	groupByStr, ok := groupBy.(string)
	if !ok {
		return
	}

	if len(groupByStr) == 0 {
		return
	}

	stateMap["group_by"] = []any{groupByStr}

	stateutil.MarshalStateMap(stateMap, resp)
	resp.Diagnostics.AddAttributeWarning(
		path.Root("group_by"),
		"Upgraded legacy group_by state",
		fmt.Sprintf("Upgraded legacy group_by from string to list: %q", groupByStr),
	)
}

func migrateV1ToV2(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	stateutil.SetDefaultState(req, resp)

	stateMap := stateutil.UnmarshalStateMap(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, ok := stateMap["settings"]
	if !ok || settings == nil {
		return
	}

	// If settings already has the new shape (object), nothing to do.
	if _, ok := settings.(map[string]any); ok {
		return
	}

	settingsList, ok := settings.([]any)
	if !ok {
		return
	}
	if len(settingsList) == 0 {
		stateMap["settings"] = nil
	} else {
		first, ok := settingsList[0].(map[string]any)
		if !ok {
			return
		}
		stateMap["settings"] = first
	}

	stateutil.MarshalStateMap(stateMap, resp)
	resp.Diagnostics.AddAttributeWarning(
		path.Root("settings"),
		"Upgraded legacy settings state",
		"Upgraded legacy settings from a nested block to a single nested attribute.",
	)
}
