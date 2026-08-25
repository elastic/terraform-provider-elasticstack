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
//
// pm always arrives with AiopsChangePointChartConfig unset (callers build state from a zero-valued
// PanelModel to avoid aliasing plan pointers), so that field cannot be used to detect whether this
// panel was previously this same type. prior.AiopsChangePointChartConfig is the only reliable signal:
// non-nil means the panel was already this type and its null intent must be honored; nil means
// there is no prior null intent for this config block (creation, import, or a type change).
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, api kbapi.KibanaHTTPAPIsAiopsChangePointChart) diag.Diagnostics {
	if prior == nil || prior.AiopsChangePointChartConfig == nil {
		pm.AiopsChangePointChartConfig = aiopsChangePointChartConfigFromAPIImport(api)
		return nil
	}

	// Same-type update: rebuild from the API, then reapply the prior config's null intent for any
	// optional field the plan/state had not set (REQ-009 null-preservation).
	existing := aiopsChangePointChartConfigFromAPIImport(api)
	existing.TimeRange = panelkit.MergeTimeRange(existing.TimeRange, api.TimeRange, prior.AiopsChangePointChartConfig.TimeRange)
	aiopsChangePointChartPreserveNullIntentFromPrior(prior.AiopsChangePointChartConfig, existing)
	pm.AiopsChangePointChartConfig = existing
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
	panelkit.NullPreserveFromPrior(prior.AggregationFunction, &existing.AggregationFunction)
	panelkit.NullPreserveFromPrior(prior.SplitField, &existing.SplitField)
	panelkit.NullPreserveFromPrior(prior.Partitions, &existing.Partitions)
	panelkit.NullPreserveFromPrior(prior.MaxSeriesToPlot, &existing.MaxSeriesToPlot)
	panelkit.NullPreserveFromPrior(prior.ViewType, &existing.ViewType)
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
