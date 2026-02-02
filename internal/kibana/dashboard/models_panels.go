package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type panelModel struct {
	Type           types.String         `tfsdk:"type"`
	Grid           panelGridModel       `tfsdk:"grid"`
	ID             types.String         `tfsdk:"id"`
	MarkdownConfig *markdownConfigModel `tfsdk:"markdown_config"`
	XYChartConfig  *xyChartConfigModel  `tfsdk:"xy_chart_config"`
	ConfigJSON     jsontypes.Normalized `tfsdk:"config_json"`
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
	handlesAPIPanelConfig(string, kbapi.DashboardPanelItem_Config) bool
	handlesTFPanelConfig(pm panelModel) bool
	populateFromAPIPanel(context.Context, *panelModel, kbapi.DashboardPanelItem_Config) diag.Diagnostics
	mapPanelToAPI(panelModel, *kbapi.DashboardPanelItem_Config) diag.Diagnostics
}

var panelConfigConverters = []panelConfigConverter{
	markdownPanelConfigConverter{},
	newXYChartPanelConfigConverter(),
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
			panel, d := m.mapPanelFromAPI(ctx, panelItem)
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
			sectionModel, d := m.mapSectionFromAPI(ctx, section)
			diags.Append(d...)
			if diags.HasError() {
				return nil, nil, diags
			}
			sections = append(sections, sectionModel)
		}
	}

	return panels, sections, diags
}

func (m *dashboardModel) mapSectionFromAPI(ctx context.Context, section kbapi.DashboardPanelSection) (sectionModel, diag.Diagnostics) {
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
			pm, d := m.mapPanelFromAPI(ctx, p)
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

func (m *dashboardModel) mapPanelFromAPI(ctx context.Context, panelItem kbapi.DashboardPanelItem) (panelModel, diag.Diagnostics) {
	pm := panelModel{
		Type: types.StringValue(panelItem.Type),
		Grid: panelGridModel{
			X: types.Int64Value(int64(panelItem.Grid.X)),
			Y: types.Int64Value(int64(panelItem.Grid.Y)),
		},
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
		if converter.handlesAPIPanelConfig(panelItem.Type, panelItem.Config) {
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

	pm.ConfigJSON = jsontypes.NewNormalizedValue(string(configBytes))
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

		if utils.IsKnown(sm.Collapsed) {
			section.Collapsed = utils.Pointer(sm.Collapsed.ValueBool())
		}
		if utils.IsKnown(sm.ID) {
			section.Uid = utils.Pointer(sm.ID.ValueString())
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

	if utils.IsKnown(pm.Grid.W) {
		w := float32(pm.Grid.W.ValueInt64())
		panelItem.Grid.W = &w
	}
	if utils.IsKnown(pm.Grid.H) {
		h := float32(pm.Grid.H.ValueInt64())
		panelItem.Grid.H = &h
	}

	if utils.IsKnown(pm.ID) {
		panelItem.Uid = utils.Pointer(pm.ID.ValueString())
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

	if !panelConfigHandled && utils.IsKnown(pm.ConfigJSON) {
		var configMap map[string]interface{}
		diags.Append(pm.ConfigJSON.Unmarshal(&configMap)...)
		if !diags.HasError() {
			if err := panelItem.Config.FromDashboardPanelItemConfig2(configMap); err != nil {
				diags.AddError("Failed to marshal panel config JSON", err.Error())
			}
		}
	}

	return panelItem, diags
}
