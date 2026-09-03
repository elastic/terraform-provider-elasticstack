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

package managedintegration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// registerLegacyPackagePoliciesGuard counts requests to the deprecated
// /api/fleet/package_policies/* surface. Successful CRUD callback tests assert
// the counter remains zero (task 8 removed the package_policies fallback).
func registerLegacyPackagePoliciesGuard(mux *http.ServeMux) *atomic.Int64 {
	var calls atomic.Int64
	record := func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "unexpected legacy package_policies call", http.StatusTeapot)
	}
	mux.HandleFunc("/api/fleet/package_policies", record)
	mux.HandleFunc("/api/fleet/package_policies/", record)
	return &calls
}

func requireNoLegacyPackagePoliciesCalls(t *testing.T, calls *atomic.Int64) {
	t.Helper()
	require.Equal(t, int64(0), calls.Load(), "legacy /api/fleet/package_policies/ must not be called")
}

// httpMethodCapture records the HTTP method of the last request via atomic.Value
// so callback tests can assert verbs without data races if handlers or
// assertions ever run concurrently.
type httpMethodCapture struct {
	method atomic.Value
}

func newHTTPMethodCapture() *httpMethodCapture {
	return &httpMethodCapture{}
}

func (c *httpMethodCapture) record(r *http.Request) {
	c.method.Store(r.Method)
}

func (c *httpMethodCapture) requireEqual(t *testing.T, want string) {
	t.Helper()
	got, _ := c.method.Load().(string)
	require.Equal(t, want, got)
}

// wrapLegacyPackagePoliciesGuard wraps a handler and rejects any
// /api/fleet/package_policies/* request while counting it.
func wrapLegacyPackagePoliciesGuard(next http.Handler) (http.Handler, *atomic.Int64) {
	var calls atomic.Int64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/fleet/package_policies") {
			calls.Add(1)
			http.Error(w, "unexpected legacy package_policies call", http.StatusTeapot)
			return
		}
		next.ServeHTTP(w, r)
	}), &calls
}

func TestRegisterLegacyPackagePoliciesGuard_countsLegacyCalls(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	legacyCalls := registerLegacyPackagePoliciesGuard(mux)
	mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/fleet/package_policies/pp-1")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusTeapot, resp.StatusCode)
	require.Equal(t, int64(1), legacyCalls.Load())

	resp, err = http.Get(srv.URL + "/api/fleet/managed_integrations/mi-1")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, int64(1), legacyCalls.Load(), "managed_integrations traffic must not increment the legacy counter")
}

func TestWrapLegacyPackagePoliciesGuard_countsLegacyCalls(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	handler, legacyCalls := wrapLegacyPackagePoliciesGuard(next)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/fleet/package_policies/x")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusTeapot, resp.StatusCode)
	require.Equal(t, int64(1), legacyCalls.Load())
	require.Equal(t, int64(0), calls.Load())

	resp, err = http.Get(srv.URL + "/ok")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, int64(1), legacyCalls.Load())
	require.Equal(t, int64(1), calls.Load())
}
