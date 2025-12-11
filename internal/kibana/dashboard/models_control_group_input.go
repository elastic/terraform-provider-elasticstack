package dashboard

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newControlGroupInputFromAPI(ctx context.Context, cgi *struct {
	AutoApplySelections  *bool                                                                      `json:"autoApplySelections,omitempty"`
	ChainingSystem       *kbapi.GetDashboardsDashboardId200DataControlGroupInputChainingSystem      `json:"chainingSystem,omitempty"`
	Controls             *[]kbapi.GetDashboardsDashboardId_200_Data_ControlGroupInput_Controls_Item `json:"controls,omitempty"`
	Enhancements         *map[string]interface{}                                                    `json:"enhancements,omitempty"`
	IgnoreParentSettings *struct {
		IgnoreFilters     *bool `json:"ignoreFilters,omitempty"`
		IgnoreQuery       *bool `json:"ignoreQuery,omitempty"`
		IgnoreTimerange   *bool `json:"ignoreTimerange,omitempty"`
		IgnoreValidations *bool `json:"ignoreValidations,omitempty"`
	} `json:"ignoreParentSettings,omitempty"`
	LabelPosition *kbapi.GetDashboardsDashboardId200DataControlGroupInputLabelPosition `json:"labelPosition,omitempty"`
}, diags *diag.Diagnostics) *controlGroupInputModel {
	if cgi == nil {
		return nil
	}

	model := &controlGroupInputModel{
		AutoApplySelections: types.BoolPointerValue(cgi.AutoApplySelections),
		ChainingSystem:      typeutils.StringishPointerValue((*string)(cgi.ChainingSystem)),
		LabelPosition:       typeutils.StringishPointerValue((*string)(cgi.LabelPosition)),
	}

	// Map enhancements
	if cgi.Enhancements != nil {
		enhancementsJSON, err := json.Marshal(cgi.Enhancements)
		if err != nil {
			diags.AddError("Failed to marshal enhancements", err.Error())
			model.Enhancements = jsontypes.NewNormalizedNull()
		} else {
			model.Enhancements = jsontypes.NewNormalizedValue(string(enhancementsJSON))
		}
	} else {
		model.Enhancements = jsontypes.NewNormalizedNull()
	}

	// Map ignore parent settings
	model.IgnoreParentSettings = newIgnoreParentSettingsFromAPI(cgi.IgnoreParentSettings)

	// Map controls
	if cgi.Controls != nil && len(*cgi.Controls) > 0 {
		var controls []controlModel
		for _, ctrl := range *cgi.Controls {
			cm := controlModel{
				ID:    typeutils.StringishPointerValue(ctrl.Id),
				Type:  types.StringValue(ctrl.Type),
				Order: types.Float64Value(float64(ctrl.Order)),
				Width: typeutils.StringishPointerValue(ctrl.Width),
				Grow:  types.BoolPointerValue(ctrl.Grow),
			}

			// Map control config
			if ctrl.ControlConfig != nil {
				configJSON, err := json.Marshal(ctrl.ControlConfig)
				if err != nil {
					diags.AddError("Failed to marshal control config", err.Error())
					cm.ControlConfig = jsontypes.NewNormalizedNull()
				} else {
					cm.ControlConfig = jsontypes.NewNormalizedValue(string(configJSON))
				}
			} else {
				cm.ControlConfig = jsontypes.NewNormalizedNull()
			}

			controls = append(controls, cm)
		}

		// Convert to types.List
		controlsList, listDiags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":             types.StringType,
				"type":           types.StringType,
				"order":          types.Float64Type,
				"width":          types.StringType,
				"grow":           types.BoolType,
				"control_config": jsontypes.Normalized{}.Type(ctx),
			},
		}, controls)
		diags.Append(listDiags...)
		model.Controls = controlsList
	} else {
		model.Controls = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":             types.StringType,
				"type":           types.StringType,
				"order":          types.Float64Type,
				"width":          types.StringType,
				"grow":           types.BoolType,
				"control_config": jsontypes.Normalized{}.Type(ctx),
			},
		})
	}

	return model
}

func newIgnoreParentSettingsFromAPI(ips *struct {
	IgnoreFilters     *bool `json:"ignoreFilters,omitempty"`
	IgnoreQuery       *bool `json:"ignoreQuery,omitempty"`
	IgnoreTimerange   *bool `json:"ignoreTimerange,omitempty"`
	IgnoreValidations *bool `json:"ignoreValidations,omitempty"`
}) *ignoreParentSettingsModel {
	if ips == nil {
		return nil
	}

	return &ignoreParentSettingsModel{
		IgnoreFilters:     types.BoolPointerValue(ips.IgnoreFilters),
		IgnoreQuery:       types.BoolPointerValue(ips.IgnoreQuery),
		IgnoreTimerange:   types.BoolPointerValue(ips.IgnoreTimerange),
		IgnoreValidations: types.BoolPointerValue(ips.IgnoreValidations),
	}
}

func (m *controlGroupInputModel) toAPICreate() *struct {
	AutoApplySelections  *bool                                                                         `json:"autoApplySelections,omitempty"`
	ChainingSystem       *kbapi.PostDashboardsDashboardJSONBodyDataControlGroupInputChainingSystem     `json:"chainingSystem,omitempty"`
	Controls             *[]kbapi.PostDashboardsDashboardJSONBody_Data_ControlGroupInput_Controls_Item `json:"controls,omitempty"`
	Enhancements         *map[string]interface{}                                                       `json:"enhancements,omitempty"`
	IgnoreParentSettings *struct {
		IgnoreFilters     *bool `json:"ignoreFilters,omitempty"`
		IgnoreQuery       *bool `json:"ignoreQuery,omitempty"`
		IgnoreTimerange   *bool `json:"ignoreTimerange,omitempty"`
		IgnoreValidations *bool `json:"ignoreValidations,omitempty"`
	} `json:"ignoreParentSettings,omitempty"`
	LabelPosition *kbapi.PostDashboardsDashboardJSONBodyDataControlGroupInputLabelPosition `json:"labelPosition,omitempty"`
} {
	if m == nil {
		return nil
	}

	result := &struct {
		AutoApplySelections  *bool                                                                         `json:"autoApplySelections,omitempty"`
		ChainingSystem       *kbapi.PostDashboardsDashboardJSONBodyDataControlGroupInputChainingSystem     `json:"chainingSystem,omitempty"`
		Controls             *[]kbapi.PostDashboardsDashboardJSONBody_Data_ControlGroupInput_Controls_Item `json:"controls,omitempty"`
		Enhancements         *map[string]interface{}                                                       `json:"enhancements,omitempty"`
		IgnoreParentSettings *struct {
			IgnoreFilters     *bool `json:"ignoreFilters,omitempty"`
			IgnoreQuery       *bool `json:"ignoreQuery,omitempty"`
			IgnoreTimerange   *bool `json:"ignoreTimerange,omitempty"`
			IgnoreValidations *bool `json:"ignoreValidations,omitempty"`
		} `json:"ignoreParentSettings,omitempty"`
		LabelPosition *kbapi.PostDashboardsDashboardJSONBodyDataControlGroupInputLabelPosition `json:"labelPosition,omitempty"`
	}{}

	if utils.IsKnown(m.AutoApplySelections) {
		result.AutoApplySelections = m.AutoApplySelections.ValueBoolPointer()
	}

	if utils.IsKnown(m.ChainingSystem) {
		chainingSystem := kbapi.PostDashboardsDashboardJSONBodyDataControlGroupInputChainingSystem(m.ChainingSystem.ValueString())
		result.ChainingSystem = &chainingSystem
	}

	if utils.IsKnown(m.LabelPosition) {
		labelPosition := kbapi.PostDashboardsDashboardJSONBodyDataControlGroupInputLabelPosition(m.LabelPosition.ValueString())
		result.LabelPosition = &labelPosition
	}

	// Map enhancements
	if utils.IsKnown(m.Enhancements) {
		var enhancements map[string]interface{}
		diags := m.Enhancements.Unmarshal(&enhancements)
		if !diags.HasError() {
			result.Enhancements = &enhancements
		}
	}

	// Map ignore parent settings
	if m.IgnoreParentSettings != nil {
		result.IgnoreParentSettings = m.IgnoreParentSettings.toAPICreate()
	}

	// Map controls
	if utils.IsKnown(m.Controls) && !m.Controls.IsNull() {
		var controls []controlModel
		// Extract controls from the list
		m.Controls.ElementsAs(context.Background(), &controls, false)

		var apiControls []kbapi.PostDashboardsDashboardJSONBody_Data_ControlGroupInput_Controls_Item
		for _, ctrl := range controls {
			apiCtrl := kbapi.PostDashboardsDashboardJSONBody_Data_ControlGroupInput_Controls_Item{
				Type:  ctrl.Type.ValueString(),
				Order: float32(ctrl.Order.ValueFloat64()),
			}

			if utils.IsKnown(ctrl.ID) {
				apiCtrl.Id = utils.Pointer(ctrl.ID.ValueString())
			}

			if utils.IsKnown(ctrl.Width) {
				width := kbapi.PostDashboardsDashboardJSONBodyDataControlGroupInputControlsWidth(ctrl.Width.ValueString())
				apiCtrl.Width = &width
			}

			if utils.IsKnown(ctrl.Grow) {
				apiCtrl.Grow = ctrl.Grow.ValueBoolPointer()
			}

			if utils.IsKnown(ctrl.ControlConfig) {
				var config map[string]interface{}
				diags := ctrl.ControlConfig.Unmarshal(&config)
				if !diags.HasError() {
					apiCtrl.ControlConfig = &config
				}
			}

			apiControls = append(apiControls, apiCtrl)
		}

		if len(apiControls) > 0 {
			result.Controls = &apiControls
		}
	}

	return result
}

func (m *controlGroupInputModel) toAPIUpdate() *struct {
	AutoApplySelections  *bool                                                                          `json:"autoApplySelections,omitempty"`
	ChainingSystem       *kbapi.PutDashboardsDashboardIdJSONBodyDataControlGroupInputChainingSystem     `json:"chainingSystem,omitempty"`
	Controls             *[]kbapi.PutDashboardsDashboardIdJSONBody_Data_ControlGroupInput_Controls_Item `json:"controls,omitempty"`
	Enhancements         *map[string]interface{}                                                        `json:"enhancements,omitempty"`
	IgnoreParentSettings *struct {
		IgnoreFilters     *bool `json:"ignoreFilters,omitempty"`
		IgnoreQuery       *bool `json:"ignoreQuery,omitempty"`
		IgnoreTimerange   *bool `json:"ignoreTimerange,omitempty"`
		IgnoreValidations *bool `json:"ignoreValidations,omitempty"`
	} `json:"ignoreParentSettings,omitempty"`
	LabelPosition *kbapi.PutDashboardsDashboardIdJSONBodyDataControlGroupInputLabelPosition `json:"labelPosition,omitempty"`
} {
	if m == nil {
		return nil
	}

	result := &struct {
		AutoApplySelections  *bool                                                                          `json:"autoApplySelections,omitempty"`
		ChainingSystem       *kbapi.PutDashboardsDashboardIdJSONBodyDataControlGroupInputChainingSystem     `json:"chainingSystem,omitempty"`
		Controls             *[]kbapi.PutDashboardsDashboardIdJSONBody_Data_ControlGroupInput_Controls_Item `json:"controls,omitempty"`
		Enhancements         *map[string]interface{}                                                        `json:"enhancements,omitempty"`
		IgnoreParentSettings *struct {
			IgnoreFilters     *bool `json:"ignoreFilters,omitempty"`
			IgnoreQuery       *bool `json:"ignoreQuery,omitempty"`
			IgnoreTimerange   *bool `json:"ignoreTimerange,omitempty"`
			IgnoreValidations *bool `json:"ignoreValidations,omitempty"`
		} `json:"ignoreParentSettings,omitempty"`
		LabelPosition *kbapi.PutDashboardsDashboardIdJSONBodyDataControlGroupInputLabelPosition `json:"labelPosition,omitempty"`
	}{}

	if utils.IsKnown(m.AutoApplySelections) {
		result.AutoApplySelections = m.AutoApplySelections.ValueBoolPointer()
	}

	if utils.IsKnown(m.ChainingSystem) {
		chainingSystem := kbapi.PutDashboardsDashboardIdJSONBodyDataControlGroupInputChainingSystem(m.ChainingSystem.ValueString())
		result.ChainingSystem = &chainingSystem
	}

	if utils.IsKnown(m.LabelPosition) {
		labelPosition := kbapi.PutDashboardsDashboardIdJSONBodyDataControlGroupInputLabelPosition(m.LabelPosition.ValueString())
		result.LabelPosition = &labelPosition
	}

	// Map enhancements
	if utils.IsKnown(m.Enhancements) {
		var enhancements map[string]interface{}
		diags := m.Enhancements.Unmarshal(&enhancements)
		if !diags.HasError() {
			result.Enhancements = &enhancements
		}
	}

	// Map ignore parent settings
	if m.IgnoreParentSettings != nil {
		result.IgnoreParentSettings = m.IgnoreParentSettings.toAPIUpdate()
	}

	// Map controls
	if utils.IsKnown(m.Controls) && !m.Controls.IsNull() {
		var controls []controlModel
		// Extract controls from the list
		m.Controls.ElementsAs(context.Background(), &controls, false)

		var apiControls []kbapi.PutDashboardsDashboardIdJSONBody_Data_ControlGroupInput_Controls_Item
		for _, ctrl := range controls {
			apiCtrl := kbapi.PutDashboardsDashboardIdJSONBody_Data_ControlGroupInput_Controls_Item{
				Type:  ctrl.Type.ValueString(),
				Order: float32(ctrl.Order.ValueFloat64()),
			}

			if utils.IsKnown(ctrl.ID) {
				apiCtrl.Id = utils.Pointer(ctrl.ID.ValueString())
			}

			if utils.IsKnown(ctrl.Width) {
				width := kbapi.PutDashboardsDashboardIdJSONBodyDataControlGroupInputControlsWidth(ctrl.Width.ValueString())
				apiCtrl.Width = &width
			}

			if utils.IsKnown(ctrl.Grow) {
				apiCtrl.Grow = ctrl.Grow.ValueBoolPointer()
			}

			if utils.IsKnown(ctrl.ControlConfig) {
				var config map[string]interface{}
				diags := ctrl.ControlConfig.Unmarshal(&config)
				if !diags.HasError() {
					apiCtrl.ControlConfig = &config
				}
			}

			apiControls = append(apiControls, apiCtrl)
		}

		if len(apiControls) > 0 {
			result.Controls = &apiControls
		}
	}

	return result
}

func (m *ignoreParentSettingsModel) toAPICreate() *struct {
	IgnoreFilters     *bool `json:"ignoreFilters,omitempty"`
	IgnoreQuery       *bool `json:"ignoreQuery,omitempty"`
	IgnoreTimerange   *bool `json:"ignoreTimerange,omitempty"`
	IgnoreValidations *bool `json:"ignoreValidations,omitempty"`
} {
	if m == nil {
		return nil
	}

	result := &struct {
		IgnoreFilters     *bool `json:"ignoreFilters,omitempty"`
		IgnoreQuery       *bool `json:"ignoreQuery,omitempty"`
		IgnoreTimerange   *bool `json:"ignoreTimerange,omitempty"`
		IgnoreValidations *bool `json:"ignoreValidations,omitempty"`
	}{}

	if utils.IsKnown(m.IgnoreFilters) {
		result.IgnoreFilters = m.IgnoreFilters.ValueBoolPointer()
	}
	if utils.IsKnown(m.IgnoreQuery) {
		result.IgnoreQuery = m.IgnoreQuery.ValueBoolPointer()
	}
	if utils.IsKnown(m.IgnoreTimerange) {
		result.IgnoreTimerange = m.IgnoreTimerange.ValueBoolPointer()
	}
	if utils.IsKnown(m.IgnoreValidations) {
		result.IgnoreValidations = m.IgnoreValidations.ValueBoolPointer()
	}

	return result
}

func (m *ignoreParentSettingsModel) toAPIUpdate() *struct {
	IgnoreFilters     *bool `json:"ignoreFilters,omitempty"`
	IgnoreQuery       *bool `json:"ignoreQuery,omitempty"`
	IgnoreTimerange   *bool `json:"ignoreTimerange,omitempty"`
	IgnoreValidations *bool `json:"ignoreValidations,omitempty"`
} {
	if m == nil {
		return nil
	}

	result := &struct {
		IgnoreFilters     *bool `json:"ignoreFilters,omitempty"`
		IgnoreQuery       *bool `json:"ignoreQuery,omitempty"`
		IgnoreTimerange   *bool `json:"ignoreTimerange,omitempty"`
		IgnoreValidations *bool `json:"ignoreValidations,omitempty"`
	}{}

	if utils.IsKnown(m.IgnoreFilters) {
		result.IgnoreFilters = m.IgnoreFilters.ValueBoolPointer()
	}
	if utils.IsKnown(m.IgnoreQuery) {
		result.IgnoreQuery = m.IgnoreQuery.ValueBoolPointer()
	}
	if utils.IsKnown(m.IgnoreTimerange) {
		result.IgnoreTimerange = m.IgnoreTimerange.ValueBoolPointer()
	}
	if utils.IsKnown(m.IgnoreValidations) {
		result.IgnoreValidations = m.IgnoreValidations.ValueBoolPointer()
	}

	return result
}
