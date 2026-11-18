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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMonitor(t *testing.T) {
	const monitorID = "e3c90fb3-46c0-46d2-8827-7524f234f58d"
	const getBody = `{"id":"` + monitorID + `","type":"http","name":"issue-4986","url":"http://localhost:5601","locations":[{"id":"us_west","label":"US West","isServiceManaged":true}]}`

	testcases := []struct {
		name         string
		putBody      string
		wantWarnings []string
	}{
		{
			name:    "push error body",
			putBody: `{"message":"error pushing monitor to the service","attributes":{"errors":[{"locationId":"us_west","error":{"reason":"boom","status":500}}]}}`,
			wantWarnings: []string{
				`The synthetics monitor "issue-4986" (my-space/` + monitorID + `) was saved in Kibana, but pushing it to the Synthetics Service failed`,
				"- us_west: boom (HTTP 500)",
				"Kibana retries syncing periodically.",
			},
		},
		{
			name:    "no push errors",
			putBody: `{"warnings":[]}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			client := newTestKibanaClient(t, func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPut:
					respondJSON(w, tc.putBody)
				case http.MethodGet:
					respondJSON(w, getBody)
				default:
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			})

			plan := httpMonitorPlan()
			plan.ID = types.StringValue("my-space/" + monitorID)
			plan.SpaceID = types.StringValue("my-space")
			result, diags := updateMonitor(context.Background(), client, entitycore.KibanaWriteRequest[tfModelV0]{
				Plan:    plan,
				SpaceID: "my-space",
				WriteID: monitorID,
			})
			require.False(t, diags.HasError(), diags)
			assert.Equal(t, "my-space/"+monitorID, result.Model.ID.ValueString())

			warnings := diags.Warnings()
			if len(tc.wantWarnings) == 0 {
				assert.Empty(t, warnings)
				return
			}
			require.Len(t, warnings, 1)
			assert.Equal(t, "Synthetics monitor saved, but syncing to Elastic-managed locations failed", warnings[0].Summary())
			for _, want := range tc.wantWarnings {
				assert.Contains(t, warnings[0].Detail(), want)
			}
		})
	}
}
