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

package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestKibanaClient(t *testing.T, handler http.HandlerFunc) *clients.KibanaScopedClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	oapiClient, err := kibanaoapi.NewClient(kibanaoapi.Config{URL: srv.URL})
	require.NoError(t, err)
	return clients.NewKibanaScopedClientForTest(oapiClient)
}

func respondJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func httpMonitorPlan() tfModelV0 {
	return tfModelV0{
		ID:        types.StringUnknown(),
		Name:      types.StringValue("issue-4986"),
		SpaceID:   types.StringUnknown(),
		Locations: []types.String{types.StringValue("us_west")},
		HTTP: &tfHTTPMonitorFieldsV0{
			URL: types.StringValue("http://localhost:5601"),
		},
	}
}

func warningsOf(diags diag.Diagnostics) diag.Diagnostics {
	return diags.Warnings()
}

func TestCreateMonitor(t *testing.T) {
	const monitorID = "e3c90fb3-46c0-46d2-8827-7524f234f58d"

	testcases := []struct {
		name         string
		spaceID      string
		body         string
		wantID       string
		wantError    string
		wantWarnings []string
	}{
		{
			name:    "push error body",
			spaceID: "my-space",
			body:    `{"message":"error pushing monitor to the service","attributes":{"errors":[{"locationId":"us_west"}]},"id":"` + monitorID + `"}`,
			wantID:  "my-space/" + monitorID,
			wantWarnings: []string{
				`The synthetics monitor "issue-4986" (my-space/` + monitorID + `) was created and is tracked in Terraform state.`,
				"- us_west (no details reported)",
				"these errors may relate to other monitors",
				"Kibana retries syncing periodically.",
			},
		},
		{
			name:   "monitor body in the default space",
			body:   `{"id":"` + monitorID + `","type":"http","name":"issue-4986","url":"http://localhost:5601"}`,
			wantID: "/" + monitorID,
		},
		{
			name:      "body without id",
			spaceID:   "my-space",
			body:      `{"message":"error pushing monitor to the service","attributes":{"errors":[{"locationId":"us_west"}]}}`,
			wantError: "did not include the monitor ID",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			client := newTestKibanaClient(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				respondJSON(w, tc.body)
			})

			plan := httpMonitorPlan()
			result, diags := createMonitor(context.Background(), client, entitycore.KibanaWriteRequest[tfModelV0]{
				Plan:    plan,
				SpaceID: tc.spaceID,
			})

			if tc.wantError != "" {
				require.True(t, diags.HasError())
				assert.Contains(t, diags.Errors()[0].Detail(), tc.wantError)
				return
			}
			require.False(t, diags.HasError(), diags)

			assert.False(t, result.SkipReadAfterWrite)
			assert.Equal(t, tc.wantID, result.Model.ID.ValueString())
			assert.Equal(t, tc.spaceID, result.Model.SpaceID.ValueString())
			assert.Equal(t, plan.Name, result.Model.Name)
			assert.Equal(t, plan.HTTP, result.Model.HTTP)

			warnings := warningsOf(diags)
			if len(tc.wantWarnings) == 0 {
				assert.Empty(t, warnings)
				return
			}
			require.Len(t, warnings, 1)
			assert.Equal(t, "Kibana reported Synthetics Service sync errors", warnings[0].Summary())
			for _, want := range tc.wantWarnings {
				assert.Contains(t, warnings[0].Detail(), want)
			}
		})
	}
}
