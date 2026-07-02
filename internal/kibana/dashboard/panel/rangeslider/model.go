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

package rangeslider

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PopulateFromAPI reads back a range slider control config from the API response and updates the
// panel model. The config is a discriminated union (Field vs ES|QL) with no explicit discriminator
// field on the wire; the variant is detected by probing for the `esql_query` key.
//
// Null-preservation semantics apply per branch: if an optional field is null in the existing TF
// state, it is not overwritten with a Kibana-returned value. If there is no existing config block,
// and Kibana returns an empty/absent config, RangeSliderControlConfig is left nil.
//
// tfPanel is the prior TF state/plan panel, or nil on import. When nil, the function populates all
// API-returned fields unconditionally (no prior intent to preserve).
func PopulateFromAPI(ctx context.Context, pm *models.PanelModel, tfPanel *models.PanelModel, rs *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl) diag.Diagnostics {
	if rs == nil {
		return nil
	}

	var diags diag.Diagnostics

	isImport := tfPanel == nil
	var priorCfg *models.RangeSliderControlConfigModel
	if !isImport {
		priorCfg = tfPanel.RangeSliderControlConfig
		if priorCfg == nil {
			// No prior intent to have this block configured; preserve nil.
			return nil
		}
	}

	raw, err := rs.Config.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to decode range slider control config", err.Error())
		return diags
	}

	// Reuse the existing config container (seeded from prior state) when one is available, rather
	// than allocating a new one, so unchanged reads preserve pointer identity.
	cfg := priorCfg
	if cfg == nil {
		cfg = &models.RangeSliderControlConfigModel{}
	}

	if isEsqlRangeSliderConfig(raw) {
		esqlCfg, err := rs.Config.AsKibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsql()
		if err != nil {
			diags.AddError("Failed to decode range slider control config", err.Error())
			return diags
		}
		var priorByEsql *models.RangeSliderControlByEsqlModel
		if priorCfg != nil {
			priorByEsql = priorCfg.ByEsql
		}
		cfg.ByField = nil
		cfg.ByEsql = populateByEsqlFromAPI(ctx, priorByEsql, &esqlCfg)
		pm.RangeSliderControlConfig = cfg
		return diags
	}

	fieldCfg, err := rs.Config.AsKibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaField()
	if err != nil {
		diags.AddError("Failed to decode range slider control config", err.Error())
		return diags
	}
	var priorByField *models.RangeSliderControlByFieldModel
	if priorCfg != nil {
		priorByField = priorCfg.ByField
	}
	cfg.ByEsql = nil
	cfg.ByField = populateByFieldFromAPI(ctx, priorByField, &fieldCfg)
	pm.RangeSliderControlConfig = cfg
	return diags
}

// isEsqlRangeSliderConfig reports whether the raw range slider config JSON is the ES|QL variant.
// The union has no explicit discriminator field, so presence of `esql_query` is used to distinguish
// it from the Field variant (`data_view_id` / `field_name`).
func isEsqlRangeSliderConfig(raw []byte) bool {
	var probe struct {
		EsqlQuery *string `json:"esql_query"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	return probe.EsqlQuery != nil
}

// rangeSliderSharedFields points at the by_field/by_esql model attributes shared by both branches
// (identical types on both the Field and ES|QL API schemas), letting populateByFieldFromAPI and
// populateByEsqlFromAPI apply the same null-preservation logic through a single implementation.
type rangeSliderSharedFields struct {
	Title             *types.String
	UseGlobalFilters  *types.Bool
	IgnoreValidations *types.Bool
	Value             *types.List
	Step              *types.Float32
}

// rangeSliderSharedAPIFields mirrors the attributes common to both
// KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchema{Field,Esql}, which share identical Go
// types for every shared attribute.
type rangeSliderSharedAPIFields struct {
	Title             *string
	UseGlobalFilters  *bool
	IgnoreValidations *bool
	Value             *[]string
	Step              *float32
}

// populateRangeSliderSharedFields applies REQ-009 null-preservation semantics to the attributes
// shared by both by_field and by_esql: on import (prior == nil), every optional field is populated
// unconditionally, defaulting absent API values to null. Otherwise, only fields previously known in
// state are updated from the API response; fields that were null in prior state remain null, and
// fields the API stops returning keep their prior value rather than resetting to null.
func populateRangeSliderSharedFields(ctx context.Context, prior *rangeSliderSharedFields, m rangeSliderSharedFields, api rangeSliderSharedAPIFields) {
	if prior == nil {
		*m.Title = types.StringPointerValue(api.Title)
		*m.UseGlobalFilters = types.BoolPointerValue(api.UseGlobalFilters)
		*m.IgnoreValidations = types.BoolPointerValue(api.IgnoreValidations)
		*m.Value = valueListFromAPI(ctx, api.Value)
		*m.Step = float32PointerValue(api.Step)
		return
	}

	*m.Title = types.StringNull()
	if typeutils.IsKnown(*prior.Title) {
		*m.Title = *prior.Title
		if api.Title != nil {
			*m.Title = types.StringValue(*api.Title)
		}
	}
	*m.UseGlobalFilters = types.BoolNull()
	if typeutils.IsKnown(*prior.UseGlobalFilters) {
		*m.UseGlobalFilters = *prior.UseGlobalFilters
		if api.UseGlobalFilters != nil {
			*m.UseGlobalFilters = types.BoolValue(*api.UseGlobalFilters)
		}
	}
	*m.IgnoreValidations = types.BoolNull()
	if typeutils.IsKnown(*prior.IgnoreValidations) {
		*m.IgnoreValidations = *prior.IgnoreValidations
		if api.IgnoreValidations != nil {
			*m.IgnoreValidations = types.BoolValue(*api.IgnoreValidations)
		}
	}
	*m.Value = types.ListNull(types.StringType)
	if typeutils.IsKnown(*prior.Value) {
		*m.Value = *prior.Value
		if api.Value != nil {
			*m.Value = valueListFromAPI(ctx, api.Value)
		}
	}
	*m.Step = types.Float32Null()
	if typeutils.IsKnown(*prior.Step) {
		*m.Step = *prior.Step
		if api.Step != nil {
			*m.Step = types.Float32Value(*api.Step)
		}
	}
}

func populateByFieldFromAPI(
	ctx context.Context,
	prior *models.RangeSliderControlByFieldModel,
	api *kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaField,
) *models.RangeSliderControlByFieldModel {
	m := &models.RangeSliderControlByFieldModel{
		DataViewID: types.StringValue(api.DataViewId),
		FieldName:  types.StringValue(api.FieldName),
	}
	var priorFields *rangeSliderSharedFields
	if prior != nil {
		priorFields = &rangeSliderSharedFields{&prior.Title, &prior.UseGlobalFilters, &prior.IgnoreValidations, &prior.Value, &prior.Step}
	}
	populateRangeSliderSharedFields(
		ctx, priorFields,
		rangeSliderSharedFields{&m.Title, &m.UseGlobalFilters, &m.IgnoreValidations, &m.Value, &m.Step},
		rangeSliderSharedAPIFields{api.Title, api.UseGlobalFilters, api.IgnoreValidations, api.Value, api.Step},
	)
	return m
}

func populateByEsqlFromAPI(
	ctx context.Context,
	prior *models.RangeSliderControlByEsqlModel,
	api *kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsql,
) *models.RangeSliderControlByEsqlModel {
	m := &models.RangeSliderControlByEsqlModel{
		EsqlQuery: types.StringValue(api.EsqlQuery),
		// The wire enum only ever legally carries "esql" (see
		// kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsqlValuesSourceEsql). We
		// normalize it back to the Terraform-facing constant so plans stay drift-free.
		ValuesSource: types.StringValue(esqlValuesSourceUserValue),
	}
	var priorFields *rangeSliderSharedFields
	if prior != nil {
		priorFields = &rangeSliderSharedFields{&prior.Title, &prior.UseGlobalFilters, &prior.IgnoreValidations, &prior.Value, &prior.Step}
	}
	populateRangeSliderSharedFields(
		ctx, priorFields,
		rangeSliderSharedFields{&m.Title, &m.UseGlobalFilters, &m.IgnoreValidations, &m.Value, &m.Step},
		rangeSliderSharedAPIFields{api.Title, api.UseGlobalFilters, api.IgnoreValidations, api.Value, api.Step},
	)
	return m
}

func valueListFromAPI(ctx context.Context, v *[]string) types.List {
	if v == nil {
		return types.ListNull(types.StringType)
	}
	l, diags := types.ListValueFrom(ctx, types.StringType, *v)
	if diags.HasError() {
		return types.ListNull(types.StringType)
	}
	return l
}

func float32PointerValue(v *float32) types.Float32 {
	if v == nil {
		return types.Float32Null()
	}
	return types.Float32Value(*v)
}

// BuildConfig writes TF model fields into the API panel payload, dispatching on whichever branch
// (ByField or ByEsql) is set on the panel's range_slider_control_config.
func BuildConfig(pm models.PanelModel, rsPanel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl) diag.Diagnostics {
	cfg := pm.RangeSliderControlConfig
	if cfg == nil {
		return nil
	}

	switch {
	case cfg.ByField != nil:
		return buildFieldConfig(cfg.ByField, rsPanel)
	case cfg.ByEsql != nil:
		return buildEsqlConfig(cfg.ByEsql, rsPanel)
	default:
		var diags diag.Diagnostics
		diags.AddError(
			"Invalid range_slider_control_config",
			"Exactly one of `by_field` or `by_esql` must be set inside `range_slider_control_config`.",
		)
		return diags
	}
}

func buildFieldConfig(cfg *models.RangeSliderControlByFieldModel, rsPanel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl) diag.Diagnostics {
	var c kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaField
	c.DataViewId = cfg.DataViewID.ValueString()
	c.FieldName = cfg.FieldName.ValueString()

	// values_source is not exposed on `by_field` and is deliberately left unset on the wire: Kibana
	// treats it as "field" when absent (its default for legacy controls, per design D2), and Kibana
	// versions below the values_source-discriminated-union schema (see
	// dashboardacctest.MinControlByFieldEsqlUnionSupport) reject the property entirely if present.
	// Omitting it keeps by_field writes compatible with every Kibana version this resource supports.

	if typeutils.IsKnown(cfg.Title) {
		c.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.UseGlobalFilters) {
		c.UseGlobalFilters = cfg.UseGlobalFilters.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.IgnoreValidations) {
		c.IgnoreValidations = cfg.IgnoreValidations.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.Value) {
		c.Value = stringListToAPI(cfg.Value)
	}
	if typeutils.IsKnown(cfg.Step) {
		v := cfg.Step.ValueFloat32()
		c.Step = &v
	}

	if err := rsPanel.Config.FromKibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaField(c); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to build range slider control config", err.Error())
		return diags
	}
	return nil
}

func buildEsqlConfig(cfg *models.RangeSliderControlByEsqlModel, rsPanel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl) diag.Diagnostics {
	var c kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsql
	c.EsqlQuery = cfg.EsqlQuery.ValueString()
	// The schema validates the user-facing values_source to a single legal value
	// (esqlValuesSourceUserValue). The wire enum's only legal value is "esql" (see
	// kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsqlValuesSourceEsql), so it is
	// hardcoded here rather than derived from the (already-validated) model string.
	c.ValuesSource = kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsqlValuesSourceEsql

	if typeutils.IsKnown(cfg.Title) {
		c.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.UseGlobalFilters) {
		c.UseGlobalFilters = cfg.UseGlobalFilters.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.IgnoreValidations) {
		c.IgnoreValidations = cfg.IgnoreValidations.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.Value) {
		c.Value = stringListToAPI(cfg.Value)
	}
	if typeutils.IsKnown(cfg.Step) {
		v := cfg.Step.ValueFloat32()
		c.Step = &v
	}

	if err := rsPanel.Config.FromKibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsql(c); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to build range slider control config", err.Error())
		return diags
	}
	return nil
}

func stringListToAPI(l types.List) *[]string {
	raw := l.Elements()
	elems := make([]string, len(raw))
	for i, e := range raw {
		elems[i] = e.(types.String).ValueString()
	}
	return &elems
}
