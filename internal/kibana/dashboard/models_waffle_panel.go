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
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
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
			hasTFChartBlock: func(blocks *models.LensByValueChartBlocks) bool {
				return blocks != nil && blocks.WaffleConfig != nil
			},
		},
	}
}

type wafflePanelConfigConverter struct {
	lensVisualizationBase
}

func (c wafflePanelConfigConverter) populateFromAttributes(
	ctx context.Context,
	dashboard *models.DashboardModel,
	tfPanel *models.PanelModel,
	blocks *models.LensByValueChartBlocks,
	attrs kbapi.KbnDashboardPanelTypeVisConfig0,
) diag.Diagnostics {
	seed := blocks.WaffleConfig

	var prior *models.WaffleConfigModel
	if seed != nil {
		cpy := *seed
		prior = &cpy
	} else if b := lensByValueChartBlocksFromPanel(tfPanel); b != nil && b.WaffleConfig != nil {
		cpy := *b.WaffleConfig
		prior = &cpy
	}

	raw, err := attrs.MarshalJSON()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	esql, err := waffleChartJSONUsesESQLDataset(raw)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	blocks.WaffleConfig = &models.WaffleConfigModel{}
	var diags diag.Diagnostics
	if esql {
		wESQL, err := attrs.AsWaffleESQL()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		diags = waffleConfigFromAPIESQL(ctx, blocks.WaffleConfig, dashboard, prior, wESQL)
	} else {
		wNoESQL, err := attrs.AsWaffleNoESQL()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		diags = waffleConfigFromAPINoESQL(ctx, blocks.WaffleConfig, dashboard, prior, wNoESQL)
	}
	mergeWaffleConfigFromPlanSeed(blocks.WaffleConfig, seed)
	return diags
}

// waffleChartJSONUsesESQLDataset reports whether lens waffle JSON is the ES|QL variant by reading
// data_source.type. AsWaffleNoESQL / AsWaffleESQL both decode the same blob without a reliable
// error distinction, so we inspect the raw panel JSON.
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
	case string(kbapi.EsqlDataSourceTypeEsql), "table":
		return true, nil
	default:
		return false, nil
	}
}

func (c wafflePanelConfigConverter) buildAttributes(blocks *models.LensByValueChartBlocks, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	configModel := *blocks.WaffleConfig
	return waffleConfigToAPI(&configModel, dashboard)
}

// normalizeKibanaLensNumberFormatJSONString trims Lens number-format defaults Kibana adds on read
// (decimals: 2, compact: false) so state matches compact Terraform jsonencode like {"type":"number"}.
func normalizeKibanaLensNumberFormatJSONString(jsonStr string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return jsonStr
	}
	typ, _ := m["type"].(string)
	if typ != "number" {
		return jsonStr
	}
	if jsonNumericEqualsLoose(m["decimals"], 2) {
		delete(m, "decimals")
	}
	if b, ok := m["compact"].(bool); ok && !b {
		delete(m, "compact")
	}
	sorted := sortJSONMapKeysRecursive(m)
	out, err := json.Marshal(sorted)
	if err != nil {
		return jsonStr
	}
	return string(out)
}

func jsonNumericEqualsLoose(v any, want float64) bool {
	switch x := v.(type) {
	case float64:
		return x == want
	case float32:
		return float64(x) == want
	case int:
		return float64(x) == want
	case int64:
		return float64(x) == want
	case json.Number:
		f, err := x.Float64()
		return err == nil && f == want
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return err == nil && f == want
	default:
		return false
	}
}

// mergeWaffleConfigFromPlanSeed restores optional fields Kibana omits or defaults on read
// so Refresh after Apply matches the Terraform plan (see mapPanelFromAPI seeding comment).
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

func waffleConfigUsesESQL(m *models.WaffleConfigModel) bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Expression.IsNull() && m.Query.Language.IsNull()
}

func waffleConfigFromAPINoESQL(ctx context.Context, m *models.WaffleConfigModel, dashboard *models.DashboardModel, prior *models.WaffleConfigModel, api kbapi.WaffleNoESQL) diag.Diagnostics {
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

	datasetBytes, err := api.DataSource.MarshalJSON()
	dv, ok := marshalToNormalized(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = dv

	m.Query = &models.FilterSimpleModel{}
	filterSimpleFromAPI(m.Query, api.Query)

	m.Filters = populateFiltersFromAPI(api.Filters, &diags)

	m.Legend = &models.WaffleLegendModel{}
	waffleLegendFromAPI(ctx, m.Legend, api.Legend)

	if api.Styling.Values.Mode != nil || api.Styling.Values.PercentDecimals != nil {
		m.ValueDisplay = &models.WaffleValueDisplay{
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
				populatePieChartMetricDefaults,
			)
			if i < len(priorMetrics) {
				cfg = preservePriorJSONWithDefaultsIfEquivalent(ctx, priorMetrics[i].Config, cfg, &diags)
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
				populateLensGroupByDefaults,
			)
			if i < len(priorGroupBy) {
				cfg = preservePriorJSONWithDefaultsIfEquivalent(ctx, priorGroupBy[i].Config, cfg, &diags)
			}
			m.GroupBy[i].Config = cfg
		}
	}

	m.EsqlMetrics = nil
	m.EsqlGroupBy = nil

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

func waffleConfigFromAPIESQL(ctx context.Context, m *models.WaffleConfigModel, dashboard *models.DashboardModel, prior *models.WaffleConfigModel, api kbapi.WaffleESQL) diag.Diagnostics {
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

	datasetBytes, err := json.Marshal(api.DataSource)
	dv, ok := marshalToNormalized(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = dv

	m.Query = nil

	m.Filters = populateFiltersFromAPI(api.Filters, &diags)

	m.Legend = &models.WaffleLegendModel{}
	waffleLegendFromAPI(ctx, m.Legend, api.Legend)

	if api.Styling.Values.Mode != nil || api.Styling.Values.PercentDecimals != nil {
		m.ValueDisplay = &models.WaffleValueDisplay{
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
				if staticColor, colorErr := met.Color.AsStaticColor(); colorErr == nil {
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
					if string(b) == jsonNullString || len(b) == 0 {
						b = []byte(defaultNumberFormatJSON)
					}
					return jsontypes.NewNormalizedValue(normalizeKibanaLensNumberFormatJSONString(string(b)))
				}(),
				Color: &models.WaffleStaticColor{
					Type:  colorType,
					Color: colorValue,
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
		m.EsqlGroupBy = make([]models.WaffleEsqlGroupBy, len(*api.GroupBy))
		for i, gb := range *api.GroupBy {
			colorBytes, err := json.Marshal(gb.Color)
			if err != nil {
				diags.AddError("Failed to marshal esql group_by color", err.Error())
				continue
			}
			formatBytes, err := json.Marshal(gb.Format)
			if err != nil {
				diags.AddError("Failed to marshal esql group_by format", err.Error())
				continue
			}
			if string(formatBytes) == jsonNullString || len(formatBytes) == 0 {
				formatBytes = []byte(defaultNumberFormatJSON)
			}
			formatStr := normalizeKibanaLensNumberFormatJSONString(string(formatBytes))
			eg := models.WaffleEsqlGroupBy{
				Column:     types.StringValue(gb.Column),
				CollapseBy: types.StringValue(string(gb.CollapseBy)),
				ColorJSON:  jsontypes.NewNormalizedValue(string(colorBytes)),
				FormatJSON: jsontypes.NewNormalizedValue(formatStr),
				Label:      types.StringNull(),
			}
			if gb.Label != nil {
				eg.Label = types.StringValue(*gb.Label)
			}
			m.EsqlGroupBy[i] = eg
		}
	}

	m.Metrics = nil
	m.GroupBy = nil

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

func waffleLegendFromAPI(ctx context.Context, m *models.WaffleLegendModel, api kbapi.WaffleLegend) {
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
	if api.Visibility != nil {
		m.Visible = types.StringValue(string(*api.Visibility))
	} else {
		m.Visible = types.StringNull()
	}
}

func waffleLegendToAPI(m *models.WaffleLegendModel) (kbapi.WaffleLegend, diag.Diagnostics) {
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
		v := kbapi.WaffleLegendVisibility(m.Visible.ValueString())
		leg.Visibility = &v
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

func waffleConfigToAPI(m *models.WaffleConfigModel, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	diags.Append(waffleConfigModeValidateDiags(waffleConfigUsesESQL(m), waffleModeListStateFromSlice(len(m.Metrics)),
		waffleModeListStateFromSlice(len(m.GroupBy)),
		waffleModeListStateFromSlice(len(m.EsqlMetrics)),
		waffleModeListStateFromSlice(len(m.EsqlGroupBy)),
		nil,
	)...)
	if diags.HasError() {
		return attrs, diags
	}

	if waffleConfigUsesESQL(m) {
		esql, d := waffleConfigToAPIESQL(m, dashboard)
		diags.Append(d...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromWaffleESQL(esql); err != nil {
			diags.AddError("Failed to build waffle ES|QL chart", err.Error())
		}
		return attrs, diags
	}

	noESQL, d := waffleConfigToAPINoESQL(m, dashboard)
	diags.Append(d...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromWaffleNoESQL(noESQL); err != nil {
		diags.AddError("Failed to build waffle chart", err.Error())
	}
	return attrs, diags
}

func waffleConfigToAPINoESQL(m *models.WaffleConfigModel, dashboard *models.DashboardModel) (kbapi.WaffleNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.WaffleNoESQL{
		Type: kbapi.WaffleNoESQLTypeWaffle,
	}

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
	api.Query = filterSimpleToAPI(m.Query)

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if m.Legend == nil {
		diags.AddError("Missing legend", "waffle_config.legend must be provided")
		return api, diags
	}
	leg, ld := waffleLegendToAPI(m.Legend)
	diags.Append(ld...)
	api.Legend = leg

	if m.ValueDisplay != nil && typeutils.IsKnown(m.ValueDisplay.Mode) {
		mode := kbapi.ValueDisplayMode(m.ValueDisplay.Mode.ValueString())
		vd := kbapi.ValueDisplay{
			Mode: &mode,
		}
		if typeutils.IsKnown(m.ValueDisplay.PercentDecimals) {
			p := float32(m.ValueDisplay.PercentDecimals.ValueFloat64())
			vd.PercentDecimals = &p
		}
		api.Styling.Values = vd
	} else {
		// Required by the Dashboard API; omitting mode yields HTTP 400.
		mode := kbapi.ValueDisplayModePercentage
		api.Styling.Values = kbapi.ValueDisplay{
			Mode: &mode,
		}
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
		items, ddDiags := decodeLensDrilldownSlice[kbapi.WaffleNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}

func waffleConfigToAPIESQL(m *models.WaffleConfigModel, dashboard *models.DashboardModel) (kbapi.WaffleESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.WaffleESQL{
		Type: kbapi.WaffleESQLTypeWaffle,
	}

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

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing dataset", "waffle_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal data_source_json", err.Error())
		return api, diags
	}

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if m.Legend == nil {
		diags.AddError("Missing legend", "waffle_config.legend must be provided")
		return api, diags
	}
	leg, ld := waffleLegendToAPI(m.Legend)
	diags.Append(ld...)
	api.Legend = leg

	if m.ValueDisplay != nil && typeutils.IsKnown(m.ValueDisplay.Mode) {
		mode := kbapi.ValueDisplayMode(m.ValueDisplay.Mode.ValueString())
		vd := kbapi.ValueDisplay{
			Mode: &mode,
		}
		if typeutils.IsKnown(m.ValueDisplay.PercentDecimals) {
			p := float32(m.ValueDisplay.PercentDecimals.ValueFloat64())
			vd.PercentDecimals = &p
		}
		api.Styling.Values = vd
	} else {
		mode := kbapi.ValueDisplayModePercentage
		api.Styling.Values = kbapi.ValueDisplay{
			Mode: &mode,
		}
	}

	metrics := make([]struct {
		Color  *kbapi.WaffleESQL_Metrics_Color `json:"color,omitempty"`
		Column string                          `json:"column"`
		Format kbapi.FormatType                `json:"format"`
		Label  *string                         `json:"label,omitempty"`
	}, len(m.EsqlMetrics))
	for i, em := range m.EsqlMetrics {
		if err := json.Unmarshal([]byte(em.FormatJSON.ValueString()), &metrics[i].Format); err != nil {
			diags.AddError("Failed to unmarshal format_json", err.Error())
		}
		metrics[i].Column = em.Column.ValueString()
		if em.Color == nil {
			diags.AddError("Missing color", "waffle_config.esql_metrics color is required")
			continue
		}
		staticColor := kbapi.StaticColor{
			Type:  kbapi.StaticColorType(em.Color.Type.ValueString()),
			Color: em.Color.Color.ValueString(),
		}
		var color kbapi.WaffleESQL_Metrics_Color
		if err := color.FromStaticColor(staticColor); err != nil {
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
		gb := make([]struct {
			CollapseBy kbapi.CollapseBy   `json:"collapse_by"`
			Color      kbapi.ColorMapping `json:"color"`
			Column     string             `json:"column"`
			Format     kbapi.FormatType   `json:"format"`
			Label      *string            `json:"label,omitempty"`
		}, len(m.EsqlGroupBy))
		for i, eg := range m.EsqlGroupBy {
			if err := json.Unmarshal([]byte(eg.ColorJSON.ValueString()), &gb[i].Color); err != nil {
				diags.AddError("Failed to unmarshal color_json", err.Error())
			}
			gb[i].Column = eg.Column.ValueString()
			gb[i].CollapseBy = kbapi.CollapseBy(eg.CollapseBy.ValueString())
			formatSrc := defaultNumberFormatJSON
			if typeutils.IsKnown(eg.FormatJSON) {
				formatSrc = eg.FormatJSON.ValueString()
			}
			if err := json.Unmarshal([]byte(formatSrc), &gb[i].Format); err != nil {
				diags.AddError("Failed to unmarshal format_json", err.Error())
			}
			if typeutils.IsKnown(eg.Label) {
				s := eg.Label.ValueString()
				gb[i].Label = &s
			}
		}
		api.GroupBy = &gb
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
		items, ddDiags := decodeLensDrilldownSlice[kbapi.WaffleESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}
