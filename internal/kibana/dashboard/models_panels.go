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
	"reflect"

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
	handlesAPIPanelConfig(ctx *panelModel, panelType string, config kbapi.DashboardPanelItem_Config) bool
	handlesTFPanelConfig(tfModel panelModel) bool
	populateFromAPIPanel(ctx context.Context, tfModel *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics
	mapPanelToAPI(tfModel panelModel, config *kbapi.DashboardPanelItem_Config) diag.Diagnostics
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
		// Try to handle as DashboardPanelItem (requires type)
		panelItem, err := item.AsDashboardPanelItem()
		if err == nil && panelItem.Type != "" {
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
			continue
		}

		// Try to handle as DashboardPanelSection
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

	pm.Type = types.StringValue(panelItem.Type)
	pm.Grid = panelGridModel{
		X: types.Int64Value(int64(panelItem.Grid.X)),
		Y: types.Int64Value(int64(panelItem.Grid.Y)),
	}
	if panelItem.Grid.W != nil {
		pm.Grid.W = types.Int64Value(int64(*panelItem.Grid.W))
	} else {
		pm.Grid.W = types.Int64Null()
	}
	if panelItem.Grid.H != nil {
		pm.Grid.H = types.Int64Value(int64(*panelItem.Grid.H))
	} else {
		pm.Grid.H = types.Int64Null()
	}

	if panelItem.Uid != nil {
		pm.ID = types.StringValue(*panelItem.Uid)
	} else {
		pm.ID = types.StringNull()
	}

	var diags diag.Diagnostics
	for _, converter := range panelConfigConverters {
		if converter.handlesAPIPanelConfig(tfPanel, panelItem.Type, panelItem.Config) {
			d := converter.populateFromAPIPanel(ctx, &pm, panelItem.Config)
			diags.Append(d...)
			if diags.HasError() {
				return panelModel{}, diags
			}

			break
		}
	}

	configBytes, err := panelItem.Config.MarshalJSON()
	if err != nil {
		diags = append(diags, diagutil.FrameworkDiagFromError(err)...)
		return panelModel{}, diags
	}

	apiConfigJSON := string(configBytes)
	if tfPanel != nil && typeutils.IsKnown(tfPanel.ConfigJSON) {
		planJSON := tfPanel.ConfigJSON.ValueString()
		if jsonSemanticallyEqual([]byte(planJSON), configBytes) {
			pm.ConfigJSON = tfPanel.ConfigJSON
		} else {
			pm.ConfigJSON = jsontypes.NewNormalizedValue(apiConfigJSON)
		}
	} else {
		pm.ConfigJSON = jsontypes.NewNormalizedValue(apiConfigJSON)
	}
	return pm, diags
}

func jsonSemanticallyEqual(a, b []byte) bool {
	var va, vb any
	if json.Unmarshal(a, &va) != nil || json.Unmarshal(b, &vb) != nil {
		return false
	}
	return reflect.DeepEqual(va, vb)
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
		err := item.FromDashboardPanelSection(section)
		if err != nil {
			diags.AddError("Failed to create dashboard section item", err.Error())
		}
		apiPanels = append(apiPanels, item)
	}

	return &apiPanels, diags
}

func (pm panelModel) toAPI() (kbapi.DashboardPanelItem, diag.Diagnostics) {
	panelItem := kbapi.DashboardPanelItem{
		Type: pm.Type.ValueString(),
		Grid: struct {
			H *float32 `json:"h,omitempty"`
			W *float32 `json:"w,omitempty"`
			X float32  `json:"x"`
			Y float32  `json:"y"`
		}{
			X: float32(pm.Grid.X.ValueInt64()),
			Y: float32(pm.Grid.Y.ValueInt64()),
		},
	}

	if typeutils.IsKnown(pm.Grid.W) {
		w := float32(pm.Grid.W.ValueInt64())
		panelItem.Grid.W = &w
	}
	if typeutils.IsKnown(pm.Grid.H) {
		h := float32(pm.Grid.H.ValueInt64())
		panelItem.Grid.H = &h
	}

	if typeutils.IsKnown(pm.ID) {
		panelItem.Uid = new(pm.ID.ValueString())
	}

	var diags diag.Diagnostics
	var panelConfigHandled bool
	for _, converter := range panelConfigConverters {
		if converter.handlesTFPanelConfig(pm) {
			d := converter.mapPanelToAPI(pm, &panelItem.Config)
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
			if err := panelItem.Config.FromDashboardPanelItemConfig8(configMap); err != nil {
				diags.AddError("Failed to marshal panel config JSON", err.Error())
			}
		}
	}

	return panelItem, diags
}
