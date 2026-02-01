package dashboard

import (
	"encoding/json"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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
	ConfigJSON     jsontypes.Normalized `tfsdk:"config_json"`
}

type panelGridModel struct {
	X types.Int64 `tfsdk:"x"`
	Y types.Int64 `tfsdk:"y"`
	W types.Int64 `tfsdk:"w"`
	H types.Int64 `tfsdk:"h"`
}

type markdownConfigModel struct {
	Content         types.String `tfsdk:"content"`
	Description     types.String `tfsdk:"description"`
	HidePanelTitles types.Bool   `tfsdk:"hide_panel_titles"`
	Title           types.String `tfsdk:"title"`
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

func (m *dashboardModel) mapPanelsFromAPI(apiPanels *kbapi.DashboardPanels) ([]panelModel, []sectionModel, diag.Diagnostics) {
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
			panels = append(panels, m.mapPanelFromAPI(panelItem))
			continue
		}

		// Try to handle as DashboardPanelSection
		section, err := item.AsDashboardPanelSection()
		if err == nil {
			sections = append(sections, m.mapSectionFromAPI(section))
		}
	}

	return panels, sections, diags
}

func (m *dashboardModel) mapSectionFromAPI(section kbapi.DashboardPanelSection) sectionModel {
	sm := sectionModel{
		Title:     types.StringValue(section.Title),
		Collapsed: types.BoolPointerValue(section.Collapsed),
		ID:        types.StringPointerValue(section.Uid),
		Grid: sectionGridModel{
			Y: types.Int64Value(int64(section.Grid.Y)),
		},
	}

	// Map section panels
	if section.Panels != nil {
		var innerPanels []panelModel
		for _, p := range *section.Panels {
			pm := panelModel{
				Type: types.StringValue(p.Type),
				ID:   types.StringPointerValue(p.Uid),
				Grid: panelGridModel{
					X: types.Int64Value(int64(p.Grid.X)),
					Y: types.Int64Value(int64(p.Grid.Y)),
				},
			}
			if p.Grid.W != nil {
				pm.Grid.W = types.Int64Value(int64(*p.Grid.W))
			} else {
				pm.Grid.W = types.Int64Null()
			}
			if p.Grid.H != nil {
				pm.Grid.H = types.Int64Value(int64(*p.Grid.H))
			} else {
				pm.Grid.H = types.Int64Null()
			}

			// Map Config
			// Similar logic to standard panels but adapting to DashboardPanelSection_Panels_Config
			var mappedToStruct bool
			config0, err := p.Config.AsDashboardPanelSectionPanelsConfig0()
			if err == nil {
				rawJSON, _ := p.Config.MarshalJSON()
				structJSON, _ := json.Marshal(config0)
				var rawMap, structMap map[string]interface{}
				if json.Unmarshal(rawJSON, &rawMap) == nil && json.Unmarshal(structJSON, &structMap) == nil {
					if reflect.DeepEqual(rawMap, structMap) {
						pm.MarkdownConfig = &markdownConfigModel{
							Content:         types.StringValue(config0.Content),
							Description:     types.StringPointerValue(config0.Description),
							HidePanelTitles: types.BoolPointerValue(config0.HidePanelTitles),
							Title:           types.StringPointerValue(config0.Title),
						}
						pm.ConfigJSON = jsontypes.NewNormalizedNull()
						mappedToStruct = true
					}
				}
			}

			if !mappedToStruct {
				pm.MarkdownConfig = nil
				configBytes, err := p.Config.MarshalJSON()
				if err == nil {
					pm.ConfigJSON = jsontypes.NewNormalizedValue(string(configBytes))
				} else {
					pm.ConfigJSON = jsontypes.NewNormalizedNull()
				}
			}

			innerPanels = append(innerPanels, pm)
		}
		sm.Panels = innerPanels
	}
	return sm
}

func (m *dashboardModel) mapPanelFromAPI(panelItem kbapi.DashboardPanelItem) panelModel {
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

	// Try to map to the structured config first
	var mappedToStruct bool
	config0, err := panelItem.Config.AsDashboardPanelItemConfig0()
	if err == nil {
		// To ensure we don't lose data, we need to check if the round-trip matches.
		// If the JSON contains fields not in Config0, we should prefer JSON.
		rawJSON, _ := panelItem.Config.MarshalJSON()
		structJSON, _ := json.Marshal(config0)

		var rawMap, structMap map[string]interface{}
		if json.Unmarshal(rawJSON, &rawMap) == nil && json.Unmarshal(structJSON, &structMap) == nil {
			if reflect.DeepEqual(rawMap, structMap) {
				// Map to markdownConfigModel
				pm.MarkdownConfig = &markdownConfigModel{
					Content:         types.StringValue(config0.Content),
					Description:     types.StringPointerValue(config0.Description),
					HidePanelTitles: types.BoolPointerValue(config0.HidePanelTitles),
					Title:           types.StringPointerValue(config0.Title),
				}
				pm.ConfigJSON = jsontypes.NewNormalizedNull()
				mappedToStruct = true
			}
		}
	}

	if !mappedToStruct {
		pm.MarkdownConfig = nil
		configBytes, err := panelItem.Config.MarshalJSON()
		if err == nil {
			pm.ConfigJSON = jsontypes.NewNormalizedValue(string(configBytes))
		} else {
			pm.ConfigJSON = jsontypes.NewNormalizedNull()
		}
	}
	return pm
}

func (m *dashboardModel) panelsToAPI() (*kbapi.DashboardPanels, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.Panels == nil && m.Sections == nil {
		return nil, diags
	}

	apiPanels := make(kbapi.DashboardPanels, 0, len(m.Panels)+len(m.Sections))

	// Process panels
	for _, pm := range m.Panels {
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

		if pm.MarkdownConfig != nil {
			configModel := *pm.MarkdownConfig

			config0 := kbapi.DashboardPanelItemConfig0{
				Content: configModel.Content.ValueString(),
			}
			if utils.IsKnown(configModel.Description) {
				config0.Description = utils.Pointer(configModel.Description.ValueString())
			}
			if utils.IsKnown(configModel.HidePanelTitles) {
				config0.HidePanelTitles = utils.Pointer(configModel.HidePanelTitles.ValueBool())
			}
			if utils.IsKnown(configModel.Title) {
				config0.Title = utils.Pointer(configModel.Title.ValueString())
			}

			if err := panelItem.Config.FromDashboardPanelItemConfig0(config0); err != nil {
				diags.AddError("Failed to marshal panel config", err.Error())
			}
		} else if utils.IsKnown(pm.ConfigJSON) {
			var configMap map[string]interface{}
			diags.Append(pm.ConfigJSON.Unmarshal(&configMap)...)
			if !diags.HasError() {
				if err := panelItem.Config.FromDashboardPanelItemConfig2(configMap); err != nil {
					diags.AddError("Failed to marshal panel config JSON", err.Error())
				}
			}
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
			innerPanels := make([]struct {
				Config kbapi.DashboardPanelSection_Panels_Config `json:"config"`
				Grid   struct {
					H *float32 `json:"h,omitempty"`
					W *float32 `json:"w,omitempty"`
					X float32  `json:"x"`
					Y float32  `json:"y"`
				} `json:"grid"`
				Type    string  `json:"type"`
				Uid     *string `json:"uid,omitempty"`
				Version *string `json:"version,omitempty"`
			}, 0, len(sm.Panels))

			for _, pm := range sm.Panels {
				p := struct {
					Config kbapi.DashboardPanelSection_Panels_Config `json:"config"`
					Grid   struct {
						H *float32 `json:"h,omitempty"`
						W *float32 `json:"w,omitempty"`
						X float32  `json:"x"`
						Y float32  `json:"y"`
					} `json:"grid"`
					Type    string  `json:"type"`
					Uid     *string `json:"uid,omitempty"`
					Version *string `json:"version,omitempty"`
				}{
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

				if !pm.Grid.W.IsNull() && !pm.Grid.W.IsUnknown() {
					w := float32(pm.Grid.W.ValueInt64())
					p.Grid.W = &w
				}
				if !pm.Grid.H.IsNull() && !pm.Grid.H.IsUnknown() {
					h := float32(pm.Grid.H.ValueInt64())
					p.Grid.H = &h
				}

				if utils.IsKnown(pm.ID) {
					p.Uid = utils.Pointer(pm.ID.ValueString())
				}

				// Map config for section panel
				if pm.MarkdownConfig != nil {
					configModel := *pm.MarkdownConfig
					config0 := kbapi.DashboardPanelSectionPanelsConfig0{
						Content: configModel.Content.ValueString(),
					}
					if utils.IsKnown(configModel.Description) {
						config0.Description = utils.Pointer(configModel.Description.ValueString())
					}
					if utils.IsKnown(configModel.HidePanelTitles) {
						config0.HidePanelTitles = utils.Pointer(configModel.HidePanelTitles.ValueBool())
					}
					if utils.IsKnown(configModel.Title) {
						config0.Title = utils.Pointer(configModel.Title.ValueString())
					}

					if err := p.Config.FromDashboardPanelSectionPanelsConfig0(config0); err != nil {
						diags.AddError("Failed to marshal section panel config", err.Error())
					}
				} else if utils.IsKnown(pm.ConfigJSON) {
					var configMap map[string]interface{}
					diags.Append(pm.ConfigJSON.Unmarshal(&configMap)...)
					if !diags.HasError() {
						if err := p.Config.FromDashboardPanelSectionPanelsConfig2(configMap); err != nil {
							diags.AddError("Failed to marshal section panel config JSON", err.Error())
						}
					}
				}

				innerPanels = append(innerPanels, p)
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
