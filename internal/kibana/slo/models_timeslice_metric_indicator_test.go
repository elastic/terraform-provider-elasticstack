package slo

import (
	"testing"

	generatedslo "github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		m := tfModel{TimesliceMetricIndicator: []tfTimesliceMetricIndicator{{
			Index:          types.StringValue("metrics-*"),
			DataViewID:     types.StringValue("dv-1"),
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
						Filter:      types.StringValue("status:200"),
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
						Filter:      types.StringValue("labels.env:prod"),
					},
				},
			}},
		}}}

		ok, ind, diags := m.timesliceMetricIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())
		require.NotNil(t, ind.IndicatorPropertiesTimesliceMetric)

		metrics := ind.IndicatorPropertiesTimesliceMetric.Params.Metric.Metrics
		require.Len(t, metrics, 3)
		require.NotNil(t, metrics[0].TimesliceMetricBasicMetricWithField)
		require.NotNil(t, metrics[0].TimesliceMetricBasicMetricWithField.Filter)
		assert.Equal(t, "status:200", *metrics[0].TimesliceMetricBasicMetricWithField.Filter)

		require.NotNil(t, metrics[1].TimesliceMetricPercentileMetric)
		assert.Equal(t, 95.0, metrics[1].TimesliceMetricPercentileMetric.Percentile)

		require.NotNil(t, metrics[2].TimesliceMetricDocCountMetric)
		require.NotNil(t, metrics[2].TimesliceMetricDocCountMetric.Filter)
		assert.Equal(t, "labels.env:prod", *metrics[2].TimesliceMetricDocCountMetric.Filter)
	})

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
	t.Run("maps metric variants including filters", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesTimesliceMetric{
			Params: generatedslo.IndicatorPropertiesTimesliceMetricParams{
				Index:          "metrics-*",
				DataViewId:     strPtr("dv-1"),
				TimestampField: "@timestamp",
				Filter:         nil,
				Metric: generatedslo.IndicatorPropertiesTimesliceMetricParamsMetric{
					Equation:   "a",
					Comparator: "GT",
					Threshold:  1.23,
					Metrics: []generatedslo.IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner{
						{
							TimesliceMetricBasicMetricWithField: &generatedslo.TimesliceMetricBasicMetricWithField{
								Name:        "a",
								Aggregation: "sum",
								Field:       "foo",
								Filter:      strPtr("status:200"),
							},
						},
						{
							TimesliceMetricPercentileMetric: &generatedslo.TimesliceMetricPercentileMetric{
								Name:        "p95",
								Aggregation: "percentile",
								Field:       "latency",
								Percentile:  95,
								Filter:      nil,
							},
						},
						{
							TimesliceMetricDocCountMetric: &generatedslo.TimesliceMetricDocCountMetric{
								Name:        "c",
								Aggregation: "doc_count",
								Filter:      strPtr("labels.env:prod"),
							},
						},
					},
				},
			},
		}

		var m tfModel
		diags := m.populateFromTimesliceMetricIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.TimesliceMetricIndicator, 1)

		ind := m.TimesliceMetricIndicator[0]
		assert.Equal(t, "dv-1", ind.DataViewID.ValueString())
		require.Len(t, ind.Metric, 1)
		require.Len(t, ind.Metric[0].Metrics, 3)

		assert.Equal(t, "sum", ind.Metric[0].Metrics[0].Aggregation.ValueString())
		assert.Equal(t, "status:200", ind.Metric[0].Metrics[0].Filter.ValueString())

		assert.Equal(t, "percentile", ind.Metric[0].Metrics[1].Aggregation.ValueString())
		assert.True(t, ind.Metric[0].Metrics[1].Filter.IsNull())
		assert.Equal(t, 95.0, ind.Metric[0].Metrics[1].Percentile.ValueFloat64())

		assert.Equal(t, "doc_count", ind.Metric[0].Metrics[2].Aggregation.ValueString())
		assert.Equal(t, "labels.env:prod", ind.Metric[0].Metrics[2].Filter.ValueString())
	})

	t.Run("sets optional fields to null when not present", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesTimesliceMetric{
			Params: generatedslo.IndicatorPropertiesTimesliceMetricParams{
				Index:          "metrics-*",
				DataViewId:     nil,
				TimestampField: "@timestamp",
				Filter:         nil,
				Metric: generatedslo.IndicatorPropertiesTimesliceMetricParamsMetric{
					Equation:   "a",
					Comparator: "GT",
					Threshold:  1,
					Metrics: []generatedslo.IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner{
						{
							TimesliceMetricBasicMetricWithField: &generatedslo.TimesliceMetricBasicMetricWithField{
								Name:        "a",
								Aggregation: "sum",
								Field:       "foo",
								Filter:      nil,
							},
						},
					},
				},
			},
		}

		var m tfModel
		diags := m.populateFromTimesliceMetricIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.TimesliceMetricIndicator, 1)

		ind := m.TimesliceMetricIndicator[0]
		assert.True(t, ind.DataViewID.IsNull())
		assert.True(t, ind.Filter.IsNull())
		assert.True(t, ind.Metric[0].Metrics[0].Filter.IsNull())
	})

	t.Run("returns empty diagnostics when api is nil", func(t *testing.T) {
		var m tfModel
		diags := m.populateFromTimesliceMetricIndicator(nil)
		require.False(t, diags.HasError())
		assert.Nil(t, m.TimesliceMetricIndicator)
	})
}
