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

package agentlesspolicy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// clearKibanaEnvOverrides prevents the config package's environment-variable
// override machinery (internal/clients/config: KIBANA_ENDPOINT,
// KIBANA_USERNAME, etc. -- see withNonURLEnvironmentOverrides /
// withURLEnvironmentOverride) from hijacking the httptest.Server endpoint
// these tests configure explicitly. Without this, a developer shell that has
// sourced .env or .env.cloud (as this OpenSpec change's Task 6 empirical
// verification required) would silently redirect these "hermetic" tests at a
// real Kibana instead of the local httptest server -- which is exactly what
// was observed while developing this test. Setting
// TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT makes the explicitly
// configured endpoint win over KIBANA_ENDPOINT; the auth vars are blanked
// because applyAuthEnvOverrides applies them unconditionally whenever they
// are merely *present* in the environment, regardless of value.
func clearKibanaEnvOverrides(t *testing.T) {
	t.Helper()
	t.Setenv("TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT", "true")
	// Deliberately NOT included in the blank-to-"" loop below: KIBANA_CA_CERTS /
	// FLEET_CA_CERTS. Unlike the auth vars below (where an empty string is
	// simply "no credential" and harmless against a plain-HTTP httptest
	// server), an *empty but present* CA_CERTS var is treated as a
	// one-element list containing a single empty file path, which
	// config.withNonURLEnvironmentOverrides then tries to open -- turning
	// "unset" into a hard error. Leaving these unset (LookupEnv ok=false) is
	// the correct no-op.
	//
	// FLEET_ENDPOINT gets the same "leave genuinely unset, don't blank"
	// treatment, but for a different and sharper reason: unlike Kibana,
	// there is no TF_ELASTICSTACK_PREFER_CONFIGURED_FLEET_ENDPOINT escape
	// valve, and fleetConfig.withEnvironmentOverrides
	// (internal/clients/config/fleet.go) applies FLEET_ENDPOINT via
	// `os.LookupEnv` -- i.e. on env-var *presence*, not on a non-empty
	// value. Blanking it to "" (as this loop used to do) therefore
	// unconditionally clobbers the Fleet client's URL -- inherited from the
	// scoped Kibana endpoint when no Fleet block is configured -- with an
	// empty string, breaking every request the Fleet client makes
	// ("unsupported protocol scheme \"\""). This went unnoticed as long as
	// every test in this file only exercised the Kibana OpenAPI client
	// (checkDeploymentTopology's `/api/status` probe); it surfaced once
	// create_test.go's TestCreateAgentlessPolicy_topologyGatesFleetCall
	// started reusing this helper to also exercise a real Fleet POST. If
	// FLEET_ENDPOINT happens to already be set in the ambient environment
	// (it is not set by this repo's .env), it is force-unset for the
	// duration of the test and restored afterwards, so these tests stay
	// hermetic either way.
	if orig, ok := os.LookupEnv("FLEET_ENDPOINT"); ok {
		require.NoError(t, os.Unsetenv("FLEET_ENDPOINT"))
		t.Cleanup(func() {
			// t.Setenv cannot express "restore to fully unset" (it can only
			// set a value), so this deliberately restores via the raw os
			// package rather than t.Setenv.
			require.NoError(t, os.Setenv("FLEET_ENDPOINT", orig)) //nolint:usetesting
		})
	}
	for _, key := range []string{
		"KIBANA_USERNAME", "KIBANA_PASSWORD", "KIBANA_API_KEY", "KIBANA_BEARER_TOKEN",
		"FLEET_USERNAME", "FLEET_PASSWORD", "FLEET_API_KEY", "FLEET_BEARER_TOKEN",
	} {
		t.Setenv(key, "")
	}
}

// newTopologyTestClient builds a real *clients.KibanaScopedClient (via the
// same public ProviderClientFactory path the provider's own Configure method
// uses) backed by an httptest.Server, so checkDeploymentTopology exercises
// its real HTTP call path rather than a hand-rolled fake. Most call sites
// only need the client; see newTopologyTestClientWithServer for the one
// subtest that also needs to control the server's lifecycle (closing it
// early to simulate an unreachable Kibana).
func newTopologyTestClient(t *testing.T, handler http.Handler) *clients.KibanaScopedClient {
	t.Helper()
	client, _ := newTopologyTestClientWithServer(t, handler)
	return client
}

// newTopologyTestClientWithServer is newTopologyTestClient, additionally
// returning the underlying *httptest.Server so a caller can control its
// lifecycle directly (e.g. closing it early to simulate an unreachable
// Kibana) instead of relying on the t.Cleanup-deferred close.
func newTopologyTestClientWithServer(t *testing.T, handler http.Handler) (*clients.KibanaScopedClient, *httptest.Server) {
	t.Helper()
	clearKibanaEnvOverrides(t)

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := config.ProviderConfiguration{
		Kibana: []config.KibanaConnection{
			{
				Endpoints: types.ListValueMust(types.StringType, []attr.Value{types.StringValue(srv.URL)}),
				CACerts:   types.ListNull(types.StringType),
			},
		},
	}
	factory, diags := clients.NewProviderClientFactoryFromFramework(context.Background(), cfg, "test")
	require.False(t, diags.HasError(), "factory build failed: %v", diags)

	scoped, diags := factory.GetKibanaClient(context.Background(), types.ListNull(types.ObjectType{}))
	require.False(t, diags.HasError(), "scoped client build failed: %v", diags)

	return scoped, srv
}

func statusHandler(body string, extraHeaders map[string]string, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		for k, v := range extraHeaders {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		if statusCode != 0 {
			w.WriteHeader(statusCode)
		}
		fmt.Fprint(w, body)
	}
}

// TestCheckDeploymentTopology covers Task 6.2's fail-closed/fail-open matrix
// (see specs/fleet-agentless-policy/spec.md, "Deployment topology preflight
// check", and design.md Decision 7). Two of these cases -- "confirmed cloud
// hosted" and "confirmed self-managed" -- were additionally verified
// empirically against a live Kibana 9.4.3 Elastic Cloud Hosted deployment
// and a live self-managed docker-compose Kibana 9.4.0 (see the Task 6 report
// for the fleet-agentless-policy OpenSpec change); this test hermetically
// pins the same decision logic so it cannot regress silently.
func TestCheckDeploymentTopology(t *testing.T) {
	t.Run("serverless build_flavor passes (confirmed cloud)", func(t *testing.T) {
		client := newTopologyTestClient(t, statusHandler(
			`{"version":{"number":"9.4.0","build_flavor":"serverless"}}`, nil, 0,
		))
		diags := checkDeploymentTopology(context.Background(), client)
		require.False(t, diags.HasError())
	})

	t.Run("Elastic Cloud proxy header present passes (confirmed cloud hosted)", func(t *testing.T) {
		client := newTopologyTestClient(t, statusHandler(
			`{"version":{"number":"9.4.3","build_flavor":"traditional"}}`,
			map[string]string{"X-Found-Handling-Cluster": "abc123"}, 0,
		))
		diags := checkDeploymentTopology(context.Background(), client)
		require.False(t, diags.HasError())
	})

	t.Run("alternate Elastic Cloud proxy header present passes", func(t *testing.T) {
		client := newTopologyTestClient(t, statusHandler(
			`{"version":{"number":"9.4.3","build_flavor":"traditional"}}`,
			map[string]string{"X-Found-Handling-Instance": "instance-0000000000"}, 0,
		))
		diags := checkDeploymentTopology(context.Background(), client)
		require.False(t, diags.HasError())
	})

	t.Run("traditional flavor with no cloud signal fails closed (confirmed self-managed)", func(t *testing.T) {
		client := newTopologyTestClient(t, statusHandler(
			`{"version":{"number":"9.4.0","build_flavor":"traditional"}}`, nil, 0,
		))
		diags := checkDeploymentTopology(context.Background(), client)
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary(), "Unsupported deployment topology")
		require.Contains(t, diags[0].Detail(), "Elastic Cloud Hosted or Serverless")
	})

	t.Run("empty build_flavor with no cloud signal fails closed", func(t *testing.T) {
		client := newTopologyTestClient(t, statusHandler(
			`{"version":{"number":"8.19.0"}}`, nil, 0,
		))
		diags := checkDeploymentTopology(context.Background(), client)
		require.True(t, diags.HasError())
	})

	t.Run("non-200 status is inconclusive and fails open", func(t *testing.T) {
		client := newTopologyTestClient(t, statusHandler(`{}`, nil, http.StatusServiceUnavailable))
		diags := checkDeploymentTopology(context.Background(), client)
		require.False(t, diags.HasError(), "an inconclusive probe must fail open, not block Create")
	})

	t.Run("malformed JSON body is inconclusive and fails open", func(t *testing.T) {
		client := newTopologyTestClient(t, statusHandler(`not json`, nil, 0))
		diags := checkDeploymentTopology(context.Background(), client)
		require.False(t, diags.HasError())
	})

	t.Run("unreachable Kibana is inconclusive and fails open", func(t *testing.T) {
		client, srv := newTopologyTestClientWithServer(t, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
		srv.Close() // Make the endpoint unreachable before the probe runs.

		diags := checkDeploymentTopology(context.Background(), client)
		require.False(t, diags.HasError(), "a failed status probe must fail open, not block Create")
	})
}
