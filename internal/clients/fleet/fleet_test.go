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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newTestClient creates a fleet.Client backed by the given test server.
func newTestClient(t *testing.T, server *httptest.Server) *fleet.Client {
	t.Helper()
	c, err := fleet.NewClient(fleet.Config{URL: server.URL})
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}
	return c
}

// newTestClientWithRoundTripper creates a fleet.Client that sends all requests
// through rt without using a real server.
func newTestClientWithRoundTripper(t *testing.T, rt http.RoundTripper) *fleet.Client {
	t.Helper()
	const endpoint = "http://kibana.test/"
	httpClient := &http.Client{Transport: rt}
	apiClient, err := kbapi.NewClientWithResponses(endpoint, kbapi.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("newTestClientWithRoundTripper: %v", err)
	}
	return &fleet.Client{
		URL:  "http://kibana.test",
		HTTP: httpClient,
		API:  apiClient,
	}
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

func TestGetPackages_NonJSONHTTP200ReturnsDiagnostic(t *testing.T) {
	tests := []struct {
		name       string
		prerelease bool
	}{
		{
			name:       "primary path returns diagnostic",
			prerelease: false,
		},
		{
			name:       "retry path returns diagnostic",
			prerelease: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestCount := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requestCount++
				w.Header().Set("Content-Type", "text/plain")
				if tc.prerelease && requestCount == 1 {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprint(w, `{"message":"definition for this key is missing: prerelease"}`)
					return
				}
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"items":[]}`)
			}))
			defer srv.Close()

			client := newTestClient(t, srv)
			items, diags := fleet.GetPackages(context.Background(), client, tc.prerelease, "")

			require.Nil(t, items)
			require.True(t, diags.HasError())
			require.Len(t, diags, 1)
			require.Equal(t, "Unexpected Fleet response", diags[0].Summary())
			require.Contains(t, diags[0].Detail(), "Fleet returned HTTP 200 for the packages list endpoint but the response body could not be decoded as JSON")
			if tc.prerelease {
				require.Equal(t, 2, requestCount)
			} else {
				require.Equal(t, 1, requestCount)
			}
		})
	}
}

func TestGetPackages_JSONHTTP200ReturnsItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"items":[{"name":"system","version":"1.0.0"}]}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	items, diags := fleet.GetPackages(context.Background(), client, false, "")

	require.False(t, diags.HasError())
	require.Len(t, items, 1)
	require.Equal(t, "system", items[0].Name)
	require.Equal(t, "1.0.0", items[0].Version)
}
