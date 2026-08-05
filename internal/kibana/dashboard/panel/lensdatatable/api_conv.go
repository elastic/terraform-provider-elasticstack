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

package lensdatatable

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func datatableNoESQLConfigFromAPI(
	ctx context.Context,
	m *models.DatatableNoESQLConfigModel,
	prior *models.DatatableNoESQLConfigModel,
	api kbapi.KibanaHTTPAPIsDatatableNoESQL,
	presentation kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0,
) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	base, ok := lenscommon.PopulateLensChartBaseFromAPI(
		api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling,
		datasetBytes, datasetErr, "data_source_json", api.Filters, &diags,
	)
	if !ok {
		return diags
	}
	m.LensChartBaseTFModel = base

	m.Styling = &models.DatatableStylingModel{}
	if stylingDiags := datatableStylingFromAPI(m.Styling, api.Styling); stylingDiags.HasError() {
		return stylingDiags
	}

	m.Query = &models.FilterSimpleModel{}
	lenscommon.FilterSimpleFromAPI(m.Query, api.Query)

	if len(api.Metrics) > 0 {
		m.Metrics = make([]models.DatatableMetricModel, len(api.Metrics))
		for i, metric := range api.Metrics {
			metricBytes, err := json.Marshal(metric)
			mv, ok := lenscommon.WrapNormalizedJSON(metricBytes, err, "metric", &diags)
			if !ok {
				return diags
			}
			m.Metrics[i].ConfigJSON = mv
		}
	}

	if api.Rows != nil && len(*api.Rows) > 0 {
		m.Rows = make([]models.DatatableRowModel, len(*api.Rows))
		for i, row := range *api.Rows {
			rowBytes, err := json.Marshal(row)
			rv, ok := lenscommon.WrapNormalizedJSON(rowBytes, err, "row", &diags)
			if !ok {
				return diags
			}
			m.Rows[i].ConfigJSON = rv
		}
	}

	if api.SplitMetricsBy != nil && len(*api.SplitMetricsBy) > 0 {
		m.SplitMetricsBy = make([]models.DatatableSplitByModel, len(*api.SplitMetricsBy))
		for i, splitBy := range *api.SplitMetricsBy {
			splitBytes, err := json.Marshal(splitBy)
			sv, ok := lenscommon.WrapNormalizedJSON(splitBytes, err, "split_metrics_by", &diags)
			if !ok {
				return diags
			}
			m.SplitMetricsBy[i].ConfigJSON = sv
		}
	}

	if !lenscommon.PopulateLensChartPresentationFromAPI(ctx, &m.LensChartPresentationTFModel, prior, presentation, &diags) {
		return diags
	}

	return diags
}

func datatableNoESQLConfigToAPI(m *models.DatatableNoESQLConfigModel) (kbapi.KibanaHTTPAPIsDatatableNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KibanaHTTPAPIsDatatableNoESQL{Type: kbapi.KibanaHTTPAPIsDatatableNoESQLTypeDataTable}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	if typeutils.IsKnown(m.DataSourceJSON) {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal datatable_config.no_esql.data_source_json", err.Error())
			return api, diags
		}
	}

	if m.Styling != nil {
		styling, stylingDiags := datatableStylingToAPI(m.Styling)
		diags.Append(stylingDiags...)
		if diags.HasError() {
			return api, diags
		}
		api.Styling = styling
	}

	if m.Query != nil {
		api.Query = lenscommon.FilterSimpleToAPI(m.Query)
	}

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if len(m.Metrics) > 0 {
		metrics := make([]struct {
			Alignment    *kbapi.KibanaHTTPAPIsDatatableNoESQLMetricsAlignment      `json:"alignment,omitempty"`
			ApplyColorTo *kbapi.KibanaHTTPAPIsDatatableNoESQL_Metrics_ApplyColorTo `json:"apply_color_to,omitempty"`
			Color        *kbapi.KibanaHTTPAPIsDatatableNoESQL_Metrics_Color        `json:"color,omitempty"`
			Summary      *struct {
				Label *string                                                  `json:"label,omitempty"`
				Type  kbapi.KibanaHTTPAPIsDatatableNoESQL_Metrics_Summary_Type `json:"type"`
			} `json:"summary,omitempty"`
			Visible *bool    `json:"visible,omitempty"`
			Width   *float32 `json:"width,omitempty"`
		}, len(m.Metrics))
		for i, metricModel := range m.Metrics {
			if typeutils.IsKnown(metricModel.ConfigJSON) {
				if err := json.Unmarshal([]byte(metricModel.ConfigJSON.ValueString()), &metrics[i]); err != nil {
					diags.AddError("Failed to unmarshal metric", err.Error())
					return api, diags
				}
			}
		}
		api.Metrics = metrics
	}

	if len(m.Rows) > 0 {
		rows := make([]struct {
			Alignment    *kbapi.KibanaHTTPAPIsDatatableNoESQLRowsAlignment      `json:"alignment,omitempty"`
			ApplyColorTo *kbapi.KibanaHTTPAPIsDatatableNoESQL_Rows_ApplyColorTo `json:"apply_color_to,omitempty"`
			ClickFilter  *bool                                                  `json:"click_filter,omitempty"`
			CollapseBy   *kbapi.KibanaHTTPAPIsCollapseBy                        `json:"collapse_by,omitempty"`
			Color        *kbapi.KibanaHTTPAPIsDatatableNoESQL_Rows_Color        `json:"color,omitempty"`
			Visible      *bool                                                  `json:"visible,omitempty"`
			Width        *float32                                               `json:"width,omitempty"`
		}, len(m.Rows))
		for i, rowModel := range m.Rows {
			if typeutils.IsKnown(rowModel.ConfigJSON) {
				if err := json.Unmarshal([]byte(rowModel.ConfigJSON.ValueString()), &rows[i]); err != nil {
					diags.AddError("Failed to unmarshal row", err.Error())
					return api, diags
				}
			}
		}
		api.Rows = &rows
	}

	if len(m.SplitMetricsBy) > 0 {
		splits := make([]kbapi.KibanaHTTPAPIsDatatableNoESQL_SplitMetricsBy_Item, len(m.SplitMetricsBy))
		for i, splitModel := range m.SplitMetricsBy {
			if typeutils.IsKnown(splitModel.ConfigJSON) {
				if err := json.Unmarshal([]byte(splitModel.ConfigJSON.ValueString()), &splits[i]); err != nil {
					diags.AddError("Failed to unmarshal split_metrics_by", err.Error())
					return api, diags
				}
			}
		}
		api.SplitMetricsBy = &splits
	}

	return api, diags
}

func datatableESQLConfigFromAPI(
	ctx context.Context,
	m *models.DatatableESQLConfigModel,
	prior *models.DatatableESQLConfigModel,
	api kbapi.KibanaHTTPAPIsDatatableESQL,
	presentation kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0,
) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	base, ok := lenscommon.PopulateLensChartBaseFromAPI(
		api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling,
		datasetBytes, datasetErr, "data_source_json", api.Filters, &diags,
	)
	if !ok {
		return diags
	}
	m.LensChartBaseTFModel = base

	m.Styling = &models.DatatableStylingModel{}
	if stylingDiags := datatableStylingFromAPI(m.Styling, api.Styling); stylingDiags.HasError() {
		return stylingDiags
	}

	if api.Metrics != nil && len(*api.Metrics) > 0 {
		m.Metrics = make([]models.DatatableMetricModel, len(*api.Metrics))
		for i, metric := range *api.Metrics {
			metricBytes, err := json.Marshal(metric)
			mv, ok := lenscommon.WrapNormalizedJSON(metricBytes, err, "metric", &diags)
			if !ok {
				return diags
			}
			m.Metrics[i].ConfigJSON = mv
		}
	}

	if api.Rows != nil && len(*api.Rows) > 0 {
		m.Rows = make([]models.DatatableRowModel, len(*api.Rows))
		for i, row := range *api.Rows {
			rowBytes, err := json.Marshal(row)
			rv, ok := lenscommon.WrapNormalizedJSON(rowBytes, err, "row", &diags)
			if !ok {
				return diags
			}
			m.Rows[i].ConfigJSON = rv
		}
	}

	if api.SplitMetricsBy != nil && len(*api.SplitMetricsBy) > 0 {
		m.SplitMetricsBy = make([]models.DatatableSplitByModel, len(*api.SplitMetricsBy))
		for i, splitBy := range *api.SplitMetricsBy {
			splitBytes, err := json.Marshal(splitBy)
			sv, ok := lenscommon.WrapNormalizedJSON(splitBytes, err, "split_metrics_by", &diags)
			if !ok {
				return diags
			}
			m.SplitMetricsBy[i].ConfigJSON = sv
		}
	}

	if !lenscommon.PopulateLensChartPresentationFromAPI(ctx, &m.LensChartPresentationTFModel, prior, presentation, &diags) {
		return diags
	}

	return diags
}

func datatableESQLConfigToAPI(m *models.DatatableESQLConfigModel) (kbapi.KibanaHTTPAPIsDatatableESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KibanaHTTPAPIsDatatableESQL{Type: kbapi.KibanaHTTPAPIsDatatableESQLTypeDataTable}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	if typeutils.IsKnown(m.DataSourceJSON) {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal datatable_config.esql.data_source_json", err.Error())
			return api, diags
		}
	}

	if m.Styling != nil {
		styling, stylingDiags := datatableStylingToAPI(m.Styling)
		diags.Append(stylingDiags...)
		if diags.HasError() {
			return api, diags
		}
		api.Styling = styling
	}

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.KibanaHTTPAPIsDatatableESQLMetric, len(m.Metrics))
		for i, metricModel := range m.Metrics {
			if typeutils.IsKnown(metricModel.ConfigJSON) {
				if err := json.Unmarshal([]byte(metricModel.ConfigJSON.ValueString()), &metrics[i]); err != nil {
					diags.AddError("Failed to unmarshal metric", err.Error())
					return api, diags
				}
			}
		}
		api.Metrics = &metrics
	}

	if len(m.Rows) > 0 {
		rows := make([]struct {
			Alignment    *kbapi.KibanaHTTPAPIsDatatableESQLRowsAlignment      `json:"alignment,omitempty"`
			ApplyColorTo *kbapi.KibanaHTTPAPIsDatatableESQL_Rows_ApplyColorTo `json:"apply_color_to,omitempty"`
			ClickFilter  *bool                                                `json:"click_filter,omitempty"`
			CollapseBy   *kbapi.KibanaHTTPAPIsCollapseBy                      `json:"collapse_by,omitempty"`
			Color        *kbapi.KibanaHTTPAPIsDatatableESQL_Rows_Color        `json:"color,omitempty"`
			Column       string                                               `json:"column"`
			Format       *kbapi.KibanaHTTPAPIsFormatType                      `json:"format,omitempty"`
			Label        *string                                              `json:"label,omitempty"`
			Visible      *bool                                                `json:"visible,omitempty"`
			Width        *float32                                             `json:"width,omitempty"`
		}, len(m.Rows))
		for i, rowModel := range m.Rows {
			if typeutils.IsKnown(rowModel.ConfigJSON) {
				if err := json.Unmarshal([]byte(rowModel.ConfigJSON.ValueString()), &rows[i]); err != nil {
					diags.AddError("Failed to unmarshal row", err.Error())
					return api, diags
				}
			}
		}
		api.Rows = &rows
	}

	if len(m.SplitMetricsBy) > 0 {
		splits := make([]struct {
			Column string                          `json:"column"`
			Format *kbapi.KibanaHTTPAPIsFormatType `json:"format,omitempty"`
			Label  *string                         `json:"label,omitempty"`
		}, len(m.SplitMetricsBy))
		for i, splitModel := range m.SplitMetricsBy {
			if typeutils.IsKnown(splitModel.ConfigJSON) {
				if err := json.Unmarshal([]byte(splitModel.ConfigJSON.ValueString()), &splits[i]); err != nil {
					diags.AddError("Failed to unmarshal split_metrics_by", err.Error())
					return api, diags
				}
			}
		}
		api.SplitMetricsBy = &splits
	}

	return api, diags
}

func datatableStylingFromAPI(m *models.DatatableStylingModel, api *kbapi.KibanaHTTPAPIsDatatableStyling) diag.Diagnostics {
	var diags diag.Diagnostics
	if api == nil {
		return diags
	}

	m.Density = &models.DatatableDensityModel{}
	if densityDiags := datatableDensityFromAPI(m.Density, api.Density); densityDiags.HasError() {
		return densityDiags
	}

	if api.SortBy != nil {
		sortBytes, err := json.Marshal(api.SortBy)
		sortV, ok := lenscommon.WrapNormalizedJSON(sortBytes, err, "sort_by", &diags)
		if !ok {
			return diags
		}
		m.SortByJSON = sortV
	} else {
		m.SortByJSON = jsontypes.NewNormalizedNull()
	}

	if api.Paging != nil {
		var paging float32
		raw, err := json.Marshal(api.Paging)
		if err != nil {
			diags.AddError("Failed to marshal paging", err.Error())
			return diags
		}
		if err := json.Unmarshal(raw, &paging); err != nil {
			diags.AddError("Failed to unmarshal paging", err.Error())
			return diags
		}
		m.Paging = types.Int64Value(int64(paging))
	} else {
		m.Paging = types.Int64Null()
	}

	return diags
}

func datatableStylingToAPI(m *models.DatatableStylingModel) (*kbapi.KibanaHTTPAPIsDatatableStyling, diag.Diagnostics) {
	if m == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	styling := &kbapi.KibanaHTTPAPIsDatatableStyling{}

	if m.Density != nil {
		density, densityDiags := datatableDensityToAPI(m.Density)
		diags.Append(densityDiags...)
		if diags.HasError() {
			return styling, diags
		}
		styling.Density = density
	}

	if typeutils.IsKnown(m.SortByJSON) {
		var sortBy kbapi.KibanaHTTPAPIsDatatableStyling_SortBy
		if err := json.Unmarshal([]byte(m.SortByJSON.ValueString()), &sortBy); err != nil {
			diags.AddError("Failed to unmarshal sort_by", err.Error())
			return styling, diags
		}
		styling.SortBy = &sortBy
	}

	if typeutils.IsKnown(m.Paging) {
		var paging kbapi.KibanaHTTPAPIsDatatableStyling_Paging
		raw, err := json.Marshal(m.Paging.ValueInt64())
		if err != nil {
			diags.AddError("Failed to marshal paging", err.Error())
			return styling, diags
		}
		if err := json.Unmarshal(raw, &paging); err != nil {
			diags.AddError("Failed to unmarshal paging", err.Error())
			return styling, diags
		}
		styling.Paging = &paging
	}

	return styling, diags
}

func datatableDensityFromAPI(m *models.DatatableDensityModel, api *kbapi.KibanaHTTPAPIsDatatableDensity) diag.Diagnostics {
	var diags diag.Diagnostics
	if api == nil {
		return diags
	}

	if api.Mode != nil {
		var mode string
		raw, err := json.Marshal(api.Mode)
		if err != nil {
			diags.AddError("Failed to marshal density mode", err.Error())
			return diags
		}
		if err := json.Unmarshal(raw, &mode); err != nil {
			diags.AddError("Failed to unmarshal density mode", err.Error())
			return diags
		}
		m.Mode = types.StringValue(mode)
	} else {
		m.Mode = types.StringNull()
	}

	if api.Height != nil {
		m.Height = &models.DatatableDensityHeightModel{}
		heightDiags := datatableDensityHeightFromAPI(m.Height, api.Height)
		diags.Append(heightDiags...)
	}

	return diags
}

func datatableDensityToAPI(m *models.DatatableDensityModel) (*kbapi.KibanaHTTPAPIsDatatableDensity, diag.Diagnostics) {
	if m == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	density := &kbapi.KibanaHTTPAPIsDatatableDensity{}

	if typeutils.IsKnown(m.Mode) {
		var mode kbapi.KibanaHTTPAPIsDatatableDensity_Mode
		raw, err := json.Marshal(m.Mode.ValueString())
		if err != nil {
			diags.AddError("Failed to marshal density mode", err.Error())
			return density, diags
		}
		if err := json.Unmarshal(raw, &mode); err != nil {
			diags.AddError("Failed to unmarshal density mode", err.Error())
			return density, diags
		}
		density.Mode = &mode
	}

	if m.Height != nil {
		height := &struct {
			Header *kbapi.KibanaHTTPAPIsDatatableDensity_Height_Header `json:"header,omitempty"`
			Value  *kbapi.KibanaHTTPAPIsDatatableDensity_Height_Value  `json:"value,omitempty"`
		}{}

		if m.Height.Header != nil {
			header, headerDiags := datatableDensityHeightHeaderToAPI(m.Height.Header)
			diags.Append(headerDiags...)
			if diags.HasError() {
				return density, diags
			}
			height.Header = header
		}

		if m.Height.Value != nil {
			value, valueDiags := datatableDensityHeightValueToAPI(m.Height.Value)
			diags.Append(valueDiags...)
			if diags.HasError() {
				return density, diags
			}
			height.Value = value
		}

		density.Height = height
	}

	return density, diags
}

func datatableDensityHeightFromAPI(m *models.DatatableDensityHeightModel, api *struct {
	Header *kbapi.KibanaHTTPAPIsDatatableDensity_Height_Header `json:"header,omitempty"`
	Value  *kbapi.KibanaHTTPAPIsDatatableDensity_Height_Value  `json:"value,omitempty"`
}) diag.Diagnostics {
	var diags diag.Diagnostics
	if api == nil {
		return diags
	}

	if api.Header != nil {
		m.Header = &models.DatatableDensityHeightHeaderModel{}
		headerDiags := datatableDensityHeightHeaderFromAPI(m.Header, api.Header)
		diags.Append(headerDiags...)
	}

	if api.Value != nil {
		m.Value = &models.DatatableDensityHeightValueModel{}
		valueDiags := datatableDensityHeightValueFromAPI(m.Value, api.Value)
		diags.Append(valueDiags...)
	}

	return diags
}

func datatableDensityHeightHeaderFromAPI(m *models.DatatableDensityHeightHeaderModel, api *kbapi.KibanaHTTPAPIsDatatableDensity_Height_Header) diag.Diagnostics {
	var diags diag.Diagnostics
	if api == nil {
		return diags
	}

	raw, err := api.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal header density", err.Error())
		return diags
	}

	var header struct {
		Type     string   `json:"type"`
		MaxLines *float32 `json:"max_lines,omitempty"`
	}

	if err := json.Unmarshal(raw, &header); err != nil {
		diags.AddError("Failed to unmarshal header density", err.Error())
		return diags
	}

	m.Type = types.StringValue(header.Type)
	if header.MaxLines != nil {
		m.MaxLines = types.Float64Value(float64(*header.MaxLines))
	} else {
		m.MaxLines = types.Float64Null()
	}

	return diags
}

func datatableDensityHeightHeaderToAPI(m *models.DatatableDensityHeightHeaderModel) (*kbapi.KibanaHTTPAPIsDatatableDensity_Height_Header, diag.Diagnostics) {
	if m == nil || !typeutils.IsKnown(m.Type) {
		return nil, nil
	}

	var diags diag.Diagnostics
	var header kbapi.KibanaHTTPAPIsDatatableDensity_Height_Header

	switch m.Type.ValueString() {
	case "auto":
		auto := kbapi.KibanaHTTPAPIsDatatableDensityHeightHeader0{Type: kbapi.KibanaHTTPAPIsDatatableDensityHeightHeader0TypeAuto}
		if err := header.FromKibanaHTTPAPIsDatatableDensityHeightHeader0(auto); err != nil {
			diags.AddError("Failed to marshal header density", err.Error())
			return nil, diags
		}
	case "custom":
		custom := kbapi.KibanaHTTPAPIsDatatableDensityHeightHeader1{Type: kbapi.KibanaHTTPAPIsDatatableDensityHeightHeader1TypeCustom}
		if typeutils.IsKnown(m.MaxLines) {
			maxLines := float32(m.MaxLines.ValueFloat64())
			custom.MaxLines = &maxLines
		}
		if err := header.FromKibanaHTTPAPIsDatatableDensityHeightHeader1(custom); err != nil {
			diags.AddError("Failed to marshal header density", err.Error())
			return nil, diags
		}
	default:
		return nil, diags
	}

	return &header, diags
}

func datatableDensityHeightValueFromAPI(m *models.DatatableDensityHeightValueModel, api *kbapi.KibanaHTTPAPIsDatatableDensity_Height_Value) diag.Diagnostics {
	var diags diag.Diagnostics
	if api == nil {
		return diags
	}

	raw, err := api.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal value density", err.Error())
		return diags
	}

	var value struct {
		Type  string   `json:"type"`
		Lines *float32 `json:"lines,omitempty"`
	}

	if err := json.Unmarshal(raw, &value); err != nil {
		diags.AddError("Failed to unmarshal value density", err.Error())
		return diags
	}

	m.Type = types.StringValue(value.Type)
	if value.Lines != nil {
		m.Lines = types.Float64Value(float64(*value.Lines))
	} else {
		m.Lines = types.Float64Null()
	}

	return diags
}

func datatableDensityHeightValueToAPI(m *models.DatatableDensityHeightValueModel) (*kbapi.KibanaHTTPAPIsDatatableDensity_Height_Value, diag.Diagnostics) {
	if m == nil || !typeutils.IsKnown(m.Type) {
		return nil, nil
	}

	var diags diag.Diagnostics
	var value kbapi.KibanaHTTPAPIsDatatableDensity_Height_Value

	switch m.Type.ValueString() {
	case "auto":
		auto := kbapi.KibanaHTTPAPIsDatatableDensityHeightValue0{Type: kbapi.KibanaHTTPAPIsDatatableDensityHeightValue0TypeAuto}
		if err := value.FromKibanaHTTPAPIsDatatableDensityHeightValue0(auto); err != nil {
			diags.AddError("Failed to marshal value density", err.Error())
			return nil, diags
		}
	case "custom":
		custom := kbapi.KibanaHTTPAPIsDatatableDensityHeightValue1{Type: kbapi.KibanaHTTPAPIsDatatableDensityHeightValue1TypeCustom}
		if typeutils.IsKnown(m.Lines) {
			lines := float32(m.Lines.ValueFloat64())
			custom.Lines = &lines
		}
		if err := value.FromKibanaHTTPAPIsDatatableDensityHeightValue1(custom); err != nil {
			diags.AddError("Failed to marshal value density", err.Error())
			return nil, diags
		}
	default:
		return nil, diags
	}

	return &value, diags
}
