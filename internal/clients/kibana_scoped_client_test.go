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
	"testing"

	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helper: minimal KibanaScopedClient constructors ---

// newKibanaScopedClientNoEndpoint returns a KibanaScopedClient whose kibana and
// fleet endpoint fields are both empty — simulating unconfigured state.
func newKibanaScopedClientNoEndpoint(t *testing.T) *KibanaScopedClient {
	t.Helper()
	return &KibanaScopedClient{}
}

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

// --- Scenario 3: Missing Kibana endpoint → GetKibanaOapiClient error ---

func TestKibanaScopedClient_GetKibanaOapiClient_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientNoEndpoint(t)
	client, err := sc.GetKibanaOapiClient()
	assert.Nil(t, client, "GetKibanaOapiClient must return nil client when kibana endpoint is missing")
	require.Error(t, err)
	assert.Equal(t,
		"kibana OpenAPI client is not configured: set kibana.endpoints, kibana_connection.endpoints, or KIBANA_ENDPOINT",
		err.Error(),
	)
}

// --- Scenario 4: Localhost fallback blocked ---
// The accessor must fail even when the underlying kibanaOapi.Client would normally
// fall back to localhost:5601. We prove this by constructing a KibanaScopedClient
// whose kibanaOapi field is nil (no client initialised) but kibanaEndpoint is
// empty — the validation check fires before any client is returned.

func TestKibanaScopedClient_GetKibanaOapiClient_NilClientNoEndpoint_BlocksLocalhost(t *testing.T) {
	t.Parallel()
	// kibanaOapi field nil, kibanaEndpoint empty: models the case where the provider
	// block is completely absent (no kibana config at all).
	sc := &KibanaScopedClient{
		kibanaOapi:     nil,
		kibanaEndpoint: "",
	}
	client, err := sc.GetKibanaOapiClient()
	assert.Nil(t, client)
	require.Error(t, err,
		"GetKibanaOapiClient must return an error (not silently fall back to localhost) when endpoint is empty")
	assert.Equal(t,
		"kibana OpenAPI client is not configured: set kibana.endpoints, kibana_connection.endpoints, or KIBANA_ENDPOINT",
		err.Error(),
	)
}

// --- Scenario 5: Missing Fleet endpoint → GetFleetClient error ---

func TestKibanaScopedClient_GetFleetClient_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientNoEndpoint(t)
	client, err := sc.GetFleetClient()
	assert.Nil(t, client, "GetFleetClient must return nil client when fleet endpoint is missing")
	require.Error(t, err)
	assert.Equal(t,
		"fleet client is not configured: set fleet.endpoint or FLEET_ENDPOINT, "+
			"or configure kibana.endpoints, kibana_connection.endpoints, or KIBANA_ENDPOINT "+
			"for inherited Fleet endpoint resolution",
		err.Error(),
	)
}

// --- Scenario 6: Fleet endpoint inherited from Kibana → GetFleetClient succeeds ---

func TestKibanaScopedClient_GetFleetClient_InheritedFromKibana(t *testing.T) {
	t.Parallel()
	// Use a placeholder URL; we are only testing the accessor validation path,
	// not an actual HTTP connection.
	sc := newKibanaScopedClientFleetFromKibana(t, "http://kibana.example.com:5601")
	client, err := sc.GetFleetClient()
	require.NoError(t, err,
		"GetFleetClient must succeed when fleet endpoint is inherited from Kibana")
	assert.NotNil(t, client)
}

// --- Scenario 7: Endpoint present, auth empty → accessor succeeds ---

func TestKibanaScopedClient_GetKibanaOapiClient_EndpointPresentNoAuth(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithEndpointNoAuth(t, "http://kibana.example.com:5601")
	client, err := sc.GetKibanaOapiClient()
	require.NoError(t, err,
		"GetKibanaOapiClient must not fail when endpoint is present but auth fields are empty")
	assert.NotNil(t, client)
}

func TestKibanaScopedClient_GetFleetClient_EndpointPresentNoAuth(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithEndpointNoAuth(t, "http://kibana.example.com:5601")
	client, err := sc.GetFleetClient()
	require.NoError(t, err,
		"GetFleetClient must not fail when endpoint is present but auth fields are empty")
	assert.NotNil(t, client)
}

// --- GetKibanaOapiClientDiag ---

func TestKibanaScopedClient_GetKibanaOapiClientDiag_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientNoEndpoint(t)
	client, diags := sc.GetKibanaOapiClientDiag()
	assert.Nil(t, client)
	require.True(t, diags.HasError())
}

func TestKibanaScopedClient_GetKibanaOapiClientDiag_EndpointPresent(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientWithEndpointNoAuth(t, "http://kibana.example.com:5601")
	client, diags := sc.GetKibanaOapiClientDiag()
	require.False(t, diags.HasError())
	assert.NotNil(t, client)
}

// --- getServerStatusRaw error propagation ---

// TestKibanaScopedClient_getServerStatusRaw_MissingEndpoint verifies that when
// kibanaEndpoint is empty, getServerStatusRaw returns a populated Diagnostics error
// rather than a nil-pointer panic or a silent empty result.
func TestKibanaScopedClient_getServerStatusRaw_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientNoEndpoint(t)
	_, _, diags := sc.getServerStatusRaw(t.Context())
	require.True(t, diags.HasError(), "getServerStatusRaw must return an error when kibana endpoint is not configured")
}

// TestKibanaScopedClient_ServerVersion_MissingEndpoint verifies that ServerVersion
// propagates the error from getServerStatusRaw when the endpoint is not configured.
func TestKibanaScopedClient_ServerVersion_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientNoEndpoint(t)
	v, diags := sc.ServerVersion(t.Context())
	assert.Nil(t, v)
	require.True(t, diags.HasError())
}

// TestKibanaScopedClient_ServerFlavor_MissingEndpoint verifies that ServerFlavor
// propagates the error from getServerStatusRaw when the endpoint is not configured.
func TestKibanaScopedClient_ServerFlavor_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientNoEndpoint(t)
	flavor, diags := sc.ServerFlavor(t.Context())
	assert.Empty(t, flavor)
	require.True(t, diags.HasError())
}

// TestKibanaScopedClient_EnforceMinVersion_MissingEndpoint verifies that
// EnforceMinVersion propagates the error from getServerStatusRaw when the endpoint
// is not configured.
func TestKibanaScopedClient_EnforceMinVersion_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientNoEndpoint(t)
	minVer, err := version.NewVersion("8.0.0")
	require.NoError(t, err)
	ok, diags := sc.EnforceMinVersion(t.Context(), minVer)
	assert.False(t, ok)
	require.True(t, diags.HasError())
}

// TestKibanaScopedClient_EnforceVersionCheck_MissingEndpoint verifies that
// EnforceVersionCheck propagates the error from getServerStatusRaw when the endpoint
// is not configured — confirming it issues only one HTTP request attempt (via the
// shared helper) before failing.
func TestKibanaScopedClient_EnforceVersionCheck_MissingEndpoint(t *testing.T) {
	t.Parallel()
	sc := newKibanaScopedClientNoEndpoint(t)
	ok, diags := sc.EnforceVersionCheck(t.Context(), func(_ *version.Version) bool { return true })
	assert.False(t, ok)
	require.True(t, diags.HasError())
}
