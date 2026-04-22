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
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type optionsListControlDisplaySettingsModel struct {
	Placeholder   types.String `tfsdk:"placeholder"`
	HideActionBar types.Bool   `tfsdk:"hide_action_bar"`
	HideExclude   types.Bool   `tfsdk:"hide_exclude"`
	HideExists    types.Bool   `tfsdk:"hide_exists"`
	HideSort      types.Bool   `tfsdk:"hide_sort"`
}

type optionsListControlSortModel struct {
	By        types.String `tfsdk:"by"`
	Direction types.String `tfsdk:"direction"`
}

type optionsListControlConfigModel struct {
	DataViewID        types.String                            `tfsdk:"data_view_id"`
	FieldName         types.String                            `tfsdk:"field_name"`
	Title             types.String                            `tfsdk:"title"`
	UseGlobalFilters  types.Bool                              `tfsdk:"use_global_filters"`
	IgnoreValidations types.Bool                              `tfsdk:"ignore_validations"`
	SingleSelect      types.Bool                              `tfsdk:"single_select"`
	Exclude           types.Bool                              `tfsdk:"exclude"`
	ExistsSelected    types.Bool                              `tfsdk:"exists_selected"`
	RunPastTimeout    types.Bool                              `tfsdk:"run_past_timeout"`
	SearchTechnique   types.String                            `tfsdk:"search_technique"`
	SelectedOptions   types.List                              `tfsdk:"selected_options"`
	DisplaySettings   *optionsListControlDisplaySettingsModel `tfsdk:"display_settings"`
	Sort              *optionsListControlSortModel            `tfsdk:"sort"`
}

// populateOptionsListControlFromAPI reads back an options list control config from the API
// response and updates the panel model. Null-preservation semantics apply: if a field is
// null in the existing TF state (pm.OptionsListControlConfig), we do not overwrite it with
// a Kibana-returned value. If there is no existing config block, and Kibana returns an
// empty/absent config, we leave OptionsListControlConfig as nil.
//
// tfPanel is the prior TF state/plan panel, or nil on import.
func populateOptionsListControlFromAPI(pm *panelModel, tfPanel *panelModel, ol *kbapi.KbnDashboardPanelTypeOptionsListControl) {
	if ol == nil {
		return
	}
	apiConfig := &ol.Config
	existing := pm.OptionsListControlConfig

	if tfPanel == nil {
		// On import: populate from API. Required fields are always set; optional fields are
		// set only if present in the API response. Fields that Kibana returns as server-side
		// defaults (e.g. use_global_filters, exclude, sort) are treated as optional and left
		// null when absent from the user's configuration — matching the null-preservation
		// behaviour used during normal apply reads.
		pm.OptionsListControlConfig = &optionsListControlConfigModel{
			DataViewID:      types.StringValue(apiConfig.DataViewId),
			FieldName:       types.StringValue(apiConfig.FieldName),
			SelectedOptions: types.ListNull(types.StringType),
		}
		existing = pm.OptionsListControlConfig
		if apiConfig.Title != nil {
			existing.Title = types.StringValue(*apiConfig.Title)
		}
		if apiConfig.SingleSelect != nil {
			existing.SingleSelect = types.BoolValue(*apiConfig.SingleSelect)
		}
		if apiConfig.SearchTechnique != nil {
			existing.SearchTechnique = types.StringValue(string(*apiConfig.SearchTechnique))
		}
		if apiConfig.SelectedOptions != nil {
			existing.SelectedOptions = selectedOptionsToList(*apiConfig.SelectedOptions)
		}
		if apiConfig.DisplaySettings != nil {
			existing.DisplaySettings = displaySettingsFromAPI(apiConfig.DisplaySettings)
		}
		return
	}

	// If the existing state has no config block, preserve nil intent.
	if existing == nil {
		return
	}

	// Block exists in state — update required fields unconditionally, optional fields only when known.
	existing.DataViewID = types.StringValue(apiConfig.DataViewId)
	existing.FieldName = types.StringValue(apiConfig.FieldName)

	if typeutils.IsKnown(existing.Title) && apiConfig.Title != nil {
		existing.Title = types.StringValue(*apiConfig.Title)
	}
	if typeutils.IsKnown(existing.UseGlobalFilters) && apiConfig.UseGlobalFilters != nil {
		existing.UseGlobalFilters = types.BoolValue(*apiConfig.UseGlobalFilters)
	}
	if typeutils.IsKnown(existing.IgnoreValidations) && apiConfig.IgnoreValidations != nil {
		existing.IgnoreValidations = types.BoolValue(*apiConfig.IgnoreValidations)
	}
	if typeutils.IsKnown(existing.SingleSelect) && apiConfig.SingleSelect != nil {
		existing.SingleSelect = types.BoolValue(*apiConfig.SingleSelect)
	}
	if typeutils.IsKnown(existing.Exclude) && apiConfig.Exclude != nil {
		existing.Exclude = types.BoolValue(*apiConfig.Exclude)
	}
	if typeutils.IsKnown(existing.ExistsSelected) && apiConfig.ExistsSelected != nil {
		existing.ExistsSelected = types.BoolValue(*apiConfig.ExistsSelected)
	}
	if typeutils.IsKnown(existing.RunPastTimeout) && apiConfig.RunPastTimeout != nil {
		existing.RunPastTimeout = types.BoolValue(*apiConfig.RunPastTimeout)
	}
	if typeutils.IsKnown(existing.SearchTechnique) && apiConfig.SearchTechnique != nil {
		existing.SearchTechnique = types.StringValue(string(*apiConfig.SearchTechnique))
	}
	if !existing.SelectedOptions.IsNull() && !existing.SelectedOptions.IsUnknown() && apiConfig.SelectedOptions != nil {
		existing.SelectedOptions = selectedOptionsToList(*apiConfig.SelectedOptions)
	}

	// Display settings: nil or empty API response => treat as omitted block.
	if existing.DisplaySettings != nil && apiConfig.DisplaySettings != nil {
		ds := existing.DisplaySettings
		api := apiConfig.DisplaySettings
		if typeutils.IsKnown(ds.Placeholder) && api.Placeholder != nil {
			ds.Placeholder = types.StringValue(*api.Placeholder)
		}
		if typeutils.IsKnown(ds.HideActionBar) && api.HideActionBar != nil {
			ds.HideActionBar = types.BoolValue(*api.HideActionBar)
		}
		if typeutils.IsKnown(ds.HideExclude) && api.HideExclude != nil {
			ds.HideExclude = types.BoolValue(*api.HideExclude)
		}
		if typeutils.IsKnown(ds.HideExists) && api.HideExists != nil {
			ds.HideExists = types.BoolValue(*api.HideExists)
		}
		if typeutils.IsKnown(ds.HideSort) && api.HideSort != nil {
			ds.HideSort = types.BoolValue(*api.HideSort)
		}
	}

	// Sort: update only when both state and API have a sort block.
	if existing.Sort != nil && apiConfig.Sort != nil {
		existing.Sort.By = types.StringValue(string(apiConfig.Sort.By))
		existing.Sort.Direction = types.StringValue(string(apiConfig.Sort.Direction))
	}
}

// buildOptionsListControlConfig writes the TF model fields into the API panel struct.
func buildOptionsListControlConfig(pm panelModel, olPanel *kbapi.KbnDashboardPanelTypeOptionsListControl) {
	cfg := pm.OptionsListControlConfig
	if cfg == nil {
		return
	}

	olPanel.Config.DataViewId = cfg.DataViewID.ValueString()
	olPanel.Config.FieldName = cfg.FieldName.ValueString()

	if typeutils.IsKnown(cfg.Title) {
		olPanel.Config.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.UseGlobalFilters) {
		olPanel.Config.UseGlobalFilters = cfg.UseGlobalFilters.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.IgnoreValidations) {
		olPanel.Config.IgnoreValidations = cfg.IgnoreValidations.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.SingleSelect) {
		olPanel.Config.SingleSelect = cfg.SingleSelect.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.Exclude) {
		olPanel.Config.Exclude = cfg.Exclude.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.ExistsSelected) {
		olPanel.Config.ExistsSelected = cfg.ExistsSelected.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.RunPastTimeout) {
		olPanel.Config.RunPastTimeout = cfg.RunPastTimeout.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.SearchTechnique) {
		st := kbapi.KbnDashboardPanelTypeOptionsListControlConfigSearchTechnique(cfg.SearchTechnique.ValueString())
		olPanel.Config.SearchTechnique = &st
	}
	if !cfg.SelectedOptions.IsNull() && !cfg.SelectedOptions.IsUnknown() {
		elems := cfg.SelectedOptions.Elements()
		items := make([]kbapi.KbnDashboardPanelTypeOptionsListControl_Config_SelectedOptions_Item, 0, len(elems))
		for _, e := range elems {
			s := e.(types.String).ValueString()
			var item kbapi.KbnDashboardPanelTypeOptionsListControl_Config_SelectedOptions_Item
			if err := item.FromKbnDashboardPanelTypeOptionsListControlConfigSelectedOptions0(s); err == nil {
				items = append(items, item)
			}
		}
		olPanel.Config.SelectedOptions = &items
	}
	if cfg.DisplaySettings != nil {
		ds := cfg.DisplaySettings
		apiDS := &struct {
			HideActionBar *bool   `json:"hide_action_bar,omitempty"`
			HideExclude   *bool   `json:"hide_exclude,omitempty"`
			HideExists    *bool   `json:"hide_exists,omitempty"`
			HideSort      *bool   `json:"hide_sort,omitempty"`
			Placeholder   *string `json:"placeholder,omitempty"`
		}{}
		if typeutils.IsKnown(ds.Placeholder) {
			apiDS.Placeholder = ds.Placeholder.ValueStringPointer()
		}
		if typeutils.IsKnown(ds.HideActionBar) {
			apiDS.HideActionBar = ds.HideActionBar.ValueBoolPointer()
		}
		if typeutils.IsKnown(ds.HideExclude) {
			apiDS.HideExclude = ds.HideExclude.ValueBoolPointer()
		}
		if typeutils.IsKnown(ds.HideExists) {
			apiDS.HideExists = ds.HideExists.ValueBoolPointer()
		}
		if typeutils.IsKnown(ds.HideSort) {
			apiDS.HideSort = ds.HideSort.ValueBoolPointer()
		}
		olPanel.Config.DisplaySettings = apiDS
	}
	if cfg.Sort != nil {
		olPanel.Config.Sort = &struct {
			By        kbapi.KbnDashboardPanelTypeOptionsListControlConfigSortBy        `json:"by"`
			Direction kbapi.KbnDashboardPanelTypeOptionsListControlConfigSortDirection `json:"direction"`
		}{
			By:        kbapi.KbnDashboardPanelTypeOptionsListControlConfigSortBy(cfg.Sort.By.ValueString()),
			Direction: kbapi.KbnDashboardPanelTypeOptionsListControlConfigSortDirection(cfg.Sort.Direction.ValueString()),
		}
	}
}

func selectedOptionsToList(items []kbapi.KbnDashboardPanelTypeOptionsListControl_Config_SelectedOptions_Item) types.List {
	values := make([]attr.Value, 0, len(items))
	for _, item := range items {
		if s, err := item.AsKbnDashboardPanelTypeOptionsListControlConfigSelectedOptions0(); err == nil {
			values = append(values, types.StringValue(s))
			continue
		}
		if n, err := item.AsKbnDashboardPanelTypeOptionsListControlConfigSelectedOptions1(); err == nil {
			values = append(values, types.StringValue(strconv.FormatFloat(float64(n), 'f', -1, 32)))
		}
	}
	return types.ListValueMust(types.StringType, values)
}

func displaySettingsFromAPI(api *struct {
	HideActionBar *bool   `json:"hide_action_bar,omitempty"`
	HideExclude   *bool   `json:"hide_exclude,omitempty"`
	HideExists    *bool   `json:"hide_exists,omitempty"`
	HideSort      *bool   `json:"hide_sort,omitempty"`
	Placeholder   *string `json:"placeholder,omitempty"`
}) *optionsListControlDisplaySettingsModel {
	if api == nil {
		return nil
	}
	ds := &optionsListControlDisplaySettingsModel{}
	if api.Placeholder != nil {
		ds.Placeholder = types.StringValue(*api.Placeholder)
	}
	if api.HideActionBar != nil {
		ds.HideActionBar = types.BoolValue(*api.HideActionBar)
	}
	if api.HideExclude != nil {
		ds.HideExclude = types.BoolValue(*api.HideExclude)
	}
	if api.HideExists != nil {
		ds.HideExists = types.BoolValue(*api.HideExists)
	}
	if api.HideSort != nil {
		ds.HideSort = types.BoolValue(*api.HideSort)
	}
	return ds
}
