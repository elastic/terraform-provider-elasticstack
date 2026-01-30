package slo

import (
	"testing"

	generatedslo "github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistogramCustomIndicator_ToAPI(t *testing.T) {
	t.Run("returns ok=false when not configured", func(t *testing.T) {
		m := tfModel{}
		ok, _, diags := m.histogramCustomIndicatorToAPI()
		require.False(t, ok)
		require.False(t, diags.HasError())
	})

	t.Run("emits error when good/total cardinality invalid", func(t *testing.T) {
		m := tfModel{HistogramCustomIndicator: []tfHistogramCustomIndicator{{
			Index:          types.StringValue("logs-*"),
			DataViewID:     types.StringNull(),
			Filter:         types.StringNull(),
			TimestampField: types.StringValue("@timestamp"),
			Good:           nil,
			Total:          nil,
		}}}

		ok, _, diags := m.histogramCustomIndicatorToAPI()
		require.True(t, ok)
		require.True(t, diags.HasError())
	})

	t.Run("maps ranges and optional pointers", func(t *testing.T) {
		m := tfModel{HistogramCustomIndicator: []tfHistogramCustomIndicator{{
			Index:          types.StringValue("logs-*"),
			DataViewID:     types.StringValue("dv-1"),
			Filter:         types.StringNull(),
			TimestampField: types.StringValue("@timestamp"),
			Good: []tfHistogramRange{{
				Aggregation: types.StringValue("sum"),
				Field:       types.StringValue("latency"),
				Filter:      types.StringValue("status:200"),
				From:        types.Float64Value(0),
				To:          types.Float64Null(),
			}},
			Total: []tfHistogramRange{{
				Aggregation: types.StringValue("sum"),
				Field:       types.StringValue("latency"),
				Filter:      types.StringNull(),
				From:        types.Float64Null(),
				To:          types.Float64Value(10),
			}},
		}}}

		ok, ind, diags := m.histogramCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())
		require.NotNil(t, ind.IndicatorPropertiesHistogram)

		params := ind.IndicatorPropertiesHistogram.Params
		require.NotNil(t, params.DataViewId)
		assert.Equal(t, "dv-1", *params.DataViewId)
		require.Nil(t, params.Filter)

		require.NotNil(t, params.Good.Filter)
		assert.Equal(t, "status:200", *params.Good.Filter)
		require.NotNil(t, params.Good.From)
		assert.Equal(t, 0.0, *params.Good.From)
		assert.Nil(t, params.Good.To)

		assert.Nil(t, params.Total.Filter)
		assert.Nil(t, params.Total.From)
		require.NotNil(t, params.Total.To)
		assert.Equal(t, 10.0, *params.Total.To)
	})
}

func TestHistogramCustomIndicator_PopulateFromAPI(t *testing.T) {
	t.Run("maps ranges and optional pointers", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesHistogram{
			Params: generatedslo.IndicatorPropertiesHistogramParams{
				Index:          "logs-*",
				DataViewId:     strPtr("dv-1"),
				Filter:         nil,
				TimestampField: "@timestamp",
				Good: generatedslo.IndicatorPropertiesHistogramParamsGood{
					Aggregation: "sum",
					Field:       "latency",
					Filter:      strPtr("status:200"),
					From:        floatPtr(0),
					To:          nil,
				},
				Total: generatedslo.IndicatorPropertiesHistogramParamsTotal{
					Aggregation: "sum",
					Field:       "latency",
					Filter:      nil,
					From:        nil,
					To:          floatPtr(10),
				},
			},
		}

		var m tfModel
		diags := m.populateFromHistogramCustomIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.HistogramCustomIndicator, 1)

		ind := m.HistogramCustomIndicator[0]
		assert.Equal(t, "dv-1", ind.DataViewID.ValueString())
		require.Len(t, ind.Good, 1)
		require.Len(t, ind.Total, 1)
		assert.Equal(t, "status:200", ind.Good[0].Filter.ValueString())
		assert.Equal(t, 0.0, ind.Good[0].From.ValueFloat64())
		assert.True(t, ind.Good[0].To.IsNull())
		assert.True(t, ind.Total[0].From.IsNull())
		assert.Equal(t, 10.0, ind.Total[0].To.ValueFloat64())
	})

	t.Run("sets optional fields to null when not present", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesHistogram{
			Params: generatedslo.IndicatorPropertiesHistogramParams{
				Index:          "logs-*",
				DataViewId:     nil,
				Filter:         nil,
				TimestampField: "@timestamp",
				Good: generatedslo.IndicatorPropertiesHistogramParamsGood{
					Aggregation: "sum",
					Field:       "latency",
					Filter:      nil,
					From:        nil,
					To:          nil,
				},
				Total: generatedslo.IndicatorPropertiesHistogramParamsTotal{
					Aggregation: "sum",
					Field:       "latency",
					Filter:      nil,
					From:        nil,
					To:          nil,
				},
			},
		}

		var m tfModel
		diags := m.populateFromHistogramCustomIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.HistogramCustomIndicator, 1)

		ind := m.HistogramCustomIndicator[0]
		assert.True(t, ind.DataViewID.IsNull())
		assert.True(t, ind.Filter.IsNull())
		assert.True(t, ind.Good[0].Filter.IsNull())
		assert.True(t, ind.Good[0].From.IsNull())
		assert.True(t, ind.Good[0].To.IsNull())
	})

	t.Run("returns empty diagnostics when api is nil", func(t *testing.T) {
		var m tfModel
		diags := m.populateFromHistogramCustomIndicator(nil)
		require.False(t, diags.HasError())
		assert.Nil(t, m.HistogramCustomIndicator)
	})
}
