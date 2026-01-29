package slo

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfMetricCustomIndicator struct {
	Index          types.String             `tfsdk:"index"`
	DataViewID     types.String             `tfsdk:"data_view_id"`
	Filter         types.String             `tfsdk:"filter"`
	TimestampField types.String             `tfsdk:"timestamp_field"`
	Good           []tfMetricCustomEquation `tfsdk:"good"`
	Total          []tfMetricCustomEquation `tfsdk:"total"`
}

type tfMetricCustomEquation struct {
	Metrics  []tfMetricCustomMetric `tfsdk:"metrics"`
	Equation types.String           `tfsdk:"equation"`
}

type tfMetricCustomMetric struct {
	Name        types.String `tfsdk:"name"`
	Aggregation types.String `tfsdk:"aggregation"`
	Field       types.String `tfsdk:"field"`
	Filter      types.String `tfsdk:"filter"`
}

func (m tfModel) metricCustomIndicatorToAPI() (bool, slo.SloWithSummaryResponseIndicator, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(m.MetricCustomIndicator) != 1 {
		return false, slo.SloWithSummaryResponseIndicator{}, diags
	}

	ind := m.MetricCustomIndicator[0]
	if len(ind.Good) != 1 || len(ind.Total) != 1 {
		diags.AddError("Invalid configuration", "metric_custom_indicator.good and .total must each have exactly 1 item")
		return true, slo.SloWithSummaryResponseIndicator{}, diags
	}

	goodMetrics := make([]slo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner, 0, len(ind.Good[0].Metrics))
	for _, metric := range ind.Good[0].Metrics {
		goodMetrics = append(goodMetrics, slo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner{
			Name:        metric.Name.ValueString(),
			Aggregation: metric.Aggregation.ValueString(),
			Field:       metric.Field.ValueString(),
			Filter:      stringPtr(metric.Filter),
		})
	}
	totalMetrics := make([]slo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner, 0, len(ind.Total[0].Metrics))
	for _, metric := range ind.Total[0].Metrics {
		totalMetrics = append(totalMetrics, slo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner{
			Name:        metric.Name.ValueString(),
			Aggregation: metric.Aggregation.ValueString(),
			Field:       metric.Field.ValueString(),
			Filter:      stringPtr(metric.Filter),
		})
	}

	return true, slo.SloWithSummaryResponseIndicator{
		IndicatorPropertiesCustomMetric: &slo.IndicatorPropertiesCustomMetric{
			Type: indicatorAddressToType["metric_custom_indicator"],
			Params: slo.IndicatorPropertiesCustomMetricParams{
				Index:          ind.Index.ValueString(),
				DataViewId:     stringPtr(ind.DataViewID),
				Filter:         stringPtr(ind.Filter),
				TimestampField: ind.TimestampField.ValueString(),
				Good: slo.IndicatorPropertiesCustomMetricParamsGood{
					Metrics:  goodMetrics,
					Equation: ind.Good[0].Equation.ValueString(),
				},
				Total: slo.IndicatorPropertiesCustomMetricParamsTotal{
					Metrics:  totalMetrics,
					Equation: ind.Total[0].Equation.ValueString(),
				},
			},
		},
	}, diags
}

func (m *tfModel) populateFromMetricCustomIndicator(apiIndicator *slo.IndicatorPropertiesCustomMetric) diag.Diagnostics {
	var diags diag.Diagnostics
	if apiIndicator == nil {
		return diags
	}

	p := apiIndicator.Params
	ind := tfMetricCustomIndicator{
		Index:          types.StringValue(p.Index),
		TimestampField: types.StringValue(p.TimestampField),
		Filter:         stringOrNull(p.Filter),
		DataViewID:     types.StringNull(),
	}
	if p.DataViewId != nil {
		ind.DataViewID = types.StringValue(*p.DataViewId)
	}

	goodMetrics := make([]tfMetricCustomMetric, 0, len(p.Good.Metrics))
	for _, mtr := range p.Good.Metrics {
		goodMetrics = append(goodMetrics, tfMetricCustomMetric{
			Name:        types.StringValue(mtr.Name),
			Aggregation: types.StringValue(mtr.Aggregation),
			Field:       types.StringValue(mtr.Field),
			Filter:      stringOrNull(mtr.Filter),
		})
	}

	totalMetrics := make([]tfMetricCustomMetric, 0, len(p.Total.Metrics))
	for _, mtr := range p.Total.Metrics {
		totalMetrics = append(totalMetrics, tfMetricCustomMetric{
			Name:        types.StringValue(mtr.Name),
			Aggregation: types.StringValue(mtr.Aggregation),
			Field:       types.StringValue(mtr.Field),
			Filter:      stringOrNull(mtr.Filter),
		})
	}

	ind.Good = []tfMetricCustomEquation{{
		Equation: types.StringValue(p.Good.Equation),
		Metrics:  goodMetrics,
	}}
	ind.Total = []tfMetricCustomEquation{{
		Equation: types.StringValue(p.Total.Equation),
		Metrics:  totalMetrics,
	}}

	m.MetricCustomIndicator = []tfMetricCustomIndicator{ind}
	return diags
}
