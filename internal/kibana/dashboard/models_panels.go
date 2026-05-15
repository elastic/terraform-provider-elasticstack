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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func dashboardMapPanelsFromAPI(ctx context.Context, m *models.DashboardModel, apiPanels *kbapi.DashboardPanels) ([]models.PanelModel, []models.SectionModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if apiPanels == nil || len(*apiPanels) == 0 {
		return nil, nil, diags
	}

	var panels []models.PanelModel
	var sections []models.SectionModel

	for _, item := range *apiPanels {
		// Try section first; this avoids treating section items as panels.
		section, err := item.AsKbnDashboardSection()
		if err == nil && section.Title != "" {
			tfSectionIndex := len(sections)
			var tfSection *models.SectionModel
			if tfSectionIndex < len(m.Sections) {
				tfSection = &m.Sections[tfSectionIndex]
			}
			sectionModel, d := dashboardMapSectionFromAPI(ctx, m, tfSection, section)
			diags.Append(d...)
			if diags.HasError() {
				return nil, nil, diags
			}
			sections = append(sections, sectionModel)
			continue
		}

		panelItem, err := item.AsDashboardPanelItem()
		if err == nil {
			tfPanelIndex := len(panels)
			var tfPanel *models.PanelModel
			if tfPanelIndex < len(m.Panels) {
				tfPanel = &m.Panels[tfPanelIndex]
			}

			panel, d := dashboardMapPanelFromAPI(ctx, m, tfPanel, panelItem)
			diags.Append(d...)
			if diags.HasError() {
				return nil, nil, diags
			}

			panels = append(panels, panel)
		}
	}

	return panels, sections, diags
}

func dashboardMapSectionFromAPI(ctx context.Context, m *models.DashboardModel, tfSection *models.SectionModel, section kbapi.KbnDashboardSection) (models.SectionModel, diag.Diagnostics) {
	collapsed := types.BoolPointerValue(section.Collapsed)
	if tfSection != nil && !typeutils.IsKnown(tfSection.Collapsed) && section.Collapsed != nil && !*section.Collapsed {
		collapsed = types.BoolNull()
	}

	sm := models.SectionModel{
		Title:     types.StringValue(section.Title),
		Collapsed: collapsed,
		ID:        types.StringPointerValue(section.Id),
		Grid: models.SectionGridModel{
			Y: types.Int64Value(int64(section.Grid.Y)),
		},
	}

	// Map section panels
	var diags diag.Diagnostics
	if section.Panels != nil {
		var innerPanels []models.PanelModel
		for _, p := range *section.Panels {
			tfPanelIndex := len(innerPanels)
			var tfPanel *models.PanelModel
			if tfSection != nil && tfPanelIndex < len(tfSection.Panels) {
				tfPanel = &tfSection.Panels[tfPanelIndex]
			}

			pm, d := dashboardMapPanelFromAPI(ctx, m, tfPanel, p)
			diags.Append(d...)
			if diags.HasError() {
				return models.SectionModel{}, diags
			}

			innerPanels = append(innerPanels, pm)
		}
		sm.Panels = innerPanels
	}
	return sm, diags
}

func setPanelGridFromAPI(pm *models.PanelModel, x, y float32, w, h *float32) {
	pm.Grid = models.PanelGridModel{
		X: types.Int64Value(int64(x)),
		Y: types.Int64Value(int64(y)),
	}
	if w != nil {
		pm.Grid.W = types.Int64Value(int64(*w))
	} else {
		pm.Grid.W = types.Int64Null()
	}
	if h != nil {
		pm.Grid.H = types.Int64Value(int64(*h))
	} else {
		pm.Grid.H = types.Int64Null()
	}
}

func panelHasTypedConfig(pm *models.PanelModel) bool {
	return pm.MarkdownConfig != nil ||
		pm.TimeSliderControlConfig != nil ||
		pm.SloBurnRateConfig != nil ||
		pm.SloOverviewConfig != nil ||
		pm.SloErrorBudgetConfig != nil ||
		pm.EsqlControlConfig != nil ||
		pm.OptionsListControlConfig != nil ||
		pm.RangeSliderControlConfig != nil ||
		pm.SyntheticsStatsOverviewConfig != nil ||
		pm.SyntheticsMonitorsConfig != nil ||
		pm.LensDashboardAppConfig != nil ||
		pm.VisConfig != nil ||
		pm.ImageConfig != nil ||
		pm.SloAlertsConfig != nil ||
		pm.DiscoverSessionConfig != nil
}

func panelUsesConfigJSONOnly(pm *models.PanelModel) bool {
	if pm == nil || !typeutils.IsKnown(pm.ConfigJSON) {
		return false
	}
	return !panelHasTypedConfig(pm)
}

func clearPanelConfigBlocks(pm *models.PanelModel) {
	pm.MarkdownConfig = nil
	pm.TimeSliderControlConfig = nil
	pm.SloBurnRateConfig = nil
	pm.SloOverviewConfig = nil
	pm.SloErrorBudgetConfig = nil
	pm.EsqlControlConfig = nil
	pm.OptionsListControlConfig = nil
	pm.RangeSliderControlConfig = nil
	pm.SyntheticsStatsOverviewConfig = nil
	pm.SyntheticsMonitorsConfig = nil
	pm.LensDashboardAppConfig = nil
	pm.VisConfig = nil
	pm.ImageConfig = nil
	pm.SloAlertsConfig = nil
	pm.DiscoverSessionConfig = nil
}

func dashboardMapPanelFromAPI(ctx context.Context, m *models.DashboardModel, tfPanel *models.PanelModel, panelItem kbapi.DashboardPanelItem) (models.PanelModel, diag.Diagnostics) {
	// Start from the existing TF model when available (plan or prior state).
	//
	// Kibana may omit optional attributes on reads even when they were provided on
	// writes. Seeding from the existing model allows individual panel converters
	// to preserve already-known values when the API response doesn't include them.
	var pm models.PanelModel
	if tfPanel != nil {
		pm = *tfPanel
	}

	discriminator, err := panelItem.Discriminator()
	if err != nil {
		return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
	}
	pm.Type = types.StringValue(discriminator)

	var diags diag.Diagnostics
	if h := LookupHandler(discriminator); h != nil {
		diags = h.FromAPI(ctx, &pm, tfPanel, panelItem)
		alignPanelStateFromPlan(ctx, tfPanel, &pm)
		return pm, diags
	}

	switch discriminator {
	case panelTypeVis:
		visPanel, err := panelItem.AsKbnDashboardPanelTypeVis()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, visPanel.Grid.X, visPanel.Grid.Y, visPanel.Grid.W, visPanel.Grid.H)
		pm.ID = types.StringPointerValue(visPanel.Id)

		configBytes, err := visPanel.Config.MarshalJSON()
		if err == nil {
			configJSON := customtypes.NewJSONWithDefaultsValue(string(configBytes), populatePanelConfigJSONDefaults)
			if tfPanel != nil {
				configJSON = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, tfPanel.ConfigJSON, configJSON, &diags)
			}
			pm.ConfigJSON = configJSON
		}

		if panelUsesConfigJSONOnly(tfPanel) {
			break
		}

		var root map[string]any
		if len(configBytes) == 0 || json.Unmarshal(configBytes, &root) != nil {
			break
		}

		visPrior := configPriorForVisRead(tfPanel, &pm)

		switch classifyLensDashboardAppConfigFromRoot(root) {
		case lensConfigClassByReference:
			cfg1, err1 := visPanel.Config.AsKbnDashboardPanelTypeVisConfig1()
			if err1 != nil {
				diags.AddError("Invalid visualization panel configuration on read", err1.Error())
				break
			}
			diags.Append(populateVisByReferenceFromAPI(ctx, visPrior, &pm, cfg1)...)

		case lensConfigClassByValueChart:
			config0, err0 := visPanel.Config.AsKbnDashboardPanelTypeVisConfig0()
			if err0 != nil {
				diags.AddError("Invalid visualization panel configuration on read", err0.Error())
				break
			}
			pm.VisConfig = &models.VisConfigModel{
				ByValue: &models.VisByValueModel{},
			}
			diags.Append(populateLensVisByValueFromTypedChartAPI(ctx, m, tfPanel, &pm.VisConfig.ByValue.LensByValueChartBlocks, config0, true)...)

		default:
			if visPrior != nil && visPrior.ByReference != nil {
				// REQ-009 / D10: ambiguous API shape — preserve prior by_reference (pm seeded from tfPanel).
				break
			}
			config0, err0 := visPanel.Config.AsKbnDashboardPanelTypeVisConfig0()
			if err0 != nil {
				break
			}
			visType := lenscommon.DetectVizType(config0)
			if visType == "" {
				break
			}
			conv := lenscommon.ForType(visType)
			if conv == nil {
				diags.AddError(
					"Unsupported visualization chart type",
					fmt.Sprintf(
						"The dashboard returned Lens visualization discriminator %q which this provider does not support as typed `vis_config.by_value`. "+
							"Use panel-level `config_json` as the escape hatch to manage this panel until support is added.",
						visType,
					),
				)
				break
			}
			pm.VisConfig = &models.VisConfigModel{
				ByValue: &models.VisByValueModel{},
			}
			seedWaffleLensByValueChartFromPriorPanel(&pm.VisConfig.ByValue.LensByValueChartBlocks, tfPanel)
			seedLensChartPriorIntoBlocks(tfPanel, &pm.VisConfig.ByValue.LensByValueChartBlocks, visType)
			diags.Append(conv.PopulateFromAttributes(ctx, lensChartResolver(m), &pm.VisConfig.ByValue.LensByValueChartBlocks, config0)...)
		}
	case panelTypeLensDashboardApp:
		ldPanel, err := panelItem.AsKbnDashboardPanelTypeLensDashboardApp()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, ldPanel.Grid.X, ldPanel.Grid.Y, ldPanel.Grid.W, ldPanel.Grid.H)
		pm.ID = types.StringPointerValue(ldPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		d := populateLensDashboardAppFromAPI(ctx, m, &pm, tfPanel, ldPanel)
		diags.Append(d...)
	case panelTypeDiscoverSession:
		dsPanel, err := panelItem.AsKbnDashboardPanelTypeDiscoverSession()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, dsPanel.Grid.X, dsPanel.Grid.Y, dsPanel.Grid.W, dsPanel.Grid.H)
		pm.ID = types.StringPointerValue(dsPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateDiscoverSessionPanelFromAPI(ctx, &pm, tfPanel, dsPanel)
	default:
		fillUnknownDashboardPanelFromAPI(ctx, tfPanel, &pm, panelItem)
	}

	alignPanelStateFromPlan(ctx, tfPanel, &pm)

	return pm, diags
}

func timeRangeModelToAPI(tr *models.TimeRangeModel) kbapi.KbnEsQueryServerTimeRangeSchema {
	if tr == nil {
		return kbapi.KbnEsQueryServerTimeRangeSchema{}
	}
	out := kbapi.KbnEsQueryServerTimeRangeSchema{
		From: tr.From.ValueString(),
		To:   tr.To.ValueString(),
	}
	if typeutils.IsKnown(tr.Mode) {
		mode := kbapi.KbnEsQueryServerTimeRangeSchemaMode(tr.Mode.ValueString())
		out.Mode = &mode
	}
	return out
}

// resolveChartTimeRange returns the API time_range for a typed Lens chart root: chart-level when set,
// otherwise copied from the dashboard-level time_range (both are required API inputs).
//
// Production dashboard writes (`dashboardPanelsToAPI` / `panelToAPI`) always pass the enclosing
// `models.DashboardModel`, so null chart-level `time_range` inherits dashboard-level values (REQ-013).
//
// The `now-15m` / `now` fallback below applies when there is no chart-level override and either
// no parent `models.DashboardModel` is in scope (e.g. isolated unit tests call `buildAttributes(..., nil)`),
// or `dashboard != nil` but `dashboard.TimeRange == nil` (unusual in production: the dashboard
// schema requires `time_range`). Optional tooling may also construct chart payloads without a parent
// dashboard. The lens-dashboard-app typed `by_value` path threads the parent dashboard via
// `lensDashboardAppToAPI` / `lensDashboardAppByValueToAPI` so it inherits like other typed charts;
// it does not rely on this fallback during normal resource updates.
func resolveChartTimeRange(dashboard *models.DashboardModel, chartLevel *models.TimeRangeModel) kbapi.KbnEsQueryServerTimeRangeSchema {
	if chartLevel != nil {
		return timeRangeModelToAPI(chartLevel)
	}
	if dashboard != nil && dashboard.TimeRange != nil {
		return timeRangeModelToAPI(dashboard.TimeRange)
	}
	return kbapi.KbnEsQueryServerTimeRangeSchema{
		From: "now-15m",
		To:   "now",
	}
}

func dashboardPanelsToAPI(ctx context.Context, m *models.DashboardModel) (*kbapi.DashboardPanels, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.Panels == nil && m.Sections == nil {
		return nil, diags
	}

	apiPanels := make(kbapi.DashboardPanels, 0, len(m.Panels)+len(m.Sections))

	// Process panels
	for _, pm := range m.Panels {
		panelItem, d := panelToAPI(ctx, pm, m)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var item kbapi.DashboardPanels_Item
		err := item.FromDashboardPanelItem(panelItem)
		if err != nil {
			diags.AddError("Failed to create dashboard panel item", err.Error())
		}

		apiPanels = append(apiPanels, item)
	}

	// Process sections
	for _, sm := range m.Sections {
		section := kbapi.KbnDashboardSection{
			Title: sm.Title.ValueString(),
			Grid: struct {
				Y float32 `json:"y"`
			}{
				Y: float32(sm.Grid.Y.ValueInt64()),
			},
		}

		if typeutils.IsKnown(sm.Collapsed) {
			section.Collapsed = new(sm.Collapsed.ValueBool())
		}
		if typeutils.IsKnown(sm.ID) {
			section.Id = new(sm.ID.ValueString())
		}

		if len(sm.Panels) > 0 {
			innerPanels := make([]kbapi.DashboardPanelItem, 0, len(sm.Panels))

			for _, pm := range sm.Panels {
				item, d := panelToAPI(ctx, pm, m)
				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}

				innerPanels = append(innerPanels, item)
			}
			section.Panels = &innerPanels
		}

		var item kbapi.DashboardPanels_Item
		err := item.FromKbnDashboardSection(section)
		if err != nil {
			diags.AddError("Failed to create dashboard section item", err.Error())
		}
		apiPanels = append(apiPanels, item)
	}

	return &apiPanels, diags
}

// panelDispatcherAllowsTypedConfigOmission reports panel types whose handlers may serialize when the Terraform
// *_config block is absent (optional block or API-default empty synthetics payload). Other registered handlers rely
// on HasConfig(...) or legacy error branches instead.
func panelDispatcherAllowsTypedConfigOmission(panelType string) bool {
	switch panelType {
	// Practitioner may omit `markdown_config` when managing the panel purely via panel-level `config_json`.
	case panelTypeMarkdown, panelTypeTimeSlider, panelTypeSyntheticsStatsOverview, panelTypeSyntheticsMonitors:
		return true
	default:
		return false
	}
}

func panelToAPI(ctx context.Context, pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	// Type-first dispatch: when pm.Type is known the registry resolves the handler in O(1).
	// Optional typed *_config blocks (allowlist) still dispatch when the block itself is absent.
	if typeutils.IsKnown(pm.Type) {
		pt := pm.Type.ValueString()
		if h := LookupHandler(pt); h != nil {
			if panelkit.HasConfig(&pm, pt+"_config") || panelDispatcherAllowsTypedConfigOmission(pt) {
				return h.ToAPI(pm, dashboard)
			}
		}
	}

	// Fallback: pm.Type may be unset while a typed *_config block is populated (e.g. legacy plan-state shapes).
	// Scan handlers so writes still resolve to the correct dispatcher.
	for _, h := range AllHandlers() {
		if panelkit.HasConfig(&pm, h.PanelType()+"_config") {
			return h.ToAPI(pm, dashboard)
		}
	}

	var diags diag.Diagnostics

	var dashTR *models.TimeRangeModel
	if dashboard != nil {
		dashTR = dashboard.TimeRange
	}

	grid := struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	}{
		X: float32(pm.Grid.X.ValueInt64()),
		Y: float32(pm.Grid.Y.ValueInt64()),
	}
	if typeutils.IsKnown(pm.Grid.W) {
		w := float32(pm.Grid.W.ValueInt64())
		grid.W = &w
	}
	if typeutils.IsKnown(pm.Grid.H) {
		h := float32(pm.Grid.H.ValueInt64())
		grid.H = &h
	}

	var panelID *string
	if typeutils.IsKnown(pm.ID) {
		panelID = new(pm.ID.ValueString())
	}

	var panelItem kbapi.DashboardPanelItem

	lensGrid := lensDashboardAPIGrid{H: grid.H, W: grid.W, X: grid.X, Y: grid.Y}
	if pm.LensDashboardAppConfig != nil {
		return lensDashboardAppToAPI(pm, lensGrid, panelID, dashboard)
	}
	if pm.VisConfig != nil {
		return visConfigToAPI(pm, dashboard, grid, panelID)
	}
	if pm.Type.ValueString() == panelTypeLensDashboardApp {
		if typeutils.IsKnown(pm.ConfigJSON) && !pm.ConfigJSON.IsNull() {
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `lens-dashboard-app` panels. "+
					"Use the `lens_dashboard_app_config` block with `by_value` or `by_reference` instead.",
			)
			return kbapi.DashboardPanelItem{}, diags
		}
		diags.AddError(
			"Missing `lens_dashboard_app_config`",
			"The `lens_dashboard_app_config` block is required for `lens-dashboard-app` panels.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}

	if pm.DiscoverSessionConfig != nil {
		return discoverSessionPanelToAPI(ctx, pm, grid, panelID, dashTR)
	}

	if pm.Type.ValueString() == panelTypeDiscoverSession {
		if typeutils.IsKnown(pm.ConfigJSON) && !pm.ConfigJSON.IsNull() {
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `discover_session` panels. Use `discover_session_config` instead.",
			)
			return kbapi.DashboardPanelItem{}, diags
		}
		diags.AddError(
			"Missing discover_session panel configuration",
			"Discover session panels require `discover_session_config`.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}

	if pm.Type.ValueString() == panelTypeSloBurnRate {
		if typeutils.IsKnown(pm.ConfigJSON) && !pm.ConfigJSON.IsNull() {
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `slo_burn_rate` panels. Use `slo_burn_rate_config` instead.",
			)
			return kbapi.DashboardPanelItem{}, diags
		}
		diags.AddError(
			"Missing SLO burn rate panel configuration",
			"SLO burn rate panels require `slo_burn_rate_config`.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}

	if pm.Type.ValueString() == panelTypeImage {
		if typeutils.IsKnown(pm.ConfigJSON) && !pm.ConfigJSON.IsNull() {
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `image` panels. Use `image_config` instead.",
			)
			return kbapi.DashboardPanelItem{}, diags
		}
		diags.AddError(
			"Missing image panel configuration",
			"Image panels require `image_config`.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}

	if pm.Type.ValueString() == panelTypeSloAlerts {
		if typeutils.IsKnown(pm.ConfigJSON) && !pm.ConfigJSON.IsNull() {
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `slo_alerts` panels. Use `slo_alerts_config` instead.",
			)
			return kbapi.DashboardPanelItem{}, diags
		}
		diags.AddError(
			"Missing SLO alerts panel configuration",
			"SLO alerts panels require `slo_alerts_config`.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}

	// Practitioner-authored `config_json`: some typed panel discriminators enumerate explicit rejects (example:
	// `slo_overview`, `synthetics_stats_overview`) while others historically fell through to the default branch's raw union
	// reconstruction (examples: `slo_error_budget`, `synthetics_monitors`). That asymmetry existed before dashboard-panel-contract;
	// Openspec dashboard-panel-contract section 5 retires this switch and normalizes behavior across handlers.
	if typeutils.IsKnown(pm.ConfigJSON) {
		configJSON := []byte(pm.ConfigJSON.ValueString())
		switch pm.Type.ValueString() {
		case panelTypeVis:
			var config kbapi.KbnDashboardPanelTypeVis_Config
			if err := config.UnmarshalJSON(configJSON); err != nil {
				diags.AddError("Failed to unmarshal visualization panel config", err.Error())
				return kbapi.DashboardPanelItem{}, diags
			}
			visPanel := kbapi.KbnDashboardPanelTypeVis{
				Config: config,
				Grid:   grid,
				Id:     panelID,
				Type:   kbapi.Vis,
			}
			if err := panelItem.FromKbnDashboardPanelTypeVis(visPanel); err != nil {
				diags.AddError("Failed to create visualization panel", err.Error())
			}
			return panelItem, diags
		case panelTypeSloBurnRate:
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `slo_burn_rate` panels. Use `slo_burn_rate_config` instead.",
			)
		case panelTypeSloOverview:
			diags.AddError(
				"Unsupported panel type for config_json",
				"The slo_overview panel type must be managed through the typed slo_overview_config block, not config_json.",
			)
		case panelTypeSyntheticsStatsOverview:
			diags.AddError(
				"Unsupported panel type for config_json",
				"The synthetics_stats_overview panel type must be managed through the typed synthetics_stats_overview_config block, not config_json.",
			)
		case panelTypeDiscoverSession:
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `discover_session` panels. Use `discover_session_config` instead.",
			)
		case panelTypeTimeSlider:
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `time_slider_control` panels. Use `time_slider_control_config` or omit config.",
			)
		case panelTypeEsqlControl:
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `esql_control` panels. Use `esql_control_config` instead.",
			)
		case panelTypeOptionsListControl:
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `options_list_control` panels. Use `options_list_control_config` instead.",
			)
		case panelTypeRangeSlider:
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `range_slider_control` panels. Use `range_slider_control_config` instead.",
			)
		default:
			// Unknown panel type: reconstruct the full panel JSON from the stored
			// config_json + grid + id + type and set it directly as the raw union.
			fullPanel := map[string]any{
				"type":   pm.Type.ValueString(),
				"grid":   grid,
				"config": json.RawMessage(configJSON),
			}
			if panelID != nil {
				fullPanel["id"] = *panelID
			}
			rawBytes, mErr := json.Marshal(fullPanel)
			if mErr != nil {
				diags.AddError("Failed to marshal unknown panel", mErr.Error())
				return kbapi.DashboardPanelItem{}, diags
			}
			if err := panelItem.UnmarshalJSON(rawBytes); err != nil {
				diags.AddError("Failed to create unknown panel type", err.Error())
				return kbapi.DashboardPanelItem{}, diags
			}
			return panelItem, diags
		}
	}

	// Distinguish between known panel types missing their config block vs
	// truly unknown panel types that have no typed config support.
	panelType := pm.Type.ValueString()
	if !typeutils.IsKnown(pm.Type) {
		// Type is unknown/null; no way to determine intent.
		diags.AddError("Unsupported panel configuration", "No panel configuration block was provided.")
		return kbapi.DashboardPanelItem{}, diags
	}

	switch panelType {
	case panelTypeMarkdown, panelTypeVis, panelTypeTimeSlider, panelTypeSloBurnRate,
		panelTypeSloErrorBudget, panelTypeEsqlControl, panelTypeOptionsListControl,
		panelTypeRangeSlider, panelTypeSyntheticsStatsOverview, panelTypeSyntheticsMonitors,
		panelTypeLensDashboardApp, panelTypeSloOverview, panelTypeDiscoverSession:
		diags.AddError("Unsupported panel configuration", "No panel configuration block was provided.")
	default:
		diags.AddError(
			"Unsupported panel type",
			fmt.Sprintf(
				"Panel type %q is not yet supported. This panel type was preserved from the API during read "+
					"but cannot be authored in configuration. To add support for this panel type, "+
					"wait for a provider update that includes a typed configuration block.",
				panelType,
			),
		)
	}
	return kbapi.DashboardPanelItem{}, diags
}
