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

package ilm

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandAction(t *testing.T) {
	t.Parallel()

	v80 := version.Must(version.NewVersion("8.0.0"))
	v82 := version.Must(version.NewVersion("8.2.0"))
	v814 := version.Must(version.NewVersion("8.14.0"))

	tests := []struct {
		name          string
		action        []any
		serverVersion *version.Version
		settings      []string
		expected      map[string]any
		errorSummary  string
		errorDetail   string
	}{
		{
			name:          "ignores nil action body",
			action:        []any{nil},
			serverVersion: v80,
			settings:      []string{"priority"},
			expected:      map[string]any{},
		},
		{
			name:          "decodes allocation filter JSON",
			serverVersion: v80,
			action: []any{map[string]any{
				"include": `{"box_type":"warm"}`,
				"exclude": `{"rack":"rack-a"}`,
				"require": `{"zone":"zone-1"}`,
			}},
			settings: []string{"include", "exclude", "require"},
			expected: map[string]any{
				"include": map[string]any{"box_type": "warm"},
				"exclude": map[string]any{"rack": "rack-a"},
				"require": map[string]any{"zone": "zone-1"},
			},
		},
		{
			name:          "filters empty values but keeps skip empty settings",
			serverVersion: v80,
			action: []any{map[string]any{
				"priority":              0,
				"number_of_replicas":    0,
				"total_shards_per_node": -1,
				"max_age":               "   ",
			}},
			settings: []string{"priority", "number_of_replicas", "total_shards_per_node", "max_age"},
			expected: map[string]any{
				"priority":              0,
				"number_of_replicas":    0,
				"total_shards_per_node": -1,
			},
		},
		{
			name:          "keeps version gated settings on supported server",
			serverVersion: v814,
			action: []any{map[string]any{
				"allow_write_after_shrink": true,
				"max_primary_shard_docs":   100,
			}},
			settings: []string{"allow_write_after_shrink", "max_primary_shard_docs"},
			expected: map[string]any{
				"allow_write_after_shrink": true,
				"max_primary_shard_docs":   100,
			},
		},
		{
			name:          "skips unsupported default values",
			serverVersion: v80,
			action: []any{map[string]any{
				"min_age":                "",
				"min_docs":               0,
				"min_primary_shard_size": "",
			}},
			settings: []string{"min_age", "min_docs", "min_primary_shard_size"},
			expected: map[string]any{},
		},
		{
			name:          "errors on unsupported non default setting",
			serverVersion: v80,
			action: []any{map[string]any{
				"min_age": "1d",
			}},
			settings:     []string{"min_age"},
			errorSummary: "Unsupported ILM setting",
			errorDetail:  "[min_age] is not supported",
		},
		{
			name:          "allows version gated setting when server version is unknown",
			serverVersion: nil,
			action: []any{map[string]any{
				"min_age": "1d",
			}},
			settings: []string{"min_age"},
			expected: map[string]any{"min_age": "1d"},
		},
		{
			name:          "errors on invalid JSON filter",
			serverVersion: v82,
			action: []any{map[string]any{
				"include": "{",
			}},
			settings:     []string{"include"},
			errorSummary: "Invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			action, diags := expandAction(tt.action, tt.serverVersion, tt.settings...)

			if tt.errorSummary != "" {
				require.True(t, diags.HasError())
				require.Nil(t, action)
				require.Contains(t, diags[0].Summary(), tt.errorSummary)
				if tt.errorDetail != "" {
					require.Contains(t, diags[0].Detail(), tt.errorDetail)
				}
				return
			}

			require.False(t, diags.HasError(), "%s", diags)
			require.Equal(t, tt.expected, action)
		})
	}
}

func TestExpandPhase(t *testing.T) {
	t.Parallel()

	v814 := version.Must(version.NewVersion("8.14.0"))

	tests := []struct {
		name         string
		input        func() map[string]any
		server       *version.Version
		expected     *models.Phase
		errorSummary string
		errorDetail  string
	}{
		{
			name:   "expands supported actions",
			server: v814,
			input: func() map[string]any {
				return map[string]any{
					"min_age": "30d",
					"allocate": []any{map[string]any{
						"number_of_replicas": 1,
						"include":            `{"box_type":"cold"}`,
					}},
					"downsample": []any{map[string]any{
						"fixed_interval": "1h",
						"wait_timeout":   "2h",
					}},
					"freeze":   []any{map[string]any{"enabled": true}},
					"readonly": []any{map[string]any{"enabled": true}},
					"searchable_snapshot": []any{map[string]any{
						"snapshot_repository": "repo-a",
						"force_merge_index":   true,
					}},
					"set_priority": []any{map[string]any{"priority": 100}},
					"shrink":       []any{map[string]any{"allow_write_after_shrink": true}},
					"unfollow":     []any{map[string]any{"enabled": true}},
				}
			},
			expected: &models.Phase{
				MinAge: "30d",
				Actions: map[string]models.Action{
					"allocate": {
						"number_of_replicas": 1,
						"include":            map[string]any{"box_type": "cold"},
					},
					"downsample": {
						"fixed_interval": "1h",
						"wait_timeout":   "2h",
					},
					"freeze":   {},
					"readonly": {},
					"searchable_snapshot": {
						"snapshot_repository": "repo-a",
						"force_merge_index":   true,
					},
					"set_priority": {"priority": 100},
					"shrink":       {"allow_write_after_shrink": true},
					"unfollow":     {},
				},
			},
		},
		{
			name:   "skips disabled and invalid actions",
			server: v814,
			input: func() map[string]any {
				return map[string]any{
					"freeze":   []any{map[string]any{"enabled": false}},
					"readonly": map[string]any{"enabled": true},
					"unfollow": []any{map[string]any{"enabled": false}},
				}
			},
			expected: &models.Phase{
				Actions: map[string]models.Action{},
			},
		},
		{
			name:   "returns unknown action error",
			server: v814,
			input: func() map[string]any {
				return map[string]any{
					"mystery": []any{map[string]any{}},
				}
			},
			errorSummary: "Unknown action defined.",
			errorDetail:  `Configured action "mystery" is not supported`,
		},
		{
			name:   "propagates action expansion error",
			server: v814,
			input: func() map[string]any {
				return map[string]any{
					"allocate": []any{map[string]any{"include": "{"}},
				}
			},
			errorSummary: "Invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase, diags := expandPhase(tt.input(), tt.server)

			if tt.errorSummary != "" {
				require.True(t, diags.HasError())
				require.Nil(t, phase)
				require.Contains(t, diags[0].Summary(), tt.errorSummary)
				if tt.errorDetail != "" {
					require.Contains(t, diags[0].Detail(), tt.errorDetail)
				}
				return
			}

			require.False(t, diags.HasError(), "%s", diags)
			require.Equal(t, tt.expected, phase)
		})
	}
}

func TestExpandIlmPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		metadata     string
		phases       func() map[string]map[string]any
		server       *version.Version
		expected     *models.Policy
		errorSummary string
		errorDetail  string
	}{
		{
			name:     "expands metadata and phases",
			metadata: `{"owner":"search-team"}`,
			server:   nil,
			phases: func() map[string]map[string]any {
				return map[string]map[string]any{
					ilmPhaseHot: {
						"min_age":      "1d",
						"set_priority": []any{map[string]any{"priority": 50}},
						"readonly":     []any{map[string]any{"enabled": true}},
					},
					ilmPhaseWarm: nil,
				}
			},
			expected: &models.Policy{
				Name:     "policy-a",
				Metadata: map[string]any{"owner": "search-team"},
				Phases: map[string]models.Phase{
					ilmPhaseHot: {
						MinAge: "1d",
						Actions: map[string]models.Action{
							"set_priority": {"priority": 50},
							"readonly":     {},
						},
					},
				},
			},
		},
		{
			name:     "ignores blank metadata",
			metadata: " \n\t ",
			server:   nil,
			phases: func() map[string]map[string]any {
				return map[string]map[string]any{}
			},
			expected: &models.Policy{
				Name:   "policy-a",
				Phases: map[string]models.Phase{},
			},
		},
		{
			name:         "returns metadata decode error",
			metadata:     "{",
			server:       nil,
			phases:       func() map[string]map[string]any { return map[string]map[string]any{} },
			errorSummary: "Invalid metadata JSON",
		},
		{
			name:     "propagates phase expansion error",
			metadata: "",
			server:   nil,
			phases: func() map[string]map[string]any {
				return map[string]map[string]any{
					ilmPhaseHot: {
						"mystery": []any{map[string]any{}},
					},
				}
			},
			errorSummary: "Unknown action defined.",
			errorDetail:  `Configured action "mystery" is not supported`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, diags := expandIlmPolicy("policy-a", tt.metadata, tt.phases(), tt.server)

			if tt.errorSummary != "" {
				require.True(t, diags.HasError())
				require.Nil(t, policy)
				assert.Contains(t, diags[0].Summary(), tt.errorSummary)
				if tt.errorDetail != "" {
					assert.Contains(t, diags[0].Detail(), tt.errorDetail)
				}
				return
			}

			require.False(t, diags.HasError(), "%s", diags)
			require.Equal(t, tt.expected, policy)
		})
	}
}
