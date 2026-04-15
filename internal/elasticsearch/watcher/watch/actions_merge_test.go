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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeActionsPreservingRedactedLeaves_nestedSecret(t *testing.T) {
	t.Parallel()
	api := map[string]any{
		"a1": map[string]any{
			"webhook": map[string]any{
				"host": "new.example",
				"auth": map[string]any{
					"basic": map[string]any{
						"username": "u",
						"password": elasticsearchWatcherRedactedSecret,
					},
				},
			},
		},
	}
	prior := map[string]any{
		"a1": map[string]any{
			"webhook": map[string]any{
				"host": "old.example",
				"auth": map[string]any{
					"basic": map[string]any{
						"username": "u",
						"password": "plain-secret",
					},
				},
			},
		},
	}
	got := mergeActionsPreservingRedactedLeaves(api, prior)
	host := got["a1"].(map[string]any)["webhook"].(map[string]any)["host"].(string)
	require.Equal(t, "new.example", host, "non-redacted API fields stay authoritative")
	pw := got["a1"].(map[string]any)["webhook"].(map[string]any)["auth"].(map[string]any)["basic"].(map[string]any)["password"].(string)
	require.Equal(t, "plain-secret", pw)
}

func TestMergeActionsPreservingRedactedLeaves_noPriorValueKeepsRedacted(t *testing.T) {
	t.Parallel()
	api := map[string]any{
		"x": map[string]any{"secret": elasticsearchWatcherRedactedSecret},
	}
	got := mergeActionsPreservingRedactedLeaves(api, map[string]any{})
	sec := got["x"].(map[string]any)["secret"].(string)
	require.Equal(t, elasticsearchWatcherRedactedSecret, sec)
}

func TestMergeActionsPreservingRedactedLeaves_priorRedactedNotReplaced(t *testing.T) {
	t.Parallel()
	api := map[string]any{"k": elasticsearchWatcherRedactedSecret}
	prior := map[string]any{"k": elasticsearchWatcherRedactedSecret}
	got := mergeActionsPreservingRedactedLeaves(api, prior)
	require.Equal(t, elasticsearchWatcherRedactedSecret, got["k"])
}

func TestMergeActionsPreservingRedactedLeaves_mismatchedPath(t *testing.T) {
	t.Parallel()
	api := map[string]any{
		"a": map[string]any{"nested": elasticsearchWatcherRedactedSecret},
	}
	prior := map[string]any{
		"b": map[string]any{"nested": "wrong-branch"},
	}
	got := mergeActionsPreservingRedactedLeaves(api, prior)
	require.Equal(t, elasticsearchWatcherRedactedSecret, got["a"].(map[string]any)["nested"])
}

func TestMergeActionsPreservingRedactedLeaves_arrayByIndex(t *testing.T) {
	t.Parallel()
	api := map[string]any{
		"list": []any{
			elasticsearchWatcherRedactedSecret,
			"visible",
		},
	}
	prior := map[string]any{
		"list": []any{"first-secret", "was-visible"},
	}
	got := mergeActionsPreservingRedactedLeaves(api, prior)
	arr := got["list"].([]any)
	require.Equal(t, "first-secret", arr[0])
	require.Equal(t, "visible", arr[1])
}

func TestMergeActionsPreservingRedactedLeaves_priorTypeMismatch(t *testing.T) {
	t.Parallel()
	api := map[string]any{"k": elasticsearchWatcherRedactedSecret}
	prior := map[string]any{"k": map[string]any{"not": "a string"}}
	got := mergeActionsPreservingRedactedLeaves(api, prior)
	require.Equal(t, elasticsearchWatcherRedactedSecret, got["k"])
}

func TestMergePreserveRedactedLeaves_nonStringLeavesUnchanged(t *testing.T) {
	t.Parallel()
	api := map[string]any{"n": float64(42), "b": true}
	prior := map[string]any{"n": float64(1), "b": false}
	got := mergePreserveRedactedLeaves(api, prior).(map[string]any)
	require.InEpsilon(t, float64(42), got["n"], 1e-9)
	require.Equal(t, true, got["b"])
}

func TestMergeActionsPreservingRedactedLeaves_roundTripJSON(t *testing.T) {
	t.Parallel()
	apiJSON := `{"w":{"webhook":{"host":"h","auth":{"basic":{"password":"::es_redacted::"}}}}}`
	priorJSON := `{"w":{"webhook":{"host":"old","auth":{"basic":{"password":"secret"}}}}}`
	var api, prior any
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
	require.NoError(t, json.Unmarshal([]byte(priorJSON), &prior))
	got := mergeActionsPreservingRedactedLeaves(api.(map[string]any), prior)
	out, err := json.Marshal(got)
	require.NoError(t, err)
	require.Contains(t, string(out), `"password":"secret"`)
	require.Contains(t, string(out), `"host":"h"`)
}
