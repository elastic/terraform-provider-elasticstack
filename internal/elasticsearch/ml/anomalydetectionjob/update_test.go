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

package anomalydetectionjob

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// fakeMLJobServer keeps a single anomaly detection job in memory and applies
// the description field of _update requests the way Elasticsearch does: a
// field that is present in the body replaces the stored value, an absent field
// leaves it untouched.
type fakeMLJobServer struct {
	mu          sync.Mutex
	jobID       string
	description string
	updates     []map[string]any
}

func (f *fakeMLJobServer) handler(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	w.Header().Set("X-Elastic-Product", "Elasticsearch")
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.URL.Path == "/":
		_ = json.NewEncoder(w).Encode(map[string]any{
			"cluster_uuid": "cluster-uuid",
			"version":      map[string]any{"number": "8.19.0", "build_flavor": "default"},
		})
	case r.Method == http.MethodPost && r.URL.Path == "/_ml/anomaly_detectors/"+f.jobID+"/_update":
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		f.updates = append(f.updates, body)
		if d, ok := body["description"].(string); ok {
			f.description = d
		}
		_ = json.NewEncoder(w).Encode(f.job())
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/_ml/anomaly_detectors/"+f.jobID):
		_ = json.NewEncoder(w).Encode(map[string]any{"count": 1, "jobs": []any{f.job()}})
	default:
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "unexpected request: " + r.Method + " " + r.URL.Path})
	}
}

func (f *fakeMLJobServer) job() map[string]any {
	job := map[string]any{
		"job_id":            f.jobID,
		"job_type":          "anomaly_detector",
		"job_version":       "8.19.0",
		"create_time":       1700000000000,
		"model_snapshot_id": "",
		"analysis_config": map[string]any{
			"bucket_span": "15m",
			"detectors":   []any{map[string]any{"function": "count", "detector_index": 0}},
		},
		"analysis_limits":    map[string]any{"model_memory_limit": "128mb"},
		"data_description":   map[string]any{"time_field": "@timestamp", "time_format": "epoch_ms"},
		"results_index_name": "shared",
		"allow_lazy_open":    false,
	}
	// Elasticsearch echoes whatever string is stored, including "".
	if f.description != "" {
		job["description"] = f.description
	}
	return job
}

func newFakeMLJobClient(t *testing.T, f *fakeMLJobServer) *clients.ElasticsearchScopedClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(f.handler))
	t.Cleanup(srv.Close)

	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{srv.URL},
		Username:  "elastic",
		Password:  "changeme",
	})
	require.NoError(t, err)
	return clients.NewElasticsearchScopedClientForTest(esClient, []string{srv.URL})
}

// Removing description from the configuration plans it as null. The update
// must clear it in Elasticsearch, otherwise the read-after-write returns the
// old value and Terraform fails with "Provider produced inconsistent result
// after apply".
func TestUpdateAnomalyDetectionJob_removedDescriptionIsCleared(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	fake := &fakeMLJobServer{jobID: "job-unit-test", description: "steady description"}
	client := newFakeMLJobClient(t, fake)

	prior := baselineTFModelForBuildFromPlan(ctx, t)
	plan := prior
	plan.Description = types.StringNull()

	result, diags := updateAnomalyDetectionJob(ctx, client, entitycore.WriteRequest[TFModel]{
		Plan:    plan,
		Prior:   &prior,
		WriteID: fake.jobID,
	})
	require.False(t, diags.HasError(), "%v", diags)
	require.Zero(t, diags.WarningsCount(), "unexpected warnings: %v", diags)
	require.Len(t, fake.updates, 1, "expected one Update Job API call")
	require.Equal(t, map[string]any{"description": ""}, fake.updates[0])

	state, found, diags := readAnomalyDetectionJob(ctx, client, fake.jobID, result.Model)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, found)
	require.True(t, state.Description.IsNull(), "description should be null after read, got %s", state.Description)
}

func TestUpdateAPIModel_BuildFromPlan_removedDescriptionSendsEmptyString(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baselineTFModelForBuildFromPlan(ctx, t)

	plan := prior
	plan.Description = types.StringNull()

	var u UpdateAPIModel
	hasChanges, diags := u.BuildFromPlan(ctx, &plan, &prior)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, hasChanges)
	require.NotNil(t, u.Description)
	require.Empty(t, *u.Description)
}

func TestUpdateAPIModel_BuildFromPlan_unknownDescriptionIsOmitted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baselineTFModelForBuildFromPlan(ctx, t)

	plan := prior
	plan.Description = types.StringUnknown()

	var u UpdateAPIModel
	hasChanges, diags := u.BuildFromPlan(ctx, &plan, &prior)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, hasChanges)
	require.Nil(t, u.Description)
}
