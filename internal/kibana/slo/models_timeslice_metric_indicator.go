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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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

func buildTimesliceMetricItem(metric tfTimesliceMetricMetric, idx int) (kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	var item kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item

	var filter *string
	if typeutils.IsKnown(metric.Filter) {
		filter = metric.Filter.ValueStringPointer()
	}

	agg := metric.Aggregation.ValueString()
	switch {
	case slices.Contains(timesliceMetricAggregationsBasic, agg):
		bm := kbapi.SLOsTimesliceMetricBasicMetricWithField{
			Name:        metric.Name.ValueString(),
			Aggregation: kbapi.SLOsTimesliceMetricBasicMetricWithFieldAggregation(agg),
			Field:       metric.Field.ValueString(),
			Filter:      filter,
		}
		if err := item.FromSLOsTimesliceMetricBasicMetricWithField(bm); err != nil {
			diags.AddError("Invalid configuration", fmt.Sprintf("metrics[%d]: %s", idx, err.Error()))
		}
	case agg == timesliceMetricAggregationPercentile:
		pm := kbapi.SLOsTimesliceMetricPercentileMetric{
			Name:        metric.Name.ValueString(),
			Aggregation: kbapi.SLOsTimesliceMetricPercentileMetricAggregation(agg),
			Field:       metric.Field.ValueString(),
			Percentile:  metric.Percentile.ValueFloat64(),
			Filter:      filter,
		}
		if err := item.FromSLOsTimesliceMetricPercentileMetric(pm); err != nil {
			diags.AddError("Invalid configuration", fmt.Sprintf("metrics[%d]: %s", idx, err.Error()))
		}
	case agg == timesliceMetricAggregationDocCount:
		dm := kbapi.SLOsTimesliceMetricDocCountMetric{
			Name:        metric.Name.ValueString(),
			Aggregation: kbapi.SLOsTimesliceMetricDocCountMetricAggregation(agg),
			Filter:      filter,
		}
		if err := item.FromSLOsTimesliceMetricDocCountMetric(dm); err != nil {
			diags.AddError("Invalid configuration", fmt.Sprintf("metrics[%d]: %s", idx, err.Error()))
		}
	default:
		diags.AddError("Invalid configuration", fmt.Sprintf("metrics[%d]: unsupported aggregation '%s'", idx, agg))
	}
	return item, diags
}

func (m tfModel) timesliceMetricIndicatorToAPI() (bool, kbapi.SLOsSloWithSummaryResponse_Indicator, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(m.TimesliceMetricIndicator) != 1 {
		return false, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	ind := m.TimesliceMetricIndicator[0]
	if len(ind.Metric) != 1 {
		diags.AddError("Invalid configuration", "timeslice_metric_indicator.metric must have exactly 1 item")
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}
	metricDef := ind.Metric[0]

	metrics := make([]kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item, 0, len(metricDef.Metrics))
	for i, metric := range metricDef.Metrics {
		item, itemDiags := buildTimesliceMetricItem(metric, i)
		diags.Append(itemDiags...)
		if diags.HasError() {
			return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
		}
		metrics = append(metrics, item)
	}

	tsIndicator := kbapi.SLOsIndicatorPropertiesTimesliceMetric{
		Type: indicatorAddressToType["timeslice_metric_indicator"],
		Params: struct {
			DataViewId *string `json:"dataViewId,omitempty"` //nolint:revive // var-naming: API struct field
			Filter     *string `json:"filter,omitempty"`
			Index      string  `json:"index"`
			Metric     struct {
				Comparator kbapi.SLOsIndicatorPropertiesTimesliceMetricParamsMetricComparator        `json:"comparator"`
				Equation   string                                                                    `json:"equation"`
				Metrics    []kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item `json:"metrics"`
				Threshold  float64                                                                   `json:"threshold"`
			} `json:"metric"`
			TimestampField string `json:"timestampField"`
		}{
			Index:          ind.Index.ValueString(),
			DataViewId:     typeutils.ValueStringPointer(ind.DataViewID),
			TimestampField: ind.TimestampField.ValueString(),
			Filter:         typeutils.ValueStringPointer(ind.Filter),
			Metric: struct {
				Comparator kbapi.SLOsIndicatorPropertiesTimesliceMetricParamsMetricComparator        `json:"comparator"`
				Equation   string                                                                    `json:"equation"`
				Metrics    []kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item `json:"metrics"`
				Threshold  float64                                                                   `json:"threshold"`
			}{
				Metrics:    metrics,
				Equation:   metricDef.Equation.ValueString(),
				Comparator: kbapi.SLOsIndicatorPropertiesTimesliceMetricParamsMetricComparator(metricDef.Comparator.ValueString()),
				Threshold:  metricDef.Threshold.ValueFloat64(),
			},
		},
	}

	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	if err := result.FromSLOsIndicatorPropertiesTimesliceMetric(tsIndicator); err != nil {
		diags.AddError("Failed to build Timeslice Metric indicator", err.Error())
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}
	return true, result, diags
}

func (m *tfModel) populateFromTimesliceMetricIndicator(apiIndicator kbapi.SLOsIndicatorPropertiesTimesliceMetric) diag.Diagnostics {
	diags := diag.Diagnostics{}

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
		// All As* calls on kbapi unions succeed because they just unmarshal raw JSON.
		// Use the aggregation field value to determine which variant to read, mirroring
		// the write-path switch in buildTimesliceMetricItem.
		if pm, err := mm.AsSLOsTimesliceMetricPercentileMetric(); err == nil && string(pm.Aggregation) == timesliceMetricAggregationPercentile {
			metric.Name = types.StringValue(pm.Name)
			metric.Aggregation = types.StringValue(string(pm.Aggregation))
			metric.Field = types.StringValue(pm.Field)
			metric.Percentile = types.Float64Value(pm.Percentile)
			metric.Filter = types.StringPointerValue(pm.Filter)
		} else if dm, err := mm.AsSLOsTimesliceMetricDocCountMetric(); err == nil && string(dm.Aggregation) == timesliceMetricAggregationDocCount {
			metric.Name = types.StringValue(dm.Name)
			metric.Aggregation = types.StringValue(string(dm.Aggregation))
			metric.Filter = types.StringPointerValue(dm.Filter)
		} else if bm, err := mm.AsSLOsTimesliceMetricBasicMetricWithField(); err == nil && bm.Name != "" {
			metric.Name = types.StringValue(bm.Name)
			metric.Aggregation = types.StringValue(string(bm.Aggregation))
			metric.Field = types.StringValue(bm.Field)
			metric.Filter = types.StringPointerValue(bm.Filter)
		} else {
			diags.AddError(
				"Unrecognized timeslice metric aggregation type",
				"Could not determine the aggregation type for a timeslice metric entry. The API returned an unrecognized metric variant.",
			)
			return diags
		}
		tm = append(tm, metric)
	}

	ind.Metric = []tfTimesliceMetricDefinition{{
		Metrics:    tm,
		Equation:   types.StringValue(p.Metric.Equation),
		Comparator: types.StringValue(string(p.Metric.Comparator)),
		Threshold:  types.Float64Value(p.Metric.Threshold),
	}}

	m.TimesliceMetricIndicator = []tfTimesliceMetricIndicator{ind}
	return diags
}
