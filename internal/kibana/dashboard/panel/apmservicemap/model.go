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

// BuildConfig writes the TF model into the API panel struct.
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap) diag.Diagnostics {
	cfg := pm.ApmServiceMapConfig
	if cfg == nil {
		return nil
	}

	var diags diag.Diagnostics

	if typeutils.IsKnown(cfg.Title) {
		panel.Config.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Description) {
		panel.Config.Description = cfg.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.HideTitle) {
		panel.Config.HideTitle = cfg.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.HideBorder) {
		panel.Config.HideBorder = cfg.HideBorder.ValueBoolPointer()
	}
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
		if vals := typeutils.StringSetElements(cfg.AlertStatusFilter, &diags); len(vals) > 0 {
			out := make([]kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAlertStatusFilter, len(vals))
			for i, v := range vals {
				out[i] = kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAlertStatusFilter(v)
			}
			panel.Config.AlertStatusFilter = &out
		}
	}
	if typeutils.IsKnown(cfg.AnomalySeverityFilter) {
		if vals := typeutils.StringSetElements(cfg.AnomalySeverityFilter, &diags); len(vals) > 0 {
			out := make([]kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAnomalySeverityFilter, len(vals))
			for i, v := range vals {
				out[i] = kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableAnomalySeverityFilter(v)
			}
			panel.Config.AnomalySeverityFilter = &out
		}
	}
	if typeutils.IsKnown(cfg.ConnectionFilter) {
		if vals := typeutils.StringSetElements(cfg.ConnectionFilter, &diags); len(vals) > 0 {
			out := make([]kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableConnectionFilter, len(vals))
			for i, v := range vals {
				out[i] = kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableConnectionFilter(v)
			}
			panel.Config.ConnectionFilter = &out
		}
	}
	if typeutils.IsKnown(cfg.SloStatusFilter) {
		if vals := typeutils.StringSetElements(cfg.SloStatusFilter, &diags); len(vals) > 0 {
			out := make([]kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableSloStatusFilter, len(vals))
			for i, v := range vals {
				out[i] = kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableSloStatusFilter(v)
			}
			panel.Config.SloStatusFilter = &out
		}
	}
	if cfg.TimeRange != nil {
		panel.Config.TimeRange = lenscommon.TimeRangeModelToAPI(cfg.TimeRange)
	}

	return diags
}

// PopulateFromAPI reads back an APM service map panel from the API response.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeApmServiceMap) diag.Diagnostics {
	cfg := apiPanel.Config

	if prior == nil {
		pm.ApmServiceMapConfig = apmServiceMapConfigFromAPIImport(cfg)
		return nil
	}

	if pm.ApmServiceMapConfig == nil && prior.ApmServiceMapConfig != nil {
		pm.ApmServiceMapConfig = apmServiceMapConfigFromAPIImport(cfg)
	}

	existing := pm.ApmServiceMapConfig
	if existing == nil {
		return nil
	}

	if !apmServiceMapConfigHasAnyField(cfg) {
		pm.ApmServiceMapConfig = nil
		return nil
	}

	existing.Title = panelkit.PreserveString(existing.Title, cfg.Title)
	existing.Description = panelkit.PreserveString(existing.Description, cfg.Description)
	existing.HideTitle = panelkit.PreserveBool(existing.HideTitle, cfg.HideTitle)
	existing.HideBorder = panelkit.PreserveBool(existing.HideBorder, cfg.HideBorder)
	existing.Environment = panelkit.PreserveString(existing.Environment, cfg.Environment)
	existing.ServiceName = panelkit.PreserveString(existing.ServiceName, cfg.ServiceName)
	existing.ServiceGroupID = panelkit.PreserveString(existing.ServiceGroupID, cfg.ServiceGroupId)
	existing.Kuery = panelkit.PreserveString(existing.Kuery, cfg.Kuery)
	existing.MapOrientation = preserveMapOrientation(existing.MapOrientation, cfg.MapOrientation)
	existing.SyncWithDashboardFilters = panelkit.PreserveBool(existing.SyncWithDashboardFilters, cfg.SyncWithDashboardFilters)
	existing.AlertStatusFilter = preserveStringSetFromEnum(existing.AlertStatusFilter, cfg.AlertStatusFilter)
	existing.AnomalySeverityFilter = preserveStringSetFromEnum(existing.AnomalySeverityFilter, cfg.AnomalySeverityFilter)
	existing.ConnectionFilter = preserveStringSetFromEnum(existing.ConnectionFilter, cfg.ConnectionFilter)
	existing.SloStatusFilter = preserveStringSetFromEnum(existing.SloStatusFilter, cfg.SloStatusFilter)

	var priorTR *models.TimeRangeModel
	if prior.ApmServiceMapConfig != nil {
		priorTR = prior.ApmServiceMapConfig.TimeRange
	}
	existing.TimeRange = panelkit.MergeTimeRange(existing.TimeRange, cfg.TimeRange, priorTR)
	if prior.ApmServiceMapConfig != nil {
		apmServiceMapPreserveNullIntentFromPrior(prior.ApmServiceMapConfig, existing)
	}

	return nil
}

func apmServiceMapConfigFromAPIImport(cfg kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable) *models.ApmServiceMapConfigModel {
	if !apmServiceMapConfigHasAnyField(cfg) {
		return nil
	}
	return &models.ApmServiceMapConfigModel{
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
}

func apmServiceMapConfigHasAnyField(cfg kbapi.KibanaHTTPAPIsApmServiceMapEmbeddable) bool {
	if cfg.Title != nil || cfg.Description != nil || cfg.HideTitle != nil || cfg.HideBorder != nil ||
		cfg.Environment != nil || cfg.ServiceName != nil || cfg.ServiceGroupId != nil || cfg.Kuery != nil ||
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

func mapOrientationFromAPI(v *kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableMapOrientation) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*v))
}

func preserveMapOrientation(existing types.String, api *kbapi.KibanaHTTPAPIsApmServiceMapEmbeddableMapOrientation) types.String {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	if api == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*api))
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

func preserveStringSetFromEnum[T ~string](existing types.Set, api *[]T) types.Set {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	if api == nil || len(*api) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	return enumSliceToStringSet(api)
}

func apmServiceMapPreserveNullIntentFromPrior(prior, existing *models.ApmServiceMapConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	if !typeutils.IsKnown(prior.Title) {
		existing.Title = types.StringNull()
	}
	if !typeutils.IsKnown(prior.Description) {
		existing.Description = types.StringNull()
	}
	if !typeutils.IsKnown(prior.HideTitle) {
		existing.HideTitle = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.HideBorder) {
		existing.HideBorder = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.Environment) {
		existing.Environment = types.StringNull()
	}
	if !typeutils.IsKnown(prior.ServiceName) {
		existing.ServiceName = types.StringNull()
	}
	if !typeutils.IsKnown(prior.ServiceGroupID) {
		existing.ServiceGroupID = types.StringNull()
	}
	if !typeutils.IsKnown(prior.Kuery) {
		existing.Kuery = types.StringNull()
	}
	if !typeutils.IsKnown(prior.MapOrientation) {
		existing.MapOrientation = types.StringNull()
	}
	if !typeutils.IsKnown(prior.SyncWithDashboardFilters) {
		existing.SyncWithDashboardFilters = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.AlertStatusFilter) {
		existing.AlertStatusFilter = types.SetNull(types.StringType)
	}
	if !typeutils.IsKnown(prior.AnomalySeverityFilter) {
		existing.AnomalySeverityFilter = types.SetNull(types.StringType)
	}
	if !typeutils.IsKnown(prior.ConnectionFilter) {
		existing.ConnectionFilter = types.SetNull(types.StringType)
	}
	if !typeutils.IsKnown(prior.SloStatusFilter) {
		existing.SloStatusFilter = types.SetNull(types.StringType)
	}
	if prior.TimeRange == nil {
		existing.TimeRange = nil
	} else if existing.TimeRange != nil && prior.TimeRange != nil {
		if !typeutils.IsKnown(prior.TimeRange.From) {
			existing.TimeRange.From = types.StringNull()
		}
		if !typeutils.IsKnown(prior.TimeRange.To) {
			existing.TimeRange.To = types.StringNull()
		}
		if !typeutils.IsKnown(prior.TimeRange.Mode) {
			existing.TimeRange.Mode = types.StringNull()
		}
	}
}
