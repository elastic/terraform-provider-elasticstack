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

package integrationpolicy_test

import (
	"context"
	"maps"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	integrationpolicy "github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

type privateData map[string]string

func (p *privateData) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if val, ok := (*p)[key]; ok {
		return []byte(val), nil
	}
	return nil, nil
}

func (p *privateData) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	(*p)[key] = string(value)
	return nil
}

type Map = map[string]any

func TestHandleRespSecrets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	private := privateData{"secrets": `{"known-secret":"secret", "known-secret-1":"secret1", "known-secret-2":"secret2"}`}

	secretRefs := &[]kbapi.PackagePolicySecretRef{
		{Id: "known-secret"},
		{Id: "known-secret-1"},
		{Id: "known-secret-2"},
	}

	tests := []struct {
		name  string
		input Map
		want  Map
	}{
		{
			name:  "converts plain",
			input: Map{"k": "v"},
			want:  Map{"k": "v"},
		},
		{
			name:  "converts wrapped",
			input: Map{"k": Map{"type": "string", "value": "v"}},
			want:  Map{"k": "v"},
		},
		{
			name:  "converts wrapped nil",
			input: Map{"k": Map{"type": "string"}},
			want:  Map{},
		},
		{
			name:  "converts secret",
			input: Map{"k": Map{"isSecretRef": true, "id": "known-secret"}},
			want:  Map{"k": "secret"},
		},
		{
			name:  "converts secret with multiple values",
			input: Map{"k": Map{"isSecretRef": true, "ids": []any{"known-secret-1", "known-secret-2"}}},
			want:  Map{"k": []any{"secret1", "secret2"}},
		},
		{
			name:  "converts wrapped secret",
			input: Map{"k": Map{"type": "password", "value": Map{"isSecretRef": true, "id": "known-secret"}}},
			want:  Map{"k": "secret"},
		},
		{
			name:  "converts unknown secret",
			input: Map{"k": Map{"isSecretRef": true, "id": "unknown-secret"}},
			want:  Map{"k": Map{"isSecretRef": true, "id": "unknown-secret"}},
		},
		{
			name:  "converts wrapped unknown secret",
			input: Map{"k": Map{"type": "password", "value": Map{"isSecretRef": true, "id": "unknown-secret"}}},
			want:  Map{"k": Map{"isSecretRef": true, "id": "unknown-secret"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := kbapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs: map[string]kbapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]kbapi.PackagePolicyInputStream{"stream1": {Vars: new(maps.Clone(tt.input))}},
						Vars:    new(maps.Clone(tt.input)),
					},
				},
				Vars: new(maps.Clone(tt.input)),
			}
			wants := kbapi.PackagePolicy{
				Inputs: map[string]kbapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]kbapi.PackagePolicyInputStream{"stream1": {Vars: new(tt.want)}},
						Vars:    &tt.want,
					},
				},
				Vars: &tt.want,
			}

			diags := integrationpolicy.HandleRespSecrets(ctx, &resp, &private)
			require.Empty(t, diags)
			// Policy vars
			got := *resp.Vars
			want := *wants.Vars
			require.Equal(t, want, got)

			// Input vars
			got = *resp.Inputs["input1"].Vars
			want = *wants.Inputs["input1"].Vars
			require.Equal(t, want, got)

			// Stream vars
			got = *(*resp.Inputs["input1"].Streams)["stream1"].Vars
			want = *(*wants.Inputs["input1"].Streams)["stream1"].Vars
			require.Equal(t, want, got)

			// privateData
			privateWants := privateData{"secrets": `{"known-secret":"secret","known-secret-1":"secret1","known-secret-2":"secret2"}`}
			require.Equal(t, privateWants, private)
		})
	}
}

func TestHandleReqRespSecrets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	secretRefs := &[]kbapi.PackagePolicySecretRef{
		{Id: "known-secret"},
		{Id: "known-secret-1"},
		{Id: "known-secret-2"},
	}

	tests := []struct {
		name      string
		reqInput  Map
		respInput Map
		want      Map
	}{
		{
			name:      "converts plain",
			reqInput:  Map{"k": "v"},
			respInput: Map{"k": "v"},
			want:      Map{"k": "v"},
		},
		{
			name:      "converts wrapped",
			reqInput:  Map{"k": "v"},
			respInput: Map{"k": Map{"type": "string", "value": "v"}},
			want:      Map{"k": "v"},
		},
		{
			name:      "converts wrapped nil",
			reqInput:  Map{"k": nil},
			respInput: Map{"k": Map{"type": "string"}},
			want:      Map{},
		},
		{
			name:      "converts secret",
			reqInput:  Map{"k": "secret"},
			respInput: Map{"k": Map{"isSecretRef": true, "id": "known-secret"}},
			want:      Map{"k": "secret"},
		},
		{
			name:      "converts secret with multiple values",
			reqInput:  Map{"k": []any{"secret1", "secret2"}},
			respInput: Map{"k": Map{"isSecretRef": true, "ids": []any{"known-secret-1", "known-secret-2"}}},
			want:      Map{"k": []any{"secret1", "secret2"}},
		},
		{
			name:      "converts wrapped secret",
			reqInput:  Map{"k": "secret"},
			respInput: Map{"k": Map{"type": "password", "value": Map{"isSecretRef": true, "id": "known-secret"}}},
			want:      Map{"k": "secret"},
		},
		{
			name:      "converts unknown secret",
			reqInput:  Map{"k": Map{"isSecretRef": true, "id": "unknown-secret"}},
			respInput: Map{"k": Map{"isSecretRef": true, "id": "unknown-secret"}},
			want:      Map{"k": Map{"isSecretRef": true, "id": "unknown-secret"}},
		},
		{
			name:      "converts wrapped unknown secret",
			reqInput:  Map{"k": Map{"isSecretRef": true, "id": "unknown-secret"}},
			respInput: Map{"k": Map{"type": "password", "value": Map{"isSecretRef": true, "id": "unknown-secret"}}},
			want:      Map{"k": Map{"isSecretRef": true, "id": "unknown-secret"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := kbapi.PackagePolicyRequest{
				Inputs: &map[string]kbapi.PackagePolicyRequestInput{
					"input1": {
						Streams: &map[string]kbapi.PackagePolicyRequestInputStream{"stream1": {Vars: new(maps.Clone(tt.reqInput))}},
						Vars:    new(maps.Clone(tt.reqInput)),
					},
				},
				Vars: new(maps.Clone(tt.reqInput)),
			}
			resp := kbapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs: map[string]kbapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]kbapi.PackagePolicyInputStream{"stream1": {Vars: new(maps.Clone(tt.respInput))}},
						Vars:    new(maps.Clone(tt.respInput)),
					},
				},
				Vars: new(maps.Clone(tt.respInput)),
			}
			wants := kbapi.PackagePolicy{
				Inputs: map[string]kbapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]kbapi.PackagePolicyInputStream{"stream1": {Vars: new(tt.want)}},
						Vars:    &tt.want,
					},
				},
				Vars: &tt.want,
			}

			private := privateData{}
			diags := integrationpolicy.HandleReqRespSecrets(ctx, req, &resp, &private)
			require.Empty(t, diags)

			// Policy vars
			got := *resp.Vars
			want := *wants.Vars
			require.Equal(t, want, got)

			// Input vars
			got = *resp.Inputs["input1"].Vars
			want = *wants.Inputs["input1"].Vars
			require.Equal(t, want, got)

			// Stream vars
			got = *(*resp.Inputs["input1"].Streams)["stream1"].Vars
			want = *(*wants.Inputs["input1"].Streams)["stream1"].Vars
			require.Equal(t, want, got)

			privateWants := privateData{"secrets": `{}`}
			if v, ok := (*req.Vars)["k"]; ok {
				if s, ok := v.(string); ok && s == "secret" {
					privateWants = privateData{"secrets": `{"known-secret":"secret"}`}
				} else if _, ok := v.([]any); ok {
					privateWants = privateData{"secrets": `{"known-secret-1":"secret1","known-secret-2":"secret2"}`}
				}
			}
			require.Equal(t, privateWants, private)
		})
	}
}
