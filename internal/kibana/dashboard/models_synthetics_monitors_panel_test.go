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

func makeGrid() struct {
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
	grid := makeGrid()
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
	grid := makeGrid()

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
	grid := makeGrid()

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
			},
		},
	}
	grid := makeGrid()

	panel := buildSyntheticsMonitorsPanel(pm, grid, nil)

	require.NotNil(t, panel.Config.Filters)
	assert.NotNil(t, panel.Config.Filters.Projects)
	assert.NotNil(t, panel.Config.Filters.Tags)
	assert.NotNil(t, panel.Config.Filters.MonitorIds)
	assert.NotNil(t, panel.Config.Filters.Locations)
	assert.NotNil(t, panel.Config.Filters.MonitorTypes)
}

// ─────────────────────────────────────────────────────────────────────────────
// populateSyntheticsMonitorsFromAPI (read path) tests
// ─────────────────────────────────────────────────────────────────────────────

// makeAPIFilters builds the API filter struct pointer used in tests.
func makeAPIFilters(projects, tags, monitorIDs, locations, monitorTypes []struct{ Label, Value string }) *struct {
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
} {
	type item = struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	f := &struct {
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
	}{}

	if len(projects) > 0 {
		s := make([]item, len(projects))
		for i, p := range projects {
			s[i] = item{Label: p.Label, Value: p.Value}
		}
		f.Projects = &s
	}
	if len(tags) > 0 {
		s := make([]item, len(tags))
		for i, p := range tags {
			s[i] = item{Label: p.Label, Value: p.Value}
		}
		f.Tags = &s
	}
	if len(monitorIDs) > 0 {
		s := make([]item, len(monitorIDs))
		for i, p := range monitorIDs {
			s[i] = item{Label: p.Label, Value: p.Value}
		}
		f.MonitorIds = &s
	}
	if len(locations) > 0 {
		s := make([]item, len(locations))
		for i, p := range locations {
			s[i] = item{Label: p.Label, Value: p.Value}
		}
		f.Locations = &s
	}
	if len(monitorTypes) > 0 {
		s := make([]item, len(monitorTypes))
		for i, p := range monitorTypes {
			s[i] = item{Label: p.Label, Value: p.Value}
		}
		f.MonitorTypes = &s
	}
	return f
}

// On import (tfPanel == nil) with no filters returned from API, config remains nil.
func Test_populateSyntheticsMonitorsFromAPI_import_noFilters(t *testing.T) {
	pm := &panelModel{}
	populateSyntheticsMonitorsFromAPI(pm, nil, nil)
	assert.Nil(t, pm.SyntheticsMonitorsConfig)
}

// On import with filter data in API response, config is populated.
func Test_populateSyntheticsMonitorsFromAPI_import_withFilters(t *testing.T) {
	pm := &panelModel{}
	apiFilters := makeAPIFilters(
		[]struct{ Label, Value string }{{"My Project", "proj-1"}},
		nil, nil, nil, nil,
	)
	populateSyntheticsMonitorsFromAPI(pm, nil, apiFilters)

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
	apiFilters := makeAPIFilters(
		[]struct{ Label, Value string }{{"P", "p"}},
		nil, nil, nil, nil,
	)
	populateSyntheticsMonitorsFromAPI(pm, tfPanel, apiFilters)
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

	// API returns present but empty filters
	f := makeAPIFilters(nil, nil, nil, nil, nil)
	populateSyntheticsMonitorsFromAPI(pm, tfPanel, f)

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

	apiFilters := makeAPIFilters(
		[]struct{ Label, Value string }{{"P1", "p1"}},
		nil, nil, nil, nil,
	)
	populateSyntheticsMonitorsFromAPI(pm, tfPanel, apiFilters)

	require.NotNil(t, pm.SyntheticsMonitorsConfig)
	require.NotNil(t, pm.SyntheticsMonitorsConfig.Filters)
	require.Len(t, pm.SyntheticsMonitorsConfig.Filters.Projects, 1)
	assert.Equal(t, "P1", pm.SyntheticsMonitorsConfig.Filters.Projects[0].Label.ValueString())
	assert.Equal(t, "p1", pm.SyntheticsMonitorsConfig.Filters.Projects[0].Value.ValueString())
}

// Prior state had config with no filters; API returns nil filters → keep filters nil.
func Test_populateSyntheticsMonitorsFromAPI_apiNilFilters_preservesNilFilters(t *testing.T) {
	existing := &syntheticsMonitorsConfigModel{Filters: nil}
	pm := &panelModel{SyntheticsMonitorsConfig: existing}
	tfPanel := &panelModel{SyntheticsMonitorsConfig: existing}

	populateSyntheticsMonitorsFromAPI(pm, tfPanel, nil)

	require.NotNil(t, pm.SyntheticsMonitorsConfig)
	assert.Nil(t, pm.SyntheticsMonitorsConfig.Filters)
}
