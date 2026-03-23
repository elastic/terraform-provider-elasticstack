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

package slo

import (
	"fmt"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfTimesliceMetricIndicator struct {
	Index          types.String                  `tfsdk:"index"`
	DataViewID     types.String                  `tfsdk:"data_view_id"`
	TimestampField types.String                  `tfsdk:"timestamp_field"`
	Filter         types.String                  `tfsdk:"filter"`
	Metric         []tfTimesliceMetricDefinition `tfsdk:"metric"`
}

type tfTimesliceMetricDefinition struct {
	Metrics    []tfTimesliceMetricMetric `tfsdk:"metrics"`
	Equation   types.String              `tfsdk:"equation"`
	Comparator types.String              `tfsdk:"comparator"`
	Threshold  types.Float64             `tfsdk:"threshold"`
}

type tfTimesliceMetricMetric struct {
	Name        types.String  `tfsdk:"name"`
	Aggregation types.String  `tfsdk:"aggregation"`
	Field       types.String  `tfsdk:"field"`
	Percentile  types.Float64 `tfsdk:"percentile"`
	Filter      types.String  `tfsdk:"filter"`
}

func (m tfModel) timesliceMetricIndicatorToAPI() (bool, slo.SloWithSummaryResponseIndicator, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(m.TimesliceMetricIndicator) != 1 {
		return false, slo.SloWithSummaryResponseIndicator{}, diags
	}

	ind := m.TimesliceMetricIndicator[0]
	if len(ind.Metric) != 1 {
		diags.AddError("Invalid configuration", "timeslice_metric_indicator.metric must have exactly 1 item")
		return true, slo.SloWithSummaryResponseIndicator{}, diags
	}
	metricDef := ind.Metric[0]

	metrics := make([]slo.IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner, 0, len(metricDef.Metrics))
	for i, metric := range metricDef.Metrics {
		var filter *string
		if typeutils.IsKnown(metric.Filter) {
			filter = metric.Filter.ValueStringPointer()
		}

		agg := metric.Aggregation.ValueString()
		switch {
		case slices.Contains(timesliceMetricAggregationsBasic, agg):
			metrics = append(metrics, slo.IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner{
				TimesliceMetricBasicMetricWithField: &slo.TimesliceMetricBasicMetricWithField{
					Name:        metric.Name.ValueString(),
					Aggregation: agg,
					Field:       metric.Field.ValueString(),
					Filter:      filter,
				},
			})
		case agg == timesliceMetricAggregationPercentile:
			metrics = append(metrics, slo.IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner{
				TimesliceMetricPercentileMetric: &slo.TimesliceMetricPercentileMetric{
					Name:        metric.Name.ValueString(),
					Aggregation: agg,
					Field:       metric.Field.ValueString(),
					Percentile:  metric.Percentile.ValueFloat64(),
					Filter:      filter,
				},
			})
		case agg == timesliceMetricAggregationDocCount:
			metrics = append(metrics, slo.IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner{
				TimesliceMetricDocCountMetric: &slo.TimesliceMetricDocCountMetric{
					Name:        metric.Name.ValueString(),
					Aggregation: agg,
					Filter:      filter,
				},
			})
		default:
			diags.AddError("Invalid configuration", fmt.Sprintf("metrics[%d]: unsupported aggregation '%s'", i, agg))
			return true, slo.SloWithSummaryResponseIndicator{}, diags
		}
	}

	return true, slo.SloWithSummaryResponseIndicator{
		IndicatorPropertiesTimesliceMetric: &slo.IndicatorPropertiesTimesliceMetric{
			Type: indicatorAddressToType["timeslice_metric_indicator"],
			Params: slo.IndicatorPropertiesTimesliceMetricParams{
				Index:          ind.Index.ValueString(),
				DataViewId:     stringPtr(ind.DataViewID),
				TimestampField: ind.TimestampField.ValueString(),
				Filter:         stringPtr(ind.Filter),
				Metric: slo.IndicatorPropertiesTimesliceMetricParamsMetric{
					Metrics:    metrics,
					Equation:   metricDef.Equation.ValueString(),
					Comparator: metricDef.Comparator.ValueString(),
					Threshold:  metricDef.Threshold.ValueFloat64(),
				},
			},
		},
	}, diags
}

func (m *tfModel) populateFromTimesliceMetricIndicator(apiIndicator *slo.IndicatorPropertiesTimesliceMetric) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if apiIndicator == nil {
		return diags
	}

	p := apiIndicator.Params
	ind := tfTimesliceMetricIndicator{
		Index:          types.StringValue(p.Index),
		TimestampField: types.StringValue(p.TimestampField),
		Filter:         stringOrNull(p.Filter),
		DataViewID:     types.StringNull(),
	}
	if p.DataViewId != nil {
		ind.DataViewID = types.StringValue(*p.DataViewId)
	}

	tm := make([]tfTimesliceMetricMetric, 0, len(p.Metric.Metrics))
	for _, mm := range p.Metric.Metrics {
		metric := tfTimesliceMetricMetric{
			Field:      types.StringNull(),
			Percentile: types.Float64Null(),
			Filter:     types.StringNull(),
		}
		if mm.TimesliceMetricBasicMetricWithField != nil {
			metric.Name = types.StringValue(mm.TimesliceMetricBasicMetricWithField.Name)
			metric.Aggregation = types.StringValue(mm.TimesliceMetricBasicMetricWithField.Aggregation)
			metric.Field = types.StringValue(mm.TimesliceMetricBasicMetricWithField.Field)
			metric.Filter = types.StringPointerValue(mm.TimesliceMetricBasicMetricWithField.Filter)
		}
		if mm.TimesliceMetricPercentileMetric != nil {
			metric.Name = types.StringValue(mm.TimesliceMetricPercentileMetric.Name)
			metric.Aggregation = types.StringValue(mm.TimesliceMetricPercentileMetric.Aggregation)
			metric.Field = types.StringValue(mm.TimesliceMetricPercentileMetric.Field)
			metric.Percentile = types.Float64Value(mm.TimesliceMetricPercentileMetric.Percentile)
			metric.Filter = types.StringPointerValue(mm.TimesliceMetricPercentileMetric.Filter)
		}
		if mm.TimesliceMetricDocCountMetric != nil {
			metric.Name = types.StringValue(mm.TimesliceMetricDocCountMetric.Name)
			metric.Aggregation = types.StringValue(mm.TimesliceMetricDocCountMetric.Aggregation)
			metric.Filter = types.StringPointerValue(mm.TimesliceMetricDocCountMetric.Filter)
		}
		tm = append(tm, metric)
	}

	ind.Metric = []tfTimesliceMetricDefinition{{
		Metrics:    tm,
		Equation:   types.StringValue(p.Metric.Equation),
		Comparator: types.StringValue(p.Metric.Comparator),
		Threshold:  types.Float64Value(p.Metric.Threshold),
	}}

	m.TimesliceMetricIndicator = []tfTimesliceMetricIndicator{ind}
	return diags
}
