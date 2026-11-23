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

package jobstate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/stretchr/testify/require"
)

// newJobStateTestClient starts a fake ES server whose ML job stats endpoint
// always reports jobState, and returns a scoped client pointed at it.
func newJobStateTestClient(t *testing.T, jobID, jobState string, statsCalls *atomic.Int32) *clients.ElasticsearchScopedClient {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"cluster_uuid": "test-cluster-uuid",
				"version": map[string]any{
					"number":       "8.19.0",
					"build_flavor": "default",
				},
			})
			return
		}
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/_stats") {
			statsCalls.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count": 1,
				"jobs": []any{
					map[string]any{
						"job_id": jobID,
						"state":  jobState,
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "unexpected request: " + r.Method + " " + r.URL.Path})
	}))
	t.Cleanup(srv.Close)

	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{srv.URL},
		Username:  "elastic",
		Password:  "changeme",
	})
	require.NoError(t, err)
	return clients.NewElasticsearchScopedClientForTest(esClient, []string{srv.URL})
}

// TestWaitForJobState_FailsFastOnMismatchedTerminalState verifies the fix for
// the inconsistency described in the issue: a job stuck in a terminal state
// (e.g. "failed") other than the desired one must fail immediately instead of
// polling until the context deadline.
func TestWaitForJobState_FailsFastOnMismatchedTerminalState(t *testing.T) {
	t.Parallel()

	var statsCalls atomic.Int32
	client := newJobStateTestClient(t, "test-job", "failed", &statsCalls)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	diags := waitForJobState(ctx, client, MLJobStateData{}, "test-job", "opened")
	elapsed := time.Since(start)

	require.True(t, diags.HasError(), "expected an error when the job settles into a mismatched terminal state")
	require.Less(t, elapsed, 1*time.Second, "must fail fast instead of polling to the context deadline")
	require.Equal(t, int32(1), statsCalls.Load())
}

// TestWaitForJobState_ReachesDesiredState verifies the happy path still works
// through the shared asyncutils.WaitForTerminalOrDesiredState helper.
func TestWaitForJobState_ReachesDesiredState(t *testing.T) {
	t.Parallel()

	var statsCalls atomic.Int32
	client := newJobStateTestClient(t, "test-job", "opened", &statsCalls)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	diags := waitForJobState(ctx, client, MLJobStateData{}, "test-job", "opened")
	require.False(t, diags.HasError(), "unexpected diagnostics: %s", diags)
}
