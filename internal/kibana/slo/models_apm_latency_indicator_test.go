package slo

import (
	"testing"

	generatedslo "github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApmLatencyIndicator_ToAPI(t *testing.T) {
	t.Run("returns ok=false when not configured", func(t *testing.T) {
		m := tfModel{}
		ok, _, diags := m.apmLatencyIndicatorToAPI()
		require.False(t, ok)
		require.False(t, diags.HasError())
	})

	t.Run("maps all fields including threshold", func(t *testing.T) {
		m := tfModel{ApmLatencyIndicator: []tfApmLatencyIndicator{{
			Index:           types.StringValue("apm-*"),
			Filter:          types.StringValue("status:200"),
			Service:         types.StringValue("svc"),
			Environment:     types.StringValue("prod"),
			TransactionType: types.StringValue("request"),
			TransactionName: types.StringValue("GET /"),
			Threshold:       types.Int64Value(500),
		}}}

		ok, ind, diags := m.apmLatencyIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())
		require.NotNil(t, ind.IndicatorPropertiesApmLatency)

		assert.Equal(t, indicatorAddressToType["apm_latency_indicator"], ind.IndicatorPropertiesApmLatency.Type)
		assert.Equal(t, 500.0, ind.IndicatorPropertiesApmLatency.Params.Threshold)
		assert.Equal(t, "svc", ind.IndicatorPropertiesApmLatency.Params.Service)
		assert.Equal(t, "apm-*", ind.IndicatorPropertiesApmLatency.Params.Index)
		require.NotNil(t, ind.IndicatorPropertiesApmLatency.Params.Filter)
		assert.Equal(t, "status:200", *ind.IndicatorPropertiesApmLatency.Params.Filter)
	})

	t.Run("omits filter when null", func(t *testing.T) {
		m := tfModel{ApmLatencyIndicator: []tfApmLatencyIndicator{{
			Index:           types.StringValue("apm-*"),
			Filter:          types.StringNull(),
			Service:         types.StringValue("svc"),
			Environment:     types.StringValue("prod"),
			TransactionType: types.StringValue("request"),
			TransactionName: types.StringValue("GET /"),
			Threshold:       types.Int64Value(500),
		}}}

		ok, ind, diags := m.apmLatencyIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())
		require.NotNil(t, ind.IndicatorPropertiesApmLatency)
		assert.Nil(t, ind.IndicatorPropertiesApmLatency.Params.Filter)
	})
}

func TestApmLatencyIndicator_PopulateFromAPI(t *testing.T) {
	t.Run("maps all fields including threshold", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesApmLatency{
			Params: generatedslo.IndicatorPropertiesApmLatencyParams{
				Service:         "svc",
				Environment:     "prod",
				TransactionType: "request",
				TransactionName: "GET /",
				Index:           "apm-*",
				Threshold:       500,
				Filter:          strPtr("status:200"),
			},
		}

		var m tfModel
		diags := m.populateFromApmLatencyIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.ApmLatencyIndicator, 1)

		ind := m.ApmLatencyIndicator[0]
		assert.Equal(t, int64(500), ind.Threshold.ValueInt64())
		assert.Equal(t, "svc", ind.Service.ValueString())
		assert.Equal(t, "apm-*", ind.Index.ValueString())
		assert.Equal(t, "status:200", ind.Filter.ValueString())
	})

	t.Run("sets filter to null when not present", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesApmLatency{
			Params: generatedslo.IndicatorPropertiesApmLatencyParams{
				Service:         "svc",
				Environment:     "prod",
				TransactionType: "request",
				TransactionName: "GET /",
				Index:           "apm-*",
				Threshold:       500,
				Filter:          nil,
			},
		}

		var m tfModel
		diags := m.populateFromApmLatencyIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.ApmLatencyIndicator, 1)
		assert.True(t, m.ApmLatencyIndicator[0].Filter.IsNull())
	})

	t.Run("returns empty diagnostics when api is nil", func(t *testing.T) {
		var m tfModel
		diags := m.populateFromApmLatencyIndicator(nil)
		require.False(t, diags.HasError())
		assert.Nil(t, m.ApmLatencyIndicator)
	})
}
