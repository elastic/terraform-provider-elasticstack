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
	"testing"

	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helper: minimal KibanaScopedClient constructors ---

// newKibanaScopedClientWithEndpoint returns a KibanaScopedClient that has a
// non-empty kibanaEndpoint and populated client objects, but empty auth fields.
// This lets us verify that endpoint-only validation does not check auth.
func newKibanaScopedClientWithEndpointNoAuth(t *testing.T, endpoint string) *KibanaScopedClient {
	t.Helper()

	kibOapi, err := kibanaoapi.NewClient(kibanaoapi.Config{
		URL: endpoint,
		// No username/password — intentionally empty.
	})
	require.NoError(t, err)

	fleetCfg := fleetclient.Config{
		URL: endpoint,
	}
	fleet, err := fleetclient.NewClient(fleetCfg)
	require.NoError(t, err)

	return &KibanaScopedClient{
		kibanaOapi:     kibOapi,
		fleet:          fleet,
		kibanaEndpoint: endpoint,
		fleetEndpoint:  endpoint,
	}
}

// newKibanaScopedClientFleetFromKibana returns a KibanaScopedClient that
// simulates the Fleet-from-Kibana resolution path: kibanaEndpoint is set, and
// fleetEndpoint is also set (derived from Kibana) but there is no explicit
// standalone Fleet endpoint. The fleet client is populated.
func newKibanaScopedClientFleetFromKibana(t *testing.T, kibanaURL string) *KibanaScopedClient {
	t.Helper()

	kibOapi, err := kibanaoapi.NewClient(kibanaoapi.Config{
		URL:      kibanaURL,
		Username: "elastic",
		Password: "changeme",
	})
	require.NoError(t, err)

	// fleetEndpoint is derived from Kibana, set to the same URL.
	fleetCfg := fleetclient.Config{
		URL:      kibanaURL,
		Username: "elastic",
		Password: "changeme",
	}
	fleet, err := fleetclient.NewClient(fleetCfg)
	require.NoError(t, err)

	return &KibanaScopedClient{
		kibanaOapi:     kibOapi,
		fleet:          fleet,
		kibanaEndpoint: kibanaURL,
		// fleetEndpoint derived from Kibana.
		fleetEndpoint: kibanaURL,
	}
}

// newKibanaScopedClientWithStatusServer returns a KibanaScopedClient backed by an
// httptest.Server that responds to GET /api/status with the given version and flavor.
func newKibanaScopedClientWithStatusServer(t *testing.T, version, flavor string) *KibanaScopedClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == kibanaStatusPath {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"version":{"number":%q,"build_flavor":%q}}`, version, flavor)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return newKibanaScopedClientWithEndpointNoAuth(t, srv.URL)
}

// newKibanaScopedClientWithStatusHandler returns a KibanaScopedClient backed by an
// httptest.Server that uses the given handler for all requests.
func newKibanaScopedClientWithStatusHandler(t *testing.T, handler http.HandlerFunc) *KibanaScopedClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return newKibanaScopedClientWithEndpointNoAuth(t, srv.URL)
}

// --- Fleet endpoint inherited from Kibana ---

func TestKibanaScopedClient_GetFleetClient_InheritedFromKibana(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientFleetFromKibana(t, "http://kibana.example.com:5601")
	require.NotNil(t, sc.GetFleetClient())
}

func TestKibanaScopedClient_GetKibanaOapiClient_EndpointPresentNoAuth(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithEndpointNoAuth(t, "http://kibana.example.com:5601")
	require.NotNil(t, sc.GetKibanaOapiClient())
}

func TestKibanaScopedClient_GetFleetClient_EndpointPresentNoAuth(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithEndpointNoAuth(t, "http://kibana.example.com:5601")
	require.NotNil(t, sc.GetFleetClient())
}

// --- EnforceMinVersion: stateful version comparisons ---

func TestKibanaScopedClient_EnforceMinVersion_StatefulBelowMin(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "8.10.0", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
	require.False(t, diags.HasError())
	assert.False(t, ok)
}

func TestKibanaScopedClient_EnforceMinVersion_StatefulAtMin(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "8.15.0", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestKibanaScopedClient_EnforceMinVersion_StatefulAboveMin(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "9.0.0", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestKibanaScopedClient_EnforceMinVersion_SnapshotBuildMeetsReleaseMinimum(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "9.5.0-SNAPSHOT", "default")
	minVer, err := version.NewVersion("9.5.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
	require.False(t, diags.HasError())
	assert.True(t, ok, "9.5.0-SNAPSHOT must satisfy a 9.5.0 release minimum")
}

func TestKibanaScopedClient_EnforceMinVersion_ServerlessShortCircuit(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "8.10.0", ServerlessFlavor)
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
	require.False(t, diags.HasError())
	assert.True(t, ok, "serverless must short-circuit to true regardless of version")
}

func TestKibanaScopedClient_EnforceMinVersion_MalformedVersionResponse(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "not-a-version", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

func TestKibanaScopedClient_EnforceMinVersion_StatusAPIError(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == kibanaStatusPath {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

func TestKibanaScopedClient_EnforceMinVersion_InvalidJSONResponse(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == kibanaStatusPath {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{invalid json`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

// --- EnforceVersionCheck: stateful version comparisons ---

func TestKibanaScopedClient_EnforceVersionCheck_StatefulBelowMin(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "8.10.0", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceVersionCheck(t.Context(), func(v *version.Version) bool {
		return v.GreaterThanOrEqual(minVer)
	})
	require.False(t, diags.HasError())
	assert.False(t, ok)
}

func TestKibanaScopedClient_EnforceVersionCheck_StatefulAtMin(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "8.15.0", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceVersionCheck(t.Context(), func(v *version.Version) bool {
		return v.GreaterThanOrEqual(minVer)
	})
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestKibanaScopedClient_EnforceVersionCheck_StatefulAboveMin(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "9.0.0", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceVersionCheck(t.Context(), func(v *version.Version) bool {
		return v.GreaterThanOrEqual(minVer)
	})
	require.False(t, diags.HasError())
	assert.True(t, ok)
}

func TestKibanaScopedClient_EnforceVersionCheck_ServerlessShortCircuit(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "8.10.0", ServerlessFlavor)

	ok, diags := sc.EnforceVersionCheck(t.Context(), func(_ *version.Version) bool { return false })
	require.False(t, diags.HasError())
	assert.True(t, ok, "serverless must short-circuit to true even when check returns false")
}

func TestKibanaScopedClient_EnforceVersionCheck_MalformedVersionResponse(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusServer(t, "not-a-version", "default")
	minVer, err := version.NewVersion("8.15.0")
	require.NoError(t, err)

	ok, diags := sc.EnforceVersionCheck(t.Context(), func(v *version.Version) bool {
		return v.GreaterThanOrEqual(minVer)
	})
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

func TestKibanaScopedClient_EnforceVersionCheck_StatusAPIError(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == kibanaStatusPath {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	ok, diags := sc.EnforceVersionCheck(t.Context(), func(_ *version.Version) bool { return true })
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

func TestKibanaScopedClient_EnforceVersionCheck_InvalidJSONResponse(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithStatusHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == kibanaStatusPath {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{invalid json`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	ok, diags := sc.EnforceVersionCheck(t.Context(), func(_ *version.Version) bool { return true })
	assert.False(t, ok)
	require.True(t, diags.HasError())
}
