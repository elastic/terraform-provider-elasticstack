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

// ─────────────────────────────────────────────────────────────────────────────
// buildSyntheticsStatsOverviewConfig (write converter) tests
// ─────────────────────────────────────────────────────────────────────────────

func Test_buildSyntheticsStatsOverviewConfig_nilConfig(t *testing.T) {
	pm := panelModel{}
	var panel kbapi.KbnDashboardPanelSyntheticsStatsOverview
	buildSyntheticsStatsOverviewConfig(pm, &panel)
	// Zero config — no panic, nothing set.
	assert.Nil(t, panel.Config.Title)
	assert.Nil(t, panel.Config.Description)
	assert.Nil(t, panel.Config.HideTitle)
	assert.Nil(t, panel.Config.HideBorder)
	assert.Nil(t, panel.Config.Drilldowns)
	assert.Nil(t, panel.Config.Filters)
}

func Test_buildSyntheticsStatsOverviewConfig_emptyConfig(t *testing.T) {
	pm := panelModel{
		SyntheticsStatsOverviewConfig: &syntheticsStatsOverviewConfigModel{
			Title:       types.StringNull(),
			Description: types.StringNull(),
			HideTitle:   types.BoolNull(),
			HideBorder:  types.BoolNull(),
			Drilldowns:  nil,
			Filters:     nil,
		},
	}
	var panel kbapi.KbnDashboardPanelSyntheticsStatsOverview
	buildSyntheticsStatsOverviewConfig(pm, &panel)

	assert.Nil(t, panel.Config.Title)
	assert.Nil(t, panel.Config.Description)
	assert.Nil(t, panel.Config.HideTitle)
	assert.Nil(t, panel.Config.HideBorder)
	assert.Nil(t, panel.Config.Drilldowns)
	assert.Nil(t, panel.Config.Filters)
}

func Test_buildSyntheticsStatsOverviewConfig_displaySettings(t *testing.T) {
	pm := panelModel{
		SyntheticsStatsOverviewConfig: &syntheticsStatsOverviewConfigModel{
			Title:       types.StringValue("My Panel"),
			Description: types.StringValue("A description"),
			HideTitle:   types.BoolValue(true),
			HideBorder:  types.BoolValue(false),
		},
	}
	var panel kbapi.KbnDashboardPanelSyntheticsStatsOverview
	buildSyntheticsStatsOverviewConfig(pm, &panel)

	require.NotNil(t, panel.Config.Title)
	assert.Equal(t, "My Panel", *panel.Config.Title)
	require.NotNil(t, panel.Config.Description)
	assert.Equal(t, "A description", *panel.Config.Description)
	require.NotNil(t, panel.Config.HideTitle)
	assert.True(t, *panel.Config.HideTitle)
	require.NotNil(t, panel.Config.HideBorder)
	assert.False(t, *panel.Config.HideBorder)
}

func Test_buildSyntheticsStatsOverviewConfig_withDrilldowns(t *testing.T) {
	pm := panelModel{
		SyntheticsStatsOverviewConfig: &syntheticsStatsOverviewConfigModel{
			Title: types.StringNull(),
			Drilldowns: []syntheticsStatsOverviewDrilldownModel{
				{
					URL:          types.StringValue("https://example.com/{{context.panel.title}}"),
					Label:        types.StringValue("View details"),
					Trigger:      types.StringValue("on_open_panel_menu"),
					Type:         types.StringValue("url_drilldown"),
					EncodeURL:    types.BoolNull(),
					OpenInNewTab: types.BoolNull(),
				},
			},
		},
	}
	var panel kbapi.KbnDashboardPanelSyntheticsStatsOverview
	buildSyntheticsStatsOverviewConfig(pm, &panel)

	require.NotNil(t, panel.Config.Drilldowns)
	require.Len(t, *panel.Config.Drilldowns, 1)
	d := (*panel.Config.Drilldowns)[0]
	assert.Equal(t, "https://example.com/{{context.panel.title}}", d.Url)
	assert.Equal(t, "View details", d.Label)
	assert.Equal(t, kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTriggerOnOpenPanelMenu, d.Trigger)
	assert.Equal(t, kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTypeUrlDrilldown, d.Type)
	assert.Nil(t, d.EncodeUrl)
	assert.Nil(t, d.OpenInNewTab)
}

func Test_buildSyntheticsStatsOverviewConfig_withDrilldowns_optionalBoolsSet(t *testing.T) {
	pm := panelModel{
		SyntheticsStatsOverviewConfig: &syntheticsStatsOverviewConfigModel{
			Drilldowns: []syntheticsStatsOverviewDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("Link"),
					Trigger:      types.StringValue("on_open_panel_menu"),
					Type:         types.StringValue("url_drilldown"),
					EncodeURL:    types.BoolValue(true),
					OpenInNewTab: types.BoolValue(false),
				},
			},
		},
	}
	var panel kbapi.KbnDashboardPanelSyntheticsStatsOverview
	buildSyntheticsStatsOverviewConfig(pm, &panel)

	require.NotNil(t, panel.Config.Drilldowns)
	d := (*panel.Config.Drilldowns)[0]
	require.NotNil(t, d.EncodeUrl)
	assert.True(t, *d.EncodeUrl)
	require.NotNil(t, d.OpenInNewTab)
	assert.False(t, *d.OpenInNewTab)
}

func Test_buildSyntheticsStatsOverviewConfig_withFilters(t *testing.T) {
	pm := panelModel{
		SyntheticsStatsOverviewConfig: &syntheticsStatsOverviewConfigModel{
			Filters: &syntheticsStatsOverviewFiltersModel{
				Projects: []syntheticsFilterItemModel{
					{Label: types.StringValue("My Project"), Value: types.StringValue("my-project")},
				},
				Tags: []syntheticsFilterItemModel{
					{Label: types.StringValue("prod"), Value: types.StringValue("prod")},
				},
				MonitorTypes: []syntheticsFilterItemModel{
					{Label: types.StringValue("HTTP"), Value: types.StringValue("http")},
				},
			},
		},
	}
	var panel kbapi.KbnDashboardPanelSyntheticsStatsOverview
	buildSyntheticsStatsOverviewConfig(pm, &panel)

	require.NotNil(t, panel.Config.Filters)
	require.NotNil(t, panel.Config.Filters.Projects)
	require.Len(t, *panel.Config.Filters.Projects, 1)
	assert.Equal(t, "My Project", (*panel.Config.Filters.Projects)[0].Label)
	assert.Equal(t, "my-project", (*panel.Config.Filters.Projects)[0].Value)
	require.NotNil(t, panel.Config.Filters.Tags)
	require.Len(t, *panel.Config.Filters.Tags, 1)
	assert.Equal(t, "prod", (*panel.Config.Filters.Tags)[0].Value)
	require.NotNil(t, panel.Config.Filters.MonitorTypes)
	require.Len(t, *panel.Config.Filters.MonitorTypes, 1)
	assert.Equal(t, "http", (*panel.Config.Filters.MonitorTypes)[0].Value)
}

func Test_buildSyntheticsStatsOverviewConfig_emptyFilters_notSent(t *testing.T) {
	// A filters block with no entries should not produce a filters payload.
	pm := panelModel{
		SyntheticsStatsOverviewConfig: &syntheticsStatsOverviewConfigModel{
			Filters: &syntheticsStatsOverviewFiltersModel{
				Projects: nil,
				Tags:     nil,
			},
		},
	}
	var panel kbapi.KbnDashboardPanelSyntheticsStatsOverview
	buildSyntheticsStatsOverviewConfig(pm, &panel)
	assert.Nil(t, panel.Config.Filters)
}

// ─────────────────────────────────────────────────────────────────────────────
// populateSyntheticsStatsOverviewFromAPI (read converter) tests
// ─────────────────────────────────────────────────────────────────────────────

// makeSyntheticsAPIConfig builds a minimal API config for test use.
func makeSyntheticsAPIConfig() kbapi.KbnDashboardPanelSyntheticsStatsOverview {
	return kbapi.KbnDashboardPanelSyntheticsStatsOverview{}
}

// Test: on import (tfPanel == nil) with empty config — block stays null.
func Test_populateSyntheticsStatsOverviewFromAPI_import_emptyConfig_blockIsNull(t *testing.T) {
	pm := &panelModel{}
	panel := makeSyntheticsAPIConfig()
	populateSyntheticsStatsOverviewFromAPI(pm, nil, panel.Config)

	assert.Nil(t, pm.SyntheticsStatsOverviewConfig, "block should be null when API config is empty on import")
}

// Test: on import with fields set — block is populated.
func Test_populateSyntheticsStatsOverviewFromAPI_import_withFields(t *testing.T) {
	pm := &panelModel{}
	panel := makeSyntheticsAPIConfig()
	title := "My Panel"
	panel.Config.Title = &title
	desc := "My desc"
	panel.Config.Description = &desc
	hideTitle := true
	panel.Config.HideTitle = &hideTitle
	hideBorder := false
	panel.Config.HideBorder = &hideBorder

	populateSyntheticsStatsOverviewFromAPI(pm, nil, panel.Config)

	require.NotNil(t, pm.SyntheticsStatsOverviewConfig)
	cfg := pm.SyntheticsStatsOverviewConfig
	assert.Equal(t, "My Panel", cfg.Title.ValueString())
	assert.Equal(t, "My desc", cfg.Description.ValueString())
	assert.Equal(t, types.BoolValue(true), cfg.HideTitle)
	assert.Equal(t, types.BoolValue(false), cfg.HideBorder)
}

// Test: prior state has no block (nil) — nil intent preserved.
func Test_populateSyntheticsStatsOverviewFromAPI_nilBlock_preservesNilIntent(t *testing.T) {
	pm := &panelModel{}
	tfPanel := &panelModel{} // no SyntheticsStatsOverviewConfig

	panel := makeSyntheticsAPIConfig()
	title := "Should not appear"
	panel.Config.Title = &title

	populateSyntheticsStatsOverviewFromAPI(pm, tfPanel, panel.Config)

	assert.Nil(t, pm.SyntheticsStatsOverviewConfig, "block should remain nil when prior state had no config block")
}

// Test: null-preservation for optional string fields.
func Test_populateSyntheticsStatsOverviewFromAPI_nullPreservation_strings(t *testing.T) {
	existing := &syntheticsStatsOverviewConfigModel{
		Title:       types.StringNull(),
		Description: types.StringNull(),
		HideTitle:   types.BoolNull(),
		HideBorder:  types.BoolNull(),
	}
	pm := &panelModel{SyntheticsStatsOverviewConfig: existing}
	tfPanel := &panelModel{SyntheticsStatsOverviewConfig: existing}

	panel := makeSyntheticsAPIConfig()
	title := "API title"
	panel.Config.Title = &title

	populateSyntheticsStatsOverviewFromAPI(pm, tfPanel, panel.Config)

	require.NotNil(t, pm.SyntheticsStatsOverviewConfig)
	// title was null in prior state — preserve null even though API returned a value.
	assert.True(t, pm.SyntheticsStatsOverviewConfig.Title.IsNull(),
		"title should stay null when not configured by practitioner")
}

// Test: when fields are explicitly set in prior state, round-trip them from API.
func Test_populateSyntheticsStatsOverviewFromAPI_explicitFields_roundTrip(t *testing.T) {
	existing := &syntheticsStatsOverviewConfigModel{
		Title:       types.StringValue("Old Title"),
		Description: types.StringValue("Old desc"),
		HideTitle:   types.BoolValue(false),
		HideBorder:  types.BoolValue(true),
	}
	pm := &panelModel{SyntheticsStatsOverviewConfig: existing}
	tfPanel := &panelModel{SyntheticsStatsOverviewConfig: existing}

	panel := makeSyntheticsAPIConfig()
	title := "New Title"
	panel.Config.Title = &title
	desc := "New desc"
	panel.Config.Description = &desc
	hideTitle := true
	panel.Config.HideTitle = &hideTitle
	hideBorder := false
	panel.Config.HideBorder = &hideBorder

	populateSyntheticsStatsOverviewFromAPI(pm, tfPanel, panel.Config)

	require.NotNil(t, pm.SyntheticsStatsOverviewConfig)
	cfg := pm.SyntheticsStatsOverviewConfig
	assert.Equal(t, "New Title", cfg.Title.ValueString())
	assert.Equal(t, "New desc", cfg.Description.ValueString())
	assert.Equal(t, types.BoolValue(true), cfg.HideTitle)
	assert.Equal(t, types.BoolValue(false), cfg.HideBorder)
}

// Test: drilldown optional bool null-preservation.
func Test_populateSyntheticsStatsOverviewFromAPI_drilldowns_nullPreservation(t *testing.T) {
	existing := &syntheticsStatsOverviewConfigModel{
		Drilldowns: []syntheticsStatsOverviewDrilldownModel{
			{
				URL:          types.StringValue("https://example.com"),
				Label:        types.StringValue("View"),
				Trigger:      types.StringValue("on_open_panel_menu"),
				Type:         types.StringValue("url_drilldown"),
				EncodeURL:    types.BoolNull(),
				OpenInNewTab: types.BoolNull(),
			},
		},
	}
	pm := &panelModel{SyntheticsStatsOverviewConfig: existing}
	tfPanel := &panelModel{SyntheticsStatsOverviewConfig: existing}

	panel := makeSyntheticsAPIConfig()
	encodeURL := true
	openInNewTab := true
	panel.Config.Drilldowns = &[]struct {
		EncodeUrl    *bool                                                                        `json:"encode_url,omitempty"`
		Label        string                                                                       `json:"label"`
		OpenInNewTab *bool                                                                        `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTrigger       `json:"trigger"`
		Type         kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsType          `json:"type"`
		Url          string                                                                       `json:"url"`
	}{
		{
			Url:          "https://example.com",
			Label:        "View",
			Trigger:      kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTriggerOnOpenPanelMenu,
			Type:         kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTypeUrlDrilldown,
			EncodeUrl:    &encodeURL,
			OpenInNewTab: &openInNewTab,
		},
	}

	populateSyntheticsStatsOverviewFromAPI(pm, tfPanel, panel.Config)

	require.NotNil(t, pm.SyntheticsStatsOverviewConfig)
	require.Len(t, pm.SyntheticsStatsOverviewConfig.Drilldowns, 1)
	d := pm.SyntheticsStatsOverviewConfig.Drilldowns[0]
	assert.True(t, d.EncodeURL.IsNull(), "encode_url should remain null when not configured by practitioner")
	assert.True(t, d.OpenInNewTab.IsNull(), "open_in_new_tab should remain null when not configured by practitioner")
}

// Test: empty API filters treated as absent block.
func Test_populateSyntheticsStatsOverviewFromAPI_emptyFilters_treatedAsAbsent(t *testing.T) {
	pm := &panelModel{}

	panel := makeSyntheticsAPIConfig()
	// Set title so import path fires, then provide empty filters.
	title := "Panel"
	panel.Config.Title = &title
	panel.Config.Filters = &struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"monitor_ids,omitempty"`
		MonitorTypes *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"monitor_types,omitempty"`
		Projects *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"projects,omitempty"`
		Tags *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"tags,omitempty"`
	}{} // all nil categories

	populateSyntheticsStatsOverviewFromAPI(pm, nil, panel.Config)

	require.NotNil(t, pm.SyntheticsStatsOverviewConfig)
	assert.Nil(t, pm.SyntheticsStatsOverviewConfig.Filters, "empty filters should not populate the filters block")
}

// Test: import with filters populated.
func Test_populateSyntheticsStatsOverviewFromAPI_import_withFilters(t *testing.T) {
	pm := &panelModel{}

	panel := makeSyntheticsAPIConfig()
	projects := []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}{
		{Label: "My Project", Value: "my-project"},
	}
	panel.Config.Filters = &struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"monitor_ids,omitempty"`
		MonitorTypes *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"monitor_types,omitempty"`
		Projects *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"projects,omitempty"`
		Tags *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"tags,omitempty"`
	}{
		Projects: &projects,
	}

	populateSyntheticsStatsOverviewFromAPI(pm, nil, panel.Config)

	require.NotNil(t, pm.SyntheticsStatsOverviewConfig)
	require.NotNil(t, pm.SyntheticsStatsOverviewConfig.Filters)
	require.Len(t, pm.SyntheticsStatsOverviewConfig.Filters.Projects, 1)
	assert.Equal(t, "My Project", pm.SyntheticsStatsOverviewConfig.Filters.Projects[0].Label.ValueString())
	assert.Equal(t, "my-project", pm.SyntheticsStatsOverviewConfig.Filters.Projects[0].Value.ValueString())
}
