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

package visconfig_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard" // register lens converters + JSON defaults
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/visconfig"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustVisPanelItem(t *testing.T, cfg0 lenscommon.VisByValueConfig0) kbapi.DashboardPanelItem {
	t.Helper()
	var cfg kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config
	require.NoError(t, cfg.FromKibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0(cfg0))
	w, h := float32(24), float32(12)
	id := "panel-id"
	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis{
		Config: cfg,
		Grid: struct {
			H *float32 `json:"h,omitempty"`
			W *float32 `json:"w,omitempty"`
			X float32  `json:"x"`
			Y float32  `json:"y"`
		}{X: 0, Y: 0, W: new(w), H: new(h)},
		Id:   &id,
		Type: kbapi.Vis,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeVis(panel))
	return item
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
	var cfg0 lenscommon.VisByValueConfig0
	require.NoError(t, json.Unmarshal([]byte(inner), &cfg0))
	item := mustVisPanelItem(t, cfg0)

	var pm models.PanelModel
	diags := visconfig.Handler{}.FromAPI(ctx, &pm, nil, item)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, pm.VisConfig)
	require.NotNil(t, pm.VisConfig.ByValue)
	require.NotNil(t, pm.VisConfig.ByValue.MetricChartConfig)
}

func TestHandler_FromAPI_byValue_datatable(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	apiJSON := `{"type":"data_table","title":"d","data_source":{"type":"esql","query":"FROM logs-* | LIMIT 5"},` +
		`"filters":[],"metrics":[{"column":"c","operation":"value","format":{"type":"number"}}],` +
		`"rows":[{"column":"r","collapse_by":"avg","format":{"type":"number"}}],` +
		`"styling":{"density":{"mode":"default","height":{"header":{"type":"auto"},"value":{"type":"auto"}}}},` +
		`"time_range":{"from":"now-7d","to":"now"}}`
	var api kbapi.KibanaHTTPAPIsDatatableESQLByValuePanel
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
	var cfg0 lenscommon.VisByValueConfig0
	require.NoError(t, cfg0.FromKibanaHTTPAPIsDatatableESQLByValuePanel(api))
	item := mustVisPanelItem(t, cfg0)

	var pm models.PanelModel
	diags := visconfig.Handler{}.FromAPI(ctx, &pm, nil, item)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, pm.VisConfig.ByValue.DatatableConfig)
}

func TestHandler_FromAPI_configJSONOnlyPreservesUnsetVisConfig(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	const apiPanelsJSON = `{
		"type": "vis",
		"grid": { "x": 0, "y": 0, "w": 8, "h": 8 },
		"config": { "wrapped": {} }
	}`
	var vis kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &vis))
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeVis(vis))

	raw := `{"wrapped": {}}`
	tfPrior := models.PanelModel{
		ConfigJSON: customtypes.NewJSONWithDefaultsValue(raw, panelkit.PanelJSONDefaultsFunc()),
		VisConfig:  nil,
	}

	var pm models.PanelModel
	diags := visconfig.Handler{}.FromAPI(ctx, &pm, &tfPrior, item)
	require.False(t, diags.HasError(), "%s", diags)
	assert.Nil(t, pm.VisConfig)
}

func TestHandler_roundTrip_byReference_withoutTimeRange(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	const apiPanelsJSON = `{
		"type": "vis",
		"grid": { "x": 0, "y": 0, "w": 24, "h": 12 },
		"id": "vis-ref-panel",
		"config": {
			"ref_id": "lens:a1b2c3",
			"title": "Linked lens"
		}
	}`
	var vis kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &vis))
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeVis(vis))

	var pm models.PanelModel
	diags := visconfig.Handler{}.FromAPI(ctx, &pm, nil, item)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, pm.VisConfig.ByReference)
	assert.Nil(t, pm.VisConfig.ByReference.TimeRange)

	out, d2 := visconfig.Handler{}.ToAPI(pm, nil)
	require.False(t, d2.HasError(), "%s", d2)
	back, err := out.AsKibanaHTTPAPIsKbnDashboardPanelTypeVis()
	require.NoError(t, err)
	cfg1, err := back.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1()
	require.NoError(t, err)
	assert.Equal(t, "lens:a1b2c3", cfg1.RefId)
	assert.Nil(t, cfg1.TimeRange)
}

func TestVisByReferenceModelToAPIConfig1_omitsTimeRangeWhenUnset(t *testing.T) {
	byRef := models.VisByReferenceModel{
		RefID: types.StringValue("panel_0"),
	}
	api1, diags := lenscommon.VisByReferenceModelToAPIConfig1(byRef, "references_json")
	require.False(t, diags.HasError())
	assert.Nil(t, api1.TimeRange)
}

func TestHandler_roundTrip_byReference(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	const apiPanelsJSON = `{
		"type": "vis",
		"grid": { "x": 0, "y": 0, "w": 24, "h": 12 },
		"id": "vis-ref-panel",
		"config": {
			"ref_id": "lens:a1b2c3",
			"time_range": { "from": "now-7d", "to": "now" },
			"title": "Linked lens"
		}
	}`
	var vis kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &vis))
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeVis(vis))

	var pm models.PanelModel
	diags := visconfig.Handler{}.FromAPI(ctx, &pm, nil, item)
	require.False(t, diags.HasError(), "%s", diags)

	out, d2 := visconfig.Handler{}.ToAPI(pm, nil)
	require.False(t, d2.HasError(), "%s", d2)
	back, err := out.AsKibanaHTTPAPIsKbnDashboardPanelTypeVis()
	require.NoError(t, err)
	cfg1, err := back.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1()
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
	var cfg0 lenscommon.VisByValueConfig0
	require.NoError(t, json.Unmarshal([]byte(inner), &cfg0))
	item := mustVisPanelItem(t, cfg0)

	var pm models.PanelModel
	require.False(t, visconfig.Handler{}.FromAPI(ctx, &pm, nil, item).HasError())

	out, d2 := visconfig.Handler{}.ToAPI(pm, nil)
	require.False(t, d2.HasError())
	v, err := out.AsKibanaHTTPAPIsKbnDashboardPanelTypeVis()
	require.NoError(t, err)
	require.NotNil(t, v.Grid.W)
	assert.InDelta(t, float64(24), float64(*v.Grid.W), 1e-6)
}
