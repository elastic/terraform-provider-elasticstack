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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mapPanelFromAPI_vis_byReference_populatesVizConfig(t *testing.T) {
	ctx := context.Background()
	const apiPanelsJSON = `[
		{
			"type": "vis",
			"grid": { "x": 0, "y": 0, "w": 24, "h": 12 },
			"id": "viz-ref-panel",
			"config": {
				"ref_id": "lens:a1b2c3",
				"time_range": { "from": "now-7d", "to": "now" },
				"title": "Linked lens"
			}
		}
	]`
	var apiPanels kbapi.DashboardPanels
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &apiPanels))

	dm := &dashboardModel{}
	panels, sections, diags := dm.mapPanelsFromAPI(ctx, &apiPanels)
	require.False(t, diags.HasError())
	require.Nil(t, sections)
	require.Len(t, panels, 1)
	pm := panels[0]

	require.NotNil(t, pm.VizConfig)
	require.Nil(t, pm.VizConfig.ByValue)
	require.NotNil(t, pm.VizConfig.ByReference)
	assert.Equal(t, "lens:a1b2c3", pm.VizConfig.ByReference.RefID.ValueString())
	assert.Equal(t, "now-7d", pm.VizConfig.ByReference.TimeRange.From.ValueString())
	assert.Equal(t, "now", pm.VizConfig.ByReference.TimeRange.To.ValueString())
	assert.Equal(t, "Linked lens", pm.VizConfig.ByReference.Title.ValueString())
	require.True(t, typeutils.IsKnown(pm.ConfigJSON))
}

func Test_mapPanelFromAPI_vis_byValue_populatesNestedChartBlock(t *testing.T) {
	ctx := context.Background()
	const apiPanelsJSON = `[
		{
			"type": "vis",
			"grid": { "x": 0, "y": 0, "w": 6, "h": 6 },
			"id": "viz-metric",
			"config": {
				"type": "metric",
				"title": "M",
				"query": { "expression": "*", "language": "kql" },
				"metrics": []
			}
		}
	]`
	var apiPanels kbapi.DashboardPanels
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &apiPanels))

	dm := &dashboardModel{}
	panels, _, diags := dm.mapPanelsFromAPI(ctx, &apiPanels)
	require.False(t, diags.HasError())
	require.Len(t, panels, 1)

	pm := panels[0]
	require.NotNil(t, pm.VizConfig)
	require.Nil(t, pm.VizConfig.ByReference)
	require.NotNil(t, pm.VizConfig.ByValue)
	require.NotNil(t, pm.VizConfig.ByValue.MetricChartConfig)
}

func Test_mapPanelFromAPI_vis_byValue_prefersAPIChartOverStalePriorXYBlock(t *testing.T) {
	ctx := context.Background()

	tfPanel := panelModel{
		Type: types.StringValue("vis"),
		VizConfig: &vizConfigModel{
			ByValue: &vizByValueModel{
				lensByValueChartBlocks: lensByValueChartBlocks{
					XYChartConfig: &xyChartConfigModel{
						Title: types.StringValue("Old XY Title"),
						Axis: &xyAxisModel{
							X: &xyAxisConfigModel{},
							Y: &yAxisConfigModel{},
						},
						Decorations: &xyDecorationsModel{},
						Fitting:     &xyFittingModel{Type: types.StringValue("none")},
						Legend:      &xyLegendModel{Inside: types.BoolValue(false), Visibility: types.StringValue("visible")},
						Query:       &filterSimpleModel{Language: types.StringValue("kql"), Expression: types.StringValue("*")},
					},
				},
			},
		},
		ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
	}

	const apiPanelsJSON = `[
		{
			"type": "vis",
			"grid": { "x": 0, "y": 0, "w": 6, "h": 6 },
			"id": "viz-chart-swap",
			"config": {
				"type": "metric",
				"title": "Metric From API",
				"query": { "expression": "*", "language": "kql" },
				"metrics": []
			}
		}
	]`
	var apiPanels kbapi.DashboardPanels
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &apiPanels))
	item := apiPanels[0]
	panelRow, err := item.AsDashboardPanelItem()
	require.NoError(t, err)

	dm := dashboardModel{}
	out, diags := dm.mapPanelFromAPI(ctx, &tfPanel, panelRow)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, out.VizConfig)
	require.NotNil(t, out.VizConfig.ByValue)
	require.Nil(t, out.VizConfig.ByValue.XYChartConfig)
	require.NotNil(t, out.VizConfig.ByValue.MetricChartConfig)
	assert.Equal(t, "Metric From API", out.VizConfig.ByValue.MetricChartConfig.Title.ValueString())
}

func Test_mapPanelFromAPI_vis_unsupportedChartDiagnostic(t *testing.T) {
	ctx := context.Background()

	original := lensVizConverters
	lensVizConverters = nil // no converters match metric (or anything)
	t.Cleanup(func() {
		lensVizConverters = original
	})

	const apiPanelsJSON = `[
		{
			"type": "vis",
			"grid": { "x": 0, "y": 0, "w": 4, "h": 4 },
			"config": {
				"type": "metric",
				"title": "M",
				"query": { "expression": "*", "language": "kql" },
				"metrics": []
			}
		}
	]`
	var apiPanels kbapi.DashboardPanels
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &apiPanels))

	dm := &dashboardModel{}
	_, _, diags := dm.mapPanelsFromAPI(ctx, &apiPanels)
	require.True(t, diags.HasError())
	found := false
	for _, d := range diags {
		if d.Summary() == "Unsupported visualization chart type" {
			found = true
			assert.Contains(t, d.Detail(), "metric")
			assert.Contains(t, d.Detail(), "config_json")
			break
		}
	}
	require.True(t, found, "expected Unsupported visualization chart type diagnostic")
}

func Test_mapPanelFromAPI_vis_ambiguousPreservesPriorByReference(t *testing.T) {
	ctx := context.Background()
	priorPanel := panelModel{
		Type: types.StringValue("vis"),
		VizConfig: &vizConfigModel{
			ByReference: &vizByReferenceModel{
				RefID: types.StringValue("saved/prior/ref"),
				TimeRange: lensDashboardAppTimeRangeModel{
					From: types.StringValue("now-30d"),
					To:   types.StringValue("now"),
				},
				Title: types.StringValue("Prior Title"),
			},
		},
	}

	const apiPanelsJSON = `[
		{
			"type": "vis",
			"grid": { "x": 0, "y": 0, "w": 12, "h": 12 },
			"config": { "_note": "no chart type root and incomplete ref linkage" }
		}
	]`
	var apiPanels kbapi.DashboardPanels
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &apiPanels))
	item := apiPanels[0]
	panelRow, err := item.AsDashboardPanelItem()
	require.NoError(t, err)

	dm := dashboardModel{}
	out, diags := dm.mapPanelFromAPI(ctx, &priorPanel, panelRow)
	require.False(t, diags.HasError())
	require.NotNil(t, out.VizConfig)
	require.Nil(t, out.VizConfig.ByValue)
	require.NotNil(t, out.VizConfig.ByReference)
	assert.Equal(t, priorPanel.VizConfig.ByReference.RefID.ValueString(), out.VizConfig.ByReference.RefID.ValueString())
	assert.Equal(t, priorPanel.VizConfig.ByReference.Title.ValueString(), out.VizConfig.ByReference.Title.ValueString())
}

func Test_mapPanelFromAPI_vis_configJSONOnlyLeavesVizUnset(t *testing.T) {
	ctx := context.Background()
	const apiPanelsJSON = `[
		{
			"type": "vis",
			"grid": { "x": 0, "y": 0, "w": 8, "h": 8 },
			"config": { "wrapped": {} }
		}
	]`
	var apiPanels kbapi.DashboardPanels
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &apiPanels))

	raw := `{"wrapped": {}}`
	tfPrior := panelModel{
		ConfigJSON: customtypes.NewJSONWithDefaultsValue(raw, populatePanelConfigJSONDefaults),
		VizConfig:  nil,
	}

	item := apiPanels[0]
	panelRow, err := item.AsDashboardPanelItem()
	require.NoError(t, err)

	dm := dashboardModel{}
	out, diags := dm.mapPanelFromAPI(ctx, &tfPrior, panelRow)
	require.False(t, diags.HasError())
	assert.Nil(t, out.VizConfig)
}

func Test_panel_toAPI_vis_byReference_writesVisConfig1(t *testing.T) {
	pm := panelModel{
		Type: types.StringValue("vis"),
		Grid: panelGridModel{X: types.Int64Value(1), Y: types.Int64Value(2), W: types.Int64Value(24), H: types.Int64Value(14)},
		ID:   types.StringValue("p-ref"),
		VizConfig: &vizConfigModel{
			ByReference: &vizByReferenceModel{
				RefID: types.StringValue("lens:out"),
				TimeRange: lensDashboardAppTimeRangeModel{
					From: types.StringValue("now-1h"),
					To:   types.StringValue("now"),
				},
			},
		},
		ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
	}

	item, diags := pm.toAPI(nil)
	require.False(t, diags.HasError())

	visPanel, err := item.AsKbnDashboardPanelTypeVis()
	require.NoError(t, err)
	cfg1, err := visPanel.Config.AsKbnDashboardPanelTypeVisConfig1()
	require.NoError(t, err)
	assert.Equal(t, "lens:out", cfg1.RefId)
	assert.Equal(t, "now-1h", cfg1.TimeRange.From)
	assert.Equal(t, "now", cfg1.TimeRange.To)
}

func Test_panel_toAPI_vis_configJSONWithoutViz_unmarshalsOpaqueConfigJSON(t *testing.T) {
	pm := panelModel{
		Type: types.StringValue("vis"),
		Grid: panelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(10), H: types.Int64Value(10)},
		ConfigJSON: customtypes.NewJSONWithDefaultsValue(
			`{"attributes":{"references":[]}}`,
			populatePanelConfigJSONDefaults,
		),
	}
	item, diags := pm.toAPI(nil)
	require.False(t, diags.HasError())
	v, err := item.AsKbnDashboardPanelTypeVis()
	require.NoError(t, err)
	raw, err := v.Config.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, string(raw), "attributes")
}
