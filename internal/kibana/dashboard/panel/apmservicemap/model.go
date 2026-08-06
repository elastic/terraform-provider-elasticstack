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

package apmservicemap

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const environmentServerDefault = "ENVIRONMENT_ALL"

// BuildConfig writes the TF model into the API panel struct.
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap) diag.Diagnostics {
	cfg := pm.ApmServiceMapConfig
	if cfg == nil {
		return nil
	}

	var diags diag.Diagnostics

	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&panel.Config.Title, &panel.Config.Description, &panel.Config.HideTitle, &panel.Config.HideBorder)
	if typeutils.IsKnown(cfg.Environment) {
		panel.Config.Environment = cfg.Environment.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.ServiceName) {
		panel.Config.ServiceName = cfg.ServiceName.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.ServiceGroupID) {
		panel.Config.ServiceGroupId = cfg.ServiceGroupID.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Kuery) {
		panel.Config.Kuery = cfg.Kuery.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.MapOrientation) {
		v := kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableMapOrientation(cfg.MapOrientation.ValueString())
		panel.Config.MapOrientation = &v
	}
	if typeutils.IsKnown(cfg.SyncWithDashboardFilters) {
		panel.Config.SyncWithDashboardFilters = cfg.SyncWithDashboardFilters.ValueBoolPointer()
	}

	if typeutils.IsKnown(cfg.AlertStatusFilter) {
		panel.Config.AlertStatusFilter = stringSetToEnumSlice[kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAlertStatusFilter](cfg.AlertStatusFilter, &diags)
	}
	if typeutils.IsKnown(cfg.AnomalySeverityFilter) {
		panel.Config.AnomalySeverityFilter = stringSetToEnumSlice[kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAnomalySeverityFilter](cfg.AnomalySeverityFilter, &diags)
	}
	if typeutils.IsKnown(cfg.ConnectionFilter) {
		panel.Config.ConnectionFilter = stringSetToEnumSlice[kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableConnectionFilter](cfg.ConnectionFilter, &diags)
	}
	if typeutils.IsKnown(cfg.SloStatusFilter) {
		panel.Config.SloStatusFilter = stringSetToEnumSlice[kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableSloStatusFilter](cfg.SloStatusFilter, &diags)
	}
	if cfg.TimeRange != nil {
		panel.Config.TimeRange = lenscommon.TimeRangeModelToAPI(cfg.TimeRange)
	}

	return diags
}

// PopulateFromAPI maps the Kibana API panel config into Terraform panel state while preserving
// prior null intent (REQ-009). prior is the prior TF state/plan panel, or nil on import.
//
// pm always arrives with ApmServiceMapConfig unset (callers build state from a zero-valued
// PanelModel to avoid aliasing plan pointers), so that field cannot be used to detect whether this
// panel was previously this same type. prior.ApmServiceMapConfig is the only reliable signal:
// non-nil means the panel was already this type and its null intent must be honored; nil means
// there is no prior null intent for this config block (creation, import, or a type change).
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap) diag.Diagnostics {
	cfg := apiPanel.Config

	if prior == nil || prior.ApmServiceMapConfig == nil {
		pm.ApmServiceMapConfig = apmServiceMapConfigFromAPIImport(cfg, true)
		return nil
	}

	// Same-type update: rebuild from the API, then reapply the prior config's null intent for any
	// optional field the plan/state had not set (REQ-009 null-preservation).
	existing := apmServiceMapConfigFromAPIImport(cfg, false)
	if existing == nil {
		pm.ApmServiceMapConfig = nil
		return nil
	}

	// BuildConfig omits both null and empty filter sets from the API payload (the API cannot
	// distinguish the two), so an empty-but-known prior set must be threaded through explicitly
	// here rather than derived from the freshly re-imported `existing` value; otherwise a
	// practitioner-configured `= []` would drift to null on every subsequent read/plan.
	priorCfg := prior.ApmServiceMapConfig
	existing.AlertStatusFilter = mergeStringSetFromEnum(cfg.AlertStatusFilter, priorCfg.AlertStatusFilter)
	existing.AnomalySeverityFilter = mergeStringSetFromEnum(cfg.AnomalySeverityFilter, priorCfg.AnomalySeverityFilter)
	existing.ConnectionFilter = mergeStringSetFromEnum(cfg.ConnectionFilter, priorCfg.ConnectionFilter)
	existing.SloStatusFilter = mergeStringSetFromEnum(cfg.SloStatusFilter, priorCfg.SloStatusFilter)

	existing.TimeRange = panelkit.MergeTimeRange(existing.TimeRange, cfg.TimeRange, priorCfg.TimeRange)
	apmServiceMapPreserveNullIntentFromPrior(priorCfg, existing)
	pm.ApmServiceMapConfig = existing
	return nil
}

func apmServiceMapConfigFromAPIImport(cfg kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable, suppressEnvironmentDefault bool) *models.ApmServiceMapConfigModel {
	if !apmServiceMapConfigHasAnyField(cfg, suppressEnvironmentDefault) {
		return nil
	}
	result := &models.ApmServiceMapConfigModel{
		Title:                    types.StringPointerValue(cfg.Title),
		Description:              types.StringPointerValue(cfg.Description),
		HideTitle:                types.BoolPointerValue(cfg.HideTitle),
		HideBorder:               types.BoolPointerValue(cfg.HideBorder),
		Environment:              types.StringPointerValue(cfg.Environment),
		ServiceName:              types.StringPointerValue(cfg.ServiceName),
		ServiceGroupID:           types.StringPointerValue(cfg.ServiceGroupId),
		Kuery:                    types.StringPointerValue(cfg.Kuery),
		MapOrientation:           mapOrientationFromAPI(cfg.MapOrientation),
		SyncWithDashboardFilters: types.BoolPointerValue(cfg.SyncWithDashboardFilters),
		AlertStatusFilter:        enumSliceToStringSet(cfg.AlertStatusFilter),
		AnomalySeverityFilter:    enumSliceToStringSet(cfg.AnomalySeverityFilter),
		ConnectionFilter:         enumSliceToStringSet(cfg.ConnectionFilter),
		SloStatusFilter:          enumSliceToStringSet(cfg.SloStatusFilter),
		TimeRange:                panelkit.TimeRangeFromAPI(cfg.TimeRange, nil),
	}
	if suppressEnvironmentDefault && result.Environment.ValueString() == environmentServerDefault {
		result.Environment = types.StringNull()
	}
	return result
}

func apmServiceMapConfigHasAnyField(cfg kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable, ignoreEnvironmentServerDefault bool) bool {
	hasEnvironment := cfg.Environment != nil
	if ignoreEnvironmentServerDefault && hasEnvironment && *cfg.Environment == environmentServerDefault {
		hasEnvironment = false
	}
	if cfg.Title != nil || cfg.Description != nil || cfg.HideTitle != nil || cfg.HideBorder != nil ||
		hasEnvironment || cfg.ServiceName != nil || cfg.ServiceGroupId != nil || cfg.Kuery != nil ||
		cfg.MapOrientation != nil || cfg.SyncWithDashboardFilters != nil {
		return true
	}
	if cfg.AlertStatusFilter != nil && len(*cfg.AlertStatusFilter) > 0 {
		return true
	}
	if cfg.AnomalySeverityFilter != nil && len(*cfg.AnomalySeverityFilter) > 0 {
		return true
	}
	if cfg.ConnectionFilter != nil && len(*cfg.ConnectionFilter) > 0 {
		return true
	}
	if cfg.SloStatusFilter != nil && len(*cfg.SloStatusFilter) > 0 {
		return true
	}
	if cfg.TimeRange != nil && (cfg.TimeRange.From != "" || cfg.TimeRange.To != "" ||
		(cfg.TimeRange.Mode != nil && string(*cfg.TimeRange.Mode) != "")) {
		return true
	}
	return false
}

// enumStringPointer converts a typed enum pointer to a *string, matching the pointer semantics
// (nil in, nil out) expected by types.StringPointerValue and panelkit.PreserveString.
func enumStringPointer[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func mapOrientationFromAPI(v *kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableMapOrientation) types.String {
	return types.StringPointerValue(enumStringPointer(v))
}

// stringSetToEnumSlice converts a validated Set of strings into a slice of the API's enum type,
// returning nil when the set has no elements (so the field is omitted from the API payload).
func stringSetToEnumSlice[T ~string](set types.Set, diags *diag.Diagnostics) *[]T {
	vals := typeutils.StringSetElements(set, diags)
	if len(vals) == 0 {
		return nil
	}
	out := make([]T, len(vals))
	for i, v := range vals {
		out[i] = T(v)
	}
	return &out
}

func enumSliceToStringSet[T ~string](slice *[]T) types.Set {
	if slice == nil || len(*slice) == 0 {
		return types.SetNull(types.StringType)
	}
	elems := make([]attr.Value, len(*slice))
	for i, v := range *slice {
		elems[i] = types.StringValue(string(v))
	}
	return types.SetValueMust(types.StringType, elems)
}

// mergeStringSetFromEnum reflects the API's values when it returns any (drift detection), and
// otherwise carries the prior config's set forward unchanged. The API cannot distinguish "unset"
// from "explicitly empty" (BuildConfig omits both), so falling back to the prior set — rather than
// to a value derived from the API alone — is what allows a known-empty set to round-trip correctly.
func mergeStringSetFromEnum[T ~string](api *[]T, priorSet types.Set) types.Set {
	if api != nil && len(*api) > 0 {
		return enumSliceToStringSet(api)
	}
	return priorSet
}

func apmServiceMapPreserveNullIntentFromPrior(prior, existing *models.ApmServiceMapConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	panelkit.NullPreservePresentationFromPrior(prior.Title, prior.Description, prior.HideTitle, prior.HideBorder,
		&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder)
	panelkit.NullPreserveStringFromPrior(prior.Environment, &existing.Environment)
	panelkit.NullPreserveStringFromPrior(prior.ServiceName, &existing.ServiceName)
	panelkit.NullPreserveStringFromPrior(prior.ServiceGroupID, &existing.ServiceGroupID)
	panelkit.NullPreserveStringFromPrior(prior.Kuery, &existing.Kuery)
	panelkit.NullPreserveStringFromPrior(prior.MapOrientation, &existing.MapOrientation)
	panelkit.NullPreserveBoolFromPrior(prior.SyncWithDashboardFilters, &existing.SyncWithDashboardFilters)
	panelkit.NullPreserveSetFromPrior(prior.AlertStatusFilter, &existing.AlertStatusFilter)
	panelkit.NullPreserveSetFromPrior(prior.AnomalySeverityFilter, &existing.AnomalySeverityFilter)
	panelkit.NullPreserveSetFromPrior(prior.ConnectionFilter, &existing.ConnectionFilter)
	panelkit.NullPreserveSetFromPrior(prior.SloStatusFilter, &existing.SloStatusFilter)
	existing.TimeRange = panelkit.PreserveTimeRangeNullIntentFromPrior(prior.TimeRange, existing.TimeRange)
}
