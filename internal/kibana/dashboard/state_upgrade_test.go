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
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func runMigrateV0ToV1(t *testing.T, raw map[string]any) map[string]any {
	t.Helper()

	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)

	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: rawJSON}}
	resp := &resource.UpgradeStateResponse{}
	migrateV0ToV1(context.Background(), req, resp)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)

	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
	return got
}

func TestMigrateV0ToV1_OptionsListControl(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"panels": []any{
			map[string]any{
				"type": panelTypeOptionsListControl,
				"id":   "ol-1",
				"options_list_control_config": map[string]any{
					"data_view_id":       "logs-view",
					"field_name":         "service.name",
					"title":              "Service",
					"use_global_filters": true,
					"ignore_validations": false,
					"single_select":      true,
					"exclude":            false,
					"exists_selected":    false,
					"run_past_timeout":   true,
					"search_technique":   "prefix",
					"selected_options":   []any{"auth-service"},
					"display_settings": map[string]any{
						"placeholder":     "Pick a service",
						"hide_action_bar": true,
						"hide_exclude":    false,
						"hide_exists":     false,
						"hide_sort":       false,
					},
					"sort": map[string]any{
						"by":        "_count",
						"direction": "desc",
					},
				},
			},
		},
	}

	got := runMigrateV0ToV1(t, raw)

	panels, ok := got["panels"].([]any)
	require.True(t, ok)
	require.Len(t, panels, 1)
	panel, ok := panels[0].(map[string]any)
	require.True(t, ok)

	cfg, ok := panel["options_list_control_config"].(map[string]any)
	require.True(t, ok)
	require.Nil(t, cfg["by_esql"])

	byField, ok := cfg["by_field"].(map[string]any)
	require.True(t, ok, "expected v0 flat attributes to be relocated under by_field")

	require.Equal(t, "logs-view", byField["data_view_id"])
	require.Equal(t, "service.name", byField["field_name"])
	require.Equal(t, "Service", byField["title"])
	require.Equal(t, true, byField["use_global_filters"])
	require.Equal(t, false, byField["ignore_validations"])
	require.Equal(t, true, byField["single_select"])
	require.Equal(t, false, byField["exclude"])
	require.Equal(t, false, byField["exists_selected"])
	require.Equal(t, true, byField["run_past_timeout"])
	require.Equal(t, "prefix", byField["search_technique"])
	require.Equal(t, []any{"auth-service"}, byField["selected_options"])

	displaySettings, ok := byField["display_settings"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Pick a service", displaySettings["placeholder"])
	require.Equal(t, true, displaySettings["hide_action_bar"])

	sort, ok := byField["sort"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "_count", sort["by"])
	require.Equal(t, "desc", sort["direction"])

	// No v0 attribute should remain directly on the config block; every key now
	// lives either at by_field/by_esql or was never a v0 attribute.
	for _, k := range optionsListV0FlatAttrs {
		_, exists := cfg[k]
		require.False(t, exists, "v0 attribute %q should have been relocated out of the config root", k)
	}
}

func TestMigrateV0ToV1_RangeSliderControl(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"panels": []any{
			map[string]any{
				"type": panelTypeRangeSlider,
				"id":   "rs-1",
				"range_slider_control_config": map[string]any{
					"data_view_id":       "orders-view",
					"field_name":         "price",
					"title":              "Price",
					"use_global_filters": true,
					"ignore_validations": false,
					"value":              []any{"10", "500"},
					"step":               5.0,
				},
			},
		},
	}

	got := runMigrateV0ToV1(t, raw)

	panels, ok := got["panels"].([]any)
	require.True(t, ok)
	panel, ok := panels[0].(map[string]any)
	require.True(t, ok)

	cfg, ok := panel["range_slider_control_config"].(map[string]any)
	require.True(t, ok)
	require.Nil(t, cfg["by_esql"])

	byField, ok := cfg["by_field"].(map[string]any)
	require.True(t, ok, "expected v0 flat attributes to be relocated under by_field")

	require.Equal(t, "orders-view", byField["data_view_id"])
	require.Equal(t, "price", byField["field_name"])
	require.Equal(t, "Price", byField["title"])
	require.Equal(t, true, byField["use_global_filters"])
	require.Equal(t, false, byField["ignore_validations"])
	require.Equal(t, []any{"10", "500"}, byField["value"])
	require.Equal(t, 5.0, byField["step"])

	for _, k := range rangeSliderV0FlatAttrs {
		_, exists := cfg[k]
		require.False(t, exists, "v0 attribute %q should have been relocated out of the config root", k)
	}
}

// TestMigrateV0ToV1_NonControlPanelsUnaffected covers the REQ-040 scenario "Non-control panels are
// unaffected by the upgrader": a markdown panel mixed alongside options_list/range_slider panels
// must be left byte-for-byte unchanged by the migration.
func TestMigrateV0ToV1_NonControlPanelsUnaffected(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"panels": []any{
			map[string]any{
				"type": "markdown",
				"id":   "md-1",
				"markdown_config": map[string]any{
					"content": "# Hello",
				},
			},
			map[string]any{
				"type": panelTypeOptionsListControl,
				"id":   "ol-1",
				"options_list_control_config": map[string]any{
					"data_view_id": "logs-view",
					"field_name":   "service.name",
				},
			},
			map[string]any{
				"type": panelTypeRangeSlider,
				"id":   "rs-1",
				"range_slider_control_config": map[string]any{
					"data_view_id": "orders-view",
					"field_name":   "price",
				},
			},
		},
	}

	got := runMigrateV0ToV1(t, raw)

	panels, ok := got["panels"].([]any)
	require.True(t, ok)
	require.Len(t, panels, 3)

	markdown, ok := panels[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "markdown", markdown["type"])
	mdCfg, ok := markdown["markdown_config"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "# Hello", mdCfg["content"])
	require.NotContains(t, markdown, "options_list_control_config")
	require.NotContains(t, markdown, "range_slider_control_config")

	ol, ok := panels[1].(map[string]any)
	require.True(t, ok)
	olCfg, ok := ol["options_list_control_config"].(map[string]any)
	require.True(t, ok)
	_, hasByField := olCfg["by_field"]
	require.True(t, hasByField)

	rs, ok := panels[2].(map[string]any)
	require.True(t, ok)
	rsCfg, ok := rs["range_slider_control_config"].(map[string]any)
	require.True(t, ok)
	_, hasByField = rsCfg["by_field"]
	require.True(t, hasByField)
}

// TestMigrateV0ToV1_PinnedPanelsAndSections covers REQ-040's requirement that pinned_panels
// entries, and panels nested inside dashboard sections, are migrated identically to top-level
// panels.
func TestMigrateV0ToV1_PinnedPanelsAndSections(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"pinned_panels": []any{
			map[string]any{
				"type": panelTypeOptionsListControl,
				"id":   "pinned-ol-1",
				"options_list_control_config": map[string]any{
					"data_view_id": "pinned-dv",
					"field_name":   "status",
				},
			},
		},
		"sections": []any{
			map[string]any{
				"id": "section-1",
				"panels": []any{
					map[string]any{
						"type": panelTypeRangeSlider,
						"id":   "sectioned-rs-1",
						"range_slider_control_config": map[string]any{
							"data_view_id": "orders-view",
							"field_name":   "price",
						},
					},
				},
			},
		},
	}

	got := runMigrateV0ToV1(t, raw)

	pinned, ok := got["pinned_panels"].([]any)
	require.True(t, ok)
	pinnedPanel, ok := pinned[0].(map[string]any)
	require.True(t, ok)
	pinnedCfg, ok := pinnedPanel["options_list_control_config"].(map[string]any)
	require.True(t, ok)
	pinnedByField, ok := pinnedCfg["by_field"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "pinned-dv", pinnedByField["data_view_id"])

	sections, ok := got["sections"].([]any)
	require.True(t, ok)
	section, ok := sections[0].(map[string]any)
	require.True(t, ok)
	sectionPanels, ok := section["panels"].([]any)
	require.True(t, ok)
	sectionPanel, ok := sectionPanels[0].(map[string]any)
	require.True(t, ok)
	sectionCfg, ok := sectionPanel["range_slider_control_config"].(map[string]any)
	require.True(t, ok)
	sectionByField, ok := sectionCfg["by_field"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "orders-view", sectionByField["data_view_id"])
}

func TestUpgradeState_RegistersV0Upgrader(t *testing.T) {
	t.Parallel()

	r := &Resource{}
	upgraders := r.UpgradeState(context.Background())
	up, ok := upgraders[0]
	require.True(t, ok, "expected a registered v0 state upgrader")
	require.NotNil(t, up.StateUpgrader)
}
