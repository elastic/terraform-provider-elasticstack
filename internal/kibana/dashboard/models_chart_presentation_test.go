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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_lensDrilldownItemToRawJSON_dashboard_and_defaults(t *testing.T) {
	item := models.LensDrilldownItemTFModel{
		DashboardDrilldown: &models.LensDashboardDrilldownTFModel{
			DashboardID:  types.StringValue("dash1"),
			Label:        types.StringValue("Go to dash"),
			UseFilters:   types.BoolValue(false),
			UseTimeRange: types.BoolValue(true),
			OpenInNewTab: types.BoolValue(true),
		},
	}
	raw, diags := lensDrilldownItemToRawJSON(item, 0)
	require.False(t, diags.HasError())

	var wire map[string]any
	require.NoError(t, json.Unmarshal(raw, &wire))
	assert.Equal(t, "dashboard_drilldown", wire["type"])
	assert.Equal(t, "on_apply_filter", wire["trigger"])
	assert.Equal(t, "dash1", wire["dashboard_id"])
	assert.Equal(t, "Go to dash", wire["label"])
	assert.Equal(t, false, wire["use_filters"])
	assert.Equal(t, true, wire["use_time_range"])
	assert.Equal(t, true, wire["open_in_new_tab"])
}

func Test_lensDrilldownItemToRawJSON_discover_defaults(t *testing.T) {
	item := models.LensDrilldownItemTFModel{
		DiscoverDrilldown: &models.LensDiscoverDrilldownTFModel{
			Label:        types.StringValue("Open Discover"),
			OpenInNewTab: types.BoolValue(false),
		},
	}
	raw, diags := lensDrilldownItemToRawJSON(item, 1)
	require.False(t, diags.HasError())

	var wire map[string]any
	require.NoError(t, json.Unmarshal(raw, &wire))
	assert.Equal(t, "discover_drilldown", wire["type"])
	assert.Equal(t, "on_apply_filter", wire["trigger"])
	assert.Equal(t, "Open Discover", wire["label"])
	assert.Equal(t, false, wire["open_in_new_tab"])
}

func Test_lensDrilldownItemToRawJSON_url_includes_trigger(t *testing.T) {
	item := models.LensDrilldownItemTFModel{
		URLDrilldown: &models.LensURLDrilldownTFModel{
			URL:          types.StringValue("https://example.test/{{event.url}}"),
			Label:        types.StringValue("External"),
			Trigger:      types.StringValue("on_click_row"),
			EncodeURL:    types.BoolValue(false),
			OpenInNewTab: types.BoolValue(false),
		},
	}
	raw, diags := lensDrilldownItemToRawJSON(item, 2)
	require.False(t, diags.HasError())

	var wire map[string]any
	require.NoError(t, json.Unmarshal(raw, &wire))
	assert.Equal(t, "url_drilldown", wire["type"])
	assert.Equal(t, "https://example.test/{{event.url}}", wire["url"])
	assert.Equal(t, "External", wire["label"])
	assert.Equal(t, "on_click_row", wire["trigger"])
	assert.Equal(t, false, wire["encode_url"])
	assert.Equal(t, false, wire["open_in_new_tab"])
}

func Test_lensDrilldownItemFromAPIJSON_dispatch_and_trigger_defaults(t *testing.T) {
	t.Run("dashboard trigger omitted defaults", func(t *testing.T) {
		raw := []byte(`{"type":"dashboard_drilldown","dashboard_id":"d1","label":"L"}`)
		item, diags := lensDrilldownItemFromAPIJSON(raw, 0)
		require.False(t, diags.HasError())
		require.NotNil(t, item.DashboardDrilldown)
		assert.Nil(t, item.DiscoverDrilldown)
		assert.Nil(t, item.URLDrilldown)
		assert.Equal(t, "d1", item.DashboardDrilldown.DashboardID.ValueString())
		assert.Equal(t, "L", item.DashboardDrilldown.Label.ValueString())
		assert.Equal(t, lensDrilldownTriggerOnApplyFilter, item.DashboardDrilldown.Trigger.ValueString())
	})

	t.Run("discover trigger omitted defaults", func(t *testing.T) {
		raw := []byte(`{"type":"discover_drilldown","label":"D"}`)
		item, diags := lensDrilldownItemFromAPIJSON(raw, 0)
		require.False(t, diags.HasError())
		require.NotNil(t, item.DiscoverDrilldown)
		assert.Equal(t, "D", item.DiscoverDrilldown.Label.ValueString())
		assert.Equal(t, lensDrilldownTriggerOnApplyFilter, item.DiscoverDrilldown.Trigger.ValueString())
	})

	t.Run("url requires explicit trigger in payload", func(t *testing.T) {
		raw := []byte(`{"type":"url_drilldown","url":"https://x","label":"U","trigger":"on_open_panel_menu"}`)
		item, diags := lensDrilldownItemFromAPIJSON(raw, 0)
		require.False(t, diags.HasError())
		require.NotNil(t, item.URLDrilldown)
		assert.Equal(t, "https://x", item.URLDrilldown.URL.ValueString())
		assert.Equal(t, "on_open_panel_menu", item.URLDrilldown.Trigger.ValueString())
	})
}

func Test_lensDrilldownsToRawJSON_variantCountErrors(t *testing.T) {
	t.Run("multiple variants set", func(t *testing.T) {
		item := models.LensDrilldownItemTFModel{
			DashboardDrilldown: &models.LensDashboardDrilldownTFModel{
				DashboardID: types.StringValue("d1"),
				Label:       types.StringValue("x"),
			},
			URLDrilldown: &models.LensURLDrilldownTFModel{
				URL:     types.StringValue("https://x"),
				Label:   types.StringValue("y"),
				Trigger: types.StringValue("on_click_row"),
			},
		}
		_, diags := lensDrilldownsToRawJSON([]models.LensDrilldownItemTFModel{item})
		require.True(t, diags.HasError())
	})

	t.Run("zero variants set", func(t *testing.T) {
		_, diags := lensDrilldownsToRawJSON([]models.LensDrilldownItemTFModel{{}})
		require.True(t, diags.HasError())
	})
}

func Test_lensDrilldownItem_wireRoundTrip_matchesTFModel(t *testing.T) {
	orig := models.LensDrilldownItemTFModel{
		DashboardDrilldown: &models.LensDashboardDrilldownTFModel{
			DashboardID:  types.StringValue("dash-1"),
			Label:        types.StringValue("Drill"),
			Trigger:      types.StringValue(lensDrilldownTriggerOnApplyFilter),
			UseFilters:   types.BoolValue(true),
			UseTimeRange: types.BoolValue(false),
			OpenInNewTab: types.BoolValue(true),
		},
	}

	raw, diags := lensDrilldownItemToRawJSON(orig, 0)
	require.False(t, diags.HasError())

	got, diags := lensDrilldownItemFromAPIJSON(raw, 0)
	require.False(t, diags.HasError())

	require.NotNil(t, got.DashboardDrilldown)
	assert.Equal(t, orig.DashboardDrilldown.DashboardID, got.DashboardDrilldown.DashboardID)
	assert.Equal(t, orig.DashboardDrilldown.Label, got.DashboardDrilldown.Label)
	assert.Equal(t, orig.DashboardDrilldown.Trigger, got.DashboardDrilldown.Trigger)
	assert.Equal(t, orig.DashboardDrilldown.UseFilters, got.DashboardDrilldown.UseFilters)
	assert.Equal(t, orig.DashboardDrilldown.UseTimeRange, got.DashboardDrilldown.UseTimeRange)
	assert.Equal(t, orig.DashboardDrilldown.OpenInNewTab, got.DashboardDrilldown.OpenInNewTab)
}
