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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
)

// newTestClient creates a fleet.Client backed by the given test server.
func newTestClient(t *testing.T, server *httptest.Server) *fleet.Client {
	t.Helper()
	c, err := fleet.NewClient(fleet.Config{URL: server.URL})
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}
	return c
}

func TestGetPackages_SpaceAwarePath(t *testing.T) {
	tests := []struct {
		name        string
		spaceID     string
		wantPathPfx string
	}{
		{
			name:        "no space id uses default path",
			spaceID:     "",
			wantPathPfx: "/api/fleet/epm/packages",
		},
		{
			name:        "default space uses default path",
			spaceID:     "default",
			wantPathPfx: "/api/fleet/epm/packages",
		},
		{
			name:        "custom space id prefixes path with /s/{space_id}",
			spaceID:     "my-space",
			wantPathPfx: "/s/my-space/api/fleet/epm/packages",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedPath string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				// Return a minimal valid response so GetPackages doesn't error.
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"items":[]}`)
			}))
			defer srv.Close()

			client := newTestClient(t, srv)
			_, diags := fleet.GetPackages(context.Background(), client, false, tc.spaceID)
			if diags.HasError() {
				t.Fatalf("GetPackages returned unexpected error: %v", diags.Errors())
			}

			if !strings.HasPrefix(capturedPath, tc.wantPathPfx) {
				t.Errorf("request path = %q, want prefix %q", capturedPath, tc.wantPathPfx)
			}
		})
	}
}
