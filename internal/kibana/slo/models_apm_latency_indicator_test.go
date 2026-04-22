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

		apiInd, err := ind.AsSLOsIndicatorPropertiesApmLatency()
		require.NoError(t, err)
		assert.Equal(t, indicatorAddressToType["apm_latency_indicator"], apiInd.Type)
		assert.InDelta(t, 500.0, apiInd.Params.Threshold, 1e-3)
		assert.Equal(t, "svc", apiInd.Params.Service)
		assert.Equal(t, "apm-*", apiInd.Params.Index)
		require.NotNil(t, apiInd.Params.Filter)
		assert.Equal(t, "status:200", *apiInd.Params.Filter)
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
		apiInd, err := ind.AsSLOsIndicatorPropertiesApmLatency()
		require.NoError(t, err)
		assert.Nil(t, apiInd.Params.Filter)
	})
}

func TestApmLatencyIndicator_PopulateFromAPI(t *testing.T) {
	t.Run("maps all fields including threshold", func(t *testing.T) {
		filter := testTimesliceSumFilter
		api := kbapi.SLOsIndicatorPropertiesApmLatency{
			Params: struct {
				Environment     string  `json:"environment"`
				Filter          *string `json:"filter,omitempty"`
				Index           string  `json:"index"`
				Service         string  `json:"service"`
				Threshold       float64 `json:"threshold"`
				TransactionName string  `json:"transactionName"`
				TransactionType string  `json:"transactionType"`
			}{
				Service:         "svc",
				Environment:     "prod",
				TransactionType: "request",
				TransactionName: "GET /",
				Index:           "apm-*",
				Threshold:       500,
				Filter:          &filter,
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
		api := kbapi.SLOsIndicatorPropertiesApmLatency{
			Params: struct {
				Environment     string  `json:"environment"`
				Filter          *string `json:"filter,omitempty"`
				Index           string  `json:"index"`
				Service         string  `json:"service"`
				Threshold       float64 `json:"threshold"`
				TransactionName string  `json:"transactionName"`
				TransactionType string  `json:"transactionType"`
			}{
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
}
