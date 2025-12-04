package integration_policy_test

import (
	"context"
	"fmt"
	"maps"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

type privateData map[string]string

func (p *privateData) GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics) {
	if val, ok := (*p)[key]; ok {
		return []byte(val), nil
	} else {
		return nil, nil
	}
}

func (p *privateData) SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics {
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
		t.Run(fmt.Sprintf("%s - mapped inputs", tt.name), func(t *testing.T) {
			var respInputs kbapi.PackagePolicy_Inputs
			require.NoError(t, respInputs.FromPackagePolicyMappedInputs(kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &map[string]kbapi.PackagePolicyMappedInputStream{"stream1": {Vars: utils.Pointer(maps.Clone(tt.input))}},
					Vars:    utils.Pointer(maps.Clone(tt.input)),
				},
			}))

			wantsMappedInputs := kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &map[string]kbapi.PackagePolicyMappedInputStream{"stream1": {Vars: utils.Pointer(tt.want)}},
					Vars:    &tt.want,
				},
			}
			var wantsInputs kbapi.PackagePolicy_Inputs
			require.NoError(t, wantsInputs.FromPackagePolicyMappedInputs(wantsMappedInputs))

			resp := kbapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs:           respInputs,
				Vars:             utils.Pointer(maps.Clone(tt.input)),
			}
			wants := kbapi.PackagePolicy{
				Inputs: wantsInputs,
				Vars:   &tt.want,
			}

			diags := integration_policy.HandleRespSecrets(ctx, &resp, &private)
			require.Empty(t, diags)
			// Policy vars
			got := *resp.Vars
			want := *wants.Vars
			require.Equal(t, want, got)

			// Re-extract mapped inputs from union to get the modified values
			finalMappedInputs, err := resp.Inputs.AsPackagePolicyMappedInputs()
			require.NoError(t, err)

			// Input vars
			got = *finalMappedInputs["input1"].Vars
			want = *wantsMappedInputs["input1"].Vars
			require.Equal(t, want, got)

			// Stream vars
			got = *(*finalMappedInputs["input1"].Streams)["stream1"].Vars
			want = *(*wantsMappedInputs["input1"].Streams)["stream1"].Vars
			require.Equal(t, want, got) // privateData
			privateWants := privateData{"secrets": `{"known-secret":"secret","known-secret-1":"secret1","known-secret-2":"secret2"}`}
			require.Equal(t, privateWants, private)
		})
		t.Run(fmt.Sprintf("%s - typed inputs", tt.name), func(t *testing.T) {
			var respInputs kbapi.PackagePolicy_Inputs
			require.NoError(t, respInputs.FromPackagePolicyTypedInputs(kbapi.PackagePolicyTypedInputs{
				{
					Type:    "input1",
					Streams: []kbapi.PackagePolicyTypedInputStream{{Vars: utils.Pointer(maps.Clone(tt.input))}},
					Vars:    utils.Pointer(maps.Clone(tt.input)),
				},
			}))

			wantsTypedInputs := kbapi.PackagePolicyTypedInputs{
				{
					Type:    "input1",
					Streams: []kbapi.PackagePolicyTypedInputStream{{Vars: utils.Pointer(tt.want)}},
					Vars:    &tt.want,
				},
			}
			var wantsInputs kbapi.PackagePolicy_Inputs
			require.NoError(t, wantsInputs.FromPackagePolicyTypedInputs(wantsTypedInputs))

			resp := kbapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs:           respInputs,
				Vars:             utils.Pointer(maps.Clone(tt.input)),
			}
			wants := kbapi.PackagePolicy{
				Inputs: wantsInputs,
				Vars:   &tt.want,
			}

			diags := integration_policy.HandleRespSecrets(ctx, &resp, &private)
			require.Empty(t, diags)
			// Policy vars
			got := *resp.Vars
			want := *wants.Vars
			require.Equal(t, want, got)

			// Re-extract mapped inputs from union to get the modified values
			finalTypedInputs, err := resp.Inputs.AsPackagePolicyTypedInputs()
			require.NoError(t, err)

			// Input vars
			got = *finalTypedInputs[0].Vars
			want = *wantsTypedInputs[0].Vars
			require.Equal(t, want, got)

			// Stream vars
			got = *finalTypedInputs[0].Streams[0].Vars
			want = *wantsTypedInputs[0].Streams[0].Vars
			require.Equal(t, want, got) // privateData
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
		t.Run(fmt.Sprintf("%s - mapped inputs", tt.name), func(t *testing.T) {
			// Build mapped request inputs union
			mappedReqInputs := map[string]kbapi.PackagePolicyRequestMappedInput{
				"input1": {
					Streams: &map[string]kbapi.PackagePolicyRequestMappedInputStream{"stream1": {Vars: utils.Pointer(maps.Clone(tt.reqInput))}},
					Vars:    utils.Pointer(maps.Clone(tt.reqInput)),
				},
			}
			mappedReq := kbapi.PackagePolicyRequestMappedInputs{
				Inputs: &mappedReqInputs,
				Vars:   utils.Pointer(maps.Clone(tt.reqInput)),
			}

			var req kbapi.PackagePolicyRequest
			require.NoError(t, req.FromPackagePolicyRequestMappedInputs(mappedReq))

			// Build mapped response inputs union
			var respInputs kbapi.PackagePolicy_Inputs
			require.NoError(t, respInputs.FromPackagePolicyMappedInputs(kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &map[string]kbapi.PackagePolicyMappedInputStream{"stream1": {Vars: utils.Pointer(maps.Clone(tt.respInput))}},
					Vars:    utils.Pointer(maps.Clone(tt.respInput)),
				},
			}))

			wantsMappedInputs := kbapi.PackagePolicyMappedInputs{
				"input1": {
					Streams: &map[string]kbapi.PackagePolicyMappedInputStream{"stream1": {Vars: utils.Pointer(tt.want)}},
					Vars:    &tt.want,
				},
			}
			var wantsInputs kbapi.PackagePolicy_Inputs
			require.NoError(t, wantsInputs.FromPackagePolicyMappedInputs(wantsMappedInputs))

			resp := kbapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs:           respInputs,
				Vars:             utils.Pointer(maps.Clone(tt.respInput)),
			}
			wants := kbapi.PackagePolicy{
				Inputs: wantsInputs,
				Vars:   &tt.want,
			}

			private := privateData{}
			diags := integration_policy.HandleReqRespSecrets(ctx, req, &resp, &private)
			require.Empty(t, diags)

			// Policy vars
			got := *resp.Vars
			want := *wants.Vars
			require.Equal(t, want, got)

			// Input vars
			finalRespMappedInputs, err := resp.Inputs.AsPackagePolicyMappedInputs()
			require.NoError(t, err)
			got = *finalRespMappedInputs["input1"].Vars
			want = *wantsMappedInputs["input1"].Vars
			require.Equal(t, want, got)

			// Stream vars
			got = *(*finalRespMappedInputs["input1"].Streams)["stream1"].Vars
			want = *(*wantsMappedInputs["input1"].Streams)["stream1"].Vars
			require.Equal(t, want, got)

			privateWants := privateData{"secrets": `{}`}
			if v, ok := tt.reqInput["k"]; ok {
				if s, ok := v.(string); ok && s == "secret" {
					privateWants = privateData{"secrets": `{"known-secret":"secret"}`}
				} else if _, ok := v.([]any); ok {
					privateWants = privateData{"secrets": `{"known-secret-1":"secret1","known-secret-2":"secret2"}`}
				}
			}
			require.Equal(t, privateWants, private)
		})
		t.Run(fmt.Sprintf("%s - typed inputs", tt.name), func(t *testing.T) {
			// Build typed request inputs union
			typedReqInputs := []kbapi.PackagePolicyRequestTypedInput{
				{
					Type:    "input1",
					Streams: &[]kbapi.PackagePolicyRequestTypedInputStream{{Vars: utils.Pointer(maps.Clone(tt.reqInput))}},
					Vars:    utils.Pointer(maps.Clone(tt.reqInput)),
				},
			}
			typedReq := kbapi.PackagePolicyRequestTypedInputs{
				Inputs: &typedReqInputs,
				Vars:   utils.Pointer(maps.Clone(tt.reqInput)),
			}

			var req kbapi.PackagePolicyRequest
			require.NoError(t, req.FromPackagePolicyRequestTypedInputs(typedReq))

			// Build typed response inputs union
			var respInputs kbapi.PackagePolicy_Inputs
			require.NoError(t, respInputs.FromPackagePolicyTypedInputs(kbapi.PackagePolicyTypedInputs{
				{
					Type:    "input1",
					Streams: []kbapi.PackagePolicyTypedInputStream{{Vars: utils.Pointer(maps.Clone(tt.respInput))}},
					Vars:    utils.Pointer(maps.Clone(tt.respInput)),
				},
			}))

			wantsTypedInputs := kbapi.PackagePolicyTypedInputs{
				{
					Type:    "input1",
					Streams: []kbapi.PackagePolicyTypedInputStream{{Vars: utils.Pointer(tt.want)}},
					Vars:    &tt.want,
				},
			}
			var wantsInputs kbapi.PackagePolicy_Inputs
			require.NoError(t, wantsInputs.FromPackagePolicyTypedInputs(wantsTypedInputs))

			resp := kbapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs:           respInputs,
				Vars:             utils.Pointer(maps.Clone(tt.respInput)),
			}
			wants := kbapi.PackagePolicy{
				Inputs: wantsInputs,
				Vars:   &tt.want,
			}

			private := privateData{}
			diags := integration_policy.HandleReqRespSecrets(ctx, req, &resp, &private)
			require.Empty(t, diags)

			// Policy vars
			got := *resp.Vars
			want := *wants.Vars
			require.Equal(t, want, got)

			// Input vars
			finalRespTypedInputs, err := resp.Inputs.AsPackagePolicyTypedInputs()
			require.NoError(t, err)
			got = *finalRespTypedInputs[0].Vars
			want = *wantsTypedInputs[0].Vars
			require.Equal(t, want, got)

			// Stream vars
			got = *(finalRespTypedInputs[0].Streams)[0].Vars
			want = *(wantsTypedInputs[0].Streams)[0].Vars
			require.Equal(t, want, got)

			privateWants := privateData{"secrets": `{}`}
			if v, ok := tt.reqInput["k"]; ok {
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
