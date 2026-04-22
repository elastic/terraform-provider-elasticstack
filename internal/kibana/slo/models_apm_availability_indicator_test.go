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

		apiInd, err := ind.AsSLOsIndicatorPropertiesApmAvailability()
		require.NoError(t, err)
		assert.Equal(t, indicatorAddressToType["apm_availability_indicator"], apiInd.Type)
		assert.Equal(t, "svc", apiInd.Params.Service)
		assert.Equal(t, "prod", apiInd.Params.Environment)
		assert.Equal(t, "request", apiInd.Params.TransactionType)
		assert.Equal(t, "GET /", apiInd.Params.TransactionName)
		assert.Equal(t, "apm-*", apiInd.Params.Index)
		require.NotNil(t, apiInd.Params.Filter)
		assert.Equal(t, "service.name:foo", *apiInd.Params.Filter)
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
		apiInd, err := ind.AsSLOsIndicatorPropertiesApmAvailability()
		require.NoError(t, err)
		assert.Nil(t, apiInd.Params.Filter)
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
		apiInd, err := ind.AsSLOsIndicatorPropertiesApmAvailability()
		require.NoError(t, err)
		assert.Nil(t, apiInd.Params.Filter)
	})
}

func TestApmAvailabilityIndicator_PopulateFromAPI(t *testing.T) {
	t.Run("maps all fields with optional filter", func(t *testing.T) {
		filter := "service.name:foo"
		api := kbapi.SLOsIndicatorPropertiesApmAvailability{
			Params: struct {
				Environment     string  `json:"environment"`
				Filter          *string `json:"filter,omitempty"`
				Index           string  `json:"index"`
				Service         string  `json:"service"`
				TransactionName string  `json:"transactionName"`
				TransactionType string  `json:"transactionType"`
			}{
				Service:         "svc",
				Environment:     "prod",
				TransactionType: "request",
				TransactionName: "GET /",
				Index:           "apm-*",
				Filter:          &filter,
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
		api := kbapi.SLOsIndicatorPropertiesApmAvailability{
			Params: struct {
				Environment     string  `json:"environment"`
				Filter          *string `json:"filter,omitempty"`
				Index           string  `json:"index"`
				Service         string  `json:"service"`
				TransactionName string  `json:"transactionName"`
				TransactionType string  `json:"transactionType"`
			}{
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
}
