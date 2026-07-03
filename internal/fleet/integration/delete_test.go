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

package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPackageName    = "system"
	testPackageVersion = "1.0.0"
	testSpaceID        = "default"
)

func newTestFleetClient(t *testing.T, server *httptest.Server) *fleet.Client {
	t.Helper()
	client, err := fleet.NewClient(fleet.Config{URL: server.URL})
	require.NoError(t, err)
	return client
}

func packageUninstallPath(name, version string) string {
	return fmt.Sprintf("/api/fleet/epm/packages/%s/%s", name, version)
}

func TestDeleteKibanaAssetsWithFallback_installSpace400FallsBackToUninstall(t *testing.T) {
	t.Parallel()

	var uninstallCalls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/kibana_assets"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"statusCode":400,"error":"Bad Request","message":"Impossible to delete kibana assets from the space where the package was installed, you must uninstall the package."}`)
		case r.Method == http.MethodDelete && r.URL.Path == packageUninstallPath(testPackageName, testPackageVersion):
			uninstallCalls++
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestFleetClient(t, srv)
	diags := deleteKibanaAssetsWithFallback(context.Background(), client, testPackageName, testPackageVersion, testSpaceID, true)

	assert.False(t, diags.HasError())
	assert.Equal(t, 1, uninstallCalls)
}

func TestDeleteKibanaAssetsWithFallback_other400DoesNotFallback(t *testing.T) {
	t.Parallel()

	var uninstallCalls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/kibana_assets"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"statusCode":400,"error":"Bad Request","message":"Some other validation error"}`)
		case r.Method == http.MethodDelete && r.URL.Path == packageUninstallPath(testPackageName, testPackageVersion):
			uninstallCalls++
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestFleetClient(t, srv)
	diags := deleteKibanaAssetsWithFallback(context.Background(), client, testPackageName, testPackageVersion, testSpaceID, true)

	require.True(t, diags.HasError())
	assert.Equal(t, 0, uninstallCalls)
	assert.Contains(t, diags.Errors()[0].Detail(), "Some other validation error")
}

func TestDeleteKibanaAssetsWithFallback_successDoesNotCallUninstall(t *testing.T) {
	t.Parallel()

	var uninstallCalls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/kibana_assets"):
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == packageUninstallPath(testPackageName, testPackageVersion):
			uninstallCalls++
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestFleetClient(t, srv)
	diags := deleteKibanaAssetsWithFallback(context.Background(), client, testPackageName, testPackageVersion, testSpaceID, true)

	assert.False(t, diags.HasError())
	assert.Equal(t, 0, uninstallCalls)
}

func TestDiagnosticsContainInstallSpaceDeleteRejection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		diags diag.Diagnostics
		want  bool
	}{
		{
			name: "matches detail with normalized whitespace",
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unexpected status code from server: got HTTP 400",
					"{\"statusCode\":400,\"message\":\"Impossible to delete kibana assets from the space\nwhere the package was installed\"}",
				),
			},
			want: true,
		},
		{
			name: "does not match unrelated error",
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unexpected status code from server: got HTTP 400",
					`{"statusCode":400,"message":"Some other validation error"}`,
				),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, diagnosticsContainInstallSpaceDeleteRejection(tc.diags))
		})
	}
}
