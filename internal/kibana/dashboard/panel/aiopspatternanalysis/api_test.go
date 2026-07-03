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

package aiopspatternanalysis_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/aiopspatternanalysis"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
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

	contracttest.Run(t, aiopspatternanalysis.Handler{}, contracttest.Config{
		FullAPIResponse: `{
			"type": "aiops_pattern_analysis",
			"grid": {"x": 0, "y": 0, "w": 24, "h": 15},
			"id": "aiops-pa-contract",
			"config": {
				"data_view_id": "logs-*",
				"field_name": "message",
				"minimum_time_range": "1_week",
				"random_sampler_mode": "on_manual",
				"random_sampler_probability": 0.01,
				"title": "Patterns"
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
		AiopsPatternAnalysisConfig: &models.AiopsPatternAnalysisConfigModel{
			DataViewID: stringVal("logs-*"),
			FieldName:  stringVal("message"),
		},
	}

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsPatternAnalysis{}
	diags := aiopspatternanalysis.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%s", diags)

	require.Equal(t, "logs-*", panel.Config.DataViewId)
	require.Equal(t, "message", panel.Config.FieldName)
	require.Nil(t, panel.Config.MinimumTimeRange)
	require.Nil(t, panel.Config.RandomSamplerMode)
	require.Nil(t, panel.Config.RandomSamplerProbability)
	require.Nil(t, panel.Config.Title)
	require.Nil(t, panel.Config.TimeRange)
}

func TestBuildConfig_allOptional(t *testing.T) {
	t.Parallel()

	pm := models.PanelModel{
		AiopsPatternAnalysisConfig: &models.AiopsPatternAnalysisConfigModel{
			DataViewID:               stringVal("logs-*"),
			FieldName:                stringVal("message"),
			MinimumTimeRange:         stringVal("1_week"),
			RandomSamplerMode:        stringVal("on_manual"),
			RandomSamplerProbability: float32Val(0.01),
			Title:                    stringVal("Patterns"),
			Description:              stringVal("Pattern panel"),
			HideTitle:                boolVal(true),
			HideBorder:               boolVal(false),
			TimeRange: &models.TimeRangeModel{
				From: stringVal("now-15m"),
				To:   stringVal("now"),
			},
		},
	}

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsPatternAnalysis{}
	diags := aiopspatternanalysis.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%s", diags)

	require.Equal(t, "logs-*", panel.Config.DataViewId)
	require.Equal(t, "message", panel.Config.FieldName)
	require.NotNil(t, panel.Config.MinimumTimeRange)
	require.Equal(t, "1_week", string(*panel.Config.MinimumTimeRange))
	require.NotNil(t, panel.Config.RandomSamplerMode)
	require.Equal(t, "on_manual", string(*panel.Config.RandomSamplerMode))
	require.NotNil(t, panel.Config.RandomSamplerProbability)
	require.InDelta(t, 0.01, float64(*panel.Config.RandomSamplerProbability), 1e-6)
	require.NotNil(t, panel.Config.TimeRange)
	require.Equal(t, "now-15m", panel.Config.TimeRange.From)
}

func TestPopulateFromAPI_nullPreservation(t *testing.T) {
	t.Parallel()

	mtr := kbapi.KibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange("1_week")
	rsm := kbapi.KibanaHTTPAPIsAiopsPatternAnalysisRandomSamplerMode("on_automatic")
	api := kbapi.KibanaHTTPAPIsAiopsPatternAnalysis{
		DataViewId:               "logs-*",
		FieldName:                "message",
		MinimumTimeRange:         &mtr,
		RandomSamplerMode:        &rsm,
		RandomSamplerProbability: new(float32(0.02)),
	}

	prior := &models.PanelModel{
		AiopsPatternAnalysisConfig: &models.AiopsPatternAnalysisConfigModel{
			DataViewID:               stringVal("logs-*"),
			FieldName:                stringVal("message"),
			MinimumTimeRange:         stringNull(),
			RandomSamplerMode:        stringNull(),
			RandomSamplerProbability: float32Null(),
			TimeRange:                nil,
		},
	}
	pm := &models.PanelModel{
		AiopsPatternAnalysisConfig: &models.AiopsPatternAnalysisConfigModel{
			DataViewID:               stringVal("logs-*"),
			FieldName:                stringVal("message"),
			MinimumTimeRange:         stringNull(),
			RandomSamplerMode:        stringNull(),
			RandomSamplerProbability: float32Null(),
			TimeRange:                nil,
		},
	}
	diags := aiopspatternanalysis.PopulateFromAPI(pm, prior, api)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := pm.AiopsPatternAnalysisConfig
	require.Equal(t, "logs-*", cfg.DataViewID.ValueString())
	require.Equal(t, "message", cfg.FieldName.ValueString())
	require.True(t, cfg.MinimumTimeRange.IsNull())
	require.True(t, cfg.RandomSamplerMode.IsNull())
	require.True(t, cfg.RandomSamplerProbability.IsNull())
	require.Nil(t, cfg.TimeRange)
}

func TestPopulateFromAPI_import(t *testing.T) {
	t.Parallel()

	mtr := kbapi.KibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange("1_month")
	api := kbapi.KibanaHTTPAPIsAiopsPatternAnalysis{
		DataViewId:       "logs-*",
		FieldName:        "message",
		MinimumTimeRange: &mtr,
	}

	pm := &models.PanelModel{}
	diags := aiopspatternanalysis.PopulateFromAPI(pm, nil, api)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := pm.AiopsPatternAnalysisConfig
	require.Equal(t, "logs-*", cfg.DataViewID.ValueString())
	require.Equal(t, "message", cfg.FieldName.ValueString())
	require.Equal(t, "1_month", cfg.MinimumTimeRange.ValueString())
	require.True(t, cfg.RandomSamplerMode.IsNull())
	require.True(t, cfg.RandomSamplerProbability.IsNull())
	require.Nil(t, cfg.TimeRange)
}

// TestPopulateFromAPI_typeChangeRecovery verifies the type-change path (pm has no config but
// prior does) rebuilds config entirely from the API, including TimeRange.
func TestPopulateFromAPI_typeChangeRecovery(t *testing.T) {
	t.Parallel()

	mtr := kbapi.KibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange("1_week")
	from, to := "now-30m", "now"
	tr := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{From: from, To: to}
	api := kbapi.KibanaHTTPAPIsAiopsPatternAnalysis{
		DataViewId:       "logs-*",
		FieldName:        "message",
		MinimumTimeRange: &mtr,
		TimeRange:        &tr,
	}

	// pm has no config (panel type changed away from this type in the plan)
	// but prior still has the config block.
	pm := &models.PanelModel{}
	prior := &models.PanelModel{
		AiopsPatternAnalysisConfig: &models.AiopsPatternAnalysisConfigModel{
			DataViewID: stringVal("old-dv"),
			FieldName:  stringVal("old.field"),
		},
	}

	diags := aiopspatternanalysis.PopulateFromAPI(pm, prior, api)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := pm.AiopsPatternAnalysisConfig
	require.NotNil(t, cfg, "type-change path should populate config from API")
	require.Equal(t, "logs-*", cfg.DataViewID.ValueString())
	require.Equal(t, "message", cfg.FieldName.ValueString())
	require.Equal(t, "1_week", cfg.MinimumTimeRange.ValueString())
	require.NotNil(t, cfg.TimeRange, "type-change path must initialise TimeRange from API")
	require.Equal(t, from, cfg.TimeRange.From.ValueString())
	require.Equal(t, to, cfg.TimeRange.To.ValueString())
	require.True(t, cfg.TimeRange.Mode.IsNull())
}

func TestToAPI_rejectsConfigJSON(t *testing.T) {
	t.Parallel()

	pm := models.PanelModel{
		Type: stringVal("aiops_pattern_analysis"),
		AiopsPatternAnalysisConfig: &models.AiopsPatternAnalysisConfigModel{
			DataViewID: stringVal("logs-*"),
			FieldName:  stringVal("message"),
		},
	}
	pm.ConfigJSON = configJSONSet("{}")

	_, diags := aiopspatternanalysis.Handler{}.ToAPI(pm, nil)
	require.True(t, diags.HasError(), "expected config_json conflict error")
	require.Contains(t, diagSummary(diags), "config_json")
}

func TestRoundtrip_viaHandler(t *testing.T) {
	t.Parallel()

	const fixtureJSON = `{
		"type": "aiops_pattern_analysis",
		"grid": {"x": 0, "y": 0, "w": 24, "h": 15},
		"id": "aiops-pa-rt",
		"config": {
			"data_view_id": "logs-*",
			"field_name": "message",
			"minimum_time_range": "1_week",
			"random_sampler_mode": "on_manual",
			"random_sampler_probability": 0.01
		}
	}`

	item0, err := contracttest.ParseDashboardPanel(fixtureJSON)
	require.NoError(t, err)

	var pm models.PanelModel
	handler := aiopspatternanalysis.Handler{}
	diags := handler.FromAPI(context.Background(), &pm, nil, item0)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.AiopsPatternAnalysisConfig)
	require.Equal(t, "logs-*", pm.AiopsPatternAnalysisConfig.DataViewID.ValueString())
	require.Equal(t, "message", pm.AiopsPatternAnalysisConfig.FieldName.ValueString())

	item1, d2 := handler.ToAPI(pm, nil)
	require.False(t, d2.HasError(), "%s", d2)

	cfg0 := configMap(t, item0)
	cfg1 := configMap(t, item1)
	require.Equal(t, cfg0["data_view_id"], cfg1["data_view_id"])
	require.Equal(t, cfg0["field_name"], cfg1["field_name"])
	require.Equal(t, cfg0["minimum_time_range"], cfg1["minimum_time_range"])
	require.Equal(t, cfg0["random_sampler_mode"], cfg1["random_sampler_mode"])
	require.Equal(t, cfg0["random_sampler_probability"], cfg1["random_sampler_probability"])
}
