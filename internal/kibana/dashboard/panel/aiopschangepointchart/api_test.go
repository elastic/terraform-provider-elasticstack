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

package aiopschangepointchart_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/aiopschangepointchart"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func stringVal(s string) types.String { return types.StringValue(s) }
func boolVal(b bool) types.Bool       { return types.BoolValue(b) }
func float32Val(f float32) types.Float32 {
	return types.Float32Value(f)
}
func stringNull() types.String { return types.StringNull() }
func float32Null() types.Float32 {
	return types.Float32Null()
}

func stringSet(values ...string) types.Set {
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}
	s, diags := types.SetValue(types.StringType, elems)
	if diags.HasError() {
		return types.SetNull(types.StringType)
	}
	return s
}

func configJSONSet(s string) customtypes.JSONWithDefaultsValue[map[string]any] {
	return customtypes.NewJSONWithDefaultsValue(s, func(m map[string]any) map[string]any { return m })
}

func configMap(t *testing.T, item kbapi.DashboardPanelItem) map[string]any {
	t.Helper()
	raw, err := json.Marshal(item)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	cfg, ok := m["config"].(map[string]any)
	require.True(t, ok, "config should be object")
	return cfg
}

func diagSummary(diags diag.Diagnostics) string {
	if diags == nil {
		return ""
	}
	var b strings.Builder
	for _, d := range diags {
		b.WriteString(d.Severity().String())
		b.WriteString(": ")
		b.WriteString(d.Summary())
		if dt := d.Detail(); dt != "" {
			b.WriteString(" — ")
			b.WriteString(dt)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func TestContract(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, aiopschangepointchart.Handler{}, contracttest.Config{
		FullAPIResponse: `{
			"type": "aiops_change_point_chart",
			"grid": {"x": 0, "y": 0, "w": 24, "h": 15},
			"id": "aiops-cpc-contract",
			"config": {
				"data_view_id": "metrics-*",
				"metric_field": "system.cpu.total.pct",
				"aggregation_function": "avg",
				"split_field": "host.name",
				"partitions": ["host-a", "host-b"],
				"max_series_to_plot": 6,
				"view_type": "charts",
				"title": "Change points"
			}
		}`,
		// The optional `time_range` SingleNestedAttribute has required inner `from`/`to` leaves, so a
		// fixture that omits the panel-level time_range cannot satisfy the harness's required-leaf walk.
		OmitRequiredLeafPresence: true,
		SkipFields:               []string{"config.time_range", "time_range"},
	})
}

func TestBuildConfig_requiredOnly(t *testing.T) {
	t.Parallel()

	pm := models.PanelModel{
		AiopsChangePointChartConfig: &models.AiopsChangePointChartConfigModel{
			DataViewID:  stringVal("metrics-*"),
			MetricField: stringVal("system.cpu.total.pct"),
			Partitions:  types.SetNull(types.StringType),
		},
	}

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart{}
	diags := aiopschangepointchart.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%s", diags)

	require.Equal(t, "metrics-*", panel.Config.DataViewId)
	require.Equal(t, "system.cpu.total.pct", panel.Config.MetricField)
	require.Nil(t, panel.Config.AggregationFunction)
	require.Nil(t, panel.Config.SplitField)
	require.Nil(t, panel.Config.Partitions)
	require.Nil(t, panel.Config.MaxSeriesToPlot)
	require.Nil(t, panel.Config.ViewType)
	require.Nil(t, panel.Config.TimeRange)
}

func TestBuildConfig_allOptional(t *testing.T) {
	t.Parallel()

	pm := models.PanelModel{
		AiopsChangePointChartConfig: &models.AiopsChangePointChartConfigModel{
			DataViewID:          stringVal("metrics-*"),
			MetricField:         stringVal("system.cpu.total.pct"),
			AggregationFunction: stringVal("avg"),
			SplitField:          stringVal("host.name"),
			Partitions:          stringSet("host-a", "host-b"),
			MaxSeriesToPlot:     float32Val(6),
			ViewType:            stringVal("charts"),
			Title:               stringVal("Change points"),
			HideBorder:          boolVal(true),
		},
	}

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart{}
	diags := aiopschangepointchart.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%s", diags)

	require.Equal(t, "metrics-*", panel.Config.DataViewId)
	require.Equal(t, "system.cpu.total.pct", panel.Config.MetricField)
	require.NotNil(t, panel.Config.AggregationFunction)
	require.Equal(t, "avg", string(*panel.Config.AggregationFunction))
	require.NotNil(t, panel.Config.SplitField)
	require.Equal(t, "host.name", *panel.Config.SplitField)
	require.NotNil(t, panel.Config.Partitions)
	require.ElementsMatch(t, []string{"host-a", "host-b"}, *panel.Config.Partitions)
	require.NotNil(t, panel.Config.MaxSeriesToPlot)
	require.InDelta(t, 6, float64(*panel.Config.MaxSeriesToPlot), 1e-6)
	require.NotNil(t, panel.Config.ViewType)
	require.Equal(t, "charts", string(*panel.Config.ViewType))
}

func TestPopulateFromAPI_nullPreservation(t *testing.T) {
	t.Parallel()

	agg := kbapi.KibanaHTTPAPIsAiopsChangePointChartAggregationFunction("avg")
	vt := kbapi.KibanaHTTPAPIsAiopsChangePointChartViewType("charts")
	api := kbapi.KibanaHTTPAPIsAiopsChangePointChart{
		DataViewId:          "metrics-*",
		MetricField:         "system.cpu.total.pct",
		AggregationFunction: &agg,
		SplitField:          new("host.name"),
		Partitions:          &[]string{"host-a", "host-b"},
		MaxSeriesToPlot:     new(float32(6)),
		ViewType:            &vt,
	}

	prior := &models.PanelModel{
		AiopsChangePointChartConfig: &models.AiopsChangePointChartConfigModel{
			DataViewID:          stringVal("metrics-*"),
			MetricField:         stringVal("system.cpu.total.pct"),
			AggregationFunction: stringNull(),
			SplitField:          stringNull(),
			Partitions:          types.SetNull(types.StringType),
			MaxSeriesToPlot:     float32Null(),
			ViewType:            stringNull(),
			TimeRange:           nil,
		},
	}
	pm := &models.PanelModel{
		AiopsChangePointChartConfig: &models.AiopsChangePointChartConfigModel{
			DataViewID:          stringVal("metrics-*"),
			MetricField:         stringVal("system.cpu.total.pct"),
			AggregationFunction: stringNull(),
			SplitField:          stringNull(),
			Partitions:          types.SetNull(types.StringType),
			MaxSeriesToPlot:     float32Null(),
			ViewType:            stringNull(),
			TimeRange:           nil,
		},
	}

	diags := aiopschangepointchart.PopulateFromAPI(pm, prior, api)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := pm.AiopsChangePointChartConfig
	require.Equal(t, "metrics-*", cfg.DataViewID.ValueString())
	require.Equal(t, "system.cpu.total.pct", cfg.MetricField.ValueString())
	require.True(t, cfg.AggregationFunction.IsNull())
	require.True(t, cfg.SplitField.IsNull())
	require.True(t, cfg.Partitions.IsNull())
	require.True(t, cfg.MaxSeriesToPlot.IsNull())
	require.True(t, cfg.ViewType.IsNull())
	require.Nil(t, cfg.TimeRange)
}

func TestPopulateFromAPI_import(t *testing.T) {
	t.Parallel()

	agg := kbapi.KibanaHTTPAPIsAiopsChangePointChartAggregationFunction("sum")
	api := kbapi.KibanaHTTPAPIsAiopsChangePointChart{
		DataViewId:          "metrics-*",
		MetricField:         "system.cpu.total.pct",
		AggregationFunction: &agg,
		Partitions:          &[]string{"host-a"},
	}

	pm := &models.PanelModel{}
	diags := aiopschangepointchart.PopulateFromAPI(pm, nil, api)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := pm.AiopsChangePointChartConfig
	require.Equal(t, "metrics-*", cfg.DataViewID.ValueString())
	require.Equal(t, "system.cpu.total.pct", cfg.MetricField.ValueString())
	require.Equal(t, "sum", cfg.AggregationFunction.ValueString())
	require.True(t, cfg.SplitField.IsNull())
	require.False(t, cfg.Partitions.IsNull())
	require.True(t, cfg.ViewType.IsNull())
	require.Nil(t, cfg.TimeRange)
}

func TestPopulateFromAPI_typeChangeRecovery(t *testing.T) {
	t.Parallel()

	agg := kbapi.KibanaHTTPAPIsAiopsChangePointChartAggregationFunction("avg")
	vt := kbapi.KibanaHTTPAPIsAiopsChangePointChartViewType("charts")
	from, to := "now-15m", "now"
	tr := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{From: from, To: to}
	api := kbapi.KibanaHTTPAPIsAiopsChangePointChart{
		DataViewId:          "logs-*",
		MetricField:         "system.memory.used.pct",
		AggregationFunction: &agg,
		ViewType:            &vt,
		TimeRange:           &tr,
	}

	// pm has no config (panel type changed away from this type in the plan)
	// but prior still has the config block with known optional fields.
	pm := &models.PanelModel{}
	prior := &models.PanelModel{
		AiopsChangePointChartConfig: &models.AiopsChangePointChartConfigModel{
			DataViewID:          stringVal("old-dv"),
			MetricField:         stringVal("old.field"),
			AggregationFunction: stringVal("sum"),
			ViewType:            stringVal("bar"),
		},
	}

	diags := aiopschangepointchart.PopulateFromAPI(pm, prior, api)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := pm.AiopsChangePointChartConfig
	require.NotNil(t, cfg, "type-change path should populate config from API")
	// Required fields always come from the API.
	require.Equal(t, "logs-*", cfg.DataViewID.ValueString())
	require.Equal(t, "system.memory.used.pct", cfg.MetricField.ValueString())
	// Optional fields that were known in prior are updated from API.
	require.Equal(t, "avg", cfg.AggregationFunction.ValueString())
	require.Equal(t, "charts", cfg.ViewType.ValueString())
	require.NotNil(t, cfg.TimeRange)
	require.Equal(t, from, cfg.TimeRange.From.ValueString())
	require.Equal(t, to, cfg.TimeRange.To.ValueString())
	require.True(t, cfg.TimeRange.Mode.IsNull())
}

func TestPopulateFromAPI_partitionsOrderInsensitive(t *testing.T) {
	t.Parallel()

	api := kbapi.KibanaHTTPAPIsAiopsChangePointChart{
		DataViewId:  "metrics-*",
		MetricField: "system.cpu.total.pct",
		Partitions:  &[]string{"host-b", "host-a", "host-c"},
	}

	prior := &models.PanelModel{
		AiopsChangePointChartConfig: &models.AiopsChangePointChartConfigModel{
			DataViewID:  stringVal("metrics-*"),
			MetricField: stringVal("system.cpu.total.pct"),
			Partitions:  stringSet("host-a", "host-b", "host-c"),
		},
	}
	pm := &models.PanelModel{
		AiopsChangePointChartConfig: &models.AiopsChangePointChartConfigModel{
			DataViewID:  stringVal("metrics-*"),
			MetricField: stringVal("system.cpu.total.pct"),
			Partitions:  stringSet("host-a", "host-b", "host-c"),
		},
	}

	diags := aiopschangepointchart.PopulateFromAPI(pm, prior, api)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := pm.AiopsChangePointChartConfig
	require.False(t, cfg.Partitions.IsNull())
	elems := cfg.Partitions.Elements()
	got := make([]string, 0, len(elems))
	for _, e := range elems {
		got = append(got, e.(types.String).ValueString())
	}
	require.ElementsMatch(t, []string{"host-a", "host-b", "host-c"}, got)
}

func TestToAPI_rejectsConfigJSON(t *testing.T) {
	t.Parallel()

	pm := models.PanelModel{
		Type: stringVal("aiops_change_point_chart"),
		AiopsChangePointChartConfig: &models.AiopsChangePointChartConfigModel{
			DataViewID:  stringVal("metrics-*"),
			MetricField: stringVal("system.cpu.total.pct"),
		},
	}
	pm.ConfigJSON = configJSONSet("{}")

	_, diags := aiopschangepointchart.Handler{}.ToAPI(pm, nil)
	require.True(t, diags.HasError(), "expected config_json conflict error")
	require.Contains(t, diagSummary(diags), "config_json")
}

func TestRoundtrip_viaHandler(t *testing.T) {
	t.Parallel()

	const fixtureJSON = `{
		"type": "aiops_change_point_chart",
		"grid": {"x": 0, "y": 0, "w": 24, "h": 15},
		"id": "aiops-cpc-rt",
		"config": {
			"data_view_id": "metrics-*",
			"metric_field": "system.cpu.total.pct",
			"aggregation_function": "avg",
			"split_field": "host.name",
			"partitions": ["host-a", "host-b"],
			"max_series_to_plot": 6,
			"view_type": "charts"
		}
	}`

	item0, err := contracttest.ParseDashboardPanel(fixtureJSON)
	require.NoError(t, err)

	var pm models.PanelModel
	handler := aiopschangepointchart.Handler{}
	diags := handler.FromAPI(context.Background(), &pm, nil, item0)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.AiopsChangePointChartConfig)
	require.Equal(t, "metrics-*", pm.AiopsChangePointChartConfig.DataViewID.ValueString())
	require.Equal(t, "system.cpu.total.pct", pm.AiopsChangePointChartConfig.MetricField.ValueString())

	item1, d2 := handler.ToAPI(pm, nil)
	require.False(t, d2.HasError(), "%s", d2)

	cfg0 := configMap(t, item0)
	cfg1 := configMap(t, item1)
	require.Equal(t, cfg0["data_view_id"], cfg1["data_view_id"])
	require.Equal(t, cfg0["metric_field"], cfg1["metric_field"])
	require.Equal(t, cfg0["aggregation_function"], cfg1["aggregation_function"])
	require.Equal(t, cfg0["split_field"], cfg1["split_field"])
	require.ElementsMatch(t, toStringSlice(cfg0["partitions"]), toStringSlice(cfg1["partitions"]))
	require.Equal(t, cfg0["view_type"], cfg1["view_type"])
}

func toStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
