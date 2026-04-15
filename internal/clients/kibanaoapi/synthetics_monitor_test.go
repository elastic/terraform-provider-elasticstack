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

package kibanaoapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient creates a minimal kibanaoapi.Client backed by the given test server.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := NewClient(Config{URL: srv.URL})
	require.NoError(t, err)
	return c
}

func TestGetMonitor_200(t *testing.T) {
	enabled := true
	monitor := SyntheticsMonitor{
		ID:        "abc123",
		Name:      "my-http-monitor",
		Type:      SyntheticsMonitorTypeHTTP,
		Namespace: "default",
		Enabled:   &enabled,
		Schedule:  &SyntheticsMonitorSchedule{Number: "5", Unit: "m"},
		URL:       "https://example.com",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "abc123")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(monitor)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, diags := GetMonitor(context.Background(), client, "default", "abc123")
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, result)
	assert.Equal(t, "abc123", result.ID)
	assert.Equal(t, "my-http-monitor", result.Name)
	assert.Equal(t, SyntheticsMonitorTypeHTTP, result.Type)
}

func TestGetMonitor_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, diags := GetMonitor(context.Background(), client, "default", "nonexistent")
	assert.False(t, diags.HasError(), diags)
	assert.Nil(t, result, "expect nil result for 404")
}

func TestCreateMonitor_200(t *testing.T) {
	req := SyntheticsMonitorRequest{
		Type:      SyntheticsMonitorTypeHTTP,
		Name:      "new-monitor",
		Schedule:  5,
		Locations: []string{"us_east"},
		Labels:    map[string]string{},
		URL:       "https://example.com",
	}
	expectedResponse := SyntheticsMonitor{
		ID:   "created-id",
		Name: "new-monitor",
		Type: SyntheticsMonitorTypeHTTP,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		// Validate the request body contains required fields
		var body map[string]any
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "http", body["type"])
		assert.Equal(t, "new-monitor", body["name"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, diags := CreateMonitor(context.Background(), client, "default", req)
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, result)
	assert.Equal(t, "created-id", result.ID)
	assert.Equal(t, SyntheticsMonitorTypeHTTP, result.Type)
}

func TestUpdateMonitor_200(t *testing.T) {
	req := SyntheticsMonitorRequest{
		Type:   SyntheticsMonitorTypeHTTP,
		Name:   "updated-monitor",
		Labels: map[string]string{},
		URL:    "https://updated.example.com",
	}
	expectedResponse := SyntheticsMonitor{
		ID:   "monitor-id",
		Name: "updated-monitor",
		Type: SyntheticsMonitorTypeHTTP,
		URL:  "https://updated.example.com",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Contains(t, r.URL.Path, "monitor-id")

		var body map[string]any
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "updated-monitor", body["name"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, diags := UpdateMonitor(context.Background(), client, "default", "monitor-id", req)
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, result)
	assert.Equal(t, "monitor-id", result.ID)
	assert.Equal(t, "updated-monitor", result.Name)
}

func TestDeleteMonitor_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		// Bulk delete endpoint: DELETE /api/synthetics/monitors with body {"ids": [...]}
		assert.Equal(t, "/api/synthetics/monitors", r.URL.Path)
		var body map[string]any
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		ids, _ := body["ids"].([]any)
		assert.Len(t, ids, 1)
		assert.Equal(t, "monitor-id", ids[0])
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	diags := DeleteMonitor(context.Background(), client, "default", "monitor-id")
	assert.False(t, diags.HasError(), diags)
}

func TestDeleteMonitor_404(t *testing.T) {
	// 404 on delete should be treated as success (already deleted)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	diags := DeleteMonitor(context.Background(), client, "default", "monitor-id")
	assert.False(t, diags.HasError(), diags)
}

func TestCreateMonitor_SpaceAwarePath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For non-default spaces, path should be /s/{spaceID}/api/synthetics/monitors
		assert.Contains(t, r.URL.Path, "/s/my-space/")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SyntheticsMonitor{
			ID:   "space-monitor-id",
			Type: SyntheticsMonitorTypeHTTP,
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	req := SyntheticsMonitorRequest{
		Type:   SyntheticsMonitorTypeHTTP,
		Name:   "space-monitor",
		Labels: map[string]string{},
		URL:    "https://example.com",
	}
	result, diags := CreateMonitor(context.Background(), client, "my-space", req)
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, result)
}

func TestSyntheticsMonitorRequestMarshal_TypeDiscriminator(t *testing.T) {
	// Verify that JSON marshaling preserves type discriminator and flat field layout
	// consistent with the Kibana API wire format
	testcases := []struct {
		name     string
		req      SyntheticsMonitorRequest
		wantType string
		wantKey  string
	}{
		{
			name: "HTTP monitor has type=http and url field",
			req: SyntheticsMonitorRequest{
				Type:   SyntheticsMonitorTypeHTTP,
				Name:   "http-mon",
				Labels: map[string]string{},
				URL:    "https://example.com",
			},
			wantType: "http",
			wantKey:  "url",
		},
		{
			name: "TCP monitor has type=tcp and host field",
			req: SyntheticsMonitorRequest{
				Type:   SyntheticsMonitorTypeTCP,
				Name:   "tcp-mon",
				Labels: map[string]string{},
				Host:   "example.com:9200",
			},
			wantType: "tcp",
			wantKey:  "host",
		},
		{
			name: "ICMP monitor has type=icmp",
			req: SyntheticsMonitorRequest{
				Type:   SyntheticsMonitorTypeICMP,
				Name:   "icmp-mon",
				Labels: map[string]string{},
				Host:   "8.8.8.8",
			},
			wantType: "icmp",
			wantKey:  "host",
		},
		{
			name: "Browser monitor has type=browser and inline_script field",
			req: SyntheticsMonitorRequest{
				Type:         SyntheticsMonitorTypeBrowser,
				Name:         "browser-mon",
				Labels:       map[string]string{},
				InlineScript: "step('go', () => page.goto('https://example.com'))",
			},
			wantType: "browser",
			wantKey:  "inline_script",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.req)
			require.NoError(t, err)

			var m map[string]any
			require.NoError(t, json.Unmarshal(data, &m))

			assert.Equal(t, tc.wantType, m["type"], "type discriminator mismatch")
			assert.Contains(t, m, tc.wantKey, "expected key %q in marshaled request", tc.wantKey)
			// Labels must always be present (even empty) per REQ-023
			assert.Contains(t, m, "labels", "labels must always be present")
		})
	}
}
