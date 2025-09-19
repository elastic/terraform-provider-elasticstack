package integration_policy_test

import (
	"context"
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
		t.Run(tt.name, func(t *testing.T) {
			resp := kbapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs: map[string]kbapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]kbapi.PackagePolicyInputStream{"stream1": {Vars: utils.Pointer(maps.Clone(tt.input))}},
						Vars:    utils.Pointer(maps.Clone(tt.input)),
					},
				},
				Vars: utils.Pointer(maps.Clone(tt.input)),
			}
			wants := kbapi.PackagePolicy{
				Inputs: map[string]kbapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]kbapi.PackagePolicyInputStream{"stream1": {Vars: utils.Pointer(tt.want)}},
						Vars:    &tt.want,
					},
				},
				Vars: &tt.want,
			}

			diags := integration_policy.HandleRespSecrets(ctx, &resp, &private)
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
						Streams: &map[string]kbapi.PackagePolicyRequestInputStream{"stream1": {Vars: utils.Pointer(maps.Clone(tt.reqInput))}},
						Vars:    utils.Pointer(maps.Clone(tt.reqInput)),
					},
				},
				Vars: utils.Pointer(maps.Clone(tt.reqInput)),
			}
			resp := kbapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs: map[string]kbapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]kbapi.PackagePolicyInputStream{"stream1": {Vars: utils.Pointer(maps.Clone(tt.respInput))}},
						Vars:    utils.Pointer(maps.Clone(tt.respInput)),
					},
				},
				Vars: utils.Pointer(maps.Clone(tt.respInput)),
			}
			wants := kbapi.PackagePolicy{
				Inputs: map[string]kbapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]kbapi.PackagePolicyInputStream{"stream1": {Vars: utils.Pointer(tt.want)}},
						Vars:    &tt.want,
					},
				},
				Vars: &tt.want,
			}

			private := privateData{}
			diags := integration_policy.HandleReqRespSecrets(ctx, req, &resp, &private)
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
