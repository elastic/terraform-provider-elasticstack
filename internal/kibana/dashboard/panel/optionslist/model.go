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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// displaySettingsAPI mirrors the identical anonymous display_settings struct shape shared by both
// the Field and ES|QL options list control API schemas (same field names, types, and JSON tags),
// letting both branches reuse the same conversion helpers.
type displaySettingsAPI = struct {
	HideActionBar *bool   `json:"hide_action_bar,omitempty"`
	HideExclude   *bool   `json:"hide_exclude,omitempty"`
	HideExists    *bool   `json:"hide_exists,omitempty"`
	HideSort      *bool   `json:"hide_sort,omitempty"`
	Placeholder   *string `json:"placeholder,omitempty"`
}

// PopulateFromAPI reads back an options list control config from the API response and updates the
// panel model. The control config is a discriminated union (Field vs ES|QL branch); the branch is
// detected by inspecting the raw API JSON for the `esql_query` key, which is only present on the
// ES|QL branch. Null-preservation semantics (REQ-009) apply to optional booleans within both
// branches: if a field is null in the existing TF state, it is not overwritten with a
// Kibana-returned value.
//
// tfPanel is the prior TF state/plan panel, or nil on import.
func PopulateFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, ol *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl) diag.Diagnostics {
	if ol == nil {
		return nil
	}

	var diags diag.Diagnostics
	raw, err := ol.Config.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to decode options list control config", err.Error())
		return diags
	}

	if panelkit.IsEsqlBranch(raw) {
		apiConfig, err := ol.Config.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql()
		if err != nil {
			diags.AddError("Failed to decode options list control config", err.Error())
			return diags
		}
		return populateEsqlFromAPI(pm, tfPanel, apiConfig)
	}

	apiConfig, err := ol.Config.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField()
	if err != nil {
		diags.AddError("Failed to decode options list control config", err.Error())
		return diags
	}
	return populateFieldFromAPI(pm, tfPanel, apiConfig)
}

// preserveKnownBool updates existing from api only when existing is already known (REQ-009
// null-preservation); an api value of nil, or existing being null/unknown, leaves existing
// unchanged.
func preserveKnownBool(existing types.Bool, api *bool) types.Bool {
	if typeutils.IsKnown(existing) && api != nil {
		return types.BoolValue(*api)
	}
	return existing
}

// preserveKnownString is the string equivalent of preserveKnownBool.
func preserveKnownString(existing types.String, api *string) types.String {
	if typeutils.IsKnown(existing) && api != nil {
		return types.StringValue(*api)
	}
	return existing
}

// sharedOptionsListAPIFields holds optional field values extracted from either the Field or ES|QL
// API config variant into a branch-neutral form, so populateSharedOptionsListFieldsFromAPI can
// apply them without knowing which branch produced them.
type sharedOptionsListAPIFields struct {
	Title             *string
	UseGlobalFilters  *bool
	IgnoreValidations *bool
	SingleSelect      *bool
	Exclude           *bool
	ExistsSelected    *bool
	RunPastTimeout    *bool
	SearchTechnique   *string     // pre-converted from the branch-specific named enum type
	SelectedOptions   *types.List // pre-converted via the branch-specific selectedOptions helper
	DisplaySettings   *displaySettingsAPI
	SortBy            *string // both nil when Sort is absent in the API response
	SortDirection     *string
}

func sharedAPIFieldsFromField(apiConfig kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField) sharedOptionsListAPIFields {
	f := sharedOptionsListAPIFields{
		Title:             apiConfig.Title,
		UseGlobalFilters:  apiConfig.UseGlobalFilters,
		IgnoreValidations: apiConfig.IgnoreValidations,
		SingleSelect:      apiConfig.SingleSelect,
		Exclude:           apiConfig.Exclude,
		ExistsSelected:    apiConfig.ExistsSelected,
		RunPastTimeout:    apiConfig.RunPastTimeout,
		DisplaySettings:   apiConfig.DisplaySettings,
	}
	if apiConfig.SearchTechnique != nil {
		st := string(*apiConfig.SearchTechnique)
		f.SearchTechnique = &st
	}
	if apiConfig.SelectedOptions != nil {
		list := selectedOptionsFieldToList(*apiConfig.SelectedOptions)
		f.SelectedOptions = &list
	}
	if apiConfig.Sort != nil {
		by := string(apiConfig.Sort.By)
		dir := string(apiConfig.Sort.Direction)
		f.SortBy, f.SortDirection = &by, &dir
	}
	return f
}

func sharedAPIFieldsFromEsql(apiConfig kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql) sharedOptionsListAPIFields {
	f := sharedOptionsListAPIFields{
		Title:             apiConfig.Title,
		UseGlobalFilters:  apiConfig.UseGlobalFilters,
		IgnoreValidations: apiConfig.IgnoreValidations,
		SingleSelect:      apiConfig.SingleSelect,
		Exclude:           apiConfig.Exclude,
		ExistsSelected:    apiConfig.ExistsSelected,
		RunPastTimeout:    apiConfig.RunPastTimeout,
		DisplaySettings:   apiConfig.DisplaySettings,
	}
	if apiConfig.SearchTechnique != nil {
		st := string(*apiConfig.SearchTechnique)
		f.SearchTechnique = &st
	}
	if apiConfig.SelectedOptions != nil {
		list := selectedOptionsEsqlToList(*apiConfig.SelectedOptions)
		f.SelectedOptions = &list
	}
	if apiConfig.Sort != nil {
		by := string(apiConfig.Sort.By)
		dir := string(apiConfig.Sort.Direction)
		f.SortBy, f.SortDirection = &by, &dir
	}
	return f
}

// populateSharedOptionsListFieldsFromAPI applies the optional field updates that are identical
// across the Field and ES|QL populate functions. model points into the branch-specific struct and
// api carries pre-processed values from the matching sharedAPIFieldsFromX adapter.
func populateSharedOptionsListFieldsFromAPI(model optionsListNullIntentFields, api sharedOptionsListAPIFields) {
	*model.Title = preserveKnownString(*model.Title, api.Title)
	*model.UseGlobalFilters = preserveKnownBool(*model.UseGlobalFilters, api.UseGlobalFilters)
	*model.IgnoreValidations = preserveKnownBool(*model.IgnoreValidations, api.IgnoreValidations)
	*model.SingleSelect = preserveKnownBool(*model.SingleSelect, api.SingleSelect)
	*model.Exclude = preserveKnownBool(*model.Exclude, api.Exclude)
	*model.ExistsSelected = preserveKnownBool(*model.ExistsSelected, api.ExistsSelected)
	*model.RunPastTimeout = preserveKnownBool(*model.RunPastTimeout, api.RunPastTimeout)
	if typeutils.IsKnown(*model.SearchTechnique) && api.SearchTechnique != nil {
		*model.SearchTechnique = types.StringValue(*api.SearchTechnique)
	}
	if typeutils.IsKnown(*model.SelectedOptions) && api.SelectedOptions != nil {
		*model.SelectedOptions = *api.SelectedOptions
	}
	if *model.DisplaySettings != nil && api.DisplaySettings != nil {
		updateDisplaySettingsFromAPI(*model.DisplaySettings, api.DisplaySettings)
	}
	if *model.Sort != nil && api.SortBy != nil {
		(*model.Sort).By = types.StringValue(*api.SortBy)
		(*model.Sort).Direction = types.StringValue(*api.SortDirection)
	}
}

// populateFieldFromAPI populates pm.OptionsListControlConfig.ByField from a Field-branch API
// response, applying the same null-preservation semantics as the pre-union implementation.
func populateFieldFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField) diag.Diagnostics {
	var existing *models.OptionsListControlByFieldModel
	if pm.OptionsListControlConfig != nil {
		existing = pm.OptionsListControlConfig.ByField
	}

	if tfPanel == nil {
		// Import: required fields and user-configurable optional fields are populated; optional
		// booleans and sort are left null since Kibana returns server-side defaults for them that
		// would otherwise cause post-import drift.
		pm.OptionsListControlConfig = &models.OptionsListControlConfigModel{
			ByField: newOptionsListFieldFromRequiredAndPresent(apiConfig),
		}
		return nil
	}

	if existing == nil {
		if tfPanel.OptionsListControlConfig == nil {
			// No prior intent to have this control configured at all; preserve nil.
			return nil
		}
		// Either there was no prior by_field state, or the remote control switched branches
		// out-of-band (e.g. from by_esql to by_field); build a fresh by_field branch from the
		// API response so state reflects the actual remote configuration instead of silently
		// keeping a stale by_esql (or empty) block.
		existing = newOptionsListFieldFromRequiredAndPresent(apiConfig)
		pm.OptionsListControlConfig = &models.OptionsListControlConfigModel{ByField: existing}
	}

	// Block exists in state — update required fields unconditionally, optional fields only when known.
	existing.DataViewID = types.StringValue(apiConfig.DataViewId)
	existing.FieldName = types.StringValue(apiConfig.FieldName)
	populateSharedOptionsListFieldsFromAPI(
		optionsListNullIntentFields{
			Title: &existing.Title, UseGlobalFilters: &existing.UseGlobalFilters,
			IgnoreValidations: &existing.IgnoreValidations, SingleSelect: &existing.SingleSelect,
			Exclude: &existing.Exclude, ExistsSelected: &existing.ExistsSelected,
			RunPastTimeout: &existing.RunPastTimeout, SearchTechnique: &existing.SearchTechnique,
			SelectedOptions: &existing.SelectedOptions, DisplaySettings: &existing.DisplaySettings,
			Sort: &existing.Sort,
		},
		sharedAPIFieldsFromField(apiConfig),
	)

	if tfPanel.OptionsListControlConfig != nil {
		preserveOptionsListFieldNullIntentFromPrior(tfPanel.OptionsListControlConfig.ByField, existing)
	}
	return nil
}

// populateEsqlFromAPI populates pm.OptionsListControlConfig.ByEsql from an ES|QL-branch API
// response, mirroring the null-preservation semantics used for the Field branch (REQ-009 /D5).
func populateEsqlFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql) diag.Diagnostics {
	var existing *models.OptionsListControlByEsqlModel
	if pm.OptionsListControlConfig != nil {
		existing = pm.OptionsListControlConfig.ByEsql
	}

	if tfPanel == nil {
		pm.OptionsListControlConfig = &models.OptionsListControlConfigModel{
			ByEsql: newOptionsListEsqlFromRequiredAndPresent(apiConfig),
		}
		return nil
	}

	if existing == nil {
		if tfPanel.OptionsListControlConfig == nil {
			// No prior intent to have this control configured at all; preserve nil.
			return nil
		}
		// Either there was no prior by_esql state, or the remote control switched branches
		// out-of-band (e.g. from by_field to by_esql); build a fresh by_esql branch from the
		// API response so state reflects the actual remote configuration instead of silently
		// keeping a stale by_field (or empty) block.
		existing = newOptionsListEsqlFromRequiredAndPresent(apiConfig)
		pm.OptionsListControlConfig = &models.OptionsListControlConfigModel{ByEsql: existing}
	}

	// Block exists in state — update required fields unconditionally, optional fields only when known.
	existing.EsqlQuery = types.StringValue(apiConfig.EsqlQuery)
	existing.ValuesSource = types.StringValue(panelkit.EsqlValuesSourceUserValue)
	populateSharedOptionsListFieldsFromAPI(
		optionsListNullIntentFields{
			Title: &existing.Title, UseGlobalFilters: &existing.UseGlobalFilters,
			IgnoreValidations: &existing.IgnoreValidations, SingleSelect: &existing.SingleSelect,
			Exclude: &existing.Exclude, ExistsSelected: &existing.ExistsSelected,
			RunPastTimeout: &existing.RunPastTimeout, SearchTechnique: &existing.SearchTechnique,
			SelectedOptions: &existing.SelectedOptions, DisplaySettings: &existing.DisplaySettings,
			Sort: &existing.Sort,
		},
		sharedAPIFieldsFromEsql(apiConfig),
	)

	if tfPanel.OptionsListControlConfig != nil {
		preserveOptionsListEsqlNullIntentFromPrior(tfPanel.OptionsListControlConfig.ByEsql, existing)
	}
	return nil
}

// newOptionsListFieldFromRequiredAndPresent builds a fresh by_field model populating required
// fields unconditionally and optional user-configurable fields (title, single_select,
// search_technique, selected_options, display_settings) only when present in the API response.
// Optional booleans and sort are intentionally left null (see populateFieldFromAPI).
func newOptionsListFieldFromRequiredAndPresent(apiConfig kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField) *models.OptionsListControlByFieldModel {
	m := &models.OptionsListControlByFieldModel{
		DataViewID:      types.StringValue(apiConfig.DataViewId),
		FieldName:       types.StringValue(apiConfig.FieldName),
		SelectedOptions: types.ListNull(types.StringType),
	}
	if apiConfig.Title != nil {
		m.Title = types.StringValue(*apiConfig.Title)
	}
	if apiConfig.SingleSelect != nil {
		m.SingleSelect = types.BoolValue(*apiConfig.SingleSelect)
	}
	if apiConfig.SearchTechnique != nil {
		m.SearchTechnique = types.StringValue(string(*apiConfig.SearchTechnique))
	}
	if apiConfig.SelectedOptions != nil {
		m.SelectedOptions = selectedOptionsFieldToList(*apiConfig.SelectedOptions)
	}
	if apiConfig.DisplaySettings != nil {
		m.DisplaySettings = displaySettingsFromAPI(apiConfig.DisplaySettings)
	}
	return m
}

// newOptionsListEsqlFromRequiredAndPresent is the ES|QL-branch analog of
// newOptionsListFieldFromRequiredAndPresent.
func newOptionsListEsqlFromRequiredAndPresent(apiConfig kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql) *models.OptionsListControlByEsqlModel {
	m := &models.OptionsListControlByEsqlModel{
		EsqlQuery: types.StringValue(apiConfig.EsqlQuery),
		// The wire enum only ever legally carries "esql" (see
		// kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlValuesSourceEsql).
		// The Terraform-facing attribute always reads back as panelkit.EsqlValuesSourceUserValue.
		ValuesSource:    types.StringValue(panelkit.EsqlValuesSourceUserValue),
		SelectedOptions: types.ListNull(types.StringType),
	}
	if apiConfig.Title != nil {
		m.Title = types.StringValue(*apiConfig.Title)
	}
	if apiConfig.SingleSelect != nil {
		m.SingleSelect = types.BoolValue(*apiConfig.SingleSelect)
	}
	if apiConfig.SearchTechnique != nil {
		m.SearchTechnique = types.StringValue(string(*apiConfig.SearchTechnique))
	}
	if apiConfig.SelectedOptions != nil {
		m.SelectedOptions = selectedOptionsEsqlToList(*apiConfig.SelectedOptions)
	}
	if apiConfig.DisplaySettings != nil {
		m.DisplaySettings = displaySettingsFromAPI(apiConfig.DisplaySettings)
	}
	return m
}

// optionsListNullIntentFields points at the by_field/by_esql model attributes subject to
// REQ-009 null-preservation, letting preserveOptionsListFieldNullIntentFromPrior and
// preserveOptionsListEsqlNullIntentFromPrior share one implementation despite operating on the two
// distinct branch model types (both use identical Go types for every field below).
type optionsListNullIntentFields struct {
	Title             *types.String
	UseGlobalFilters  *types.Bool
	IgnoreValidations *types.Bool
	SingleSelect      *types.Bool
	Exclude           *types.Bool
	ExistsSelected    *types.Bool
	RunPastTimeout    *types.Bool
	SearchTechnique   *types.String
	SelectedOptions   *types.List
	DisplaySettings   **models.OptionsListControlDisplaySettingsModel
	Sort              **models.OptionsListControlSortModel
}

// preserveOptionsListNullIntentFromPrior resets each field on existing to null/nil when the
// corresponding prior field was null/nil, so fields never explicitly configured by the user stay
// null instead of drifting to a Kibana-returned value.
func preserveOptionsListNullIntentFromPrior(prior, existing optionsListNullIntentFields) {
	if !typeutils.IsKnown(*prior.Title) {
		*existing.Title = types.StringNull()
	}
	if !typeutils.IsKnown(*prior.UseGlobalFilters) {
		*existing.UseGlobalFilters = types.BoolNull()
	}
	if !typeutils.IsKnown(*prior.IgnoreValidations) {
		*existing.IgnoreValidations = types.BoolNull()
	}
	if !typeutils.IsKnown(*prior.SingleSelect) {
		*existing.SingleSelect = types.BoolNull()
	}
	if !typeutils.IsKnown(*prior.Exclude) {
		*existing.Exclude = types.BoolNull()
	}
	if !typeutils.IsKnown(*prior.ExistsSelected) {
		*existing.ExistsSelected = types.BoolNull()
	}
	if !typeutils.IsKnown(*prior.RunPastTimeout) {
		*existing.RunPastTimeout = types.BoolNull()
	}
	if !typeutils.IsKnown(*prior.SearchTechnique) {
		*existing.SearchTechnique = types.StringNull()
	}
	if !typeutils.IsKnown(*prior.SelectedOptions) {
		*existing.SelectedOptions = types.ListNull(types.StringType)
	}
	if *prior.DisplaySettings == nil {
		*existing.DisplaySettings = nil
	}
	if *prior.Sort == nil {
		*existing.Sort = nil
	}
}

func preserveOptionsListFieldNullIntentFromPrior(prior, existing *models.OptionsListControlByFieldModel) {
	if prior == nil || existing == nil {
		return
	}
	preserveOptionsListNullIntentFromPrior(
		optionsListNullIntentFields{
			&prior.Title, &prior.UseGlobalFilters, &prior.IgnoreValidations, &prior.SingleSelect,
			&prior.Exclude, &prior.ExistsSelected, &prior.RunPastTimeout, &prior.SearchTechnique,
			&prior.SelectedOptions, &prior.DisplaySettings, &prior.Sort,
		},
		optionsListNullIntentFields{
			&existing.Title, &existing.UseGlobalFilters, &existing.IgnoreValidations, &existing.SingleSelect,
			&existing.Exclude, &existing.ExistsSelected, &existing.RunPastTimeout, &existing.SearchTechnique,
			&existing.SelectedOptions, &existing.DisplaySettings, &existing.Sort,
		},
	)
}

func preserveOptionsListEsqlNullIntentFromPrior(prior, existing *models.OptionsListControlByEsqlModel) {
	if prior == nil || existing == nil {
		return
	}
	preserveOptionsListNullIntentFromPrior(
		optionsListNullIntentFields{
			&prior.Title, &prior.UseGlobalFilters, &prior.IgnoreValidations, &prior.SingleSelect,
			&prior.Exclude, &prior.ExistsSelected, &prior.RunPastTimeout, &prior.SearchTechnique,
			&prior.SelectedOptions, &prior.DisplaySettings, &prior.Sort,
		},
		optionsListNullIntentFields{
			&existing.Title, &existing.UseGlobalFilters, &existing.IgnoreValidations, &existing.SingleSelect,
			&existing.Exclude, &existing.ExistsSelected, &existing.RunPastTimeout, &existing.SearchTechnique,
			&existing.SelectedOptions, &existing.DisplaySettings, &existing.Sort,
		},
	)
}

// BuildConfig writes TF model fields into the API panel payload, dispatching to the by_field or
// by_esql branch builder depending on which is set on the model (exactly one is guaranteed by the
// schema-level ExactlyOneOf validator).
func BuildConfig(pm models.PanelModel, olPanel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl) diag.Diagnostics {
	cfg := pm.OptionsListControlConfig
	if cfg == nil {
		return nil
	}

	switch {
	case cfg.ByField != nil:
		return buildFieldConfig(cfg.ByField, olPanel)
	case cfg.ByEsql != nil:
		return buildEsqlConfig(cfg.ByEsql, olPanel)
	default:
		var diags diag.Diagnostics
		diags.AddError(
			"Invalid options_list_control_config",
			"Exactly one of `by_field` or `by_esql` must be set inside `options_list_control_config`.",
		)
		return diags
	}
}

// buildFieldConfig writes the by_field branch into the API payload. values_source is not exposed as
// a user-facing attribute for this branch, and is deliberately left unset on the wire: Kibana treats
// it as "field" when absent (its default for legacy controls, per design D2), and Kibana versions
// below the values_source-discriminated-union schema (see
// dashboardacctest.MinControlByFieldEsqlUnionSupport) reject the property entirely if present.
// Omitting it keeps by_field writes compatible with every Kibana version this resource supports.
func buildFieldConfig(cfg *models.OptionsListControlByFieldModel, olPanel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl) diag.Diagnostics {
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
		c.SelectedOptions = buildSelectedOptionsField(cfg.SelectedOptions)
	}
	if cfg.DisplaySettings != nil {
		c.DisplaySettings = buildDisplaySettingsAPI(cfg.DisplaySettings)
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
	if err := olPanel.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(c); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to build options list control config", err.Error())
		return diags
	}
	return nil
}

// buildEsqlConfig writes the by_esql branch into the API payload. values_source is schema-validated
// to be panelkit.EsqlValuesSourceUserValue ("esql_query") but the wire enum's only legal value is "esql" (see
// kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlValuesSourceEsql), so the
// wire constant is always sent regardless of the model value.
func buildEsqlConfig(cfg *models.OptionsListControlByEsqlModel, olPanel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl) diag.Diagnostics {
	var c kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql
	c.EsqlQuery = cfg.EsqlQuery.ValueString()
	c.ValuesSource = kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlValuesSourceEsql

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
		st := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSearchTechnique(cfg.SearchTechnique.ValueString())
		c.SearchTechnique = &st
	}
	if !cfg.SelectedOptions.IsNull() && !cfg.SelectedOptions.IsUnknown() {
		c.SelectedOptions = buildSelectedOptionsEsql(cfg.SelectedOptions)
	}
	if cfg.DisplaySettings != nil {
		c.DisplaySettings = buildDisplaySettingsAPI(cfg.DisplaySettings)
	}
	if cfg.Sort != nil {
		c.Sort = &struct {
			By        kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSortBy        `json:"by"`
			Direction kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSortDirection `json:"direction"`
		}{
			By:        kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSortBy(cfg.Sort.By.ValueString()),
			Direction: kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSortDirection(cfg.Sort.Direction.ValueString()),
		}
	}
	if err := olPanel.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql(c); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to build options list control config", err.Error())
		return diags
	}
	return nil
}

func selectedOptionsFieldToList(items []kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item) types.List {
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

func selectedOptionsEsqlToList(items []kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql_SelectedOptions_Item) types.List {
	values := make([]attr.Value, 0, len(items))
	for _, item := range items {
		if s, err := item.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSelectedOptions0(); err == nil {
			values = append(values, types.StringValue(s))
			continue
		}
		if n, err := item.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSelectedOptions1(); err == nil {
			values = append(values, types.StringValue(strconv.FormatFloat(float64(n), 'f', -1, 32)))
		}
	}
	return types.ListValueMust(types.StringType, values)
}

func buildSelectedOptionsField(list types.List) *[]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item {
	elems := list.Elements()
	items := make([]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item, 0, len(elems))
	for _, e := range elems {
		s := e.(types.String).ValueString()
		var item kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item
		if err := item.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions0(s); err == nil {
			items = append(items, item)
		}
	}
	return &items
}

func buildSelectedOptionsEsql(list types.List) *[]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql_SelectedOptions_Item {
	elems := list.Elements()
	items := make([]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql_SelectedOptions_Item, 0, len(elems))
	for _, e := range elems {
		s := e.(types.String).ValueString()
		var item kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql_SelectedOptions_Item
		if err := item.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSelectedOptions0(s); err == nil {
			items = append(items, item)
		}
	}
	return &items
}

func displaySettingsFromAPI(api *displaySettingsAPI) *models.OptionsListControlDisplaySettingsModel {
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

func updateDisplaySettingsFromAPI(ds *models.OptionsListControlDisplaySettingsModel, api *displaySettingsAPI) {
	if ds == nil || api == nil {
		return
	}
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

func buildDisplaySettingsAPI(ds *models.OptionsListControlDisplaySettingsModel) *displaySettingsAPI {
	apiDS := &displaySettingsAPI{}
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
	return apiDS
}
