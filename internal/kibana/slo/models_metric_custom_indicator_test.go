package slo

import (
	"testing"

	generatedslo "github.com/elastic/terraform-provider-elasticstack/generated/slo"
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
					Filter:      types.StringValue("status:200"),
				}},
			}},
		}}}

		ok, ind, diags := m.metricCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())
		require.NotNil(t, ind.IndicatorPropertiesCustomMetric)

		params := ind.IndicatorPropertiesCustomMetric.Params
		require.NotNil(t, params.DataViewId)
		assert.Equal(t, "dv-1", *params.DataViewId)
		require.NotNil(t, params.Filter)
		assert.Equal(t, "labels.env:prod", *params.Filter)

		require.Len(t, params.Good.Metrics, 1)
		assert.Equal(t, "a", params.Good.Metrics[0].Name)
		assert.Nil(t, params.Good.Metrics[0].Filter)
		assert.Equal(t, "a / b", params.Good.Equation)

		require.Len(t, params.Total.Metrics, 1)
		assert.Equal(t, "c", params.Total.Metrics[0].Name)
		require.NotNil(t, params.Total.Metrics[0].Filter)
		assert.Equal(t, "status:200", *params.Total.Metrics[0].Filter)
		assert.Equal(t, "c", params.Total.Equation)
	})
}

func TestMetricCustomIndicator_PopulateFromAPI(t *testing.T) {
	t.Run("maps equations, metrics and optional pointers", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesCustomMetric{
			Params: generatedslo.IndicatorPropertiesCustomMetricParams{
				Index:          "metrics-*",
				DataViewId:     strPtr("dv-1"),
				Filter:         strPtr("labels.env:prod"),
				TimestampField: "@timestamp",
				Good: generatedslo.IndicatorPropertiesCustomMetricParamsGood{
					Equation: "a / b",
					Metrics: []generatedslo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner{{
						Name:        "a",
						Aggregation: "sum",
						Field:       "good",
						Filter:      nil,
					}},
				},
				Total: generatedslo.IndicatorPropertiesCustomMetricParamsTotal{
					Equation: "c",
					Metrics: []generatedslo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner{{
						Name:        "c",
						Aggregation: "sum",
						Field:       "total",
						Filter:      strPtr("status:200"),
					}},
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
		assert.Equal(t, "status:200", ind.Total[0].Metrics[0].Filter.ValueString())
	})

	t.Run("sets optional fields to null when not present", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesCustomMetric{
			Params: generatedslo.IndicatorPropertiesCustomMetricParams{
				Index:          "metrics-*",
				DataViewId:     nil,
				Filter:         nil,
				TimestampField: "@timestamp",
				Good: generatedslo.IndicatorPropertiesCustomMetricParamsGood{
					Equation: "a",
					Metrics:  []generatedslo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner{},
				},
				Total: generatedslo.IndicatorPropertiesCustomMetricParamsTotal{
					Equation: "b",
					Metrics:  []generatedslo.IndicatorPropertiesCustomMetricParamsGoodMetricsInner{},
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

	t.Run("returns empty diagnostics when api is nil", func(t *testing.T) {
		var m tfModel
		diags := m.populateFromMetricCustomIndicator(nil)
		require.False(t, diags.HasError())
		assert.Nil(t, m.MetricCustomIndicator)
	})
}
