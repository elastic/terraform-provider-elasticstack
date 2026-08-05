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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
	m.Sampling = lenscommon.SamplingFromAPIWithDefault(sampling, 1.0)
	m.DonutHole = typeutils.StringishPointerValue(donutHole)
	m.LabelPosition = typeutils.StringishPointerValue(labelPosition)
	dv, ok := lenscommon.WrapNormalizedJSON(datasetBytes, datasetErr, "data_source_json", diags)
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
	_ context.Context,
	m *models.PieChartConfigModel,
	_ *models.PieChartConfigModel,
	apiChart kbapi.KibanaHTTPAPIsPieNoESQL,
) diag.Diagnostics {
	var diags diag.Diagnostics

	var donutHole *string
	var labelPosition *string
	if apiChart.Styling != nil {
		if apiChart.Styling.DonutHole != nil {
			donutHole = pieUnionString(apiChart.Styling.DonutHole)
		}
		if apiChart.Styling.Labels != nil && apiChart.Styling.Labels.Position != nil {
			labelPosition = pieUnionString(apiChart.Styling.Labels.Position)
		}
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

	return diags
}

func pieChartConfigFromAPIESQL(
	_ context.Context,
	m *models.PieChartConfigModel,
	_ *models.PieChartConfigModel,
	apiChart kbapi.KibanaHTTPAPIsPieESQL,
) diag.Diagnostics {
	var diags diag.Diagnostics

	var donutHole *string
	var labelPosition *string
	if apiChart.Styling != nil {
		if apiChart.Styling.DonutHole != nil {
			donutHole = pieUnionString(apiChart.Styling.DonutHole)
		}
		if apiChart.Styling.Labels != nil && apiChart.Styling.Labels.Position != nil {
			labelPosition = pieUnionString(apiChart.Styling.Labels.Position)
		}
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

	return diags
}

// populatePieStyling sets the default styling mode and applies optional donut hole, label
// position, and legend fields onto styling. It avoids callers duplicating this setup across
// NoESQL and ESQL branches.
func populatePieStyling(m *models.PieChartConfigModel, styling **kbapi.KibanaHTTPAPIsPieStyling, legend **kbapi.KibanaHTTPAPIsPieLegend) {
	*styling = &kbapi.KibanaHTTPAPIsPieStyling{
		Values: &kbapi.KibanaHTTPAPIsValueDisplay{Mode: pieValueDisplayMode("percentage")},
	}

	if !m.DonutHole.IsNull() {
		val := kbapi.KibanaHTTPAPIsPieStyling_DonutHole{}
		if err := pieSetUnionString(&val, m.DonutHole.ValueString()); err != nil {
			return
		}
		(*styling).DonutHole = &val
	}

	if !m.LabelPosition.IsNull() {
		pos := kbapi.KibanaHTTPAPIsPieStyling_Labels_Position{}
		if err := pieSetUnionString(&pos, m.LabelPosition.ValueString()); err != nil {
			return
		}
		(*styling).Labels = &struct {
			Position *kbapi.KibanaHTTPAPIsPieStyling_Labels_Position `json:"position,omitempty"`
			Visible  *bool                                           `json:"visible,omitempty"`
		}{Position: &pos}
	}

	if m.Legend != nil {
		*legend = lenscommon.PartitionLegendToPieLegend(m.Legend)
	}
	if *legend != nil {
		sizeValue := pieUnionString((*legend).Size)
		if sizeValue == nil || *sizeValue == "" {
			size := kbapi.KibanaHTTPAPIsLegendSize{}
			if err := pieSetUnionString(&size, "auto"); err == nil {
				(*legend).Size = &size
			}
		}
	}
}

func pieChartConfigToAPINoESQL(m *models.PieChartConfigModel) (kbapi.KibanaHTTPAPIsPieNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var chart kbapi.KibanaHTTPAPIsPieNoESQL

	chart.Title, chart.Description, chart.IgnoreGlobalFilters, chart.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	populatePieStyling(m, &chart.Styling, &chart.Legend)

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing dataset", "pie_chart_config.data_source_json must be provided")
		return chart, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &chart.DataSource); err != nil {
		diags.AddError("Failed to unmarshal pie_chart_config.data_source_json", err.Error())
		return chart, diags
	}

	chart.Query = lenscommon.FilterSimpleToAPI(m.Query)
	chart.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if len(m.Metrics) > 0 {
		metrics := make([]json.RawMessage, len(m.Metrics))
		for i, metric := range m.Metrics {
			metrics[i] = json.RawMessage(metric.Config.ValueString())
			if !json.Valid(metrics[i]) {
				diags.AddError("Invalid metric", "metric configuration must be valid JSON")
			}
		}
		raw, err := json.Marshal(struct {
			Metrics []json.RawMessage `json:"metrics"`
		}{Metrics: metrics})
		if err == nil {
			err = json.Unmarshal(raw, &chart)
		}
		if err != nil {
			diags.AddError("Failed to unmarshal metrics", err.Error())
		}
	}

	if len(m.GroupBy) > 0 {
		groupBy := make([]json.RawMessage, len(m.GroupBy))
		for i, grp := range m.GroupBy {
			groupBy[i] = json.RawMessage(grp.Config.ValueString())
			if !json.Valid(groupBy[i]) {
				diags.AddError("Invalid group_by", "group_by configuration must be valid JSON")
			}
		}
		raw, err := json.Marshal(struct {
			GroupBy []json.RawMessage `json:"group_by"`
		}{GroupBy: groupBy})
		if err == nil {
			err = json.Unmarshal(raw, &chart)
		}
		if err != nil {
			diags.AddError("Failed to unmarshal group_by", err.Error())
		}
	}

	chart.Type = kbapi.KibanaHTTPAPIsPieNoESQLTypePie

	return chart, diags
}

func pieChartConfigToAPIESQL(m *models.PieChartConfigModel) (kbapi.KibanaHTTPAPIsPieESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var chart kbapi.KibanaHTTPAPIsPieESQL

	chart.Title, chart.Description, chart.IgnoreGlobalFilters, chart.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	populatePieStyling(m, &chart.Styling, &chart.Legend)

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing dataset", "pie_chart_config.data_source_json must be provided")
		return chart, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &chart.DataSource); err != nil {
		diags.AddError("Failed to unmarshal pie_chart_config.data_source_json", err.Error())
		return chart, diags
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
		rawEntries := make([]lenscommon.EsqlGroupByAPIFields, len(m.GroupBy))
		for i, grp := range m.GroupBy {
			if err := json.Unmarshal([]byte(grp.Config.ValueString()), &rawEntries[i]); err != nil {
				diags.AddError("Failed to unmarshal group_by", err.Error())
			}
			if rawEntries[i].Format != nil {
				fb, _ := json.Marshal(rawEntries[i].Format)
				if string(fb) == lenscommon.JSONNullString || len(fb) == 0 {
					var format kbapi.KibanaHTTPAPIsFormatType
					_ = format.FromKibanaHTTPAPIsNumericFormat(kbapi.KibanaHTTPAPIsNumericFormat{Type: kbapi.Number})
					rawEntries[i].Format = &format
				}
			}
		}
		lenscommon.SetEsqlGroupByOnAPI(rawEntries, &chart.GroupBy, &diags)
	}

	chart.Type = kbapi.KibanaHTTPAPIsPieESQLTypePie

	return chart, diags
}

func pieChartConfigToAPI(m *models.PieChartConfigModel) (lenscommon.LensByValueConfig, diag.Diagnostics) {
	var result lenscommon.LensByValueConfig
	var diags diag.Diagnostics
	if m == nil {
		return result, diags
	}

	if lenscommon.ConfigUsesESQL(m.Query) {
		chart, chartDiags := pieChartConfigToAPIESQL(m)
		diags.Append(chartDiags...)
		if chartDiags.HasError() {
			return result, diags
		}
		var err error
		result.Chart, err = pieChartToLensAPI(chart)
		if err != nil {
			diags.AddError("Failed to create Lens chart config", err.Error())
			return result, diags
		}
	} else {
		chart, chartDiags := pieChartConfigToAPINoESQL(m)
		diags.Append(chartDiags...)
		if chartDiags.HasError() {
			return result, diags
		}
		var err error
		result.Chart, err = pieChartToLensAPI(chart)
		if err != nil {
			diags.AddError("Failed to create Lens chart config", err.Error())
			return result, diags
		}
	}
	if err := preservePieRawDimensions(&result.Chart, m); err != nil {
		diags.AddError("Failed to preserve pie chart dimensions", err.Error())
		return result, diags
	}

	writes, presentationDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presentationDiags...)
	if presentationDiags.HasError() {
		return result, diags
	}
	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_0_Drilldowns_Item](
		writes, &result.Presentation.TimeRange, &result.Presentation.HideTitle, &result.Presentation.HideBorder,
		&result.Presentation.References, &result.Presentation.Drilldowns,
	)...)

	return result, diags
}

func preservePieRawDimensions(chart *kbapi.KibanaHTTPAPIsLensApiConfig, m *models.PieChartConfigModel) error {
	raw, err := json.Marshal(chart)
	if err != nil {
		return err
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if len(m.Metrics) > 0 {
		metrics := make([]json.RawMessage, len(m.Metrics))
		for i, metric := range m.Metrics {
			metrics[i] = json.RawMessage(metric.Config.ValueString())
		}
		payload["metrics"], err = json.Marshal(metrics)
		if err != nil {
			return err
		}
	}
	if len(m.GroupBy) > 0 {
		groupBy := make([]json.RawMessage, len(m.GroupBy))
		for i, dimension := range m.GroupBy {
			groupBy[i] = json.RawMessage(dimension.Config.ValueString())
		}
		payload["group_by"], err = json.Marshal(groupBy)
		if err != nil {
			return err
		}
	}
	raw, err = json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, chart)
}
