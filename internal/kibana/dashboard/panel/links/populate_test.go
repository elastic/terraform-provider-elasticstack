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

package links_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/links"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustLinksPanelItem(t *testing.T, apiJSON string) kbapi.DashboardPanelItem {
	t.Helper()
	var panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &panel))
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeLinks(panel))
	return item
}

func TestLinksByValueFromAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	apiJSON := `{
		"type": "links",
		"grid": { "x": 0, "y": 0, "w": 24, "h": 10 },
		"id": "links-panel-id",
		"config": {
			"layout": "vertical",
			"title": "API title",
			"description": "API description",
			"hide_title": true,
			"hide_border": false,
			"links": [
				{
					"type": "dashboardLink",
					"destination": "dashboard-id-1",
					"label": "Dashboard link",
					"options": {
						"open_in_new_tab": true,
						"use_filters": false,
						"use_time_range": true
					}
				},
				{
					"type": "externalLink",
					"destination": "https://example.com",
					"label": "External link",
					"options": {
						"open_in_new_tab": true,
						"encode_url": false
					}
				}
			]
		}
	}`

	prior := &models.PanelModel{
		LinksConfig: &models.LinksPanelConfigModel{
			ByValue: &models.LinksPanelByValueModel{
				Layout:      types.StringValue("horizontal"),
				Title:       types.StringValue("Prior title"),
				Description: types.StringValue("Prior description"),
				HideTitle:   types.BoolValue(false),
				HideBorder:  types.BoolValue(true),
				Links: []models.LinkItemModel{
					{
						Type:         types.StringValue("dashboard"),
						Destination:  types.StringValue("dashboard-id-1"),
						Label:        types.StringValue("Dashboard link"),
						OpenInNewTab: types.BoolValue(false),
						UseFilters:   types.BoolValue(true),
						UseTimeRange: types.BoolValue(false),
						EncodeURL:    types.BoolNull(),
					},
					{
						Type:         types.StringValue("external"),
						Destination:  types.StringValue("https://example.com"),
						Label:        types.StringValue("External link"),
						OpenInNewTab: types.BoolValue(false),
						EncodeURL:    types.BoolValue(true),
						UseFilters:   types.BoolNull(),
						UseTimeRange: types.BoolNull(),
					},
				},
			},
		},
	}

	var pm models.PanelModel
	item := mustLinksPanelItem(t, apiJSON)
	diags := links.Handler{}.FromAPI(ctx, &pm, prior, item)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.LinksConfig)
	require.NotNil(t, pm.LinksConfig.ByValue)
	assert.True(t, pm.LinksConfig.ByValue.Layout.Equal(types.StringValue("vertical")))
	assert.True(t, pm.LinksConfig.ByValue.Title.Equal(types.StringValue("API title")))
	assert.True(t, pm.LinksConfig.ByValue.Description.Equal(types.StringValue("API description")))
	assert.True(t, pm.LinksConfig.ByValue.HideTitle.Equal(types.BoolValue(true)))
	assert.True(t, pm.LinksConfig.ByValue.HideBorder.Equal(types.BoolValue(false)))
	require.Len(t, pm.LinksConfig.ByValue.Links, 2)

	dashLink := pm.LinksConfig.ByValue.Links[0]
	assert.Equal(t, "dashboard", dashLink.Type.ValueString())
	assert.Equal(t, "dashboard-id-1", dashLink.Destination.ValueString())
	assert.True(t, dashLink.OpenInNewTab.Equal(types.BoolValue(true)))
	assert.True(t, dashLink.UseFilters.Equal(types.BoolValue(false)))
	assert.True(t, dashLink.UseTimeRange.Equal(types.BoolValue(true)))

	extLink := pm.LinksConfig.ByValue.Links[1]
	assert.Equal(t, "external", extLink.Type.ValueString())
	assert.Equal(t, "https://example.com", extLink.Destination.ValueString())
	assert.True(t, extLink.OpenInNewTab.Equal(types.BoolValue(true)))
	assert.True(t, extLink.EncodeURL.Equal(types.BoolValue(false)))
}

func TestLinksByReferenceFromAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	apiJSON := `{
		"type": "links",
		"grid": { "x": 0, "y": 0, "w": 24, "h": 10 },
		"id": "links-ref-panel-id",
		"config": {
			"ref_id": "links-ref-1",
			"title": "Linked title",
			"description": "Linked description",
			"hide_title": true,
			"hide_border": false
		}
	}`

	prior := &models.PanelModel{
		LinksConfig: &models.LinksPanelConfigModel{
			ByReference: &models.LinksPanelByReferenceModel{
				RefID:       types.StringValue("links-ref-1"),
				Title:       types.StringValue("Prior title"),
				Description: types.StringValue("Prior description"),
				HideTitle:   types.BoolValue(false),
				HideBorder:  types.BoolValue(true),
			},
		},
	}

	var pm models.PanelModel
	item := mustLinksPanelItem(t, apiJSON)
	diags := links.Handler{}.FromAPI(ctx, &pm, prior, item)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.LinksConfig)
	require.NotNil(t, pm.LinksConfig.ByReference)
	assert.True(t, pm.LinksConfig.ByReference.RefID.Equal(types.StringValue("links-ref-1")))
	assert.True(t, pm.LinksConfig.ByReference.Title.Equal(types.StringValue("Linked title")))
	assert.True(t, pm.LinksConfig.ByReference.Description.Equal(types.StringValue("Linked description")))
	assert.True(t, pm.LinksConfig.ByReference.HideTitle.Equal(types.BoolValue(true)))
	assert.True(t, pm.LinksConfig.ByReference.HideBorder.Equal(types.BoolValue(false)))
}

func TestLinksByValueNullPreservation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	apiJSON := `{
		"type": "links",
		"grid": { "x": 0, "y": 0, "w": 24, "h": 10 },
		"config": {
			"layout": "vertical",
			"title": "API title",
			"description": "API description",
			"hide_title": true,
			"hide_border": false,
			"links": [
				{
					"type": "externalLink",
					"destination": "https://example.com"
				}
			]
		}
	}`

	prior := &models.PanelModel{
		LinksConfig: &models.LinksPanelConfigModel{
			ByValue: &models.LinksPanelByValueModel{
				Layout:      types.StringValue("vertical"),
				Title:       types.StringNull(),
				Description: types.StringValue("Prior description"),
				HideTitle:   types.BoolValue(false),
				HideBorder:  types.BoolNull(),
				Links: []models.LinkItemModel{
					{
						Type:        types.StringValue("external"),
						Destination: types.StringValue("https://example.com"),
					},
				},
			},
		},
	}

	var pm models.PanelModel
	item := mustLinksPanelItem(t, apiJSON)
	diags := links.Handler{}.FromAPI(ctx, &pm, prior, item)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.LinksConfig.ByValue)
	assert.True(t, pm.LinksConfig.Title.IsNull())
	assert.True(t, pm.LinksConfig.ByValue.Title.IsNull())
	assert.True(t, pm.LinksConfig.ByValue.HideBorder.IsNull())
	assert.True(t, pm.LinksConfig.ByValue.Description.Equal(types.StringValue("API description")))
	assert.True(t, pm.LinksConfig.ByValue.HideTitle.Equal(types.BoolValue(true)))
}

func TestLinksByValueImport(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	apiJSON := `{
		"type": "links",
		"grid": { "x": 1, "y": 2, "w": 20, "h": 8 },
		"id": "links-import-id",
		"config": {
			"layout": "horizontal",
			"links": [
				{
					"type": "dashboardLink",
					"destination": "dashboard-import"
				},
				{
					"type": "externalLink",
					"destination": "https://example.com/import"
				}
			]
		}
	}`

	var pm models.PanelModel
	item := mustLinksPanelItem(t, apiJSON)
	diags := links.Handler{}.FromAPI(ctx, &pm, nil, item)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.LinksConfig)
	require.NotNil(t, pm.LinksConfig.ByValue)
	assert.True(t, pm.LinksConfig.ByValue.Layout.Equal(types.StringValue("horizontal")))
	assert.True(t, pm.LinksConfig.ByValue.Title.IsNull())
	assert.True(t, pm.LinksConfig.ByValue.Description.IsNull())
	assert.True(t, pm.LinksConfig.ByValue.HideTitle.IsNull())
	assert.True(t, pm.LinksConfig.ByValue.HideBorder.IsNull())
	require.Len(t, pm.LinksConfig.ByValue.Links, 2)
	assert.Equal(t, "dashboard", pm.LinksConfig.ByValue.Links[0].Type.ValueString())
	assert.Equal(t, "external", pm.LinksConfig.ByValue.Links[1].Type.ValueString())
	assert.Equal(t, "dashboard-import", pm.LinksConfig.ByValue.Links[0].Destination.ValueString())
	assert.Equal(t, "https://example.com/import", pm.LinksConfig.ByValue.Links[1].Destination.ValueString())
}
