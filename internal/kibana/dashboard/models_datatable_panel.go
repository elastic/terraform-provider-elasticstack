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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newDatatablePanelConfigConverter() datatablePanelConfigConverter {
	return datatablePanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.DatatableNoESQLTypeDataTable),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.DatatableConfig != nil },
		},
	}
}

type datatablePanelConfigConverter struct {
	lensVisualizationBase
}

func (c datatablePanelConfigConverter) populateFromAttributes(ctx context.Context, pm *panelModel, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	pm.DatatableConfig = &datatableConfigModel{}

	if datatableNoESQL, err := attrs.AsDatatableNoESQL(); err == nil && !isDatatableNoESQLCandidateActuallyESQL(datatableNoESQL) {
		pm.DatatableConfig.NoESQL = &datatableNoESQLConfigModel{}
		return pm.DatatableConfig.NoESQL.fromAPI(ctx, datatableNoESQL)
	}
	datatableESQL, err := attrs.AsDatatableESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.DatatableConfig.ESQL = &datatableESQLConfigModel{}
	return pm.DatatableConfig.ESQL.fromAPI(ctx, datatableESQL)
}

func (c datatablePanelConfigConverter) buildAttributes(pm panelModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	if pm.DatatableConfig == nil {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0

	switch {
	case pm.DatatableConfig.NoESQL != nil:
		noESQL, noDiags := pm.DatatableConfig.NoESQL.toAPI()
		diags.Append(noDiags...)
		if diags.HasError() {
			return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
		}

		if err := attrs.FromDatatableNoESQL(noESQL); err != nil {
			diags.AddError("Failed to convert datatable no-esql config", err.Error())
			return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
		}
	case pm.DatatableConfig.ESQL != nil:
		esql, esqlDiags := pm.DatatableConfig.ESQL.toAPI()
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
		}

		if err := attrs.FromDatatableESQL(esql); err != nil {
			diags.AddError("Failed to convert datatable esql config", err.Error())
			return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
		}
	default:
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	return attrs, diags
}

type datatableConfigModel struct {
	NoESQL *datatableNoESQLConfigModel `tfsdk:"no_esql"`
	ESQL   *datatableESQLConfigModel   `tfsdk:"esql"`
}

type datatableNoESQLConfigModel struct {
	Title               types.String            `tfsdk:"title"`
	Description         types.String            `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized    `tfsdk:"data_source_json"`
	Density             *datatableDensityModel  `tfsdk:"density"`
	IgnoreGlobalFilters types.Bool              `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64           `tfsdk:"sampling"`
	Query               *filterSimpleModel      `tfsdk:"query"`
	Filters             []chartFilterJSONModel  `tfsdk:"filters"`
	Metrics             []datatableMetricModel  `tfsdk:"metrics"`
	Rows                []datatableRowModel     `tfsdk:"rows"`
	SplitMetricsBy      []datatableSplitByModel `tfsdk:"split_metrics_by"`
	SortByJSON          jsontypes.Normalized    `tfsdk:"sort_by_json"`
	Paging              types.Int64             `tfsdk:"paging"`
}

type datatableESQLConfigModel struct {
	Title               types.String            `tfsdk:"title"`
	Description         types.String            `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized    `tfsdk:"data_source_json"`
	Density             *datatableDensityModel  `tfsdk:"density"`
	IgnoreGlobalFilters types.Bool              `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64           `tfsdk:"sampling"`
	Filters             []chartFilterJSONModel  `tfsdk:"filters"`
	Metrics             []datatableMetricModel  `tfsdk:"metrics"`
	Rows                []datatableRowModel     `tfsdk:"rows"`
	SplitMetricsBy      []datatableSplitByModel `tfsdk:"split_metrics_by"`
	SortByJSON          jsontypes.Normalized    `tfsdk:"sort_by_json"`
	Paging              types.Int64             `tfsdk:"paging"`
}

type datatableMetricModel struct {
	ConfigJSON jsontypes.Normalized `tfsdk:"config_json"`
}

type datatableRowModel struct {
	ConfigJSON jsontypes.Normalized `tfsdk:"config_json"`
}

type datatableSplitByModel struct {
	ConfigJSON jsontypes.Normalized `tfsdk:"config_json"`
}

func isDatatableNoESQLCandidateActuallyESQL(apiTable kbapi.DatatableNoESQL) bool {
	body, err := json.Marshal(apiTable.DataSource)
	if err != nil {
		return false
	}

	var dataset struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &dataset); err != nil {
		return false
	}

	return dataset.Type == legacyMetricDatasetTypeESQL || dataset.Type == legacyMetricDatasetTypeTable
}

type datatableDensityModel struct {
	Mode   types.String                 `tfsdk:"mode"`
	Height *datatableDensityHeightModel `tfsdk:"height"`
}

type datatableDensityHeightModel struct {
	Header *datatableDensityHeightHeaderModel `tfsdk:"header"`
	Value  *datatableDensityHeightValueModel  `tfsdk:"value"`
}

type datatableDensityHeightHeaderModel struct {
	Type     types.String  `tfsdk:"type"`
	MaxLines types.Float64 `tfsdk:"max_lines"`
}

type datatableDensityHeightValueModel struct {
	Type  types.String  `tfsdk:"type"`
	Lines types.Float64 `tfsdk:"lines"`
}

func (m *datatableNoESQLConfigModel) fromAPI(ctx context.Context, api kbapi.DatatableNoESQL) diag.Diagnostics {
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

	m.Density = &datatableDensityModel{}
	if densityDiags := m.Density.fromAPI(api.Density); densityDiags.HasError() {
		return densityDiags
	}

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	m.Filters = populateFiltersFromAPI(api.Filters, &diags)

	if len(api.Metrics) > 0 {
		m.Metrics = make([]datatableMetricModel, len(api.Metrics))
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
		m.Rows = make([]datatableRowModel, len(*api.Rows))
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
		m.SplitMetricsBy = make([]datatableSplitByModel, len(*api.SplitMetricsBy))
		for i, splitBy := range *api.SplitMetricsBy {
			splitBytes, err := json.Marshal(splitBy)
			sv, ok := marshalToNormalized(splitBytes, err, "split_metrics_by", &diags)
			if !ok {
				return diags
			}
			m.SplitMetricsBy[i].ConfigJSON = sv
		}
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

func (m *datatableNoESQLConfigModel) toAPI() (kbapi.DatatableNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.DatatableNoESQL{Type: kbapi.DatatableNoESQLTypeDataTable}
	api.TimeRange = lensPanelTimeRange()

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

	if m.Density != nil {
		density, densityDiags := m.Density.toAPI()
		diags.Append(densityDiags...)
		if diags.HasError() {
			return api, diags
		}
		api.Density = density
	}

	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}

	if typeutils.IsKnown(m.Sampling) {
		sampling := float32(m.Sampling.ValueFloat64())
		api.Sampling = &sampling
	}

	if m.Query != nil {
		api.Query = m.Query.toAPI()
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

	if typeutils.IsKnown(m.SortByJSON) {
		var sortBy kbapi.DatatableNoESQL_SortBy
		if err := json.Unmarshal([]byte(m.SortByJSON.ValueString()), &sortBy); err != nil {
			diags.AddError("Failed to unmarshal sort_by", err.Error())
			return api, diags
		}
		api.SortBy = &sortBy
	}

	if typeutils.IsKnown(m.Paging) {
		paging := kbapi.DatatableNoESQLPaging(m.Paging.ValueInt64())
		api.Paging = &paging
	}

	return api, diags
}

func (m *datatableESQLConfigModel) fromAPI(ctx context.Context, api kbapi.DatatableESQL) diag.Diagnostics {
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

	m.Density = &datatableDensityModel{}
	if densityDiags := m.Density.fromAPI(api.Density); densityDiags.HasError() {
		return densityDiags
	}

	m.Filters = populateFiltersFromAPI(api.Filters, &diags)

	if api.Metrics != nil && len(*api.Metrics) > 0 {
		m.Metrics = make([]datatableMetricModel, len(*api.Metrics))
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
		m.Rows = make([]datatableRowModel, len(*api.Rows))
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
		m.SplitMetricsBy = make([]datatableSplitByModel, len(*api.SplitMetricsBy))
		for i, splitBy := range *api.SplitMetricsBy {
			splitBytes, err := json.Marshal(splitBy)
			sv, ok := marshalToNormalized(splitBytes, err, "split_metrics_by", &diags)
			if !ok {
				return diags
			}
			m.SplitMetricsBy[i].ConfigJSON = sv
		}
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

func (m *datatableESQLConfigModel) toAPI() (kbapi.DatatableESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.DatatableESQL{Type: kbapi.DatatableESQLTypeDataTable}
	api.TimeRange = lensPanelTimeRange()

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

	if m.Density != nil {
		density, densityDiags := m.Density.toAPI()
		diags.Append(densityDiags...)
		if diags.HasError() {
			return api, diags
		}
		api.Density = density
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

	if typeutils.IsKnown(m.SortByJSON) {
		var sortBy kbapi.DatatableESQL_SortBy
		if err := json.Unmarshal([]byte(m.SortByJSON.ValueString()), &sortBy); err != nil {
			diags.AddError("Failed to unmarshal sort_by", err.Error())
			return api, diags
		}
		api.SortBy = &sortBy
	}

	if typeutils.IsKnown(m.Paging) {
		paging := kbapi.DatatableESQLPaging(m.Paging.ValueInt64())
		api.Paging = &paging
	}

	return api, diags
}

func (m *datatableDensityModel) fromAPI(api kbapi.DatatableDensity) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Mode = typeutils.StringishPointerValue(api.Mode)

	if api.Height != nil {
		m.Height = &datatableDensityHeightModel{}
		heightDiags := m.Height.fromAPI(api.Height)
		diags.Append(heightDiags...)
	}

	return diags
}

func (m *datatableDensityModel) toAPI() (kbapi.DatatableDensity, diag.Diagnostics) {
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
			header, headerDiags := m.Height.Header.toAPI()
			diags.Append(headerDiags...)
			if diags.HasError() {
				return density, diags
			}
			height.Header = header
		}

		if m.Height.Value != nil {
			value, valueDiags := m.Height.Value.toAPI()
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

func (m *datatableDensityHeightModel) fromAPI(api *struct {
	Header *kbapi.DatatableDensity_Height_Header `json:"header,omitempty"`
	Value  *kbapi.DatatableDensity_Height_Value  `json:"value,omitempty"`
}) diag.Diagnostics {
	var diags diag.Diagnostics
	if api == nil {
		return diags
	}

	if api.Header != nil {
		m.Header = &datatableDensityHeightHeaderModel{}
		headerDiags := m.Header.fromAPI(api.Header)
		diags.Append(headerDiags...)
	}

	if api.Value != nil {
		m.Value = &datatableDensityHeightValueModel{}
		valueDiags := m.Value.fromAPI(api.Value)
		diags.Append(valueDiags...)
	}

	return diags
}

func (m *datatableDensityHeightHeaderModel) fromAPI(api *kbapi.DatatableDensity_Height_Header) diag.Diagnostics {
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

func (m *datatableDensityHeightHeaderModel) toAPI() (*kbapi.DatatableDensity_Height_Header, diag.Diagnostics) {
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

func (m *datatableDensityHeightValueModel) fromAPI(api *kbapi.DatatableDensity_Height_Value) diag.Diagnostics {
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

func (m *datatableDensityHeightValueModel) toAPI() (*kbapi.DatatableDensity_Height_Value, diag.Diagnostics) {
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
