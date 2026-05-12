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

package settings

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// migrateClusterSettingsStateV0ToV1 reshapes state written by the SDKv2-based
// implementation (schema version 0) so it matches the Plugin Framework
// implementation (schema version 1):
//
//  1. The persistent and transient blocks were stored as a list-of-one
//     (TypeList with MaxItems=1); they are now SingleNestedBlock objects.
//     Unwrap [obj] -> obj and remove empty lists.
//  2. Each setting object had value="" / value_list=[] for the unused
//     alternative; the framework's null representation is required for
//     set-element identity to match the new flatten output.
func migrateClusterSettingsStateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	var stateMap map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &stateMap); err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not unmarshal prior state: "+err.Error())
		return
	}

	for _, category := range []string{"persistent", "transient"} {
		if err := unwrapAndNormaliseCategoryBlock(stateMap, category); err != nil {
			resp.Diagnostics.AddError("State upgrade error", err.Error())
			return
		}
	}

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not marshal new state: "+err.Error())
		return
	}
	resp.DynamicValue = &tfprotov6.DynamicValue{JSON: stateJSON}
}

// unwrapAndNormaliseCategoryBlock unwraps a persistent/transient list-of-one
// into a single object (or removes the key entirely when the list is empty
// or absent), and normalises every nested setting's value/value_list.
func unwrapAndNormaliseCategoryBlock(stateMap map[string]any, category string) error {
	raw, ok := stateMap[category]
	if !ok || raw == nil {
		delete(stateMap, category)
		return nil
	}
	blocks, ok := raw.([]any)
	if !ok {
		return fmt.Errorf("expected %q to be a list from prior SDK state, got %T", category, raw)
	}
	if len(blocks) == 0 {
		delete(stateMap, category)
		return nil
	}
	if len(blocks) > 1 {
		return fmt.Errorf("expected %q to contain at most one block from prior SDK state, got %d", category, len(blocks))
	}
	blockObj, ok := blocks[0].(map[string]any)
	if !ok {
		delete(stateMap, category)
		return nil
	}
	if settings, ok := blockObj["setting"].([]any); ok {
		for _, s := range settings {
			settingObj, ok := s.(map[string]any)
			if !ok {
				continue
			}
			normaliseSettingValues(settingObj)
		}
	}
	stateMap[category] = blockObj
	return nil
}

// normaliseSettingValues converts SDK zero-value representations to nulls.
// JSON-encoded null is represented as the absence of the key, so deleting the
// key produces the desired null value when the framework decodes the state.
func normaliseSettingValues(setting map[string]any) {
	if v, ok := setting["value"]; ok {
		if s, isStr := v.(string); isStr && s == "" {
			delete(setting, "value")
		}
	}
	if v, ok := setting["value_list"]; ok {
		switch t := v.(type) {
		case nil:
			delete(setting, "value_list")
		case []any:
			if len(t) == 0 {
				delete(setting, "value_list")
			}
		}
	}
}
