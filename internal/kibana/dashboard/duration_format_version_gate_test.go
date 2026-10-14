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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

const durationFormatParamsMinVersionMessage = `format.type "duration" with "decimals"/"compact" requires Elastic Stack 9.5.2 or later; remove these properties or upgrade Kibana.`

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

func requireDurationFormatRejected(t *testing.T, diags diag.Diagnostics, resultID types.String, dashboardWrites *atomic.Int32, pathParts ...string) {
	t.Helper()
	require.True(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	text := diagnosticText(diags)
	for _, part := range pathParts {
		require.Contains(t, text, part)
	}
	require.Contains(t, text, durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(0), dashboardWrites.Load())
	require.True(t, resultID.IsNull() || resultID.ValueString() == "")
}

func TestCreateDashboard_rejectsDurationFormatWithDecimalsAndCompactOnOlderStack(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.5.1"), &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{durationFormatXYPanel()}

	result, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	requireDurationFormatRejected(t, diags, result.Model.ID, &dashboardWrites, "panels[0]", `id "latency"`, "layer 0", "metric 0")
}

func TestUpdateDashboard_rejectsDurationFormatWithDecimalsAndCompactOnOlderStack(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.5.1"), &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringValue("dash-1"))
	planModel.Panels = []models.PanelModel{durationFormatXYPanel()}

	result, diags := updateDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	requireDurationFormatRejected(t, diags, result.Model.ID, &dashboardWrites, "panels[0]", `id "latency"`, "layer 0", "metric 0")
}

func TestCreateDashboard_propagatesStatusRetrievalFailureDuringVersionCheck(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusInternalServerError, "kibana status unavailable", &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{durationFormatXYPanel()}

	_, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.True(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	text := diagnosticText(diags)
	require.Contains(t, text, "Unexpected status code from server: got HTTP 500")
	require.Contains(t, text, "kibana status unavailable")
	require.NotContains(t, text, durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(0), dashboardWrites.Load())
}

func TestCreateDashboard_propagatesVersionParsingFailureDuringVersionCheck(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("not-a-version"), &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{durationFormatXYPanel()}

	_, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.True(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	text := diagnosticText(diags)
	require.Contains(t, text, "malformed version: not-a-version")
	require.NotContains(t, text, durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(0), dashboardWrites.Load())
}

func TestCreateDashboard_rejectsDurationFormatNestedInSectionOnOlderStack(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.5.1"), &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Sections = []models.SectionModel{{
		ID:     types.StringValue("overview"),
		Panels: []models.PanelModel{durationFormatXYPanel()},
	}}

	result, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	requireDurationFormatRejected(t, diags, result.Model.ID, &dashboardWrites, "sections[0].panels[0]", `id "latency"`, "layer 0", "metric 0")
}

func TestCreateDashboard_allowsDurationFormatParamsAtMinimumVersion(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.5.2"), &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{durationFormatXYPanel()}

	_, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.NotContains(t, diagnosticText(diags), durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(1), dashboardWrites.Load())
}

func TestUpdateDashboard_allowsDurationFormatParamsAtMinimumVersion(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.5.2"), &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringValue("dash-1"))
	planModel.Panels = []models.PanelModel{durationFormatXYPanel()}

	_, diags := updateDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.NotContains(t, diagnosticText(diags), durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(1), dashboardWrites.Load())
}

func TestUpdateDashboard_rejectsDurationFormatNestedInSectionOnOlderStack(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.5.1"), &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringValue("dash-1"))
	planModel.Sections = []models.SectionModel{{
		ID:     types.StringValue("overview"),
		Panels: []models.PanelModel{durationFormatXYPanel()},
	}}

	result, diags := updateDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	requireDurationFormatRejected(t, diags, result.Model.ID, &dashboardWrites, "sections[0].panels[0]", `id "latency"`, "layer 0", "metric 0")
}

func TestCreateDashboard_skipsVersionCheckWithoutDurationFormatParams(t *testing.T) {
	var dashboardWrites atomic.Int32
	var statusCalls atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.4.0"), &dashboardWrites, &statusCalls)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringNull())

	_, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.Equal(t, int32(1), dashboardWrites.Load())
	require.Equal(t, int32(0), statusCalls.Load())
}

func TestUpdateDashboard_skipsVersionCheckWithoutDurationFormatParams(t *testing.T) {
	var dashboardWrites atomic.Int32
	var statusCalls atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.4.0"), &dashboardWrites, &statusCalls)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringValue("dash-1"))

	_, diags := updateDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.Equal(t, int32(1), dashboardWrites.Load())
	require.Equal(t, int32(0), statusCalls.Load())
}

func TestUpdateDashboard_propagatesStatusRetrievalFailureDuringVersionCheck(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusInternalServerError, "kibana status unavailable", &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringValue("dash-1"))
	planModel.Panels = []models.PanelModel{durationFormatXYPanel()}

	_, diags := updateDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.True(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	text := diagnosticText(diags)
	require.Contains(t, text, "Unexpected status code from server: got HTTP 500")
	require.Contains(t, text, "kibana status unavailable")
	require.NotContains(t, text, durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(0), dashboardWrites.Load())
}

func TestUpdateDashboard_propagatesVersionParsingFailureDuringVersionCheck(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("not-a-version"), &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringValue("dash-1"))
	planModel.Panels = []models.PanelModel{durationFormatXYPanel()}

	_, diags := updateDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.True(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	text := diagnosticText(diags)
	require.Contains(t, text, "malformed version: not-a-version")
	require.NotContains(t, text, durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(0), dashboardWrites.Load())
}

func TestCreateDashboard_namesNestedLayerAndMetricIndexes(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.5.1"), &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
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

	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{{}, offending}

	result, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	requireDurationFormatRejected(t, diags, result.Model.ID, &dashboardWrites, "panels[1]", `id "latency"`, "layer 1", "metric 1")
}

func TestCreateDashboard_allowsDurationFormatParamsOnServerless(t *testing.T) {
	var dashboardWrites atomic.Int32
	statusBody := map[string]any{
		"version": map[string]any{
			"number":       "9.4.0",
			"build_flavor": "serverless",
		},
	}
	server := newDashboardWriteServer(t, http.StatusOK, statusBody, &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{durationFormatXYPanel()}

	_, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.NotContains(t, diagnosticText(diags), durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(1), dashboardWrites.Load())
}

func TestUpdateDashboard_allowsDurationFormatParamsOnServerless(t *testing.T) {
	var dashboardWrites atomic.Int32
	statusBody := map[string]any{
		"version": map[string]any{
			"number":       "9.4.0",
			"build_flavor": "serverless",
		},
	}
	server := newDashboardWriteServer(t, http.StatusOK, statusBody, &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	planModel := testDashboardPlanModel(types.StringValue("dash-1"))
	planModel.Panels = []models.PanelModel{durationFormatXYPanel()}

	_, diags := updateDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.NotContains(t, diagnosticText(diags), durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(1), dashboardWrites.Load())
}

func TestCreateDashboard_doesNotGateNumberFormatOnOlderStack(t *testing.T) {
	var dashboardWrites atomic.Int32
	var statusCalls atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.4.0"), &dashboardWrites, &statusCalls)

	scopedClient := newTestKibanaScopedClient(t, server)
	panel := durationFormatXYPanel()
	panel.VisConfig.ByValue.XYChartConfig.Layers[0].DataLayer.Y[0].ConfigJSON = jsontypes.NewNormalizedValue(
		`{"format":{"type":"number","decimals":0,"compact":true}}`,
	)
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{panel}

	_, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.Equal(t, int32(1), dashboardWrites.Load())
	require.Equal(t, int32(0), statusCalls.Load())
}

func TestCreateDashboard_skipsVersionCheckForNullFormat(t *testing.T) {
	var dashboardWrites atomic.Int32
	var statusCalls atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.4.0"), &dashboardWrites, &statusCalls)

	scopedClient := newTestKibanaScopedClient(t, server)
	panel := durationFormatXYPanel()
	panel.VisConfig.ByValue.XYChartConfig.Layers[0].DataLayer.Y[0].ConfigJSON = jsontypes.NewNormalizedValue(`{"format":null}`)
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{panel}

	_, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.NotContains(t, diagnosticText(diags), durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(1), dashboardWrites.Load())
	require.Equal(t, int32(0), statusCalls.Load())
}

func TestCreateDashboard_skipsVersionCheckForDurationFormatWithoutParams(t *testing.T) {
	var dashboardWrites atomic.Int32
	var statusCalls atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.4.0"), &dashboardWrites, &statusCalls)

	scopedClient := newTestKibanaScopedClient(t, server)
	panel := durationFormatXYPanel()
	panel.VisConfig.ByValue.XYChartConfig.Layers[0].DataLayer.Y[0].ConfigJSON = jsontypes.NewNormalizedValue(
		`{"format":{"type":"duration","from":"s","to":"ms"}}`,
	)
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{panel}

	_, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.NotContains(t, diagnosticText(diags), durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(1), dashboardWrites.Load())
	require.Equal(t, int32(0), statusCalls.Load())
}

func TestUpdateDashboard_doesNotGateNumberFormatOnOlderStack(t *testing.T) {
	var dashboardWrites atomic.Int32
	var statusCalls atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.4.0"), &dashboardWrites, &statusCalls)

	scopedClient := newTestKibanaScopedClient(t, server)
	panel := durationFormatXYPanel()
	panel.VisConfig.ByValue.XYChartConfig.Layers[0].DataLayer.Y[0].ConfigJSON = jsontypes.NewNormalizedValue(
		`{"format":{"type":"number","decimals":0,"compact":true}}`,
	)
	planModel := testDashboardPlanModel(types.StringValue("dash-1"))
	planModel.Panels = []models.PanelModel{panel}

	_, diags := updateDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.NotContains(t, diagnosticText(diags), durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(1), dashboardWrites.Load())
	require.Equal(t, int32(0), statusCalls.Load())
}

func TestCreateDashboard_doesNotGatePercentFormatOnOlderStack(t *testing.T) {
	var dashboardWrites atomic.Int32
	var statusCalls atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.4.0"), &dashboardWrites, &statusCalls)

	scopedClient := newTestKibanaScopedClient(t, server)
	panel := durationFormatXYPanel()
	panel.VisConfig.ByValue.XYChartConfig.Layers[0].DataLayer.Y[0].ConfigJSON = jsontypes.NewNormalizedValue(
		`{"format":{"type":"percent","decimals":0,"compact":true}}`,
	)
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{panel}

	_, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.NotContains(t, diagnosticText(diags), durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(1), dashboardWrites.Load())
	require.Equal(t, int32(0), statusCalls.Load())
}

func TestUpdateDashboard_skipsVersionCheckForDurationFormatWithoutParams(t *testing.T) {
	var dashboardWrites atomic.Int32
	var statusCalls atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.4.0"), &dashboardWrites, &statusCalls)

	scopedClient := newTestKibanaScopedClient(t, server)
	panel := durationFormatXYPanel()
	panel.VisConfig.ByValue.XYChartConfig.Layers[0].DataLayer.Y[0].ConfigJSON = jsontypes.NewNormalizedValue(
		`{"format":{"type":"duration","from":"s","to":"ms"}}`,
	)
	planModel := testDashboardPlanModel(types.StringValue("dash-1"))
	planModel.Panels = []models.PanelModel{panel}

	_, diags := updateDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	require.False(t, diags.HasError(), "diagnostics: %s", diagnosticText(diags))
	require.NotContains(t, diagnosticText(diags), durationFormatParamsMinVersionMessage)
	require.Equal(t, int32(1), dashboardWrites.Load())
	require.Equal(t, int32(0), statusCalls.Load())
}

func TestCreateDashboard_rejectsDurationFormatShapesOnOlderStack(t *testing.T) {
	tests := []struct {
		name       string
		configJSON string
	}{
		{
			name:       "format id duration with decimals only",
			configJSON: `{"format":{"id":"duration","decimals":0}}`,
		},
		{
			name:       "format id duration with compact only",
			configJSON: `{"format":{"id":"duration","compact":true}}`,
		},
		{
			name:       "format id duration with decimals and compact",
			configJSON: `{"format":{"id":"duration","decimals":0,"compact":true}}`,
		},
		{
			name:       "format id duration with decimals under params",
			configJSON: `{"format":{"id":"duration","params":{"decimals":0}}}`,
		},
		{
			name:       "format id duration with compact under params",
			configJSON: `{"format":{"id":"duration","params":{"compact":true}}}`,
		},
		{
			name:       "format type duration with decimals only",
			configJSON: `{"format":{"type":"duration","decimals":0}}`,
		},
		{
			name:       "format type duration with compact only",
			configJSON: `{"format":{"type":"duration","compact":true}}`,
		},
		{
			name:       "format type duration with decimals under params",
			configJSON: `{"format":{"type":"duration","params":{"decimals":0}}}`,
		},
		{
			name:       "format type duration with compact under params",
			configJSON: `{"format":{"type":"duration","params":{"compact":true}}}`,
		},
		{
			name:       "format type duration with decimals and compact under params",
			configJSON: `{"format":{"type":"duration","params":{"decimals":0,"compact":true}}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var dashboardWrites atomic.Int32
			server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.5.1"), &dashboardWrites, nil)

			scopedClient := newTestKibanaScopedClient(t, server)
			panel := durationFormatXYPanel()
			panel.VisConfig.ByValue.XYChartConfig.Layers[0].DataLayer.Y[0].ConfigJSON = jsontypes.NewNormalizedValue(tc.configJSON)
			planModel := testDashboardPlanModel(types.StringNull())
			planModel.Panels = []models.PanelModel{panel}

			result, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
				Plan:    planModel,
				WriteID: planModel.GetResourceID().ValueString(),
				SpaceID: planModel.GetSpaceID().ValueString(),
			})

			requireDurationFormatRejected(t, diags, result.Model.ID, &dashboardWrites, "panels[0]", `id "latency"`, "layer 0", "metric 0")
		})
	}
}

func TestCreateDashboard_rejectsDurationFormatWithNullPanelIDOnOlderStack(t *testing.T) {
	var dashboardWrites atomic.Int32
	server := newDashboardWriteServer(t, http.StatusOK, statefulStatusBody("9.5.1"), &dashboardWrites, nil)

	scopedClient := newTestKibanaScopedClient(t, server)
	panel := durationFormatXYPanel()
	panel.ID = types.StringNull()
	planModel := testDashboardPlanModel(types.StringNull())
	planModel.Panels = []models.PanelModel{panel}

	result, diags := createDashboard(t.Context(), scopedClient, entitycore.KibanaWriteRequest[models.DashboardModel]{
		Plan:    planModel,
		WriteID: planModel.GetResourceID().ValueString(),
		SpaceID: planModel.GetSpaceID().ValueString(),
	})

	requireDurationFormatRejected(t, diags, result.Model.ID, &dashboardWrites, "panels[0]", "layer 0", "metric 0")
}
