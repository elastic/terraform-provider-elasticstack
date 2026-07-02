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

package mlanomalycharts_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/mlanomalycharts"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestContract(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, mlanomalycharts.Handler{}, contracttest.Config{
		FullAPIResponse: `{
			"type": "ml_anomaly_charts",
			"grid": {"x": 0, "y": 0, "w": 24, "h": 8},
			"id": "ml-charts-contract",
			"config": {
				"job_ids": ["job-a"],
				"severity_threshold": [{"min": 75}, {"min": 50, "max": 75}],
				"title": "Anomaly Charts"
			}
		}`,
		OmitRequiredLeafPresence: true,
		OmitValidateRequiredZero: true,
		SkipFields:               []string{"config.time_range", "time_range", "severity_threshold"},
	})
}

func TestNamedSeverity_roundTrip(t *testing.T) {
	t.Parallel()

	pm := panelWithConfig(&models.MlAnomalyChartsConfigModel{
		JobIDs: []types.String{types.StringValue("job-a")},
		SeverityThreshold: []models.MlAnomalyChartsSeverityThresholdModel{
			{Severity: types.StringValue("major")},
		},
	})

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{}
	diags := mlanomalycharts.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%s", diags)
	require.Equal(t, []map[string]any{{"min": float64(50), "max": float64(75)}}, severityThresholdMaps(t, panel.Config.SeverityThreshold))

	readBack := readPanelViaHandlerWithPrior(t, panel, &pm)
	require.Equal(t, "major", readBack.MlAnomalyChartsConfig.SeverityThreshold[0].Severity.ValueString())
	require.True(t, readBack.MlAnomalyChartsConfig.SeverityThreshold[0].Min.IsNull())
	require.True(t, readBack.MlAnomalyChartsConfig.SeverityThreshold[0].Max.IsNull())
}

func TestCriticalSeverity_roundTrip(t *testing.T) {
	t.Parallel()

	pm := panelWithConfig(&models.MlAnomalyChartsConfigModel{
		JobIDs: []types.String{types.StringValue("job-a")},
		SeverityThreshold: []models.MlAnomalyChartsSeverityThresholdModel{
			{Severity: types.StringValue("critical")},
		},
	})

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{}
	diags := mlanomalycharts.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%s", diags)
	require.Equal(t, []map[string]any{{"min": float64(75)}}, severityThresholdMaps(t, panel.Config.SeverityThreshold))

	readBack := readPanelViaHandlerWithPrior(t, panel, &pm)
	require.Equal(t, "critical", readBack.MlAnomalyChartsConfig.SeverityThreshold[0].Severity.ValueString())
}

func TestRawRange_roundTrip(t *testing.T) {
	t.Parallel()

	pm := panelWithConfig(&models.MlAnomalyChartsConfigModel{
		JobIDs: []types.String{types.StringValue("job-a")},
		SeverityThreshold: []models.MlAnomalyChartsSeverityThresholdModel{
			{Min: types.Int64Value(10), Max: types.Int64Value(20)},
		},
	})

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{}
	diags := mlanomalycharts.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%s", diags)
	require.Equal(t, []map[string]any{{"min": float64(10), "max": float64(20)}}, severityThresholdMaps(t, panel.Config.SeverityThreshold))

	readBack := readPanelViaHandlerWithPrior(t, panel, &pm)
	item := readBack.MlAnomalyChartsConfig.SeverityThreshold[0]
	require.True(t, item.Severity.IsNull())
	require.Equal(t, int64(10), item.Min.ValueInt64())
	require.Equal(t, int64(20), item.Max.ValueInt64())
}

func TestFormPreservation_priorRawCoincidingWithWarning(t *testing.T) {
	t.Parallel()

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{
		Type: kbapi.MlAnomalyCharts,
		Config: kbapi.KibanaHTTPAPIsMlAnomalyCharts{
			JobIds:            []string{"job-a"},
			SeverityThreshold: severityThresholdItems(t, []map[string]any{{"min": 3, "max": 25}}),
		},
	}
	prior := panelWithConfig(&models.MlAnomalyChartsConfigModel{
		JobIDs: []types.String{types.StringValue("job-a")},
		SeverityThreshold: []models.MlAnomalyChartsSeverityThresholdModel{
			{Min: types.Int64Value(3), Max: types.Int64Value(25)},
		},
	})

	readBack := readPanelViaHandlerWithPrior(t, panel, &prior)
	item := readBack.MlAnomalyChartsConfig.SeverityThreshold[0]
	require.True(t, item.Severity.IsNull())
	require.Equal(t, int64(3), item.Min.ValueInt64())
	require.Equal(t, int64(25), item.Max.ValueInt64())
}

func TestFormPreservation_priorNamedWarning(t *testing.T) {
	t.Parallel()

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{
		Type: kbapi.MlAnomalyCharts,
		Config: kbapi.KibanaHTTPAPIsMlAnomalyCharts{
			JobIds:            []string{"job-a"},
			SeverityThreshold: severityThresholdItems(t, []map[string]any{{"min": 3, "max": 25}}),
		},
	}
	prior := panelWithConfig(&models.MlAnomalyChartsConfigModel{
		JobIDs: []types.String{types.StringValue("job-a")},
		SeverityThreshold: []models.MlAnomalyChartsSeverityThresholdModel{
			{Severity: types.StringValue("warning")},
		},
	})

	readBack := readPanelViaHandlerWithPrior(t, panel, &prior)
	item := readBack.MlAnomalyChartsConfig.SeverityThreshold[0]
	require.Equal(t, "warning", item.Severity.ValueString())
	require.True(t, item.Min.IsNull())
	require.True(t, item.Max.IsNull())
}

func TestFormPreservation_criticalRaw(t *testing.T) {
	t.Parallel()

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{
		Type: kbapi.MlAnomalyCharts,
		Config: kbapi.KibanaHTTPAPIsMlAnomalyCharts{
			JobIds:            []string{"job-a"},
			SeverityThreshold: severityThresholdItems(t, []map[string]any{{"min": 75}}),
		},
	}
	prior := panelWithConfig(&models.MlAnomalyChartsConfigModel{
		JobIDs: []types.String{types.StringValue("job-a")},
		SeverityThreshold: []models.MlAnomalyChartsSeverityThresholdModel{
			{Min: types.Int64Value(75)},
		},
	})

	readBack := readPanelViaHandlerWithPrior(t, panel, &prior)
	item := readBack.MlAnomalyChartsConfig.SeverityThreshold[0]
	require.True(t, item.Severity.IsNull())
	require.Equal(t, int64(75), item.Min.ValueInt64())
	require.True(t, item.Max.IsNull())
}

func TestImportDefault_namedForCanonicalBand(t *testing.T) {
	t.Parallel()

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{
		Type: kbapi.MlAnomalyCharts,
		Config: kbapi.KibanaHTTPAPIsMlAnomalyCharts{
			JobIds:            []string{"job-a"},
			SeverityThreshold: severityThresholdItems(t, []map[string]any{{"min": 3, "max": 25}}),
		},
	}

	readBack := readPanelViaHandler(t, panel)
	item := readBack.MlAnomalyChartsConfig.SeverityThreshold[0]
	require.Equal(t, "warning", item.Severity.ValueString())
}

func TestNullPreservation_maxSeriesToPlot(t *testing.T) {
	t.Parallel()

	maxSeries := float32(12)
	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{
		Type: kbapi.MlAnomalyCharts,
		Config: kbapi.KibanaHTTPAPIsMlAnomalyCharts{
			JobIds:          []string{"job-a"},
			MaxSeriesToPlot: &maxSeries,
		},
	}
	prior := panelWithConfig(&models.MlAnomalyChartsConfigModel{
		JobIDs:          []types.String{types.StringValue("job-a")},
		MaxSeriesToPlot: types.Int64Null(),
	})

	readBack := readPanelViaHandlerWithPrior(t, panel, &prior)
	require.True(t, readBack.MlAnomalyChartsConfig.MaxSeriesToPlot.IsNull())
}

func TestNullPreservation_timeRangeMode(t *testing.T) {
	t.Parallel()

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{
		Type: kbapi.MlAnomalyCharts,
		Config: kbapi.KibanaHTTPAPIsMlAnomalyCharts{
			JobIds: []string{"job-a"},
			TimeRange: &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
				From: "now-7d",
				To:   "now",
			},
		},
	}
	prior := panelWithConfig(&models.MlAnomalyChartsConfigModel{
		JobIDs: []types.String{types.StringValue("job-a")},
		TimeRange: &models.TimeRangeModel{
			From: types.StringValue("now-7d"),
			To:   types.StringValue("now"),
			Mode: types.StringNull(),
		},
	})

	readBack := readPanelViaHandlerWithPrior(t, panel, &prior)
	require.NotNil(t, readBack.MlAnomalyChartsConfig.TimeRange)
	require.True(t, readBack.MlAnomalyChartsConfig.TimeRange.Mode.IsNull())
}

func TestFormPreservation_mixedListAcrossRefresh(t *testing.T) {
	t.Parallel()

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{
		Type: kbapi.MlAnomalyCharts,
		Config: kbapi.KibanaHTTPAPIsMlAnomalyCharts{
			JobIds: []string{"job-a"},
			SeverityThreshold: severityThresholdItems(t, []map[string]any{
				{"min": 50, "max": 75},
				{"min": 10, "max": 20},
			}),
		},
	}
	prior := panelWithConfig(&models.MlAnomalyChartsConfigModel{
		JobIDs: []types.String{types.StringValue("job-a")},
		SeverityThreshold: []models.MlAnomalyChartsSeverityThresholdModel{
			{Severity: types.StringValue("major")},
			{Min: types.Int64Value(10), Max: types.Int64Value(20)},
		},
	})

	readBack := readPanelViaHandlerWithPrior(t, panel, &prior)
	require.Len(t, readBack.MlAnomalyChartsConfig.SeverityThreshold, 2)
	require.Equal(t, "major", readBack.MlAnomalyChartsConfig.SeverityThreshold[0].Severity.ValueString())
	require.True(t, readBack.MlAnomalyChartsConfig.SeverityThreshold[0].Min.IsNull())
	raw := readBack.MlAnomalyChartsConfig.SeverityThreshold[1]
	require.True(t, raw.Severity.IsNull())
	require.Equal(t, int64(10), raw.Min.ValueInt64())
	require.Equal(t, int64(20), raw.Max.ValueInt64())
}

func TestFormPreservation_priorNamedDriftFallback(t *testing.T) {
	t.Parallel()

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts{
		Type: kbapi.MlAnomalyCharts,
		Config: kbapi.KibanaHTTPAPIsMlAnomalyCharts{
			JobIds:            []string{"job-a"},
			SeverityThreshold: severityThresholdItems(t, []map[string]any{{"min": 50, "max": 80}}),
		},
	}
	prior := panelWithConfig(&models.MlAnomalyChartsConfigModel{
		JobIDs: []types.String{types.StringValue("job-a")},
		SeverityThreshold: []models.MlAnomalyChartsSeverityThresholdModel{
			{Severity: types.StringValue("major")},
		},
	})

	readBack := readPanelViaHandlerWithPrior(t, panel, &prior)
	item := readBack.MlAnomalyChartsConfig.SeverityThreshold[0]
	require.True(t, item.Severity.IsNull())
	require.Equal(t, int64(50), item.Min.ValueInt64())
	require.Equal(t, int64(80), item.Max.ValueInt64())
}

func TestToAPI_rejectsConfigJSON(t *testing.T) {
	t.Parallel()

	pm := models.PanelModel{
		Type: stringVal("ml_anomaly_charts"),
		MlAnomalyChartsConfig: &models.MlAnomalyChartsConfigModel{
			JobIDs: []types.String{types.StringValue("job-a")},
		},
	}
	pm.ConfigJSON = configJSONSet("{}")

	_, diags := mlanomalycharts.Handler{}.ToAPI(pm, nil)
	require.True(t, diags.HasError(), "expected config_json conflict error")
	require.Contains(t, diagSummary(diags), "config_json")
}

func panelWithConfig(cfg *models.MlAnomalyChartsConfigModel) models.PanelModel {
	return models.PanelModel{
		Type:                  types.StringValue("ml_anomaly_charts"),
		MlAnomalyChartsConfig: cfg,
	}
}

func readPanelViaHandler(t *testing.T, panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts) models.PanelModel {
	t.Helper()
	return readPanelViaHandlerWithPrior(t, panel, nil)
}

func readPanelViaHandlerWithPrior(t *testing.T, panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts, prior *models.PanelModel) models.PanelModel {
	t.Helper()
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts(panel))

	var pm models.PanelModel
	handler := mlanomalycharts.Handler{}
	diags := handler.FromAPI(context.Background(), &pm, prior, item)
	require.False(t, diags.HasError(), "%s", diags)
	return pm
}

func severityThresholdItems(t *testing.T, specs []map[string]any) *[]kbapi.KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item {
	t.Helper()
	items := make([]kbapi.KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item, len(specs))
	for i, spec := range specs {
		raw, err := json.Marshal(spec)
		require.NoError(t, err)
		require.NoError(t, items[i].UnmarshalJSON(raw))
	}
	return &items
}

func severityThresholdMaps(t *testing.T, items *[]kbapi.KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item) []map[string]any {
	t.Helper()
	if items == nil {
		return nil
	}
	out := make([]map[string]any, len(*items))
	for i, item := range *items {
		raw, err := json.Marshal(item)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(raw, &out[i]))
	}
	return out
}
