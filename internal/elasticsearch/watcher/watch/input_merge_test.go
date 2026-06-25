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

package watch

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/require"
)

// httpInputAPI returns the shape Elasticsearch Get Watch produces for an HTTP
// input with basic auth: the password leaf is the redacted sentinel, while
// non-secret fields (host, path) come back authoritative from the API.
func httpInputAPI() map[string]any {
	return map[string]any{
		"http": map[string]any{
			"request": map[string]any{
				"scheme": "http",
				"host":   "api.example",
				"port":   float64(9200),
				"path":   "/v1/data",
				"auth": map[string]any{
					"basic": map[string]any{
						"username": "acc-input-user",
						"password": elasticsearchWatcherRedactedSecret,
					},
				},
			},
		},
	}
}

func httpInputPrior(password any) map[string]any {
	return map[string]any{
		"http": map[string]any{
			"request": map[string]any{
				"scheme": "http",
				"host":   "old.example", // stale relative to API; must NOT win
				"port":   float64(9200),
				"path":   "/v1/data",
				"auth": map[string]any{
					"basic": map[string]any{
						"username": "acc-input-user",
						"password": password,
					},
				},
			},
		},
	}
}

// TestMergePreserveRedactedLeaves_inputBasicAuthPasswordPreserved covers the
// primary reported case: an HTTP input basic-auth password redacted by the
// Watcher API is restored from the prior known value, while non-redacted
// sibling input fields remain authoritative from the API response.
func TestMergePreserveRedactedLeaves_inputBasicAuthPasswordPreserved(t *testing.T) {
	t.Parallel()
	got := mergePreserveRedactedLeaves(httpInputAPI(), httpInputPrior("plain-input-secret"))
	req := got.(map[string]any)["http"].(map[string]any)["request"].(map[string]any)
	require.Equal(t, "api.example", req["host"], "non-redacted API host stays authoritative")
	require.Equal(t, "/v1/data", req["path"])
	pw := req["auth"].(map[string]any)["basic"].(map[string]any)["password"].(string)
	require.Equal(t, "plain-input-secret", pw, "redacted password restored from prior value")
}

// TestMergePreserveRedactedLeaves_inputPriorSentinelKeptAsIs: when the prior
// value at the redacted path is itself the sentinel (e.g. no prior concrete
// value ever existed), the sentinel is stored unchanged rather than replaced.
func TestMergePreserveRedactedLeaves_inputPriorSentinelKeptAsIs(t *testing.T) {
	t.Parallel()
	got := mergePreserveRedactedLeaves(httpInputAPI(), httpInputPrior(elasticsearchWatcherRedactedSecret))
	pw := got.(map[string]any)["http"].(map[string]any)["request"].(map[string]any)["auth"].(map[string]any)["basic"].(map[string]any)["password"]
	require.Equal(t, elasticsearchWatcherRedactedSecret, pw, "prior sentinel must not be replaced")
}

// TestMergePreserveRedactedLeaves_inputPriorObjectReplacesRedacted: a non-string
// prior value (e.g. an object) at a redacted path is preserved, mirroring the
// actions script-reference case.
func TestMergePreserveRedactedLeaves_inputPriorObjectReplacesRedacted(t *testing.T) {
	t.Parallel()
	api := map[string]any{
		"http": map[string]any{
			"request": map[string]any{
				"headers": map[string]any{
					"Authorization": elasticsearchWatcherRedactedSecret,
				},
			},
		},
	}
	prior := map[string]any{
		"http": map[string]any{
			"request": map[string]any{
				"headers": map[string]any{
					"Authorization": map[string]any{
						"source": "return 'Bearer ' + ctx.metadata.token",
						"lang":   "painless",
					},
				},
			},
		},
	}
	got := mergePreserveRedactedLeaves(api, prior)
	auth := got.(map[string]any)["http"].(map[string]any)["request"].(map[string]any)["headers"].(map[string]any)["Authorization"]
	require.Equal(t, map[string]any{
		"source": "return 'Bearer ' + ctx.metadata.token",
		"lang":   "painless",
	}, auth)
}

// TestMergePreserveRedactedLeaves_inputRoundTripJSON exercises the full
// JSON marshal/unmarshal round trip for an input payload.
func TestMergePreserveRedactedLeaves_inputRoundTripJSON(t *testing.T) {
	t.Parallel()
	apiJSON := `{"http":{"request":{"host":"api.example","path":"/v1/data","auth":{"basic":{"username":"u","password":"::es_redacted::"}}}}}`
	priorJSON := `{"http":{"request":{"host":"old.example","path":"/v1/data","auth":{"basic":{"username":"u","password":"secret"}}}}}`
	var api, prior any
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
	require.NoError(t, json.Unmarshal([]byte(priorJSON), &prior))
	got := mergePreserveRedactedLeaves(api, prior)
	out, err := json.Marshal(got)
	require.NoError(t, err)
	require.Contains(t, string(out), `"password":"secret"`)
	require.Contains(t, string(out), `"host":"api.example"`)
	require.NotContains(t, string(out), elasticsearchWatcherRedactedSecret)
}

// newInputTestWatch builds a minimal *models.Watch with an HTTP input body
// sufficient to drive fromAPIModel for the input redaction tests.
func newInputTestWatch(input map[string]any) *models.Watch {
	w := &models.Watch{WatchID: "test-watch"}
	w.Status.State.Active = false
	w.Body.Trigger = map[string]any{"schedule": map[string]any{"cron": "0 0/1 * * * ?"}}
	w.Body.Input = input
	return w
}

// TestFromAPIModel_inputRedactedWithPriorPreserved verifies the full fromAPIModel
// wiring: priorInput is applied to a redacted input password and the merged
// value (concrete password, authoritative API host) is stored in state JSON.
func TestFromAPIModel_inputRedactedWithPriorPreserved(t *testing.T) {
	t.Parallel()
	priorInput := jsontypes.NewNormalizedValue(`{"http":{"request":{"host":"old.example","path":"/v1/data","auth":{"basic":{"username":"acc-input-user","password":"plain-input-secret"}}}}}`)
	d := &Data{}
	diags := d.fromAPIModel(context.Background(), newInputTestWatch(httpInputAPI()), jsontypes.NewNormalizedNull(), priorInput)
	require.False(t, diags.HasError(), "diags: %v", diags)

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(d.Input.ValueString()), &got))
	req := got["http"].(map[string]any)["request"].(map[string]any)
	require.Equal(t, "api.example", req["host"], "non-redacted API host stays authoritative")
	pw := req["auth"].(map[string]any)["basic"].(map[string]any)["password"].(string)
	require.Equal(t, "plain-input-secret", pw, "redacted password restored from prior state")
}

// TestFromAPIModel_inputRedactedNoPriorKeepsSentinel: with no prior input
// (unknown/null), the redacted sentinel from the API is stored as-is.
func TestFromAPIModel_inputRedactedNoPriorKeepsSentinel(t *testing.T) {
	t.Parallel()
	d := &Data{}
	diags := d.fromAPIModel(context.Background(), newInputTestWatch(httpInputAPI()), jsontypes.NewNormalizedNull(), jsontypes.NewNormalizedNull())
	require.False(t, diags.HasError(), "diags: %v", diags)

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(d.Input.ValueString()), &got))
	pw := got["http"].(map[string]any)["request"].(map[string]any)["auth"].(map[string]any)["basic"].(map[string]any)["password"]
	require.Equal(t, elasticsearchWatcherRedactedSecret, pw, "sentinel stored when no prior concrete value exists")
}

// TestFromAPIModel_inputNilDefaultsToNone: a nil API input still defaults to
// {"none":{}} regardless of priorInput.
func TestFromAPIModel_inputNilDefaultsToNone(t *testing.T) {
	t.Parallel()
	d := &Data{}
	diags := d.fromAPIModel(context.Background(), newInputTestWatch(nil), jsontypes.NewNormalizedNull(), jsontypes.NewNormalizedValue(`{"http":{"request":{"host":"h"}}}`))
	require.False(t, diags.HasError(), "diags: %v", diags)
	require.JSONEq(t, `{"none":{}}`, d.Input.ValueString())
}

// TestFromAPIModel_inputPriorObjectReplacesRedacted: at the fromAPIModel level,
// a non-string prior value at a redacted input path is preserved.
func TestFromAPIModel_inputPriorObjectReplacesRedacted(t *testing.T) {
	t.Parallel()
	apiInput := map[string]any{
		"http": map[string]any{
			"request": map[string]any{
				"headers": map[string]any{
					"Authorization": elasticsearchWatcherRedactedSecret,
				},
			},
		},
	}
	priorInput := jsontypes.NewNormalizedValue(`{"http":{"request":{"headers":{"Authorization":{"source":"return 'Bearer x'","lang":"painless"}}}}}`)
	d := &Data{}
	diags := d.fromAPIModel(context.Background(), newInputTestWatch(apiInput), jsontypes.NewNormalizedNull(), priorInput)
	require.False(t, diags.HasError(), "diags: %v", diags)

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(d.Input.ValueString()), &got))
	auth := got["http"].(map[string]any)["request"].(map[string]any)["headers"].(map[string]any)["Authorization"]
	require.Equal(t, map[string]any{"source": "return 'Bearer x'", "lang": "painless"}, auth)
}
