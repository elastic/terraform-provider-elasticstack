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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
			FilterKql:      types.ObjectNull(tfKqlKqlObjectAttrTypes),
			Good:           types.StringValue("status:200"),
			GoodKql:        types.ObjectNull(tfKqlKqlObjectAttrTypes),
			Total:          types.StringValue("*"),
			TotalKql:       types.ObjectNull(tfKqlKqlObjectAttrTypes),
			TimestampField: types.StringValue("@timestamp"),
		}}}

		ok, ind, diags := m.kqlCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())

		apiInd, err := ind.AsSLOsIndicatorPropertiesCustomKql()
		require.NoError(t, err)

		params := apiInd.Params
		assert.Equal(t, "logs-*", params.Index)
		require.NotNil(t, params.DataViewId)
		assert.Equal(t, "dv-123", *params.DataViewId)
		require.NotNil(t, params.Filter)
		filterStr, ferr := params.Filter.AsSLOsKqlWithFilters0()
		require.NoError(t, ferr)
		assert.Equal(t, "service.name:foo", filterStr)
		goodStr, gerr := params.Good.AsSLOsKqlWithFiltersGood0()
		require.NoError(t, gerr)
		assert.Equal(t, "status:200", goodStr)
		totalStr, terr := params.Total.AsSLOsKqlWithFiltersTotal0()
		require.NoError(t, terr)
		assert.Equal(t, "*", totalStr)
		assert.Equal(t, "@timestamp", params.TimestampField)
	})

	t.Run("handles unknown/null values by defaulting to empty strings for required fields", func(t *testing.T) {
		m := tfModel{KqlCustomIndicator: []tfKqlCustomIndicator{{
			Index:          types.StringValue("logs-*"),
			DataViewID:     types.StringNull(),
			Filter:         types.StringUnknown(),
			FilterKql:      types.ObjectNull(tfKqlKqlObjectAttrTypes),
			Good:           types.StringUnknown(),
			GoodKql:        types.ObjectNull(tfKqlKqlObjectAttrTypes),
			Total:          types.StringNull(),
			TotalKql:       types.ObjectNull(tfKqlKqlObjectAttrTypes),
			TimestampField: types.StringValue("@timestamp"),
		}}}

		ok, ind, diags := m.kqlCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())

		apiInd, err := ind.AsSLOsIndicatorPropertiesCustomKql()
		require.NoError(t, err)

		params := apiInd.Params
		assert.Equal(t, "logs-*", params.Index)
		assert.Nil(t, params.DataViewId)
		assert.Nil(t, params.Filter)
		// Good and Total are required fields, so they default to empty strings
		goodStr, gerr := params.Good.AsSLOsKqlWithFiltersGood0()
		require.NoError(t, gerr)
		assert.Empty(t, goodStr)
		totalStr, terr := params.Total.AsSLOsKqlWithFiltersTotal0()
		require.NoError(t, terr)
		assert.Empty(t, totalStr)
		assert.Equal(t, "@timestamp", params.TimestampField)
	})

	t.Run("preserves empty strings when explicitly provided", func(t *testing.T) {
		m := tfModel{KqlCustomIndicator: []tfKqlCustomIndicator{{
			Index:          types.StringValue("logs-*"),
			FilterKql:      types.ObjectNull(tfKqlKqlObjectAttrTypes),
			Good:           types.StringValue(""),
			GoodKql:        types.ObjectNull(tfKqlKqlObjectAttrTypes),
			Total:          types.StringValue(""),
			TotalKql:       types.ObjectNull(tfKqlKqlObjectAttrTypes),
			TimestampField: types.StringValue("@timestamp"),
		}}}

		ok, ind, diags := m.kqlCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())

		apiInd, err := ind.AsSLOsIndicatorPropertiesCustomKql()
		require.NoError(t, err)

		goodStr, gerr := apiInd.Params.Good.AsSLOsKqlWithFiltersGood0()
		require.NoError(t, gerr)
		assert.Empty(t, goodStr)
		totalStr, terr := apiInd.Params.Total.AsSLOsKqlWithFiltersTotal0()
		require.NoError(t, terr)
		assert.Empty(t, totalStr)
	})

	t.Run("serializes good_kql object form with filters", func(t *testing.T) {
		q := jsontypes.NewNormalizedValue(`{"match_all":{}}`)
		row, d := types.ObjectValue(tfKqlFilterRowObjectType.AttrTypes, map[string]attr.Value{"query": q})
		require.False(t, d.HasError())
		list, d := types.ListValue(tfKqlFilterRowObjectType, []attr.Value{row})
		require.False(t, d.HasError())
		kqlObj, d := types.ObjectValue(tfKqlKqlObjectAttrTypes, map[string]attr.Value{
			"kql_query": types.StringValue("event.outcome: success"),
			"filters":   list,
		})
		require.False(t, d.HasError())

		m := tfModel{KqlCustomIndicator: []tfKqlCustomIndicator{{
			Index:          types.StringValue("logs-*"),
			FilterKql:      types.ObjectNull(tfKqlKqlObjectAttrTypes),
			Good:           types.StringNull(),
			GoodKql:        kqlObj,
			Total:          types.StringValue("*"),
			TotalKql:       types.ObjectNull(tfKqlKqlObjectAttrTypes),
			TimestampField: types.StringValue("@timestamp"),
		}}}

		ok, ind, diags := m.kqlCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, diags.HasError())

		apiInd, err := ind.AsSLOsIndicatorPropertiesCustomKql()
		require.NoError(t, err)
		g1, err := apiInd.Params.Good.AsSLOsKqlWithFiltersGood1()
		require.NoError(t, err)
		require.NotNil(t, g1.Filters)
		require.Len(t, *g1.Filters, 1)
		require.NotNil(t, g1.KqlQuery)
		assert.Equal(t, "event.outcome: success", *g1.KqlQuery)
	})
}

func TestKqlCustomIndicator_PopulateFromAPI(t *testing.T) {
	t.Run("maps all optional fields", func(t *testing.T) {
		dvID := "dv-123"
		var filter kbapi.SLOsKqlWithFilters
		require.NoError(t, filter.FromSLOsKqlWithFilters0("service.name:foo"))
		var good kbapi.SLOsKqlWithFiltersGood
		require.NoError(t, good.FromSLOsKqlWithFiltersGood0("status:200"))
		var total kbapi.SLOsKqlWithFiltersTotal
		require.NoError(t, total.FromSLOsKqlWithFiltersTotal0("*"))

		api := kbapi.SLOsIndicatorPropertiesCustomKql{
			Params: struct {
				DataViewId     *string                       `json:"dataViewId,omitempty"` //nolint:revive // var-naming: API struct field
				Filter         *kbapi.SLOsKqlWithFilters     `json:"filter,omitempty"`
				Good           kbapi.SLOsKqlWithFiltersGood  `json:"good"`
				Index          string                        `json:"index"`
				TimestampField string                        `json:"timestampField"`
				Total          kbapi.SLOsKqlWithFiltersTotal `json:"total"`
			}{
				Index:          "logs-*",
				DataViewId:     &dvID,
				Filter:         &filter,
				Good:           good,
				Total:          total,
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
		assert.True(t, ind.FilterKql.IsNull())
		assert.True(t, ind.GoodKql.IsNull())
		assert.True(t, ind.TotalKql.IsNull())
	})

	t.Run("sets optional fields to null when not present", func(t *testing.T) {
		var emptyGood kbapi.SLOsKqlWithFiltersGood
		var emptyTotal kbapi.SLOsKqlWithFiltersTotal

		api := kbapi.SLOsIndicatorPropertiesCustomKql{
			Params: struct {
				DataViewId     *string                       `json:"dataViewId,omitempty"` //nolint:revive // var-naming: API struct field
				Filter         *kbapi.SLOsKqlWithFilters     `json:"filter,omitempty"`
				Good           kbapi.SLOsKqlWithFiltersGood  `json:"good"`
				Index          string                        `json:"index"`
				TimestampField string                        `json:"timestampField"`
				Total          kbapi.SLOsKqlWithFiltersTotal `json:"total"`
			}{
				Index:          "logs-*",
				DataViewId:     nil,
				Filter:         nil,
				Good:           emptyGood,
				Total:          emptyTotal,
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
		// Empty unions will fail As* — so Good and Total will be null
		assert.True(t, ind.Good.IsNull())
		assert.True(t, ind.Total.IsNull())
		assert.True(t, ind.FilterKql.IsNull())
		assert.True(t, ind.GoodKql.IsNull())
		assert.True(t, ind.TotalKql.IsNull())
	})

	t.Run("maps good as object form when API returns filters", func(t *testing.T) {
		var good kbapi.SLOsKqlWithFiltersGood
		filters := []kbapi.SLOsFilter{{Query: func() *map[string]interface{} {
			m := map[string]interface{}{"match_all": map[string]interface{}{}}
			return &m
		}()}}
		require.NoError(t, good.FromSLOsKqlWithFiltersGood1(kbapi.SLOsKqlWithFiltersGood1{
			KqlQuery: strPtr("event.outcome: success"),
			Filters:  &filters,
		}))

		api := kbapi.SLOsIndicatorPropertiesCustomKql{
			Params: struct {
				DataViewId     *string                       `json:"dataViewId,omitempty"` //nolint:revive // var-naming: API struct field
				Filter         *kbapi.SLOsKqlWithFilters     `json:"filter,omitempty"`
				Good           kbapi.SLOsKqlWithFiltersGood  `json:"good"`
				Index          string                        `json:"index"`
				TimestampField string                        `json:"timestampField"`
				Total          kbapi.SLOsKqlWithFiltersTotal `json:"total"`
			}{
				Index:          "logs-*",
				Good:           good,
				Total:          mustKqlTotalFromString(t, "*"),
				TimestampField: "@timestamp",
			},
		}

		var m tfModel
		diags := m.populateFromKqlCustomIndicator(api)
		require.False(t, diags.HasError())
		ind := m.KqlCustomIndicator[0]
		assert.True(t, ind.Good.IsNull())
		assert.False(t, ind.GoodKql.IsNull())
		kqq := ind.GoodKql.Attributes()["kql_query"].(types.String)
		assert.Equal(t, "event.outcome: success", kqq.ValueString())
	})
}

func strPtr(s string) *string { return &s }

func mustKqlTotalFromString(t *testing.T, s string) kbapi.SLOsKqlWithFiltersTotal {
	t.Helper()
	var total kbapi.SLOsKqlWithFiltersTotal
	require.NoError(t, total.FromSLOsKqlWithFiltersTotal0(s))
	return total
}
