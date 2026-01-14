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
	Type                 types.String           `tfsdk:"type"`
	Grid                 panelGridModel         `tfsdk:"grid"`
	PanelID              types.String           `tfsdk:"panel_id"`
	EmbeddableConfig     *embeddableConfigModel `tfsdk:"embeddable_config"`
	EmbeddableConfigJSON jsontypes.Normalized   `tfsdk:"embeddable_config_json"`
}

type panelGridModel struct {
	X types.Int64 `tfsdk:"x"`
	Y types.Int64 `tfsdk:"y"`
	W types.Int64 `tfsdk:"w"`
	H types.Int64 `tfsdk:"h"`
}

type embeddableConfigModel struct {
	Content         types.String `tfsdk:"content"`
	Description     types.String `tfsdk:"description"`
	HidePanelTitles types.Bool   `tfsdk:"hide_panel_titles"`
	Title           types.String `tfsdk:"title"`
}

func (m *dashboardModel) mapPanelsFromAPI(apiPanels *kbapi.DashboardPanels) ([]panelModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if apiPanels == nil || len(*apiPanels) == 0 {
		return nil, diags
	}

	var panels []panelModel
	for _, item := range *apiPanels {
		// Try to handle as DashboardPanelItem
		panelItem, err := item.AsDashboardPanelItem()
		if err == nil {
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
				pm.PanelID = types.StringValue(*panelItem.Uid)
			} else {
				pm.PanelID = types.StringNull()
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
						// Map to embeddableConfigModel
						pm.EmbeddableConfig = &embeddableConfigModel{
							Content:         types.StringValue(config0.Content),
							Description:     types.StringPointerValue(config0.Description),
							HidePanelTitles: types.BoolPointerValue(config0.HidePanelTitles),
							Title:           types.StringPointerValue(config0.Title),
						}
						pm.EmbeddableConfigJSON = jsontypes.NewNormalizedNull()
						mappedToStruct = true
					}
				}
			}

			if !mappedToStruct {
				pm.EmbeddableConfig = nil
				configBytes, err := panelItem.Config.MarshalJSON()
				if err == nil {
					pm.EmbeddableConfigJSON = jsontypes.NewNormalizedValue(string(configBytes))
				} else {
					pm.EmbeddableConfigJSON = jsontypes.NewNormalizedNull()
				}
			}

			panels = append(panels, pm)
		}
	}

	return panels, diags
}

func (m *dashboardModel) panelsToAPI() (*kbapi.DashboardPanels, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.Panels == nil {
		return nil, diags
	}

	panels := m.Panels
	apiPanels := make(kbapi.DashboardPanels, 0, len(panels))
	for _, pm := range panels {
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

		if !pm.Grid.W.IsNull() && !pm.Grid.W.IsUnknown() {
			w := float32(pm.Grid.W.ValueInt64())
			panelItem.Grid.W = &w
		}
		if !pm.Grid.H.IsNull() && !pm.Grid.H.IsUnknown() {
			h := float32(pm.Grid.H.ValueInt64())
			panelItem.Grid.H = &h
		}

		if utils.IsKnown(pm.PanelID) {
			panelItem.Uid = utils.Pointer(pm.PanelID.ValueString())
		}

		if pm.EmbeddableConfig != nil {
			configModel := *pm.EmbeddableConfig

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
		} else if utils.IsKnown(pm.EmbeddableConfigJSON) {
			var configMap map[string]interface{}
			diags.Append(pm.EmbeddableConfigJSON.Unmarshal(&configMap)...)
			if !diags.HasError() {
				if err := panelItem.Config.FromDashboardPanelItemConfig1(configMap); err != nil {
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

	return &apiPanels, diags
}
