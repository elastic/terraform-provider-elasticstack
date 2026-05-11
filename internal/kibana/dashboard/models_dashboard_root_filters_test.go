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

package dashboard

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func rootFiltersListNull() types.List {
	return types.ListNull(dashboardRootSavedFiltersElementType())
}

func rootFiltersListEmpty(ctx context.Context, t *testing.T) types.List {
	t.Helper()
	var diags diag.Diagnostics
	l := typeutils.ListValueFrom(ctx, []chartFilterJSONModel{}, dashboardRootSavedFiltersElementType(), path.Root("filters"), &diags)
	require.False(t, diags.HasError())
	return l
}

func mustUnmarshalFilterItem(t *testing.T, rawJSON string) kbapi.DashboardFilters_Item {
	t.Helper()
	var item kbapi.DashboardFilters_Item
	require.NoError(t, json.Unmarshal([]byte(rawJSON), &item))
	return item
}

func canonicalJSONFromRaw(t *testing.T, rawJSON string) string {
	t.Helper()
	var root any
	require.NoError(t, json.Unmarshal([]byte(rawJSON), &root))
	b, err := json.Marshal(sortJSONMapKeysRecursive(root))
	require.NoError(t, err)
	return string(b)
}

func filterModelsFromRawJSON(t *testing.T, raw ...string) []chartFilterJSONModel {
	t.Helper()
	out := make([]chartFilterJSONModel, len(raw))
	for i, r := range raw {
		out[i] = chartFilterJSONModel{FilterJSON: jsontypes.NewNormalizedValue(r)}
	}
	return out
}

func Test_mapDashboardFiltersFromAPI_nullPreserved_whenAPINilOrEmpty(t *testing.T) {
	ctx := context.Background()

	t.Run("API Filters nil", func(t *testing.T) {
		m := &dashboardModel{Filters: rootFiltersListNull()}
		var diags diag.Diagnostics
		m.mapDashboardFiltersFromAPI(ctx, &kbapi.KbnDashboardData{Filters: nil}, &diags)
		require.False(t, diags.HasError())
		require.True(t, m.Filters.IsNull())
	})

	t.Run("API Filters pointer to empty slice", func(t *testing.T) {
		m := &dashboardModel{Filters: rootFiltersListNull()}
		empty := []kbapi.DashboardFilters_Item{}
		var diags diag.Diagnostics
		m.mapDashboardFiltersFromAPI(ctx, &kbapi.KbnDashboardData{Filters: &empty}, &diags)
		require.False(t, diags.HasError())
		require.True(t, m.Filters.IsNull())
	})

	t.Run("known empty list stays empty when API nil", func(t *testing.T) {
		m := &dashboardModel{Filters: rootFiltersListEmpty(ctx, t)}
		var diags diag.Diagnostics
		m.mapDashboardFiltersFromAPI(ctx, &kbapi.KbnDashboardData{Filters: nil}, &diags)
		require.False(t, diags.HasError())
		require.False(t, m.Filters.IsNull())
		require.False(t, m.Filters.IsUnknown())
		require.Empty(t, m.Filters.Elements())
	})

	t.Run("known empty list stays empty when API empty slice", func(t *testing.T) {
		m := &dashboardModel{Filters: rootFiltersListEmpty(ctx, t)}
		empty := []kbapi.DashboardFilters_Item{}
		var diags diag.Diagnostics
		m.mapDashboardFiltersFromAPI(ctx, &kbapi.KbnDashboardData{Filters: &empty}, &diags)
		require.False(t, diags.HasError())
		require.False(t, m.Filters.IsNull())
		require.Empty(t, m.Filters.Elements())
	})
}

func Test_mapDashboardFiltersFromAPI_orderPreserved(t *testing.T) {
	ctx := context.Background()
	raws := []string{
		`{"type":"condition","condition":{"field":"host.name","operator":"is","value":"a"}}`,
		`{"type":"condition","condition":{"field":"service.name","operator":"is","value":"b"}}`,
		`{"type":"condition","condition":{"field":"kubernetes.pod.name","operator":"is","value":"c"}}`,
	}
	items := make([]kbapi.DashboardFilters_Item, len(raws))
	for i, r := range raws {
		items[i] = mustUnmarshalFilterItem(t, r)
	}
	m := &dashboardModel{Filters: rootFiltersListNull()}
	var diags diag.Diagnostics
	m.mapDashboardFiltersFromAPI(ctx, &kbapi.KbnDashboardData{Filters: &items}, &diags)
	require.False(t, diags.HasError())

	elems := typeutils.ListTypeAs[chartFilterJSONModel](ctx, m.Filters, path.Root("filters"), &diags)
	require.False(t, diags.HasError())
	require.Len(t, elems, len(raws))
	for i, r := range raws {
		require.JSONEq(t, canonicalJSONFromRaw(t, r), elems[i].FilterJSON.ValueString())
	}
}

func Test_mapDashboardFiltersFromAPI_normalizesKeyOrder(t *testing.T) {
	ctx := context.Background()
	reordered := `{"condition":{"value":"web-01","field":"host.name","operator":"is"},"type":"condition"}`
	item := mustUnmarshalFilterItem(t, reordered)
	m := &dashboardModel{Filters: rootFiltersListNull()}
	var diags diag.Diagnostics
	m.mapDashboardFiltersFromAPI(ctx, &kbapi.KbnDashboardData{Filters: &[]kbapi.DashboardFilters_Item{item}}, &diags)
	require.False(t, diags.HasError())

	elems := typeutils.ListTypeAs[chartFilterJSONModel](ctx, m.Filters, path.Root("filters"), &diags)
	require.False(t, diags.HasError())
	require.Len(t, elems, 1)
	want := canonicalJSONFromRaw(t, reordered)
	require.Equal(t, want, elems[0].FilterJSON.ValueString())
	require.Contains(t, elems[0].FilterJSON.ValueString(), `"field":"host.name"`)
}

func Test_dashboardModel_dashboardFiltersToCreateAPI_writeSemantics(t *testing.T) {
	ctx := context.Background()
	raws := []string{
		`{"type":"condition","condition":{"field":"a.field","operator":"is","value":"1"}}`,
		`{"type":"condition","condition":{"field":"b.field","operator":"is","value":"2"}}`,
		`{"type":"condition","condition":{"field":"c.field","operator":"is","value":"3"}}`,
	}

	t.Run("null Filters omits request filters", func(t *testing.T) {
		m := &dashboardModel{Filters: rootFiltersListNull()}
		var diags diag.Diagnostics
		var req kbapi.PostDashboardsJSONRequestBody
		m.dashboardFiltersToCreateAPI(ctx, &req, &diags)
		require.False(t, diags.HasError())
		require.Nil(t, req.Filters)
	})

	t.Run("known empty sends non-nil empty slice", func(t *testing.T) {
		m := &dashboardModel{Filters: rootFiltersListEmpty(ctx, t)}
		var diags diag.Diagnostics
		var req kbapi.PostDashboardsJSONRequestBody
		m.dashboardFiltersToCreateAPI(ctx, &req, &diags)
		require.False(t, diags.HasError())
		require.NotNil(t, req.Filters)
		require.Empty(t, *req.Filters)
	})

	t.Run("order preserved and round-trips JSON", func(t *testing.T) {
		elems := filterModelsFromRawJSON(t, raws...)
		var diags diag.Diagnostics
		m := &dashboardModel{
			Filters: typeutils.ListValueFrom(ctx, elems, dashboardRootSavedFiltersElementType(), path.Root("filters"), &diags),
		}
		require.False(t, diags.HasError())
		var req kbapi.PostDashboardsJSONRequestBody
		m.dashboardFiltersToCreateAPI(ctx, &req, &diags)
		require.False(t, diags.HasError())
		require.NotNil(t, req.Filters)
		require.Len(t, *req.Filters, len(raws))
		for i := range raws {
			gotBytes, err := json.Marshal((*req.Filters)[i])
			require.NoError(t, err)
			require.JSONEq(t, canonicalJSONFromRaw(t, raws[i]), string(gotBytes))
		}
	})
}

func Test_dashboardModel_dashboardFiltersToUpdateAPI_writeSemantics(t *testing.T) {
	ctx := context.Background()
	raws := []string{
		`{"type":"condition","condition":{"field":"x.field","operator":"is","value":"9"}}`,
		`{"type":"condition","condition":{"field":"y.field","operator":"is","value":"8"}}`,
		`{"type":"condition","condition":{"field":"z.field","operator":"is","value":"7"}}`,
	}

	t.Run("null Filters omits request filters", func(t *testing.T) {
		m := &dashboardModel{Filters: rootFiltersListNull()}
		var diags diag.Diagnostics
		var req kbapi.PutDashboardsIdJSONRequestBody
		m.dashboardFiltersToUpdateAPI(ctx, &req, &diags)
		require.False(t, diags.HasError())
		require.Nil(t, req.Filters)
	})

	t.Run("known empty sends non-nil empty slice", func(t *testing.T) {
		m := &dashboardModel{Filters: rootFiltersListEmpty(ctx, t)}
		var diags diag.Diagnostics
		var req kbapi.PutDashboardsIdJSONRequestBody
		m.dashboardFiltersToUpdateAPI(ctx, &req, &diags)
		require.False(t, diags.HasError())
		require.NotNil(t, req.Filters)
		require.Empty(t, *req.Filters)
	})

	t.Run("order preserved and round-trips JSON", func(t *testing.T) {
		elems := filterModelsFromRawJSON(t, raws...)
		var diags diag.Diagnostics
		m := &dashboardModel{
			Filters: typeutils.ListValueFrom(ctx, elems, dashboardRootSavedFiltersElementType(), path.Root("filters"), &diags),
		}
		require.False(t, diags.HasError())
		var req kbapi.PutDashboardsIdJSONRequestBody
		m.dashboardFiltersToUpdateAPI(ctx, &req, &diags)
		require.False(t, diags.HasError())
		require.NotNil(t, req.Filters)
		require.Len(t, *req.Filters, len(raws))
		for i := range raws {
			gotBytes, err := json.Marshal((*req.Filters)[i])
			require.NoError(t, err)
			require.JSONEq(t, canonicalJSONFromRaw(t, raws[i]), string(gotBytes))
		}
	})
}
