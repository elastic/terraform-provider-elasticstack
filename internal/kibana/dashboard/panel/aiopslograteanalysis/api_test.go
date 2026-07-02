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

package aiopslograteanalysis_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/aiopslograteanalysis"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/stretchr/testify/require"
)

func TestContract(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, aiopslograteanalysis.Handler{}, contracttest.Config{
		FullAPIResponse: `{
			"type": "aiops_log_rate_analysis",
			"grid": {"x": 0, "y": 0, "w": 24, "h": 15},
			"id": "aiops-lra-contract",
			"config": {
				"data_view_id": "logs-*",
				"title": "Log spikes",
				"hide_title": true
			}
		}`,
		// The optional `time_range` SingleNestedAttribute has required inner `from`/`to` leaves, so a
		// fixture that omits the panel-level time_range cannot satisfy the harness's required-leaf-presence
		// walk (flat JSON vs nested TF paths). Disable that phase; time_range round-trip/null-preservation
		// is covered by the dedicated unit tests below.
		OmitRequiredLeafPresence: true,
		// time_range is a nested object whose round-trip/null-preserve walking is handled by the
		// dedicated null-preservation test below; skip it in the flat harness checks.
		SkipFields: []string{"config.time_range", "time_range"},
	})
}

// TestBuildConfig_requiredOnly verifies a required-only config serializes data_view_id and omits
// optional fields from the API payload.
func TestBuildConfig_requiredOnly(t *testing.T) {
	t.Parallel()

	pm := models.PanelModel{
		AiopsLogRateAnalysisConfig: &models.AiopsLogRateAnalysisConfigModel{
			DataViewID: stringVal("logs-*"),
		},
	}

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsLogRateAnalysis{}
	diags := aiopslograteanalysis.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%s", diags)

	require.Equal(t, "logs-*", panel.Config.DataViewId)
	require.Nil(t, panel.Config.Title)
	require.Nil(t, panel.Config.Description)
	require.Nil(t, panel.Config.HideTitle)
	require.Nil(t, panel.Config.HideBorder)
	require.Nil(t, panel.Config.TimeRange)
}

// TestBuildConfig_allOptional verifies all optional fields serialize into the API payload.
func TestBuildConfig_allOptional(t *testing.T) {
	t.Parallel()

	hideTitle := true
	hideBorder := false
	pm := models.PanelModel{
		AiopsLogRateAnalysisConfig: &models.AiopsLogRateAnalysisConfigModel{
			DataViewID:  stringVal("logs-*"),
			Title:       stringVal("Log spikes"),
			Description: stringVal("Spike panel"),
			HideTitle:   boolVal(hideTitle),
			HideBorder:  boolVal(hideBorder),
			TimeRange: &models.TimeRangeModel{
				From: stringVal("now-15m"),
				To:   stringVal("now"),
				Mode: stringVal("relative"),
			},
		},
	}

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsLogRateAnalysis{}
	diags := aiopslograteanalysis.BuildConfig(pm, &panel)
	require.False(t, diags.HasError(), "%s", diags)

	require.Equal(t, "logs-*", panel.Config.DataViewId)
	require.Equal(t, "Log spikes", *panel.Config.Title)
	require.Equal(t, "Spike panel", *panel.Config.Description)
	require.True(t, *panel.Config.HideTitle)
	require.False(t, *panel.Config.HideBorder)
	require.NotNil(t, panel.Config.TimeRange)
	require.Equal(t, "now-15m", panel.Config.TimeRange.From)
	require.Equal(t, "now", panel.Config.TimeRange.To)
	require.NotNil(t, panel.Config.TimeRange.Mode)
	require.Equal(t, "relative", string(*panel.Config.TimeRange.Mode))
}

// TestPopulateFromAPI_nullPreservation verifies REQ-009: optional fields that were null in prior
// state stay null after read even when the API returns server-side values.
func TestPopulateFromAPI_nullPreservation(t *testing.T) {
	t.Parallel()

	api := kbapi.KibanaHTTPAPIsAiopsLogRateAnalysis{
		DataViewId:  "logs-*",
		Title:       new("Server default title"),
		Description: new("Server default description"),
		HideTitle:   new(true),
		HideBorder:  new(false),
	}

	// Prior state: practitioner omitted all optional fields (null).
	prior := &models.PanelModel{
		AiopsLogRateAnalysisConfig: &models.AiopsLogRateAnalysisConfigModel{
			DataViewID:  stringVal("logs-*"),
			Title:       stringNull(),
			Description: stringNull(),
			HideTitle:   boolNull(),
			HideBorder:  boolNull(),
			TimeRange:   nil,
		},
	}

	pm := &models.PanelModel{
		AiopsLogRateAnalysisConfig: &models.AiopsLogRateAnalysisConfigModel{
			DataViewID:  stringVal("logs-*"),
			Title:       stringNull(),
			Description: stringNull(),
			HideTitle:   boolNull(),
			HideBorder:  boolNull(),
			TimeRange:   nil,
		},
	}

	diags := aiopslograteanalysis.PopulateFromAPI(pm, prior, api)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := pm.AiopsLogRateAnalysisConfig
	require.Equal(t, "logs-*", cfg.DataViewID.ValueString())
	// Optional fields must stay null despite the API returning values.
	require.True(t, cfg.Title.IsNull())
	require.True(t, cfg.Description.IsNull())
	require.True(t, cfg.HideTitle.IsNull())
	require.True(t, cfg.HideBorder.IsNull())
	require.Nil(t, cfg.TimeRange)
}

// TestPopulateFromAPI_import verifies import (prior == nil) populates required fields and optional
// fields only when the API returns non-nil values.
func TestPopulateFromAPI_import(t *testing.T) {
	t.Parallel()

	api := kbapi.KibanaHTTPAPIsAiopsLogRateAnalysis{
		DataViewId: "logs-*",
		Title:      new("Imported title"),
	}

	pm := &models.PanelModel{}
	diags := aiopslograteanalysis.PopulateFromAPI(pm, nil, api)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := pm.AiopsLogRateAnalysisConfig
	require.Equal(t, "logs-*", cfg.DataViewID.ValueString())
	require.Equal(t, "Imported title", cfg.Title.ValueString())
	require.True(t, cfg.Description.IsNull())
	require.True(t, cfg.HideTitle.IsNull())
	require.Nil(t, cfg.TimeRange)
}

// TestPopulateFromAPI_typeChangeRecovery verifies the type-change path (pm has no config but
// prior does) initialises config from the API including TimeRange.
func TestPopulateFromAPI_typeChangeRecovery(t *testing.T) {
	t.Parallel()

	from, to := "now-30m", "now"
	tr := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{From: from, To: to}
	api := kbapi.KibanaHTTPAPIsAiopsLogRateAnalysis{
		DataViewId: "logs-*",
		Title:      new("Recovered title"),
		TimeRange:  &tr,
	}

	pm := &models.PanelModel{}
	// Prior has known Title so that null-preservation doesn't wipe it out after
	// the type-change init falls through to the merge path.
	prior := &models.PanelModel{
		AiopsLogRateAnalysisConfig: &models.AiopsLogRateAnalysisConfigModel{
			DataViewID: stringVal("old-dv"),
			Title:      stringVal("old-title"),
		},
	}

	diags := aiopslograteanalysis.PopulateFromAPI(pm, prior, api)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := pm.AiopsLogRateAnalysisConfig
	require.NotNil(t, cfg, "type-change path should populate config from API")
	require.Equal(t, "logs-*", cfg.DataViewID.ValueString())
	// Title was known in prior so it gets updated from the API.
	require.Equal(t, "Recovered title", cfg.Title.ValueString())
	require.NotNil(t, cfg.TimeRange)
	require.Equal(t, from, cfg.TimeRange.From.ValueString())
	require.Equal(t, to, cfg.TimeRange.To.ValueString())
	require.True(t, cfg.TimeRange.Mode.IsNull())

// TestToAPI_rejectsConfigJSON verifies simultaneous typed config and config_json is rejected.
func TestToAPI_rejectsConfigJSON(t *testing.T) {
	t.Parallel()

	pm := models.PanelModel{
		Type: stringVal("aiops_log_rate_analysis"),
		AiopsLogRateAnalysisConfig: &models.AiopsLogRateAnalysisConfigModel{
			DataViewID: stringVal("logs-*"),
		},
	}
	pm.ConfigJSON = configJSONSet("{}")

	_, diags := aiopslograteanalysis.Handler{}.ToAPI(pm, nil)
	require.True(t, diags.HasError(), "expected config_json conflict error")
	require.Contains(t, diagSummary(diags), "config_json")
}

// TestRoundtrip_viaHandler verifies FromAPI followed by ToAPI reproduces the API config.
func TestRoundtrip_viaHandler(t *testing.T) {
	t.Parallel()

	const fixtureJSON = `{
		"type": "aiops_log_rate_analysis",
		"grid": {"x": 0, "y": 0, "w": 24, "h": 15},
		"id": "aiops-lra-rt",
		"config": {
			"data_view_id": "logs-*",
			"title": "Log spikes",
			"hide_border": true
		}
	}`

	item0, err := contracttest.ParseDashboardPanel(fixtureJSON)
	require.NoError(t, err)

	var pm models.PanelModel
	handler := aiopslograteanalysis.Handler{}
	diags := handler.FromAPI(context.Background(), &pm, nil, item0)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.AiopsLogRateAnalysisConfig)
	require.Equal(t, "logs-*", pm.AiopsLogRateAnalysisConfig.DataViewID.ValueString())

	item1, d2 := handler.ToAPI(pm, nil)
	require.False(t, d2.HasError(), "%s", d2)

	cfg0 := configMap(t, item0)
	cfg1 := configMap(t, item1)
	require.Equal(t, cfg0["data_view_id"], cfg1["data_view_id"])
	require.Equal(t, cfg0["title"], cfg1["title"])
	require.Equal(t, cfg0["hide_border"], cfg1["hide_border"])
}
