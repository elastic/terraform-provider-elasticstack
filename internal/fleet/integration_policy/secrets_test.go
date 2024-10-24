package integration_policy_test

import (
	"context"
	"maps"
	"testing"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
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
	private := privateData{"secrets": `{"known-secret":"secret"}`}

	secretRefs := &[]fleetapi.PackagePolicySecretRef{
		{Id: "known-secret"},
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
			resp := fleetapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs: map[string]fleetapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]fleetapi.PackagePolicyInputStream{"stream1": fleetapi.PackagePolicyInputStream{Vars: utils.Pointer(maps.Clone(tt.input))}},
						Vars:    utils.Pointer(maps.Clone(tt.input)),
					},
				},
				Vars: utils.Pointer(maps.Clone(tt.input)),
			}
			wants := fleetapi.PackagePolicy{
				Inputs: map[string]fleetapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]fleetapi.PackagePolicyInputStream{"stream1": fleetapi.PackagePolicyInputStream{Vars: utils.Pointer(tt.want)}},
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
			privateWants := privateData{"secrets": `{"known-secret":"secret"}`}
			require.Equal(t, privateWants, private)
		})
	}
}

func TestHandleReqRespSecrets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	secretRefs := &[]fleetapi.PackagePolicySecretRef{
		{Id: "known-secret"},
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
			req := fleetapi.PackagePolicyRequest{
				Inputs: &map[string]fleetapi.PackagePolicyRequestInput{
					"input1": {
						Streams: &map[string]fleetapi.PackagePolicyRequestInputStream{"stream1": {Vars: utils.Pointer(maps.Clone(tt.reqInput))}},
						Vars:    utils.Pointer(maps.Clone(tt.reqInput)),
					},
				},
				Vars: utils.Pointer(maps.Clone(tt.reqInput)),
			}
			resp := fleetapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs: map[string]fleetapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]fleetapi.PackagePolicyInputStream{"stream1": fleetapi.PackagePolicyInputStream{Vars: utils.Pointer(maps.Clone(tt.respInput))}},
						Vars:    utils.Pointer(maps.Clone(tt.respInput)),
					},
				},
				Vars: utils.Pointer(maps.Clone(tt.respInput)),
			}
			wants := fleetapi.PackagePolicy{
				Inputs: map[string]fleetapi.PackagePolicyInput{
					"input1": {
						Streams: &map[string]fleetapi.PackagePolicyInputStream{"stream1": fleetapi.PackagePolicyInputStream{Vars: utils.Pointer(tt.want)}},
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

			if v, ok := (*req.Vars)["k"]; ok && v == "secret" {
				privateWants := privateData{"secrets": `{"known-secret":"secret"}`}
				require.Equal(t, privateWants, private)
			} else {
				privateWants := privateData{"secrets": `{}`}
				require.Equal(t, privateWants, private)
			}
		})
	}
}
