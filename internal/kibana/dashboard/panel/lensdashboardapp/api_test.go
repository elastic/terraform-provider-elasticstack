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

package lensdashboardapp_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard" // register lens converters + JSON defaults
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensdashboardapp"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustLensDashboardAppPanelItem(t *testing.T, vis0 kbapi.KbnDashboardPanelTypeVisConfig0) kbapi.DashboardPanelItem {
	t.Helper()
	payload, err := json.Marshal(vis0)
	require.NoError(t, err)
	var lens0 kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0
	require.NoError(t, json.Unmarshal(payload, &lens0))
	var cfg kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	require.NoError(t, cfg.FromKbnDashboardPanelTypeLensDashboardAppConfig0(lens0))
	w, h := float32(24), float32(12)
	id := "panel-id"
	panel := kbapi.KbnDashboardPanelTypeLensDashboardApp{
		Config: cfg,
		Grid: kbapi.KbnDashboardPanelGrid{
			X: 0, Y: 0, W: &w, H: &h,
		},
		Id:   &id,
		Type: kbapi.LensDashboardApp,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKbnDashboardPanelTypeLensDashboardApp(panel))
	return item
}

func panelModelWithGrid() models.PanelModel {
	return models.PanelModel{
		Grid: models.PanelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
			W: types.Int64Value(24),
			H: types.Int64Value(12),
		},
		ID: types.StringValue("panel-id"),
	}
}

func TestTerraformChartBlockKey_coversAllConverters(t *testing.T) {
	t.Parallel()
	for _, c := range lenscommon.All() {
		key := lenscommon.TerraformChartBlockKey(c.VizType())
		require.NotEmpty(t, key, "missing terraform attribute key for viz type %q", c.VizType())
	}
}

func TestHandler_FromAPI_byValue_metric(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	const inner = `{
		"type": "metric",
		"title": "M",
		"query": { "expression": "*", "language": "kql" },
		"metrics": []
	}`
	var cfg0 kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, json.Unmarshal([]byte(inner), &cfg0))
	item := mustLensDashboardAppPanelItem(t, cfg0)

	var pm models.PanelModel
	diags := lensdashboardapp.Handler{}.FromAPI(ctx, &pm, nil, item)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, pm.LensDashboardAppConfig)
	require.NotNil(t, pm.LensDashboardAppConfig.ByValue)
	require.True(t, typeutils.IsKnown(pm.LensDashboardAppConfig.ByValue.ConfigJSON))
	var root map[string]any
	require.NoError(t, json.Unmarshal([]byte(pm.LensDashboardAppConfig.ByValue.ConfigJSON.ValueString()), &root))
	assert.Equal(t, "metric", root["type"])
}

func TestHandler_FromAPI_byValue_datatable(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	apiJSON := `{"type":"data_table","title":"d","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 5"},` +
		`"filters":[],"metrics":[{"column":"c","operation":"value","format":{"type":"number"}}],` +
		`"rows":[{"column":"r","collapse_by":"avg","format":{"type":"number"}}],` +
		`"styling":{"density":{"mode":"default","height":{"header":{"type":"auto"},"value":{"type":"auto"}}}},` +
		`"time_range":{"from":"now-7d","to":"now"}}`
	var api kbapi.DatatableESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
	var cfg0 kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, cfg0.FromDatatableESQL(api))
	item := mustLensDashboardAppPanelItem(t, cfg0)

	var pm models.PanelModel
	diags := lensdashboardapp.Handler{}.FromAPI(ctx, &pm, nil, item)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, pm.LensDashboardAppConfig.ByValue)
	require.True(t, typeutils.IsKnown(pm.LensDashboardAppConfig.ByValue.ConfigJSON))
	var root map[string]any
	require.NoError(t, json.Unmarshal([]byte(pm.LensDashboardAppConfig.ByValue.ConfigJSON.ValueString()), &root))
	assert.Equal(t, "data_table", root["type"])
}

func TestHandler_roundTrip_byReference(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	const apiPanelsJSON = `{
		"type": "lens-dashboard-app",
		"grid": { "x": 0, "y": 0, "w": 24, "h": 12 },
		"id": "lens-ref-panel",
		"config": {
			"ref_id": "lens:a1b2c3",
			"time_range": { "from": "now-7d", "to": "now" },
			"title": "Linked lens"
		}
	}`
	var ld kbapi.KbnDashboardPanelTypeLensDashboardApp
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &ld))
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKbnDashboardPanelTypeLensDashboardApp(ld))

	var pm models.PanelModel
	diags := lensdashboardapp.Handler{}.FromAPI(ctx, &pm, nil, item)
	require.False(t, diags.HasError(), "%s", diags)

	out, d2 := lensdashboardapp.Handler{}.ToAPI(pm, nil)
	require.False(t, d2.HasError(), "%s", d2)
	back, err := out.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	cfg1, err := back.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig1()
	require.NoError(t, err)
	assert.Equal(t, "lens:a1b2c3", cfg1.RefId)
	require.NotNil(t, cfg1.Title)
	assert.Equal(t, "Linked lens", *cfg1.Title)
}

func TestHandler_ToAPI_byValue_xy_roundTripGrid(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	const inner = `{
		"type": "xy",
		"title": "XY",
		"query": { "expression": "*", "language": "kql" },
		"layers": [{"type": "line", "y": [{"operation": "count", "axis": "left"}]}]
	}`
	var cfg0 kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, json.Unmarshal([]byte(inner), &cfg0))
	item := mustLensDashboardAppPanelItem(t, cfg0)

	var pm models.PanelModel
	require.False(t, lensdashboardapp.Handler{}.FromAPI(ctx, &pm, nil, item).HasError())

	out, d2 := lensdashboardapp.Handler{}.ToAPI(pm, nil)
	require.False(t, d2.HasError())
	ld, err := out.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	require.NotNil(t, ld.Grid.W)
	assert.InDelta(t, float64(24), float64(*ld.Grid.W), 1e-6)
}

func TestHandler_ToAPI_FromAPI_byValue_configJSONOnly(t *testing.T) {
	dash := &models.DashboardModel{}
	ctx := iface.WithEnclosingDashboard(context.Background(), dash)
	raw := `{"type":"metric","title":"M","query":{"expression":"*","language":"kql"},"metrics":[]}`
	pm := panelModelWithGrid()
	pm.LensDashboardAppConfig = &models.LensDashboardAppConfigModel{
		ByValue: &models.LensDashboardAppByValueModel{
			ConfigJSON: jsontypes.NewNormalizedValue(raw),
		},
	}

	out, d := lensdashboardapp.Handler{}.ToAPI(pm, dash)
	require.False(t, d.HasError(), "%s", d)

	var pm2 models.PanelModel
	diags := lensdashboardapp.Handler{}.FromAPI(ctx, &pm2, nil, out)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, pm2.LensDashboardAppConfig)
	require.NotNil(t, pm2.LensDashboardAppConfig.ByValue)
	require.True(t, typeutils.IsKnown(pm2.LensDashboardAppConfig.ByValue.ConfigJSON))
	assert.Contains(t, pm2.LensDashboardAppConfig.ByValue.ConfigJSON.ValueString(), `"type":"metric"`)
}
