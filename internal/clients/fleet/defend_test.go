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

// Tests for the Defend-specific Fleet package policy helpers.
// These verify query-format selection: Defend helpers do NOT send
// format=simplified (typed path) while generic helpers DO (mapped path).

package fleet_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
)

// newTestClient creates a fleet.Client wired to the given test server.
func newTestClient(t *testing.T, server *httptest.Server) *fleetclient.Client {
	t.Helper()
	client, err := fleetclient.NewClient(fleetclient.Config{
		URL: server.URL,
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	return client
}

// TestGetDefendPackagePolicyDoesNotUseSimplifiedFormat verifies that
// GetDefendPackagePolicy does NOT append ?format=simplified to the request URL.
// This ensures the typed Defend response (with typed inputs, config, version) is
// returned rather than the simplified mapped-input format.
func TestGetDefendPackagePolicyDoesNotUseSimplifiedFormat(t *testing.T) {
	var capturedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		// Return a minimal valid Defend policy response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"item": map[string]interface{}{
				"id":      "policy-123",
				"name":    "test-endpoint",
				"enabled": true,
				"inputs":  []interface{}{},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server)
	ctx := t.Context()

	_, _ = fleetclient.GetDefendPackagePolicy(ctx, client, "policy-123", "")

	// The Defend path must NOT include format=simplified
	if contains(capturedURL, "format=simplified") {
		t.Errorf("GetDefendPackagePolicy sent format=simplified; typed Defend responses require raw format. URL: %s", capturedURL)
	}

	// The path should include the policy ID
	if !contains(capturedURL, "policy-123") {
		t.Errorf("GetDefendPackagePolicy URL does not contain policy ID. URL: %s", capturedURL)
	}
}

// TestGetPackagePolicyUsesSimplifiedFormat verifies that the generic
// GetPackagePolicy helper sends format=simplified. This is a regression guard
// ensuring the mapped-input path remains unchanged after Defend was added.
func TestGetPackagePolicyUsesSimplifiedFormat(t *testing.T) {
	var capturedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"item": map[string]interface{}{
				"id":       "policy-456",
				"name":     "generic-policy",
				"enabled":  true,
				"inputs":   map[string]interface{}{},
				"revision": 1,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server)
	ctx := t.Context()

	_, _ = fleetclient.GetPackagePolicy(ctx, client, "policy-456", "")

	// Generic path MUST include format=simplified
	if !contains(capturedURL, "format=simplified") {
		t.Errorf("GetPackagePolicy did not send format=simplified; the mapped-input path requires simplified format. URL: %s", capturedURL)
	}
}

// TestCreateDefendPackagePolicyDoesNotUseSimplifiedFormat verifies that
// CreateDefendPackagePolicy does NOT send format=simplified.
func TestCreateDefendPackagePolicyDoesNotUseSimplifiedFormat(t *testing.T) {
	var capturedURL string
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"item": map[string]interface{}{
				"id":      "new-policy-789",
				"name":    "endpoint-policy",
				"enabled": true,
				"inputs":  []interface{}{},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server)
	ctx := t.Context()

	req := kbapi.DefendPackagePolicyRequest{
		Name: "endpoint-policy",
		Package: kbapi.PackagePolicyRequestPackage{
			Name:    "endpoint",
			Version: "8.14.0",
		},
		Inputs: []kbapi.DefendPackagePolicyRequestInput{
			{
				Type:    "ENDPOINT_INTEGRATION_CONFIG",
				Enabled: true,
				Streams: []interface{}{},
			},
		},
	}

	_, _ = fleetclient.CreateDefendPackagePolicy(ctx, client, "", req)

	if contains(capturedURL, "format=simplified") {
		t.Errorf("CreateDefendPackagePolicy sent format=simplified; typed path must not use simplified format. URL: %s", capturedURL)
	}

	// Verify the request body contains typed inputs (array, not map)
	var body map[string]interface{}
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("could not unmarshal request body: %v", err)
	}

	inputs, ok := body["inputs"]
	if !ok {
		t.Fatal("request body does not contain inputs")
	}

	if _, ok := inputs.([]interface{}); !ok {
		t.Errorf("Defend create request inputs must be a list, got %T", inputs)
	}
}

// TestSpaceAwareDefendPath verifies that Defend helpers correctly prepend
// /s/{spaceID} when a space ID is provided.
func TestSpaceAwareDefendPath(t *testing.T) {
	var capturedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"item": map[string]interface{}{
				"id":      "policy-123",
				"name":    "test",
				"enabled": true,
				"inputs":  []interface{}{},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server)
	ctx := t.Context()

	_, _ = fleetclient.GetDefendPackagePolicy(ctx, client, "policy-123", "my-space")

	if !contains(capturedURL, "/s/my-space/") {
		t.Errorf("Defend helper did not prepend space path. URL: %s", capturedURL)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
