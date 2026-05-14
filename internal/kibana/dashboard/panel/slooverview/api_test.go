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

package slooverview_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/slooverview"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/stretchr/testify/require"
)

func TestContract(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, slooverview.Handler{}, contracttest.Config{
		// Required-leaf-presence compares raw fixture.config to Terraform slo_overview_config.single|groups.* paths via a struct
		// navigator. Kibana discriminator output flattens single-mode (`overview_mode`, `slo_id`, top-level drilldowns) whereas
		// Terraform nests display + drilldowns under exactly one child block (`single` or `groups`); omit the phase here.
		FullAPIResponse: `{
			"type": "slo_overview",
			"grid": {"x": 0, "y": 0, "w": 24, "h": 8},
			"id": "slo-overview-contract",
			"config": {
				"overview_mode": "single",
				"slo_id": "my-slo",
				"title": "Overview",
				"drilldowns": [{
					"type": "url_drilldown",
					"trigger": "on_open_panel_menu",
					"url": "https://example.com/panel",
					"label": "Open"
				}]
			}
		}`,
		OmitRequiredLeafPresence: true,
		SkipFields: []string{
			// Baseline fixture is single-mode; optional groups.* branches are unreadable on the hydrated model baseline.
			"groups",
			"config.drilldowns",
			"single.drilldowns",
			"groups.drilldowns",
			"drilldowns",
			"single.hide_title",
			"single.hide_border",
			"single.slo_instance_id",
			"single.remote_name",
			"single.title",
			"single.description",
			"groups.hide_title",
			"groups.hide_border",
			"groups.title",
			"groups.description",
			"config.title",
			"title",
		},
	})
}

func TestSloOverview_singleMode_drilldowns_roundTrip_viaHandler(t *testing.T) {
	t.Parallel()

	const fixtureJSON = `{
	  "type": "slo_overview",
	  "grid": { "x": 0, "y": 0, "w": 24, "h": 8 },
	  "id": "slo-ov-single",
	  "config": {
	    "overview_mode": "single",
	    "slo_id": "svc-slo-alpha",
	    "title": "One SLO",
	    "description": "Unit test panel",
	    "drilldowns": [
	      {
	        "type": "url_drilldown",
	        "trigger": "on_open_panel_menu",
	        "url": "https://example.com/{{context.panel.title}}",
	        "label": "Open context"
	      },
	      {
	        "type": "url_drilldown",
	        "trigger": "on_open_panel_menu",
	        "url": "https://kibana/app",
	        "label": "Kibana home",
	        "encode_url": true,
	        "open_in_new_tab": false
	      }
	    ]
	  }
	}`

	item0, err := contracttest.ParseDashboardPanel(fixtureJSON)
	require.NoError(t, err)

	var pm models.PanelModel
	diags := slooverview.Handler{}.FromAPI(context.Background(), &pm, nil, item0)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.SloOverviewConfig)
	require.NotNil(t, pm.SloOverviewConfig.Single)
	require.Len(t, pm.SloOverviewConfig.Single.Drilldowns, 2)
	ddA := pm.SloOverviewConfig.Single.Drilldowns
	require.Equal(t, "https://example.com/{{context.panel.title}}", ddA[0].URL.ValueString())
	require.True(t, ddA[0].EncodeURL.IsNull())
	ddB := ddA[1]
	require.True(t, ddB.EncodeURL.ValueBool())
	require.False(t, ddB.OpenInNewTab.ValueBool())

	item1, td := slooverview.Handler{}.ToAPI(pm, nil)
	require.False(t, td.HasError(), "%s", td)

	api0 := jsonPanelMap(t, item0)
	api1 := jsonPanelMap(t, item1)
	cfg0 := mustConfigMap(t, api0["config"])
	cfg1 := mustConfigMap(t, api1["config"])
	d0, d1 := drillsFromConfig(cfg0), drillsFromConfig(cfg1)
	require.Len(t, d1, len(d0))
	require.Equal(t, d0[0]["url"], d1[0]["url"])
	require.Equal(t, d0[0]["label"], d1[0]["label"])
	require.Equal(t, string(kbapi.SloSingleOverviewEmbeddableDrilldownsTriggerOnOpenPanelMenu), d1[0]["trigger"])
	require.Equal(t, string(kbapi.SloSingleOverviewEmbeddableDrilldownsTypeUrlDrilldown), d1[0]["type"])
	require.Equal(t, true, d1[1]["encode_url"])
	require.Equal(t, false, d1[1]["open_in_new_tab"])
}

func TestSloOverview_groupsMode_drilldowns_roundTrip_viaHandler(t *testing.T) {
	t.Parallel()

	const fixtureJSON = `{
	  "type": "slo_overview",
	  "grid": { "x": 0, "y": 0, "w": 24, "h": 10 },
	  "id": "slo-ov-groups",
	  "config": {
	    "overview_mode": "groups",
	    "title": "Grouped overview",
	    "group_filters": { "group_by": "slo.tags" },
	    "drilldowns": [
	      {
	        "type": "url_drilldown",
	        "trigger": "on_open_panel_menu",
	        "url": "https://example.com/groups",
	        "label": "Explore group"
	      }
	    ]
	  }
	}`

	item0, err := contracttest.ParseDashboardPanel(fixtureJSON)
	require.NoError(t, err)

	var pm models.PanelModel
	diags := slooverview.Handler{}.FromAPI(context.Background(), &pm, nil, item0)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.SloOverviewConfig)
	require.NotNil(t, pm.SloOverviewConfig.Groups)
	require.Len(t, pm.SloOverviewConfig.Groups.Drilldowns, 1)
	require.Equal(t, "https://example.com/groups", pm.SloOverviewConfig.Groups.Drilldowns[0].URL.ValueString())

	item1, td := slooverview.Handler{}.ToAPI(pm, nil)
	require.False(t, td.HasError(), "%s", td)

	api0 := jsonPanelMap(t, item0)
	api1 := jsonPanelMap(t, item1)
	cfg0 := mustConfigMap(t, api0["config"])
	cfg1 := mustConfigMap(t, api1["config"])
	d0, d1 := drillsFromConfig(cfg0), drillsFromConfig(cfg1)
	require.Len(t, d1, len(d0))
	require.Equal(t, d0[0]["url"], d1[0]["url"])
	require.Equal(t, d0[0]["label"], d1[0]["label"])
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
