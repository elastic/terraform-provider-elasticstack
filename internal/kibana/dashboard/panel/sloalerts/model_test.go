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

package sloalerts

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sloAlertsEmbeddableFromJSON(t *testing.T, raw string) kbapi.SloAlertsEmbeddable {
	t.Helper()
	var emb kbapi.SloAlertsEmbeddable
	require.NoError(t, json.Unmarshal([]byte(raw), &emb))
	return emb
}

func sloAlertsPanelModel(cfg *models.SloAlertsPanelConfigModel, x, y int64) models.PanelModel {
	return models.PanelModel{
		Type:            types.StringValue(panelType),
		Grid:            models.PanelGridModel{X: types.Int64Value(x), Y: types.Int64Value(y)},
		SloAlertsConfig: cfg,
	}
}

func Test_sloAlertsPanelToAPI_minimal(t *testing.T) {
	pm := sloAlertsPanelModel(&models.SloAlertsPanelConfigModel{
		Slos: []models.SloAlertsPanelSloModel{
			{SloID: types.StringValue("slo-1")},
		},
	}, 0, 0)
	item, diags := Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError())

	sa, err := item.AsKbnDashboardPanelTypeSloAlerts()
	require.NoError(t, err)
	require.NotNil(t, sa.Config.Slos)
	require.Len(t, *sa.Config.Slos, 1)
	assert.Equal(t, "slo-1", (*sa.Config.Slos)[0].SloId)
	assert.Nil(t, (*sa.Config.Slos)[0].SloInstanceId)
}

func Test_sloAlertsPanel_roundTrip_minimal(t *testing.T) {
	pm := sloAlertsPanelModel(&models.SloAlertsPanelConfigModel{
		Slos: []models.SloAlertsPanelSloModel{
			{SloID: types.StringValue("slo-1"), SloInstanceID: types.StringNull()},
		},
	}, 0, 0)
	item, diags := Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError())
	apiPanel, err := item.AsKbnDashboardPanelTypeSloAlerts()
	require.NoError(t, err)

	apiPanel.Config = sloAlertsEmbeddableFromJSON(t, `{"slos":[{"slo_id":"slo-1","slo_instance_id":"*"}]}`)

	next := pm
	PopulateFromAPI(&next, &pm, apiPanel)
	require.NotNil(t, next.SloAlertsConfig)
	require.Len(t, next.SloAlertsConfig.Slos, 1)
	assert.Equal(t, "slo-1", next.SloAlertsConfig.Slos[0].SloID.ValueString())
	assert.True(t, next.SloAlertsConfig.Slos[0].SloInstanceID.IsNull())
}

func Test_sloAlertsPanel_roundTrip_multipleSlos(t *testing.T) {
	pm := sloAlertsPanelModel(&models.SloAlertsPanelConfigModel{
		Slos: []models.SloAlertsPanelSloModel{
			{SloID: types.StringValue("a"), SloInstanceID: types.StringNull()},
			{SloID: types.StringValue("b"), SloInstanceID: types.StringValue("inst-2")},
		},
	}, 1, 2)
	item, diags := Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError())
	apiPanel, err := item.AsKbnDashboardPanelTypeSloAlerts()
	require.NoError(t, err)

	next := pm
	PopulateFromAPI(&next, &pm, apiPanel)
	require.Len(t, next.SloAlertsConfig.Slos, 2)
	assert.Equal(t, "b", next.SloAlertsConfig.Slos[1].SloID.ValueString())
	assert.Equal(t, "inst-2", next.SloAlertsConfig.Slos[1].SloInstanceID.ValueString())
}

func Test_populateSloAlertsPanelFromAPI_sloInstanceID_nullPreserved_refreshAndImport(t *testing.T) {
	apiPanel := kbapi.KbnDashboardPanelTypeSloAlerts{
		Config: sloAlertsEmbeddableFromJSON(t, `{"slos":[{"slo_id":"slo-1","slo_instance_id":"*" }]}`),
	}

	t.Run("refresh", func(t *testing.T) {
		prior := models.PanelModel{
			SloAlertsConfig: &models.SloAlertsPanelConfigModel{
				Slos: []models.SloAlertsPanelSloModel{
					{SloID: types.StringValue("slo-1"), SloInstanceID: types.StringNull()},
				},
			},
		}
		next := prior
		PopulateFromAPI(&next, &prior, apiPanel)
		assert.True(t, next.SloAlertsConfig.Slos[0].SloInstanceID.IsNull())
	})

	t.Run("import", func(t *testing.T) {
		pm := models.PanelModel{}
		PopulateFromAPI(&pm, nil, apiPanel)
		require.NotNil(t, pm.SloAlertsConfig)
		assert.True(t, pm.SloAlertsConfig.Slos[0].SloInstanceID.IsNull())
	})
}

func Test_sloAlerts_drilldown_roundTrip_defaultsNull_refreshAndImport(t *testing.T) {
	pm := sloAlertsPanelModel(&models.SloAlertsPanelConfigModel{
		Slos: []models.SloAlertsPanelSloModel{{SloID: types.StringValue("slo-1")}},
		Drilldowns: []models.URLDrilldownModel{
			{
				URL:          types.StringValue("https://kibana.example/drill"),
				Label:        types.StringValue("investigate"),
				EncodeURL:    types.BoolNull(),
				OpenInNewTab: types.BoolNull(),
			},
		},
	}, 0, 0)
	item, diags := Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError())
	apiPanel, err := item.AsKbnDashboardPanelTypeSloAlerts()
	require.NoError(t, err)

	apiPanel.Config = sloAlertsEmbeddableFromJSON(t,
		`{"slos":[{"slo_id":"slo-1"}],"drilldowns":[`+
			`{"url":"https://kibana.example/drill","label":"investigate",`+
			`"trigger":"on_open_panel_menu","type":"url_drilldown",`+
			`"encode_url":true,"open_in_new_tab":false}]}`)

	t.Run("refresh", func(t *testing.T) {
		next := pm
		PopulateFromAPI(&next, &pm, apiPanel)
		d := next.SloAlertsConfig.Drilldowns[0]
		assert.True(t, d.EncodeURL.IsNull())
		assert.True(t, d.OpenInNewTab.IsNull())
	})

	t.Run("import", func(t *testing.T) {
		pmImport := models.PanelModel{}
		PopulateFromAPI(&pmImport, nil, apiPanel)
		d := pmImport.SloAlertsConfig.Drilldowns[0]
		assert.True(t, d.EncodeURL.IsNull())
		assert.True(t, d.OpenInNewTab.IsNull())
	})
}

func Test_sloAlertsPanelToAPI_drilldownWritesTrigger(t *testing.T) {
	pm := sloAlertsPanelModel(&models.SloAlertsPanelConfigModel{
		Slos: []models.SloAlertsPanelSloModel{{SloID: types.StringValue("x")}},
		Drilldowns: []models.URLDrilldownModel{
			{URL: types.StringValue("https://z"), Label: types.StringValue("lbl")},
		},
	}, 0, 0)
	item, diags := Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError())
	sa, err := item.AsKbnDashboardPanelTypeSloAlerts()
	require.NoError(t, err)
	require.NotNil(t, sa.Config.Drilldowns)
	d := (*sa.Config.Drilldowns)[0]
	assert.Equal(t, kbapi.SloAlertsEmbeddableDrilldownsTriggerOnOpenPanelMenu, d.Trigger)
	assert.Equal(t, kbapi.SloAlertsEmbeddableDrilldownsTypeUrlDrilldown, d.Type)
}

func Test_sloAlerts_slos_emptyList_rejected(t *testing.T) {
	ctx := context.Background()
	v := listvalidator.SizeAtLeast(1)
	sloObj := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"slo_id":          types.StringType,
			"slo_instance_id": types.StringType,
		},
	}
	empty := types.ListValueMust(sloObj, []attr.Value{})
	var resp validator.ListResponse
	v.ValidateList(ctx, validator.ListRequest{
		Path:           path.Root("slos"),
		PathExpression: path.MatchRoot("slos"),
		ConfigValue:    empty,
	}, &resp)
	require.True(t, resp.Diagnostics.HasError())
}

func Test_sloAlerts_slos_nonempty_accepted(t *testing.T) {
	ctx := context.Background()
	v := listvalidator.SizeAtLeast(1)
	sloObj := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"slo_id":          types.StringType,
			"slo_instance_id": types.StringType,
		},
	}
	row := types.ObjectValueMust(sloObj.AttrTypes, map[string]attr.Value{
		"slo_id":          types.StringValue("slo-1"),
		"slo_instance_id": types.StringNull(),
	})
	list := types.ListValueMust(sloObj, []attr.Value{row})
	var resp validator.ListResponse
	v.ValidateList(ctx, validator.ListRequest{
		Path:           path.Root("slos"),
		PathExpression: path.MatchRoot("slos"),
		ConfigValue:    list,
	}, &resp)
	require.False(t, resp.Diagnostics.HasError())
}

func Test_populateSloAlertsPanelFromAPI_import_preservesDrilldownDefaults(t *testing.T) {
	raw := `{"slos":[{"slo_id":"slo-1"}],"drilldowns":[{"url":"https://example.com","label":"open","trigger":"on_open_panel_menu","type":"url_drilldown","encode_url":true,"open_in_new_tab":false}]}`

	pm := models.PanelModel{}
	PopulateFromAPI(&pm, nil, kbapi.KbnDashboardPanelTypeSloAlerts{
		Config: sloAlertsEmbeddableFromJSON(t, raw),
	})
	require.NotNil(t, pm.SloAlertsConfig)
	assert.True(t, pm.SloAlertsConfig.Slos[0].SloInstanceID.IsNull())
	d := pm.SloAlertsConfig.Drilldowns[0]
	assert.True(t, d.EncodeURL.IsNull())
	assert.True(t, d.OpenInNewTab.IsNull())
}
