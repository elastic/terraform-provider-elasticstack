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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newWafflePanelConfigConverter() wafflePanelConfigConverter {
	return wafflePanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.WaffleNoESQLTypeWaffle),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.WaffleConfig != nil },
		},
	}
}

type wafflePanelConfigConverter struct {
	lensVisualizationBase
}

func (c wafflePanelConfigConverter) populateFromAttributes(ctx context.Context, pm *panelModel, attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes) diag.Diagnostics {
	seed := pm.WaffleConfig
	waffleChart, err := attrs.AsWaffleChart()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.WaffleConfig = &waffleConfigModel{}
	raw, err := json.Marshal(waffleChart)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	esql, err := waffleChartJSONUsesESQLDataset(raw)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var diags diag.Diagnostics
	if esql {
		wESQL, err := waffleChart.AsWaffleESQL()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		diags = pm.WaffleConfig.fromAPIESQL(ctx, wESQL)
	} else {
		wNoESQL, err := waffleChart.AsWaffleNoESQL()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		diags = pm.WaffleConfig.fromAPINoESQL(ctx, wNoESQL)
	}
	mergeWaffleConfigFromPlanSeed(pm.WaffleConfig, seed)
	return diags
}

// waffleChartJSONUsesESQLDataset reports whether waffle chart JSON is the ES|QL variant by reading
// dataset.type. kbapi.WaffleChart AsWaffleNoESQL / AsWaffleESQL both json.Unmarshal the same blob
// into different structs and do not indicate the variant via their error return value.
func waffleChartJSONUsesESQLDataset(waffleChartJSON []byte) (bool, error) {
	var top struct {
		Dataset json.RawMessage `json:"dataset"`
	}
	if err := json.Unmarshal(waffleChartJSON, &top); err != nil {
		return false, err
	}
	var ds struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(top.Dataset, &ds); err != nil {
		return false, err
	}
	switch ds.Type {
	case string(kbapi.EsqlDatasetTypeEsql), string(kbapi.Table):
		return true, nil
	default:
		return false, nil
	}
}

func (c wafflePanelConfigConverter) buildAttributes(pm panelModel) (kbapi.KbnDashboardPanelLens_Config_0_Attributes, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *pm.WaffleConfig

	waffleChart, wDiags := configModel.toAPI()
	diags.Append(wDiags...)
	if diags.HasError() {
		return kbapi.KbnDashboardPanelLens_Config_0_Attributes{}, diags
	}

	var attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	if err := attrs.FromWaffleChart(waffleChart); err != nil {
		diags.AddError("Failed to create waffle chart attributes", err.Error())
		return kbapi.KbnDashboardPanelLens_Config_0_Attributes{}, diags
	}

	return attrs, diags
}

// mergeWaffleConfigFromPlanSeed restores optional fields Kibana omits or defaults on read
// so Refresh after Apply matches the Terraform plan (see mapPanelFromAPI seeding comment).
func mergeWaffleConfigFromPlanSeed(cur, seed *waffleConfigModel) {
	if cur == nil || seed == nil {
		return
	}
	if typeutils.IsKnown(seed.IgnoreGlobalFilters) {
		if !typeutils.IsKnown(cur.IgnoreGlobalFilters) || cur.IgnoreGlobalFilters.IsNull() ||
			cur.IgnoreGlobalFilters.ValueBool() != seed.IgnoreGlobalFilters.ValueBool() {
			cur.IgnoreGlobalFilters = seed.IgnoreGlobalFilters
		}
	}
	if typeutils.IsKnown(seed.Sampling) {
		if !typeutils.IsKnown(cur.Sampling) || cur.Sampling.IsNull() ||
			cur.Sampling.ValueFloat64() != seed.Sampling.ValueFloat64() {
			cur.Sampling = seed.Sampling
		}
	}
	if cur.Legend != nil && seed.Legend != nil {
		if cur.Legend.Values.IsNull() && !seed.Legend.Values.IsNull() && !seed.Legend.Values.IsUnknown() {
			cur.Legend.Values = seed.Legend.Values
		}
		if seed.Legend.Visible.IsNull() && typeutils.IsKnown(cur.Legend.Visible) {
			cur.Legend.Visible = types.StringNull()
		}
	}
	if seed.ValueDisplay == nil && cur.ValueDisplay != nil {
		cur.ValueDisplay = nil
	}
}

type waffleConfigModel struct {
	Title               types.String         `tfsdk:"title"`
	Description         types.String         `tfsdk:"description"`
	DatasetJSON         jsontypes.Normalized `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool           `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64        `tfsdk:"sampling"`
	Query               *filterSimpleModel   `tfsdk:"query"`
	Filters             []searchFilterModel  `tfsdk:"filters"`
	Legend              *waffleLegendModel   `tfsdk:"legend"`
	ValueDisplay        *waffleValueDisplay  `tfsdk:"value_display"`
	Metrics             []waffleDSLMetric    `tfsdk:"metrics"`
	GroupBy             []waffleDSLGroupBy   `tfsdk:"group_by"`
	EsqlMetrics         []waffleEsqlMetric   `tfsdk:"esql_metrics"`
	EsqlGroupBy         []waffleEsqlGroupBy  `tfsdk:"esql_group_by"`
}

type waffleDSLMetric struct {
	Config customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config"`
}

type waffleDSLGroupBy struct {
	Config customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config"`
}

type waffleLegendModel struct {
	Size               types.String `tfsdk:"size"`
	TruncateAfterLines types.Int64  `tfsdk:"truncate_after_lines"`
	Values             types.List   `tfsdk:"values"`
	Visible            types.String `tfsdk:"visible"`
}

type waffleValueDisplay struct {
	Mode            types.String  `tfsdk:"mode"`
	PercentDecimals types.Float64 `tfsdk:"percent_decimals"`
}

type waffleEsqlMetric struct {
	Column     types.String         `tfsdk:"column"`
	Operation  types.String         `tfsdk:"operation"`
	Label      types.String         `tfsdk:"label"`
	FormatJSON jsontypes.Normalized `tfsdk:"format_json"`
	Color      *waffleStaticColor   `tfsdk:"color"`
}

type waffleStaticColor struct {
	Type  types.String `tfsdk:"type"`
	Color types.String `tfsdk:"color"`
}

type waffleEsqlGroupBy struct {
	Column     types.String         `tfsdk:"column"`
	Operation  types.String         `tfsdk:"operation"`
	CollapseBy types.String         `tfsdk:"collapse_by"`
	ColorJSON  jsontypes.Normalized `tfsdk:"color_json"`
}

func (m *waffleConfigModel) usesESQL() bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Query.IsNull() && m.Query.Language.IsNull()
}

func (m *waffleConfigModel) fromAPINoESQL(ctx context.Context, api kbapi.WaffleNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)

	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset_json", err.Error())
		return diags
	}
	m.DatasetJSON = jsontypes.NewNormalizedValue(string(datasetBytes))

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, 0, len(*api.Filters))
		for _, f := range *api.Filters {
			fm := searchFilterModel{}
			fd := fm.fromAPI(f)
			diags.Append(fd...)
			if !fd.HasError() {
				m.Filters = append(m.Filters, fm)
			}
		}
	}

	m.Legend = &waffleLegendModel{}
	m.Legend.fromAPI(ctx, api.Legend)

	if api.ValueDisplay != nil {
		m.ValueDisplay = &waffleValueDisplay{
			Mode: types.StringValue(string(api.ValueDisplay.Mode)),
		}
		if api.ValueDisplay.PercentDecimals != nil {
			m.ValueDisplay.PercentDecimals = types.Float64Value(float64(*api.ValueDisplay.PercentDecimals))
		} else {
			m.ValueDisplay.PercentDecimals = types.Float64Null()
		}
	}

	if len(api.Metrics) > 0 {
		m.Metrics = make([]waffleDSLMetric, len(api.Metrics))
		for i, metric := range api.Metrics {
			b, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				continue
			}
			m.Metrics[i].Config = customtypes.NewJSONWithDefaultsValue[map[string]any](
				string(b),
				populatePieChartMetricDefaults,
			)
		}
	}

	if api.GroupBy != nil && len(*api.GroupBy) > 0 {
		m.GroupBy = make([]waffleDSLGroupBy, len(*api.GroupBy))
		for i, gb := range *api.GroupBy {
			b, err := json.Marshal(gb)
			if err != nil {
				diags.AddError("Failed to marshal group_by", err.Error())
				continue
			}
			m.GroupBy[i].Config = customtypes.NewJSONWithDefaultsValue[map[string]any](
				string(b),
				populateLensGroupByDefaults,
			)
		}
	}

	m.EsqlMetrics = nil
	m.EsqlGroupBy = nil
	return diags
}

func (m *waffleConfigModel) fromAPIESQL(ctx context.Context, api kbapi.WaffleESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)

	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset_json", err.Error())
		return diags
	}
	m.DatasetJSON = jsontypes.NewNormalizedValue(string(datasetBytes))

	m.Query = nil

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, 0, len(*api.Filters))
		for _, f := range *api.Filters {
			fm := searchFilterModel{}
			fd := fm.fromAPI(f)
			diags.Append(fd...)
			if !fd.HasError() {
				m.Filters = append(m.Filters, fm)
			}
		}
	}

	m.Legend = &waffleLegendModel{}
	m.Legend.fromAPI(ctx, api.Legend)

	if api.ValueDisplay != nil {
		m.ValueDisplay = &waffleValueDisplay{
			Mode: types.StringValue(string(api.ValueDisplay.Mode)),
		}
		if api.ValueDisplay.PercentDecimals != nil {
			m.ValueDisplay.PercentDecimals = types.Float64Value(float64(*api.ValueDisplay.PercentDecimals))
		} else {
			m.ValueDisplay.PercentDecimals = types.Float64Null()
		}
	}

	if len(api.Metrics) > 0 {
		m.EsqlMetrics = make([]waffleEsqlMetric, len(api.Metrics))
		for i, met := range api.Metrics {
			em := waffleEsqlMetric{
				Column:    types.StringValue(met.Column),
				Operation: types.StringValue(string(met.Operation)),
				FormatJSON: func() jsontypes.Normalized {
					b, err := json.Marshal(met.Format)
					if err != nil {
						return jsontypes.NewNormalizedNull()
					}
					// Kibana may omit format on saved-object round-trip, leaving Format as an empty union.
					if string(b) == "null" || len(b) == 0 {
						b = []byte(`{"type":"number"}`)
					}
					return jsontypes.NewNormalizedValue(string(b))
				}(),
				Color: &waffleStaticColor{
					Type:  types.StringValue(string(met.Color.Type)),
					Color: types.StringValue(met.Color.Color),
				},
			}
			if met.Label != nil {
				em.Label = types.StringValue(*met.Label)
			} else {
				em.Label = types.StringNull()
			}
			m.EsqlMetrics[i] = em
		}
	}

	if api.GroupBy != nil && len(*api.GroupBy) > 0 {
		m.EsqlGroupBy = make([]waffleEsqlGroupBy, len(*api.GroupBy))
		for i, gb := range *api.GroupBy {
			colorBytes, err := json.Marshal(gb.Color)
			if err != nil {
				diags.AddError("Failed to marshal esql group_by color", err.Error())
				continue
			}
			m.EsqlGroupBy[i] = waffleEsqlGroupBy{
				Column:     types.StringValue(gb.Column),
				Operation:  types.StringValue(string(gb.Operation)),
				CollapseBy: types.StringValue(string(gb.CollapseBy)),
				ColorJSON:  jsontypes.NewNormalizedValue(string(colorBytes)),
			}
		}
	}

	m.Metrics = nil
	m.GroupBy = nil
	return diags
}

func (m *waffleLegendModel) fromAPI(ctx context.Context, api kbapi.WaffleLegend) {
	_ = ctx
	m.Size = types.StringValue(string(api.Size))
	if api.TruncateAfterLines != nil {
		m.TruncateAfterLines = types.Int64Value(int64(*api.TruncateAfterLines))
	} else {
		m.TruncateAfterLines = types.Int64Null()
	}
	if api.Values != nil && len(*api.Values) > 0 {
		elems := make([]attr.Value, len(*api.Values))
		for i, v := range *api.Values {
			elems[i] = types.StringValue(string(v))
		}
		lv, diags := types.ListValue(types.StringType, elems)
		if !diags.HasError() {
			m.Values = lv
		} else {
			m.Values = types.ListNull(types.StringType)
		}
	} else {
		m.Values = types.ListNull(types.StringType)
	}
	if api.Visible != nil {
		m.Visible = types.StringValue(string(*api.Visible))
	} else {
		m.Visible = types.StringNull()
	}
}

func (m *waffleLegendModel) toAPI() (kbapi.WaffleLegend, diag.Diagnostics) {
	var diags diag.Diagnostics
	var leg kbapi.WaffleLegend
	if m == nil {
		diags.AddError("Missing legend", "waffle_config.legend must be provided")
		return leg, diags
	}
	if typeutils.IsKnown(m.Size) {
		leg.Size = kbapi.LegendSize(m.Size.ValueString())
	} else {
		diags.AddError("Missing legend size", "waffle_config.legend.size must be provided")
	}
	if typeutils.IsKnown(m.TruncateAfterLines) {
		v := float32(m.TruncateAfterLines.ValueInt64())
		leg.TruncateAfterLines = &v
	}
	if typeutils.IsKnown(m.Visible) {
		v := kbapi.WaffleLegendVisible(m.Visible.ValueString())
		leg.Visible = &v
	}
	if !m.Values.IsNull() && !m.Values.IsUnknown() {
		elems := m.Values.Elements()
		vals := make([]kbapi.WaffleLegendValues, 0, len(elems))
		for _, e := range elems {
			sv, ok := e.(types.String)
			if ok && typeutils.IsKnown(sv) {
				vals = append(vals, kbapi.WaffleLegendValues(sv.ValueString()))
			}
		}
		if len(vals) > 0 {
			leg.Values = &vals
		}
	}
	return leg, diags
}

func (m *waffleConfigModel) toAPI() (kbapi.WaffleChart, diag.Diagnostics) {
	var diags diag.Diagnostics
	var chart kbapi.WaffleChart

	if m == nil {
		return chart, diags
	}

	diags.Append(waffleConfigModeValidateDiags(m.usesESQL(),
		waffleModeListStateFromSlice(len(m.Metrics)),
		waffleModeListStateFromSlice(len(m.GroupBy)),
		waffleModeListStateFromSlice(len(m.EsqlMetrics)),
		waffleModeListStateFromSlice(len(m.EsqlGroupBy)),
		nil,
	)...)
	if diags.HasError() {
		return chart, diags
	}

	if m.usesESQL() {
		esql, d := m.toAPIESQL()
		diags.Append(d...)
		if diags.HasError() {
			return chart, diags
		}
		if err := chart.FromWaffleESQL(esql); err != nil {
			diags.AddError("Failed to build waffle ES|QL chart", err.Error())
		}
		return chart, diags
	}

	noESQL, d := m.toAPINoESQL()
	diags.Append(d...)
	if diags.HasError() {
		return chart, diags
	}
	if err := chart.FromWaffleNoESQL(noESQL); err != nil {
		diags.AddError("Failed to build waffle chart", err.Error())
	}
	return chart, diags
}

func (m *waffleConfigModel) toAPINoESQL() (kbapi.WaffleNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.WaffleNoESQL{Type: kbapi.WaffleNoESQLTypeWaffle}

	if typeutils.IsKnown(m.Title) {
		api.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = new(m.Description.ValueString())
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = new(m.IgnoreGlobalFilters.ValueBool())
	}
	if typeutils.IsKnown(m.Sampling) {
		api.Sampling = new(float32(m.Sampling.ValueFloat64()))
	}

	if m.DatasetJSON.IsNull() {
		diags.AddError("Missing dataset", "waffle_config.dataset_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &api.Dataset); err != nil {
		diags.AddError("Failed to unmarshal dataset_json", err.Error())
		return api, diags
	}

	if m.Query == nil {
		diags.AddError("Missing query", "waffle_config.query must be set for non-ES|QL waffles (or omit `query` entirely for ES|QL mode)")
		return api, diags
	}
	api.Query = m.Query.toAPI()

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilter, 0, len(m.Filters))
		for _, f := range m.Filters {
			sf, fd := f.toAPI()
			diags.Append(fd...)
			filters = append(filters, sf)
		}
		api.Filters = &filters
	}

	if m.Legend == nil {
		diags.AddError("Missing legend", "waffle_config.legend must be provided")
		return api, diags
	}
	leg, ld := m.Legend.toAPI()
	diags.Append(ld...)
	api.Legend = leg

	if m.ValueDisplay != nil && typeutils.IsKnown(m.ValueDisplay.Mode) {
		vd := struct {
			Mode            kbapi.WaffleNoESQLValueDisplayMode `json:"mode"`
			PercentDecimals *float32                           `json:"percent_decimals,omitempty"`
		}{
			Mode: kbapi.WaffleNoESQLValueDisplayMode(m.ValueDisplay.Mode.ValueString()),
		}
		if typeutils.IsKnown(m.ValueDisplay.PercentDecimals) {
			p := float32(m.ValueDisplay.PercentDecimals.ValueFloat64())
			vd.PercentDecimals = &p
		}
		api.ValueDisplay = &vd
	}

	metrics := make([]kbapi.WaffleNoESQL_Metrics_Item, len(m.Metrics))
	for i, met := range m.Metrics {
		if err := json.Unmarshal([]byte(met.Config.ValueString()), &metrics[i]); err != nil {
			diags.AddError("Failed to unmarshal metric config", err.Error())
		}
	}
	api.Metrics = metrics

	if len(m.GroupBy) > 0 {
		gb := make([]kbapi.WaffleNoESQL_GroupBy_Item, len(m.GroupBy))
		for i, g := range m.GroupBy {
			if err := json.Unmarshal([]byte(g.Config.ValueString()), &gb[i]); err != nil {
				diags.AddError("Failed to unmarshal group_by config", err.Error())
			}
		}
		api.GroupBy = &gb
	}

	return api, diags
}

func (m *waffleConfigModel) toAPIESQL() (kbapi.WaffleESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.WaffleESQL{Type: kbapi.WaffleESQLTypeWaffle}

	if typeutils.IsKnown(m.Title) {
		api.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = new(m.Description.ValueString())
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = new(m.IgnoreGlobalFilters.ValueBool())
	}
	if typeutils.IsKnown(m.Sampling) {
		api.Sampling = new(float32(m.Sampling.ValueFloat64()))
	}

	if m.DatasetJSON.IsNull() {
		diags.AddError("Missing dataset", "waffle_config.dataset_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &api.Dataset); err != nil {
		diags.AddError("Failed to unmarshal dataset_json", err.Error())
		return api, diags
	}

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilter, 0, len(m.Filters))
		for _, f := range m.Filters {
			sf, fd := f.toAPI()
			diags.Append(fd...)
			filters = append(filters, sf)
		}
		api.Filters = &filters
	}

	if m.Legend == nil {
		diags.AddError("Missing legend", "waffle_config.legend must be provided")
		return api, diags
	}
	leg, ld := m.Legend.toAPI()
	diags.Append(ld...)
	api.Legend = leg

	if m.ValueDisplay != nil && typeutils.IsKnown(m.ValueDisplay.Mode) {
		vd := struct {
			Mode            kbapi.WaffleESQLValueDisplayMode `json:"mode"`
			PercentDecimals *float32                         `json:"percent_decimals,omitempty"`
		}{
			Mode: kbapi.WaffleESQLValueDisplayMode(m.ValueDisplay.Mode.ValueString()),
		}
		if typeutils.IsKnown(m.ValueDisplay.PercentDecimals) {
			p := float32(m.ValueDisplay.PercentDecimals.ValueFloat64())
			vd.PercentDecimals = &p
		}
		api.ValueDisplay = &vd
	}

	metrics := make([]struct {
		Color     kbapi.StaticColor                `json:"color"`
		Column    string                           `json:"column"`
		Format    kbapi.FormatType                 `json:"format"`
		Label     *string                          `json:"label,omitempty"`
		Operation kbapi.WaffleESQLMetricsOperation `json:"operation"`
	}, len(m.EsqlMetrics))
	for i, em := range m.EsqlMetrics {
		if err := json.Unmarshal([]byte(em.FormatJSON.ValueString()), &metrics[i].Format); err != nil {
			diags.AddError("Failed to unmarshal format_json", err.Error())
		}
		metrics[i].Column = em.Column.ValueString()
		metrics[i].Operation = kbapi.WaffleESQLMetricsOperation(em.Operation.ValueString())
		if em.Color == nil {
			diags.AddError("Missing color", "waffle_config.esql_metrics color is required")
			continue
		}
		metrics[i].Color = kbapi.StaticColor{
			Type:  kbapi.StaticColorType(em.Color.Type.ValueString()),
			Color: em.Color.Color.ValueString(),
		}
		if typeutils.IsKnown(em.Label) {
			s := em.Label.ValueString()
			metrics[i].Label = &s
		}
	}
	api.Metrics = metrics

	if len(m.EsqlGroupBy) > 0 {
		gb := make([]struct {
			CollapseBy kbapi.CollapseBy                 `json:"collapse_by"`
			Color      kbapi.ColorMapping               `json:"color"`
			Column     string                           `json:"column"`
			Operation  kbapi.WaffleESQLGroupByOperation `json:"operation"`
		}, len(m.EsqlGroupBy))
		for i, eg := range m.EsqlGroupBy {
			if err := json.Unmarshal([]byte(eg.ColorJSON.ValueString()), &gb[i].Color); err != nil {
				diags.AddError("Failed to unmarshal color_json", err.Error())
			}
			gb[i].Column = eg.Column.ValueString()
			gb[i].Operation = kbapi.WaffleESQLGroupByOperation(eg.Operation.ValueString())
			gb[i].CollapseBy = kbapi.CollapseBy(eg.CollapseBy.ValueString())
		}
		api.GroupBy = &gb
	}

	return api, diags
}
