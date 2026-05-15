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

package dashboard

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func datatableNoESQLConfigFromAPI(
	ctx context.Context,
	m *models.DatatableNoESQLConfigModel,
	dashboard *models.DashboardModel,
	prior *models.DatatableNoESQLConfigModel,
	api kbapi.DatatableNoESQL,
) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

	datasetBytes, err := json.Marshal(api.DataSource)
	dv, ok := marshalToNormalized(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = dv

	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	m.Styling = &models.DatatableStylingModel{}
	if stylingDiags := datatableStylingFromAPI(m.Styling, api.Styling); stylingDiags.HasError() {
		return stylingDiags
	}

	m.Query = &models.FilterSimpleModel{}
	filterSimpleFromAPI(m.Query, api.Query)

	m.Filters = populateFiltersFromAPI(api.Filters, &diags)

	if len(api.Metrics) > 0 {
		m.Metrics = make([]models.DatatableMetricModel, len(api.Metrics))
		for i, metric := range api.Metrics {
			metricBytes, err := json.Marshal(metric)
			mv, ok := marshalToNormalized(metricBytes, err, "metric", &diags)
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
			rv, ok := marshalToNormalized(rowBytes, err, "row", &diags)
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
			sv, ok := marshalToNormalized(splitBytes, err, "split_metrics_by", &diags)
			if !ok {
				return diags
			}
			m.SplitMetricsBy[i].ConfigJSON = sv
		}
	}

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lensDrilldownsAPIToWire(api.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lensChartPresentationReadsFor(ctx, dashboard, priorLens, api.TimeRange, api.HideTitle, api.HideBorder, api.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres

	return diags
}

func datatableNoESQLConfigToAPI(m *models.DatatableNoESQLConfigModel, dashboard *models.DashboardModel) (kbapi.DatatableNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.DatatableNoESQL{Type: kbapi.DatatableNoESQLTypeDataTable}

	if typeutils.IsKnown(m.Title) {
		api.Title = m.Title.ValueStringPointer()
	}

	if typeutils.IsKnown(m.Description) {
		api.Description = m.Description.ValueStringPointer()
	}

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

	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}

	if typeutils.IsKnown(m.Sampling) {
		sampling := float32(m.Sampling.ValueFloat64())
		api.Sampling = &sampling
	}

	if m.Query != nil {
		api.Query = filterSimpleToAPI(m.Query)
	}

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.DatatableNoESQL_Metrics_Item, len(m.Metrics))
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
		rows := make([]kbapi.DatatableNoESQL_Rows_Item, len(m.Rows))
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
		splits := make([]kbapi.DatatableNoESQL_SplitMetricsBy_Item, len(m.SplitMetricsBy))
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

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	api.TimeRange = writes.TimeRange
	if writes.HideTitle != nil {
		api.HideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		api.HideBorder = writes.HideBorder
	}
	if writes.References != nil {
		api.References = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := decodeLensDrilldownSlice[kbapi.DatatableNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}

func datatableESQLConfigFromAPI(
	ctx context.Context,
	m *models.DatatableESQLConfigModel,
	dashboard *models.DashboardModel,
	prior *models.DatatableESQLConfigModel,
	api kbapi.DatatableESQL,
) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

	datasetBytes, err := json.Marshal(api.DataSource)
	dv, ok := marshalToNormalized(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = dv

	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	m.Styling = &models.DatatableStylingModel{}
	if stylingDiags := datatableStylingFromAPI(m.Styling, api.Styling); stylingDiags.HasError() {
		return stylingDiags
	}

	m.Filters = populateFiltersFromAPI(api.Filters, &diags)

	if api.Metrics != nil && len(*api.Metrics) > 0 {
		m.Metrics = make([]models.DatatableMetricModel, len(*api.Metrics))
		for i, metric := range *api.Metrics {
			metricBytes, err := json.Marshal(metric)
			mv, ok := marshalToNormalized(metricBytes, err, "metric", &diags)
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
			rv, ok := marshalToNormalized(rowBytes, err, "row", &diags)
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
			sv, ok := marshalToNormalized(splitBytes, err, "split_metrics_by", &diags)
			if !ok {
				return diags
			}
			m.SplitMetricsBy[i].ConfigJSON = sv
		}
	}

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lensDrilldownsAPIToWire(api.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lensChartPresentationReadsFor(ctx, dashboard, priorLens, api.TimeRange, api.HideTitle, api.HideBorder, api.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres

	return diags
}

func datatableESQLConfigToAPI(m *models.DatatableESQLConfigModel, dashboard *models.DashboardModel) (kbapi.DatatableESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.DatatableESQL{Type: kbapi.DatatableESQLTypeDataTable}

	if typeutils.IsKnown(m.Title) {
		api.Title = m.Title.ValueStringPointer()
	}

	if typeutils.IsKnown(m.Description) {
		api.Description = m.Description.ValueStringPointer()
	}

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

	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}

	if typeutils.IsKnown(m.Sampling) {
		sampling := float32(m.Sampling.ValueFloat64())
		api.Sampling = &sampling
	}

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.DatatableESQLMetric, len(m.Metrics))
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
			Alignment    *kbapi.DatatableESQLRowsAlignment    `json:"alignment,omitempty"`
			ApplyColorTo *kbapi.DatatableESQLRowsApplyColorTo `json:"apply_color_to,omitempty"`
			ClickFilter  *bool                                `json:"click_filter,omitempty"`
			CollapseBy   kbapi.CollapseBy                     `json:"collapse_by"`
			Color        *kbapi.DatatableESQL_Rows_Color      `json:"color,omitempty"`
			Column       string                               `json:"column"`
			Format       kbapi.FormatType                     `json:"format"`
			Label        *string                              `json:"label,omitempty"`
			Visible      *bool                                `json:"visible,omitempty"`
			Width        *float32                             `json:"width,omitempty"`
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
			Column string           `json:"column"`
			Format kbapi.FormatType `json:"format"`
			Label  *string          `json:"label,omitempty"`
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

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	api.TimeRange = writes.TimeRange
	if writes.HideTitle != nil {
		api.HideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		api.HideBorder = writes.HideBorder
	}
	if writes.References != nil {
		api.References = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := decodeLensDrilldownSlice[kbapi.DatatableESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}

func datatableStylingFromAPI(m *models.DatatableStylingModel, api kbapi.DatatableStyling) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Density = &models.DatatableDensityModel{}
	if densityDiags := datatableDensityFromAPI(m.Density, api.Density); densityDiags.HasError() {
		return densityDiags
	}

	if api.SortBy != nil {
		sortBytes, err := json.Marshal(api.SortBy)
		sortV, ok := marshalToNormalized(sortBytes, err, "sort_by", &diags)
		if !ok {
			return diags
		}
		m.SortByJSON = sortV
	} else {
		m.SortByJSON = jsontypes.NewNormalizedNull()
	}

	if api.Paging != nil {
		m.Paging = types.Int64Value(int64(*api.Paging))
	} else {
		m.Paging = types.Int64Null()
	}

	return diags
}

func datatableStylingToAPI(m *models.DatatableStylingModel) (kbapi.DatatableStyling, diag.Diagnostics) {
	if m == nil {
		return kbapi.DatatableStyling{}, nil
	}

	var diags diag.Diagnostics
	var styling kbapi.DatatableStyling

	if m.Density != nil {
		density, densityDiags := datatableDensityToAPI(m.Density)
		diags.Append(densityDiags...)
		if diags.HasError() {
			return styling, diags
		}
		styling.Density = density
	}

	if typeutils.IsKnown(m.SortByJSON) {
		var sortBy kbapi.DatatableStyling_SortBy
		if err := json.Unmarshal([]byte(m.SortByJSON.ValueString()), &sortBy); err != nil {
			diags.AddError("Failed to unmarshal sort_by", err.Error())
			return styling, diags
		}
		styling.SortBy = &sortBy
	}

	if typeutils.IsKnown(m.Paging) {
		paging := kbapi.DatatableStylingPaging(m.Paging.ValueInt64())
		styling.Paging = &paging
	}

	return styling, diags
}

func datatableDensityFromAPI(m *models.DatatableDensityModel, api kbapi.DatatableDensity) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Mode = typeutils.StringishPointerValue(api.Mode)

	if api.Height != nil {
		m.Height = &models.DatatableDensityHeightModel{}
		heightDiags := datatableDensityHeightFromAPI(m.Height, api.Height)
		diags.Append(heightDiags...)
	}

	return diags
}

func datatableDensityToAPI(m *models.DatatableDensityModel) (kbapi.DatatableDensity, diag.Diagnostics) {
	if m == nil {
		return kbapi.DatatableDensity{}, nil
	}

	var diags diag.Diagnostics
	var density kbapi.DatatableDensity

	if typeutils.IsKnown(m.Mode) {
		mode := kbapi.DatatableDensityMode(m.Mode.ValueString())
		density.Mode = &mode
	}

	if m.Height != nil {
		height := &struct {
			Header *kbapi.DatatableDensity_Height_Header `json:"header,omitempty"`
			Value  *kbapi.DatatableDensity_Height_Value  `json:"value,omitempty"`
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
	Header *kbapi.DatatableDensity_Height_Header `json:"header,omitempty"`
	Value  *kbapi.DatatableDensity_Height_Value  `json:"value,omitempty"`
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

func datatableDensityHeightHeaderFromAPI(m *models.DatatableDensityHeightHeaderModel, api *kbapi.DatatableDensity_Height_Header) diag.Diagnostics {
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

func datatableDensityHeightHeaderToAPI(m *models.DatatableDensityHeightHeaderModel) (*kbapi.DatatableDensity_Height_Header, diag.Diagnostics) {
	if m == nil || !typeutils.IsKnown(m.Type) {
		return nil, nil
	}

	var diags diag.Diagnostics
	var header kbapi.DatatableDensity_Height_Header

	switch m.Type.ValueString() {
	case "auto":
		auto := kbapi.DatatableDensityHeightHeader0{Type: kbapi.DatatableDensityHeightHeader0TypeAuto}
		if err := header.FromDatatableDensityHeightHeader0(auto); err != nil {
			diags.AddError("Failed to marshal header density", err.Error())
			return nil, diags
		}
	case "custom":
		custom := kbapi.DatatableDensityHeightHeader1{Type: kbapi.DatatableDensityHeightHeader1TypeCustom}
		if typeutils.IsKnown(m.MaxLines) {
			maxLines := float32(m.MaxLines.ValueFloat64())
			custom.MaxLines = &maxLines
		}
		if err := header.FromDatatableDensityHeightHeader1(custom); err != nil {
			diags.AddError("Failed to marshal header density", err.Error())
			return nil, diags
		}
	default:
		return nil, diags
	}

	return &header, diags
}

func datatableDensityHeightValueFromAPI(m *models.DatatableDensityHeightValueModel, api *kbapi.DatatableDensity_Height_Value) diag.Diagnostics {
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

func datatableDensityHeightValueToAPI(m *models.DatatableDensityHeightValueModel) (*kbapi.DatatableDensity_Height_Value, diag.Diagnostics) {
	if m == nil || !typeutils.IsKnown(m.Type) {
		return nil, nil
	}

	var diags diag.Diagnostics
	var value kbapi.DatatableDensity_Height_Value

	switch m.Type.ValueString() {
	case "auto":
		auto := kbapi.DatatableDensityHeightValue0{Type: kbapi.DatatableDensityHeightValue0TypeAuto}
		if err := value.FromDatatableDensityHeightValue0(auto); err != nil {
			diags.AddError("Failed to marshal value density", err.Error())
			return nil, diags
		}
	case "custom":
		custom := kbapi.DatatableDensityHeightValue1{Type: kbapi.DatatableDensityHeightValue1TypeCustom}
		if typeutils.IsKnown(m.Lines) {
			lines := float32(m.Lines.ValueFloat64())
			custom.Lines = &lines
		}
		if err := value.FromDatatableDensityHeightValue1(custom); err != nil {
			diags.AddError("Failed to marshal value density", err.Error())
			return nil, diags
		}
	default:
		return nil, diags
	}

	return &value, diags
}
