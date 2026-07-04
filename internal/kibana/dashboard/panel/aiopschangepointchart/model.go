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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
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
	if typeutils.IsKnown(cfg.Partitions) {
		elems := cfg.Partitions.Elements()
		items := make([]string, 0, len(elems))
		for _, e := range elems {
			items = append(items, e.(types.String).ValueString())
		}
		panel.Config.Partitions = &items
	}
	if typeutils.IsKnown(cfg.MaxSeriesToPlot) {
		v := cfg.MaxSeriesToPlot.ValueFloat32()
		panel.Config.MaxSeriesToPlot = &v
	}
	if typeutils.IsKnown(cfg.ViewType) {
		v := kbapi.KibanaHTTPAPIsAiopsChangePointChartViewType(cfg.ViewType.ValueString())
		panel.Config.ViewType = &v
	}

	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&panel.Config.Title, &panel.Config.Description, &panel.Config.HideTitle, &panel.Config.HideBorder)
	panel.Config.TimeRange = lenscommon.TimeRangeModelToAPI(cfg.TimeRange)

	return nil
}

// PopulateFromAPI maps the Kibana API panel config into Terraform panel state while preserving
// prior null intent (REQ-009). prior is the prior TF state/plan panel, or nil on import.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, api kbapi.KibanaHTTPAPIsAiopsChangePointChart) diag.Diagnostics {
	// On import (prior == nil): populate required fields unconditionally; optional fields only when API non-nil.
	if prior == nil {
		pm.AiopsChangePointChartConfig = aiopsChangePointChartConfigFromAPIImport(api)
		return nil
	}

	// Type-change recovery: the plan dropped this config block but prior still has it.
	// Rebuild entirely from the API and skip null-preservation, since there is no
	// current-plan null intent to honor.
	if pm.AiopsChangePointChartConfig == nil && prior.AiopsChangePointChartConfig != nil {
		pm.AiopsChangePointChartConfig = aiopsChangePointChartConfigFromAPIImport(api)
		return nil
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
	existing.MaxSeriesToPlot = panelkit.PreserveFloat32(existing.MaxSeriesToPlot, api.MaxSeriesToPlot)

	// Partitions set: null-preserve. When the practitioner omitted it (null/unknown in state), keep null
	// regardless of API-returned values; otherwise refresh from the API set (order-insensitive).
	if typeutils.IsKnown(existing.Partitions) {
		existing.Partitions = changePointPartitionsFromAPI(api.Partitions)
	}

	panelkit.ApplyPresentationFromAPI(&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder,
		api.Title, api.Description, api.HideTitle, api.HideBorder)

	var priorTR *models.TimeRangeModel
	if prior.AiopsChangePointChartConfig != nil {
		priorTR = prior.AiopsChangePointChartConfig.TimeRange
	}
	existing.TimeRange = panelkit.MergeTimeRange(existing.TimeRange, api.TimeRange, priorTR)

	if prior.AiopsChangePointChartConfig != nil {
		aiopsChangePointChartPreserveNullIntentFromPrior(prior.AiopsChangePointChartConfig, existing)
	}
	return nil
}

func aiopsChangePointChartConfigFromAPIImport(api kbapi.KibanaHTTPAPIsAiopsChangePointChart) *models.AiopsChangePointChartConfigModel {
	cfg := &models.AiopsChangePointChartConfigModel{
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
	cfg.MaxSeriesToPlot = types.Float32PointerValue(api.MaxSeriesToPlot)
	cfg.TimeRange = panelkit.TimeRangeFromAPI(api.TimeRange, nil)
	return cfg
}

func aiopsChangePointChartPreserveNullIntentFromPrior(prior, existing *models.AiopsChangePointChartConfigModel) {
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
		existing.MaxSeriesToPlot = types.Float32Null()
	}
	if !typeutils.IsKnown(prior.ViewType) {
		existing.ViewType = types.StringNull()
	}
	panelkit.NullPreservePresentationFromPrior(prior.Title, prior.Description, prior.HideTitle, prior.HideBorder,
		&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder)
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
