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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mapPanelFromAPI_vis_byReference_populatesVisConfig(t *testing.T) {
	ctx := context.Background()
	const apiPanelsJSON = `[
		{
			"type": "vis",
			"grid": { "x": 0, "y": 0, "w": 24, "h": 12 },
			"id": "vis-ref-panel",
			"config": {
				"ref_id": "lens:a1b2c3",
				"time_range": { "from": "now-7d", "to": "now" },
				"title": "Linked lens"
			}
		}
	]`
	var apiPanels kbapi.DashboardPanels
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &apiPanels))

	dm := &models.DashboardModel{}
	panels, sections, diags := dashboardMapPanelsFromAPI(ctx, dm, &apiPanels)
	require.False(t, diags.HasError())
	require.Nil(t, sections)
	require.Len(t, panels, 1)
	pm := panels[0]

	require.NotNil(t, pm.VisConfig)
	require.Nil(t, pm.VisConfig.ByValue)
	require.NotNil(t, pm.VisConfig.ByReference)
	assert.Equal(t, "lens:a1b2c3", pm.VisConfig.ByReference.RefID.ValueString())
	assert.Equal(t, "now-7d", pm.VisConfig.ByReference.TimeRange.From.ValueString())
	assert.Equal(t, "now", pm.VisConfig.ByReference.TimeRange.To.ValueString())
	assert.Equal(t, "Linked lens", pm.VisConfig.ByReference.Title.ValueString())
	require.True(t, typeutils.IsKnown(pm.ConfigJSON))
}

func Test_mapPanelFromAPI_vis_byValue_populatesNestedChartBlock(t *testing.T) {
	ctx := context.Background()
	const apiPanelsJSON = `[
		{
			"type": "vis",
			"grid": { "x": 0, "y": 0, "w": 6, "h": 6 },
			"id": "vis-metric",
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

	dm := &models.DashboardModel{}
	panels, _, diags := dashboardMapPanelsFromAPI(ctx, dm, &apiPanels)
	require.False(t, diags.HasError())
	require.Len(t, panels, 1)

	pm := panels[0]
	require.NotNil(t, pm.VisConfig)
	require.Nil(t, pm.VisConfig.ByReference)
	require.NotNil(t, pm.VisConfig.ByValue)
	require.NotNil(t, pm.VisConfig.ByValue.MetricChartConfig)
}

func Test_mapPanelFromAPI_vis_byValue_prefersAPIChartOverStalePriorXYBlock(t *testing.T) {
	ctx := context.Background()

	tfPanel := models.PanelModel{
		Type: types.StringValue("vis"),
		VisConfig: &models.VisConfigModel{
			ByValue: &models.VisByValueModel{
				LensByValueChartBlocks: models.LensByValueChartBlocks{
					XYChartConfig: &models.XYChartConfigModel{
						Title: types.StringValue("Old XY Title"),
						Axis: &models.XYAxisModel{
							X: &models.XYAxisConfigModel{},
							Y: &models.YAxisConfigModel{},
						},
						Decorations: &models.XYDecorationsModel{},
						Fitting:     &models.XYFittingModel{Type: types.StringValue("none")},
						Legend:      &models.XYLegendModel{Inside: types.BoolValue(false), Visibility: types.StringValue("visible")},
						Query:       &models.FilterSimpleModel{Language: types.StringValue("kql"), Expression: types.StringValue("*")},
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
			"id": "vis-chart-swap",
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

	dm := models.DashboardModel{}
	out, diags := dashboardMapPanelFromAPI(ctx, &dm, &tfPanel, panelRow)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, out.VisConfig)
	require.NotNil(t, out.VisConfig.ByValue)
	require.Nil(t, out.VisConfig.ByValue.XYChartConfig)
	require.NotNil(t, out.VisConfig.ByValue.MetricChartConfig)
	assert.Equal(t, "Metric From API", out.VisConfig.ByValue.MetricChartConfig.Title.ValueString())
}

func Test_mapPanelFromAPI_vis_unsupportedChartDiagnostic(t *testing.T) {
	ctx := context.Background()

	original := lensVisConverters
	lensVisConverters = nil // no converters match metric (or anything)
	t.Cleanup(func() {
		lensVisConverters = original
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

	dm := &models.DashboardModel{}
	_, _, diags := dashboardMapPanelsFromAPI(ctx, dm, &apiPanels)
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
	priorPanel := models.PanelModel{
		Type: types.StringValue("vis"),
		VisConfig: &models.VisConfigModel{
			ByReference: &models.VisByReferenceModel{
				RefID: types.StringValue("saved/prior/ref"),
				TimeRange: models.LensDashboardAppTimeRangeModel{
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

	dm := models.DashboardModel{}
	out, diags := dashboardMapPanelFromAPI(ctx, &dm, &priorPanel, panelRow)
	require.False(t, diags.HasError())
	require.NotNil(t, out.VisConfig)
	require.Nil(t, out.VisConfig.ByValue)
	require.NotNil(t, out.VisConfig.ByReference)
	assert.Equal(t, priorPanel.VisConfig.ByReference.RefID.ValueString(), out.VisConfig.ByReference.RefID.ValueString())
	assert.Equal(t, priorPanel.VisConfig.ByReference.Title.ValueString(), out.VisConfig.ByReference.Title.ValueString())
}

func Test_mapPanelFromAPI_vis_configJSONOnlyLeavesVisUnset(t *testing.T) {
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
	tfPrior := models.PanelModel{
		ConfigJSON: customtypes.NewJSONWithDefaultsValue(raw, populatePanelConfigJSONDefaults),
		VisConfig:  nil,
	}

	item := apiPanels[0]
	panelRow, err := item.AsDashboardPanelItem()
	require.NoError(t, err)

	dm := models.DashboardModel{}
	out, diags := dashboardMapPanelFromAPI(ctx, &dm, &tfPrior, panelRow)
	require.False(t, diags.HasError())
	assert.Nil(t, out.VisConfig)
}

func Test_panel_toAPI_vis_byReference_writesVisConfig1(t *testing.T) {
	pm := models.PanelModel{
		Type: types.StringValue("vis"),
		Grid: models.PanelGridModel{X: types.Int64Value(1), Y: types.Int64Value(2), W: types.Int64Value(24), H: types.Int64Value(14)},
		ID:   types.StringValue("p-ref"),
		VisConfig: &models.VisConfigModel{
			ByReference: &models.VisByReferenceModel{
				RefID: types.StringValue("lens:out"),
				TimeRange: models.LensDashboardAppTimeRangeModel{
					From: types.StringValue("now-1h"),
					To:   types.StringValue("now"),
				},
			},
		},
		ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
	}

	item, diags := panelToAPI(context.Background(), pm, nil)
	require.False(t, diags.HasError())

	visPanel, err := item.AsKbnDashboardPanelTypeVis()
	require.NoError(t, err)
	cfg1, err := visPanel.Config.AsKbnDashboardPanelTypeVisConfig1()
	require.NoError(t, err)
	assert.Equal(t, "lens:out", cfg1.RefId)
	assert.Equal(t, "now-1h", cfg1.TimeRange.From)
	assert.Equal(t, "now", cfg1.TimeRange.To)
}

func Test_panel_toAPI_vis_configJSONWithoutVis_unmarshalsOpaqueConfigJSON(t *testing.T) {
	pm := models.PanelModel{
		Type: types.StringValue("vis"),
		Grid: models.PanelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(10), H: types.Int64Value(10)},
		ConfigJSON: customtypes.NewJSONWithDefaultsValue(
			`{"attributes":{"references":[]}}`,
			populatePanelConfigJSONDefaults,
		),
	}
	item, diags := panelToAPI(context.Background(), pm, nil)
	require.False(t, diags.HasError())
	v, err := item.AsKbnDashboardPanelTypeVis()
	require.NoError(t, err)
	raw, err := v.Config.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, string(raw), "attributes")
}

func Test_visConfigToAPI_missingVisConfig_diagnostic(t *testing.T) {
	pm := models.PanelModel{
		Type:       types.StringValue("vis"),
		Grid:       models.PanelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(10), H: types.Int64Value(10)},
		VisConfig:  nil,
		ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
	}
	_, diags := visConfigToAPI(pm, nil, struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	}{}, nil)
	require.True(t, diags.HasError())
	found := false
	for _, d := range diags {
		if d.Summary() == "Missing `vis_config`" {
			found = true
			break
		}
	}
	require.True(t, found, "expected Missing vis_config diagnostic")
}

func Test_visConfigToAPI_byValue_missingConverter_diagnostic(t *testing.T) {
	pm := models.PanelModel{
		Type: types.StringValue("vis"),
		Grid: models.PanelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0), W: types.Int64Value(10), H: types.Int64Value(10)},
		VisConfig: &models.VisConfigModel{
			ByValue: &models.VisByValueModel{
				LensByValueChartBlocks: models.LensByValueChartBlocks{},
			},
		},
		ConfigJSON: customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults),
	}
	_, diags := visConfigToAPI(pm, nil, struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	}{}, nil)
	require.True(t, diags.HasError())
	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid `vis_config.by_value`" {
			found = true
			break
		}
	}
	require.True(t, found, "expected Invalid vis_config.by_value diagnostic")
}

func Test_visByReferenceToAPI_invalidReferencesJSON_diagnostic(t *testing.T) {
	byRef := models.LensDashboardAppByReferenceModel{
		RefID: types.StringValue("lens:out"),
		TimeRange: models.LensDashboardAppTimeRangeModel{
			From: types.StringValue("now-1h"),
			To:   types.StringValue("now"),
		},
		ReferencesJSON: jsontypes.NewNormalizedValue(`not-valid-json`),
	}
	_, diags := visByReferenceToAPI(byRef, struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	}{X: 0, Y: 0}, nil)
	require.True(t, diags.HasError())
	found := false
	for _, d := range diags {
		if d.Summary() == "Invalid `vis_config.by_reference.references_json`" {
			found = true
			break
		}
	}
	require.True(t, found, "expected Invalid references_json diagnostic")
}

func Test_populateVisByReferenceFromAPI_emptyDrilldownsSlice(t *testing.T) {
	ctx := context.Background()
	cfg1 := kbapi.KbnDashboardPanelTypeVisConfig1{
		RefId: "lens:ref",
		TimeRange: kbapi.KbnEsQueryServerTimeRangeSchema{
			From: "now-7d",
			To:   "now",
		},
		Drilldowns: &[]kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item{},
	}
	pm := &models.PanelModel{}
	diags := populateVisByReferenceFromAPI(ctx, nil, pm, cfg1)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.VisConfig)
	require.NotNil(t, pm.VisConfig.ByReference)
	// API returns empty slice → populated as empty models.DrilldownsModel (not nil)
	assert.NotNil(t, pm.VisConfig.ByReference.Drilldowns)
	assert.Empty(t, pm.VisConfig.ByReference.Drilldowns)
}

func Test_populateVisByReferenceFromAPI_nilDrilldownsFallsBackToPrior(t *testing.T) {
	ctx := context.Background()
	cfg1 := kbapi.KbnDashboardPanelTypeVisConfig1{
		RefId: "lens:ref",
		TimeRange: kbapi.KbnEsQueryServerTimeRangeSchema{
			From: "now-7d",
			To:   "now",
		},
		// Drilldowns intentionally nil
	}
	prior := &models.VisConfigModel{
		ByReference: &models.LensDashboardAppByReferenceModel{
			RefID: types.StringValue("prior"),
			TimeRange: models.LensDashboardAppTimeRangeModel{
				From: types.StringValue("now-30d"),
				To:   types.StringValue("now"),
			},
		},
	}
	pm := &models.PanelModel{}
	diags := populateVisByReferenceFromAPI(ctx, prior, pm, cfg1)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.VisConfig)
	require.NotNil(t, pm.VisConfig.ByReference)
	assert.Equal(t, "lens:ref", pm.VisConfig.ByReference.RefID.ValueString())
}
