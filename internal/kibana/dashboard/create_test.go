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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

const dashboardAPIPathPrefix = "/api/dashboards/"

var (
	testServerNormalizedID = "server-normalized-id"
	testEmptyDashboardID   = ""
)

func testDashboardPlanModel(dashboardID types.String) models.DashboardModel {
	return models.DashboardModel{
		SpaceID:     types.StringValue("default"),
		DashboardID: dashboardID,
		Title:       types.StringValue("Test Dashboard"),
		TimeRange: &models.TimeRangeModel{
			From: types.StringValue("now-15m"),
			To:   types.StringValue("now"),
		},
		RefreshInterval: &models.RefreshIntervalModel{
			Pause: types.BoolValue(true),
			Value: types.Int64Value(90000),
		},
		Query: &models.DashboardQueryModel{
			Language: types.StringValue("kql"),
			Text:     types.StringValue(""),
		},
	}
}

func newTestKibanaScopedClient(t *testing.T, server *httptest.Server) *clients.KibanaScopedClient {
	t.Helper()
	t.Setenv("KIBANA_ENDPOINT", server.URL)

	scopedClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)
	require.NotNil(t, scopedClient.GetKibanaOapiClient())
	return scopedClient
}

func TestCreateDashboard(t *testing.T) {
	type testCase struct {
		name              string
		dashboardID       types.String
		putStatusCode     int
		putResponseID     *string
		expectPOST        bool
		expectPUT         bool
		expectPUTID       string
		expectDashboardID string
		expectCompositeID string
		expectError       bool
	}

	tests := []testCase{
		{
			name:              "POST when dashboard_id is null",
			dashboardID:       types.StringNull(),
			expectPOST:        true,
			expectDashboardID: "auto-generated-uuid",
			expectCompositeID: "default/auto-generated-uuid",
		},
		{
			name:              "POST when dashboard_id is unknown",
			dashboardID:       types.StringUnknown(),
			expectPOST:        true,
			expectDashboardID: "auto-generated-uuid",
			expectCompositeID: "default/auto-generated-uuid",
		},
		{
			name:              "PUT when dashboard_id is known",
			dashboardID:       types.StringValue("my-team-overview"),
			putStatusCode:     http.StatusCreated,
			expectPUT:         true,
			expectPUTID:       "my-team-overview",
			expectDashboardID: "my-team-overview",
			expectCompositeID: "default/my-team-overview",
		},
		{
			name:              "PUT accepts 200 OK response",
			dashboardID:       types.StringValue("existing-dashboard"),
			putStatusCode:     http.StatusOK,
			expectPUT:         true,
			expectPUTID:       "existing-dashboard",
			expectDashboardID: "existing-dashboard",
			expectCompositeID: "default/existing-dashboard",
		},
		{
			name:              "PUT uses server-returned id when response differs from request",
			dashboardID:       types.StringValue("my-team-overview"),
			putStatusCode:     http.StatusCreated,
			putResponseID:     &testServerNormalizedID,
			expectPUT:         true,
			expectPUTID:       "my-team-overview",
			expectDashboardID: "server-normalized-id",
			expectCompositeID: "default/server-normalized-id",
		},
		{
			name:          "PUT rejects empty dashboard id in response",
			dashboardID:   types.StringValue("my-id"),
			putStatusCode: http.StatusCreated,
			putResponseID: &testEmptyDashboardID,
			expectPUT:     true,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var postCalls atomic.Int32
			var putCalls atomic.Int32
			var putPathID atomic.Value

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("Content-Type", "application/json")

				switch {
				case req.Method == http.MethodPost && req.URL.Path == "/api/dashboards":
					postCalls.Add(1)
					rw.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(rw).Encode(map[string]string{"id": "auto-generated-uuid"})
				case req.Method == http.MethodPut && strings.HasPrefix(req.URL.Path, dashboardAPIPathPrefix):
					putCalls.Add(1)
					id := strings.TrimPrefix(req.URL.Path, dashboardAPIPathPrefix)
					putPathID.Store(id)
					responseID := id
					if tt.putResponseID != nil {
						responseID = *tt.putResponseID
					}
					rw.WriteHeader(tt.putStatusCode)
					_ = json.NewEncoder(rw).Encode(map[string]string{"id": responseID})
				default:
					t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
				}
			}))
			defer server.Close()

			scopedClient := newTestKibanaScopedClient(t, server)
			planModel := testDashboardPlanModel(tt.dashboardID)

			writeReq := entitycore.KibanaWriteRequest[models.DashboardModel]{
				Plan: planModel,
			}
			if typeutils.IsKnown(planModel.DashboardID) {
				writeReq.WriteID = planModel.DashboardID.ValueString()
			}

			result, diags := createDashboard(t.Context(), scopedClient, writeReq)

			require.Equal(t, tt.expectError, diags.HasError(), "diagnostics: %v", diags)
			if tt.expectError {
				require.Contains(t, diags[0].Summary(), "Dashboard create returned empty id")
				return
			}

			if tt.expectPOST {
				require.Equal(t, int32(1), postCalls.Load())
				require.Equal(t, int32(0), putCalls.Load())
			}
			if tt.expectPUT {
				require.Equal(t, int32(0), postCalls.Load())
				require.Equal(t, int32(1), putCalls.Load())
				require.Equal(t, tt.expectPUTID, putPathID.Load())
			}

			require.Equal(t, tt.expectCompositeID, result.Model.ID.ValueString())
			require.Equal(t, tt.expectDashboardID, result.Model.DashboardID.ValueString())
		})
	}
}
