package slo

import (
	"testing"

	generatedslo "github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKqlCustomIndicator_ToAPI(t *testing.T) {
	t.Run("returns ok=false when not configured", func(t *testing.T) {
		m := tfModel{}
		ok, _, diags := m.kqlCustomIndicatorToAPI()
		require.False(t, ok)
		require.False(t, diags.HasError())
	})

	t.Run("maps all fields with optional data_view_id", func(t *testing.T) {
		m := tfModel{KqlCustomIndicator: []tfKqlCustomIndicator{{
			Index:          types.StringValue("logs-*"),
			DataViewID:     types.StringValue("dv-123"),
			Filter:         types.StringValue("service.name:foo"),
			Good:           types.StringValue("status:200"),
			Total:          types.StringValue("*"),
			TimestampField: types.StringValue("@timestamp"),
		}}}

		ok, ind, diags := m.kqlCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())
		require.NotNil(t, ind.IndicatorPropertiesCustomKql)

		params := ind.IndicatorPropertiesCustomKql.Params
		assert.Equal(t, "logs-*", params.Index)
		require.NotNil(t, params.DataViewId)
		assert.Equal(t, "dv-123", *params.DataViewId)
		require.NotNil(t, params.Filter)
		require.NotNil(t, params.Filter.String)
		assert.Equal(t, "service.name:foo", *params.Filter.String)
		require.NotNil(t, params.Good.String)
		assert.Equal(t, "status:200", *params.Good.String)
		require.NotNil(t, params.Total.String)
		assert.Equal(t, "*", *params.Total.String)
		assert.Equal(t, "@timestamp", params.TimestampField)
	})

	t.Run("handles unknown values by omitting pointers", func(t *testing.T) {
		m := tfModel{KqlCustomIndicator: []tfKqlCustomIndicator{{
			Index:          types.StringValue("logs-*"),
			DataViewID:     types.StringNull(),
			Filter:         types.StringUnknown(),
			Good:           types.StringUnknown(),
			Total:          types.StringNull(),
			TimestampField: types.StringValue("@timestamp"),
		}}}

		ok, ind, diags := m.kqlCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())
		require.NotNil(t, ind.IndicatorPropertiesCustomKql)

		params := ind.IndicatorPropertiesCustomKql.Params
		assert.Equal(t, "logs-*", params.Index)
		assert.Nil(t, params.DataViewId)
		assert.Nil(t, params.Filter)
		assert.Nil(t, params.Good.String)
		assert.Nil(t, params.Total.String)
		assert.Equal(t, "@timestamp", params.TimestampField)
	})
}

func TestKqlCustomIndicator_PopulateFromAPI(t *testing.T) {
	t.Run("maps all optional fields", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesCustomKql{
			Params: generatedslo.IndicatorPropertiesCustomKqlParams{
				Index:          "logs-*",
				DataViewId:     strPtr("dv-123"),
				Filter:         &generatedslo.KqlWithFilters{String: strPtr("service.name:foo")},
				Good:           generatedslo.KqlWithFiltersGood{String: strPtr("status:200")},
				Total:          generatedslo.KqlWithFiltersTotal{String: strPtr("*")},
				TimestampField: "@timestamp",
			},
		}

		var m tfModel
		diags := m.populateFromKqlCustomIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.KqlCustomIndicator, 1)

		ind := m.KqlCustomIndicator[0]
		assert.Equal(t, "logs-*", ind.Index.ValueString())
		assert.Equal(t, "dv-123", ind.DataViewID.ValueString())
		assert.Equal(t, "service.name:foo", ind.Filter.ValueString())
		assert.Equal(t, "status:200", ind.Good.ValueString())
		assert.Equal(t, "*", ind.Total.ValueString())
		assert.Equal(t, "@timestamp", ind.TimestampField.ValueString())
	})

	t.Run("sets optional fields to null when not present", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesCustomKql{
			Params: generatedslo.IndicatorPropertiesCustomKqlParams{
				Index:          "logs-*",
				DataViewId:     nil,
				Filter:         nil,
				Good:           generatedslo.KqlWithFiltersGood{String: nil},
				Total:          generatedslo.KqlWithFiltersTotal{String: nil},
				TimestampField: "@timestamp",
			},
		}

		var m tfModel
		diags := m.populateFromKqlCustomIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.KqlCustomIndicator, 1)

		ind := m.KqlCustomIndicator[0]
		assert.True(t, ind.DataViewID.IsNull())
		assert.True(t, ind.Filter.IsNull())
		assert.True(t, ind.Good.IsNull())
		assert.True(t, ind.Total.IsNull())
	})

	t.Run("returns empty diagnostics when api is nil", func(t *testing.T) {
		var m tfModel
		diags := m.populateFromKqlCustomIndicator(nil)
		require.False(t, diags.HasError())
		assert.Nil(t, m.KqlCustomIndicator)
	})
}
