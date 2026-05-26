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

func isPieNoESQLCandidateActuallyESQL(apiChart kbapi.KibanaHTTPAPIsPieNoESQL) bool {
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
	legend *kbapi.KibanaHTTPAPIsPieLegend,
	filters *kbapi.KibanaHTTPAPIsLensPanelFilters,
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

func pieChartConfigFromAPINoESQL(
	ctx context.Context,
	m *models.PieChartConfigModel,
	resolver lenscommon.Resolver,
	prior *models.PieChartConfigModel,
	apiChart kbapi.KibanaHTTPAPIsPieNoESQL,
) diag.Diagnostics {
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

func pieChartConfigFromAPIESQL(
	ctx context.Context,
	m *models.PieChartConfigModel,
	resolver lenscommon.Resolver,
	prior *models.PieChartConfigModel,
	apiChart kbapi.KibanaHTTPAPIsPieESQL,
) diag.Diagnostics {
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

func pieChartConfigToAPI(m *models.PieChartConfigModel, resolver lenscommon.Resolver) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs lenscommon.VisByValueConfig0
	if m == nil {
		return attrs, diags
	}

	isNoESQL := m.Query != nil

	if isNoESQL {
		var chart kbapi.KibanaHTTPAPIsPieNoESQL

		defaultMode := kbapi.KibanaHTTPAPIsValueDisplayModePercentage
		chart.Styling = &kbapi.KibanaHTTPAPIsPieStyling{
			Values: &kbapi.KibanaHTTPAPIsValueDisplay{Mode: &defaultMode},
		}

		chart.Title = m.Title.ValueStringPointer()
		chart.Description = m.Description.ValueStringPointer()
		chart.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()

		if !m.Sampling.IsNull() {
			val := float32(m.Sampling.ValueFloat64())
			chart.Sampling = &val
		}

		if !m.DonutHole.IsNull() {
			val := kbapi.KibanaHTTPAPIsPieStylingDonutHole(m.DonutHole.ValueString())
			if chart.Styling == nil {
				chart.Styling = &kbapi.KibanaHTTPAPIsPieStyling{}
			}
			chart.Styling.DonutHole = &val
		}

		if !m.LabelPosition.IsNull() {
			pos := kbapi.KibanaHTTPAPIsPieStylingLabelsPosition(m.LabelPosition.ValueString())
			if chart.Styling == nil {
				chart.Styling = &kbapi.KibanaHTTPAPIsPieStyling{}
			}
			chart.Styling.Labels = &struct {
				Position *kbapi.KibanaHTTPAPIsPieStylingLabelsPosition `json:"position,omitempty"`
				Visible  *bool                                         `json:"visible,omitempty"`
			}{Position: &pos}
		}

		if m.Legend != nil {
			chart.Legend = lenscommon.PartitionLegendToPieLegend(m.Legend)
		}
		if chart.Legend != nil && (chart.Legend.Size == nil || *chart.Legend.Size == "") {
			size := kbapi.KibanaHTTPAPIsLegendSizeAuto
			chart.Legend.Size = &size
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
			metrics := make([]kbapi.KibanaHTTPAPIsPieNoESQL_Metrics_Item, len(m.Metrics))
			for i, metric := range m.Metrics {
				if err := json.Unmarshal([]byte(metric.Config.ValueString()), &metrics[i]); err != nil {
					diags.AddError("Failed to unmarshal metric", err.Error())
				}
			}
			chart.Metrics = metrics
		}

		if len(m.GroupBy) > 0 {
			groupBy := make([]kbapi.KibanaHTTPAPIsPieNoESQL_GroupBy_Item, len(m.GroupBy))
			for i, grp := range m.GroupBy {
				if err := json.Unmarshal([]byte(grp.Config.ValueString()), &groupBy[i]); err != nil {
					diags.AddError("Failed to unmarshal group_by", err.Error())
				}
			}
			chart.GroupBy = &groupBy
		}

		chart.Type = kbapi.KibanaHTTPAPIsPieNoESQLTypePie

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
			items, ddDiags := lenscommon.DecodeLensDrilldownSlice[kbapi.KibanaHTTPAPIsPieNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
			diags.Append(ddDiags...)
			if !ddDiags.HasError() {
				chart.Drilldowns = &items
			}
		}

		if err := attrs.FromKibanaHTTPAPIsPieNoESQL(chart); err != nil {
			diags.AddError("Failed to create PieNoESQL schema", err.Error())
		}
	} else {
		var chart kbapi.KibanaHTTPAPIsPieESQL

		defaultMode := kbapi.KibanaHTTPAPIsValueDisplayModePercentage
		chart.Styling = &kbapi.KibanaHTTPAPIsPieStyling{
			Values: &kbapi.KibanaHTTPAPIsValueDisplay{Mode: &defaultMode},
		}

		chart.Title = m.Title.ValueStringPointer()
		chart.Description = m.Description.ValueStringPointer()
		chart.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()

		if !m.Sampling.IsNull() {
			val := float32(m.Sampling.ValueFloat64())
			chart.Sampling = &val
		}

		if !m.DonutHole.IsNull() {
			val := kbapi.KibanaHTTPAPIsPieStylingDonutHole(m.DonutHole.ValueString())
			if chart.Styling == nil {
				chart.Styling = &kbapi.KibanaHTTPAPIsPieStyling{}
			}
			chart.Styling.DonutHole = &val
		}

		if !m.LabelPosition.IsNull() {
			pos := kbapi.KibanaHTTPAPIsPieStylingLabelsPosition(m.LabelPosition.ValueString())
			if chart.Styling == nil {
				chart.Styling = &kbapi.KibanaHTTPAPIsPieStyling{}
			}
			chart.Styling.Labels = &struct {
				Position *kbapi.KibanaHTTPAPIsPieStylingLabelsPosition `json:"position,omitempty"`
				Visible  *bool                                         `json:"visible,omitempty"`
			}{Position: &pos}
		}

		if m.Legend != nil {
			chart.Legend = lenscommon.PartitionLegendToPieLegend(m.Legend)
		}
		if chart.Legend != nil && (chart.Legend.Size == nil || *chart.Legend.Size == "") {
			size := kbapi.KibanaHTTPAPIsLegendSizeAuto
			chart.Legend.Size = &size
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
				Color  *kbapi.KibanaHTTPAPIsPieESQL_Metrics_Color `json:"color,omitempty"`
				Column string                                     `json:"column"`
				Format *kbapi.KibanaHTTPAPIsFormatType            `json:"format,omitempty"`
				Label  *string                                    `json:"label,omitempty"`
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
				CollapseBy *kbapi.KibanaHTTPAPIsCollapseBy   `json:"collapse_by,omitempty"`
				Color      *kbapi.KibanaHTTPAPIsColorMapping `json:"color,omitempty"`
				Column     string                            `json:"column"`
				Format     *kbapi.KibanaHTTPAPIsFormatType   `json:"format,omitempty"`
				Label      *string                           `json:"label,omitempty"`
			}, len(m.GroupBy))
			for i, grp := range m.GroupBy {
				if err := json.Unmarshal([]byte(grp.Config.ValueString()), &groupBy[i]); err != nil {
					diags.AddError("Failed to unmarshal group_by", err.Error())
				}
				if groupBy[i].Format != nil {
					fb, _ := json.Marshal(groupBy[i].Format)
					if string(fb) == jsonNullString || len(fb) == 0 {
						var format kbapi.KibanaHTTPAPIsFormatType
						_ = format.FromKibanaHTTPAPIsNumericFormat(kbapi.KibanaHTTPAPIsNumericFormat{Type: kbapi.Number})
						groupBy[i].Format = &format
					}
				}
			}
			chart.GroupBy = &groupBy
		}

		chart.Type = kbapi.KibanaHTTPAPIsPieESQLTypePie

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
			items, ddDiags := lenscommon.DecodeLensDrilldownSlice[kbapi.KibanaHTTPAPIsPieESQL_Drilldowns_Item](writes.DrilldownsRaw)
			diags.Append(ddDiags...)
			if !ddDiags.HasError() {
				chart.Drilldowns = &items
			}
		}

		if err := attrs.FromKibanaHTTPAPIsPieESQL(chart); err != nil {
			diags.AddError("Failed to create PieESQL schema", err.Error())
		}
	}

	return attrs, diags
}
