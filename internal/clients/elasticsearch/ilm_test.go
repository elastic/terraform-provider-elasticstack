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

package elasticsearch

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeGetIlmResponse(t *testing.T) {
	t.Parallel()

	// 2^53+1 cannot be represented exactly as float64; decoding _meta through
	// map[string]any would rewrite this integer.
	const largeInt = "9007199254740993"

	tests := []struct {
		name            string
		policyName      string
		body            string
		wantErr         bool
		wantErrContains string
		wantDate        string
		wantMetaInt     string
		wantMinAge      string
		wantRepo        string
		wantClone       any
		wantMetaEmpty   bool
	}{
		{
			name:       "preserves integer _meta and populates phases",
			policyName: "my-policy",
			body: `{
				"my-policy": {
					"modified_date": "2024-06-01T12:00:00.000Z",
					"policy": {
						"_meta": {"version": ` + largeInt + `, "owner": "search"},
						"phases": {
							"hot": {
								"min_age": "1d",
								"actions": {
									"searchable_snapshot": {
										"snapshot_repository": "repo-a",
										"force_merge_on_clone": false
									}
								}
							}
						}
					}
				}
			}`,
			wantDate:    "2024-06-01T12:00:00.000Z",
			wantMetaInt: largeInt,
			wantMinAge:  "1d",
			wantRepo:    "repo-a",
			wantClone:   false,
		},
		{
			name:       "omitted _meta stays empty",
			policyName: "my-policy",
			body: `{
				"my-policy": {
					"modified_date": "2024-06-01T12:00:00.000Z",
					"policy": {
						"phases": {
							"warm": {
								"min_age": "7d",
								"actions": {}
							}
						}
					}
				}
			}`,
			wantDate:      "2024-06-01T12:00:00.000Z",
			wantMetaEmpty: true,
			wantMinAge:    "7d",
		},
		{
			name:            "missing policy name is an error",
			policyName:      "other-policy",
			body:            `{"my-policy": {"modified_date": "2024-06-01T12:00:00.000Z", "policy": {"phases": {}}}}`,
			wantErr:         true,
			wantErrContains: "Unable to find",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, diags := decodeGetIlmResponse(tt.policyName, strings.NewReader(tt.body))
			if tt.wantErr {
				require.True(t, diags.HasError())
				require.Nil(t, got)
				if tt.wantErrContains != "" {
					assert.Contains(t, diags[0].Summary()+diags[0].Detail(), tt.wantErrContains)
				}
				return
			}

			require.False(t, diags.HasError(), "%s", diags)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantDate, got.ModifiedDate)

			if tt.wantMetaEmpty {
				assert.True(t, len(got.Metadata) == 0 || string(got.Metadata) == jsonNullLiteral)
			} else if tt.wantMetaInt != "" {
				assert.Contains(t, string(got.Metadata), tt.wantMetaInt)
				assert.NotContains(t, string(got.Metadata), "e+")
			}

			if tt.wantMinAge == "" {
				return
			}

			var phase IlmPhase
			var ok bool
			for _, p := range got.Phases {
				phase = p
				ok = true
				break
			}
			require.True(t, ok)
			assert.Equal(t, tt.wantMinAge, phase.MinAge)

			if tt.wantRepo == "" {
				return
			}
			ss, ok := phase.Actions["searchable_snapshot"]
			require.True(t, ok)
			assert.Equal(t, tt.wantRepo, ss["snapshot_repository"])
			assert.Equal(t, tt.wantClone, ss["force_merge_on_clone"])
		})
	}
}
