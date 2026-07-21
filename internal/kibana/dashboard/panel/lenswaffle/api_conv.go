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

package lenswaffle

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func waffleChartJSONUsesESQLDataset(waffleChartJSON []byte) (bool, error) {
	var top struct {
		DataSource json.RawMessage `json:"data_source"`
	}
	if err := json.Unmarshal(waffleChartJSON, &top); err != nil {
		return false, err
	}
	var ds struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(top.DataSource, &ds); err != nil {
		return false, err
	}
	switch ds.Type {
	case string(kbapi.KibanaHTTPAPIsEsqlDataSourceTypeEsql), "table":
		return true, nil
	default:
		return false, nil
	}
}

func mergeWaffleConfigFromPlanSeed(cur, seed *models.WaffleConfigModel) {
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

func waffleConfigFromAPINoESQL(ctx context.Context, m *models.WaffleConfigModel, prior *models.WaffleConfigModel, api kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanel) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)

	m.Sampling = typeutils.Float32PointerToFloat64Value(api.Sampling)

	datasetBytes, err := api.DataSource.MarshalJSON()
	dv, ok := lenscommon.WrapNormalizedJSON(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = dv

	m.Query = &models.FilterSimpleModel{}
	lenscommon.FilterSimpleFromAPI(m.Query, api.Query)

	m.Filters = lenscommon.PopulateFiltersFromAPI(api.Filters, &diags)

	m.Legend = &models.WaffleLegendModel{}
	waffleLegendFromAPI(ctx, m.Legend, api.Legend)

	if api.Styling != nil && api.Styling.Values != nil && (api.Styling.Values.Mode != nil || api.Styling.Values.PercentDecimals != nil) {
		m.ValueDisplay = &models.PartitionValueDisplay{
			Mode: typeutils.StringishPointerValue(api.Styling.Values.Mode),
		}
		if api.Styling.Values.PercentDecimals != nil {
			m.ValueDisplay.PercentDecimals = types.Float64Value(float64(*api.Styling.Values.PercentDecimals))
		} else {
			m.ValueDisplay.PercentDecimals = types.Float64Null()
		}
	}

	if len(api.Metrics) > 0 {
		priorMetrics := m.Metrics
		m.Metrics = make([]models.WaffleDSLMetric, len(api.Metrics))
		for i, metric := range api.Metrics {
			b, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				continue
			}
			cfg := customtypes.NewJSONWithDefaultsValue(
				string(b),
				lenscommon.PopulatePieChartMetricDefaults,
			)
			if i < len(priorMetrics) {
				cfg = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, priorMetrics[i].Config, cfg, &diags)
			}
			m.Metrics[i].Config = cfg
		}
	}

	if api.GroupBy != nil && len(*api.GroupBy) > 0 {
		priorGroupBy := m.GroupBy
		m.GroupBy = make([]models.WaffleDSLGroupBy, len(*api.GroupBy))
		for i, gb := range *api.GroupBy {
			b, err := json.Marshal(gb)
			if err != nil {
				diags.AddError("Failed to marshal group_by", err.Error())
				continue
			}
			cfg := customtypes.NewJSONWithDefaultsValue(
				string(b),
				lenscommon.PopulateLensGroupByDefaults,
			)
			if i < len(priorGroupBy) {
				cfg = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, priorGroupBy[i].Config, cfg, &diags)
			}
			m.GroupBy[i].Config = cfg
		}
	}

	m.EsqlMetrics = nil
	m.EsqlGroupBy = nil

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}

	return diags
}

func waffleConfigFromAPIESQL(ctx context.Context, m *models.WaffleConfigModel, prior *models.WaffleConfigModel, api kbapi.KibanaHTTPAPIsWaffleESQLByValuePanel) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)

	m.Sampling = typeutils.Float32PointerToFloat64Value(api.Sampling)

	datasetBytes, err := json.Marshal(api.DataSource)
	dv, ok := lenscommon.WrapNormalizedJSON(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = dv

	m.Query = nil

	m.Filters = lenscommon.PopulateFiltersFromAPI(api.Filters, &diags)

	m.Legend = &models.WaffleLegendModel{}
	waffleLegendFromAPI(ctx, m.Legend, api.Legend)

	if api.Styling != nil && api.Styling.Values != nil && (api.Styling.Values.Mode != nil || api.Styling.Values.PercentDecimals != nil) {
		m.ValueDisplay = &models.PartitionValueDisplay{
			Mode: typeutils.StringishPointerValue(api.Styling.Values.Mode),
		}
		if api.Styling.Values.PercentDecimals != nil {
			m.ValueDisplay.PercentDecimals = types.Float64Value(float64(*api.Styling.Values.PercentDecimals))
		} else {
			m.ValueDisplay.PercentDecimals = types.Float64Null()
		}
	}

	if len(api.Metrics) > 0 {
		m.EsqlMetrics = make([]models.WaffleEsqlMetric, len(api.Metrics))
		for i, met := range api.Metrics {
			colorType := types.StringNull()
			colorValue := types.StringNull()
			if met.Color != nil {
				if staticColor, colorErr := met.Color.AsKibanaHTTPAPIsStaticColor(); colorErr == nil {
					colorType = types.StringValue(string(staticColor.Type))
					colorValue = types.StringValue(staticColor.Color)
				}
			}
			em := models.WaffleEsqlMetric{
				Column: types.StringValue(met.Column),
				FormatJSON: func() jsontypes.Normalized {
					b, err := json.Marshal(met.Format)
					if err != nil {
						return jsontypes.NewNormalizedNull()
					}
					// Kibana may omit format on saved-object round-trip, leaving Format as an empty union.
					if string(b) == lenscommon.JSONNullString || len(b) == 0 {
						b = []byte(lenscommon.DefaultLensNumberFormatJSON)
					}
					return jsontypes.NewNormalizedValue(lenscommon.NormalizeKibanaLensNumberFormatJSONString(string(b)))
				}(),
				Color: &models.LensStaticColorModel{
					Type:  colorType,
					Color: colorValue,
				},
			}
			em.Label = typeutils.StringishPointerValue(met.Label)
			m.EsqlMetrics[i] = em
		}
	}

	if api.GroupBy != nil && len(*api.GroupBy) > 0 {
		src := make([]lenscommon.EsqlGroupByAPIFields, len(*api.GroupBy))
		for i, gb := range *api.GroupBy {
			src[i] = lenscommon.EsqlGroupByAPIFields{CollapseBy: gb.CollapseBy, Color: gb.Color, Column: gb.Column, Format: gb.Format, Label: gb.Label}
		}
		m.EsqlGroupBy = lenscommon.PopulatePartitionEsqlGroupByFromAPI(src, &diags)
	}

	m.Metrics = nil
	m.GroupBy = nil

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}

	return diags
}

func waffleLegendFromAPI(ctx context.Context, m *models.WaffleLegendModel, api *kbapi.KibanaHTTPAPIsWaffleLegend) {
	_ = ctx
	if api == nil {
		m.Size = types.StringNull()
		m.TruncateAfterLines = types.Int64Null()
		m.Values = types.ListNull(types.StringType)
		m.Visible = types.StringNull()
		return
	}
	if api.Size != nil {
		m.Size = types.StringValue(string(*api.Size))
	} else {
		m.Size = types.StringNull()
	}
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
	if api.Visibility != nil {
		m.Visible = types.StringValue(string(*api.Visibility))
	} else {
		m.Visible = types.StringNull()
	}
}

func waffleLegendToAPI(m *models.WaffleLegendModel) (*kbapi.KibanaHTTPAPIsWaffleLegend, diag.Diagnostics) {
	var diags diag.Diagnostics
	leg := &kbapi.KibanaHTTPAPIsWaffleLegend{}
	if m == nil {
		diags.AddError("Missing legend", "waffle_config.legend must be provided")
		return nil, diags
	}
	if typeutils.IsKnown(m.Size) {
		size := kbapi.KibanaHTTPAPIsLegendSize(m.Size.ValueString())
		leg.Size = &size
	} else {
		diags.AddError("Missing legend size", "waffle_config.legend.size must be provided")
	}
	if typeutils.IsKnown(m.TruncateAfterLines) {
		v := float32(m.TruncateAfterLines.ValueInt64())
		leg.TruncateAfterLines = &v
	}
	if typeutils.IsKnown(m.Visible) {
		v := kbapi.KibanaHTTPAPIsWaffleLegendVisibility(m.Visible.ValueString())
		leg.Visibility = &v
	}
	if !m.Values.IsNull() && !m.Values.IsUnknown() {
		elems := m.Values.Elements()
		vals := make([]kbapi.KibanaHTTPAPIsWaffleLegendValues, 0, len(elems))
		for _, e := range elems {
			sv, ok := e.(types.String)
			if ok && typeutils.IsKnown(sv) {
				vals = append(vals, kbapi.KibanaHTTPAPIsWaffleLegendValues(sv.ValueString()))
			}
		}
		if len(vals) > 0 {
			leg.Values = &vals
		}
	}
	return leg, diags
}

func waffleConfigToAPI(m *models.WaffleConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var attrs lenscommon.VisByValueConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	diags.Append(WaffleConfigModeValidateDiags(lenscommon.ConfigUsesESQL(m.Query), WaffleModeListStateFromSlice(len(m.Metrics)),
		WaffleModeListStateFromSlice(len(m.GroupBy)),
		WaffleModeListStateFromSlice(len(m.EsqlMetrics)),
		WaffleModeListStateFromSlice(len(m.EsqlGroupBy)),
	)...)
	if diags.HasError() {
		return attrs, diags
	}

	if lenscommon.ConfigUsesESQL(m.Query) {
		esql, d := waffleConfigToAPIESQL(m)
		diags.Append(d...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromKibanaHTTPAPIsWaffleESQLByValuePanel(esql); err != nil {
			diags.AddError("Failed to build waffle ES|QL chart", err.Error())
		}
		return attrs, diags
	}

	noESQL, d := waffleConfigToAPINoESQL(m)
	diags.Append(d...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromKibanaHTTPAPIsWaffleNoESQLByValuePanel(noESQL); err != nil {
		diags.AddError("Failed to build waffle chart", err.Error())
	}
	return attrs, diags
}

func waffleConfigToAPINoESQL(m *models.WaffleConfigModel) (kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanel{
		Type: kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanelTypeWaffle,
	}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing dataset", "waffle_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal data_source_json", err.Error())
		return api, diags
	}

	if m.Query == nil {
		diags.AddError("Missing query", "waffle_config.query must be set for non-ES|QL waffles (or omit `query` entirely for ES|QL mode)")
		return api, diags
	}
	api.Query = lenscommon.FilterSimpleToAPI(m.Query)

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if m.Legend == nil {
		diags.AddError("Missing legend", "waffle_config.legend must be provided")
		return api, diags
	}
	leg, ld := waffleLegendToAPI(m.Legend)
	diags.Append(ld...)
	api.Legend = leg

	if m.ValueDisplay != nil && typeutils.IsKnown(m.ValueDisplay.Mode) {
		api.Styling = &kbapi.KibanaHTTPAPIsWaffleStyling{Values: lenscommon.PartitionValueDisplayToAPI(m.ValueDisplay)}
	} else {
		// Required by the Dashboard API; omitting mode yields HTTP 400.
		mode := kbapi.KibanaHTTPAPIsValueDisplayModePercentage
		api.Styling = &kbapi.KibanaHTTPAPIsWaffleStyling{
			Values: &kbapi.KibanaHTTPAPIsValueDisplay{Mode: &mode},
		}
	}

	metrics := make([]kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanel_Metrics_Item, len(m.Metrics))
	for i, met := range m.Metrics {
		if err := json.Unmarshal([]byte(met.Config.ValueString()), &metrics[i]); err != nil {
			diags.AddError("Failed to unmarshal metric config", err.Error())
		}
	}
	api.Metrics = metrics

	if len(m.GroupBy) > 0 {
		gb := make([]kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanel_GroupBy_Item, len(m.GroupBy))
		for i, g := range m.GroupBy {
			if err := json.Unmarshal([]byte(g.Config.ValueString()), &gb[i]); err != nil {
				diags.AddError("Failed to unmarshal group_by config", err.Error())
			}
		}
		api.GroupBy = &gb
	}

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsWaffleNoESQLByValuePanel_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	return api, diags
}

func waffleConfigToAPIESQL(m *models.WaffleConfigModel) (kbapi.KibanaHTTPAPIsWaffleESQLByValuePanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KibanaHTTPAPIsWaffleESQLByValuePanel{
		Type: kbapi.KibanaHTTPAPIsWaffleESQLByValuePanelTypeWaffle,
	}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing dataset", "waffle_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal data_source_json", err.Error())
		return api, diags
	}

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if m.Legend == nil {
		diags.AddError("Missing legend", "waffle_config.legend must be provided")
		return api, diags
	}
	leg, ld := waffleLegendToAPI(m.Legend)
	diags.Append(ld...)
	api.Legend = leg

	if m.ValueDisplay != nil && typeutils.IsKnown(m.ValueDisplay.Mode) {
		api.Styling = &kbapi.KibanaHTTPAPIsWaffleStyling{Values: lenscommon.PartitionValueDisplayToAPI(m.ValueDisplay)}
	} else {
		mode := kbapi.KibanaHTTPAPIsValueDisplayModePercentage
		api.Styling = &kbapi.KibanaHTTPAPIsWaffleStyling{
			Values: &kbapi.KibanaHTTPAPIsValueDisplay{Mode: &mode},
		}
	}

	metrics := make([]struct {
		Color  *kbapi.KibanaHTTPAPIsWaffleESQLByValuePanel_Metrics_Color `json:"color,omitempty"`
		Column string                                                    `json:"column"`
		Format *kbapi.KibanaHTTPAPIsFormatType                           `json:"format,omitempty"`
		Label  *string                                                   `json:"label,omitempty"`
	}, len(m.EsqlMetrics))
	for i, em := range m.EsqlMetrics {
		var format kbapi.KibanaHTTPAPIsFormatType
		if err := json.Unmarshal([]byte(em.FormatJSON.ValueString()), &format); err != nil {
			diags.AddError("Failed to unmarshal format_json", err.Error())
		} else {
			metrics[i].Format = &format
		}
		metrics[i].Column = em.Column.ValueString()
		if em.Color == nil {
			diags.AddError("Missing color", "waffle_config.esql_metrics color is required")
			continue
		}
		staticColor := kbapi.KibanaHTTPAPIsStaticColor{
			Type:  kbapi.KibanaHTTPAPIsStaticColorType(em.Color.Type.ValueString()),
			Color: em.Color.Color.ValueString(),
		}
		var color kbapi.KibanaHTTPAPIsWaffleESQLByValuePanel_Metrics_Color
		if err := color.FromKibanaHTTPAPIsStaticColor(staticColor); err != nil {
			diags.AddError("Failed to marshal metric color", err.Error())
			continue
		}
		metrics[i].Color = &color
		if typeutils.IsKnown(em.Label) {
			s := em.Label.ValueString()
			metrics[i].Label = &s
		}
	}
	api.Metrics = metrics

	if len(m.EsqlGroupBy) > 0 {
		entries := lenscommon.BuildPartitionEsqlGroupByForAPI(m.EsqlGroupBy, &diags)
		if diags.HasError() {
			return api, diags
		}
		groupBy := lenscommon.BuildEsqlGroupBySliceForAPI[struct {
			CollapseBy *kbapi.KibanaHTTPAPIsCollapseBy   `json:"collapse_by,omitempty"`
			Color      *kbapi.KibanaHTTPAPIsColorMapping `json:"color,omitempty"`
			Column     string                            `json:"column"`
			Format     *kbapi.KibanaHTTPAPIsFormatType   `json:"format,omitempty"`
			Label      *string                           `json:"label,omitempty"`
		}](entries, &diags)
		if diags.HasError() {
			return api, diags
		}
		api.GroupBy = &groupBy
	}

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsWaffleESQLByValuePanel_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	return api, diags
}
