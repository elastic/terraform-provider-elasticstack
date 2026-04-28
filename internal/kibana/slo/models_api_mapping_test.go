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

package slo

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func apmAvailabilityIndicatorUnion(t *testing.T) kbapi.SLOsSloWithSummaryResponse_Indicator {
	t.Helper()
	ind := kbapi.SLOsIndicatorPropertiesApmAvailability{
		Type: "sli.apm.transactionErrorRate",
		Params: struct {
			Environment     string  `json:"environment"`
			Filter          *string `json:"filter,omitempty"`
			Index           string  `json:"index"`
			Service         string  `json:"service"`
			TransactionName string  `json:"transactionName"`
			TransactionType string  `json:"transactionType"`
		}{
			Service:         "svc",
			Environment:     "env",
			TransactionType: "req",
			TransactionName: "tx",
			Index:           "apm-*",
		},
	}
	var out kbapi.SLOsSloWithSummaryResponse_Indicator
	require.NoError(t, out.FromSLOsIndicatorPropertiesApmAvailability(ind))
	return out
}

func baseAPIModel(ind kbapi.SLOsSloWithSummaryResponse_Indicator) *models.Slo {
	return &models.Slo{
		SloID:           "slo-1",
		SpaceID:         "default",
		Name:            "n",
		Description:     "d",
		BudgetingMethod: kbapi.Occurrences,
		Indicator:       ind,
		TimeWindow:      kbapi.SLOsTimeWindow{Duration: "7d", Type: "rolling"},
		Objective:       kbapi.SLOsObjective{Target: 0.99},
	}
}

func TestTfArtifactsToAPIModel(t *testing.T) {
	t.Run("success maps dashboard ids", func(t *testing.T) {
		obj, odi := types.ObjectValue(tfSloArtifactDashboardObjectType.AttrTypes, map[string]attr.Value{
			"id": types.StringValue("dash-1"),
		})
		require.False(t, odi.HasError())
		lv, ldi := types.ListValue(tfSloArtifactDashboardObjectType, []attr.Value{obj})
		require.False(t, ldi.HasError())
		art, d := types.ObjectValue(tfArtifactsAttrTypes, map[string]attr.Value{
			"dashboards": lv,
		})
		require.False(t, d.HasError())
		api, diags := tfArtifactsToAPIModel(art)
		require.False(t, diags.HasError())
		require.NotNil(t, api)
		require.NotNil(t, api.Dashboards)
		require.Len(t, *api.Dashboards, 1)
		assert.Equal(t, "dash-1", (*api.Dashboards)[0].Id)
	})
	t.Run("error when id unknown", func(t *testing.T) {
		obj, odi := types.ObjectValue(tfSloArtifactDashboardObjectType.AttrTypes, map[string]attr.Value{
			"id": types.StringUnknown(),
		})
		require.False(t, odi.HasError())
		lv, ldi := types.ListValue(tfSloArtifactDashboardObjectType, []attr.Value{obj})
		require.False(t, ldi.HasError())
		art, d := types.ObjectValue(tfArtifactsAttrTypes, map[string]attr.Value{
			"dashboards": lv,
		})
		require.False(t, d.HasError())
		_, diags := tfArtifactsToAPIModel(art)
		require.True(t, diags.HasError(), "expected diagnostics for unknown id")
	})
	t.Run("null dashboards list clears artifacts", func(t *testing.T) {
		art, d := types.ObjectValue(tfArtifactsAttrTypes, map[string]attr.Value{
			"dashboards": types.ListNull(tfSloArtifactDashboardObjectType),
		})
		require.False(t, d.HasError())
		api, diags := tfArtifactsToAPIModel(art)
		require.False(t, diags.HasError())
		require.NotNil(t, api)
		require.NotNil(t, api.Dashboards)
		assert.Empty(t, *api.Dashboards)
	})
	t.Run("error when id null", func(t *testing.T) {
		obj, odi := types.ObjectValue(tfSloArtifactDashboardObjectType.AttrTypes, map[string]attr.Value{
			"id": types.StringNull(),
		})
		require.False(t, odi.HasError())
		lv, ldi := types.ListValue(tfSloArtifactDashboardObjectType, []attr.Value{obj})
		require.False(t, ldi.HasError())
		art, d := types.ObjectValue(tfArtifactsAttrTypes, map[string]attr.Value{
			"dashboards": lv,
		})
		require.False(t, d.HasError())
		_, diags := tfArtifactsToAPIModel(art)
		require.True(t, diags.HasError(), "expected diagnostics for null id")
	})
}

func TestPopulateFromAPI_settingsAndEnabled(t *testing.T) {
	sync := "1m"
	sf := "@timestamp"
	api := baseAPIModel(apmAvailabilityIndicatorUnion(t))
	api.Settings = &kbapi.SLOsSettings{SyncDelay: &sync, SyncField: &sf}
	api.Enabled = true

	m := tfModel{
		Settings:  mustSettingsObject(t, "1m", "", "@timestamp", nil),
		Artifacts: types.ObjectNull(tfArtifactsAttrTypes),
	}
	d := m.populateFromAPI(api)
	require.False(t, d.HasError())
	sfState := m.Settings.Attributes()["sync_field"].(types.String)
	assert.Equal(t, "@timestamp", sfState.ValueString())
	assert.True(t, m.Enabled.ValueBool())
}

func TestPopulateFromAPI_artifacts(t *testing.T) {
	t.Run("maps API references when state had not configured artifacts", func(t *testing.T) {
		api := baseAPIModel(apmAvailabilityIndicatorUnion(t))
		api.Artifacts = &kbapi.SLOsArtifacts{Dashboards: &[]struct {
			//nolint:revive // kibana: generated field
			Id string `json:"id"`
		}{{
			Id: "dash-from-api",
		}}}

		m := tfModel{Artifacts: types.ObjectNull(tfArtifactsAttrTypes)}
		d := m.populateFromAPI(api)
		require.False(t, d.HasError())
		dl, ok := m.Artifacts.Attributes()["dashboards"].(types.List)
		require.True(t, ok, "expected dashboards list, got m.Artifacts=%+v", m.Artifacts)
		require.Len(t, dl.Elements(), 1)
		ero := dl.Elements()[0].(types.Object)
		idv := ero.Attributes()["id"].(types.String)
		assert.Equal(t, "dash-from-api", idv.ValueString())
	})

	t.Run("null when API has no references and no prior block", func(t *testing.T) {
		api := baseAPIModel(apmAvailabilityIndicatorUnion(t))
		api.Artifacts = nil
		m := tfModel{Artifacts: types.ObjectNull(tfArtifactsAttrTypes)}
		require.False(t, m.populateFromAPI(api).HasError())
		assert.True(t, m.Artifacts.IsNull())
	})

	t.Run("empty list when state had the block and API has no references", func(t *testing.T) {
		api := baseAPIModel(apmAvailabilityIndicatorUnion(t))
		api.Artifacts = &kbapi.SLOsArtifacts{Dashboards: &[]struct {
			//nolint:revive // kibana: generated field
			Id string `json:"id"`
		}{}}
		empty, el := types.ListValue(tfSloArtifactDashboardObjectType, []attr.Value{})
		require.False(t, el.HasError())
		art, eo := types.ObjectValue(tfArtifactsAttrTypes, map[string]attr.Value{"dashboards": empty})
		require.False(t, eo.HasError())
		m := tfModel{Artifacts: art}
		require.False(t, m.populateFromAPI(api).HasError())
		dl, ok := m.Artifacts.Attributes()["dashboards"].(types.List)
		require.True(t, ok)
		assert.Empty(t, dl.Elements())
	})
}

func mustSettingsObject(t *testing.T, syncDelay, frequency, syncField string, prevent *bool) types.Object {
	t.Helper()
	var pib types.Bool
	if prevent == nil {
		pib = types.BoolNull()
	} else {
		pib = types.BoolValue(*prevent)
	}
	o, odi := types.ObjectValue(tfSettingsAttrTypes, map[string]attr.Value{
		"sync_delay":               types.StringValue(syncDelay),
		"frequency":                stringOrEmpty(frequency),
		"sync_field":               stringOrEmpty(syncField),
		"prevent_initial_backfill": pib,
	})
	require.False(t, odi.HasError())
	return o
}

func stringOrEmpty(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func TestTfSettings_toAPIModel(t *testing.T) {
	t.Run("includes sync_field when set", func(t *testing.T) {
		s := tfSettings{
			SyncDelay:              types.StringValue("1m"),
			Frequency:              types.StringNull(),
			SyncField:              types.StringValue("@timestamp"),
			PreventInitialBackfill: types.BoolNull(),
		}
		api := s.toAPIModel()
		require.NotNil(t, api)
		require.NotNil(t, api.SyncField)
		assert.Equal(t, "@timestamp", *api.SyncField)
	})
	t.Run("returns nil when no known settings attributes", func(t *testing.T) {
		s := tfSettings{
			SyncDelay:              types.StringNull(),
			Frequency:              types.StringNull(),
			SyncField:              types.StringNull(),
			PreventInitialBackfill: types.BoolNull(),
		}
		assert.Nil(t, s.toAPIModel())
	})
}

func TestTfModel_toAPIModel_kqlWithSettingsAndArtifacts(t *testing.T) {
	art := mustArtifactObject(t, "dash-acc-test")
	m := tfModel{
		Name:         types.StringValue("n"),
		Description:  types.StringValue("d"),
		SpaceID:      types.StringValue("default"),
		BudgetMethod: types.StringValue("timeslices"),
		TimeWindow: []tfTimeWindow{{
			Duration: types.StringValue("7d"),
			Type:     types.StringValue("rolling"),
		}},
		Objective: []tfObjective{{
			Target:          types.Float64Value(0.95),
			TimesliceTarget: types.Float64Value(0.95),
			TimesliceWindow: types.StringValue("5m"),
		}},
		Settings:  mustSettingsObject(t, "2m", "1m", "event.ingested", nil),
		Artifacts: art,
		KqlCustomIndicator: []tfKqlCustomIndicator{{
			Index:          types.StringValue("logs"),
			FilterKql:      types.ObjectNull(tfKqlKqlObjectAttrTypes),
			Filter:         types.StringValue("a:b"),
			Good:           types.StringValue("c:d"),
			GoodKql:        types.ObjectNull(tfKqlKqlObjectAttrTypes),
			Total:          types.StringValue("*"),
			TotalKql:       types.ObjectNull(tfKqlKqlObjectAttrTypes),
			TimestampField: types.StringValue("@timestamp"),
		}},
	}
	api, di := m.toAPIModel()
	require.False(t, di.HasError())
	require.NotNil(t, api.Settings)
	require.NotNil(t, api.Settings.SyncField)
	assert.Equal(t, "event.ingested", *api.Settings.SyncField)
	require.NotNil(t, api.Artifacts)
	require.NotNil(t, api.Artifacts.Dashboards)
	require.Len(t, *api.Artifacts.Dashboards, 1)
	assert.Equal(t, "dash-acc-test", (*api.Artifacts.Dashboards)[0].Id)
}

func mustArtifactObject(t *testing.T, dashboardID string) types.Object {
	t.Helper()
	row, odi := types.ObjectValue(tfSloArtifactDashboardObjectType.AttrTypes, map[string]attr.Value{
		"id": types.StringValue(dashboardID),
	})
	require.False(t, odi.HasError())
	lv, ldi := types.ListValue(tfSloArtifactDashboardObjectType, []attr.Value{row})
	require.False(t, ldi.HasError())
	art, d := types.ObjectValue(tfArtifactsAttrTypes, map[string]attr.Value{"dashboards": lv})
	require.False(t, d.HasError())
	return art
}
