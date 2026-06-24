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

package optionslist

import (
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PopulateFromAPI reads back an options list control config from the API
// response and updates the panel model. Null-preservation semantics apply: if a field is
// null in the existing TF state (pm.OptionsListControlConfig), we do not overwrite it with
// a Kibana-returned value. If there is no existing config block, and Kibana returns an
// empty/absent config, we leave OptionsListControlConfig as nil.
//
// tfPanel is the prior TF state/plan panel, or nil on import.
func PopulateFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, ol *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl) diag.Diagnostics {
	if ol == nil {
		return nil
	}
	// The control config is a discriminated union (field-based vs ES|QL).
	// The TF model only describes the field-based variant; a valid ES|QL config
	// is left unchanged. If the config matches neither variant it is malformed
	// and surfaced as an error.
	apiConfig, err := ol.Config.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField()
	if err != nil {
		if _, esqlErr := ol.Config.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql(); esqlErr != nil {
			var diags diag.Diagnostics
			diags.AddError("Failed to decode options list control config", err.Error())
			return diags
		}
		return nil
	}
	existing := pm.OptionsListControlConfig

	if tfPanel == nil {
		// On import: populate from API. Required fields are always set; optional fields are
		// set only if present in the API response. Fields that Kibana returns as server-side
		// defaults (e.g. use_global_filters, exclude, sort) are treated as optional and left
		// null when absent from the user's configuration — matching the null-preservation
		// behaviour used during normal apply reads.
		pm.OptionsListControlConfig = &models.OptionsListControlConfigModel{
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
		return nil
	}

	if existing == nil {
		if tfPanel == nil || tfPanel.OptionsListControlConfig == nil {
			return nil
		}
		pm.OptionsListControlConfig = &models.OptionsListControlConfigModel{
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

	if tfPanel != nil && tfPanel.OptionsListControlConfig != nil {
		optionsListPreserveNullIntentFromPrior(tfPanel.OptionsListControlConfig, existing)
	}
	return nil
}

func optionsListPreserveNullIntentFromPrior(prior, existing *models.OptionsListControlConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	if !typeutils.IsKnown(prior.Title) {
		existing.Title = types.StringNull()
	}
	if !typeutils.IsKnown(prior.UseGlobalFilters) {
		existing.UseGlobalFilters = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.IgnoreValidations) {
		existing.IgnoreValidations = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.SingleSelect) {
		existing.SingleSelect = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.Exclude) {
		existing.Exclude = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.ExistsSelected) {
		existing.ExistsSelected = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.RunPastTimeout) {
		existing.RunPastTimeout = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.SearchTechnique) {
		existing.SearchTechnique = types.StringNull()
	}
	if !typeutils.IsKnown(prior.SelectedOptions) {
		existing.SelectedOptions = types.ListNull(types.StringType)
	}
	if prior.DisplaySettings == nil {
		existing.DisplaySettings = nil
	}
	if prior.Sort == nil {
		existing.Sort = nil
	}
}

// BuildConfig writes TF model fields into the API panel payload.
func BuildConfig(pm models.PanelModel, olPanel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl) {
	cfg := pm.OptionsListControlConfig
	if cfg == nil {
		return
	}

	var c kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField
	c.DataViewId = cfg.DataViewID.ValueString()
	c.FieldName = cfg.FieldName.ValueString()

	if typeutils.IsKnown(cfg.Title) {
		c.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.UseGlobalFilters) {
		c.UseGlobalFilters = cfg.UseGlobalFilters.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.IgnoreValidations) {
		c.IgnoreValidations = cfg.IgnoreValidations.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.SingleSelect) {
		c.SingleSelect = cfg.SingleSelect.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.Exclude) {
		c.Exclude = cfg.Exclude.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.ExistsSelected) {
		c.ExistsSelected = cfg.ExistsSelected.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.RunPastTimeout) {
		c.RunPastTimeout = cfg.RunPastTimeout.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.SearchTechnique) {
		st := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSearchTechnique(cfg.SearchTechnique.ValueString())
		c.SearchTechnique = &st
	}
	if !cfg.SelectedOptions.IsNull() && !cfg.SelectedOptions.IsUnknown() {
		elems := cfg.SelectedOptions.Elements()
		items := make([]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item, 0, len(elems))
		for _, e := range elems {
			s := e.(types.String).ValueString()
			var item kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item
			if err := item.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions0(s); err == nil {
				items = append(items, item)
			}
		}
		c.SelectedOptions = &items
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
		c.DisplaySettings = apiDS
	}
	if cfg.Sort != nil {
		c.Sort = &struct {
			By        kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSortBy        `json:"by"`
			Direction kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSortDirection `json:"direction"`
		}{
			By:        kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSortBy(cfg.Sort.By.ValueString()),
			Direction: kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSortDirection(cfg.Sort.Direction.ValueString()),
		}
	}
	olPanel.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(c)
}

func selectedOptionsToList(items []kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item) types.List {
	values := make([]attr.Value, 0, len(items))
	for _, item := range items {
		if s, err := item.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions0(); err == nil {
			values = append(values, types.StringValue(s))
			continue
		}
		if n, err := item.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions1(); err == nil {
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
}) *models.OptionsListControlDisplaySettingsModel {
	if api == nil {
		return nil
	}
	ds := &models.OptionsListControlDisplaySettingsModel{}
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
