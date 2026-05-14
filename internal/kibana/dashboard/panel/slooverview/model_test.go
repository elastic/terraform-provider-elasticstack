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

package slooverview

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_sloSingleToAPI_basic(t *testing.T) {
	m := &models.SloOverviewSingleModel{
		SloID:         types.StringValue("my-slo-id"),
		SloInstanceID: types.StringValue("instance-1"),
		RemoteName:    types.StringValue("remote"),
		Title:         types.StringValue("My SLO"),
		Description:   types.StringValue("A description"),
		HideTitle:     types.BoolValue(true),
		HideBorder:    types.BoolValue(false),
		Drilldowns: []models.URLDrilldownModel{
			{
				URL:          types.StringValue("https://example.com"),
				Label:        types.StringValue("Open dashboard"),
				EncodeURL:    types.BoolValue(true),
				OpenInNewTab: types.BoolValue(true),
			},
		},
	}

	api, diags := singleToAPI(m)
	require.False(t, diags.HasError())

	assert.Equal(t, kbapi.SloSingleOverviewEmbeddableOverviewModeSingle, api.OverviewMode)
	assert.Equal(t, "my-slo-id", api.SloId)
	require.NotNil(t, api.SloInstanceId)
	assert.Equal(t, "instance-1", *api.SloInstanceId)
	require.NotNil(t, api.RemoteName)
	assert.Equal(t, "remote", *api.RemoteName)
	require.NotNil(t, api.Title)
	assert.Equal(t, "My SLO", *api.Title)
	require.NotNil(t, api.HideTitle)
	assert.True(t, *api.HideTitle)
	require.NotNil(t, api.Drilldowns)
	require.Len(t, *api.Drilldowns, 1)
	dd := (*api.Drilldowns)[0]
	assert.Equal(t, "https://example.com", dd.Url)
	assert.Equal(t, "Open dashboard", dd.Label)
	require.NotNil(t, dd.EncodeUrl)
	assert.True(t, *dd.EncodeUrl)
	require.NotNil(t, dd.OpenInNewTab)
	assert.True(t, *dd.OpenInNewTab)
}

func Test_sloSingleToAPI_no_optional_fields(t *testing.T) {
	m := &models.SloOverviewSingleModel{
		SloID:         types.StringValue("only-slo-id"),
		SloInstanceID: types.StringNull(),
		RemoteName:    types.StringNull(),
		Title:         types.StringNull(),
		Description:   types.StringNull(),
		HideTitle:     types.BoolNull(),
		HideBorder:    types.BoolNull(),
	}

	api, diags := singleToAPI(m)
	require.False(t, diags.HasError())

	assert.Equal(t, "only-slo-id", api.SloId)
	assert.Nil(t, api.SloInstanceId)
	assert.Nil(t, api.RemoteName)
	assert.Nil(t, api.Title)
	assert.Nil(t, api.HideTitle)
	assert.Nil(t, api.Drilldowns)
}

func Test_sloGroupsToAPI_with_group_filters(t *testing.T) {
	m := &models.SloOverviewGroupsModel{
		Title:       types.StringValue("Groups Overview"),
		Description: types.StringNull(),
		HideTitle:   types.BoolNull(),
		HideBorder:  types.BoolNull(),
		GroupFilters: &models.SloGroupFiltersModel{
			GroupBy:     types.StringValue("status"),
			KQLQuery:    types.StringValue("slo.name: my-*"),
			FiltersJSON: jsontypes.NewNormalizedNull(),
		},
	}

	api, diags := groupsToAPI(m)
	require.False(t, diags.HasError())

	assert.Equal(t, kbapi.Groups, api.OverviewMode)
	require.NotNil(t, api.Title)
	assert.Equal(t, "Groups Overview", *api.Title)
	require.NotNil(t, api.GroupFilters)
	require.NotNil(t, api.GroupFilters.GroupBy)
	assert.Equal(t, kbapi.SloGroupOverviewEmbeddableGroupFiltersGroupByStatus, *api.GroupFilters.GroupBy)
	require.NotNil(t, api.GroupFilters.KqlQuery)
	assert.Equal(t, "slo.name: my-*", *api.GroupFilters.KqlQuery)
}

func Test_sloSingleFromAPI_roundtrip(t *testing.T) {
	instanceID := "instance-abc"
	title := "SLO Panel"
	apiSingle := kbapi.SloSingleOverviewEmbeddable{
		OverviewMode:  kbapi.SloSingleOverviewEmbeddableOverviewModeSingle,
		SloId:         "slo-123",
		SloInstanceId: &instanceID,
		Title:         &title,
	}

	var config kbapi.KbnDashboardPanelTypeSloOverview_Config
	require.NoError(t, config.FromSloSingleOverviewEmbeddable(apiSingle))

	panel := kbapi.KbnDashboardPanelTypeSloOverview{
		Config: config,
		Grid:   kbapi.KbnDashboardPanelGrid{X: 0, Y: 0},
		Type:   kbapi.SloOverview,
	}

	pm := &models.PanelModel{}
	diags := PopulateFromAPI(pm, nil, panel)
	require.False(t, diags.HasError())

	require.NotNil(t, pm.SloOverviewConfig)
	require.NotNil(t, pm.SloOverviewConfig.Single)
	assert.Nil(t, pm.SloOverviewConfig.Groups)

	s := pm.SloOverviewConfig.Single
	assert.Equal(t, types.StringValue("slo-123"), s.SloID)
	assert.Equal(t, types.StringValue("instance-abc"), s.SloInstanceID)
	assert.Equal(t, types.StringValue("SLO Panel"), s.Title)
}

func Test_sloGroupsFromAPI_roundtrip(t *testing.T) {
	groupBy := kbapi.SloGroupOverviewEmbeddableGroupFiltersGroupByStatus
	kql := "slo.name: *"
	apiGroups := kbapi.SloGroupOverviewEmbeddable{
		OverviewMode: kbapi.Groups,
		GroupFilters: &struct {
			Filters  *[]kbapi.SloGroupOverviewEmbeddable_GroupFilters_Filters_Item `json:"filters,omitempty"`
			GroupBy  *kbapi.SloGroupOverviewEmbeddableGroupFiltersGroupBy          `json:"group_by,omitempty"`
			Groups   *[]string                                                     `json:"groups,omitempty"`
			KqlQuery *string                                                       `json:"kql_query,omitempty"`
		}{
			GroupBy:  &groupBy,
			KqlQuery: &kql,
		},
	}

	var config kbapi.KbnDashboardPanelTypeSloOverview_Config
	require.NoError(t, config.FromSloGroupOverviewEmbeddable(apiGroups))

	panel := kbapi.KbnDashboardPanelTypeSloOverview{
		Config: config,
		Grid: struct {
			H *float32 `json:"h,omitempty"`
			W *float32 `json:"w,omitempty"`
			X float32  `json:"x"`
			Y float32  `json:"y"`
		}{X: 0, Y: 0},
		Type: kbapi.SloOverview,
	}

	pm := &models.PanelModel{}
	diags := PopulateFromAPI(pm, nil, panel)
	require.False(t, diags.HasError())

	require.NotNil(t, pm.SloOverviewConfig)
	require.NotNil(t, pm.SloOverviewConfig.Groups)
	assert.Nil(t, pm.SloOverviewConfig.Single)

	g := pm.SloOverviewConfig.Groups
	require.NotNil(t, g.GroupFilters)
	assert.Equal(t, types.StringValue("status"), g.GroupFilters.GroupBy)
	assert.Equal(t, types.StringValue("slo.name: *"), g.GroupFilters.KQLQuery)
}

func Test_sloInstanceID_null_preservation(t *testing.T) {
	defaultInstanceID := "*"
	apiSingle := kbapi.SloSingleOverviewEmbeddable{
		OverviewMode:  kbapi.SloSingleOverviewEmbeddableOverviewModeSingle,
		SloId:         "slo-456",
		SloInstanceId: &defaultInstanceID,
	}

	var config kbapi.KbnDashboardPanelTypeSloOverview_Config
	require.NoError(t, config.FromSloSingleOverviewEmbeddable(apiSingle))

	panel := kbapi.KbnDashboardPanelTypeSloOverview{
		Config: config,
		Grid: struct {
			H *float32 `json:"h,omitempty"`
			W *float32 `json:"w,omitempty"`
			X float32  `json:"x"`
			Y float32  `json:"y"`
		}{X: 0, Y: 0},
		Type: kbapi.SloOverview,
	}

	tfPanel := &models.PanelModel{
		SloOverviewConfig: &models.SloOverviewConfigModel{
			Single: &models.SloOverviewSingleModel{
				SloID:         types.StringValue("slo-456"),
				SloInstanceID: types.StringNull(),
			},
		},
	}

	pm := &models.PanelModel{}
	*pm = *tfPanel
	diags := PopulateFromAPI(pm, tfPanel, panel)
	require.False(t, diags.HasError())

	require.NotNil(t, pm.SloOverviewConfig)
	require.NotNil(t, pm.SloOverviewConfig.Single)
	assert.True(t, pm.SloOverviewConfig.Single.SloInstanceID.IsNull())
}

func Test_sloInstanceID_explicit_value_preserved(t *testing.T) {
	instanceID := "instance-1"
	apiSingle := kbapi.SloSingleOverviewEmbeddable{
		OverviewMode:  kbapi.SloSingleOverviewEmbeddableOverviewModeSingle,
		SloId:         "slo-789",
		SloInstanceId: &instanceID,
	}

	var config kbapi.KbnDashboardPanelTypeSloOverview_Config
	require.NoError(t, config.FromSloSingleOverviewEmbeddable(apiSingle))

	panel := kbapi.KbnDashboardPanelTypeSloOverview{
		Config: config,
		Grid: struct {
			H *float32 `json:"h,omitempty"`
			W *float32 `json:"w,omitempty"`
			X float32  `json:"x"`
			Y float32  `json:"y"`
		}{X: 0, Y: 0},
		Type: kbapi.SloOverview,
	}

	tfPanel := &models.PanelModel{
		SloOverviewConfig: &models.SloOverviewConfigModel{
			Single: &models.SloOverviewSingleModel{
				SloID:         types.StringValue("slo-789"),
				SloInstanceID: types.StringValue("instance-1"),
			},
		},
	}

	pm := &models.PanelModel{}
	*pm = *tfPanel
	diags := PopulateFromAPI(pm, tfPanel, panel)
	require.False(t, diags.HasError())

	require.NotNil(t, pm.SloOverviewConfig.Single)
	assert.Equal(t, types.StringValue("instance-1"), pm.SloOverviewConfig.Single.SloInstanceID)
}

func Test_sloOverview_handlerToAPI_single(t *testing.T) {
	pm := models.PanelModel{
		Type: types.StringValue(panelType),
		Grid: models.PanelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
			W: types.Int64Value(24),
			H: types.Int64Value(10),
		},
		SloOverviewConfig: &models.SloOverviewConfigModel{
			Single: &models.SloOverviewSingleModel{
				SloID:         types.StringValue("test-slo"),
				SloInstanceID: types.StringNull(),
				RemoteName:    types.StringNull(),
				Title:         types.StringNull(),
				Description:   types.StringNull(),
				HideTitle:     types.BoolNull(),
				HideBorder:    types.BoolNull(),
			},
		},
	}

	var h Handler
	item, diags := h.ToAPI(pm, nil)
	require.False(t, diags.HasError())

	panel, err := item.AsKbnDashboardPanelTypeSloOverview()
	require.NoError(t, err)
	assert.Equal(t, kbapi.KbnDashboardPanelTypeSloOverviewType("slo_overview"), panel.Type)
	single, err := panel.Config.AsSloSingleOverviewEmbeddable()
	require.NoError(t, err)
	assert.Equal(t, "test-slo", single.SloId)
	assert.Equal(t, kbapi.SloSingleOverviewEmbeddableOverviewModeSingle, single.OverviewMode)
}

func Test_sloOverview_handlerToAPI_groups(t *testing.T) {
	pm := models.PanelModel{
		Type: types.StringValue(panelType),
		Grid: models.PanelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
			W: types.Int64Value(24),
			H: types.Int64Value(10),
		},
		SloOverviewConfig: &models.SloOverviewConfigModel{
			Groups: &models.SloOverviewGroupsModel{
				Title:       types.StringNull(),
				Description: types.StringNull(),
				HideTitle:   types.BoolNull(),
				HideBorder:  types.BoolNull(),
				GroupFilters: &models.SloGroupFiltersModel{
					GroupBy:     types.StringValue("slo.tags"),
					KQLQuery:    types.StringNull(),
					FiltersJSON: jsontypes.NewNormalizedNull(),
				},
			},
		},
	}

	var h Handler
	item, diags := h.ToAPI(pm, nil)
	require.False(t, diags.HasError())

	panel, err := item.AsKbnDashboardPanelTypeSloOverview()
	require.NoError(t, err)
	groups, err := panel.Config.AsSloGroupOverviewEmbeddable()
	require.NoError(t, err)
	assert.Equal(t, kbapi.Groups, groups.OverviewMode)
	require.NotNil(t, groups.GroupFilters)
	require.NotNil(t, groups.GroupFilters.GroupBy)
	assert.Equal(t, kbapi.SloGroupOverviewEmbeddableGroupFiltersGroupBySloTags, *groups.GroupFilters.GroupBy)
}
