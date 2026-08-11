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

package fleet

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsManifestYAML(t *testing.T) {
	tests := []struct {
		name      string
		entryName string
		want      bool
	}{
		{"bare filename", "manifest.yml", true},
		{"top-level dir prefix", "mypackage-1.0.0/manifest.yml", true},
		{"nested", "mypackage/subdir/manifest.yml", true},
		{"nested manifest under data_stream", "mypackage-1.0.0/data_stream/logs/manifest.yml", true},
		{"different file", "mypackage-1.0.0/README.md", false},
		{"different yaml extension", "mypackage-1.0.0/manifest.yaml", false},
		{"empty string", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isManifestYAML(tc.entryName); got != tc.want {
				t.Errorf("isManifestYAML(%q) = %v, want %v", tc.entryName, got, tc.want)
			}
		})
	}
}

func TestExtractPackageNameVersion(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantName    string
		wantVersion string
	}{
		{
			name:        "plain values",
			content:     "name: my-package\nversion: 1.2.3\n",
			wantName:    "my-package",
			wantVersion: "1.2.3",
		},
		{
			name:        "quoted version",
			content:     "name: my-package\nversion: \"1.2.3\"\n",
			wantName:    "my-package",
			wantVersion: "1.2.3",
		},
		{
			name:        "single-quoted version",
			content:     "name: my-package\nversion: '1.0.0'\n",
			wantName:    "my-package",
			wantVersion: "1.0.0",
		},
		{
			name:        "name only",
			content:     "name: only-name\n",
			wantName:    "only-name",
			wantVersion: "",
		},
		{
			name:        "no name field",
			content:     "version: 1.0.0\nformat_version: 1.0.0\n",
			wantName:    "",
			wantVersion: "1.0.0",
		},
		{
			name:        "empty content",
			content:     "",
			wantName:    "",
			wantVersion: "",
		},
		{
			name:        "extra whitespace after colon",
			content:     "name:   my-package\nversion:   2.0.0\n",
			wantName:    "my-package",
			wantVersion: "2.0.0",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotName, gotVersion := extractPackageNameVersion([]byte(tc.content))
			if gotName != tc.wantName {
				t.Errorf("name = %q, want %q", gotName, tc.wantName)
			}
			if gotVersion != tc.wantVersion {
				t.Errorf("version = %q, want %q", gotVersion, tc.wantVersion)
			}
		})
	}
}

func TestContainsInstallSpaceDeleteRejection(t *testing.T) {
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
			name: "matches summary",
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Impossible to delete kibana assets from the space where the package was installed",
					"{\"statusCode\":400}",
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
		{
			name: "does not match warning severity even with sentinel text",
			diags: diag.Diagnostics{
				diag.NewWarningDiagnostic(
					"Unexpected status code from server: got HTTP 400",
					"Impossible to delete kibana assets from the space where the package was installed",
				),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ContainsInstallSpaceDeleteRejection(tc.diags); got != tc.want {
				t.Errorf("ContainsInstallSpaceDeleteRejection() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsPackageInstalled(t *testing.T) {
	tests := []struct {
		name string
		pkg  *kbapi.KibanaHTTPAPIsGetPackageInfo
		want bool
	}{
		{"nil package", nil, false},
		{"no installation info, no status", &kbapi.KibanaHTTPAPIsGetPackageInfo{}, false},
		{
			"enum installed",
			&kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstalled,
				},
			},
			true,
		},
		{
			"enum install_failed short-circuits even with legacy status installed",
			&kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstallFailed,
				},
				Status: new("installed"),
			},
			false,
		},
		{
			"enum installing falls back to legacy status installed",
			&kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstalling,
				},
				Status: new("Installed"),
			},
			true,
		},
		{"legacy status only, case-insensitive", &kbapi.KibanaHTTPAPIsGetPackageInfo{Status: new("INSTALLED")}, true},
		{"legacy status install_failed", &kbapi.KibanaHTTPAPIsGetPackageInfo{Status: new("install_failed")}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsPackageInstalled(tc.pkg))
		})
	}
}

func TestIsPackageInstallFailed(t *testing.T) {
	tests := []struct {
		name string
		pkg  *kbapi.KibanaHTTPAPIsGetPackageInfo
		want bool
	}{
		{"nil package", nil, false},
		{"no installation info, no status", &kbapi.KibanaHTTPAPIsGetPackageInfo{}, false},
		{
			"enum install_failed",
			&kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstallFailed,
				},
			},
			true,
		},
		{
			"enum installed short-circuits even with legacy status install_failed",
			&kbapi.KibanaHTTPAPIsGetPackageInfo{
				InstallationInfo: &kbapi.KibanaHTTPAPIsPackageInfoInstallationInfo{
					InstallStatus: kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstalled,
				},
				Status: new("install_failed"),
			},
			false,
		},
		{"legacy status install_failed, case-insensitive", &kbapi.KibanaHTTPAPIsGetPackageInfo{Status: new("INSTALL_FAILED")}, true},
		{"legacy status installed", &kbapi.KibanaHTTPAPIsGetPackageInfo{Status: new("installed")}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsPackageInstallFailed(tc.pkg))
		})
	}
}

func newTestPackagesClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	client, err := NewClient(Config{URL: server.URL})
	require.NoError(t, err)
	return client
}

func TestWaitForPackageInstalled_succeedsFromPackageInfo(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"item":{"assets":{},"name":"system","title":"System","version":"1.0.0","status":"installed"}}`)
	}))
	defer srv.Close()

	err := WaitForPackageInstalled(t.Context(), newTestPackagesClient(t, srv), "system", "1.0.0", "", 10*time.Second, false)
	require.NoError(t, err)
}

func TestWaitForPackageInstalled_installFailedReturnsError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"item":{"assets":{},"name":"system","title":"System","version":"1.0.0","status":"install_failed"}}`)
	}))
	defer srv.Close()

	err := WaitForPackageInstalled(t.Context(), newTestPackagesClient(t, srv), "system", "1.0.0", "", 10*time.Second, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "installation failed")
}

func TestWaitForPackageInstalled_fallbackToListScan(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/fleet/epm/packages" {
			fmt.Fprint(w, `{"items":[{"id":"system","name":"system","title":"System","version":"1.0.0","status":"installed"}]}`)
			return
		}
		// The package-info endpoint has not caught up yet.
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	err := WaitForPackageInstalled(t.Context(), newTestPackagesClient(t, srv), "system", "1.0.0", "", 10*time.Second, true)
	require.NoError(t, err)
}

func TestWaitForPackageInstalled_noFallbackTimesOut(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	err := WaitForPackageInstalled(t.Context(), newTestPackagesClient(t, srv), "system", "1.0.0", "", 2500*time.Millisecond, false)
	require.Error(t, err)
}
