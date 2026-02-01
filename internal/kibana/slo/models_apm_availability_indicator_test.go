package slo

import (
	"testing"

	generatedslo "github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApmAvailabilityIndicator_ToAPI(t *testing.T) {
	t.Run("returns ok=false when not configured", func(t *testing.T) {
		m := tfModel{}
		ok, _, diags := m.apmAvailabilityIndicatorToAPI()
		require.False(t, ok)
		require.False(t, diags.HasError())
	})

	t.Run("maps all fields with optional filter", func(t *testing.T) {
		m := tfModel{ApmAvailabilityIndicator: []tfApmAvailabilityIndicator{{
			Index:           types.StringValue("apm-*"),
			Filter:          types.StringValue("service.name:foo"),
			Service:         types.StringValue("svc"),
			Environment:     types.StringValue("prod"),
			TransactionType: types.StringValue("request"),
			TransactionName: types.StringValue("GET /"),
		}}}

		ok, ind, diags := m.apmAvailabilityIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())

		require.NotNil(t, ind.IndicatorPropertiesApmAvailability)
		assert.Equal(t, indicatorAddressToType["apm_availability_indicator"], ind.IndicatorPropertiesApmAvailability.Type)
		assert.Equal(t, "svc", ind.IndicatorPropertiesApmAvailability.Params.Service)
		assert.Equal(t, "prod", ind.IndicatorPropertiesApmAvailability.Params.Environment)
		assert.Equal(t, "request", ind.IndicatorPropertiesApmAvailability.Params.TransactionType)
		assert.Equal(t, "GET /", ind.IndicatorPropertiesApmAvailability.Params.TransactionName)
		assert.Equal(t, "apm-*", ind.IndicatorPropertiesApmAvailability.Params.Index)
		require.NotNil(t, ind.IndicatorPropertiesApmAvailability.Params.Filter)
		assert.Equal(t, "service.name:foo", *ind.IndicatorPropertiesApmAvailability.Params.Filter)
	})

	t.Run("omits filter when unknown", func(t *testing.T) {
		m := tfModel{ApmAvailabilityIndicator: []tfApmAvailabilityIndicator{{
			Index:           types.StringValue("apm-*"),
			Filter:          types.StringUnknown(),
			Service:         types.StringValue("svc"),
			Environment:     types.StringValue("prod"),
			TransactionType: types.StringValue("request"),
			TransactionName: types.StringValue("GET /"),
		}}}

		ok, ind, diags := m.apmAvailabilityIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())
		require.NotNil(t, ind.IndicatorPropertiesApmAvailability)
		assert.Nil(t, ind.IndicatorPropertiesApmAvailability.Params.Filter)
	})

	t.Run("omits filter when null", func(t *testing.T) {
		m := tfModel{ApmAvailabilityIndicator: []tfApmAvailabilityIndicator{{
			Index:           types.StringValue("apm-*"),
			Filter:          types.StringNull(),
			Service:         types.StringValue("svc"),
			Environment:     types.StringValue("prod"),
			TransactionType: types.StringValue("request"),
			TransactionName: types.StringValue("GET /"),
		}}}

		ok, ind, diags := m.apmAvailabilityIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())
		require.NotNil(t, ind.IndicatorPropertiesApmAvailability)
		assert.Nil(t, ind.IndicatorPropertiesApmAvailability.Params.Filter)
	})
}

func TestApmAvailabilityIndicator_PopulateFromAPI(t *testing.T) {
	t.Run("maps all fields with optional filter", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesApmAvailability{
			Params: generatedslo.IndicatorPropertiesApmAvailabilityParams{
				Service:         "svc",
				Environment:     "prod",
				TransactionType: "request",
				TransactionName: "GET /",
				Index:           "apm-*",
				Filter:          strPtr("service.name:foo"),
			},
		}

		var m tfModel
		diags := m.populateFromApmAvailabilityIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.ApmAvailabilityIndicator, 1)

		ind := m.ApmAvailabilityIndicator[0]
		assert.Equal(t, "svc", ind.Service.ValueString())
		assert.Equal(t, "prod", ind.Environment.ValueString())
		assert.Equal(t, "request", ind.TransactionType.ValueString())
		assert.Equal(t, "GET /", ind.TransactionName.ValueString())
		assert.Equal(t, "apm-*", ind.Index.ValueString())
		assert.Equal(t, "service.name:foo", ind.Filter.ValueString())
	})

	t.Run("sets filter to null when not present", func(t *testing.T) {
		api := &generatedslo.IndicatorPropertiesApmAvailability{
			Params: generatedslo.IndicatorPropertiesApmAvailabilityParams{
				Service:         "svc",
				Environment:     "prod",
				TransactionType: "request",
				TransactionName: "GET /",
				Index:           "apm-*",
				Filter:          nil,
			},
		}

		var m tfModel
		diags := m.populateFromApmAvailabilityIndicator(api)
		require.False(t, diags.HasError())
		require.Len(t, m.ApmAvailabilityIndicator, 1)
		assert.True(t, m.ApmAvailabilityIndicator[0].Filter.IsNull())
	})

	t.Run("returns empty diagnostics when api is nil", func(t *testing.T) {
		var m tfModel
		diags := m.populateFromApmAvailabilityIndicator(nil)
		require.False(t, diags.HasError())
		assert.Nil(t, m.ApmAvailabilityIndicator)
	})
}
