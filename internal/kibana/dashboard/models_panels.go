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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var lensVisConverters = []lensVisualizationConverter{
	newXYChartPanelConfigConverter(),
	newTreemapPanelConfigConverter(),
	newMosaicPanelConfigConverter(),
	newDatatablePanelConfigConverter(),
	newTagcloudPanelConfigConverter(),
	newHeatmapPanelConfigConverter(),
	newRegionMapPanelConfigConverter(),
	newLegacyMetricPanelConfigConverter(),
	newGaugePanelConfigConverter(),
	newMetricChartPanelConfigConverter(),
	newPieChartPanelConfigConverter(),
	newWafflePanelConfigConverter(),
}

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
	switch discriminator {
	case panelTypeMarkdown:
		markdownPanel, err := panelItem.AsKbnDashboardPanelTypeMarkdown()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, markdownPanel.Grid.X, markdownPanel.Grid.Y, markdownPanel.Grid.W, markdownPanel.Grid.H)
		pm.ID = types.StringPointerValue(markdownPanel.Id)
		if !panelUsesConfigJSONOnly(tfPanel) {
			rawConfig, rawErr := markdownPanel.Config.MarshalJSON()
			branch := markdownConfigBranchUnknown
			if rawErr != nil {
				diags.AddWarning(
					"Markdown panel configuration",
					fmt.Sprintf(
						"Could not marshal panel config for markdown branch classification: %v. Using union decode fallback.",
						rawErr,
					),
				)
			} else {
				var err error
				branch, err = classifyMarkdownConfigFromRoot(rawConfig)
				if err != nil {
					diags.AddWarning(
						"Markdown panel configuration",
						fmt.Sprintf(
							"Could not parse panel config JSON for markdown branch classification: %v. Using union decode fallback.",
							err,
						),
					)
					branch = markdownConfigBranchUnknown
				}
			}

			decodeMarkdownFails := func() {
				diags.AddError(
					"Invalid markdown panel config",
					"Could not decode markdown panel config as by-value or by-reference.",
				)
			}

			switch branch {
			case markdownConfigBranchByReference:
				if populateMarkdownFromAPIAttemptByReference(&pm, tfPanel, markdownPanel.Config, true) {
					break
				}
				if !populateMarkdownFromAPIAttemptByValue(&pm, tfPanel, markdownPanel.Config, true) {
					decodeMarkdownFails()
				}
			case markdownConfigBranchByValue:
				if populateMarkdownFromAPIAttemptByValue(&pm, tfPanel, markdownPanel.Config, true) {
					break
				}
				if !populateMarkdownFromAPIAttemptByReference(&pm, tfPanel, markdownPanel.Config, true) {
					decodeMarkdownFails()
				}
			default:
				if populateMarkdownFromAPIAttemptByValue(&pm, tfPanel, markdownPanel.Config, false) {
					break
				}
				if !populateMarkdownFromAPIAttemptByReference(&pm, tfPanel, markdownPanel.Config, false) {
					decodeMarkdownFails()
				}
			}
		}
		configBytes, err := markdownPanel.Config.MarshalJSON()
		if err == nil {
			configJSON := customtypes.NewJSONWithDefaultsValue(string(configBytes), populatePanelConfigJSONDefaults)
			if tfPanel != nil {
				configJSON = preservePriorJSONWithDefaultsIfEquivalent(ctx, tfPanel.ConfigJSON, configJSON, &diags)
			}
			pm.ConfigJSON = configJSON
		}
	case panelTypeSloOverview:
		sloPanel, err := panelItem.AsKbnDashboardPanelTypeSloOverview()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, sloPanel.Grid.X, sloPanel.Grid.Y, sloPanel.Grid.W, sloPanel.Grid.H)
		pm.ID = types.StringPointerValue(sloPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		d := sloOverviewFromAPI(&pm, tfPanel, sloPanel)
		diags.Append(d...)
	case panelTypeTimeSlider:
		tsPanel, err := panelItem.AsKbnDashboardPanelTypeTimeSliderControl()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, tsPanel.Grid.X, tsPanel.Grid.Y, tsPanel.Grid.W, tsPanel.Grid.H)
		pm.ID = types.StringPointerValue(tsPanel.Id)
		// Computed read-back only: practitioner-authored config_json is not supported for
		// time_slider_control (see `config_json` type-allowlist validators on the panel schema).
		if configBytes, err := json.Marshal(tsPanel.Config); err == nil {
			pm.ConfigJSON = customtypes.NewJSONWithDefaultsValue(string(configBytes), populatePanelConfigJSONDefaults)
		}
		populateTimeSliderControlFromAPI(&pm, tfPanel, tsPanel.Config)
	case panelTypeSloBurnRate:
		sbrPanel, err := panelItem.AsKbnDashboardPanelTypeSloBurnRate()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, sbrPanel.Grid.X, sbrPanel.Grid.Y, sbrPanel.Grid.W, sbrPanel.Grid.H)
		pm.ID = types.StringPointerValue(sbrPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateSloBurnRateFromAPI(&pm, tfPanel, sbrPanel.Config)
	case panelTypeEsqlControl:
		esqlPanel, err := panelItem.AsKbnDashboardPanelTypeEsqlControl()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, esqlPanel.Grid.X, esqlPanel.Grid.Y, esqlPanel.Grid.W, esqlPanel.Grid.H)
		pm.ID = types.StringPointerValue(esqlPanel.Id)
		// ES|QL control panels are managed via esql_control_config; config_json remains unset.
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateEsqlControlFromAPI(&pm, tfPanel, esqlPanel.Config)
	case panelTypeOptionsListControl:
		olPanel, err := panelItem.AsKbnDashboardPanelTypeOptionsListControl()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, olPanel.Grid.X, olPanel.Grid.Y, olPanel.Grid.W, olPanel.Grid.H)
		pm.ID = types.StringPointerValue(olPanel.Id)
		if configBytes, err := json.Marshal(olPanel.Config); err == nil {
			pm.ConfigJSON = customtypes.NewJSONWithDefaultsValue(string(configBytes), populatePanelConfigJSONDefaults)
		}
		populateOptionsListControlFromAPI(&pm, tfPanel, &olPanel)
	case panelTypeRangeSlider:
		rsPanel, err := panelItem.AsKbnDashboardPanelTypeRangeSliderControl()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, rsPanel.Grid.X, rsPanel.Grid.Y, rsPanel.Grid.W, rsPanel.Grid.H)
		pm.ID = types.StringPointerValue(rsPanel.Id)
		// Range slider control panels are managed via range_slider_control_config; config_json remains unset.
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateRangeSliderControlFromAPI(ctx, &pm, tfPanel, &rsPanel)
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
				configJSON = preservePriorJSONWithDefaultsIfEquivalent(ctx, tfPanel.ConfigJSON, configJSON, &diags)
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
			visType := detectLensVisType(config0)
			if visType == "" {
				diags.AddError(
					"Unsupported visualization chart type",
					"The `vis` panel config has a top-level chart discriminator but could not resolve a Lens chart kind from the union; use panel-level `config_json` until this shape is modeled.",
				)
				break
			}
			converter := lensVisConverterForType(visType)
			if converter == nil {
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
			d := converter.populateFromAttributes(ctx, m, tfPanel, &pm.VisConfig.ByValue.LensByValueChartBlocks, config0)
			diags.Append(d...)

		default:
			if visPrior != nil && visPrior.ByReference != nil {
				// REQ-009 / D10: ambiguous API shape — preserve prior by_reference (pm seeded from tfPanel).
				break
			}
			config0, err0 := visPanel.Config.AsKbnDashboardPanelTypeVisConfig0()
			if err0 != nil {
				break
			}
			visType := detectLensVisType(config0)
			if visType == "" {
				break
			}
			converter := lensVisConverterForType(visType)
			if converter == nil {
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
			d := converter.populateFromAttributes(ctx, m, tfPanel, &pm.VisConfig.ByValue.LensByValueChartBlocks, config0)
			diags.Append(d...)
		}
	case panelTypeSloErrorBudget:
		sebPanel, err := panelItem.AsKbnDashboardPanelTypeSloErrorBudget()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, sebPanel.Grid.X, sebPanel.Grid.Y, sebPanel.Grid.W, sebPanel.Grid.H)
		pm.ID = types.StringPointerValue(sebPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateSloErrorBudgetFromAPI(&pm, tfPanel, sebPanel.Config)
	case panelTypeSyntheticsStatsOverview:
		ssoPanel, err := panelItem.AsKbnDashboardPanelTypeSyntheticsStatsOverview()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, ssoPanel.Grid.X, ssoPanel.Grid.Y, ssoPanel.Grid.W, ssoPanel.Grid.H)
		pm.ID = types.StringPointerValue(ssoPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateSyntheticsStatsOverviewFromAPI(&pm, tfPanel, ssoPanel)
	case panelTypeSyntheticsMonitors:
		smPanel, err := panelItem.AsKbnDashboardPanelTypeSyntheticsMonitors()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, smPanel.Grid.X, smPanel.Grid.Y, smPanel.Grid.W, smPanel.Grid.H)
		pm.ID = types.StringPointerValue(smPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateSyntheticsMonitorsFromAPI(&pm, tfPanel, smPanel)
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
	case panelTypeImage:
		imgPanel, err := panelItem.AsKbnDashboardPanelTypeImage()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, imgPanel.Grid.X, imgPanel.Grid.Y, imgPanel.Grid.W, imgPanel.Grid.H)
		pm.ID = types.StringPointerValue(imgPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateImagePanelFromAPI(&pm, tfPanel, imgPanel)
	case panelTypeSloAlerts:
		saPanel, err := panelItem.AsKbnDashboardPanelTypeSloAlerts()
		if err != nil {
			return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, saPanel.Grid.X, saPanel.Grid.Y, saPanel.Grid.W, saPanel.Grid.H)
		pm.ID = types.StringPointerValue(saPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateSloAlertsPanelFromAPI(&pm, tfPanel, saPanel)
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
		// Round-trip stability for panel types without a typed config block.
		pm.ID = types.StringNull()
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		pm.Grid = models.PanelGridModel{}
		rawBytes, err := panelItem.MarshalJSON()
		if err == nil {
			var rawObj map[string]any
			if err := json.Unmarshal(rawBytes, &rawObj); err == nil {
				if grid, ok := rawObj["grid"].(map[string]any); ok {
					x, _ := grid["x"].(float64)
					y, _ := grid["y"].(float64)
					var wPtr, hPtr *float32
					if wVal, ok := grid["w"].(float64); ok {
						wPtr = typeutils.Float32Ptr(wVal)
					}
					if hVal, ok := grid["h"].(float64); ok {
						hPtr = typeutils.Float32Ptr(hVal)
					}
					setPanelGridFromAPI(&pm, float32(x), float32(y), wPtr, hPtr)
				}
				if id, ok := rawObj["id"].(string); ok && id != "" {
					pm.ID = types.StringValue(id)
				}
				if config, ok := rawObj["config"]; ok {
					configBytes, mErr := json.Marshal(config)
					if mErr == nil {
						pm.ConfigJSON = customtypes.NewJSONWithDefaultsValue(string(configBytes), populatePanelConfigJSONDefaults)
					}
				}
			}
		}
		clearPanelConfigBlocks(&pm)
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

func panelToAPI(ctx context.Context, pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
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
	if pm.MarkdownConfig != nil {
		switch {
		case pm.MarkdownConfig.ByReference != nil:
			config1 := buildMarkdownConfigByReference(pm)
			var config kbapi.KbnDashboardPanelTypeMarkdown_Config
			if err := config.FromKbnDashboardPanelTypeMarkdownConfig1(config1); err != nil {
				return kbapi.DashboardPanelItem{}, diagutil.FrameworkDiagFromError(err)
			}
			markdownPanel := kbapi.KbnDashboardPanelTypeMarkdown{
				Config: config,
				Grid:   grid,
				Id:     panelID,
			}
			if err := panelItem.FromKbnDashboardPanelTypeMarkdown(markdownPanel); err != nil {
				diags.AddError("Failed to create markdown panel", err.Error())
			}
			return panelItem, diags
		case pm.MarkdownConfig.ByValue != nil:
			config0 := buildMarkdownConfig(pm)
			var config kbapi.KbnDashboardPanelTypeMarkdown_Config
			if err := config.FromKbnDashboardPanelTypeMarkdownConfig0(config0); err != nil {
				return kbapi.DashboardPanelItem{}, diagutil.FrameworkDiagFromError(err)
			}
			markdownPanel := kbapi.KbnDashboardPanelTypeMarkdown{
				Config: config,
				Grid:   grid,
				Id:     panelID,
			}
			if err := panelItem.FromKbnDashboardPanelTypeMarkdown(markdownPanel); err != nil {
				diags.AddError("Failed to create markdown panel", err.Error())
			}
			return panelItem, diags
		default:
			diags.AddError(
				"Invalid markdown_config",
				"Set `markdown_config.by_value` or `markdown_config.by_reference` (exactly one).",
			)
			return kbapi.DashboardPanelItem{}, diags
		}
	}

	if pm.SloOverviewConfig != nil {
		return sloOverviewToAPI(pm, grid, panelID)
	}

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

	if pm.ImageConfig != nil {
		return imagePanelToAPI(pm, grid, panelID)
	}

	if pm.SloAlertsConfig != nil {
		return sloAlertsPanelToAPI(pm, grid, panelID)
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

	if pm.Type.ValueString() == panelTypeRangeSlider || pm.RangeSliderControlConfig != nil {
		if pm.RangeSliderControlConfig == nil {
			diags.AddError(
				"Missing range slider control panel configuration",
				"Range slider control panels require `range_slider_control_config`.",
			)
			return kbapi.DashboardPanelItem{}, diags
		}
		rsPanel := kbapi.KbnDashboardPanelTypeRangeSliderControl{
			Grid: grid,
			Id:   panelID,
		}
		buildRangeSliderControlConfig(pm, &rsPanel)
		if err := panelItem.FromKbnDashboardPanelTypeRangeSliderControl(rsPanel); err != nil {
			diags.AddError("Failed to create range slider control panel", err.Error())
		}
		return panelItem, diags
	}

	if pm.Type.ValueString() == panelTypeTimeSlider || pm.TimeSliderControlConfig != nil {
		tsPanel := kbapi.KbnDashboardPanelTypeTimeSliderControl{
			Grid: grid,
			Id:   panelID,
			Config: struct {
				EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
				IsAnchored                 *bool    `json:"is_anchored,omitempty"`
				StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
			}{},
		}
		buildTimeSliderControlConfig(pm, &tsPanel)
		if err := panelItem.FromKbnDashboardPanelTypeTimeSliderControl(tsPanel); err != nil {
			diags.AddError("Failed to create time slider control panel", err.Error())
		}
		return panelItem, diags
	}

	if pm.Type.ValueString() == panelTypeSloBurnRate || pm.SloBurnRateConfig != nil {
		if pm.SloBurnRateConfig == nil {
			diags.AddError(
				"Missing SLO burn rate panel configuration",
				"SLO burn rate panels require `slo_burn_rate_config`.",
			)
			return kbapi.DashboardPanelItem{}, diags
		}
		sbrPanel := kbapi.KbnDashboardPanelTypeSloBurnRate{
			Grid: grid,
			Id:   panelID,
		}
		buildSloBurnRateConfig(pm, &sbrPanel)
		if err := panelItem.FromKbnDashboardPanelTypeSloBurnRate(sbrPanel); err != nil {
			diags.AddError("Failed to create SLO burn rate panel", err.Error())
		}
		return panelItem, diags
	}

	if pm.SloErrorBudgetConfig != nil {
		sebPanel := kbapi.KbnDashboardPanelTypeSloErrorBudget{
			Grid: grid,
			Id:   panelID,
		}
		buildSloErrorBudgetConfig(pm, &sebPanel)
		if err := panelItem.FromKbnDashboardPanelTypeSloErrorBudget(sebPanel); err != nil {
			diags.AddError("Failed to create SLO error budget panel", err.Error())
		}
		return panelItem, diags
	}

	if pm.SyntheticsStatsOverviewConfig != nil {
		ssoPanel := kbapi.KbnDashboardPanelTypeSyntheticsStatsOverview{
			Grid: grid,
			Id:   panelID,
		}
		buildSyntheticsStatsOverviewConfig(pm, &ssoPanel)
		if err := panelItem.FromKbnDashboardPanelTypeSyntheticsStatsOverview(ssoPanel); err != nil {
			diags.AddError("Failed to create synthetics stats overview panel", err.Error())
		}
		return panelItem, diags
	}

	if pm.Type.ValueString() == panelTypeSyntheticsStatsOverview {
		// Panel type is synthetics_stats_overview with no config block: send empty config.
		ssoPanel := kbapi.KbnDashboardPanelTypeSyntheticsStatsOverview{
			Grid: grid,
			Id:   panelID,
		}
		if err := panelItem.FromKbnDashboardPanelTypeSyntheticsStatsOverview(ssoPanel); err != nil {
			diags.AddError("Failed to create synthetics stats overview panel", err.Error())
		}
		return panelItem, diags
	}
	if pm.Type.ValueString() == panelTypeSyntheticsMonitors || pm.SyntheticsMonitorsConfig != nil {
		smPanel := buildSyntheticsMonitorsPanel(pm, grid, panelID)
		if err := panelItem.FromKbnDashboardPanelTypeSyntheticsMonitors(smPanel); err != nil {
			diags.AddError("Failed to create synthetics monitors panel", err.Error())
		}
		return panelItem, diags
	}

	if pm.EsqlControlConfig != nil {
		esqlPanel := kbapi.KbnDashboardPanelTypeEsqlControl{
			Grid: grid,
			Id:   panelID,
		}
		diags.Append(buildEsqlControlConfig(pm, &esqlPanel)...)
		if diags.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		if err := panelItem.FromKbnDashboardPanelTypeEsqlControl(esqlPanel); err != nil {
			diags.AddError("Failed to create esql control panel", err.Error())
		}
		return panelItem, diags
	}

	if pm.Type.ValueString() == panelTypeOptionsListControl || pm.OptionsListControlConfig != nil {
		olPanel := kbapi.KbnDashboardPanelTypeOptionsListControl{
			Grid: grid,
			Id:   panelID,
		}
		buildOptionsListControlConfig(pm, &olPanel)
		if err := panelItem.FromKbnDashboardPanelTypeOptionsListControl(olPanel); err != nil {
			diags.AddError("Failed to create options list control panel", err.Error())
		}
		return panelItem, diags
	}

	if typeutils.IsKnown(pm.ConfigJSON) {
		configJSON := []byte(pm.ConfigJSON.ValueString())
		switch pm.Type.ValueString() {
		case panelTypeMarkdown:
			var config kbapi.KbnDashboardPanelTypeMarkdown_Config
			if err := config.UnmarshalJSON(configJSON); err != nil {
				diags.AddError("Failed to unmarshal markdown panel config", err.Error())
				return kbapi.DashboardPanelItem{}, diags
			}
			markdownPanel := kbapi.KbnDashboardPanelTypeMarkdown{
				Config: config,
				Grid:   grid,
				Id:     panelID,
			}
			if err := panelItem.FromKbnDashboardPanelTypeMarkdown(markdownPanel); err != nil {
				diags.AddError("Failed to create markdown panel", err.Error())
			}
			return panelItem, diags
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
		case panelTypeImage:
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `image` panels. Use `image_config` instead.",
			)
		case panelTypeSloAlerts:
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `slo_alerts` panels. Use `slo_alerts_config` instead.",
			)
		case panelTypeDiscoverSession:
			diags.AddError(
				"Unsupported panel type for config_json",
				"Panel-level `config_json` is not supported for `discover_session` panels. Use `discover_session_config` instead.",
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
		panelTypeLensDashboardApp, panelTypeSloOverview, panelTypeImage, panelTypeSloAlerts, panelTypeDiscoverSession:
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
