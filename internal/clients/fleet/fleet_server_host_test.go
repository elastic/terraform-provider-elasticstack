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

package fleet_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/stretchr/testify/require"
)

func TestListFleetServerHosts(t *testing.T) {
	const twoHosts = `{"id":"a","name":"One","host_urls":["https://one:8220"],"is_default":true},` +
		`{"id":"b","name":"Two","host_urls":["https://two:8220"],"is_default":false}`

	tests := []struct {
		name      string
		spaceID   string
		status    int
		body      string
		wantPath  string
		wantIDs   []string
		wantError string
	}{
		{
			name:     "returns every host",
			status:   http.StatusOK,
			body:     `{"items":[` + twoHosts + `],"page":1,"perPage":10000,"total":2}`,
			wantPath: "/api/fleet/fleet_server_hosts",
			wantIDs:  []string{"a", "b"},
		},
		{
			name:     "sends the request through the given space",
			spaceID:  "ops",
			status:   http.StatusOK,
			body:     `{"items":[` + twoHosts + `],"page":1,"perPage":10000,"total":2}`,
			wantPath: "/s/ops/api/fleet/fleet_server_hosts",
			wantIDs:  []string{"a", "b"},
		},
		{
			name:     "no hosts",
			status:   http.StatusOK,
			body:     `{"items":[],"page":1,"perPage":10000,"total":0}`,
			wantPath: "/api/fleet/fleet_server_hosts",
			wantIDs:  []string{},
		},
		{
			name:      "fewer hosts returned than reported",
			status:    http.StatusOK,
			body:      `{"items":[` + twoHosts + `],"page":1,"perPage":2,"total":3}`,
			wantPath:  "/api/fleet/fleet_server_hosts",
			wantError: "the Fleet API returned 2 of 3 Fleet server hosts",
		},
		{
			name:      "not found is an error",
			status:    http.StatusNotFound,
			body:      `{"statusCode":404,"error":"Not Found","message":"missing"}`,
			wantPath:  "/api/fleet/fleet_server_hosts",
			wantError: "missing",
		},
		{
			name:      "server error",
			status:    http.StatusInternalServerError,
			body:      `{"statusCode":500,"error":"Internal Server Error","message":"boom"}`,
			wantPath:  "/api/fleet/fleet_server_hosts",
			wantError: "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			defer srv.Close()

			hosts, diags := fleet.ListFleetServerHosts(t.Context(), newTestClient(t, srv), tt.spaceID)

			require.Equal(t, tt.wantPath, gotPath)
			if tt.wantError != "" {
				require.True(t, diags.HasError())
				require.Contains(t, diags[0].Summary()+" "+diags[0].Detail(), tt.wantError)
				require.Nil(t, hosts)
				return
			}
			require.False(t, diags.HasError(), diags)
			ids := make([]string, len(hosts))
			for i, host := range hosts {
				ids[i] = host.Id
			}
			require.Equal(t, tt.wantIDs, ids)
		})
	}
}
