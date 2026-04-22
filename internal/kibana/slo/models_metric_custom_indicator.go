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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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

func buildGoodMetricItem(metric tfMetricCustomMetric) (kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item, error) {
	var item kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item
	if metric.Aggregation.ValueString() == string(kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1AggregationDocCount) {
		if !metric.Field.IsNull() && !metric.Field.IsUnknown() {
			return kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item{}, fmt.Errorf("field must not be set when aggregation is doc_count")
		}
		m1 := kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1{
			Name:        metric.Name.ValueString(),
			Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1Aggregation(metric.Aggregation.ValueString()),
			Filter:      stringPtr(metric.Filter),
		}
		if err := item.FromSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1(m1); err != nil {
			return kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item{}, err
		}
	} else {
		if metric.Field.IsNull() || metric.Field.IsUnknown() || metric.Field.ValueString() == "" {
			return kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item{}, fmt.Errorf("field is required when aggregation is not doc_count")
		}
		m0 := kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0{
			Name:        metric.Name.ValueString(),
			Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0Aggregation(metric.Aggregation.ValueString()),
			Field:       metric.Field.ValueString(),
			Filter:      stringPtr(metric.Filter),
		}
		if err := item.FromSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0(m0); err != nil {
			return kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item{}, err
		}
	}
	return item, nil
}

func buildTotalMetricItem(metric tfMetricCustomMetric) (kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item, error) {
	var item kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item
	if metric.Aggregation.ValueString() == string(kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1AggregationDocCount) {
		if !metric.Field.IsNull() && !metric.Field.IsUnknown() {
			return kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item{}, fmt.Errorf("field must not be set when aggregation is doc_count")
		}
		m1 := kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1{
			Name:        metric.Name.ValueString(),
			Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1Aggregation(metric.Aggregation.ValueString()),
			Filter:      stringPtr(metric.Filter),
		}
		if err := item.FromSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1(m1); err != nil {
			return kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item{}, err
		}
	} else {
		if metric.Field.IsNull() || metric.Field.IsUnknown() || metric.Field.ValueString() == "" {
			return kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item{}, fmt.Errorf("field is required when aggregation is not doc_count")
		}
		m0 := kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0{
			Name:        metric.Name.ValueString(),
			Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0Aggregation(metric.Aggregation.ValueString()),
			Field:       metric.Field.ValueString(),
			Filter:      stringPtr(metric.Filter),
		}
		if err := item.FromSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0(m0); err != nil {
			return kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item{}, err
		}
	}
	return item, nil
}

func (m tfModel) metricCustomIndicatorToAPI() (bool, kbapi.SLOsSloWithSummaryResponse_Indicator, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(m.MetricCustomIndicator) != 1 {
		return false, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	ind := m.MetricCustomIndicator[0]
	if len(ind.Good) != 1 || len(ind.Total) != 1 {
		diags.AddError("Invalid configuration", "metric_custom_indicator.good and .total must each have exactly 1 item")
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	goodMetrics := make([]kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item, 0, len(ind.Good[0].Metrics))
	for i, metric := range ind.Good[0].Metrics {
		item, err := buildGoodMetricItem(metric)
		if err != nil {
			diags.AddError("Invalid configuration", fmt.Sprintf("good.metrics[%d]: %s", i, err.Error()))
			return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
		}
		goodMetrics = append(goodMetrics, item)
	}

	totalMetrics := make([]kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item, 0, len(ind.Total[0].Metrics))
	for i, metric := range ind.Total[0].Metrics {
		item, err := buildTotalMetricItem(metric)
		if err != nil {
			diags.AddError("Invalid configuration", fmt.Sprintf("total.metrics[%d]: %s", i, err.Error()))
			return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
		}
		totalMetrics = append(totalMetrics, item)
	}

	metricIndicator := kbapi.SLOsIndicatorPropertiesCustomMetric{
		Type: indicatorAddressToType["metric_custom_indicator"],
		Params: struct {
			DataViewId *string `json:"dataViewId,omitempty"` //nolint:revive // var-naming: API struct field
			Filter     *string `json:"filter,omitempty"`
			Good       struct {
				Equation string                                                               `json:"equation"`
				Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item `json:"metrics"`
			} `json:"good"`
			Index          string `json:"index"`
			TimestampField string `json:"timestampField"`
			Total          struct {
				Equation string                                                                `json:"equation"`
				Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item `json:"metrics"`
			} `json:"total"`
		}{
			Index:          ind.Index.ValueString(),
			DataViewId:     stringPtr(ind.DataViewID),
			Filter:         stringPtr(ind.Filter),
			TimestampField: ind.TimestampField.ValueString(),
			Good: struct {
				Equation string                                                               `json:"equation"`
				Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item `json:"metrics"`
			}{
				Metrics:  goodMetrics,
				Equation: ind.Good[0].Equation.ValueString(),
			},
			Total: struct {
				Equation string                                                                `json:"equation"`
				Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item `json:"metrics"`
			}{
				Metrics:  totalMetrics,
				Equation: ind.Total[0].Equation.ValueString(),
			},
		},
	}

	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	if err := result.FromSLOsIndicatorPropertiesCustomMetric(metricIndicator); err != nil {
		diags.AddError("Failed to build Custom Metric indicator", err.Error())
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}
	return true, result, diags
}

func (m *tfModel) populateFromMetricCustomIndicator(apiIndicator kbapi.SLOsIndicatorPropertiesCustomMetric) diag.Diagnostics {
	diags := diag.Diagnostics{}

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
	for i, mtr := range p.Good.Metrics {
		// Dispatch on aggregation value: doc_count → Metrics1, otherwise → Metrics0.
		// As* calls always unmarshal; we check the aggregation field to pick the right variant.
		m1, _ := mtr.AsSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1()
		if string(m1.Aggregation) == string(kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1AggregationDocCount) {
			goodMetrics = append(goodMetrics, tfMetricCustomMetric{
				Name:        types.StringValue(m1.Name),
				Aggregation: types.StringValue(string(m1.Aggregation)),
				Field:       types.StringNull(),
				Filter:      stringOrNull(m1.Filter),
			})
		} else {
			m0, _ := mtr.AsSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0()
			if m0.Name == "" {
				diags.AddError("Unexpected API response", fmt.Sprintf("good.metrics[%d]: unrecognised metric variant", i))
				return diags
			}
			goodMetrics = append(goodMetrics, tfMetricCustomMetric{
				Name:        types.StringValue(m0.Name),
				Aggregation: types.StringValue(string(m0.Aggregation)),
				Field:       types.StringValue(m0.Field),
				Filter:      stringOrNull(m0.Filter),
			})
		}
	}

	totalMetrics := make([]tfMetricCustomMetric, 0, len(p.Total.Metrics))
	for i, mtr := range p.Total.Metrics {
		// Dispatch on aggregation value: doc_count → Metrics1, otherwise → Metrics0.
		m1, _ := mtr.AsSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1()
		if string(m1.Aggregation) == string(kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1AggregationDocCount) {
			totalMetrics = append(totalMetrics, tfMetricCustomMetric{
				Name:        types.StringValue(m1.Name),
				Aggregation: types.StringValue(string(m1.Aggregation)),
				Field:       types.StringNull(),
				Filter:      stringOrNull(m1.Filter),
			})
		} else {
			m0, _ := mtr.AsSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0()
			if m0.Name == "" {
				diags.AddError("Unexpected API response", fmt.Sprintf("total.metrics[%d]: unrecognised metric variant", i))
				return diags
			}
			totalMetrics = append(totalMetrics, tfMetricCustomMetric{
				Name:        types.StringValue(m0.Name),
				Aggregation: types.StringValue(string(m0.Aggregation)),
				Field:       types.StringValue(m0.Field),
				Filter:      stringOrNull(m0.Filter),
			})
		}
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
