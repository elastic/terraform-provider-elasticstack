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

package aiopschangepointchart

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes Terraform state from pm into the typed API panel config.
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart) diag.Diagnostics {
	cfg := pm.AiopsChangePointChartConfig
	if cfg == nil {
		return nil
	}

	panel.Config.DataViewId = cfg.DataViewID.ValueString()
	panel.Config.MetricField = cfg.MetricField.ValueString()

	if typeutils.IsKnown(cfg.AggregationFunction) {
		v := kbapi.KibanaHTTPAPIsAiopsChangePointChartAggregationFunction(cfg.AggregationFunction.ValueString())
		panel.Config.AggregationFunction = &v
	}
	if typeutils.IsKnown(cfg.SplitField) {
		panel.Config.SplitField = cfg.SplitField.ValueStringPointer()
	}
	if !cfg.Partitions.IsNull() && !cfg.Partitions.IsUnknown() {
		elems := cfg.Partitions.Elements()
		items := make([]string, 0, len(elems))
		for _, e := range elems {
			items = append(items, e.(types.String).ValueString())
		}
		if len(items) > 0 {
			panel.Config.Partitions = &items
		}
	}
	if typeutils.IsKnown(cfg.MaxSeriesToPlot) {
		v := float32(cfg.MaxSeriesToPlot.ValueFloat64())
		panel.Config.MaxSeriesToPlot = &v
	}
	if typeutils.IsKnown(cfg.ViewType) {
		v := kbapi.KibanaHTTPAPIsAiopsChangePointChartViewType(cfg.ViewType.ValueString())
		panel.Config.ViewType = &v
	}

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
	panel.Config.TimeRange = panelkit.TimeRangeToAPI(cfg.TimeRange)

	return nil
}

// PopulateFromAPI maps the Kibana API panel config into Terraform panel state while preserving
// prior null intent (REQ-009). prior is the prior TF state/plan panel, or nil on import.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, api kbapi.KibanaHTTPAPIsAiopsChangePointChart) diag.Diagnostics {
	// On import (prior == nil): populate required fields unconditionally; optional fields only when API non-nil.
	if prior == nil {
		pm.AiopsChangePointChartConfig = &models.AiopsChangePointChartConfigModel{
			DataViewID:          types.StringValue(api.DataViewId),
			MetricField:         types.StringValue(api.MetricField),
			AggregationFunction: changePointAggregationFunctionValue(api.AggregationFunction),
			SplitField:          types.StringPointerValue(api.SplitField),
			Partitions:          changePointPartitionsFromAPI(api.Partitions),
			ViewType:            changePointViewTypeValue(api.ViewType),
			Title:               types.StringPointerValue(api.Title),
			Description:         types.StringPointerValue(api.Description),
			HideTitle:           types.BoolPointerValue(api.HideTitle),
			HideBorder:          types.BoolPointerValue(api.HideBorder),
		}
		if api.MaxSeriesToPlot != nil {
			pm.AiopsChangePointChartConfig.MaxSeriesToPlot = types.Float64Value(float64(*api.MaxSeriesToPlot))
		} else {
			pm.AiopsChangePointChartConfig.MaxSeriesToPlot = types.Float64Null()
		}
		if api.TimeRange != nil {
			pm.AiopsChangePointChartConfig.TimeRange = &models.TimeRangeModel{
				From: types.StringValue(api.TimeRange.From),
				To:   types.StringValue(api.TimeRange.To),
			}
			if api.TimeRange.Mode != nil {
				pm.AiopsChangePointChartConfig.TimeRange.Mode = types.StringValue(string(*api.TimeRange.Mode))
			}
		}
		return nil
	}

	if pm.AiopsChangePointChartConfig == nil && prior.AiopsChangePointChartConfig != nil {
		pm.AiopsChangePointChartConfig = &models.AiopsChangePointChartConfigModel{
			DataViewID:          types.StringValue(api.DataViewId),
			MetricField:         types.StringValue(api.MetricField),
			AggregationFunction: changePointAggregationFunctionValue(api.AggregationFunction),
			SplitField:          types.StringPointerValue(api.SplitField),
			Partitions:          changePointPartitionsFromAPI(api.Partitions),
			ViewType:            changePointViewTypeValue(api.ViewType),
			Title:               types.StringPointerValue(api.Title),
			Description:         types.StringPointerValue(api.Description),
			HideTitle:           types.BoolPointerValue(api.HideTitle),
			HideBorder:          types.BoolPointerValue(api.HideBorder),
		}
		if api.MaxSeriesToPlot != nil {
			pm.AiopsChangePointChartConfig.MaxSeriesToPlot = types.Float64Value(float64(*api.MaxSeriesToPlot))
		} else {
			pm.AiopsChangePointChartConfig.MaxSeriesToPlot = types.Float64Null()
		}
	}

	existing := pm.AiopsChangePointChartConfig
	if existing == nil {
		return nil
	}

	// Required fields always update from the API.
	existing.DataViewID = types.StringValue(api.DataViewId)
	existing.MetricField = types.StringValue(api.MetricField)

	// Optional enum/string fields: only update from API when already known in state (REQ-009 null-preservation).
	if typeutils.IsKnown(existing.AggregationFunction) {
		existing.AggregationFunction = changePointAggregationFunctionValue(api.AggregationFunction)
	}
	existing.SplitField = panelkit.PreserveString(existing.SplitField, api.SplitField)
	if typeutils.IsKnown(existing.ViewType) {
		existing.ViewType = changePointViewTypeValue(api.ViewType)
	}
	existing.MaxSeriesToPlot = panelkit.PreserveFloat64(existing.MaxSeriesToPlot, float32PtrToFloat64Ptr(api.MaxSeriesToPlot))

	// Partitions set: null-preserve. When the practitioner omitted it (null/unknown in state), keep null
	// regardless of API-returned values; otherwise refresh from the API set (order-insensitive).
	if typeutils.IsKnown(existing.Partitions) {
		existing.Partitions = changePointPartitionsFromAPI(api.Partitions)
	}

	existing.Title = panelkit.PreserveString(existing.Title, api.Title)
	existing.Description = panelkit.PreserveString(existing.Description, api.Description)
	existing.HideTitle = panelkit.PreserveBool(existing.HideTitle, api.HideTitle)
	existing.HideBorder = panelkit.PreserveBool(existing.HideBorder, api.HideBorder)

	var priorTR *models.TimeRangeModel
	if prior.AiopsChangePointChartConfig != nil {
		priorTR = prior.AiopsChangePointChartConfig.TimeRange
	}
	existing.TimeRange = panelkit.TimeRangeFromAPI(priorTR, api.TimeRange)

	if prior.AiopsChangePointChartConfig != nil {
		preserveNullIntentFromPrior(prior.AiopsChangePointChartConfig, existing)
	}
	return nil
}

func preserveNullIntentFromPrior(prior, existing *models.AiopsChangePointChartConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	if !typeutils.IsKnown(prior.AggregationFunction) {
		existing.AggregationFunction = types.StringNull()
	}
	if !typeutils.IsKnown(prior.SplitField) {
		existing.SplitField = types.StringNull()
	}
	if !typeutils.IsKnown(prior.Partitions) {
		existing.Partitions = types.SetNull(types.StringType)
	}
	if !typeutils.IsKnown(prior.MaxSeriesToPlot) {
		existing.MaxSeriesToPlot = types.Float64Null()
	}
	if !typeutils.IsKnown(prior.ViewType) {
		existing.ViewType = types.StringNull()
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
	if prior.TimeRange == nil {
		existing.TimeRange = nil
	}
}

func changePointAggregationFunctionValue(v *kbapi.KibanaHTTPAPIsAiopsChangePointChartAggregationFunction) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*v))
}

func changePointViewTypeValue(v *kbapi.KibanaHTTPAPIsAiopsChangePointChartViewType) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*v))
}

// changePointPartitionsFromAPI builds a types.Set from the API *[]string. Returns a null set when
// the API value is nil or empty so state reflects an omitted partitions field.
func changePointPartitionsFromAPI(v *[]string) types.Set {
	if v == nil || len(*v) == 0 {
		return types.SetNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(*v))
	for _, p := range *v {
		elems = append(elems, types.StringValue(p))
	}
	s, diags := types.SetValue(types.StringType, elems)
	if diags.HasError() {
		return types.SetNull(types.StringType)
	}
	return s
}

// float32PtrToFloat64Ptr converts a *float32 API field to a *float64 so it can be used with
// panelkit.PreserveFloat64 (which takes *float64). Returns nil when the input is nil.
func float32PtrToFloat64Ptr(v *float32) *float64 {
	if v == nil {
		return nil
	}
	f := float64(*v)
	return &f
}
