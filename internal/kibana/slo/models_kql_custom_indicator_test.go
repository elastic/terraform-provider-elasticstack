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
	"encoding/json"
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
		filters := []kbapi.SLOsFilter{{Query: func() *map[string]any {
			m := map[string]any{"match_all": map[string]any{}}
			return &m
		}()}}
		kq := "event.outcome: success"
		require.NoError(t, good.FromSLOsKqlWithFiltersGood1(kbapi.SLOsKqlWithFiltersGood1{
			KqlQuery: &kq,
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

	t.Run("maps filter and total as object form when API returns filters", func(t *testing.T) {
		var f kbapi.SLOsKqlWithFilters
		fKq := "host.name: *"
		rowFilters := []kbapi.SLOsFilter{{
			Query: func() *map[string]any { m := map[string]any{"match_all": map[string]any{}}; return &m }(),
		}}
		require.NoError(t, f.FromSLOsKqlWithFilters1(kbapi.SLOsKqlWithFilters1{KqlQuery: &fKq, Filters: &rowFilters}))
		tot := kbapi.SLOsKqlWithFiltersTotal{}
		tKq := "event.category: *"
		t1 := kbapi.SLOsKqlWithFiltersTotal1{
			KqlQuery: &tKq,
			Filters: &[]kbapi.SLOsFilter{{
				Query: func() *map[string]any { m := map[string]any{"bool": map[string]any{}}; return &m }(),
			}},
		}
		require.NoError(t, tot.FromSLOsKqlWithFiltersTotal1(t1))
		var g kbapi.SLOsKqlWithFiltersGood
		require.NoError(t, g.FromSLOsKqlWithFiltersGood0("ok"))

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
				Filter:         &f,
				Good:           g,
				Total:          tot,
				TimestampField: "@timestamp",
			},
		}

		var m tfModel
		require.False(t, m.populateFromKqlCustomIndicator(api).HasError())
		ind := m.KqlCustomIndicator[0]
		assert.True(t, ind.Filter.IsNull())
		assert.False(t, ind.FilterKql.IsNull())
		assert.True(t, ind.Total.IsNull())
		assert.False(t, ind.TotalKql.IsNull())
		fkq := ind.FilterKql.Attributes()["kql_query"].(types.String)
		assert.Equal(t, "host.name: *", fkq.ValueString())
		tkq := ind.TotalKql.Attributes()["kql_query"].(types.String)
		assert.Equal(t, "event.category: *", tkq.ValueString())
	})

	t.Run("round_trip object form filter and total to API and back", func(t *testing.T) {
		q := jsontypes.NewNormalizedValue(`{"match_all":{}}`)
		row, d := types.ObjectValue(tfKqlFilterRowObjectType.AttrTypes, map[string]attr.Value{"query": q})
		require.False(t, d.HasError())
		list, d := types.ListValue(tfKqlFilterRowObjectType, []attr.Value{row})
		require.False(t, d.HasError())
		filterKql, d := types.ObjectValue(tfKqlKqlObjectAttrTypes, map[string]attr.Value{
			"kql_query": types.StringValue(`@timestamp: *`),
			"filters":   list,
		})
		require.False(t, d.HasError())
		totKql, d := types.ObjectValue(tfKqlKqlObjectAttrTypes, map[string]attr.Value{
			"kql_query": types.StringValue(`*`),
			"filters":   list,
		})
		require.False(t, d.HasError())

		m1 := tfModel{KqlCustomIndicator: []tfKqlCustomIndicator{{
			Index:          types.StringValue("logs-*"),
			Filter:         types.StringNull(),
			FilterKql:      filterKql,
			Good:           types.StringValue("g"),
			GoodKql:        types.ObjectNull(tfKqlKqlObjectAttrTypes),
			Total:          types.StringNull(),
			TotalKql:       totKql,
			TimestampField: types.StringValue("@timestamp"),
		}}}

		ok, indUnion, di := m1.kqlCustomIndicatorToAPI()
		require.True(t, ok)
		require.False(t, di.HasError(), "%+v", di)
		kind, err := indUnion.AsSLOsIndicatorPropertiesCustomKql()
		require.NoError(t, err)
		var m2 tfModel
		require.False(t, m2.populateFromKqlCustomIndicator(kind).HasError())
		out := m2.KqlCustomIndicator[0]
		assert.Equal(t, `@timestamp: *`, out.FilterKql.Attributes()["kql_query"].(types.String).ValueString())
		assert.Equal(t, `*`, out.TotalKql.Attributes()["kql_query"].(types.String).ValueString())
		assert.Equal(t, "g", out.Good.ValueString())

		fltList := out.FilterKql.Attributes()["filters"].(types.List)
		require.Len(t, fltList.Elements(), 1)
		totList := out.TotalKql.Attributes()["filters"].(types.List)
		require.Len(t, totList.Elements(), 1)
		for i, l := range []types.List{fltList, totList} {
			row := l.Elements()[0].(types.Object)
			qn := row.Attributes()["query"].(jsontypes.Normalized)
			var m map[string]any
			require.NoError(t, json.Unmarshal([]byte(qn.ValueString()), &m), "filter row %d", i)
			_, has := m["match_all"]
			assert.True(t, has, "expected match_all in nested filters[%d] query", i)
		}
	})
}

func TestKqlTFFormToAPI1_filterQueryDiagnostics(t *testing.T) {
	unknownQueryRow, d := types.ObjectValue(tfKqlFilterRowObjectType.AttrTypes, map[string]attr.Value{
		"query": jsontypes.NewNormalizedUnknown(),
	})
	require.False(t, d.HasError())
	list, d := types.ListValue(tfKqlFilterRowObjectType, []attr.Value{unknownQueryRow})
	require.False(t, d.HasError())
	obj, d := types.ObjectValue(tfKqlKqlObjectAttrTypes, map[string]attr.Value{
		"kql_query": types.StringValue("a:b"),
		"filters":   list,
	})
	require.False(t, d.HasError())
	_, diags := kqlTFFormToAPI1(obj, "kql_test")
	require.True(t, diags.HasError(), "expected error for unknown filter query")
	assert.Contains(t, diags[0].Detail(), "is not yet known", "%s", diags[0].Detail())

	nullQueryRow, d := types.ObjectValue(tfKqlFilterRowObjectType.AttrTypes, map[string]attr.Value{
		"query": jsontypes.NewNormalizedNull(),
	})
	require.False(t, d.HasError())
	list2, d := types.ListValue(tfKqlFilterRowObjectType, []attr.Value{nullQueryRow})
	require.False(t, d.HasError())
	obj2, d := types.ObjectValue(tfKqlKqlObjectAttrTypes, map[string]attr.Value{
		"kql_query": types.StringValue("a:b"),
		"filters":   list2,
	})
	require.False(t, d.HasError())
	_, diags2 := kqlTFFormToAPI1(obj2, "kql_test2")
	require.True(t, diags2.HasError(), "expected error for null filter query")
	assert.Contains(t, diags2[0].Detail(), "null is not valid", "%s", diags2[0].Detail())
}

func mustKqlTotalFromString(t *testing.T, s string) kbapi.SLOsKqlWithFiltersTotal {
	t.Helper()
	var total kbapi.SLOsKqlWithFiltersTotal
	require.NoError(t, total.FromSLOsKqlWithFiltersTotal0(s))
	return total
}
