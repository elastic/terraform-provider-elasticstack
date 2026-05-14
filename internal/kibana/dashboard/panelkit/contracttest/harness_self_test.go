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

package contracttest

import (
	"context"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

var jsonDefaultsNoOp = func(m map[string]any) map[string]any { return m }

type synthStatsHandler struct{}

func (synthStatsHandler) PanelType() string { return "synthetics_stats_overview" }

func (synthStatsHandler) SchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{Optional: true},
		},
	}
}

func (synthStatsHandler) FromAPI(_ context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	apiPanel, err := item.AsKbnDashboardPanelTypeSyntheticsStatsOverview()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("panel union", err.Error())
		return d
	}

	pm.Type = types.StringValue("synthetics_stats_overview")
	pm.Grid = panelkit.GridFromAPI(apiPanel.Grid.X, apiPanel.Grid.Y, apiPanel.Grid.W, apiPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(apiPanel.Id)
	pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(jsonDefaultsNoOp)

	cfg := apiPanel.Config
	if prior == nil {
		if cfg.Title == nil {
			return nil
		}
		pm.SyntheticsStatsOverviewConfig = &models.SyntheticsStatsOverviewConfigModel{
			Title: types.StringPointerValue(cfg.Title),
		}
		return nil
	}

	existing := pm.SyntheticsStatsOverviewConfig
	if existing == nil {
		return nil
	}
	if typeutils.IsKnown(existing.Title) {
		existing.Title = types.StringPointerValue(cfg.Title)
	}
	pm.SyntheticsStatsOverviewConfig = existing
	return nil
}

func (synthStatsHandler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	sso := kbapi.KbnDashboardPanelTypeSyntheticsStatsOverview{
		Grid: grid,
		Id:   id,
		Type: kbapi.SyntheticsStatsOverview,
	}
	if cfg := pm.SyntheticsStatsOverviewConfig; cfg != nil && typeutils.IsKnown(cfg.Title) {
		t := cfg.Title.ValueString()
		sso.Config.Title = &t
	}

	var out kbapi.DashboardPanelItem
	if err := out.FromKbnDashboardPanelTypeSyntheticsStatsOverview(sso); err != nil {
		var d diag.Diagnostics
		d.AddError("ToAPI", err.Error())
		return kbapi.DashboardPanelItem{}, d
	}
	return out, nil
}

func (synthStatsHandler) ValidatePanelConfig(_ context.Context, _ map[string]attr.Value, _ path.Path) diag.Diagnostics {
	return nil
}

func (synthStatsHandler) AlignStateFromPlan(_ context.Context, _, _ *models.PanelModel) {}

func (synthStatsHandler) ClassifyJSON(_ map[string]any) bool { return false }

func (synthStatsHandler) PopulateJSONDefaults(config map[string]any) map[string]any { return config }

func (synthStatsHandler) PinnedHandler() iface.PinnedHandler { return nil }

func sloBurnRateHarnessSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"slo_id": schema.StringAttribute{
				Required: true,
			},
			"duration": schema.StringAttribute{
				Required: true,
			},
			"slo_instance_id": schema.StringAttribute{Optional: true},
			"title":           schema.StringAttribute{Optional: true},
		},
	}
}

func populateSLOBurnHarness(pm *models.PanelModel, tfPanel *models.PanelModel, apiConfig kbapi.SloBurnRateEmbeddable) {
	if tfPanel == nil {
		cfg := &models.SloBurnRateConfigModel{
			SloID:    types.StringValue(apiConfig.SloId),
			Duration: types.StringValue(apiConfig.Duration),
		}
		if apiConfig.SloInstanceId != nil && *apiConfig.SloInstanceId != "*" {
			cfg.SloInstanceID = types.StringValue(*apiConfig.SloInstanceId)
		} else {
			cfg.SloInstanceID = types.StringNull()
		}
		cfg.Title = types.StringPointerValue(apiConfig.Title)
		pm.SloBurnRateConfig = cfg
		return
	}

	existing := pm.SloBurnRateConfig
	if existing == nil {
		return
	}

	existing.SloID = types.StringValue(apiConfig.SloId)
	existing.Duration = types.StringValue(apiConfig.Duration)
	if typeutils.IsKnown(existing.SloInstanceID) {
		existing.SloInstanceID = types.StringPointerValue(apiConfig.SloInstanceId)
	}
	if typeutils.IsKnown(existing.Title) {
		existing.Title = types.StringPointerValue(apiConfig.Title)
	}
}

func buildSLOBurnHarnessPanel(pm models.PanelModel) kbapi.KbnDashboardPanelTypeSloBurnRate {
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	panel := kbapi.KbnDashboardPanelTypeSloBurnRate{
		Grid: grid,
		Id:   id,
		Type: kbapi.SloBurnRate,
	}
	cfg := pm.SloBurnRateConfig
	if cfg == nil {
		return panel
	}
	embed := kbapi.SloBurnRateEmbeddable{
		SloId:    cfg.SloID.ValueString(),
		Duration: cfg.Duration.ValueString(),
	}
	if typeutils.IsKnown(cfg.SloInstanceID) {
		embed.SloInstanceId = cfg.SloInstanceID.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Title) && cfg.Title.ValueStringPointer() != nil {
		embed.Title = cfg.Title.ValueStringPointer()
	}
	panel.Config = embed
	return panel
}

type sloBurnHarnessBase struct{}

func (sloBurnHarnessBase) PanelType() string { return "slo_burn_rate" }

func (sloBurnHarnessBase) SchemaAttribute() schema.Attribute { return sloBurnRateHarnessSchema() }

func (sloBurnHarnessBase) sloFromAPI(pm, prior *models.PanelModel, item kbapi.DashboardPanelItem, mut func(*models.SloBurnRateConfigModel, kbapi.SloBurnRateEmbeddable)) diag.Diagnostics {
	apiPanel, err := item.AsKbnDashboardPanelTypeSloBurnRate()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("panel union", err.Error())
		return d
	}
	pm.Type = types.StringValue("slo_burn_rate")
	pm.Grid = panelkit.GridFromAPI(apiPanel.Grid.X, apiPanel.Grid.Y, apiPanel.Grid.W, apiPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(apiPanel.Id)
	pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(jsonDefaultsNoOp)
	populateSLOBurnHarness(pm, prior, apiPanel.Config)
	if mut != nil && pm.SloBurnRateConfig != nil {
		mut(pm.SloBurnRateConfig, apiPanel.Config)
	}
	return nil
}

func (sloBurnHarnessBase) sloToAPI(pm models.PanelModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	panel := buildSLOBurnHarnessPanel(pm)
	var out kbapi.DashboardPanelItem
	if err := out.FromKbnDashboardPanelTypeSloBurnRate(panel); err != nil {
		var d diag.Diagnostics
		d.AddError("ToAPI", err.Error())
		return kbapi.DashboardPanelItem{}, d
	}
	return out, nil
}

func (sloBurnHarnessBase) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, _ path.Path) diag.Diagnostics {
	var d diag.Diagnostics
	sv, ok := attrs["slo_id"].(types.String)
	if !ok || sv.IsUnknown() || sv.IsNull() {
		d.AddError("validation", "slo_id is required")
	}
	dv, ok := attrs["duration"].(types.String)
	if !ok || dv.IsUnknown() || dv.IsNull() {
		d.AddError("validation", "duration is required")
	}
	return d
}

func (sloBurnHarnessBase) AlignStateFromPlan(_ context.Context, _, _ *models.PanelModel) {}

func (sloBurnHarnessBase) ClassifyJSON(_ map[string]any) bool { return false }

func (sloBurnHarnessBase) PopulateJSONDefaults(config map[string]any) map[string]any { return config }

func (sloBurnHarnessBase) PinnedHandler() iface.PinnedHandler { return nil }

type brokenSLOBurnReflect struct{ sloBurnHarnessBase }

func (brokenSLOBurnReflect) FromAPI(_ context.Context, pm *models.PanelModel, _ *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	apiPanel, err := item.AsKbnDashboardPanelTypeSloBurnRate()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("panel union", err.Error())
		return d
	}
	pm.Type = types.StringValue("slo_burn_rate")
	pm.Grid = panelkit.GridFromAPI(apiPanel.Grid.X, apiPanel.Grid.Y, apiPanel.Grid.W, apiPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(apiPanel.Id)
	pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(jsonDefaultsNoOp)
	pm.SloBurnRateConfig = nil
	return nil
}

func (brokenSLOBurnReflect) ToAPI(models.PanelModel, *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	out, err := ParseDashboardPanel(sloBurnHarnessFixtureJSON)
	if err != nil {
		var d diag.Diagnostics
		d.AddError("ToAPI", err.Error())
		return kbapi.DashboardPanelItem{}, d
	}
	return out, nil
}

type brokenSLOBurnRoundTrip struct{ sloBurnHarnessBase }

func (h brokenSLOBurnRoundTrip) FromAPI(_ context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return h.sloFromAPI(pm, prior, item, nil)
}

func (brokenSLOBurnRoundTrip) ToAPI(pm models.PanelModel, _ *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	if cfg := pm.SloBurnRateConfig; cfg != nil {
		dupCfg := *cfg
		dupCfg.Duration = types.StringValue("999d")
		pmDup := pm
		pmDup.SloBurnRateConfig = &dupCfg
		return sloBurnHarnessBase{}.sloToAPI(pmDup)
	}
	return sloBurnHarnessBase{}.sloToAPI(pm)
}

type brokenSLOBurnSchema struct{ sloBurnHarnessBase }

func (h brokenSLOBurnSchema) FromAPI(_ context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return h.sloFromAPI(pm, prior, item, nil)
}

func (h brokenSLOBurnSchema) ToAPI(pm models.PanelModel, _ *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	return h.sloToAPI(pm)
}

func (brokenSLOBurnSchema) ValidatePanelConfig(_ context.Context, _ map[string]attr.Value, _ path.Path) diag.Diagnostics {
	return nil
}

type brokenSLOBurnNullPreserve struct{ sloBurnHarnessBase }

func (h brokenSLOBurnNullPreserve) FromAPI(_ context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return h.sloFromAPI(pm, prior, item, func(cfg *models.SloBurnRateConfigModel, api kbapi.SloBurnRateEmbeddable) {
		cfg.Title = types.StringPointerValue(api.Title)
	})
}

func (h brokenSLOBurnNullPreserve) ToAPI(pm models.PanelModel, _ *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	return h.sloToAPI(pm)
}

const sloBurnHarnessFixtureJSON = `{
  "type": "slo_burn_rate",
  "grid": { "x": 0, "y": 0, "w": 6, "h": 4 },
  "id": "slo-br-contract-harness",
  "config": { "sloId": "slo-contract-id", "duration": "5m", "title": "api-title-harness" }
}`

func TestContractHarness_syntheticsStatsOverviewSmoke(t *testing.T) {
	const fixture = `{
  "type": "synthetics_stats_overview",
  "grid": { "x": 0, "y": 0, "w": 24, "h": 8 },
  "id": "synth-contract",
  "config": { "title": "Hello synthetics" }
}`
	Run(t, synthStatsHandler{}, Config{FullAPIResponse: fixture})
	require.False(t, t.Failed(), "expected harness to pass for minimal synthetics handler")
}

func TestContractHarness_PlantedFailuresSLOBurnShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		h    iface.Handler
		tag  string
	}{
		{"reflect_stub", brokenSLOBurnReflect{}, "[Reflect]"},
		{"roundtrip_stub", brokenSLOBurnRoundTrip{}, "[RoundTrip]"},
		{"schema_stub", brokenSLOBurnSchema{}, "[Schema]"},
		{"null_preserve_stub", brokenSLOBurnNullPreserve{}, "[NullPreserve]"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			msgs := runChecks(ctx, tc.h, Config{FullAPIResponse: sloBurnHarnessFixtureJSON})
			require.True(t,
				slicesContainAnyPrefix(msgs, tc.tag),
				"want at least one issue containing prefix %s; messages:\n%s", tc.tag, strings.Join(msgs, "\n"))
		})
	}
}

func slicesContainAnyPrefix(xs []string, prefix string) bool {
	for _, s := range xs {
		if strings.Contains(s, prefix) {
			return true
		}
	}
	return false
}
