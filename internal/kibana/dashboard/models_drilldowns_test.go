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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_structuredDrilldowns_dashboardRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	want := models.DrilldownItemModel{
		Dashboard: &models.DrilldownDashboardBlockModel{
			DashboardID:  types.StringValue("dash-id-1"),
			Label:        types.StringValue("Open detail dashboard"),
			UseFilters:   types.BoolValue(false),
			UseTimeRange: types.BoolValue(true),
			OpenInNewTab: types.BoolValue(true),
		},
	}
	api, diags := toAPI(models.DrilldownsModel{want})
	require.False(t, diags.HasError())
	require.NotNil(t, api)
	got, diags := fromAPI(ctx, api)
	require.False(t, diags.HasError())
	require.Len(t, got, 1)
	assertDashboardBlocksEqual(t, want.Dashboard, got[0].Dashboard)
	require.Nil(t, got[0].Discover)
	require.Nil(t, got[0].URL)
}

func Test_structuredDrilldowns_discoverRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	want := models.DrilldownItemModel{
		Discover: &models.DrilldownDiscoverBlockModel{
			Label:        types.StringValue("Open in Discover"),
			OpenInNewTab: types.BoolValue(false),
		},
	}
	api, diags := toAPI(models.DrilldownsModel{want})
	require.False(t, diags.HasError())
	got, diags := fromAPI(ctx, api)
	require.False(t, diags.HasError())
	require.Len(t, got, 1)
	require.Nil(t, got[0].Dashboard)
	assertDiscoverBlocksEqual(t, want.Discover, got[0].Discover)
	require.Nil(t, got[0].URL)
}

func Test_structuredDrilldowns_urlRoundTrip_explicitTrigger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	want := models.DrilldownItemModel{
		URL: &models.DrilldownURLBlockModel{
			URL:          types.StringValue("https://example.com/{{event.field}}"),
			Label:        types.StringValue("External"),
			Trigger:      types.StringValue("on_click_value"),
			EncodeURL:    types.BoolValue(false),
			OpenInNewTab: types.BoolValue(true),
		},
	}
	api, diags := toAPI(models.DrilldownsModel{want})
	require.False(t, diags.HasError())
	got, diags := fromAPI(ctx, api)
	require.False(t, diags.HasError())
	require.Len(t, got, 1)
	require.Nil(t, got[0].Dashboard)
	require.Nil(t, got[0].Discover)
	assertURLBlocksEqual(t, want.URL, got[0].URL)
}

func Test_structuredDrilldowns_mixedThreeItemsPreserveOrderAndKind(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	want := models.DrilldownsModel{
		{
			Dashboard: &models.DrilldownDashboardBlockModel{
				DashboardID: types.StringValue("dash-a"),
				Label:       types.StringValue("D1"),
				UseFilters:  types.BoolNull(),
			},
		},
		{
			URL: &models.DrilldownURLBlockModel{
				URL:       types.StringValue("https://a.example"),
				Label:     types.StringValue("url1"),
				Trigger:   types.StringValue("on_open_panel_menu"),
				EncodeURL: types.BoolValue(true),
			},
		},
		{
			Discover: &models.DrilldownDiscoverBlockModel{
				Label:        types.StringValue("disc"),
				OpenInNewTab: types.BoolNull(),
			},
		},
	}
	lensAPI, diags := toAPI(want)
	require.False(t, diags.HasError())
	gotLens, diags := fromAPI(ctx, lensAPI)
	require.False(t, diags.HasError())
	require.Len(t, gotLens, 3)

	visAPI, diags := drilldownsToVisByRefAPI(want)
	require.False(t, diags.HasError())
	gotVis, diags := drilldownsFromVisByRefAPI(ctx, visAPI)
	require.False(t, diags.HasError())
	require.Len(t, gotVis, 3)

	require.NotNil(t, gotVis[0].Dashboard)
	require.NotNil(t, gotVis[1].URL)
	require.NotNil(t, gotVis[2].Discover)

	for i := range want {
		assertDrilldownItemEqualKinds(t, want[i], gotLens[i])
		assertDrilldownItemEqualKinds(t, want[i], gotVis[i])
	}
}

func Test_structuredDrilldowns_toAPI_urlInvalidTrigger(t *testing.T) {
	t.Parallel()
	bad := models.DrilldownsModel{
		{
			URL: &models.DrilldownURLBlockModel{
				URL:     types.StringValue("https://x"),
				Label:   types.StringValue("lbl"),
				Trigger: types.StringValue("not_a_real_trigger"),
			},
		},
	}
	_, diags := toAPI(bad)
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Detail(), "Unsupported URL drilldown `trigger`")
}

func Test_structuredDrilldowns_toAPI_unknownOptionalsSkipWireFields_andNoPanic(t *testing.T) {
	t.Parallel()
	m := models.DrilldownItemModel{
		Dashboard: &models.DrilldownDashboardBlockModel{
			DashboardID:  types.StringValue("d1"),
			Label:        types.StringValue("lbl"),
			UseFilters:   types.BoolUnknown(),
			UseTimeRange: types.BoolValue(true),
			OpenInNewTab: types.BoolNull(),
		},
	}
	require.NotPanics(t, func() {
		api, diags := toAPI(models.DrilldownsModel{m})
		require.False(t, diags.HasError())
		require.Len(t, *api, 1)
		raw, err := json.Marshal((*api)[0])
		require.NoError(t, err)
		var wire map[string]any
		require.NoError(t, json.Unmarshal(raw, &wire))
		if _, ok := wire["use_filters"]; ok {
			t.Fatalf("expected use_filters omitted when unknown, got wire=%v", wire)
		}
		require.Equal(t, true, wire["use_time_range"])
	})

	urlm := models.DrilldownItemModel{
		URL: &models.DrilldownURLBlockModel{
			URL:          types.StringValue("https://example.com"),
			Label:        types.StringValue("x"),
			Trigger:      types.StringUnknown(),
			EncodeURL:    types.BoolUnknown(),
			OpenInNewTab: types.BoolValue(false),
		},
	}
	require.NotPanics(t, func() {
		api, diags := toAPI(models.DrilldownsModel{urlm})
		require.False(t, diags.HasError())
		raw, err := json.Marshal((*api)[0])
		require.NoError(t, err)
		var wire map[string]any
		require.NoError(t, json.Unmarshal(raw, &wire))
		if _, ok := wire["trigger"]; ok {
			t.Fatalf("expected trigger omitted when unknown")
		}
		if _, ok := wire["encode_url"]; ok {
			t.Fatalf("expected encode_url omitted when unknown")
		}
		require.Equal(t, false, wire["open_in_new_tab"])
	})

	disc := models.DrilldownItemModel{
		Discover: &models.DrilldownDiscoverBlockModel{
			Label:        types.StringValue("d"),
			OpenInNewTab: types.BoolUnknown(),
		},
	}
	require.NotPanics(t, func() {
		api, diags := toAPI(models.DrilldownsModel{disc})
		require.False(t, diags.HasError())
		raw, err := json.Marshal((*api)[0])
		require.NoError(t, err)
		var wire map[string]any
		require.NoError(t, json.Unmarshal(raw, &wire))
		if _, ok := wire["open_in_new_tab"]; ok {
			t.Fatalf("expected open_in_new_tab omitted when unknown")
		}
	})
}

func Test_structuredDrilldowns_fromAPI_accumulatesMultipleItemErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	items := []kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item{
		drilldownLensItemFromJSON(t, `{"type":"bad_a"}`),
		drilldownLensItemFromJSON(t, `{"type":"bad_b"}`),
	}
	_, diags := fromAPI(ctx, &items)
	require.True(t, diags.HasError())
	errs := diags.Errors()
	require.Len(t, errs, 2)
	require.Contains(t, errs[0].Detail(), "Unsupported API drilldown `type`")
	require.Contains(t, errs[1].Detail(), "Unsupported API drilldown `type`")

	visItems := []kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item{
		drilldownVisItemFromJSON(t, `{"type":"bad_a"}`),
		drilldownVisItemFromJSON(t, `{"type":"bad_b"}`),
	}
	_, diagsVis := drilldownsFromVisByRefAPI(ctx, &visItems)
	require.True(t, diagsVis.HasError())
	errsVis := diagsVis.Errors()
	require.Len(t, errsVis, 2)
}

func Test_structuredDrilldowns_fromAPI_unsupportedDrilldownType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	itemLens := drilldownLensItemFromJSON(t, `{"type":"future_drilldown_kind","extra":1}`)
	_, diags := fromAPI(ctx, &[]kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item{itemLens})
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Detail(), "Unsupported API drilldown `type`")

	itemVis := drilldownVisItemFromJSON(t, `{"type":"future_drilldown_kind"}`)
	_, diagsVis := drilldownsFromVisByRefAPI(ctx, &[]kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item{itemVis})
	require.True(t, diagsVis.HasError())
	require.Contains(t, diagsVis.Errors()[0].Detail(), "Unsupported API drilldown `type`")
}

func Test_structuredDrilldowns_fromAPI_missingType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	item := drilldownLensItemFromJSON(t, `{"dashboard_id":"a","label":"b","trigger":"on_apply_filter"}`)
	_, diags := fromAPI(ctx, &[]kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item{item})
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Detail(), "missing required discriminator field `type`")
}

func Test_structuredDrilldowns_fromAPI_dashboardWrongTriggerLossless(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	item := drilldownLensItemFromJSON(t, `{
	  "type": "dashboard_drilldown",
	  "trigger": "on_click_row",
	  "dashboard_id": "abc",
	  "label": "x"
	}`)
	_, diags := fromAPI(ctx, &[]kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item{item})
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Detail(), "Dashboard drilldown API `trigger`")
}

func Test_structuredDrilldowns_fromAPI_urlMissingTrigger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	payload := `{
	  "type": "url_drilldown",
	  "url": "https://example.com/none",
	  "label": "x"
	}`
	itemLens := drilldownLensItemFromJSON(t, payload)
	_, diagsLens := fromAPI(ctx, &[]kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item{itemLens})
	require.True(t, diagsLens.HasError())
	require.Contains(t, diagsLens.Errors()[0].Detail(), "omits required field `trigger`")

	itemVis := drilldownVisItemFromJSON(t, payload)
	_, diagsVis := drilldownsFromVisByRefAPI(ctx, &[]kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item{itemVis})
	require.True(t, diagsVis.HasError())
	require.Contains(t, diagsVis.Errors()[0].Detail(), "omits required field `trigger`")
}

func Test_structuredDrilldowns_fromAPI_urlUnreadableTrigger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	item := drilldownLensItemFromJSON(t, `{
	  "type": "url_drilldown",
	  "trigger": "__not_known__",
	  "url": "https://x",
	  "label": "lbl"
	}`)
	_, diags := fromAPI(ctx, &[]kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item{item})
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Detail(), "unsupported API `trigger`")
}

func drilldownLensItemFromJSON(t *testing.T, payload string) kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item {
	t.Helper()
	var item kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item
	require.NoError(t, json.Unmarshal([]byte(payload), &item))
	return item
}

func drilldownVisItemFromJSON(t *testing.T, payload string) kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item {
	t.Helper()
	var item kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item
	require.NoError(t, json.Unmarshal([]byte(payload), &item))
	return item
}

func assertDrilldownItemEqualKinds(t *testing.T, want, got models.DrilldownItemModel) {
	t.Helper()
	if want.Dashboard != nil {
		require.NotNil(t, got.Dashboard)
		assertDashboardBlocksEqual(t, want.Dashboard, got.Dashboard)
		require.Nil(t, got.Discover)
		require.Nil(t, got.URL)
		return
	}
	if want.Discover != nil {
		require.NotNil(t, got.Discover)
		assertDiscoverBlocksEqual(t, want.Discover, got.Discover)
		require.Nil(t, got.Dashboard)
		require.Nil(t, got.URL)
		return
	}
	if want.URL != nil {
		require.NotNil(t, got.URL)
		assertURLBlocksEqual(t, want.URL, got.URL)
		require.Nil(t, got.Dashboard)
		require.Nil(t, got.Discover)
	}
}

func assertDashboardBlocksEqual(t *testing.T, want, got *models.DrilldownDashboardBlockModel) {
	t.Helper()
	require.Equal(t, want.DashboardID.ValueString(), got.DashboardID.ValueString())
	require.Equal(t, want.Label.ValueString(), got.Label.ValueString())
	require.Equal(t, want.UseFilters.IsNull(), got.UseFilters.IsNull())
	if !want.UseFilters.IsNull() {
		require.Equal(t, want.UseFilters.ValueBool(), got.UseFilters.ValueBool())
	}
	require.Equal(t, want.UseTimeRange.IsNull(), got.UseTimeRange.IsNull())
	if !want.UseTimeRange.IsNull() {
		require.Equal(t, want.UseTimeRange.ValueBool(), got.UseTimeRange.ValueBool())
	}
	require.Equal(t, want.OpenInNewTab.IsNull(), got.OpenInNewTab.IsNull())
	if !want.OpenInNewTab.IsNull() {
		require.Equal(t, want.OpenInNewTab.ValueBool(), got.OpenInNewTab.ValueBool())
	}
}

func assertDiscoverBlocksEqual(t *testing.T, want, got *models.DrilldownDiscoverBlockModel) {
	t.Helper()
	require.Equal(t, want.Label.ValueString(), got.Label.ValueString())
	require.Equal(t, want.OpenInNewTab.IsNull(), got.OpenInNewTab.IsNull())
	if !want.OpenInNewTab.IsNull() {
		require.Equal(t, want.OpenInNewTab.ValueBool(), got.OpenInNewTab.ValueBool())
	}
}

func assertURLBlocksEqual(t *testing.T, want, got *models.DrilldownURLBlockModel) {
	t.Helper()
	require.Equal(t, want.URL.ValueString(), got.URL.ValueString())
	require.Equal(t, want.Label.ValueString(), got.Label.ValueString())
	require.Equal(t, want.Trigger.IsNull(), got.Trigger.IsNull())
	if !want.Trigger.IsNull() {
		require.Equal(t, want.Trigger.ValueString(), got.Trigger.ValueString())
	}
	require.Equal(t, want.EncodeURL.IsNull(), got.EncodeURL.IsNull())
	if !want.EncodeURL.IsNull() {
		require.Equal(t, want.EncodeURL.ValueBool(), got.EncodeURL.ValueBool())
	}
	require.Equal(t, want.OpenInNewTab.IsNull(), got.OpenInNewTab.IsNull())
	if !want.OpenInNewTab.IsNull() {
		require.Equal(t, want.OpenInNewTab.ValueBool(), got.OpenInNewTab.ValueBool())
	}
}
