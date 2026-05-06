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

package sourcemap_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	tftest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ────────────────────────────────────────────────────────────────────────────
// Response body helpers
// ────────────────────────────────────────────────────────────────────────────

// artifactResponseJSON builds an APMUISourceMapsResponse JSON body with a
// single artifact entry.
func artifactResponseJSON(id, bundleFilepath, serviceName, serviceVersion string) map[string]any {
	return map[string]any{
		"artifacts": []any{
			map[string]any{
				"id": id,
				"body": map[string]any{
					"bundleFilepath": bundleFilepath,
					"serviceName":    serviceName,
					"serviceVersion": serviceVersion,
				},
			},
		},
	}
}

// emptyArtifactsJSON returns a response with an empty artifacts list.
func emptyArtifactsJSON() map[string]any {
	return map[string]any{
		"artifacts": []any{},
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Multipart body parsing helpers
// ────────────────────────────────────────────────────────────────────────────

// parseMultipart reads all parts from a multipart/form-data body and returns a
// map from form-field name to raw content bytes.
func parseMultipart(contentType string, body []byte) (map[string][]byte, error) {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Type %q: %w", contentType, err)
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, fmt.Errorf("Content-Type %q is missing boundary parameter", contentType)
	}

	mr := multipart.NewReader(bytes.NewReader(body), boundary)
	result := make(map[string][]byte)
	for {
		p, nextErr := mr.NextPart()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return nil, nextErr
		}
		name := p.FormName()
		data, readErr := io.ReadAll(p)
		if readErr != nil {
			return nil, readErr
		}
		if _, exists := result[name]; !exists {
			result[name] = data
		}
	}
	return result, nil
}

// unitTestSourceMapJSON is the source map JSON string used in all unit tests.
const unitTestSourceMapJSON = `{"version":3,"file":"test.min.js","sources":["test.js"],"mappings":"AAAA"}`

// Terraform HCL config helpers for unit tests (provider configured inline).

func unitTestConfigJSON(bundleFilepath, serviceName, serviceVersion, kibanaEndpoint string) string {
	return `
provider "elasticstack" {
  elasticsearch {}
  kibana {
    endpoints = ["` + kibanaEndpoint + `"]
    username  = "elastic"
    password  = "changeme"
  }
}

resource "elasticstack_apm_source_map" "unit" {
  bundle_filepath = "` + bundleFilepath + `"
  service_name    = "` + serviceName + `"
  service_version = "` + serviceVersion + `"
  sourcemap = {
    json = "` + jsonEscape(unitTestSourceMapJSON) + `"
  }
}
`
}

func unitTestConfigBinary(bundleFilepath, serviceName, serviceVersion, sourceMapBase64, kibanaEndpoint string) string {
	return `
provider "elasticstack" {
  elasticsearch {}
  kibana {
    endpoints = ["` + kibanaEndpoint + `"]
    username  = "elastic"
    password  = "changeme"
  }
}

resource "elasticstack_apm_source_map" "unit" {
  bundle_filepath = "` + bundleFilepath + `"
  service_name    = "` + serviceName + `"
  service_version = "` + serviceVersion + `"
  sourcemap = {
    binary = "` + sourceMapBase64 + `"
  }
}
`
}

// jsonEscape escapes a string for safe embedding in a double-quoted HCL string.
func jsonEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`)
}

// ────────────────────────────────────────────────────────────────────────────
// 3.7 — Multipart form construction unit tests
//
// These tests start a local httptest server to capture the exact multipart body
// that the Create handler sends to the Kibana API, then assert on field names
// and content without needing a real Kibana instance.
// ────────────────────────────────────────────────────────────────────────────

// TestSourceMapCreate_MultipartJSON verifies that when sourcemap_json is set the
// multipart body contains the expected field names and the raw JSON string as the
// sourcemap file content.
func TestSourceMapCreate_MultipartJSON(t *testing.T) {
	t.Setenv("TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT", "true")

	const (
		targetID       = "artifact-json-001"
		bundleFilepath = "/static/js/app.min.js"
		serviceName    = "my-svc"
		serviceVersion = "1.2.3"
		sourceMapJSON  = `{"version":3,"file":"test.min.js","sources":["test.js"],"mappings":"AAAA"}`
	)

	var capturedBody []byte
	var capturedContentType string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "sourcemaps"):
			raw, _ := io.ReadAll(r.Body)
			capturedBody = raw
			capturedContentType = r.Header.Get("Content-Type")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": targetID})

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "sourcemaps"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(artifactResponseJSON(targetID, bundleFilepath, serviceName, serviceVersion))

		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "sourcemaps"):
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []tftest.TestStep{
			{
				Config: unitTestConfigJSON(bundleFilepath, serviceName, serviceVersion, srv.URL),
				Check: tftest.ComposeTestCheckFunc(
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "id", targetID),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "bundle_filepath", bundleFilepath),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "service_name", serviceName),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "service_version", serviceVersion),
					func(_ *terraform.State) error {
						if len(capturedBody) == 0 {
							return fmt.Errorf("no multipart body was captured from the create request")
						}
						parts, err := parseMultipart(capturedContentType, capturedBody)
						require.NoError(t, err)

						for field, expected := range map[string]string{
							"bundle_filepath": bundleFilepath,
							"service_name":    serviceName,
							"service_version": serviceVersion,
						} {
							v, ok := parts[field]
							require.True(t, ok, "multipart body must contain field %q", field)
							assert.Equal(t, expected, string(v), "field %q value mismatch", field)
						}

						smContent, ok := parts["sourcemap"]
						require.True(t, ok, "multipart body must contain 'sourcemap' file field")
						assert.JSONEq(t, sourceMapJSON, string(smContent),
							"sourcemap field must contain the raw JSON string from sourcemap_json")
						return nil
					},
				),
			},
		},
	})
}

// TestSourceMapCreate_MultipartBinary verifies that when sourcemap_binary is set
// the multipart body contains the decoded bytes (not the base64 string) as the
// sourcemap file field content.
func TestSourceMapCreate_MultipartBinary(t *testing.T) {
	t.Setenv("TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT", "true")

	const (
		targetID       = "artifact-bin-002"
		bundleFilepath = "/static/js/app.min.js"
		serviceName    = "bin-svc"
		serviceVersion = "2.0.0"
		sourceMapJSON  = `{"version":3,"file":"test.min.js","sources":["test.js"],"mappings":"AAAA"}`
	)
	sourceMapBase64 := base64.StdEncoding.EncodeToString([]byte(sourceMapJSON))
	expectedDecodedBytes := []byte(sourceMapJSON)

	var capturedBody []byte
	var capturedContentType string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "sourcemaps"):
			raw, _ := io.ReadAll(r.Body)
			capturedBody = raw
			capturedContentType = r.Header.Get("Content-Type")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": targetID})

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "sourcemaps"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(artifactResponseJSON(targetID, bundleFilepath, serviceName, serviceVersion))

		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "sourcemaps"):
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []tftest.TestStep{
			{
				Config: unitTestConfigBinary(bundleFilepath, serviceName, serviceVersion, sourceMapBase64, srv.URL),
				Check: tftest.ComposeTestCheckFunc(
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "id", targetID),
					func(_ *terraform.State) error {
						if len(capturedBody) == 0 {
							return fmt.Errorf("no multipart body was captured from the create request")
						}
						parts, err := parseMultipart(capturedContentType, capturedBody)
						require.NoError(t, err)

						smContent, ok := parts["sourcemap"]
						require.True(t, ok, "multipart body must contain 'sourcemap' file field")
						assert.Equal(t, expectedDecodedBytes, smContent,
							"sourcemap field must contain decoded bytes, not the base64 string")
						return nil
					},
				),
			},
		},
	})
}

// TestSourceMapCreate_MultipartBoundaryAndFieldNames verifies that the multipart
// Content-Type header contains a boundary parameter and all expected form-field
// names are present.
func TestSourceMapCreate_MultipartBoundaryAndFieldNames(t *testing.T) {
	t.Setenv("TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT", "true")

	const (
		targetID       = "artifact-boundary-003"
		bundleFilepath = "/static/js/main.js"
		serviceName    = "boundary-svc"
		serviceVersion = "3.0.0"
		sourceMapJSON  = `{"version":3,"file":"test.min.js","sources":["test.js"],"mappings":"AAAA"}`
	)

	var capturedContentType string
	var capturedBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "sourcemaps"):
			raw, _ := io.ReadAll(r.Body)
			capturedBody = raw
			capturedContentType = r.Header.Get("Content-Type")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": targetID})

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "sourcemaps"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(artifactResponseJSON(targetID, bundleFilepath, serviceName, serviceVersion))

		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "sourcemaps"):
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []tftest.TestStep{
			{
				Config: unitTestConfigJSON(bundleFilepath, serviceName, serviceVersion, srv.URL),
				Check: tftest.ComposeTestCheckFunc(
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "id", targetID),
					func(_ *terraform.State) error {
						require.NotEmpty(t, capturedContentType, "Content-Type header must be set")

						mediaType, params, err := mime.ParseMediaType(capturedContentType)
						require.NoError(t, err, "Content-Type must be a valid media type")
						assert.Equal(t, "multipart/form-data", mediaType,
							"Content-Type must be multipart/form-data")

						_, hasBoundary := params["boundary"]
						assert.True(t, hasBoundary, "Content-Type must include a boundary parameter")

						parts, parseErr := parseMultipart(capturedContentType, capturedBody)
						require.NoError(t, parseErr)

						for _, expectedField := range []string{"bundle_filepath", "service_name", "service_version", "sourcemap"} {
							_, ok := parts[expectedField]
							assert.True(t, ok, "multipart body must contain field %q", expectedField)
						}
						return nil
					},
				),
			},
		},
	})
}

// ────────────────────────────────────────────────────────────────────────────
// 3.8 — Read pagination loop unit tests
// ────────────────────────────────────────────────────────────────────────────

// TestSourceMapRead_FoundOnPage1 verifies that when the target artifact appears
// in the first page, the resource is populated with the correct attributes.
func TestSourceMapRead_FoundOnPage1(t *testing.T) {
	t.Setenv("TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT", "true")

	const (
		targetID       = "page1-artifact"
		bundleFilepath = "/static/js/app.min.js"
		serviceName    = "svc-page1"
		serviceVersion = "1.0.0"
		sourceMapJSON  = `{"version":3,"file":"test.min.js","sources":["test.js"],"mappings":"AAAA"}`
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "sourcemaps"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": targetID})

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "sourcemaps"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(artifactResponseJSON(targetID, bundleFilepath, serviceName, serviceVersion))

		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "sourcemaps"):
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []tftest.TestStep{
			{
				Config: unitTestConfigJSON(bundleFilepath, serviceName, serviceVersion, srv.URL),
				Check: tftest.ComposeTestCheckFunc(
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "id", targetID),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "bundle_filepath", bundleFilepath),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "service_name", serviceName),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "service_version", serviceVersion),
				),
			},
		},
	})
}

// TestSourceMapRead_FoundOnPage2 verifies that when the target artifact appears
// only on page 2 (page 1 returns a full page of other artifacts), read still
// finds and populates it correctly, exercising the pagination loop.
func TestSourceMapRead_FoundOnPage2(t *testing.T) {
	t.Setenv("TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT", "true")

	const (
		targetID       = "page2-artifact"
		bundleFilepath = "/static/js/app.min.js"
		serviceName    = "svc-page2"
		serviceVersion = "2.0.0"
		sourceMapJSON  = `{"version":3,"file":"test.min.js","sources":["test.js"],"mappings":"AAAA"}`
		// readPageSize must match the constant in read.go
		readPageSize = 100
	)

	var readCallCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "sourcemaps"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": targetID})

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "sourcemaps"):
			callN := readCallCount.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			if callN == 1 {
				// Page 1: return a full page of "other" artifacts so the loop continues.
				otherArtifacts := make([]any, readPageSize)
				for i := range otherArtifacts {
					otherArtifacts[i] = map[string]any{
						"id": fmt.Sprintf("other-%d", i),
						"body": map[string]any{
							"bundleFilepath": "/other.js",
							"serviceName":    "other-svc",
							"serviceVersion": "0.0.0",
						},
					}
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"artifacts": otherArtifacts})
			} else {
				// Page 2: return the target artifact.
				_ = json.NewEncoder(w).Encode(artifactResponseJSON(targetID, bundleFilepath, serviceName, serviceVersion))
			}

		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "sourcemaps"):
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []tftest.TestStep{
			{
				Config: unitTestConfigJSON(bundleFilepath, serviceName, serviceVersion, srv.URL),
				Check: tftest.ComposeTestCheckFunc(
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "id", targetID),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "bundle_filepath", bundleFilepath),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "service_name", serviceName),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "service_version", serviceVersion),
				),
			},
		},
	})
}

// TestSourceMapRead_NotFoundRemovesFromState verifies that when no artifact
// matches the state id on a refresh, the resource is removed from state without
// error, causing a non-empty plan.
func TestSourceMapRead_NotFoundRemovesFromState(t *testing.T) {
	t.Setenv("TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT", "true")

	const (
		targetID       = "not-found-artifact"
		bundleFilepath = "/static/js/app.min.js"
		serviceName    = "svc-notfound"
		serviceVersion = "1.0.0"
		sourceMapJSON  = `{"version":3,"file":"test.min.js","sources":["test.js"],"mappings":"AAAA"}`
	)

	// Track how many GET calls have been made.
	// The first two GETs return the artifact:
	//   call 1: inside Create (after POST)
	//   call 2: post-step refresh run by the test framework after step 1
	// All subsequent GETs return empty, simulating the artifact disappearing and
	// causing the resource to be removed from state on the next plan in step 2.
	var getCallCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "sourcemaps"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": targetID})

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "sourcemaps"):
			n := getCallCount.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if n <= 2 {
				// First two GETs: return the artifact so step 1 completes cleanly.
				_ = json.NewEncoder(w).Encode(artifactResponseJSON(targetID, bundleFilepath, serviceName, serviceVersion))
			} else {
				// Subsequent GETs: artifact gone, triggers removal from state.
				_ = json.NewEncoder(w).Encode(emptyArtifactsJSON())
			}

		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "sourcemaps"):
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []tftest.TestStep{
			// Step 1: Create succeeds; first read inside Create finds the artifact.
			{
				Config: unitTestConfigJSON(bundleFilepath, serviceName, serviceVersion, srv.URL),
				Check: tftest.ComposeTestCheckFunc(
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "id", targetID),
				),
			},
			// Step 2: On the next plan cycle the refresh GET returns empty, so the
			// resource is removed from state and a new create is planned.
			{
				Config:             unitTestConfigJSON(bundleFilepath, serviceName, serviceVersion, srv.URL),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestSourceMapRead_ArtifactBodyNil verifies that when the artifact is found but
// its Body field is nil, the read completes without error and the resource
// remains in state (body-derived attributes retain their prior values from the
// framework state rather than being cleared, because read.go only updates them
// when Body != nil).
func TestSourceMapRead_ArtifactBodyNil(t *testing.T) {
	t.Setenv("TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT", "true")

	const (
		targetID       = "nil-body-artifact"
		bundleFilepath = "/static/js/app.min.js"
		serviceName    = "svc-nilbody"
		serviceVersion = "1.0.0"
		sourceMapJSON  = `{"version":3,"file":"test.min.js","sources":["test.js"],"mappings":"AAAA"}`
	)

	var getCallCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "sourcemaps"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": targetID})

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "sourcemaps"):
			n := getCallCount.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if n == 1 {
				// First GET (inside Create): return artifact with full body.
				_ = json.NewEncoder(w).Encode(artifactResponseJSON(targetID, bundleFilepath, serviceName, serviceVersion))
			} else {
				// Subsequent GETs: artifact present but body is nil.
				_ = json.NewEncoder(w).Encode(map[string]any{
					"artifacts": []any{
						map[string]any{"id": targetID},
					},
				})
			}

		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "sourcemaps"):
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []tftest.TestStep{
			// Step 1: Create succeeds; initial read returns full body.
			{
				Config: unitTestConfigJSON(bundleFilepath, serviceName, serviceVersion, srv.URL),
				Check: tftest.ComposeTestCheckFunc(
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "id", targetID),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "bundle_filepath", bundleFilepath),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "service_name", serviceName),
					tftest.TestCheckResourceAttr("elasticstack_apm_source_map.unit", "service_version", serviceVersion),
				),
			},
			// Step 2: Refresh reads an artifact with nil body. The resource must
			// remain in state (id still set) and the plan must be empty (no drift).
			{
				Config:             unitTestConfigJSON(bundleFilepath, serviceName, serviceVersion, srv.URL),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
