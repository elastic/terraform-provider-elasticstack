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
// buildSyntheticsMonitorsPanel (write path) tests
// ─────────────────────────────────────────────────────────────────────────────

func makeTestGrid() struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
} {
	w := float32(24)
	h := float32(15)
	return struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	}{H: &h, W: &w, X: 0, Y: 0}
}

func Test_buildSyntheticsMonitorsPanel_noConfig(t *testing.T) {
	pm := panelModel{
		Type: types.StringValue(panelTypeSyntheticsMonitors),
	}
	grid := makeTestGrid()
	uid := "panel-1"

	panel := buildSyntheticsMonitorsPanel(pm, grid, &uid)

	assert.Equal(t, kbapi.SyntheticsMonitors, panel.Type)
	require.NotNil(t, panel.Uid)
	assert.Equal(t, "panel-1", *panel.Uid)
	assert.Nil(t, panel.Config.Filters)
}

func Test_buildSyntheticsMonitorsPanel_emptyConfigBlock(t *testing.T) {
	pm := panelModel{
		Type:                     types.StringValue(panelTypeSyntheticsMonitors),
		SyntheticsMonitorsConfig: &syntheticsMonitorsConfigModel{},
	}
	grid := makeTestGrid()

	panel := buildSyntheticsMonitorsPanel(pm, grid, nil)

	assert.Nil(t, panel.Config.Filters)
}

func Test_buildSyntheticsMonitorsPanel_withFilters(t *testing.T) {
	pm := panelModel{
		Type: types.StringValue(panelTypeSyntheticsMonitors),
		SyntheticsMonitorsConfig: &syntheticsMonitorsConfigModel{
			Filters: &syntheticsMonitorsFiltersModel{
				Projects: []syntheticsFilterItemModel{
					{Label: types.StringValue("My Project"), Value: types.StringValue("proj-1")},
				},
				Tags: []syntheticsFilterItemModel{
					{Label: types.StringValue("prod"), Value: types.StringValue("prod")},
				},
			},
		},
	}
	grid := makeTestGrid()

	panel := buildSyntheticsMonitorsPanel(pm, grid, nil)

	require.NotNil(t, panel.Config.Filters)
	require.NotNil(t, panel.Config.Filters.Projects)
	require.Len(t, *panel.Config.Filters.Projects, 1)
	assert.Equal(t, "My Project", (*panel.Config.Filters.Projects)[0].Label)
	assert.Equal(t, "proj-1", (*panel.Config.Filters.Projects)[0].Value)
	require.NotNil(t, panel.Config.Filters.Tags)
	require.Len(t, *panel.Config.Filters.Tags, 1)
	assert.Equal(t, "prod", (*panel.Config.Filters.Tags)[0].Value)
}

func Test_buildSyntheticsMonitorsPanel_allFilterDimensions(t *testing.T) {
	pm := panelModel{
		Type: types.StringValue(panelTypeSyntheticsMonitors),
		SyntheticsMonitorsConfig: &syntheticsMonitorsConfigModel{
			Filters: &syntheticsMonitorsFiltersModel{
				Projects:     []syntheticsFilterItemModel{{Label: types.StringValue("P1"), Value: types.StringValue("p1")}},
				Tags:         []syntheticsFilterItemModel{{Label: types.StringValue("T1"), Value: types.StringValue("t1")}},
				MonitorIDs:   []syntheticsFilterItemModel{{Label: types.StringValue("M1"), Value: types.StringValue("m1")}},
				Locations:    []syntheticsFilterItemModel{{Label: types.StringValue("L1"), Value: types.StringValue("l1")}},
				MonitorTypes: []syntheticsFilterItemModel{{Label: types.StringValue("http"), Value: types.StringValue("http")}},
				Statuses:     []syntheticsFilterItemModel{{Label: types.StringValue("Up"), Value: types.StringValue("up")}},
			},
		},
	}
	grid := makeTestGrid()

	panel := buildSyntheticsMonitorsPanel(pm, grid, nil)

	require.NotNil(t, panel.Config.Filters)
	assert.NotNil(t, panel.Config.Filters.Projects)
	assert.NotNil(t, panel.Config.Filters.Tags)
	assert.NotNil(t, panel.Config.Filters.MonitorIds)
	assert.NotNil(t, panel.Config.Filters.Locations)
	assert.NotNil(t, panel.Config.Filters.MonitorTypes)
	assert.NotNil(t, panel.Config.Filters.Statuses)
}

// ─────────────────────────────────────────────────────────────────────────────
// populateSyntheticsMonitorsFromAPI (read path) tests
// ─────────────────────────────────────────────────────────────────────────────

// makeSyntheticsPanel builds a KbnDashboardPanelSyntheticsMonitors for use in tests.
func makeSyntheticsPanel() kbapi.KbnDashboardPanelSyntheticsMonitors {
	return kbapi.KbnDashboardPanelSyntheticsMonitors{
		Type: kbapi.SyntheticsMonitors,
	}
}

// On import (tfPanel == nil) with no filters returned from API, config remains nil.
func Test_populateSyntheticsMonitorsFromAPI_import_noFilters(t *testing.T) {
	pm := &panelModel{}
	populateSyntheticsMonitorsFromAPI(pm, nil, makeSyntheticsPanel())
	assert.Nil(t, pm.SyntheticsMonitorsConfig)
}

// On import with project filter data in API response, config is populated.
func Test_populateSyntheticsMonitorsFromAPI_import_withFilters(t *testing.T) {
	pm := &panelModel{}
	apiPanel := makeSyntheticsPanel()
	projects := []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}{{Label: "My Project", Value: "proj-1"}}
	apiPanel.Config.Filters = &struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct { //nolint:revive
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
		Statuses *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"statuses,omitempty"`
		Tags *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"tags,omitempty"`
	}{
		Projects: &projects,
	}
	populateSyntheticsMonitorsFromAPI(pm, nil, apiPanel)

	require.NotNil(t, pm.SyntheticsMonitorsConfig)
	require.NotNil(t, pm.SyntheticsMonitorsConfig.Filters)
	require.Len(t, pm.SyntheticsMonitorsConfig.Filters.Projects, 1)
	assert.Equal(t, "My Project", pm.SyntheticsMonitorsConfig.Filters.Projects[0].Label.ValueString())
	assert.Equal(t, "proj-1", pm.SyntheticsMonitorsConfig.Filters.Projects[0].Value.ValueString())
}

// Null-preservation: prior state has no config block; API returns filters.
// The config block should remain nil (preserve practitioner intent).
func Test_populateSyntheticsMonitorsFromAPI_nilBlock_preservesNilIntent(t *testing.T) {
	pm := &panelModel{}
	tfPanel := &panelModel{} // no SyntheticsMonitorsConfig
	apiPanel := makeSyntheticsPanel()
	projects := []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}{{Label: "P", Value: "p"}}
	apiPanel.Config.Filters = &struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct { //nolint:revive
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
		Statuses *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"statuses,omitempty"`
		Tags *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"tags,omitempty"`
	}{
		Projects: &projects,
	}
	populateSyntheticsMonitorsFromAPI(pm, tfPanel, apiPanel)
	assert.Nil(t, pm.SyntheticsMonitorsConfig, "config block should remain nil when prior state had no config block")
}

// Null-preservation: prior state had config block with no filters. API returns empty filters.
// The filters should remain nil.
func Test_populateSyntheticsMonitorsFromAPI_emptyAPIFilters_nullPreservation(t *testing.T) {
	existing := &syntheticsMonitorsConfigModel{
		Filters: nil, // practitioner wrote synthetics_monitors_config = {}
	}
	pm := &panelModel{SyntheticsMonitorsConfig: existing}
	tfPanel := &panelModel{SyntheticsMonitorsConfig: existing}

	// API returns present but empty filters struct
	apiPanel := makeSyntheticsPanel()
	apiPanel.Config.Filters = &struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct { //nolint:revive
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
		Statuses *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"statuses,omitempty"`
		Tags *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"tags,omitempty"`
	}{}
	populateSyntheticsMonitorsFromAPI(pm, tfPanel, apiPanel)

	require.NotNil(t, pm.SyntheticsMonitorsConfig)
	assert.Nil(t, pm.SyntheticsMonitorsConfig.Filters, "filters should remain nil when API returns empty filters")
}

// Prior state had filters configured; API round-trips them back.
func Test_populateSyntheticsMonitorsFromAPI_filtersRoundTrip(t *testing.T) {
	existing := &syntheticsMonitorsConfigModel{
		Filters: &syntheticsMonitorsFiltersModel{
			Projects: []syntheticsFilterItemModel{
				{Label: types.StringValue("P1"), Value: types.StringValue("p1")},
			},
		},
	}
	pm := &panelModel{SyntheticsMonitorsConfig: existing}
	tfPanel := &panelModel{SyntheticsMonitorsConfig: existing}

	apiPanel := makeSyntheticsPanel()
	projects := []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}{{Label: "P1", Value: "p1"}}
	apiPanel.Config.Filters = &struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct { //nolint:revive
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
		Statuses *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"statuses,omitempty"`
		Tags *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"tags,omitempty"`
	}{
		Projects: &projects,
	}
	populateSyntheticsMonitorsFromAPI(pm, tfPanel, apiPanel)

	require.NotNil(t, pm.SyntheticsMonitorsConfig)
	require.NotNil(t, pm.SyntheticsMonitorsConfig.Filters)
	require.Len(t, pm.SyntheticsMonitorsConfig.Filters.Projects, 1)
	assert.Equal(t, "P1", pm.SyntheticsMonitorsConfig.Filters.Projects[0].Label.ValueString())
	assert.Equal(t, "p1", pm.SyntheticsMonitorsConfig.Filters.Projects[0].Value.ValueString())
}

// Prior state had an explicit empty filters block (filters = {}).
// API returns an empty filters struct — the empty filters block is preserved to avoid a
// perpetual diff.
func Test_populateSyntheticsMonitorsFromAPI_emptyFiltersBlock_preserved(t *testing.T) {
	emptyFilters := &syntheticsMonitorsFiltersModel{} // all slices nil
	existing := &syntheticsMonitorsConfigModel{Filters: emptyFilters}
	pm := &panelModel{SyntheticsMonitorsConfig: existing}
	tfPanel := &panelModel{SyntheticsMonitorsConfig: existing}

	// API returns an empty filters struct (all dimensions absent).
	apiPanel := makeSyntheticsPanel()
	apiPanel.Config.Filters = &struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct { //nolint:revive
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
		Statuses *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"statuses,omitempty"`
		Tags *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"tags,omitempty"`
	}{}
	populateSyntheticsMonitorsFromAPI(pm, tfPanel, apiPanel)

	require.NotNil(t, pm.SyntheticsMonitorsConfig)
	// Empty filters block should be preserved (not dropped) to avoid a perpetual diff.
	assert.NotNil(t, pm.SyntheticsMonitorsConfig.Filters, "empty filters block should be preserved on refresh")
}

// Prior state had config with no filters; API returns nil filters → keep filters nil.
func Test_populateSyntheticsMonitorsFromAPI_apiNilFilters_preservesNilFilters(t *testing.T) {
	existing := &syntheticsMonitorsConfigModel{Filters: nil}
	pm := &panelModel{SyntheticsMonitorsConfig: existing}
	tfPanel := &panelModel{SyntheticsMonitorsConfig: existing}

	populateSyntheticsMonitorsFromAPI(pm, tfPanel, makeSyntheticsPanel())

	require.NotNil(t, pm.SyntheticsMonitorsConfig)
	assert.Nil(t, pm.SyntheticsMonitorsConfig.Filters)
}
