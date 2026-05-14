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

package sloburnrate_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/sloburnrate"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/stretchr/testify/require"
)

func TestContract(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, sloburnrate.Handler{}, contracttest.Config{
		FullAPIResponse: `{
			"type": "slo_burn_rate",
			"grid": {"x": 0, "y": 0, "w": 24, "h": 6},
			"id": "slo-burn-contract",
			"config": {
				"slo_id": "my-slo-id",
				"duration": "5m",
				"title": "Burn Rate",
				"drilldowns": [{
					"type": "url_drilldown",
					"trigger": "on_open_panel_menu",
					"url": "https://example.com/panel",
					"label": "Open"
				}]
			}
		}`,
		// Drilldowns are always refreshed from the API when slo_burn_rate_config is present — null-preservation
		// on the drilldown list is not modeled; omit harness checks on that collection (config path for round-trip; leaf name for NullPreserve walker).
		SkipFields: []string{"config.drilldowns", "drilldowns"},
	})
}

func TestDrilldowns_roundTrip_viaHandler(t *testing.T) {
	t.Parallel()

	const fixtureJSON = `{
	  "type": "slo_burn_rate",
	  "grid": { "x": 0, "y": 0, "w": 24, "h": 8 },
	  "id": "slo-burn-dd",
	  "config": {
	    "sloId": "slo-1",
	    "duration": "6d",
	    "title": "With drilldown",
	    "drilldowns": [
	      {
	        "type": "url_drilldown",
	        "trigger": "on_open_panel_menu",
	        "url": "https://example.com/{{context.panel.title}}",
	        "label": "View details"
	      },
	      {
	        "type": "url_drilldown",
	        "trigger": "on_open_panel_menu",
	        "url": "https://kibana/host",
	        "label": "Host",
	        "encode_url": true,
	        "open_in_new_tab": false
	      }
	    ]
	  }
	}`

	item0, err := contracttest.ParseDashboardPanel(fixtureJSON)
	require.NoError(t, err)

	var pm models.PanelModel
	handler := sloburnrate.Handler{}
	diags := handler.FromAPI(context.Background(), &pm, nil, item0)
	require.False(t, diags.HasError(), "%s", diags)

	require.Len(t, pm.SloBurnRateConfig.Drilldowns, 2)
	dd0 := pm.SloBurnRateConfig.Drilldowns[0]
	require.Equal(t, "https://example.com/{{context.panel.title}}", dd0.URL.ValueString())
	require.Equal(t, "View details", dd0.Label.ValueString())
	require.True(t, dd0.EncodeURL.IsNull())
	require.True(t, dd0.OpenInNewTab.IsNull())

	dd1 := pm.SloBurnRateConfig.Drilldowns[1]
	require.True(t, dd1.EncodeURL.ValueBool())
	require.False(t, dd1.OpenInNewTab.ValueBool())

	item1, d2 := handler.ToAPI(pm, nil)
	require.False(t, d2.HasError(), "%s", d2)

	api0 := jsonPanelMap(t, item0)
	api1 := jsonPanelMap(t, item1)

	cfg0 := mustConfigMap(t, api0["config"])
	cfg1 := mustConfigMap(t, api1["config"])
	ddA, ddB := drillsFromConfig(cfg0), drillsFromConfig(cfg1)

	require.Len(t, ddB, len(ddA))
	require.Equal(t, ddA[0]["url"], ddB[0]["url"])
	require.Equal(t, ddA[0]["label"], ddB[0]["label"])
	require.Equal(t, string(kbapi.SloBurnRateEmbeddableDrilldownsTriggerOnOpenPanelMenu), ddB[0]["trigger"])
	require.Equal(t, string(kbapi.SloBurnRateEmbeddableDrilldownsTypeUrlDrilldown), ddB[0]["type"])

	require.Equal(t, true, drillsFromConfig(cfg1)[1]["encode_url"])
	require.Equal(t, false, drillsFromConfig(cfg1)[1]["open_in_new_tab"])
}

func jsonPanelMap(t *testing.T, item kbapi.DashboardPanelItem) map[string]any {
	t.Helper()
	raw, err := json.Marshal(item)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	return m
}

func mustConfigMap(t *testing.T, v any) map[string]any {
	t.Helper()
	cfg, ok := v.(map[string]any)
	require.True(t, ok, "config should be object")
	return cfg
}

func drillsFromConfig(cfg map[string]any) []map[string]any {
	raw, ok := cfg["drilldowns"].([]any)
	if !ok || len(raw) == 0 {
		return nil
	}
	out := make([]map[string]any, len(raw))
	for i, e := range raw {
		out[i], _ = e.(map[string]any)
	}
	return out
}
