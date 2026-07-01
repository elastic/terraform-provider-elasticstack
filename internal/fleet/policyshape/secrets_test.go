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

package policyshape_test

import (
	"context"
	"encoding/json"
	"maps"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
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
func buildPackagePolicyWithMappedInputs(t *testing.T, secretRefs *[]kbapi.PackagePolicySecretRef, inputs kbapi.PackagePolicyMappedInputs, vars *map[string]any) kbapi.PackagePolicy {
	t.Helper()
	p := kbapi.PackagePolicy{
		SecretReferences: secretRefs,
	}
	if vars != nil {
		p.Vars = policyshape.VarsMapToUnionWrapper[kbapi.KibanaHTTPAPIsPackagePolicyResponse_Vars](*vars)
	}
	require.NoError(t, p.Inputs.FromPackagePolicyMappedInputs(inputs))
	return p
}

// buildPackagePolicyRequestMapped creates a PackagePolicyRequest from mapped inputs.
func buildPackagePolicyRequestMapped(t *testing.T, inputs *map[string]kbapi.PackagePolicyRequestMappedInput, vars *map[string]any) kbapi.PackagePolicyRequest {
	t.Helper()
	mapped := kbapi.PackagePolicyRequestMappedInputs{
		Name:    "test",
		Package: kbapi.PackagePolicyRequestPackage{Name: "test", Version: "1.0.0"},
		Inputs:  inputs,
	}
	if vars != nil {
		mapped.Vars = policyshape.VarsMapToTypedMap[kbapi.KibanaHTTPAPIsSimplifiedCreatePackagePolicyRequest_Vars_AdditionalProperties](*vars)
	}
	var req kbapi.PackagePolicyRequest
	require.NoError(t, req.FromPackagePolicyRequestMappedInputs(mapped))
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
				"stream1": {Vars: policyshape.VarsMapToTypedMap[kbapi.PackagePolicyMappedInputStream_Vars_AdditionalProperties](maps.Clone(tt.input))},
			}
			mappedInputs := kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &inputStreams,
					Vars:    policyshape.VarsMapToTypedMap[kbapi.PackagePolicyMappedInput_Vars_AdditionalProperties](maps.Clone(tt.input)),
				},
			}
			cloned := maps.Clone(tt.input)
			resp := buildPackagePolicyWithMappedInputs(t, secretRefs, mappedInputs, &cloned)

			wantInputStreams := map[string]kbapi.PackagePolicyMappedInputStream{
				"stream1": {Vars: policyshape.VarsMapToTypedMap[kbapi.PackagePolicyMappedInputStream_Vars_AdditionalProperties](tt.want)},
			}
			wantMappedInputs := kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &wantInputStreams,
					Vars:    policyshape.VarsMapToTypedMap[kbapi.PackagePolicyMappedInput_Vars_AdditionalProperties](tt.want),
				},
			}
			wants := buildPackagePolicyWithMappedInputs(t, nil, wantMappedInputs, &tt.want)

			diags := policyshape.HandleRespSecrets(ctx, &resp, &private)
			require.Empty(t, diags)
			// The typed Vars wrappers differ across endpoints (policy / input /
			// stream) so use a JSON-equal comparison to avoid threading three
			// different concrete types through the assertions.
			requireVarsEqual := func(want, got any) {
				wantJSON, err := json.Marshal(want)
				require.NoError(t, err)
				gotJSON, err := json.Marshal(got)
				require.NoError(t, err)
				require.JSONEq(t, string(wantJSON), string(gotJSON))
			}
			// Policy vars
			requireVarsEqual(wants.Vars, resp.Vars)

			// Extract inputs to check
			respMapped, err := resp.Inputs.AsPackagePolicyMappedInputs()
			require.NoError(t, err)
			wantsMapped, err := wants.Inputs.AsPackagePolicyMappedInputs()
			require.NoError(t, err)

			// Input vars
			requireVarsEqual(wantsMapped["input1"].Vars, respMapped["input1"].Vars)

			// Stream vars
			requireVarsEqual((*wantsMapped["input1"].Streams)["stream1"].Vars, (*respMapped["input1"].Streams)["stream1"].Vars)

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
				"stream1": {Vars: policyshape.VarsMapToTypedMap[kbapi.PackagePolicyRequestMappedInputStream_Vars_AdditionalProperties](maps.Clone(tt.reqInput))},
			}
			reqInputs := &map[string]kbapi.PackagePolicyRequestMappedInput{
				"input1": {
					Streams: &reqInputStreams,
					Vars:    policyshape.VarsMapToTypedMap[kbapi.PackagePolicyRequestMappedInput_Vars_AdditionalProperties](maps.Clone(tt.reqInput)),
				},
			}
			reqVars := maps.Clone(tt.reqInput)
			req := buildPackagePolicyRequestMapped(t, reqInputs, &reqVars)

			respInputStreams := map[string]kbapi.PackagePolicyMappedInputStream{
				"stream1": {Vars: policyshape.VarsMapToTypedMap[kbapi.PackagePolicyMappedInputStream_Vars_AdditionalProperties](maps.Clone(tt.respInput))},
			}
			respMappedInputs := kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &respInputStreams,
					Vars:    policyshape.VarsMapToTypedMap[kbapi.PackagePolicyMappedInput_Vars_AdditionalProperties](maps.Clone(tt.respInput)),
				},
			}
			respVars := maps.Clone(tt.respInput)
			resp := buildPackagePolicyWithMappedInputs(t, secretRefs, respMappedInputs, &respVars)

			wantInputStreams := map[string]kbapi.PackagePolicyMappedInputStream{
				"stream1": {Vars: policyshape.VarsMapToTypedMap[kbapi.PackagePolicyMappedInputStream_Vars_AdditionalProperties](tt.want)},
			}
			wantMappedInputs := kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &wantInputStreams,
					Vars:    policyshape.VarsMapToTypedMap[kbapi.PackagePolicyMappedInput_Vars_AdditionalProperties](tt.want),
				},
			}
			wants := buildPackagePolicyWithMappedInputs(t, nil, wantMappedInputs, &tt.want)

			private := privateData{}
			diags := policyshape.HandleReqRespSecrets(ctx, req, &resp, &private)
			require.Empty(t, diags)

			// The typed Vars wrappers differ across endpoints (policy / input /
			// stream) so use a JSON-equal comparison.
			requireVarsEqual := func(want, got any) {
				wantJSON, err := json.Marshal(want)
				require.NoError(t, err)
				gotJSON, err := json.Marshal(got)
				require.NoError(t, err)
				require.JSONEq(t, string(wantJSON), string(gotJSON))
			}

			// Policy vars
			requireVarsEqual(wants.Vars, resp.Vars)

			// Extract inputs to check
			respMapped, err := resp.Inputs.AsPackagePolicyMappedInputs()
			require.NoError(t, err)
			wantsMapped, err := wants.Inputs.AsPackagePolicyMappedInputs()
			require.NoError(t, err)

			// Input vars
			requireVarsEqual(wantsMapped["input1"].Vars, respMapped["input1"].Vars)

			// Stream vars
			requireVarsEqual((*wantsMapped["input1"].Streams)["stream1"].Vars, (*respMapped["input1"].Streams)["stream1"].Vars)

			// Check private data based on req vars (round-trip the typed
			// wrapper through a generic map for inspection).
			reqMapped, err := req.AsPackagePolicyRequestMappedInputs()
			require.NoError(t, err)
			privateWants := privateData{"secrets": `{}`}
			reqVarsMap := map[string]any{}
			if reqMapped.Vars != nil {
				bytes, err := json.Marshal(*reqMapped.Vars)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(bytes, &reqVarsMap))
			}
			if v, ok := reqVarsMap["k"]; ok {
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
