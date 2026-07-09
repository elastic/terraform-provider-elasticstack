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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/links"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPanelModelWithGrid() models.PanelModel {
	return models.PanelModel{
		Type: types.StringValue("links"),
		Grid: models.PanelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
			W: types.Int64Value(24),
			H: types.Int64Value(10),
		},
		ID: types.StringValue("links-panel-id"),
	}
}

func TestLinksByValueToAPI(t *testing.T) {
	t.Parallel()

	pm := newPanelModelWithGrid()
	pm.LinksConfig = &models.LinksPanelConfigModel{
		ByValue: &models.LinksPanelByValueModel{
			Layout:      types.StringValue("vertical"),
			Title:       types.StringNull(),
			Description: types.StringNull(),
			HideTitle:   types.BoolNull(),
			HideBorder:  types.BoolNull(),
			Links: []models.LinkItemModel{
				{
					Type:         types.StringValue("dashboard"),
					Destination:  types.StringValue("dashboard-id-1"),
					Label:        types.StringValue("Dashboard link"),
					OpenInNewTab: types.BoolValue(true),
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
	}

	item, diags := links.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	back, err := item.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinks()
	require.NoError(t, err)
	assert.Equal(t, kbapi.Links, back.Type)
	require.NotNil(t, back.Grid.W)
	assert.InDelta(t, float64(24), float64(*back.Grid.W), 1e-6)

	cfg0, err := back.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0()
	require.NoError(t, err)
	require.NotNil(t, cfg0.Layout)
	assert.Equal(t, kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0Layout("vertical"), *cfg0.Layout)
	require.Len(t, cfg0.Links, 2)

	dash, err := cfg0.Links[0].AsKibanaHTTPAPIsKbnLinkPanelTypeDashboardLink()
	require.NoError(t, err)
	assert.Equal(t, kbapi.DashboardLink, dash.Type)
	assert.Equal(t, "dashboard-id-1", dash.Destination)
	require.NotNil(t, dash.Label)
	assert.Equal(t, "Dashboard link", *dash.Label)
	require.NotNil(t, dash.Options)
	require.NotNil(t, dash.Options.OpenInNewTab)
	assert.True(t, *dash.Options.OpenInNewTab)
	require.NotNil(t, dash.Options.UseFilters)
	assert.True(t, *dash.Options.UseFilters)
	require.NotNil(t, dash.Options.UseTimeRange)
	assert.False(t, *dash.Options.UseTimeRange)

	ext, err := cfg0.Links[1].AsKibanaHTTPAPIsKbnLinkTypeExternalLink()
	require.NoError(t, err)
	assert.Equal(t, kbapi.ExternalLink, ext.Type)
	assert.Equal(t, "https://example.com", ext.Destination)
	require.NotNil(t, ext.Label)
	assert.Equal(t, "External link", *ext.Label)
	require.NotNil(t, ext.Options)
	require.NotNil(t, ext.Options.OpenInNewTab)
	assert.False(t, *ext.Options.OpenInNewTab)
	require.NotNil(t, ext.Options.EncodeUrl)
	assert.True(t, *ext.Options.EncodeUrl)
}

func TestLinksByReferenceToAPI(t *testing.T) {
	t.Parallel()

	pm := newPanelModelWithGrid()
	pm.LinksConfig = &models.LinksPanelConfigModel{
		ByReference: &models.LinksPanelByReferenceModel{
			RefID:       types.StringValue("links-ref-1"),
			Title:       types.StringValue("Linked links panel"),
			Description: types.StringValue("From the library"),
			HideTitle:   types.BoolValue(true),
			HideBorder:  types.BoolValue(false),
		},
	}

	item, diags := links.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	back, err := item.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinks()
	require.NoError(t, err)
	assert.Equal(t, kbapi.Links, back.Type)

	cfg1, err := back.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig1()
	require.NoError(t, err)
	assert.Equal(t, "links-ref-1", cfg1.RefId)
	require.NotNil(t, cfg1.Title)
	assert.Equal(t, "Linked links panel", *cfg1.Title)
	require.NotNil(t, cfg1.Description)
	assert.Equal(t, "From the library", *cfg1.Description)
	require.NotNil(t, cfg1.HideTitle)
	assert.True(t, *cfg1.HideTitle)
	require.NotNil(t, cfg1.HideBorder)
	assert.False(t, *cfg1.HideBorder)
}

func TestLinksByValueNullDisplayFields(t *testing.T) {
	t.Parallel()

	pm := newPanelModelWithGrid()
	pm.LinksConfig = &models.LinksPanelConfigModel{
		ByValue: &models.LinksPanelByValueModel{
			Layout:      types.StringValue("horizontal"),
			Title:       types.StringNull(),
			Description: types.StringNull(),
			HideTitle:   types.BoolNull(),
			HideBorder:  types.BoolNull(),
			Links: []models.LinkItemModel{
				{
					Type:        types.StringValue("external"),
					Destination: types.StringValue("https://example.com"),
				},
			},
		},
	}

	item, diags := links.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	back, err := item.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinks()
	require.NoError(t, err)

	cfg0, err := back.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0()
	require.NoError(t, err)
	assert.Nil(t, cfg0.Title)
	assert.Nil(t, cfg0.Description)
	assert.Nil(t, cfg0.HideTitle)
	assert.Nil(t, cfg0.HideBorder)
}
