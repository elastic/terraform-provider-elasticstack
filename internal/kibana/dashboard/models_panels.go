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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type panelModel struct {
	Type               types.String             `tfsdk:"type"`
	Grid               panelGridModel           `tfsdk:"grid"`
	ID                 types.String             `tfsdk:"id"`
	MarkdownConfig     *markdownConfigModel     `tfsdk:"markdown_config"`
	XYChartConfig      *xyChartConfigModel      `tfsdk:"xy_chart_config"`
	TreemapConfig      *treemapConfigModel      `tfsdk:"treemap_config"`
	DatatableConfig    *datatableConfigModel    `tfsdk:"datatable_config"`
	TagcloudConfig     *tagcloudConfigModel     `tfsdk:"tagcloud_config"`
	MetricChartConfig  *metricChartConfigModel  `tfsdk:"metric_chart_config"`
	PieChartConfig     *pieChartConfigModel     `tfsdk:"pie_chart_config"`
	GaugeConfig        *gaugeConfigModel        `tfsdk:"gauge_config"`
	LegacyMetricConfig *legacyMetricConfigModel `tfsdk:"legacy_metric_config"`
	RegionMapConfig    *regionMapConfigModel    `tfsdk:"region_map_config"`
	HeatmapConfig      *heatmapConfigModel      `tfsdk:"heatmap_config"`
	ConfigJSON         jsontypes.Normalized     `tfsdk:"config_json"`
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

type panelConfigConverter interface {
	handlesAPIPanelConfig(ctx *panelModel, panelType string, config apiPanelConfig) bool
	handlesTFPanelConfig(tfModel panelModel) bool
	populateFromAPIPanel(ctx context.Context, tfModel *panelModel, config apiPanelConfig) diag.Diagnostics
	mapPanelToAPI(tfModel panelModel, config *apiPanelConfig) diag.Diagnostics
}

var panelConfigConverters = []panelConfigConverter{
	markdownPanelConfigConverter{},
	newXYChartPanelConfigConverter(),
	newTreemapPanelConfigConverter(),
	newDatatablePanelConfigConverter(),
	newTagcloudPanelConfigConverter(),
	newHeatmapPanelConfigConverter(),
	newRegionMapPanelConfigConverter(),
	newLegacyMetricPanelConfigConverter(),
	newGaugePanelConfigConverter(),
	newMetricChartPanelConfigConverter(),
	newPieChartPanelConfigConverter(),
}

func (m *dashboardModel) mapPanelsFromAPI(ctx context.Context, apiPanels *kbapi.DashboardPanels) ([]panelModel, []sectionModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if apiPanels == nil || len(*apiPanels) == 0 {
		return nil, nil, diags
	}

	var panels []panelModel
	var sections []sectionModel

	for _, item := range *apiPanels {
		rawItem, err := decodeJSONMap(item)
		if err != nil {
			diags.Append(diagutil.FrameworkDiagFromError(err)...)
			return nil, nil, diags
		}

		if _, isPanel := rawItem["type"]; isPanel {
			panelType, gridX, gridY, gridW, gridH, uid, config, panelDiags := convertRawPanel(rawItem)
			diags.Append(panelDiags...)
			if diags.HasError() {
				return nil, nil, diags
			}

			tfPanelIndex := len(panels)
			var tfPanel *panelModel
			if tfPanelIndex < len(m.Panels) {
				tfPanel = &m.Panels[tfPanelIndex]
			}

			panel, d := m.mapPanelFromAPI(ctx, tfPanel, panelType, gridX, gridY, gridW, gridH, uid, config)
			diags.Append(d...)
			if diags.HasError() {
				return nil, nil, diags
			}

			panels = append(panels, panel)
			continue
		}

		// Try to handle as DashboardPanelSection.
		section, err := item.AsDashboardPanelSection()
		if err == nil {
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
		}
	}

	return panels, sections, diags
}

func (m *dashboardModel) mapSectionFromAPI(ctx context.Context, tfSection *sectionModel, section kbapi.DashboardPanelSection) (sectionModel, diag.Diagnostics) {
	sm := sectionModel{
		Title:     types.StringValue(section.Title),
		Collapsed: types.BoolPointerValue(section.Collapsed),
		ID:        types.StringPointerValue(section.Uid),
		Grid: sectionGridModel{
			Y: types.Int64Value(int64(section.Grid.Y)),
		},
	}

	// Map section panels
	var diags diag.Diagnostics
	if section.Panels != nil {
		var innerPanels []panelModel
		for _, p := range *section.Panels {
			rawPanel, err := decodeJSONMap(p)
			if err != nil {
				diags.Append(diagutil.FrameworkDiagFromError(err)...)
				return sectionModel{}, diags
			}
			panelType, gridX, gridY, gridW, gridH, uid, config, panelDiags := convertRawPanel(rawPanel)
			diags.Append(panelDiags...)
			if diags.HasError() {
				return sectionModel{}, diags
			}

			tfPanelIndex := len(innerPanels)
			var tfPanel *panelModel
			if tfSection != nil && tfPanelIndex < len(tfSection.Panels) {
				tfPanel = &tfSection.Panels[tfPanelIndex]
			}

			pm, d := m.mapPanelFromAPI(ctx, tfPanel, panelType, gridX, gridY, gridW, gridH, uid, config)
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

func (m *dashboardModel) mapPanelFromAPI(ctx context.Context, tfPanel *panelModel, panelType string, gridX, gridY float32, gridW, gridH *float32, uid *string, config apiPanelConfig) (panelModel, diag.Diagnostics) {
	// Start from the existing TF model when available (plan or prior state).
	//
	// Kibana may omit optional attributes on reads even when they were provided on
	// writes. Seeding from the existing model allows individual panel converters
	// to preserve already-known values when the API response doesn't include them.
	var pm panelModel
	if tfPanel != nil {
		pm = *tfPanel
	}

	pm.Type = types.StringValue(panelType)
	pm.Grid = panelGridModel{
		X: types.Int64Value(int64(gridX)),
		Y: types.Int64Value(int64(gridY)),
	}
	if gridW != nil {
		pm.Grid.W = types.Int64Value(int64(*gridW))
	} else {
		pm.Grid.W = types.Int64Null()
	}
	if gridH != nil {
		pm.Grid.H = types.Int64Value(int64(*gridH))
	} else {
		pm.Grid.H = types.Int64Null()
	}

	if uid != nil {
		pm.ID = types.StringValue(*uid)
	} else {
		pm.ID = types.StringNull()
	}

	var diags diag.Diagnostics
	var panelConfigHandled bool
	for _, converter := range panelConfigConverters {
		if converter.handlesAPIPanelConfig(tfPanel, panelType, config) {
			d := converter.populateFromAPIPanel(ctx, &pm, config)
			diags.Append(d...)
			if diags.HasError() {
				return panelModel{}, diags
			}

			panelConfigHandled = true
			break
		}
	}

	if !panelConfigHandled && tfPanel != nil && typeutils.IsKnown(tfPanel.ConfigJSON) {
		// Preserve user-authored config_json when Kibana enriches defaults that are
		// semantically equivalent but produce perpetual drift.
		pm.ConfigJSON = tfPanel.ConfigJSON
		return pm, diags
	}

	configJSON, err := config.jsonString()
	if err != nil {
		diags = append(diags, diagutil.FrameworkDiagFromError(err)...)
		return panelModel{}, diags
	}
	pm.ConfigJSON = jsontypes.NewNormalizedValue(configJSON)
	return pm, diags
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
		section := kbapi.DashboardPanelSection{
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
			section.Uid = new(sm.ID.ValueString())
		}

		if len(sm.Panels) > 0 {
			innerPanels := make([]kbapi.DashboardSectionPanelItem, 0, len(sm.Panels))

			for _, pm := range sm.Panels {
				panelItem, d := pm.toAPI()
				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}

				var sectionItem kbapi.DashboardSectionPanelItem
				b, err := panelItem.MarshalJSON()
				if err != nil {
					diags.AddError("Failed to marshal section panel item", err.Error())
					return nil, diags
				}
				if err := sectionItem.UnmarshalJSON(b); err != nil {
					diags.AddError("Failed to create dashboard section panel item", err.Error())
					return nil, diags
				}
				innerPanels = append(innerPanels, sectionItem)
			}
			section.Panels = &innerPanels
		}

		var item kbapi.DashboardPanels_Item
		err := item.FromDashboardPanelSection(section)
		if err != nil {
			diags.AddError("Failed to create dashboard section item", err.Error())
		}
		apiPanels = append(apiPanels, item)
	}

	return &apiPanels, diags
}

func (pm panelModel) toAPI() (kbapi.DashboardPanelItem, diag.Diagnostics) {
	panelType := pm.Type.ValueString()

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

	var diags diag.Diagnostics
	var panelConfig apiPanelConfig
	var panelConfigHandled bool
	for _, converter := range panelConfigConverters {
		if converter.handlesTFPanelConfig(pm) {
			d := converter.mapPanelToAPI(pm, &panelConfig)
			diags.Append(d...)
			if diags.HasError() {
				return kbapi.DashboardPanelItem{}, diags
			}

			panelConfigHandled = true
			break
		}
	}

	if !panelConfigHandled && typeutils.IsKnown(pm.ConfigJSON) {
		var configMap map[string]any
		diags.Append(pm.ConfigJSON.Unmarshal(&configMap)...)
		if !diags.HasError() {
			typedConfig, err := panelConfigFromMap(panelType, configMap)
			if err != nil {
				diags.AddError("Failed to marshal panel config JSON", err.Error())
			} else {
				panelConfig = typedConfig
			}
		}
	}

	var panelItem kbapi.DashboardPanelItem
	switch panelType {
	case "lens":
		p := kbapi.KbnDashboardPanelLens{
			Grid: grid,
			Type: kbapi.KbnDashboardPanelLensType("lens"),
		}
		if typeutils.IsKnown(pm.ID) {
			p.Uid = new(pm.ID.ValueString())
		}
		if panelConfig.Lens != nil {
			p.Config = *panelConfig.Lens
		}
		panelJSON, err := json.Marshal(p)
		if err != nil {
			diags.AddError("Failed to marshal dashboard lens panel", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		if err := panelItem.UnmarshalJSON(panelJSON); err != nil {
			diags.AddError("Failed to create dashboard lens panel", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
	case "DASHBOARD_MARKDOWN":
		p := kbapi.KbnDashboardPanelDASHBOARDMARKDOWN{
			Grid: grid,
			Type: kbapi.KbnDashboardPanelDASHBOARDMARKDOWNType("DASHBOARD_MARKDOWN"),
		}
		if typeutils.IsKnown(pm.ID) {
			p.Uid = new(pm.ID.ValueString())
		}
		if panelConfig.Markdown != nil {
			p.Config = panelConfig.Markdown
		}
		panelJSON, err := json.Marshal(p)
		if err != nil {
			diags.AddError("Failed to marshal dashboard markdown panel", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		if err := panelItem.UnmarshalJSON(panelJSON); err != nil {
			diags.AddError("Failed to create dashboard markdown panel", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
	default:
		diags.AddError("Unsupported panel type", "Panel type "+panelType+" is not supported by this provider.")
		return kbapi.DashboardPanelItem{}, diags
	}

	return panelItem, diags
}

func convertRawPanel(raw map[string]any) (string, float32, float32, *float32, *float32, *string, apiPanelConfig, diag.Diagnostics) {
	rawType, ok := raw["type"].(string)
	if !ok || rawType == "" {
		var diags diag.Diagnostics
		diags.AddError("Invalid dashboard panel", "Missing required 'type' field in dashboard panel payload.")
		return "", 0, 0, nil, nil, nil, apiPanelConfig{}, diags
	}

	gridMap, ok := raw["grid"].(map[string]any)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid dashboard panel", "Missing required 'grid' object in dashboard panel payload.")
		return "", 0, 0, nil, nil, nil, apiPanelConfig{}, diags
	}

	gridX, xok := jsonNumberToFloat32(gridMap["x"])
	gridY, yok := jsonNumberToFloat32(gridMap["y"])
	if !xok || !yok {
		var diags diag.Diagnostics
		diags.AddError("Invalid dashboard panel", "Dashboard panel grid must include numeric x and y values.")
		return "", 0, 0, nil, nil, nil, apiPanelConfig{}, diags
	}

	var gridW *float32
	if w, wok := jsonNumberToFloat32(gridMap["w"]); wok {
		gridW = &w
	}
	var gridH *float32
	if h, hok := jsonNumberToFloat32(gridMap["h"]); hok {
		gridH = &h
	}

	var uid *string
	if rawUID, ok := raw["uid"].(string); ok {
		uid = &rawUID
	}

	var config apiPanelConfig
	if rawConfig, ok := raw["config"]; ok {
		configMap, ok := rawConfig.(map[string]any)
		if !ok {
			var diags diag.Diagnostics
			diags.AddError("Invalid dashboard panel", "Panel config must be an object.")
			return "", 0, 0, nil, nil, nil, apiPanelConfig{}, diags
		}
		var err error
		config, err = panelConfigFromMap(normalizePanelType(rawType), configMap)
		if err != nil {
			return "", 0, 0, nil, nil, nil, apiPanelConfig{}, diagutil.FrameworkDiagFromError(err)
		}
	}

	return normalizePanelType(rawType), gridX, gridY, gridW, gridH, uid, config, nil
}

func decodeJSONMap(union json.Marshaler) (map[string]any, error) {
	b, err := union.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	return raw, nil
}

func jsonNumberToFloat32(v any) (float32, bool) {
	n, ok := v.(float64)
	if !ok {
		return 0, false
	}
	return float32(n), true
}

func normalizePanelType(apiType string) string {
	switch apiType {
	case "kbn-dashboard-panel-lens":
		return "lens"
	case "kbn-dashboard-panel-DASHBOARD_MARKDOWN":
		return "DASHBOARD_MARKDOWN"
	default:
		return apiType
	}
}
