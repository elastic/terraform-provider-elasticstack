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

package clients

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newKibanaScopedClientWithStatusCounter returns a KibanaScopedClient backed by
// an httptest.Server that responds to GET /api/status with the given version
// and flavor, together with an atomic counter the caller can read to assert how
// many times the status endpoint was hit.
func newKibanaScopedClientWithStatusCounter(t *testing.T, version, flavor string) (*KibanaScopedClient, *atomic.Int64) {
	t.Helper()
	var count atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == kibanaStatusPath {
			count.Add(1)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"version":{"number":%q,"build_flavor":%q}}`, version, flavor)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return newKibanaScopedClientWithEndpointNoAuth(t, srv.URL), &count
}

// TestKibanaScopedClient_getServerStatusRaw_CachesSuccessfulResult verifies
// that multiple EnforceMinVersion and EnforceVersionCheck calls on a single
// KibanaScopedClient instance share one /api/status round trip.
func TestKibanaScopedClient_getServerStatusRaw_CachesSuccessfulResult(t *testing.T) {
	t.Parallel()
	sc, callCount := newKibanaScopedClientWithStatusCounter(t, "8.15.0", "default")
	minVer, err := version.NewVersion("8.0.0")
	require.NoError(t, err)

	for range 5 {
		ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
		require.False(t, diags.HasError())
		assert.True(t, ok)
	}

	// EnforceVersionCheck must share the same cache as EnforceMinVersion.
	for range 5 {
		ok, diags := sc.EnforceVersionCheck(t.Context(), func(v *version.Version) bool {
			return v.GreaterThanOrEqual(minVer)
		})
		require.False(t, diags.HasError())
		assert.True(t, ok)
	}

	assert.EqualValues(t, 1, callCount.Load(),
		"all EnforceMinVersion + EnforceVersionCheck calls must share one /api/status fetch")
}

// TestKibanaScopedClient_getServerStatusRaw_DoesNotCacheErrors verifies that a
// transient failure on the first call is not cached: a subsequent call hits
// the backend again. Once a call succeeds, the result is cached.
func TestKibanaScopedClient_getServerStatusRaw_DoesNotCacheErrors(t *testing.T) {
	t.Parallel()
	var (
		callCount atomic.Int64
		failNext  atomic.Bool
	)
	failNext.Store(true)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == kibanaStatusPath {
			callCount.Add(1)
			if failNext.Swap(false) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"version":{"number":"8.15.0","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	sc := newKibanaScopedClientWithEndpointNoAuth(t, srv.URL)
	minVer, err := version.NewVersion("8.0.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
	assert.False(t, ok)
	require.True(t, diags.HasError(), "first call must surface the 500")
	require.EqualValues(t, 1, callCount.Load())

	ok, diags = sc.EnforceMinVersion(t.Context(), minVer)
	require.False(t, diags.HasError(), "second call must re-issue the status request (no error caching)")
	assert.True(t, ok)
	require.EqualValues(t, 2, callCount.Load())

	// Subsequent calls hit the cache populated by the successful second call.
	ok, diags = sc.EnforceMinVersion(t.Context(), minVer)
	require.False(t, diags.HasError())
	assert.True(t, ok)
	assert.EqualValues(t, 2, callCount.Load(),
		"third call must be served from the cache populated by the successful second call")
}

// TestKibanaScopedClient_getServerStatusRaw_MissingEndpointNotCached verifies
// that a missing-endpoint failure (no HTTP request issued at all) does not
// populate the cache.
func TestKibanaScopedClient_getServerStatusRaw_MissingEndpointNotCached(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientNoEndpoint(t)
	minVer, err := version.NewVersion("8.0.0")
	require.NoError(t, err)

	_, diags := sc.EnforceMinVersion(t.Context(), minVer)
	require.True(t, diags.HasError())

	// statusCached must remain false so that a later call (e.g. after the
	// endpoint has been populated by something out-of-band) still attempts the
	// fetch instead of returning the empty cached values.
	sc.statusMu.Lock()
	assert.False(t, sc.statusCached, "missing-endpoint failures must not populate the cache")
	sc.statusMu.Unlock()
}

// TestKibanaScopedClient_getServerStatusRaw_ConcurrentCallsShareOneFetch
// verifies that N goroutines calling EnforceMinVersion concurrently still
// share a single /api/status fetch.
func TestKibanaScopedClient_getServerStatusRaw_ConcurrentCallsShareOneFetch(t *testing.T) {
	t.Parallel()
	sc, callCount := newKibanaScopedClientWithStatusCounter(t, "8.15.0", "default")
	minVer, err := version.NewVersion("8.0.0")
	require.NoError(t, err)

	const goroutines = 32
	var wg sync.WaitGroup
	wg.Add(goroutines)
	start := make(chan struct{})

	for range goroutines {
		go func() {
			defer wg.Done()
			<-start
			ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
			// Use assert (not require) inside spawned goroutines so a failure
			// records on the test but does not call t.FailNow from outside the
			// test goroutine.
			assert.False(t, diags.HasError())
			assert.True(t, ok)
		}()
	}
	close(start)
	wg.Wait()

	assert.EqualValues(t, 1, callCount.Load(),
		"concurrent EnforceMinVersion calls must share one /api/status fetch")
}
