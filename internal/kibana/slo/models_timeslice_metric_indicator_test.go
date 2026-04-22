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

const (
	testTimesliceSumFilter = "status:200"
	testTimesliceDocFilter = "labels.env:prod"
	testTimesliceDvID      = "dv-1"
)

// buildBasicMetricItem builds a timeslice metric item containing a basic metric with field.
func buildBasicMetricItem(t *testing.T, name, agg, field string, filter *string) kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item {
	t.Helper()
	var item kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item
	bm := kbapi.SLOsTimesliceMetricBasicMetricWithField{
		Name:        name,
		Aggregation: kbapi.SLOsTimesliceMetricBasicMetricWithFieldAggregation(agg),
		Field:       field,
		Filter:      filter,
	}
	require.NoError(t, item.FromSLOsTimesliceMetricBasicMetricWithField(bm))
	return item
}

// buildPercentileItem builds a timeslice metric item containing a percentile metric.
func buildPercentileItem(t *testing.T, name, agg, field string, percentile float64, filter *string) kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item {
	t.Helper()
	var item kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item
	pm := kbapi.SLOsTimesliceMetricPercentileMetric{
		Name:        name,
		Aggregation: kbapi.SLOsTimesliceMetricPercentileMetricAggregation(agg),
		Field:       field,
		Percentile:  percentile,
		Filter:      filter,
	}
	require.NoError(t, item.FromSLOsTimesliceMetricPercentileMetric(pm))
	return item
}

// buildDocCountItem builds a timeslice metric item containing a doc_count metric.
func buildDocCountItem(t *testing.T, name string, filter *string) kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item {
	t.Helper()
	var item kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item
	dm := kbapi.SLOsTimesliceMetricDocCountMetric{
		Name:        name,
		Aggregation: "doc_count",
		Filter:      filter,
	}
	require.NoError(t, item.FromSLOsTimesliceMetricDocCountMetric(dm))
	return item
}

// buildTimesliceAPI constructs a kbapi.SLOsIndicatorPropertiesTimesliceMetric.
func buildTimesliceAPI(dvID *string, filter *string, items []kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item) kbapi.SLOsIndicatorPropertiesTimesliceMetric {
	return kbapi.SLOsIndicatorPropertiesTimesliceMetric{
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
			Index:          "metrics-*",
			DataViewId:     dvID,
			Filter:         filter,
			TimestampField: "@timestamp",
			Metric: struct {
				Comparator kbapi.SLOsIndicatorPropertiesTimesliceMetricParamsMetricComparator        `json:"comparator"`
				Equation   string                                                                    `json:"equation"`
				Metrics    []kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item `json:"metrics"`
				Threshold  float64                                                                   `json:"threshold"`
			}{
				Equation:   "a",
				Comparator: "GT",
				Threshold:  1.23,
				Metrics:    items,
			},
		},
	}
}

func TestTimesliceMetricIndicator_ToAPI(t *testing.T) {
	t.Run("returns ok=false when not configured", func(t *testing.T) {
		m := tfModel{}
		ok, _, diags := m.timesliceMetricIndicatorToAPI()
		require.False(t, ok)
		require.False(t, diags.HasError())
	})

	t.Run("emits error when metric cardinality invalid", func(t *testing.T) {
		m := tfModel{TimesliceMetricIndicator: []tfTimesliceMetricIndicator{{
			Index:          types.StringValue("metrics-*"),
			DataViewID:     types.StringNull(),
			TimestampField: types.StringValue("@timestamp"),
			Filter:         types.StringNull(),
			Metric:         nil,
		}}}

		ok, _, diags := m.timesliceMetricIndicatorToAPI()
		require.True(t, ok)
		require.True(t, diags.HasError())
	})

	t.Run("maps all supported metric variants", func(t *testing.T) {
		sumFilter := testTimesliceSumFilter
		docFilter := testTimesliceDocFilter
		m := tfModel{TimesliceMetricIndicator: []tfTimesliceMetricIndicator{{
			Index:          types.StringValue("metrics-*"),
			DataViewID:     types.StringValue(testTimesliceDvID),
			TimestampField: types.StringValue("@timestamp"),
			Filter:         types.StringNull(),
			Metric: []tfTimesliceMetricDefinition{{
				Equation:   types.StringValue("a"),
				Comparator: types.StringValue("GT"),
				Threshold:  types.Float64Value(1.23),
				Metrics: []tfTimesliceMetricMetric{
					{
						Name:        types.StringValue("a"),
						Aggregation: types.StringValue("sum"),
						Field:       types.StringValue("foo"),
						Percentile:  types.Float64Null(),
						Filter:      types.StringValue(sumFilter),
					},
					{
						Name:        types.StringValue("p95"),
						Aggregation: types.StringValue("percentile"),
						Field:       types.StringValue("latency"),
						Percentile:  types.Float64Value(95),
						Filter:      types.StringNull(),
					},
					{
						Name:        types.StringValue("c"),
						Aggregation: types.StringValue("doc_count"),
						Field:       types.StringNull(),
						Percentile:  types.Float64Null(),
						Filter:      types.StringValue(docFilter),
					},
				},
			}},
		}}}

		ok, ind, diags := m.timesliceMetricIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())

		apiInd, err := ind.AsSLOsIndicatorPropertiesTimesliceMetric()
		require.NoError(t, err)

		metrics := apiInd.Params.Metric.Metrics
		require.Len(t, metrics, 3)

		bm, err := metrics[0].AsSLOsTimesliceMetricBasicMetricWithField()
		require.NoError(t, err)
		require.NotNil(t, bm.Filter)
		assert.Equal(t, sumFilter, *bm.Filter)

		pm, err := metrics[1].AsSLOsTimesliceMetricPercentileMetric()
		require.NoError(t, err)
		assert.InDelta(t, 95.0, pm.Percentile, 1e-3)

		dm, err := metrics[2].AsSLOsTimesliceMetricDocCountMetric()
		require.NoError(t, err)
		require.NotNil(t, dm.Filter)
		assert.Equal(t, docFilter, *dm.Filter)
	})

	for _, agg := range []string{"last_value", "cardinality", "std_deviation"} {
		t.Run("maps "+agg+" aggregation as basic metric with field", func(t *testing.T) {
			m := tfModel{TimesliceMetricIndicator: []tfTimesliceMetricIndicator{{
				Index:          types.StringValue("metrics-*"),
				DataViewID:     types.StringNull(),
				TimestampField: types.StringValue("@timestamp"),
				Filter:         types.StringNull(),
				Metric: []tfTimesliceMetricDefinition{{
					Equation:   types.StringValue("A"),
					Comparator: types.StringValue("GT"),
					Threshold:  types.Float64Value(0),
					Metrics: []tfTimesliceMetricMetric{{
						Name:        types.StringValue("A"),
						Aggregation: types.StringValue(agg),
						Field:       types.StringValue("some.field"),
						Percentile:  types.Float64Null(),
						Filter:      types.StringNull(),
					}},
				}},
			}}}

			ok, ind, diags := m.timesliceMetricIndicatorToAPI()
			require.True(t, ok)
			require.False(t, diags.HasError())

			apiInd, err := ind.AsSLOsIndicatorPropertiesTimesliceMetric()
			require.NoError(t, err)

			metrics := apiInd.Params.Metric.Metrics
			require.Len(t, metrics, 1)
			bm, err := metrics[0].AsSLOsTimesliceMetricBasicMetricWithField()
			require.NoError(t, err)
			assert.Equal(t, agg, string(bm.Aggregation))
			assert.Equal(t, "some.field", bm.Field)
		})
	}

	t.Run("emits error on unsupported aggregation", func(t *testing.T) {
		m := tfModel{TimesliceMetricIndicator: []tfTimesliceMetricIndicator{{
			Index:          types.StringValue("metrics-*"),
			DataViewID:     types.StringNull(),
			TimestampField: types.StringValue("@timestamp"),
			Filter:         types.StringNull(),
			Metric: []tfTimesliceMetricDefinition{{
				Equation:   types.StringValue("a"),
				Comparator: types.StringValue("GT"),
				Threshold:  types.Float64Value(1),
				Metrics: []tfTimesliceMetricMetric{{
					Name:        types.StringValue("a"),
					Aggregation: types.StringValue("median"),
					Field:       types.StringValue("foo"),
					Percentile:  types.Float64Null(),
					Filter:      types.StringNull(),
				}},
			}},
		}}}

		ok, _, diags := m.timesliceMetricIndicatorToAPI()
		require.True(t, ok)
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), "unsupported aggregation")
	})
}

func TestTimesliceMetricIndicator_PopulateFromAPI(t *testing.T) {
	sumFilter := testTimesliceSumFilter
	docFilter := testTimesliceDocFilter
	dvID := testTimesliceDvID

	t.Run("maps metric variants including filters", func(t *testing.T) {
		api := buildTimesliceAPI(&dvID, nil, []kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item{
			buildBasicMetricItem(t, "a", "sum", "foo", &sumFilter),
			buildPercentileItem(t, "p95", "percentile", "latency", 95, nil),
			buildDocCountItem(t, "c", &docFilter),
		})

		var m tfModel
		diags := m.populateFromTimesliceMetricIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.TimesliceMetricIndicator, 1)

		ind := m.TimesliceMetricIndicator[0]
		assert.Equal(t, testTimesliceDvID, ind.DataViewID.ValueString())
		require.Len(t, ind.Metric, 1)
		require.Len(t, ind.Metric[0].Metrics, 3)

		assert.Equal(t, "sum", ind.Metric[0].Metrics[0].Aggregation.ValueString())
		assert.Equal(t, sumFilter, ind.Metric[0].Metrics[0].Filter.ValueString())

		assert.Equal(t, "percentile", ind.Metric[0].Metrics[1].Aggregation.ValueString())
		assert.True(t, ind.Metric[0].Metrics[1].Filter.IsNull())
		assert.InDelta(t, 95.0, ind.Metric[0].Metrics[1].Percentile.ValueFloat64(), 1e-9)

		assert.Equal(t, "doc_count", ind.Metric[0].Metrics[2].Aggregation.ValueString())
		assert.Equal(t, docFilter, ind.Metric[0].Metrics[2].Filter.ValueString())
	})

	t.Run("sets optional fields to null when not present", func(t *testing.T) {
		api := buildTimesliceAPI(nil, nil, []kbapi.SLOsIndicatorPropertiesTimesliceMetric_Params_Metric_Metrics_Item{
			buildBasicMetricItem(t, "a", "sum", "foo", nil),
		})

		var m tfModel
		diags := m.populateFromTimesliceMetricIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.TimesliceMetricIndicator, 1)

		ind := m.TimesliceMetricIndicator[0]
		assert.True(t, ind.DataViewID.IsNull())
		assert.True(t, ind.Filter.IsNull())
		assert.True(t, ind.Metric[0].Metrics[0].Filter.IsNull())
	})
}
