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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricCustomIndicator_ToAPI(t *testing.T) {
	t.Run("returns ok=false when not configured", func(t *testing.T) {
		m := tfModel{}
		ok, _, diags := m.metricCustomIndicatorToAPI()
		require.False(t, ok)
		require.False(t, diags.HasError())
	})

	t.Run("emits error when good/total cardinality invalid", func(t *testing.T) {
		m := tfModel{MetricCustomIndicator: []tfMetricCustomIndicator{{
			Index:          types.StringValue("metrics-*"),
			DataViewID:     types.StringNull(),
			Filter:         types.StringNull(),
			TimestampField: types.StringValue("@timestamp"),
			Good:           nil,
			Total:          nil,
		}}}

		ok, _, diags := m.metricCustomIndicatorToAPI()
		require.True(t, ok)
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Summary(), "Invalid configuration")
	})

	t.Run("maps equations, metrics and optional pointers", func(t *testing.T) {
		m := tfModel{MetricCustomIndicator: []tfMetricCustomIndicator{{
			Index:          types.StringValue("metrics-*"),
			DataViewID:     types.StringValue("dv-1"),
			Filter:         types.StringValue("labels.env:prod"),
			TimestampField: types.StringValue("@timestamp"),
			Good: []tfMetricCustomEquation{{
				Equation: types.StringValue("a / b"),
				Metrics: []tfMetricCustomMetric{{
					Name:        types.StringValue("a"),
					Aggregation: types.StringValue("sum"),
					Field:       types.StringValue("good"),
					Filter:      types.StringNull(),
				}},
			}},
			Total: []tfMetricCustomEquation{{
				Equation: types.StringValue("c"),
				Metrics: []tfMetricCustomMetric{{
					Name:        types.StringValue("c"),
					Aggregation: types.StringValue("sum"),
					Field:       types.StringValue("total"),
					Filter:      types.StringValue(testTimesliceSumFilter),
				}},
			}},
		}}}

		ok, ind, diags := m.metricCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())

		apiInd, err := ind.AsSLOsIndicatorPropertiesCustomMetric()
		require.NoError(t, err)

		params := apiInd.Params
		require.NotNil(t, params.DataViewId)
		assert.Equal(t, "dv-1", *params.DataViewId)
		require.NotNil(t, params.Filter)
		assert.Equal(t, "labels.env:prod", *params.Filter)

		require.Len(t, params.Good.Metrics, 1)
		goodMetric, err := params.Good.Metrics[0].AsSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0()
		require.NoError(t, err)
		assert.Equal(t, "a", goodMetric.Name)
		assert.Nil(t, goodMetric.Filter)
		assert.Equal(t, "a / b", params.Good.Equation)

		require.Len(t, params.Total.Metrics, 1)
		totalMetric, err := params.Total.Metrics[0].AsSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0()
		require.NoError(t, err)
		assert.Equal(t, "c", totalMetric.Name)
		require.NotNil(t, totalMetric.Filter)
		assert.Equal(t, testTimesliceSumFilter, *totalMetric.Filter)
		assert.Equal(t, "c", params.Total.Equation)
	})

	t.Run("uses Metrics1 for doc_count aggregation", func(t *testing.T) {
		goodFilter := testTimesliceSumFilter
		m := tfModel{MetricCustomIndicator: []tfMetricCustomIndicator{{
			Index:          types.StringValue("metrics-*"),
			DataViewID:     types.StringNull(),
			Filter:         types.StringNull(),
			TimestampField: types.StringValue("@timestamp"),
			Good: []tfMetricCustomEquation{{
				Equation: types.StringValue("A"),
				Metrics: []tfMetricCustomMetric{{
					Name:        types.StringValue("A"),
					Aggregation: types.StringValue("doc_count"),
					Field:       types.StringNull(),
					Filter:      types.StringValue(goodFilter),
				}},
			}},
			Total: []tfMetricCustomEquation{{
				Equation: types.StringValue("B"),
				Metrics: []tfMetricCustomMetric{{
					Name:        types.StringValue("B"),
					Aggregation: types.StringValue("doc_count"),
					Field:       types.StringNull(),
					Filter:      types.StringNull(),
				}},
			}},
		}}}

		ok, ind, diags := m.metricCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())

		apiInd, err := ind.AsSLOsIndicatorPropertiesCustomMetric()
		require.NoError(t, err)

		params := apiInd.Params

		require.Len(t, params.Good.Metrics, 1)
		goodMetric, err := params.Good.Metrics[0].AsSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1()
		require.NoError(t, err)
		assert.Equal(t, "A", goodMetric.Name)
		assert.Equal(t, kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1AggregationDocCount, goodMetric.Aggregation)
		require.NotNil(t, goodMetric.Filter)
		assert.Equal(t, goodFilter, *goodMetric.Filter)

		require.Len(t, params.Total.Metrics, 1)
		totalMetric, err := params.Total.Metrics[0].AsSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1()
		require.NoError(t, err)
		assert.Equal(t, "B", totalMetric.Name)
		assert.Equal(t, kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1AggregationDocCount, totalMetric.Aggregation)
		assert.Nil(t, totalMetric.Filter)
	})
}

func TestMetricCustomIndicator_PopulateFromAPI(t *testing.T) {
	buildGoodItem := func(t *testing.T, name, agg, field string, filter *string) kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item {
		t.Helper()
		var item kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item
		m0 := kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0{
			Name:        name,
			Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0Aggregation(agg),
			Field:       field,
			Filter:      filter,
		}
		require.NoError(t, item.FromSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics0(m0))
		return item
	}
	buildTotalItemFn := func(t *testing.T, name, agg, field string, filter *string) kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item {
		t.Helper()
		var item kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item
		m0 := kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0{
			Name:        name,
			Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0Aggregation(agg),
			Field:       field,
			Filter:      filter,
		}
		require.NoError(t, item.FromSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics0(m0))
		return item
	}
	buildGoodItemDocCount := func(t *testing.T, name string, filter *string) kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item {
		t.Helper()
		var item kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item
		m1 := kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1{
			Name:        name,
			Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1AggregationDocCount,
			Filter:      filter,
		}
		require.NoError(t, item.FromSLOsIndicatorPropertiesCustomMetricParamsGoodMetrics1(m1))
		return item
	}
	buildTotalItemDocCount := func(t *testing.T, name string, filter *string) kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item {
		t.Helper()
		var item kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item
		m1 := kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1{
			Name:        name,
			Aggregation: kbapi.SLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1AggregationDocCount,
			Filter:      filter,
		}
		require.NoError(t, item.FromSLOsIndicatorPropertiesCustomMetricParamsTotalMetrics1(m1))
		return item
	}

	t.Run("maps equations, metrics and optional pointers", func(t *testing.T) {
		dvID := "dv-1"
		overallFilter := "labels.env:prod"
		totalFilter := testTimesliceSumFilter

		api := kbapi.SLOsIndicatorPropertiesCustomMetric{
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
				Index:          "metrics-*",
				DataViewId:     &dvID,
				Filter:         &overallFilter,
				TimestampField: "@timestamp",
				Good: struct {
					Equation string                                                               `json:"equation"`
					Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item `json:"metrics"`
				}{
					Equation: "a / b",
					Metrics: []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item{
						buildGoodItem(t, "a", "sum", "good", nil),
					},
				},
				Total: struct {
					Equation string                                                                `json:"equation"`
					Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item `json:"metrics"`
				}{
					Equation: "c",
					Metrics: []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item{
						buildTotalItemFn(t, "c", "sum", "total", &totalFilter),
					},
				},
			},
		}

		var m tfModel
		diags := m.populateFromMetricCustomIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.MetricCustomIndicator, 1)

		ind := m.MetricCustomIndicator[0]
		assert.Equal(t, "metrics-*", ind.Index.ValueString())
		assert.Equal(t, "dv-1", ind.DataViewID.ValueString())
		assert.Equal(t, "labels.env:prod", ind.Filter.ValueString())
		assert.Equal(t, "@timestamp", ind.TimestampField.ValueString())

		require.Len(t, ind.Good, 1)
		require.Len(t, ind.Good[0].Metrics, 1)
		assert.Equal(t, "a / b", ind.Good[0].Equation.ValueString())
		assert.True(t, ind.Good[0].Metrics[0].Filter.IsNull())

		require.Len(t, ind.Total, 1)
		require.Len(t, ind.Total[0].Metrics, 1)
		assert.Equal(t, "c", ind.Total[0].Equation.ValueString())
		assert.Equal(t, testTimesliceSumFilter, ind.Total[0].Metrics[0].Filter.ValueString())
	})

	t.Run("sets optional fields to null when not present", func(t *testing.T) {
		api := kbapi.SLOsIndicatorPropertiesCustomMetric{
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
				Index:          "metrics-*",
				DataViewId:     nil,
				Filter:         nil,
				TimestampField: "@timestamp",
				Good: struct {
					Equation string                                                               `json:"equation"`
					Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item `json:"metrics"`
				}{
					Equation: "a",
					Metrics:  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item{},
				},
				Total: struct {
					Equation string                                                                `json:"equation"`
					Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item `json:"metrics"`
				}{
					Equation: "b",
					Metrics:  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item{},
				},
			},
		}

		var m tfModel
		diags := m.populateFromMetricCustomIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.MetricCustomIndicator, 1)

		ind := m.MetricCustomIndicator[0]
		assert.True(t, ind.DataViewID.IsNull())
		assert.True(t, ind.Filter.IsNull())
	})

	t.Run("maps doc_count metrics without field", func(t *testing.T) {
		goodFilter := testTimesliceSumFilter

		api := kbapi.SLOsIndicatorPropertiesCustomMetric{
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
				Index:          "metrics-*",
				TimestampField: "@timestamp",
				Good: struct {
					Equation string                                                               `json:"equation"`
					Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item `json:"metrics"`
				}{
					Equation: "A",
					Metrics: []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Good_Metrics_Item{
						buildGoodItemDocCount(t, "A", &goodFilter),
					},
				},
				Total: struct {
					Equation string                                                                `json:"equation"`
					Metrics  []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item `json:"metrics"`
				}{
					Equation: "B",
					Metrics: []kbapi.SLOsIndicatorPropertiesCustomMetric_Params_Total_Metrics_Item{
						buildTotalItemDocCount(t, "B", nil),
					},
				},
			},
		}

		var m tfModel
		diags := m.populateFromMetricCustomIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.MetricCustomIndicator, 1)

		ind := m.MetricCustomIndicator[0]
		require.Len(t, ind.Good, 1)
		require.Len(t, ind.Good[0].Metrics, 1)
		goodMetric := ind.Good[0].Metrics[0]
		assert.Equal(t, "A", goodMetric.Name.ValueString())
		assert.Equal(t, "doc_count", goodMetric.Aggregation.ValueString())
		assert.True(t, goodMetric.Field.IsNull())
		assert.Equal(t, goodFilter, goodMetric.Filter.ValueString())

		require.Len(t, ind.Total, 1)
		require.Len(t, ind.Total[0].Metrics, 1)
		totalMetric := ind.Total[0].Metrics[0]
		assert.Equal(t, "B", totalMetric.Name.ValueString())
		assert.Equal(t, "doc_count", totalMetric.Aggregation.ValueString())
		assert.True(t, totalMetric.Field.IsNull())
		assert.True(t, totalMetric.Filter.IsNull())
	})
}
