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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func durationFormatConfigJSON() jsontypes.Normalized {
	return jsontypes.NewNormalizedValue(`{"format":{"type":"duration","from":"s","to":"ms","decimals":0,"compact":true}}`)
}

func durationFormatXYPanel() models.PanelModel {
	return models.PanelModel{
		ID: types.StringValue("latency"),
		Grid: models.PanelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
			W: types.Int64Value(24),
			H: types.Int64Value(12),
		},
		VisConfig: &models.VisConfigModel{
			ByValue: &models.VisByValueModel{
				XYChartConfig: &models.XYChartConfigModel{
					Layers: []models.XYLayerModel{{
						Type: types.StringValue("line"),
						DataLayer: &models.DataLayerModel{
							Y: []models.YMetricModel{{
								ConfigJSON: durationFormatConfigJSON(),
							}},
						},
					}},
				},
			},
		},
	}
}

func diagnosticText(diags diag.Diagnostics) string {
	var b strings.Builder
	for _, d := range diags {
		b.WriteString(d.Summary())
		b.WriteByte('\n')
		b.WriteString(d.Detail())
		b.WriteByte('\n')
	}
	return b.String()
}

func newDashboardWriteServer(t *testing.T, statusCode int, statusBody any, dashboardWrites, statusCalls *atomic.Int32) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/api/status":
			if statusCalls != nil {
				statusCalls.Add(1)
			}
			rw.WriteHeader(statusCode)
			if statusBody != nil {
				_ = json.NewEncoder(rw).Encode(statusBody)
			}
		case req.Method == http.MethodPost && req.URL.Path == "/api/dashboards":
			dashboardWrites.Add(1)
			rw.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(rw).Encode(map[string]string{"id": "auto-generated-uuid"})
		case req.Method == http.MethodPut && strings.HasPrefix(req.URL.Path, dashboardAPIPathPrefix):
			dashboardWrites.Add(1)
			rw.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(rw).Encode(map[string]string{"id": "dash-1"})
		default:
			rw.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func statefulStatusBody(versionNumber string) map[string]any {
	return map[string]any{
		"version": map[string]any{
			"number":       versionNumber,
			"build_flavor": "default",
		},
	}
}

type dashboardWriteFunc func(context.Context, *clients.KibanaScopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]) (entitycore.KibanaWriteResult[models.DashboardModel], diag.Diagnostics)

func forEachDashboardWrite(t *testing.T, fn func(t *testing.T, id types.String, write dashboardWriteFunc)) {
	t.Helper()
	for _, op := range []struct {
		name  string
		id    types.String
		write dashboardWriteFunc
	}{
		{name: "create", id: types.StringNull(), write: createDashboard},
		{name: "update", id: types.StringValue("dash-1"), write: updateDashboard},
	} {
		t.Run(op.name, func(t *testing.T) {
			fn(t, op.id, op.write)
		})
	}
}

func writeDashboard(
	t *testing.T,
	write dashboardWriteFunc,
	statusCode int,
	statusBody any,
	plan models.DashboardModel,
	countStatus bool,
) (entitycore.KibanaWriteResult[models.DashboardModel], diag.Diagnostics, *atomic.Int32, *atomic.Int32) {
	t.Helper()
	var dashboardWrites atomic.Int32
	var statusCalls atomic.Int32
	var statusCounter *atomic.Int32
	if countStatus {
		statusCounter = &statusCalls
	}
	server := newDashboardWriteServer(t, statusCode, statusBody, &dashboardWrites, statusCounter)
	result, diags := write(t.Context(), newTestKibanaScopedClient(t, server), entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    plan,
		WriteID: plan.GetResourceID().ValueString(),
		SpaceID: plan.GetSpaceID().ValueString(),
	})
	return result, diags, &dashboardWrites, &statusCalls
}

func planWithPanel(id types.String, panel models.PanelModel) models.DashboardModel {
	plan := testDashboardPlanModel(id)
	plan.Panels = []models.PanelModel{panel}
	return plan
}

func panelWithConfigJSON(configJSON string) models.PanelModel {
	panel := durationFormatXYPanel()
	panel.VisConfig.ByValue.XYChartConfig.Layers[0].DataLayer.Y[0].ConfigJSON = jsontypes.NewNormalizedValue(configJSON)
	return panel
}

func requireDurationFormatRejected(t *testing.T, diags diag.Diagnostics, resultID types.String, dashboardWrites *atomic.Int32, pathParts ...string) {
	t.Helper()
	require.True(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	text := diagnosticText(diags)
	for _, part := range pathParts {
		require.Contains(t, text, part)
	}
	require.Contains(t, text, durationFormatParamsRequirement)
	require.Equal(t, int32(0), dashboardWrites.Load())
	require.True(t, resultID.IsNull() || resultID.ValueString() == "")
}

func requireWriteProceeded(t *testing.T, diags diag.Diagnostics, dashboardWrites, statusCalls *atomic.Int32) {
	t.Helper()
	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.NotContains(t, diagnosticText(diags), durationFormatParamsRequirement)
	require.Equal(t, int32(1), dashboardWrites.Load())
	if statusCalls != nil {
		require.Equal(t, int32(0), statusCalls.Load())
	}
}

func TestDashboardWrite_rejectsDurationFormatWithDecimalsAndCompactOnOlderStack(t *testing.T) {
	forEachDashboardWrite(t, func(t *testing.T, id types.String, write dashboardWriteFunc) {
		result, diags, writes, _ := writeDashboard(t, write, http.StatusOK, statefulStatusBody("9.5.1"), planWithPanel(id, durationFormatXYPanel()), false)
		requireDurationFormatRejected(t, diags, result.Model.ID, writes, "panels[0]", `id "latency"`, "layer 0", "metric 0")
	})
}

func TestDashboardWrite_propagatesStatusRetrievalFailureDuringVersionCheck(t *testing.T) {
	forEachDashboardWrite(t, func(t *testing.T, id types.String, write dashboardWriteFunc) {
		_, diags, writes, _ := writeDashboard(t, write, http.StatusInternalServerError, "kibana status unavailable", planWithPanel(id, durationFormatXYPanel()), false)
		text := diagnosticText(diags)
		require.True(t, diags.HasError(), "diagnostics: %s", text)
		require.Contains(t, text, "Unexpected status code from server: got HTTP 500")
		require.Contains(t, text, "kibana status unavailable")
		require.NotContains(t, text, durationFormatParamsRequirement)
		require.Equal(t, int32(0), writes.Load())
	})
}

func TestDashboardWrite_propagatesVersionParsingFailureDuringVersionCheck(t *testing.T) {
	forEachDashboardWrite(t, func(t *testing.T, id types.String, write dashboardWriteFunc) {
		_, diags, writes, _ := writeDashboard(t, write, http.StatusOK, statefulStatusBody("not-a-version"), planWithPanel(id, durationFormatXYPanel()), false)
		text := diagnosticText(diags)
		require.True(t, diags.HasError(), "diagnostics: %s", text)
		require.Contains(t, text, "malformed version: not-a-version")
		require.NotContains(t, text, durationFormatParamsRequirement)
		require.Equal(t, int32(0), writes.Load())
	})
}

func TestDashboardWrite_rejectsDurationFormatNestedInSectionOnOlderStack(t *testing.T) {
	forEachDashboardWrite(t, func(t *testing.T, id types.String, write dashboardWriteFunc) {
		plan := testDashboardPlanModel(id)
		plan.Sections = []models.SectionModel{{
			ID:     types.StringValue("overview"),
			Panels: []models.PanelModel{durationFormatXYPanel()},
		}}
		result, diags, writes, _ := writeDashboard(t, write, http.StatusOK, statefulStatusBody("9.5.1"), plan, false)
		requireDurationFormatRejected(t, diags, result.Model.ID, writes, "sections[0].panels[0]", `id "latency"`, "layer 0", "metric 0")
	})
}

func TestDashboardWrite_allowsDurationFormatParamsAtMinimumVersion(t *testing.T) {
	forEachDashboardWrite(t, func(t *testing.T, id types.String, write dashboardWriteFunc) {
		_, diags, writes, _ := writeDashboard(t, write, http.StatusOK, statefulStatusBody("9.5.2"), planWithPanel(id, durationFormatXYPanel()), false)
		requireWriteProceeded(t, diags, writes, nil)
	})
}

func TestDashboardWrite_skipsVersionCheckWithoutDurationFormatParams(t *testing.T) {
	forEachDashboardWrite(t, func(t *testing.T, id types.String, write dashboardWriteFunc) {
		_, diags, writes, statusCalls := writeDashboard(t, write, http.StatusOK, statefulStatusBody("9.4.0"), testDashboardPlanModel(id), true)
		requireWriteProceeded(t, diags, writes, statusCalls)
	})
}

func TestDashboardWrite_allowsDurationFormatParamsOnServerless(t *testing.T) {
	statusBody := map[string]any{
		"version": map[string]any{
			"number":       "9.4.0",
			"build_flavor": "serverless",
		},
	}
	forEachDashboardWrite(t, func(t *testing.T, id types.String, write dashboardWriteFunc) {
		_, diags, writes, _ := writeDashboard(t, write, http.StatusOK, statusBody, planWithPanel(id, durationFormatXYPanel()), false)
		requireWriteProceeded(t, diags, writes, nil)
	})
}

func TestDashboardWrite_doesNotGateNumberFormatOnOlderStack(t *testing.T) {
	forEachDashboardWrite(t, func(t *testing.T, id types.String, write dashboardWriteFunc) {
		plan := planWithPanel(id, panelWithConfigJSON(`{"format":{"type":"number","decimals":0,"compact":true}}`))
		_, diags, writes, statusCalls := writeDashboard(t, write, http.StatusOK, statefulStatusBody("9.4.0"), plan, true)
		requireWriteProceeded(t, diags, writes, statusCalls)
	})
}

func TestDashboardWrite_skipsVersionCheckForDurationFormatWithoutParams(t *testing.T) {
	forEachDashboardWrite(t, func(t *testing.T, id types.String, write dashboardWriteFunc) {
		plan := planWithPanel(id, panelWithConfigJSON(`{"format":{"type":"duration","from":"s","to":"ms"}}`))
		_, diags, writes, statusCalls := writeDashboard(t, write, http.StatusOK, statefulStatusBody("9.4.0"), plan, true)
		requireWriteProceeded(t, diags, writes, statusCalls)
	})
}

func TestCreateDashboard_namesNestedLayerAndMetricIndexes(t *testing.T) {
	offending := durationFormatXYPanel()
	offendingLayer := offending.VisConfig.ByValue.XYChartConfig.Layers[0]
	offending.VisConfig.ByValue.XYChartConfig.Layers = []models.XYLayerModel{
		{
			Type: types.StringValue("bar"),
			DataLayer: &models.DataLayerModel{
				Y: []models.YMetricModel{{
					ConfigJSON: jsontypes.NewNormalizedValue(`{"format":{"type":"number","decimals":2}}`),
				}},
			},
		},
		{
			Type: types.StringValue("line"),
			DataLayer: &models.DataLayerModel{
				Y: []models.YMetricModel{
					{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count"}`)},
					offendingLayer.DataLayer.Y[0],
				},
			},
		},
	}
	plan := testDashboardPlanModel(types.StringNull())
	plan.Panels = []models.PanelModel{{}, offending}

	result, diags, writes, _ := writeDashboard(t, createDashboard, http.StatusOK, statefulStatusBody("9.5.1"), plan, false)
	requireDurationFormatRejected(t, diags, result.Model.ID, writes, "panels[1]", `id "latency"`, "layer 1", "metric 1")
}

func TestCreateDashboard_skipsVersionCheckForNullFormat(t *testing.T) {
	plan := planWithPanel(types.StringNull(), panelWithConfigJSON(`{"format":null}`))
	_, diags, writes, statusCalls := writeDashboard(t, createDashboard, http.StatusOK, statefulStatusBody("9.4.0"), plan, true)
	requireWriteProceeded(t, diags, writes, statusCalls)
}

func TestCreateDashboard_doesNotGatePercentFormatOnOlderStack(t *testing.T) {
	plan := planWithPanel(types.StringNull(), panelWithConfigJSON(`{"format":{"type":"percent","decimals":0,"compact":true}}`))
	_, diags, writes, statusCalls := writeDashboard(t, createDashboard, http.StatusOK, statefulStatusBody("9.4.0"), plan, true)
	requireWriteProceeded(t, diags, writes, statusCalls)
}

func TestCreateDashboard_rejectsDurationFormatWithNullPanelIDOnOlderStack(t *testing.T) {
	panel := durationFormatXYPanel()
	panel.ID = types.StringNull()
	result, diags, writes, _ := writeDashboard(t, createDashboard, http.StatusOK, statefulStatusBody("9.5.1"), planWithPanel(types.StringNull(), panel), false)
	text := diagnosticText(diags)
	requireDurationFormatRejected(t, diags, result.Model.ID, writes, "panels[0]", "layer 0", "metric 0")
	require.NotContains(t, text, "id ")
}

func TestMetricUsesUnsupportedDurationParamsJSONShapes(t *testing.T) {
	tests := []struct {
		name       string
		configJSON string
	}{
		{name: "format id duration with decimals only", configJSON: `{"format":{"id":"duration","decimals":0}}`},
		{name: "format id duration with compact only", configJSON: `{"format":{"id":"duration","compact":true}}`},
		{name: "format id duration with decimals and compact", configJSON: `{"format":{"id":"duration","decimals":0,"compact":true}}`},
		{name: "format id duration with decimals under params", configJSON: `{"format":{"id":"duration","params":{"decimals":0}}}`},
		{name: "format id duration with compact under params", configJSON: `{"format":{"id":"duration","params":{"compact":true}}}`},
		{name: "format type duration with decimals only", configJSON: `{"format":{"type":"duration","decimals":0}}`},
		{name: "format type duration with compact only", configJSON: `{"format":{"type":"duration","compact":true}}`},
		{name: "format type duration with decimals under params", configJSON: `{"format":{"type":"duration","params":{"decimals":0}}}`},
		{name: "format type duration with compact under params", configJSON: `{"format":{"type":"duration","params":{"compact":true}}}`},
		{name: "format type duration with decimals and compact under params", configJSON: `{"format":{"type":"duration","params":{"decimals":0,"compact":true}}}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.True(t, metricUsesUnsupportedDurationParams(jsontypes.NewNormalizedValue(tc.configJSON)))
		})
	}
}
