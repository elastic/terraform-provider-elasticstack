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

func makeSloErrorBudgetAPIConfig(opts ...func(*kbapi.SloErrorBudgetEmbeddable)) kbapi.SloErrorBudgetEmbeddable {
	cfg := kbapi.SloErrorBudgetEmbeddable{SloId: "my-slo-id"}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

func withSloInstanceID(id string) func(*kbapi.SloErrorBudgetEmbeddable) {
	return func(c *kbapi.SloErrorBudgetEmbeddable) { c.SloInstanceId = new(id) }
}

func withSloTitle(t string) func(*kbapi.SloErrorBudgetEmbeddable) {
	return func(c *kbapi.SloErrorBudgetEmbeddable) { c.Title = new(t) }
}

func withSloDescription(d string) func(*kbapi.SloErrorBudgetEmbeddable) {
	return func(c *kbapi.SloErrorBudgetEmbeddable) { c.Description = new(d) }
}

func withHideTitle(v bool) func(*kbapi.SloErrorBudgetEmbeddable) {
	return func(c *kbapi.SloErrorBudgetEmbeddable) { c.HideTitle = new(v) }
}

func withHideBorder(v bool) func(*kbapi.SloErrorBudgetEmbeddable) {
	return func(c *kbapi.SloErrorBudgetEmbeddable) { c.HideBorder = new(v) }
}

func withSloDrilldown(url, label string, encodeURL, openInNewTab *bool) func(*kbapi.SloErrorBudgetEmbeddable) {
	return func(c *kbapi.SloErrorBudgetEmbeddable) {
		d := struct {
			EncodeUrl    *bool                                           `json:"encode_url,omitempty"` //nolint:revive
			Label        string                                          `json:"label"`
			OpenInNewTab *bool                                           `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.SloErrorBudgetEmbeddableDrilldownsTrigger `json:"trigger"`
			Type         kbapi.SloErrorBudgetEmbeddableDrilldownsType    `json:"type"`
			Url          string                                          `json:"url"` //nolint:revive
		}{
			Url:          url,
			Label:        label,
			Trigger:      kbapi.SloErrorBudgetEmbeddableDrilldownsTriggerOnOpenPanelMenu,
			Type:         kbapi.SloErrorBudgetEmbeddableDrilldownsTypeUrlDrilldown,
			EncodeUrl:    encodeURL,
			OpenInNewTab: openInNewTab,
		}
		if c.Drilldowns == nil {
			c.Drilldowns = &[]struct {
				EncodeUrl    *bool                                           `json:"encode_url,omitempty"` //nolint:revive
				Label        string                                          `json:"label"`
				OpenInNewTab *bool                                           `json:"open_in_new_tab,omitempty"`
				Trigger      kbapi.SloErrorBudgetEmbeddableDrilldownsTrigger `json:"trigger"`
				Type         kbapi.SloErrorBudgetEmbeddableDrilldownsType    `json:"type"`
				Url          string                                          `json:"url"` //nolint:revive
			}{}
		}
		*c.Drilldowns = append(*c.Drilldowns, d)
	}
}

// ---- buildSloErrorBudgetConfig ----

func Test_buildSloErrorBudgetConfig_minimal(t *testing.T) {
	pm := panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue("my-slo-id"),
		},
	}
	var sebPanel kbapi.KbnDashboardPanelSloErrorBudget
	buildSloErrorBudgetConfig(pm, &sebPanel)
	assert.Equal(t, "my-slo-id", sebPanel.Config.SloId)
	assert.Nil(t, sebPanel.Config.SloInstanceId)
	assert.Nil(t, sebPanel.Config.Title)
	assert.Nil(t, sebPanel.Config.Description)
	assert.Nil(t, sebPanel.Config.HideTitle)
	assert.Nil(t, sebPanel.Config.HideBorder)
	assert.Nil(t, sebPanel.Config.Drilldowns)
}

func Test_buildSloErrorBudgetConfig_withSloInstanceID(t *testing.T) {
	pm := panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID:         types.StringValue("my-slo-id"),
			SloInstanceID: types.StringValue("my-instance"),
		},
	}
	var sebPanel kbapi.KbnDashboardPanelSloErrorBudget
	buildSloErrorBudgetConfig(pm, &sebPanel)
	require.NotNil(t, sebPanel.Config.SloInstanceId)
	assert.Equal(t, "my-instance", *sebPanel.Config.SloInstanceId)
}

func Test_buildSloErrorBudgetConfig_nullSloInstanceID(t *testing.T) {
	pm := panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID:         types.StringValue("my-slo-id"),
			SloInstanceID: types.StringNull(),
		},
	}
	var sebPanel kbapi.KbnDashboardPanelSloErrorBudget
	buildSloErrorBudgetConfig(pm, &sebPanel)
	assert.Nil(t, sebPanel.Config.SloInstanceId)
}

func Test_buildSloErrorBudgetConfig_withDisplayFields(t *testing.T) {
	pm := panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID:       types.StringValue("my-slo-id"),
			Title:       types.StringValue("My Title"),
			Description: types.StringValue("My Description"),
			HideTitle:   types.BoolValue(true),
			HideBorder:  types.BoolValue(false),
		},
	}
	var sebPanel kbapi.KbnDashboardPanelSloErrorBudget
	buildSloErrorBudgetConfig(pm, &sebPanel)
	require.NotNil(t, sebPanel.Config.Title)
	assert.Equal(t, "My Title", *sebPanel.Config.Title)
	require.NotNil(t, sebPanel.Config.Description)
	assert.Equal(t, "My Description", *sebPanel.Config.Description)
	require.NotNil(t, sebPanel.Config.HideTitle)
	assert.True(t, *sebPanel.Config.HideTitle)
	require.NotNil(t, sebPanel.Config.HideBorder)
	assert.False(t, *sebPanel.Config.HideBorder)
}

func Test_buildSloErrorBudgetConfig_withDrilldowns(t *testing.T) {
	pm := panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue("my-slo-id"),
			Drilldowns: []sloErrorBudgetDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("Open in example"),
					EncodeURL:    types.BoolValue(true),
					OpenInNewTab: types.BoolValue(false),
				},
			},
		},
	}
	var sebPanel kbapi.KbnDashboardPanelSloErrorBudget
	buildSloErrorBudgetConfig(pm, &sebPanel)
	require.NotNil(t, sebPanel.Config.Drilldowns)
	require.Len(t, *sebPanel.Config.Drilldowns, 1)
	d := (*sebPanel.Config.Drilldowns)[0]
	assert.Equal(t, "https://example.com", d.Url)
	assert.Equal(t, "Open in example", d.Label)
	assert.Equal(t, kbapi.SloErrorBudgetEmbeddableDrilldownsTriggerOnOpenPanelMenu, d.Trigger)
	assert.Equal(t, kbapi.SloErrorBudgetEmbeddableDrilldownsTypeUrlDrilldown, d.Type)
	require.NotNil(t, d.EncodeUrl)
	assert.True(t, *d.EncodeUrl)
	require.NotNil(t, d.OpenInNewTab)
	assert.False(t, *d.OpenInNewTab)
}

func Test_buildSloErrorBudgetConfig_drilldownsWithNullOptionalBools(t *testing.T) {
	pm := panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue("my-slo-id"),
			Drilldowns: []sloErrorBudgetDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("Go"),
					EncodeURL:    types.BoolNull(),
					OpenInNewTab: types.BoolNull(),
				},
			},
		},
	}
	var sebPanel kbapi.KbnDashboardPanelSloErrorBudget
	buildSloErrorBudgetConfig(pm, &sebPanel)
	require.NotNil(t, sebPanel.Config.Drilldowns)
	d := (*sebPanel.Config.Drilldowns)[0]
	assert.Nil(t, d.EncodeUrl)
	assert.Nil(t, d.OpenInNewTab)
}

// ---- populateSloErrorBudgetFromAPI ----

func Test_populateSloErrorBudgetFromAPI_minimal(t *testing.T) {
	pm := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue(""),
		},
	}
	tfPanel := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue(""),
		},
	}
	apiCfg := makeSloErrorBudgetAPIConfig()
	populateSloErrorBudgetFromAPI(pm, tfPanel, apiCfg)
	require.NotNil(t, pm.SloErrorBudgetConfig)
	assert.Equal(t, "my-slo-id", pm.SloErrorBudgetConfig.SloID.ValueString())
}

func Test_populateSloErrorBudgetFromAPI_sloInstanceID_nullPreservation(t *testing.T) {
	// Prior state had slo_instance_id == null; API returns "*". Should remain null.
	pm := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID:         types.StringValue(""),
			SloInstanceID: types.StringNull(),
		},
	}
	tfPanel := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID:         types.StringValue(""),
			SloInstanceID: types.StringNull(),
		},
	}
	apiCfg := makeSloErrorBudgetAPIConfig(withSloInstanceID("*"))
	populateSloErrorBudgetFromAPI(pm, tfPanel, apiCfg)
	require.NotNil(t, pm.SloErrorBudgetConfig)
	assert.True(t, pm.SloErrorBudgetConfig.SloInstanceID.IsNull(), "slo_instance_id should remain null")
}

func Test_populateSloErrorBudgetFromAPI_sloInstanceID_writtenWhenKnown(t *testing.T) {
	// Prior state had slo_instance_id set; API returns a value. Should be updated.
	pm := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID:         types.StringValue(""),
			SloInstanceID: types.StringValue("old-instance"),
		},
	}
	tfPanel := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID:         types.StringValue(""),
			SloInstanceID: types.StringValue("old-instance"),
		},
	}
	apiCfg := makeSloErrorBudgetAPIConfig(withSloInstanceID("new-instance"))
	populateSloErrorBudgetFromAPI(pm, tfPanel, apiCfg)
	assert.Equal(t, "new-instance", pm.SloErrorBudgetConfig.SloInstanceID.ValueString())
}

func Test_populateSloErrorBudgetFromAPI_import_populatesAll(t *testing.T) {
	// tfPanel == nil means import. Should populate all API-returned fields.
	pm := &panelModel{}
	apiCfg := makeSloErrorBudgetAPIConfig(
		withSloInstanceID("my-instance"),
		withSloTitle("My Title"),
		withSloDescription("My Desc"),
		withHideTitle(true),
		withHideBorder(false),
	)
	populateSloErrorBudgetFromAPI(pm, nil, apiCfg)
	require.NotNil(t, pm.SloErrorBudgetConfig)
	assert.Equal(t, "my-slo-id", pm.SloErrorBudgetConfig.SloID.ValueString())
	assert.Equal(t, "my-instance", pm.SloErrorBudgetConfig.SloInstanceID.ValueString())
	assert.Equal(t, "My Title", pm.SloErrorBudgetConfig.Title.ValueString())
	assert.Equal(t, "My Desc", pm.SloErrorBudgetConfig.Description.ValueString())
	assert.True(t, pm.SloErrorBudgetConfig.HideTitle.ValueBool())
	assert.False(t, pm.SloErrorBudgetConfig.HideBorder.ValueBool())
}

func Test_populateSloErrorBudgetFromAPI_nilPriorBlock_preservesNil(t *testing.T) {
	// Prior state had no config block (nil). Should not create one.
	pm := &panelModel{}
	tfPanel := &panelModel{} // SloErrorBudgetConfig is nil
	apiCfg := makeSloErrorBudgetAPIConfig()
	populateSloErrorBudgetFromAPI(pm, tfPanel, apiCfg)
	assert.Nil(t, pm.SloErrorBudgetConfig)
}

func Test_populateSloErrorBudgetFromAPI_drilldowns_roundTrip(t *testing.T) {
	pm := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue(""),
			Drilldowns: []sloErrorBudgetDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("Go"),
					EncodeURL:    types.BoolNull(),
					OpenInNewTab: types.BoolNull(),
				},
			},
		},
	}
	tfPanel := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue(""),
			Drilldowns: []sloErrorBudgetDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("Go"),
					EncodeURL:    types.BoolNull(), // omitted by practitioner
					OpenInNewTab: types.BoolNull(), // omitted by practitioner
				},
			},
		},
	}
	// Kibana returns default true for encode_url and open_in_new_tab
	apiCfg := makeSloErrorBudgetAPIConfig(
		withSloDrilldown("https://example.com", "Go", new(true), new(true)),
	)
	populateSloErrorBudgetFromAPI(pm, tfPanel, apiCfg)
	require.NotNil(t, pm.SloErrorBudgetConfig)
	require.Len(t, pm.SloErrorBudgetConfig.Drilldowns, 1)
	d := pm.SloErrorBudgetConfig.Drilldowns[0]
	assert.Equal(t, "https://example.com", d.URL.ValueString())
	assert.Equal(t, "Go", d.Label.ValueString())
	// encode_url and open_in_new_tab were null in prior state; API returned true (default).
	// They should remain null (no drift).
	assert.True(t, d.EncodeURL.IsNull(), "encode_url should remain null (API default normalization)")
	assert.True(t, d.OpenInNewTab.IsNull(), "open_in_new_tab should remain null (API default normalization)")
}

func Test_populateSloErrorBudgetFromAPI_drilldowns_falseValueWritten(t *testing.T) {
	// If API returns false for encode_url/open_in_new_tab (non-default), it should be written.
	pm := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue(""),
			Drilldowns: []sloErrorBudgetDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("Go"),
					EncodeURL:    types.BoolNull(),
					OpenInNewTab: types.BoolNull(),
				},
			},
		},
	}
	tfPanel := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue(""),
			Drilldowns: []sloErrorBudgetDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("Go"),
					EncodeURL:    types.BoolNull(),
					OpenInNewTab: types.BoolNull(),
				},
			},
		},
	}
	// API returns false for both (non-default)
	apiCfg := makeSloErrorBudgetAPIConfig(
		withSloDrilldown("https://example.com", "Go", new(false), new(false)),
	)
	populateSloErrorBudgetFromAPI(pm, tfPanel, apiCfg)
	d := pm.SloErrorBudgetConfig.Drilldowns[0]
	// false is non-default, so it should be written even when prior state was null
	assert.False(t, d.EncodeURL.IsNull(), "encode_url false should be written")
	assert.False(t, d.EncodeURL.ValueBool())
	assert.False(t, d.OpenInNewTab.IsNull(), "open_in_new_tab false should be written")
	assert.False(t, d.OpenInNewTab.ValueBool())
}

func Test_populateSloErrorBudgetFromAPI_drilldowns_knownEncodeURLUpdated(t *testing.T) {
	// If prior state had encode_url = true (explicitly set), API update should be written.
	pm := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue(""),
			Drilldowns: []sloErrorBudgetDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("Go"),
					EncodeURL:    types.BoolValue(true),
					OpenInNewTab: types.BoolValue(true),
				},
			},
		},
	}
	tfPanel := &panelModel{
		SloErrorBudgetConfig: &sloErrorBudgetConfigModel{
			SloID: types.StringValue(""),
			Drilldowns: []sloErrorBudgetDrilldownModel{
				{
					URL:          types.StringValue("https://example.com"),
					Label:        types.StringValue("Go"),
					EncodeURL:    types.BoolValue(true),
					OpenInNewTab: types.BoolValue(true),
				},
			},
		},
	}
	apiCfg := makeSloErrorBudgetAPIConfig(
		withSloDrilldown("https://example.com", "Go", new(true), new(true)),
	)
	populateSloErrorBudgetFromAPI(pm, tfPanel, apiCfg)
	d := pm.SloErrorBudgetConfig.Drilldowns[0]
	assert.False(t, d.EncodeURL.IsNull())
	assert.True(t, d.EncodeURL.ValueBool())
	assert.False(t, d.OpenInNewTab.IsNull())
	assert.True(t, d.OpenInNewTab.ValueBool())
}
