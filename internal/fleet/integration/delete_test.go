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
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPackageName          = "system"
	testPackageVersion       = "1.0.0"
	testSpaceID              = "default"
	testPackageUninstallPath = "/api/fleet/epm/packages/system/1.0.0"
	installSpace400Body      = `{"statusCode":400,"error":"Bad Request","message":"Impossible to delete kibana assets from the space where the package was installed, you must uninstall the package."}`
	other400Body             = `{"statusCode":400,"error":"Bad Request","message":"Some other validation error"}`
)

func newTestFleetClient(t *testing.T, server *httptest.Server) *fleet.Client {
	t.Helper()
	client, err := fleet.NewClient(fleet.Config{URL: server.URL})
	require.NoError(t, err)
	return client
}

type testMinVersionClient struct {
	supported           bool
	calls               int
	requestedMinVersion *version.Version
}

func (c *testMinVersionClient) EnforceMinVersion(_ context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	c.calls++
	c.requestedMinVersion = minVersion
	return c.supported, nil
}

func assertDiagnosticsDoNotContainInstallSpaceRejection(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	assert.False(t, fleet.ContainsInstallSpaceDeleteRejection(diags), "diagnostics must not contain the install-space rejection message")
}

// newFallbackServer returns an httptest.Server that handles DELETE /kibana_assets
// with the provided status/body and optionally delegates DELETE uninstall requests
// to uninstallHandler.
func newFallbackServer(t *testing.T, kibanaAssetsStatus int, kibanaAssetsBody string, uninstallHandler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/kibana_assets"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(kibanaAssetsStatus)
			fmt.Fprint(w, kibanaAssetsBody)
		case r.Method == http.MethodDelete && r.URL.Path == testPackageUninstallPath:
			if uninstallHandler != nil {
				uninstallHandler(w, r)
				return
			}
			t.Errorf("unexpected uninstall request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestDeleteKibanaAssetsWithFallback_installSpace400FallsBackToUninstall(t *testing.T) {
	t.Parallel()

	var uninstallCalls int
	var uninstallForceParam string

	srv := newFallbackServer(t, http.StatusBadRequest, installSpace400Body, func(w http.ResponseWriter, r *http.Request) {
		uninstallCalls++
		uninstallForceParam = r.URL.Query().Get("force")
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	client := newTestFleetClient(t, srv)
	diags := deleteKibanaAssetsWithFallback(context.Background(), client, testPackageName, testPackageVersion, testSpaceID, true)

	assert.False(t, diags.HasError())
	assert.Equal(t, 1, uninstallCalls)
	assert.Equal(t, "true", uninstallForceParam)
	assertDiagnosticsDoNotContainInstallSpaceRejection(t, diags)
}

func TestDeleteKibanaAssetsWithFallback_installSpace400WithoutForceReturnsActionableError(t *testing.T) {
	t.Parallel()

	var uninstallCalls int

	srv := newFallbackServer(t, http.StatusBadRequest, installSpace400Body, func(w http.ResponseWriter, _ *http.Request) {
		uninstallCalls++
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	client := newTestFleetClient(t, srv)
	diags := deleteKibanaAssetsWithFallback(context.Background(), client, testPackageName, testPackageVersion, testSpaceID, false)

	require.True(t, diags.HasError())
	assert.Equal(t, 0, uninstallCalls, "Uninstall must not be called without force, since it would affect other spaces")
	require.Len(t, diags.Errors(), 1)
	errDetail := diags.Errors()[0].Detail()
	assert.Contains(t, errDetail, "install space")
	assert.Contains(t, errDetail, "force")
	assert.NotContains(t, errDetail, `"statusCode":400`, "the raw Fleet 400 body must not be leaked verbatim")
}

func TestDeleteKibanaAssetsWithFallback_installSpace400UninstallFailureReturnsUninstallError(t *testing.T) {
	t.Parallel()

	var uninstallCalls int

	srv := newFallbackServer(t, http.StatusBadRequest, installSpace400Body, func(w http.ResponseWriter, _ *http.Request) {
		uninstallCalls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"statusCode":500,"error":"Internal Server Error","message":"uninstall failed"}`)
	})
	defer srv.Close()

	client := newTestFleetClient(t, srv)
	diags := deleteKibanaAssetsWithFallback(context.Background(), client, testPackageName, testPackageVersion, testSpaceID, true)

	require.True(t, diags.HasError())
	assert.Equal(t, 1, uninstallCalls)
	assert.Contains(t, diags.Errors()[0].Detail(), "uninstall failed")
	assertDiagnosticsDoNotContainInstallSpaceRejection(t, diags)
}

func TestDeleteKibanaAssetsWithFallback_other400DoesNotFallback(t *testing.T) {
	t.Parallel()

	var uninstallCalls int

	srv := newFallbackServer(t, http.StatusBadRequest, other400Body, func(w http.ResponseWriter, _ *http.Request) {
		uninstallCalls++
		w.WriteHeader(http.StatusOK)
	})
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

	srv := newFallbackServer(t, http.StatusOK, "{}", func(w http.ResponseWriter, _ *http.Request) {
		uninstallCalls++
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	client := newTestFleetClient(t, srv)
	diags := deleteKibanaAssetsWithFallback(context.Background(), client, testPackageName, testPackageVersion, testSpaceID, true)

	assert.False(t, diags.HasError())
	assert.Equal(t, 0, uninstallCalls)
}

func TestDeleteIntegration_defaultSpacePreservesMultiSpaceGuard(t *testing.T) {
	t.Parallel()

	const deleteAssetsPath = testPackageUninstallPath + "/kibana_assets"
	const packageResponse = `{
		"item": {
			"assets": {},
			"name": "system",
			"title": "System",
			"version": "1.0.0",
			"status": "installed",
			"installationInfo": {
				"additional_spaces_installed_kibana": {"other-space": []},
				"install_source": "registry",
				"install_status": "installed",
				"installed_es": [],
				"installed_kibana": [],
				"installed_kibana_space_id": "default",
				"name": "system",
				"type": "integration",
				"verification_status": "verified",
				"version": "1.0.0"
			}
		}
	}`

	var getCalls, deleteAssetsCalls, uninstallCalls int
	var getPath, deleteAssetsPathSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == testPackageUninstallPath:
			getCalls++
			getPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, packageResponse)
		case r.Method == http.MethodDelete && r.URL.Path == deleteAssetsPath:
			deleteAssetsCalls++
			deleteAssetsPathSeen = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, installSpace400Body)
		case r.Method == http.MethodDelete && r.URL.Path == testPackageUninstallPath:
			uninstallCalls++
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	versionClient := &testMinVersionClient{supported: true}
	model := integrationModel{
		Name:        types.StringValue(testPackageName),
		Version:     types.StringValue(testPackageVersion),
		SpaceID:     types.StringValue(testSpaceID),
		Force:       types.BoolValue(false),
		SkipDestroy: types.BoolValue(false),
	}

	diags := deleteIntegrationWithClients(t.Context(), versionClient, newTestFleetClient(t, srv), model)

	require.True(t, diags.HasError())
	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, 1, versionClient.calls)
	require.NotNil(t, versionClient.requestedMinVersion)
	assert.True(t, versionClient.requestedMinVersion.Equal(MinVersionSpaceAwareIntegration))
	assert.Equal(t, 1, getCalls)
	assert.Equal(t, testPackageUninstallPath, getPath)
	assert.Equal(t, 1, deleteAssetsCalls)
	assert.Equal(t, deleteAssetsPath, deleteAssetsPathSeen)
	assert.Equal(t, 0, uninstallCalls, "default-space deletion must not globally uninstall a package used in other spaces")
	assert.Contains(t, diags.Errors()[0].Detail(), "install space")
	assert.Contains(t, diags.Errors()[0].Detail(), "force")
	assert.NotContains(t, diags.Errors()[0].Detail(), `"statusCode":400`)
	assertDiagnosticsDoNotContainInstallSpaceRejection(t, diags)
}
