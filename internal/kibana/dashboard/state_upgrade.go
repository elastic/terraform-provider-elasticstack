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

package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/optionslist"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/rangeslider"
	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// UpgradeState migrates dashboard resource state from schema version 0 to
// version 1. Version 0 stored options_list_control_config and
// range_slider_control_config as a flat set of attributes; version 1
// restructures both blocks into a `by_field {}` / `by_esql {}` union so the
// ES|QL control variant can be represented. See REQ-040.
func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: migrateV0ToV1,
		},
	}
}

func migrateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	stateutil.SetDefaultState(req, resp)

	state := stateutil.UnmarshalStateMap(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	migratePanelList(state[attrPanels])
	migratePanelList(state[attrPinnedPanels])

	// Panels nested inside dashboard `sections` share the exact same panel
	// envelope (type + *_config blocks) as top-level panels, so they need the
	// same relocation to avoid silently dropping control-panel state for
	// sectioned dashboards.
	if sections, ok := state[attrSections].([]any); ok {
		for _, raw := range sections {
			section, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			migratePanelList(section[attrPanels])
		}
	}

	stateutil.MarshalStateMap(state, resp)
}

// migratePanelList relocates the v0 flat control-config attributes for every
// options_list_control / range_slider_control entry in a panels-shaped list
// (used for the top-level `panels`, `pinned_panels`, and `sections[].panels`
// lists, which all share the same per-panel envelope). Entries of any other
// panel type are left untouched.
func migratePanelList(raw any) {
	panels, ok := raw.([]any)
	if !ok {
		return
	}
	for _, p := range panels {
		panel, ok := p.(map[string]any)
		if !ok {
			continue
		}

		panelType, _ := panel[attrPanelType].(string)
		switch panelType {
		case panelTypeOptionsListControl:
			relocateToByField(panel, controlBlockOptionsList, optionslist.BranchByField, optionslist.BranchByEsql, optionslist.ByFieldAttributeNames())
		case panelTypeRangeSlider:
			relocateToByField(panel, controlBlockRangeSlider, rangeslider.BranchByField, rangeslider.BranchByEsql, rangeslider.ByFieldAttributeNames())
		}
	}
}

// relocateToByField moves the given v0 flat attribute keys out of the
// configKey block on panel and into a nested `by_field {}` object, leaving
// `by_esql` unset (null) since v0 state never had an ES|QL branch. Values are
// moved as-is (no re-encoding), so numeric types (e.g. `step` as float64 post
// json.Unmarshal) are preserved unchanged. Missing or null configKey blocks
// (e.g. because the panel is of a different type, or the block was never set)
// are left untouched. byFieldKey/byEsqlKey and flatAttrs come from the owning
// panel package (optionslist/rangeslider) so this stays in sync with the
// schema instead of duplicating its attribute names and branch keys.
func relocateToByField(panel map[string]any, configKey, byFieldKey, byEsqlKey string, flatAttrs []string) {
	rawConfig, ok := panel[configKey]
	if !ok || rawConfig == nil {
		return
	}
	config, ok := rawConfig.(map[string]any)
	if !ok {
		return
	}

	byField := make(map[string]any, len(flatAttrs))
	for _, attr := range flatAttrs {
		if v, exists := config[attr]; exists {
			byField[attr] = v
			delete(config, attr)
		}
	}
	// Defensive: Terraform state normally serialises every attribute of an
	// object type (present, possibly null), but ensure all expected keys are
	// present so the v1 SingleNestedAttribute decodes cleanly even if a hand
	// crafted or otherwise incomplete v0 state omitted one.
	stateutil.EnsureMapKeys(byField, flatAttrs...)

	config[byFieldKey] = byField
	config[byEsqlKey] = nil
	panel[configKey] = config
}
