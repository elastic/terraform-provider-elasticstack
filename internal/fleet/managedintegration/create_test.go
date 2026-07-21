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
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// minimalManagedIntegrationCreateResponse is a well-formed
// PostFleetManagedIntegrations 200 response body, just complete enough for
// the create callback to decode the assigned policy id. The exact field values
// are not asserted by the tests below -- they only care whether the fleet
// POST endpoint was hit at all.
const minimalManagedIntegrationCreateResponse = `{"item":{` +
	`"id":"pp-1","name":"test-policy",` +
	`"created_at":"2026-01-01T00:00:00.000Z","created_by":"elastic",` +
	`"updated_at":"2026-01-01T00:00:00.000Z","updated_by":"elastic",` +
	`"inputs":{},` +
	`"package":{"name":"cloud_security_posture","version":"3.4.0"}` +
	`}}`

// fleetCreateCallRecorder builds an http.Handler that serves both the
// `/api/status` topology probe and the `/api/fleet/managed_integrations`
// create endpoint from a single httptest server (see
// newTopologyTestClient in topology_test.go, which this reuses
// unmodified -- it already accepts any http.Handler). The returned
// *bool reports whether the fleet create endpoint was ever hit, which is
// the thing these tests care about: did checkDeploymentTopology's verdict
// (or skip_topology_check bypassing it) actually gate the POST call in
// createAgentlessPolicy, or not.
func fleetCreateCallRecorder(statusBody string, statusHeaders map[string]string) (http.Handler, *bool) {
	fleetPostCalled := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, _ *http.Request) {
		for k, v := range statusHeaders {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, statusBody)
	})
	mux.HandleFunc("/api/fleet/managed_integrations", func(w http.ResponseWriter, _ *http.Request) {
		fleetPostCalled = true
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, minimalManagedIntegrationCreateResponse)
	})
	return mux, &fleetPostCalled
}

// selfManagedStatusBody and cloudStatusBody mirror the fixtures already used
// by TestCheckDeploymentTopology in topology_test.go: a
// "traditional" build_flavor with no Elastic Cloud proxy header is
// classified self-managed (fail closed); the same flavor with the proxy
// header present is classified Elastic Cloud Hosted (fail open, i.e. not
// blocked).
const selfManagedStatusBody = `{"version":{"number":"9.4.0","build_flavor":"traditional"}}`

// cloudProxyHeader reuses topology.go's own cloudProxyResponseHeaders slice
// (rather than a fresh "X-Found-Handling-Cluster" string literal) so this
// file doesn't push that header name's literal-occurrence count over
// goconst's threshold -- see the identical header literal already used by
// TestCheckDeploymentTopology in this package.
var cloudProxyHeader = map[string]string{cloudProxyResponseHeaders[0]: "abc123"}

// TestCreateAgentlessPolicy_topologyGatesFleetCall is Task 6's reviewer
// finding #2 closed: it drives checkDeploymentTopology through
// createAgentlessPolicy's actual call site (rather than only unit-testing
// checkDeploymentTopology in isolation, as TestCheckDeploymentTopology
// does), and it is also the test for the skip_topology_check escape hatch
// added to resolve design.md's Open Question 6: a self-managed-shaped
// deployment normally blocks Create, but skip_topology_check = true bypasses
// checkDeploymentTopology entirely (and therefore also skips its live
// /api/status HTTP call -- see the fleetPostCalled assertions below, which
// prove the fleet POST behavior in all three cases, not just presence/
// absence of an error diagnostic).
// Subtests below deliberately do not call t.Parallel(): they build clients
// via newTopologyTestClient, which calls clearKibanaEnvOverrides -> t.Setenv,
// and t.Setenv is documented as incompatible with parallel tests (matching
// the non-parallel style already used by TestCheckDeploymentTopology in
// topology_test.go, for the same reason).
func TestCreateAgentlessPolicy_topologyGatesFleetCall(t *testing.T) {
	t.Run("self-managed topology and skip_topology_check=false (default) blocks the fleet POST", func(t *testing.T) {
		handler, fleetPostCalled := fleetCreateCallRecorder(selfManagedStatusBody, nil)
		client := newTopologyTestClient(t, handler)

		plan := baseTestModel(t)
		plan.SkipTopologyCheck = types.BoolValue(false)

		result, diags := createAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
			Plan:    plan,
			SpaceID: "default",
		})

		require.True(t, diags.HasError(), "self-managed topology must fail closed")
		require.Contains(t, diags[0].Summary(), "Unsupported deployment topology")
		require.False(t, *fleetPostCalled, "the fleet create endpoint must not be called once the topology check fails closed")
		require.Equal(t, agentlessPolicyModel{}, result.Model)
	})

	t.Run("self-managed topology with skip_topology_check=true proceeds to the fleet POST", func(t *testing.T) {
		handler, fleetPostCalled := fleetCreateCallRecorder(selfManagedStatusBody, nil)
		client := newTopologyTestClient(t, handler)

		plan := baseTestModel(t)
		plan.SkipTopologyCheck = types.BoolValue(true)

		_, diags := createAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
			Plan:    plan,
			SpaceID: "default",
		})

		require.False(t, diags.HasError(), "skip_topology_check=true must bypass what would otherwise be a fail-closed classification")
		require.True(t, *fleetPostCalled, "Create must proceed to the fleet POST once the topology check is explicitly skipped")
	})

	t.Run("cloud-hosted topology proceeds to the fleet POST regardless of skip_topology_check", func(t *testing.T) {
		for _, skip := range []bool{false, true} {
			handler, fleetPostCalled := fleetCreateCallRecorder(selfManagedStatusBody /* unused when headers set */, cloudProxyHeader)
			client := newTopologyTestClient(t, handler)

			plan := baseTestModel(t)
			plan.SkipTopologyCheck = types.BoolValue(skip)

			_, diags := createAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
				Plan:    plan,
				SpaceID: "default",
			})

			require.False(t, diags.HasError(), "a confirmed cloud-hosted topology must never block Create (skip_topology_check=%v)", skip)
			require.True(t, *fleetPostCalled, "Create must reach the fleet POST for a confirmed cloud-hosted topology (skip_topology_check=%v)", skip)
		}
	})
}
