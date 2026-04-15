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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers

func makeSloBurnRateAPIConfig(sloID, duration string, opts ...func(*kbapi.SloBurnRateEmbeddable)) kbapi.SloBurnRateEmbeddable {
	c := kbapi.SloBurnRateEmbeddable{
		SloId:    sloID,
		Duration: duration,
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

func withSloInstanceID(id string) func(*kbapi.SloBurnRateEmbeddable) {
	return func(c *kbapi.SloBurnRateEmbeddable) { c.SloInstanceId = new(id) }
}

func withTitle(t string) func(*kbapi.SloBurnRateEmbeddable) {
	return func(c *kbapi.SloBurnRateEmbeddable) { c.Title = new(t) }
}

func withDescription(d string) func(*kbapi.SloBurnRateEmbeddable) {
	return func(c *kbapi.SloBurnRateEmbeddable) { c.Description = new(d) }
}

func withHideTitle(v bool) func(*kbapi.SloBurnRateEmbeddable) {
	return func(c *kbapi.SloBurnRateEmbeddable) { c.HideTitle = new(v) }
}

func withHideBorder(v bool) func(*kbapi.SloBurnRateEmbeddable) {
	return func(c *kbapi.SloBurnRateEmbeddable) { c.HideBorder = new(v) }
}

// ─────────────────────────────────────────────────────────────────────────────
// buildSloBurnRateConfig tests
// ─────────────────────────────────────────────────────────────────────────────

func Test_buildSloBurnRateConfig_requiredFieldsOnly(t *testing.T) {
	pm := panelModel{
		SloBurnRateConfig: &sloBurnRateConfigModel{
			SloID:         types.StringValue("my-slo-id"),
			Duration:      types.StringValue("72h"),
			SloInstanceID: types.StringNull(),
			Title:         types.StringNull(),
			Description:   types.StringNull(),
			HideTitle:     types.BoolNull(),
			HideBorder:    types.BoolNull(),
		},
	}
	var panel kbapi.KbnDashboardPanelTypeSloBurnRate
	buildSloBurnRateConfig(pm, &panel)

	assert.Equal(t, "my-slo-id", panel.Config.SloId)
	assert.Equal(t, "72h", panel.Config.Duration)
	assert.Nil(t, panel.Config.SloInstanceId)
	assert.Nil(t, panel.Config.Title)
	assert.Nil(t, panel.Config.Description)
	assert.Nil(t, panel.Config.HideTitle)
	assert.Nil(t, panel.Config.HideBorder)
	assert.Nil(t, panel.Config.Drilldowns)
}

func Test_buildSloBurnRateConfig_allOptionalFields(t *testing.T) {
	pm := panelModel{
		SloBurnRateConfig: &sloBurnRateConfigModel{
			SloID:         types.StringValue("my-slo"),
			Duration:      types.StringValue("5m"),
			SloInstanceID: types.StringValue("host-a"),
			Title:         types.StringValue("My Panel"),
			Description:   types.StringValue("Desc"),
			HideTitle:     types.BoolValue(true),
			HideBorder:    types.BoolValue(false),
		},
	}
	var panel kbapi.KbnDashboardPanelTypeSloBurnRate
	buildSloBurnRateConfig(pm, &panel)

	require.NotNil(t, panel.Config.SloInstanceId)
	assert.Equal(t, "host-a", *panel.Config.SloInstanceId)
	require.NotNil(t, panel.Config.Title)
	assert.Equal(t, "My Panel", *panel.Config.Title)
	require.NotNil(t, panel.Config.Description)
	assert.Equal(t, "Desc", *panel.Config.Description)
	require.NotNil(t, panel.Config.HideTitle)
	assert.True(t, *panel.Config.HideTitle)
	require.NotNil(t, panel.Config.HideBorder)
	assert.False(t, *panel.Config.HideBorder)
}

func Test_buildSloBurnRateConfig_withDrilldowns(t *testing.T) {
	pm := panelModel{
		SloBurnRateConfig: &sloBurnRateConfigModel{
			SloID:    types.StringValue("slo-1"),
			Duration: types.StringValue("3h"),
			Drilldowns: []sloBurnRateDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("View details"),
					EncodeURL:    types.BoolNull(),
					OpenInNewTab: types.BoolNull(),
				},
			},
		},
	}
	var panel kbapi.KbnDashboardPanelTypeSloBurnRate
	buildSloBurnRateConfig(pm, &panel)

	require.NotNil(t, panel.Config.Drilldowns)
	require.Len(t, *panel.Config.Drilldowns, 1)
	d := (*panel.Config.Drilldowns)[0]
	assert.Equal(t, "https://example.com", d.Url)
	assert.Equal(t, "View details", d.Label)
	assert.Equal(t, kbapi.SloBurnRateEmbeddableDrilldownsTriggerOnOpenPanelMenu, d.Trigger)
	assert.Equal(t, kbapi.SloBurnRateEmbeddableDrilldownsTypeUrlDrilldown, d.Type)
	assert.Nil(t, d.EncodeUrl)
	assert.Nil(t, d.OpenInNewTab)
}

func Test_buildSloBurnRateConfig_withDrilldowns_optionalBoolsSet(t *testing.T) {
	pm := panelModel{
		SloBurnRateConfig: &sloBurnRateConfigModel{
			SloID:    types.StringValue("slo-1"),
			Duration: types.StringValue("3h"),
			Drilldowns: []sloBurnRateDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("Link"),
					EncodeURL:    types.BoolValue(true),
					OpenInNewTab: types.BoolValue(false),
				},
			},
		},
	}
	var panel kbapi.KbnDashboardPanelTypeSloBurnRate
	buildSloBurnRateConfig(pm, &panel)

	require.NotNil(t, panel.Config.Drilldowns)
	d := (*panel.Config.Drilldowns)[0]
	require.NotNil(t, d.EncodeUrl)
	assert.True(t, *d.EncodeUrl)
	require.NotNil(t, d.OpenInNewTab)
	assert.False(t, *d.OpenInNewTab)
}

func Test_buildSloBurnRateConfig_nilConfig(t *testing.T) {
	pm := panelModel{}
	var panel kbapi.KbnDashboardPanelTypeSloBurnRate
	buildSloBurnRateConfig(pm, &panel)
	// Should be empty/zero config — no panic.
	assert.Empty(t, panel.Config.SloId)
}

// ─────────────────────────────────────────────────────────────────────────────
// populateSloBurnRateFromAPI tests
// ─────────────────────────────────────────────────────────────────────────────

// On import (tfPanel == nil), populate all fields from API.
func Test_populateSloBurnRateFromAPI_import_allFields(t *testing.T) {
	pm := &panelModel{}
	apiCfg := makeSloBurnRateAPIConfig("slo-1", "72h",
		withSloInstanceID("host-a"),
		withTitle("My SLO"),
		withDescription("A description"),
		withHideTitle(true),
		withHideBorder(false),
	)
	populateSloBurnRateFromAPI(pm, nil, apiCfg)

	require.NotNil(t, pm.SloBurnRateConfig)
	cfg := pm.SloBurnRateConfig
	assert.Equal(t, "slo-1", cfg.SloID.ValueString())
	assert.Equal(t, "72h", cfg.Duration.ValueString())
	assert.Equal(t, "host-a", cfg.SloInstanceID.ValueString())
	assert.Equal(t, "My SLO", cfg.Title.ValueString())
	assert.Equal(t, "A description", cfg.Description.ValueString())
	assert.Equal(t, types.BoolValue(true), cfg.HideTitle)
	assert.Equal(t, types.BoolValue(false), cfg.HideBorder)
}

// On import with minimal API response, optional fields are null.
func Test_populateSloBurnRateFromAPI_import_requiredFieldsOnly(t *testing.T) {
	pm := &panelModel{}
	apiCfg := makeSloBurnRateAPIConfig("slo-2", "5m")
	populateSloBurnRateFromAPI(pm, nil, apiCfg)

	require.NotNil(t, pm.SloBurnRateConfig)
	cfg := pm.SloBurnRateConfig
	assert.Equal(t, "slo-2", cfg.SloID.ValueString())
	assert.Equal(t, "5m", cfg.Duration.ValueString())
	assert.True(t, cfg.SloInstanceID.IsNull())
	assert.True(t, cfg.Title.IsNull())
	assert.True(t, cfg.Description.IsNull())
	assert.True(t, cfg.HideTitle.IsNull())
	assert.True(t, cfg.HideBorder.IsNull())
}

// Key: slo_instance_id null-preservation. tfPanel has no slo_instance_id (null), API returns "*".
// Provider must keep null in state (not pollute with API sentinel).
func Test_populateSloBurnRateFromAPI_sloInstanceID_nullPreservation(t *testing.T) {
	existing := &sloBurnRateConfigModel{
		SloID:         types.StringValue("slo-1"),
		Duration:      types.StringValue("72h"),
		SloInstanceID: types.StringNull(), // not configured by practitioner
		Title:         types.StringNull(),
		Description:   types.StringNull(),
		HideTitle:     types.BoolNull(),
		HideBorder:    types.BoolNull(),
	}
	pm := &panelModel{SloBurnRateConfig: existing}
	tfPanel := &panelModel{SloBurnRateConfig: existing}

	// API returns "*" (all-instances sentinel)
	apiCfg := makeSloBurnRateAPIConfig("slo-1", "72h", withSloInstanceID("*"))
	populateSloBurnRateFromAPI(pm, tfPanel, apiCfg)

	require.NotNil(t, pm.SloBurnRateConfig)
	// Must remain null — not updated to "*"
	assert.True(t, pm.SloBurnRateConfig.SloInstanceID.IsNull(), "slo_instance_id should remain null when not configured by practitioner")
}

// When slo_instance_id is explicitly configured, round-trip normally.
func Test_populateSloBurnRateFromAPI_sloInstanceID_explicitValue_roundTrips(t *testing.T) {
	existing := &sloBurnRateConfigModel{
		SloID:         types.StringValue("slo-1"),
		Duration:      types.StringValue("72h"),
		SloInstanceID: types.StringValue("host-a"),
		Title:         types.StringNull(),
		Description:   types.StringNull(),
		HideTitle:     types.BoolNull(),
		HideBorder:    types.BoolNull(),
	}
	pm := &panelModel{SloBurnRateConfig: existing}
	tfPanel := &panelModel{SloBurnRateConfig: existing}

	apiCfg := makeSloBurnRateAPIConfig("slo-1", "72h", withSloInstanceID("host-a"))
	populateSloBurnRateFromAPI(pm, tfPanel, apiCfg)

	assert.Equal(t, "host-a", pm.SloBurnRateConfig.SloInstanceID.ValueString())
}

// When slo_instance_id is explicitly configured to "*", round-trip it normally.
func Test_populateSloBurnRateFromAPI_sloInstanceID_explicitWildcard_roundTrips(t *testing.T) {
	existing := &sloBurnRateConfigModel{
		SloID:         types.StringValue("slo-1"),
		Duration:      types.StringValue("72h"),
		SloInstanceID: types.StringValue("*"),
		Title:         types.StringNull(),
		Description:   types.StringNull(),
		HideTitle:     types.BoolNull(),
		HideBorder:    types.BoolNull(),
	}
	pm := &panelModel{SloBurnRateConfig: existing}
	tfPanel := &panelModel{SloBurnRateConfig: existing}

	apiCfg := makeSloBurnRateAPIConfig("slo-1", "72h", withSloInstanceID("*"))
	populateSloBurnRateFromAPI(pm, tfPanel, apiCfg)

	assert.Equal(t, "*", pm.SloBurnRateConfig.SloInstanceID.ValueString())
}

// When prior state has no config block (nil), preserve nil intent.
func Test_populateSloBurnRateFromAPI_nilBlock_preservesNilIntent(t *testing.T) {
	pm := &panelModel{}
	tfPanel := &panelModel{} // no SloBurnRateConfig

	apiCfg := makeSloBurnRateAPIConfig("slo-1", "72h")
	populateSloBurnRateFromAPI(pm, tfPanel, apiCfg)

	assert.Nil(t, pm.SloBurnRateConfig, "SloBurnRateConfig should remain nil when prior state had no config block")
}

// Required fields (slo_id, duration) are always updated from API response.
func Test_populateSloBurnRateFromAPI_requiredFieldsAlwaysUpdated(t *testing.T) {
	existing := &sloBurnRateConfigModel{
		SloID:         types.StringValue("old-slo"),
		Duration:      types.StringValue("1h"),
		SloInstanceID: types.StringNull(),
		Title:         types.StringNull(),
		Description:   types.StringNull(),
		HideTitle:     types.BoolNull(),
		HideBorder:    types.BoolNull(),
	}
	pm := &panelModel{SloBurnRateConfig: existing}
	tfPanel := &panelModel{SloBurnRateConfig: existing}

	apiCfg := makeSloBurnRateAPIConfig("new-slo", "24h")
	populateSloBurnRateFromAPI(pm, tfPanel, apiCfg)

	assert.Equal(t, "new-slo", pm.SloBurnRateConfig.SloID.ValueString())
	assert.Equal(t, "24h", pm.SloBurnRateConfig.Duration.ValueString())
}

// Drilldown optional bool null-preservation: if encode_url / open_in_new_tab were null in
// prior state and API returns a value, preserve null.
func Test_populateSloBurnRateFromAPI_drilldowns_optionalBoolNullPreservation(t *testing.T) {
	existing := &sloBurnRateConfigModel{
		SloID:    types.StringValue("slo-1"),
		Duration: types.StringValue("6d"),
		Drilldowns: []sloBurnRateDrilldownModel{
			{
				URL:          types.StringValue("https://example.com"),
				Label:        types.StringValue("View"),
				EncodeURL:    types.BoolNull(), // not configured
				OpenInNewTab: types.BoolNull(), // not configured
			},
		},
	}
	pm := &panelModel{SloBurnRateConfig: existing}
	tfPanel := &panelModel{SloBurnRateConfig: existing}

	// API returns drilldown with encode_url and open_in_new_tab set to true
	apiDrilldowns := &[]struct {
		EncodeUrl    *bool                                        `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                       `json:"label"`
		OpenInNewTab *bool                                        `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.SloBurnRateEmbeddableDrilldownsTrigger `json:"trigger"`
		Type         kbapi.SloBurnRateEmbeddableDrilldownsType    `json:"type"`
		Url          string                                       `json:"url"` //nolint:revive
	}{
		{
			Url:          "https://example.com",
			Label:        "View",
			Trigger:      kbapi.SloBurnRateEmbeddableDrilldownsTriggerOnOpenPanelMenu,
			Type:         kbapi.SloBurnRateEmbeddableDrilldownsTypeUrlDrilldown,
			EncodeUrl:    new(true),
			OpenInNewTab: new(true),
		},
	}
	apiCfg := makeSloBurnRateAPIConfig("slo-1", "6d")
	apiCfg.Drilldowns = apiDrilldowns

	populateSloBurnRateFromAPI(pm, tfPanel, apiCfg)

	require.Len(t, pm.SloBurnRateConfig.Drilldowns, 1)
	d := pm.SloBurnRateConfig.Drilldowns[0]
	assert.True(t, d.EncodeURL.IsNull(), "encode_url should remain null when not configured by practitioner")
	assert.True(t, d.OpenInNewTab.IsNull(), "open_in_new_tab should remain null when not configured by practitioner")
}

// When drilldown optional bools were explicitly set in prior state, round-trip from API.
func Test_populateSloBurnRateFromAPI_drilldowns_optionalBoolsExplicit_roundTrip(t *testing.T) {
	existing := &sloBurnRateConfigModel{
		SloID:    types.StringValue("slo-1"),
		Duration: types.StringValue("6d"),
		Drilldowns: []sloBurnRateDrilldownModel{
			{
				URL:          types.StringValue("https://example.com"),
				Label:        types.StringValue("View"),
				EncodeURL:    types.BoolValue(true),
				OpenInNewTab: types.BoolValue(false),
			},
		},
	}
	pm := &panelModel{SloBurnRateConfig: existing}
	tfPanel := &panelModel{SloBurnRateConfig: existing}

	apiDrilldowns := &[]struct {
		EncodeUrl    *bool                                        `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                       `json:"label"`
		OpenInNewTab *bool                                        `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.SloBurnRateEmbeddableDrilldownsTrigger `json:"trigger"`
		Type         kbapi.SloBurnRateEmbeddableDrilldownsType    `json:"type"`
		Url          string                                       `json:"url"` //nolint:revive
	}{
		{
			Url:          "https://example.com",
			Label:        "View",
			Trigger:      kbapi.SloBurnRateEmbeddableDrilldownsTriggerOnOpenPanelMenu,
			Type:         kbapi.SloBurnRateEmbeddableDrilldownsTypeUrlDrilldown,
			EncodeUrl:    new(true),
			OpenInNewTab: new(false),
		},
	}
	apiCfg := makeSloBurnRateAPIConfig("slo-1", "6d")
	apiCfg.Drilldowns = apiDrilldowns

	populateSloBurnRateFromAPI(pm, tfPanel, apiCfg)

	require.Len(t, pm.SloBurnRateConfig.Drilldowns, 1)
	d := pm.SloBurnRateConfig.Drilldowns[0]
	assert.Equal(t, types.BoolValue(true), d.EncodeURL)
	assert.Equal(t, types.BoolValue(false), d.OpenInNewTab)
}
