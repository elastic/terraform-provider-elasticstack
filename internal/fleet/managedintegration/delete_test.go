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
	"sync/atomic"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConflictHintDiagnostics covers a bug found in review: this function
// used to take the full diag.Diagnostics from fleetclient.DeleteManagedIntegration
// and infer a conflict by pattern-matching diagutil.ReportUnknownHTTPError's
// generated summary text ("... got HTTP 409 ..."), which was brittle against
// wording changes or a switch to a different error-reporting helper (e.g.
// diagutil.ReportKibanaBoomHTTPError, whose summary is caller-supplied and
// might not contain "HTTP 409" at all). fleetclient.DeleteManagedIntegration now
// derives the conflict signal from the final HTTP status code observed across
// retries and reports it as a plain bool (see internal/clients/fleet/
// managed_integration.go and TestDeleteManagedIntegration/
// max_retries_exhausted_returns_error and
// transport_error_after_409_resets_is_conflict), so this function no longer needs to
// inspect diagnostic text at all -- it is now a pure "build the hint" helper
// that deleteAgentlessPolicy only calls once it already knows, authoritatively,
// that the delete failed with a conflict.
func TestConflictHintDiagnostics(t *testing.T) {
	t.Parallel()

	hint := conflictHintDiagnostics()
	require.Len(t, hint, 1)
	assert.True(t, hint.HasError())
	assert.Contains(t, hint[0].Summary(), "conflict")
	assert.Contains(t, hint[0].Detail(), "force_delete")
	assert.Contains(t, hint[0].Detail(), "force=true")
}

func deleteCallbackTestModel(t *testing.T, forceDelete bool) agentlessPolicyModel {
	t.Helper()
	m := baseTestModel(t)
	m.PolicyID = types.StringValue("policy-1")
	m.ForceDelete = types.BoolValue(forceDelete)
	return m
}

// TestDeleteAgentlessPolicy_callback exercises deleteAgentlessPolicy against
// httptest Fleet DELETE /api/fleet/managed_integrations/{id} (DeleteManagedIntegration).
func TestDeleteAgentlessPolicy_callback(t *testing.T) {
	t.Run("200 success", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"policy-1"}`)
		})
		client := newTopologyTestClient(t, mux)

		diags := deleteAgentlessPolicy(context.Background(), client, "policy-1", "default", deleteCallbackTestModel(t, false))
		require.False(t, diags.HasError(), "%v", diags)
	})

	t.Run("404 is idempotent no-op", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"not found"}`)
		})
		client := newTopologyTestClient(t, mux)

		diags := deleteAgentlessPolicy(context.Background(), client, "policy-1", "default", deleteCallbackTestModel(t, false))
		require.False(t, diags.HasError(), "%v", diags)
	})

	t.Run("persistent 409 with force_delete=false appends force_delete hint", func(t *testing.T) {
		var calls atomic.Int64
		mux := http.NewServeMux()
		mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `{"statusCode":409,"error":"Conflict","message":"provisioning"}`)
		})
		client := newTopologyTestClient(t, mux)

		diags := deleteAgentlessPolicy(context.Background(), client, "policy-1", "default", deleteCallbackTestModel(t, false))
		require.True(t, diags.HasError())
		require.Equal(t, int64(kibanautil.ConflictMaxAttempts), calls.Load())
		require.GreaterOrEqual(t, len(diags.Errors()), 2, "expect HTTP error plus force_delete hint")
		assert.Contains(t, diags.Errors()[len(diags.Errors())-1].Summary(), "conflict")
		assert.Contains(t, diags.Errors()[len(diags.Errors())-1].Detail(), "force_delete")
	})

	t.Run("persistent 409 with force_delete=true does not append hint", func(t *testing.T) {
		var calls atomic.Int64
		var capturedQuery string
		mux := http.NewServeMux()
		mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			capturedQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `{"statusCode":409,"error":"Conflict","message":"provisioning"}`)
		})
		client := newTopologyTestClient(t, mux)

		diags := deleteAgentlessPolicy(context.Background(), client, "policy-1", "default", deleteCallbackTestModel(t, true))
		require.True(t, diags.HasError())
		require.Equal(t, int64(kibanautil.ConflictMaxAttempts), calls.Load())
		require.Len(t, diags.Errors(), 1, "force_delete=true must not append conflictHintDiagnostics")
		assert.NotContains(t, diags.Errors()[0].Summary(), "Managed integration delete conflict")
		assert.Contains(t, capturedQuery, "force=true", "force_delete=true must map to ?force=true")
	})
}
