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

package integrationpolicy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFleetClient(t *testing.T, handler http.Handler) *fleet.Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	api, err := kbapi.NewClientWithResponses(server.URL+"/", kbapi.WithHTTPClient(server.Client()))
	require.NoError(t, err)

	return &fleet.Client{
		URL:  server.URL,
		HTTP: server.Client(),
		API:  api,
	}
}

func TestGetPackageInfo_PackageNotFound(t *testing.T) {
	knownPackages.Delete(getPackageCacheKey("tcp", "3.1.10"))

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	pkg, diags := getPackageInfo(context.Background(), client, "tcp", "3.1.10")

	assert.Nil(t, pkg)
	require.False(t, diags.HasError(), "expected no errors, got: %v", diags.Errors())
	require.Len(t, diags, 1)
	assert.Contains(t, diags[0].Summary(), "Package not found")
	assert.Contains(t, diags[0].Detail(), "tcp")
	assert.Contains(t, diags[0].Detail(), "3.1.10")
}

func TestGetPackageInfo_Success(t *testing.T) {
	knownPackages.Delete(getPackageCacheKey("tcp", "3.1.11"))

	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := struct {
			Item kbapi.PackageInfo `json:"item"`
		}{
			Item: kbapi.PackageInfo{
				Name:    "tcp",
				Version: "3.1.11",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))

	pkg, diags := getPackageInfo(context.Background(), client, "tcp", "3.1.11")

	require.False(t, diags.HasError())
	require.NotNil(t, pkg)
	assert.Equal(t, "tcp", pkg.Name)
	assert.Equal(t, "3.1.11", pkg.Version)
}

func TestGetPackageInfo_CacheHit(t *testing.T) {
	cached := kbapi.PackageInfo{Name: "tcp", Version: "3.1.11"}
	knownPackages.Store(getPackageCacheKey("tcp", "3.1.11"), cached)
	t.Cleanup(func() { knownPackages.Delete(getPackageCacheKey("tcp", "3.1.11")) })

	client := newTestFleetClient(t, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("HTTP request should not be made when cache is hit")
	}))

	pkg, diags := getPackageInfo(context.Background(), client, "tcp", "3.1.11")

	require.False(t, diags.HasError())
	require.NotNil(t, pkg)
	assert.Equal(t, "tcp", pkg.Name)
}
