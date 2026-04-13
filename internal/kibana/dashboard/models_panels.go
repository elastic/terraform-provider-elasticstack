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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type panelModel struct {
	Type                     types.String                                      `tfsdk:"type"`
	Grid                     panelGridModel                                    `tfsdk:"grid"`
	ID                       types.String                                      `tfsdk:"id"`
	MarkdownConfig           *markdownConfigModel                              `tfsdk:"markdown_config"`
	XYChartConfig            *xyChartConfigModel                               `tfsdk:"xy_chart_config"`
	TreemapConfig            *treemapConfigModel                               `tfsdk:"treemap_config"`
	MosaicConfig             *mosaicConfigModel                                `tfsdk:"mosaic_config"`
	DatatableConfig          *datatableConfigModel                             `tfsdk:"datatable_config"`
	TagcloudConfig           *tagcloudConfigModel                              `tfsdk:"tagcloud_config"`
	MetricChartConfig        *metricChartConfigModel                           `tfsdk:"metric_chart_config"`
	PieChartConfig           *pieChartConfigModel                              `tfsdk:"pie_chart_config"`
	GaugeConfig              *gaugeConfigModel                                 `tfsdk:"gauge_config"`
	LegacyMetricConfig       *legacyMetricConfigModel                          `tfsdk:"legacy_metric_config"`
	RegionMapConfig          *regionMapConfigModel                             `tfsdk:"region_map_config"`
	HeatmapConfig            *heatmapConfigModel                               `tfsdk:"heatmap_config"`
	WaffleConfig             *waffleConfigModel                                `tfsdk:"waffle_config"`
	TimeSliderControlConfig  *timeSliderControlConfigModel                     `tfsdk:"time_slider_control_config"`
	SloBurnRateConfig        *sloBurnRateConfigModel                           `tfsdk:"slo_burn_rate_config"`
	SloOverviewConfig        *sloOverviewConfigModel                           `tfsdk:"slo_overview_config"`
	SloErrorBudgetConfig     *sloErrorBudgetConfigModel                        `tfsdk:"slo_error_budget_config"`
	EsqlControlConfig        *esqlControlConfigModel                           `tfsdk:"esql_control_config"`
	OptionsListControlConfig *optionsListControlConfigModel                    `tfsdk:"options_list_control_config"`
	RangeSliderControlConfig *rangeSliderControlConfigModel                    `tfsdk:"range_slider_control_config"`
	SyntheticsMonitorsConfig *syntheticsMonitorsConfigModel                    `tfsdk:"synthetics_monitors_config"`
	ConfigJSON               customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config_json"`
}

type panelGridModel struct {
	X types.Int64 `tfsdk:"x"`
	Y types.Int64 `tfsdk:"y"`
	W types.Int64 `tfsdk:"w"`
	H types.Int64 `tfsdk:"h"`
}

type sectionModel struct {
	Title     types.String     `tfsdk:"title"`
	ID        types.String     `tfsdk:"id"`
	Collapsed types.Bool       `tfsdk:"collapsed"`
	Grid      sectionGridModel `tfsdk:"grid"`
	Panels    []panelModel     `tfsdk:"panels"`
}

type sectionGridModel struct {
	Y types.Int64 `tfsdk:"y"`
}

var lensVizConverters = []lensVisualizationConverter{
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

func (m *dashboardModel) mapPanelsFromAPI(ctx context.Context, apiPanels *kbapi.DashboardPanels) ([]panelModel, []sectionModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if apiPanels == nil || len(*apiPanels) == 0 {
		return nil, nil, diags
	}

	var panels []panelModel
	var sections []sectionModel

	for _, item := range *apiPanels {
		// Try section first; this avoids treating section items as panels.
		section, err := item.AsKbnDashboardSection()
		if err == nil && section.Title != "" {
			tfSectionIndex := len(sections)
			var tfSection *sectionModel
			if tfSectionIndex < len(m.Sections) {
				tfSection = &m.Sections[tfSectionIndex]
			}
			sectionModel, d := m.mapSectionFromAPI(ctx, tfSection, section)
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
			var tfPanel *panelModel
			if tfPanelIndex < len(m.Panels) {
				tfPanel = &m.Panels[tfPanelIndex]
			}

			panel, d := m.mapPanelFromAPI(ctx, tfPanel, panelItem)
			diags.Append(d...)
			if diags.HasError() {
				return nil, nil, diags
			}

			panels = append(panels, panel)
		}
	}

	return panels, sections, diags
}

func (m *dashboardModel) mapSectionFromAPI(ctx context.Context, tfSection *sectionModel, section kbapi.KbnDashboardSection) (sectionModel, diag.Diagnostics) {
	collapsed := types.BoolPointerValue(section.Collapsed)
	if tfSection != nil && !typeutils.IsKnown(tfSection.Collapsed) && section.Collapsed != nil && !*section.Collapsed {
		collapsed = types.BoolNull()
	}

	sm := sectionModel{
		Title:     types.StringValue(section.Title),
		Collapsed: collapsed,
		ID:        types.StringPointerValue(section.Id),
		Grid: sectionGridModel{
			Y: types.Int64Value(int64(section.Grid.Y)),
		},
	}

	// Map section panels
	var diags diag.Diagnostics
	if section.Panels != nil {
		var innerPanels []panelModel
		for _, p := range *section.Panels {
			tfPanelIndex := len(innerPanels)
			var tfPanel *panelModel
			if tfSection != nil && tfPanelIndex < len(tfSection.Panels) {
				tfPanel = &tfSection.Panels[tfPanelIndex]
			}

			pm, d := m.mapPanelFromAPI(ctx, tfPanel, p)
			diags.Append(d...)
			if diags.HasError() {
				return sectionModel{}, diags
			}

			innerPanels = append(innerPanels, pm)
		}
		sm.Panels = innerPanels
	}
	return sm, diags
}

func setPanelGridFromAPI(pm *panelModel, x, y float32, w, h *float32) {
	pm.Grid = panelGridModel{
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

func panelUsesConfigJSONOnly(pm *panelModel) bool {
	if pm == nil || !typeutils.IsKnown(pm.ConfigJSON) {
		return false
	}
	return pm.MarkdownConfig == nil &&
		pm.XYChartConfig == nil &&
		pm.TreemapConfig == nil &&
		pm.DatatableConfig == nil &&
		pm.TagcloudConfig == nil &&
		pm.MetricChartConfig == nil &&
		pm.PieChartConfig == nil &&
		pm.GaugeConfig == nil &&
		pm.LegacyMetricConfig == nil &&
		pm.RegionMapConfig == nil &&
		pm.HeatmapConfig == nil &&
		pm.WaffleConfig == nil &&
		pm.TimeSliderControlConfig == nil &&
		pm.SloBurnRateConfig == nil &&
		pm.SloOverviewConfig == nil &&
		pm.SloErrorBudgetConfig == nil &&
		pm.EsqlControlConfig == nil &&
		pm.OptionsListControlConfig == nil &&
		pm.RangeSliderControlConfig == nil &&
		pm.SyntheticsMonitorsConfig == nil
}

func (m *dashboardModel) mapPanelFromAPI(ctx context.Context, tfPanel *panelModel, panelItem kbapi.DashboardPanelItem) (panelModel, diag.Diagnostics) {
	// Start from the existing TF model when available (plan or prior state).
	//
	// Kibana may omit optional attributes on reads even when they were provided on
	// writes. Seeding from the existing model allows individual panel converters
	// to preserve already-known values when the API response doesn't include them.
	var pm panelModel
	if tfPanel != nil {
		pm = *tfPanel
	}

	discriminator, err := panelItem.Discriminator()
	if err != nil {
		return panelModel{}, diagutil.FrameworkDiagFromError(err)
	}
	pm.Type = types.StringValue(discriminator)

	var diags diag.Diagnostics
	switch discriminator {
	case panelTypeMarkdown:
		markdownPanel, err := panelItem.AsKbnDashboardPanelTypeMarkdown()
		if err != nil {
			return panelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, markdownPanel.Grid.X, markdownPanel.Grid.Y, markdownPanel.Grid.W, markdownPanel.Grid.H)
		pm.ID = types.StringPointerValue(markdownPanel.Id)
		if markdownPanel.Config != nil {
			if !panelUsesConfigJSONOnly(tfPanel) {
				config0, err := markdownPanel.Config.AsKbnDashboardPanelTypeMarkdownConfig0()
				if err != nil {
					// Kibana may return inline markdown fields without the union discriminator
					// expected by AsKbnDashboardPanelTypeMarkdownConfig0; fall back to unmarshalling
					// the raw config JSON into the inline schema.
					if b, mErr := markdownPanel.Config.MarshalJSON(); mErr == nil {
						var inline kbapi.KbnDashboardPanelTypeMarkdownConfig0
						if json.Unmarshal(b, &inline) == nil {
							config0 = inline
							err = nil
						}
					}
				}
				if err == nil {
					populateMarkdownFromAPI(&pm, config0)
				}
			}
			configBytes, err := markdownPanel.Config.MarshalJSON()
			if err == nil {
				pm.ConfigJSON = customtypes.NewJSONWithDefaultsValue(string(configBytes), populatePanelConfigJSONDefaults)
			}
		}
	case panelTypeSloOverview:
		sloPanel, err := panelItem.AsKbnDashboardPanelTypeSloOverview()
		if err != nil {
			return panelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, sloPanel.Grid.X, sloPanel.Grid.Y, sloPanel.Grid.W, sloPanel.Grid.H)
		pm.ID = types.StringPointerValue(sloPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		d := sloOverviewFromAPI(&pm, tfPanel, sloPanel)
		diags.Append(d...)
	case panelTypeTimeSlider:
		tsPanel, err := panelItem.AsKbnDashboardPanelTypeTimeSliderControl()
		if err != nil {
			return panelModel{}, diagutil.FrameworkDiagFromError(err)
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
			return panelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, sbrPanel.Grid.X, sbrPanel.Grid.Y, sbrPanel.Grid.W, sbrPanel.Grid.H)
		pm.ID = types.StringPointerValue(sbrPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateSloBurnRateFromAPI(&pm, tfPanel, sbrPanel.Config)
	case panelTypeEsqlControl:
		esqlPanel, err := panelItem.AsKbnDashboardPanelTypeEsqlControl()
		if err != nil {
			return panelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, esqlPanel.Grid.X, esqlPanel.Grid.Y, esqlPanel.Grid.W, esqlPanel.Grid.H)
		pm.ID = types.StringPointerValue(esqlPanel.Id)
		// ES|QL control panels are managed via esql_control_config; config_json remains unset.
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateEsqlControlFromAPI(&pm, tfPanel, esqlPanel.Config)
	case panelTypeOptionsListControl:
		olPanel, err := panelItem.AsKbnDashboardPanelTypeOptionsListControl()
		if err != nil {
			return panelModel{}, diagutil.FrameworkDiagFromError(err)
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
			return panelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, rsPanel.Grid.X, rsPanel.Grid.Y, rsPanel.Grid.W, rsPanel.Grid.H)
		pm.ID = types.StringPointerValue(rsPanel.Id)
		// Range slider control panels are managed via range_slider_control_config; config_json remains unset.
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateRangeSliderControlFromAPI(ctx, &pm, tfPanel, &rsPanel)
	case panelTypeVis:
		visPanel, err := panelItem.AsKbnDashboardPanelTypeVis()
		if err != nil {
			return panelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, visPanel.Grid.X, visPanel.Grid.Y, visPanel.Grid.W, visPanel.Grid.H)
		pm.ID = types.StringPointerValue(visPanel.Id)

		configBytes, err := visPanel.Config.MarshalJSON()
		if err == nil {
			pm.ConfigJSON = customtypes.NewJSONWithDefaultsValue(string(configBytes), populatePanelConfigJSONDefaults)
		}

		config0, err := visPanel.Config.AsKbnDashboardPanelTypeVisConfig0()
		if err == nil && !panelUsesConfigJSONOnly(tfPanel) {
			vizType := detectLensVizType(config0)
			for _, converter := range lensVizConverters {
				if converter.vizType() != vizType {
					continue
				}
				if tfPanel != nil && !converter.handlesTFConfig(*tfPanel) {
					continue
				}

				d := converter.populateFromAttributes(ctx, &pm, config0)
				diags.Append(d...)
				break
			}
		}
	case panelTypeSloErrorBudget:
		sebPanel, err := panelItem.AsKbnDashboardPanelTypeSloErrorBudget()
		if err != nil {
			return panelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, sebPanel.Grid.X, sebPanel.Grid.Y, sebPanel.Grid.W, sebPanel.Grid.H)
		pm.ID = types.StringPointerValue(sebPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateSloErrorBudgetFromAPI(&pm, tfPanel, sebPanel.Config)
	case panelTypeSyntheticsMonitors:
		smPanel, err := panelItem.AsKbnDashboardPanelTypeSyntheticsMonitors()
		if err != nil {
			return panelModel{}, diagutil.FrameworkDiagFromError(err)
		}
		setPanelGridFromAPI(&pm, smPanel.Grid.X, smPanel.Grid.Y, smPanel.Grid.W, smPanel.Grid.H)
		pm.ID = types.StringPointerValue(smPanel.Id)
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
		populateSyntheticsMonitorsFromAPI(&pm, tfPanel, smPanel)
	default:
		// No typed mapping yet; keep only the panel type.
		pm.ID = types.StringNull()
		pm.Grid = panelGridModel{
			X: types.Int64Null(),
			Y: types.Int64Null(),
			W: types.Int64Null(),
			H: types.Int64Null(),
		}
	}

	alignPanelStateFromPlan(tfPanel, &pm)

	return pm, diags
}

func lensPanelTimeRange() kbapi.KbnEsQueryServerTimeRangeSchema {
	return kbapi.KbnEsQueryServerTimeRangeSchema{
		From: "now-15m",
		To:   "now",
	}
}

func (m *dashboardModel) panelsToAPI() (*kbapi.DashboardPanels, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.Panels == nil && m.Sections == nil {
		return nil, diags
	}

	apiPanels := make(kbapi.DashboardPanels, 0, len(m.Panels)+len(m.Sections))

	// Process panels
	for _, pm := range m.Panels {
		panelItem, d := pm.toAPI()
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
				item, d := pm.toAPI()
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

func (pm panelModel) toAPI() (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics

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
		config0 := buildMarkdownConfig(pm)
		var config kbapi.KbnDashboardPanelTypeMarkdown_Config
		if err := config.FromKbnDashboardPanelTypeMarkdownConfig0(config0); err != nil {
			return kbapi.DashboardPanelItem{}, diagutil.FrameworkDiagFromError(err)
		}
		markdownPanel := kbapi.KbnDashboardPanelTypeMarkdown{
			Config: &config,
			Grid:   grid,
			Id:     panelID,
		}
		if err := panelItem.FromKbnDashboardPanelTypeMarkdown(markdownPanel); err != nil {
			diags.AddError("Failed to create markdown panel", err.Error())
		}
		return panelItem, diags
	}

	if pm.SloOverviewConfig != nil {
		return sloOverviewToAPI(pm, grid, panelID)
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

	for _, converter := range lensVizConverters {
		if !converter.handlesTFConfig(pm) {
			continue
		}

		config0, d := converter.buildAttributes(pm)
		diags.Append(d...)
		if diags.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}

		var config kbapi.KbnDashboardPanelTypeVis_Config
		if err := config.FromKbnDashboardPanelTypeVisConfig0(config0); err != nil {
			diags.AddError("Failed to create visualization panel config", err.Error())
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
				Config: &config,
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
		default:
			diags.AddError(
				"Unsupported panel type for config_json",
				"Only markdown and vis panel types are currently supported with config_json. "+
					"The esql_control panel type must be managed using the esql_control_config block. "+
					"The synthetics_monitors panel type must be managed using the synthetics_monitors_config block.",
			)
			return kbapi.DashboardPanelItem{}, diags
		}
	}

	diags.AddError("Unsupported panel configuration", "No panel configuration block was provided.")
	return kbapi.DashboardPanelItem{}, diags
}
