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

package lenspie

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const jsonNullString = "null"

func isPieNoESQLCandidateActuallyESQL(apiChart kbapi.PieNoESQL) bool {
	body, err := json.Marshal(apiChart.DataSource)
	if err != nil {
		return false
	}

	var dataset struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &dataset); err != nil {
		return false
	}

	return dataset.Type == lenscommon.LensDatasetTypeESQL || dataset.Type == lenscommon.LensDatasetTypeTable
}

func pieChartConfigPopulateCommonFields(
	m *models.PieChartConfigModel,
	title, description *string,
	ignoreGlobalFilters *bool,
	sampling *float32,
	donutHole, labelPosition *string,
	datasetBytes []byte,
	datasetErr error,
	legend kbapi.PieLegend,
	filters []kbapi.LensPanelFilters_Item,
	diags *diag.Diagnostics,
) bool {
	m.Title = types.StringPointerValue(title)
	m.Description = types.StringPointerValue(description)
	if ignoreGlobalFilters != nil {
		m.IgnoreGlobalFilters = types.BoolValue(*ignoreGlobalFilters)
	} else {
		m.IgnoreGlobalFilters = types.BoolValue(false)
	}
	if sampling != nil {
		m.Sampling = types.Float64Value(float64(*sampling))
	} else {
		m.Sampling = types.Float64Value(1.0)
	}
	if donutHole != nil {
		m.DonutHole = types.StringValue(*donutHole)
	} else {
		m.DonutHole = types.StringNull()
	}
	if labelPosition != nil {
		m.LabelPosition = types.StringValue(*labelPosition)
	} else {
		m.LabelPosition = types.StringNull()
	}
	dv, ok := lenscommon.MarshalToNormalized(datasetBytes, datasetErr, "data_source_json", diags)
	if !ok {
		return false
	}
	m.DataSourceJSON = dv
	m.Legend = &models.PartitionLegendModel{}
	lenscommon.PartitionLegendFromPieLegend(m.Legend, legend)
	m.Filters = lenscommon.PopulateFiltersFromAPI(filters, diags)
	return !diags.HasError()
}

func pieChartConfigFromAPINoESQL(ctx context.Context, m *models.PieChartConfigModel, resolver lenscommon.Resolver, prior *models.PieChartConfigModel, apiChart kbapi.PieNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	var donutHole *string
	if apiChart.Styling.DonutHole != nil {
		s := string(*apiChart.Styling.DonutHole)
		donutHole = &s
	}
	var labelPosition *string
	if apiChart.Styling.Labels != nil && apiChart.Styling.Labels.Position != nil {
		s := string(*apiChart.Styling.Labels.Position)
		labelPosition = &s
	}
	datasetBytes, datasetErr := json.Marshal(apiChart.DataSource)

	if !pieChartConfigPopulateCommonFields(m,
		apiChart.Title, apiChart.Description, apiChart.IgnoreGlobalFilters, apiChart.Sampling,
		donutHole, labelPosition,
		datasetBytes, datasetErr, apiChart.Legend,
		apiChart.Filters, &diags,
	) {
		return diags
	}

	m.Query = &models.FilterSimpleModel{}
	lenscommon.FilterSimpleFromAPI(m.Query, apiChart.Query)

	if len(apiChart.Metrics) > 0 {
		m.Metrics = make([]models.PieMetricModel, len(apiChart.Metrics))
		for i, metric := range apiChart.Metrics {
			metricJSON, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				continue
			}
			m.Metrics[i].Config = customtypes.NewJSONWithDefaultsValue(
				string(metricJSON),
				lenscommon.PopulatePieChartMetricDefaults,
			)
		}
	}

	if apiChart.GroupBy != nil && len(*apiChart.GroupBy) > 0 {
		m.GroupBy = make([]models.PieGroupByModel, len(*apiChart.GroupBy))
		for i, groupBy := range *apiChart.GroupBy {
			groupByJSON, err := json.Marshal(groupBy)
			if err != nil {
				diags.AddError("Failed to marshal group_by", err.Error())
				continue
			}
			m.GroupBy[i].Config = customtypes.NewJSONWithDefaultsValue(
				string(groupByJSON),
				lenscommon.PopulateLensGroupByDefaults,
			)
		}
	}

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lenscommon.LensDrilldownsAPIToWire(apiChart.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lenscommon.LensChartPresentationReadsFor(ctx, resolver, priorLens, apiChart.TimeRange, apiChart.HideTitle, apiChart.HideBorder, apiChart.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres

	return diags
}

func pieChartConfigFromAPIESQL(ctx context.Context, m *models.PieChartConfigModel, resolver lenscommon.Resolver, prior *models.PieChartConfigModel, apiChart kbapi.PieESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	var donutHole *string
	if apiChart.Styling.DonutHole != nil {
		s := string(*apiChart.Styling.DonutHole)
		donutHole = &s
	}
	var labelPosition *string
	if apiChart.Styling.Labels != nil && apiChart.Styling.Labels.Position != nil {
		s := string(*apiChart.Styling.Labels.Position)
		labelPosition = &s
	}
	datasetBytes, datasetErr := json.Marshal(apiChart.DataSource)

	if !pieChartConfigPopulateCommonFields(m,
		apiChart.Title, apiChart.Description, apiChart.IgnoreGlobalFilters, apiChart.Sampling,
		donutHole, labelPosition,
		datasetBytes, datasetErr, apiChart.Legend,
		apiChart.Filters, &diags,
	) {
		return diags
	}

	m.Query = nil

	if len(apiChart.Metrics) > 0 {
		m.Metrics = make([]models.PieMetricModel, len(apiChart.Metrics))
		for i, metric := range apiChart.Metrics {
			metricJSON, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				continue
			}
			m.Metrics[i].Config = customtypes.NewJSONWithDefaultsValue(
				string(metricJSON),
				lenscommon.PopulatePieChartMetricDefaults,
			)
		}
	}

	if apiChart.GroupBy != nil && len(*apiChart.GroupBy) > 0 {
		m.GroupBy = make([]models.PieGroupByModel, len(*apiChart.GroupBy))
		for i, groupBy := range *apiChart.GroupBy {
			groupByJSON, err := json.Marshal(groupBy)
			if err != nil {
				diags.AddError("Failed to marshal group_by", err.Error())
				continue
			}
			m.GroupBy[i].Config = customtypes.NewJSONWithDefaultsValue(
				string(groupByJSON),
				lenscommon.PopulateLensGroupByDefaults,
			)
		}
	}

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lenscommon.LensDrilldownsAPIToWire(apiChart.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lenscommon.LensChartPresentationReadsFor(ctx, resolver, priorLens, apiChart.TimeRange, apiChart.HideTitle, apiChart.HideBorder, apiChart.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres

	return diags
}

func pieChartConfigToAPI(m *models.PieChartConfigModel, resolver lenscommon.Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	if m == nil {
		return attrs, diags
	}

	isNoESQL := m.Query != nil

	if isNoESQL {
		var chart kbapi.PieNoESQL

		defaultMode := kbapi.ValueDisplayModePercentage
		chart.Styling.Values = kbapi.ValueDisplay{Mode: &defaultMode}

		chart.Title = m.Title.ValueStringPointer()
		chart.Description = m.Description.ValueStringPointer()
		chart.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()

		if !m.Sampling.IsNull() {
			val := float32(m.Sampling.ValueFloat64())
			chart.Sampling = &val
		}

		if !m.DonutHole.IsNull() {
			val := kbapi.PieStylingDonutHole(m.DonutHole.ValueString())
			chart.Styling.DonutHole = &val
		}

		if !m.LabelPosition.IsNull() {
			pos := kbapi.PieStylingLabelsPosition(m.LabelPosition.ValueString())
			chart.Styling.Labels = &struct {
				Position *kbapi.PieStylingLabelsPosition `json:"position,omitempty"`
				Visible  *bool                           `json:"visible,omitempty"`
			}{Position: &pos}
		}

		if m.Legend != nil {
			chart.Legend = lenscommon.PartitionLegendToPieLegend(m.Legend)
		}
		if chart.Legend.Size == "" {
			chart.Legend.Size = kbapi.LegendSizeAuto
		}

		if m.DataSourceJSON.IsNull() {
			diags.AddError("Missing dataset", "pie_chart_config.data_source_json must be provided")
			return attrs, diags
		}
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &chart.DataSource); err != nil {
			diags.AddError("Failed to unmarshal pie_chart_config.data_source_json", err.Error())
			return attrs, diags
		}

		chart.Query = lenscommon.FilterSimpleToAPI(m.Query)

		chart.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

		if len(m.Metrics) > 0 {
			metrics := make([]kbapi.PieNoESQL_Metrics_Item, len(m.Metrics))
			for i, metric := range m.Metrics {
				if err := json.Unmarshal([]byte(metric.Config.ValueString()), &metrics[i]); err != nil {
					diags.AddError("Failed to unmarshal metric", err.Error())
				}
			}
			chart.Metrics = metrics
		}

		if len(m.GroupBy) > 0 {
			groupBy := make([]kbapi.PieNoESQL_GroupBy_Item, len(m.GroupBy))
			for i, grp := range m.GroupBy {
				if err := json.Unmarshal([]byte(grp.Config.ValueString()), &groupBy[i]); err != nil {
					diags.AddError("Failed to unmarshal group_by", err.Error())
				}
			}
			chart.GroupBy = &groupBy
		}

		chart.Type = kbapi.PieNoESQLTypePie

		writes, presDiags := lenscommon.LensChartPresentationWritesFor(resolver, m.LensChartPresentationTFModel)
		diags.Append(presDiags...)
		if presDiags.HasError() {
			return attrs, diags
		}

		chart.TimeRange = writes.TimeRange
		if writes.HideTitle != nil {
			chart.HideTitle = writes.HideTitle
		}
		if writes.HideBorder != nil {
			chart.HideBorder = writes.HideBorder
		}
		if writes.References != nil {
			chart.References = writes.References
		}
		if len(writes.DrilldownsRaw) > 0 {
			items, ddDiags := lenscommon.DecodeLensDrilldownSlice[kbapi.PieNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
			diags.Append(ddDiags...)
			if !ddDiags.HasError() {
				chart.Drilldowns = &items
			}
		}

		if err := attrs.FromPieNoESQL(chart); err != nil {
			diags.AddError("Failed to create PieNoESQL schema", err.Error())
		}
	} else {
		var chart kbapi.PieESQL

		defaultMode := kbapi.ValueDisplayModePercentage
		chart.Styling.Values = kbapi.ValueDisplay{Mode: &defaultMode}

		chart.Title = m.Title.ValueStringPointer()
		chart.Description = m.Description.ValueStringPointer()
		chart.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()

		if !m.Sampling.IsNull() {
			val := float32(m.Sampling.ValueFloat64())
			chart.Sampling = &val
		}

		if !m.DonutHole.IsNull() {
			val := kbapi.PieStylingDonutHole(m.DonutHole.ValueString())
			chart.Styling.DonutHole = &val
		}

		if !m.LabelPosition.IsNull() {
			pos := kbapi.PieStylingLabelsPosition(m.LabelPosition.ValueString())
			chart.Styling.Labels = &struct {
				Position *kbapi.PieStylingLabelsPosition `json:"position,omitempty"`
				Visible  *bool                           `json:"visible,omitempty"`
			}{Position: &pos}
		}

		if m.Legend != nil {
			chart.Legend = lenscommon.PartitionLegendToPieLegend(m.Legend)
		}
		if chart.Legend.Size == "" {
			chart.Legend.Size = kbapi.LegendSizeAuto
		}

		if m.DataSourceJSON.IsNull() {
			diags.AddError("Missing dataset", "pie_chart_config.data_source_json must be provided")
			return attrs, diags
		}
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &chart.DataSource); err != nil {
			diags.AddError("Failed to unmarshal pie_chart_config.data_source_json", err.Error())
			return attrs, diags
		}

		chart.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

		if len(m.Metrics) > 0 {
			metrics := make([]struct {
				Color  *kbapi.PieESQL_Metrics_Color `json:"color,omitempty"`
				Column string                       `json:"column"`
				Format kbapi.FormatType             `json:"format"`
				Label  *string                      `json:"label,omitempty"`
			}, len(m.Metrics))
			for i, metric := range m.Metrics {
				if err := json.Unmarshal([]byte(metric.Config.ValueString()), &metrics[i]); err != nil {
					diags.AddError("Failed to unmarshal metric", err.Error())
				}
			}
			chart.Metrics = metrics
		}

		if len(m.GroupBy) > 0 {
			groupBy := make([]struct {
				CollapseBy kbapi.CollapseBy   `json:"collapse_by"`
				Color      kbapi.ColorMapping `json:"color"`
				Column     string             `json:"column"`
				Format     kbapi.FormatType   `json:"format"`
				Label      *string            `json:"label,omitempty"`
			}, len(m.GroupBy))
			for i, grp := range m.GroupBy {
				if err := json.Unmarshal([]byte(grp.Config.ValueString()), &groupBy[i]); err != nil {
					diags.AddError("Failed to unmarshal group_by", err.Error())
				}
				fb, _ := json.Marshal(groupBy[i].Format)
				if string(fb) == jsonNullString || len(fb) == 0 {
					_ = groupBy[i].Format.FromNumericFormat(kbapi.NumericFormat{Type: kbapi.Number})
				}
			}
			chart.GroupBy = &groupBy
		}

		chart.Type = kbapi.PieESQLTypePie

		writes, presDiags := lenscommon.LensChartPresentationWritesFor(resolver, m.LensChartPresentationTFModel)
		diags.Append(presDiags...)
		if presDiags.HasError() {
			return attrs, diags
		}

		chart.TimeRange = writes.TimeRange
		if writes.HideTitle != nil {
			chart.HideTitle = writes.HideTitle
		}
		if writes.HideBorder != nil {
			chart.HideBorder = writes.HideBorder
		}
		if writes.References != nil {
			chart.References = writes.References
		}
		if len(writes.DrilldownsRaw) > 0 {
			items, ddDiags := lenscommon.DecodeLensDrilldownSlice[kbapi.PieESQL_Drilldowns_Item](writes.DrilldownsRaw)
			diags.Append(ddDiags...)
			if !ddDiags.HasError() {
				chart.Drilldowns = &items
			}
		}

		if err := attrs.FromPieESQL(chart); err != nil {
			diags.AddError("Failed to create PieESQL schema", err.Error())
		}
	}

	return attrs, diags
}
