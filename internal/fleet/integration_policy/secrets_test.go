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

// buildPackagePolicyWithMappedInputs creates a PackagePolicy with mapped inputs
// for use in tests. The Inputs field is set via the union type.
func buildPackagePolicyWithMappedInputs(secretRefs *[]kbapi.PackagePolicySecretRef, inputs kbapi.PackagePolicyMappedInputs, vars *map[string]any) kbapi.PackagePolicy {
	p := kbapi.PackagePolicy{
		SecretReferences: secretRefs,
		Vars:             vars,
	}
	if err := p.Inputs.FromPackagePolicyMappedInputs(inputs); err != nil {
		panic("failed to set mapped inputs: " + err.Error())
	}
	return p
}

// buildPackagePolicyRequestMapped creates a PackagePolicyRequest from mapped inputs.
func buildPackagePolicyRequestMapped(inputs *map[string]kbapi.PackagePolicyRequestMappedInput, vars *map[string]any) kbapi.PackagePolicyRequest {
	mapped := kbapi.PackagePolicyRequestMappedInputs{
		Name:    "test",
		Package: kbapi.PackagePolicyRequestPackage{Name: "test", Version: "1.0.0"},
		Inputs:  inputs,
		Vars:    vars,
	}
	var req kbapi.PackagePolicyRequest
	if err := req.FromPackagePolicyRequestMappedInputs(mapped); err != nil {
		panic("failed to build mapped request: " + err.Error())
	}
	return req
}

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
			inputStreams := map[string]kbapi.PackagePolicyMappedInputStream{
				"stream1": {Vars: new(maps.Clone(tt.input))},
			}
			mappedInputs := kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &inputStreams,
					Vars:    new(maps.Clone(tt.input)),
				},
			}
			resp := buildPackagePolicyWithMappedInputs(secretRefs, mappedInputs, new(maps.Clone(tt.input)))

			wantInputStreams := map[string]kbapi.PackagePolicyMappedInputStream{
				"stream1": {Vars: new(tt.want)},
			}
			wantMappedInputs := kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &wantInputStreams,
					Vars:    &tt.want,
				},
			}
			wants := buildPackagePolicyWithMappedInputs(nil, wantMappedInputs, &tt.want)

			diags := integrationpolicy.HandleRespSecrets(ctx, &resp, &private)
			require.Empty(t, diags)
			// Policy vars
			got := *resp.Vars
			want := *wants.Vars
			require.Equal(t, want, got)

			// Extract inputs to check
			respMapped, err := resp.Inputs.AsPackagePolicyMappedInputs()
			require.NoError(t, err)
			wantsMapped, err := wants.Inputs.AsPackagePolicyMappedInputs()
			require.NoError(t, err)

			// Input vars
			got = *respMapped["input1"].Vars
			want = *wantsMapped["input1"].Vars
			require.Equal(t, want, got)

			// Stream vars
			got = *(*respMapped["input1"].Streams)["stream1"].Vars
			want = *(*wantsMapped["input1"].Streams)["stream1"].Vars
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
			reqInputStreams := map[string]kbapi.PackagePolicyRequestMappedInputStream{
				"stream1": {Vars: new(maps.Clone(tt.reqInput))},
			}
			reqInputs := &map[string]kbapi.PackagePolicyRequestMappedInput{
				"input1": {
					Streams: &reqInputStreams,
					Vars:    new(maps.Clone(tt.reqInput)),
				},
			}
			req := buildPackagePolicyRequestMapped(reqInputs, new(maps.Clone(tt.reqInput)))

			respInputStreams := map[string]kbapi.PackagePolicyMappedInputStream{
				"stream1": {Vars: new(maps.Clone(tt.respInput))},
			}
			respMappedInputs := kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &respInputStreams,
					Vars:    new(maps.Clone(tt.respInput)),
				},
			}
			resp := buildPackagePolicyWithMappedInputs(secretRefs, respMappedInputs, new(maps.Clone(tt.respInput)))

			wantInputStreams := map[string]kbapi.PackagePolicyMappedInputStream{
				"stream1": {Vars: new(tt.want)},
			}
			wantMappedInputs := kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &wantInputStreams,
					Vars:    &tt.want,
				},
			}
			wants := buildPackagePolicyWithMappedInputs(nil, wantMappedInputs, &tt.want)

			private := privateData{}
			diags := integrationpolicy.HandleReqRespSecrets(ctx, req, &resp, &private)
			require.Empty(t, diags)

			// Policy vars
			got := *resp.Vars
			want := *wants.Vars
			require.Equal(t, want, got)

			// Extract inputs to check
			respMapped, err := resp.Inputs.AsPackagePolicyMappedInputs()
			require.NoError(t, err)
			wantsMapped, err := wants.Inputs.AsPackagePolicyMappedInputs()
			require.NoError(t, err)

			// Input vars
			got = *respMapped["input1"].Vars
			want = *wantsMapped["input1"].Vars
			require.Equal(t, want, got)

			// Stream vars
			got = *(*respMapped["input1"].Streams)["stream1"].Vars
			want = *(*wantsMapped["input1"].Streams)["stream1"].Vars
			require.Equal(t, want, got)

			// Check private data based on req vars
			reqMapped, err := req.AsPackagePolicyRequestMappedInputs()
			require.NoError(t, err)
			privateWants := privateData{"secrets": `{}`}
			if reqMapped.Vars != nil {
				if v, ok := (*reqMapped.Vars)["k"]; ok {
					if s, ok := v.(string); ok && s == "secret" {
						privateWants = privateData{"secrets": `{"known-secret":"secret"}`}
					} else if _, ok := v.([]any); ok {
						privateWants = privateData{"secrets": `{"known-secret-1":"secret1","known-secret-2":"secret2"}`}
					}
				}
			}
			require.Equal(t, privateWants, private)
		})
	}
}
