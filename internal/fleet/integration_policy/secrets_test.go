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
	private := privateData{"secrets": `{"known-secret":"secret"}`}

	secretRefs := &[]kbapi.PackagePolicySecretRef{
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
		{
			name:  "converts list secret",
			input: Map{"k": Map{"isSecretRef": true, "ids": []any{"known-secret"}}},
			want:  Map{"k": []any{"secret"}},
		},
		{
			name:  "converts multiple list secrets",
			input: Map{"k": Map{"isSecretRef": true, "ids": []any{"known-secret", "known-secret"}}},
			want:  Map{"k": []any{"secret", "secret"}},
		},
		{
			name:  "converts mixed list secrets",
			input: Map{"k": Map{"isSecretRef": true, "ids": []any{"known-secret", "unknown-secret"}}},
			want:  Map{"k": []any{"secret"}},
		},
		{
			name:  "converts wrapped list secret",
			input: Map{"k": Map{"type": "array", "value": Map{"isSecretRef": true, "ids": []any{"known-secret"}}}},
			want:  Map{"k": []any{"secret"}},
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
			privateWants := privateData{"secrets": `{"known-secret":"secret"}`}
			require.Equal(t, privateWants, private)
		})
	}
}

func TestHandleReqRespSecrets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	secretRefs := &[]kbapi.PackagePolicySecretRef{
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
		{
			name:      "converts list secret",
			reqInput:  Map{"k": []any{"secret1", "secret2"}},
			respInput: Map{"k": Map{"isSecretRef": true, "ids": []any{"ref1", "ref2"}}},
			want:      Map{"k": []any{"secret1", "secret2"}},
		},
		{
			name:      "converts wrapped list secret",
			reqInput:  Map{"k": []any{"secret1", "secret2"}},
			respInput: Map{"k": Map{"type": "array", "value": Map{"isSecretRef": true, "ids": []any{"ref1", "ref2"}}}},
			want:      Map{"k": []any{"secret1", "secret2"}},
		},
		{
			name:      "converts partial list secret",
			reqInput:  Map{"k": []any{"secret1"}},
			respInput: Map{"k": Map{"isSecretRef": true, "ids": []any{"ref1", "ref2"}}},
			want:      Map{"k": []any{"secret1"}},
		},
		{
			name:      "converts empty list secret",
			reqInput:  Map{"k": []any{}},
			respInput: Map{"k": Map{"isSecretRef": true, "ids": []any{"ref1"}}},
			want:      Map{"k": []any{}},
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

			// Check private data expectations based on test case
			switch tt.name {
			case "converts secret", "converts wrapped secret":
				privateWants := privateData{"secrets": `{"known-secret":"secret"}`}
				require.Equal(t, privateWants, private)
			case "converts list secret", "converts wrapped list secret":
				privateWants := privateData{"secrets": `{"ref1":"secret1","ref2":"secret2"}`}
				require.Equal(t, privateWants, private)
			case "converts partial list secret":
				privateWants := privateData{"secrets": `{"ref1":"secret1"}`}
				require.Equal(t, privateWants, private)
			case "converts empty list secret":
				privateWants := privateData{"secrets": `{}`}
				require.Equal(t, privateWants, private)
			default:
				privateWants := privateData{"secrets": `{}`}
				require.Equal(t, privateWants, private)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("handles deeply nested secret references", func(t *testing.T) {
		private := privateData{"secrets": `{"nested-ref":"nested-secret"}`}

		resp := &kbapi.PackagePolicy{
			SecretReferences: &[]kbapi.PackagePolicySecretRef{{Id: "nested-ref"}},
			Vars: &map[string]any{
				"level1": map[string]any{
					"type": "object",
					"value": map[string]any{
						"level2": map[string]any{
							"isSecretRef": true,
							"id":          "nested-ref",
						},
					},
				},
			},
		}

		diags := integration_policy.HandleRespSecrets(ctx, resp, &private)
		require.Empty(t, diags)

		expected := map[string]any{
			"level1": map[string]any{
				"level2": "nested-secret",
			},
		}
		require.Equal(t, expected, *resp.Vars)
	})

	t.Run("handles multiple input streams", func(t *testing.T) {
		private := privateData{"secrets": `{"stream1-ref":"stream1-secret","stream2-ref":"stream2-secret"}`}

		resp := &kbapi.PackagePolicy{
			SecretReferences: &[]kbapi.PackagePolicySecretRef{
				{Id: "stream1-ref"}, {Id: "stream2-ref"},
			},
			Inputs: map[string]kbapi.PackagePolicyInput{
				"input1": {
					Streams: &map[string]kbapi.PackagePolicyInputStream{
						"stream1": {
							Vars: &map[string]any{
								"secret1": map[string]any{"isSecretRef": true, "id": "stream1-ref"},
							},
						},
						"stream2": {
							Vars: &map[string]any{
								"secret2": map[string]any{"isSecretRef": true, "id": "stream2-ref"},
							},
						},
					},
				},
			},
		}

		diags := integration_policy.HandleRespSecrets(ctx, resp, &private)
		require.Empty(t, diags)

		streams := *resp.Inputs["input1"].Streams
		require.Equal(t, "stream1-secret", (*streams["stream1"].Vars)["secret1"])
		require.Equal(t, "stream2-secret", (*streams["stream2"].Vars)["secret2"])
	})

	t.Run("handles invalid JSON in private data", func(t *testing.T) {
		private := privateData{"secrets": `{"invalid": json}`}

		resp := &kbapi.PackagePolicy{
			SecretReferences: &[]kbapi.PackagePolicySecretRef{},
			Vars:             &map[string]any{},
		}

		diags := integration_policy.HandleRespSecrets(ctx, resp, &private)
		require.True(t, diags.HasError(), "Expected error diagnostics")
	})
}

func TestMigrationScenarios(t *testing.T) {
	t.Parallel()

	// Test pre-0.11.7 migration scenarios mentioned in HandleReqRespSecrets
	ctx := context.Background()

	t.Run("handles importing existing secret refs", func(t *testing.T) {
		// Simulate importing a resource that already has secret refs in request
		req := kbapi.PackagePolicyRequest{
			Vars: &map[string]any{
				"existing_secret": map[string]any{
					"isSecretRef": true,
					"id":          "existing-ref",
				},
			},
		}

		resp := &kbapi.PackagePolicy{
			SecretReferences: &[]kbapi.PackagePolicySecretRef{{Id: "new-ref"}},
			Vars: &map[string]any{
				"existing_secret": map[string]any{
					"isSecretRef": true,
					"id":          "new-ref",
				},
			},
		}

		private := privateData{}
		diags := integration_policy.HandleReqRespSecrets(ctx, req, resp, &private)
		require.Empty(t, diags)

		// Should preserve the original secret ref structure since original is also a secret ref
		expected := map[string]any{
			"existing_secret": map[string]any{
				"isSecretRef": true,
				"id":          "existing-ref",
			},
		}
		require.Equal(t, expected, *resp.Vars)
	})

	t.Run("handles migration from plain text to secret ref", func(t *testing.T) {
		// Simulate migration where plain text becomes secret ref
		req := kbapi.PackagePolicyRequest{
			Vars: &map[string]any{
				"password": "plain-text-password",
			},
		}

		resp := &kbapi.PackagePolicy{
			SecretReferences: &[]kbapi.PackagePolicySecretRef{{Id: "password-ref"}},
			Vars: &map[string]any{
				"password": map[string]any{
					"isSecretRef": true,
					"id":          "password-ref",
				},
			},
		}

		private := privateData{}
		diags := integration_policy.HandleReqRespSecrets(ctx, req, resp, &private)
		require.Empty(t, diags)

		// Should replace secret ref with original plain text
		expected := map[string]any{"password": "plain-text-password"}
		require.Equal(t, expected, *resp.Vars)

		// Should save the mapping for future use
		expectedPrivate := `{"password-ref":"plain-text-password"}`
		require.Equal(t, expectedPrivate, private["secrets"])
	})

	t.Run("handles list migration scenarios", func(t *testing.T) {
		req := kbapi.PackagePolicyRequest{
			Vars: &map[string]any{
				"hosts": []any{"host1.example.com", "host2.example.com", "host3.example.com"},
			},
		}

		resp := &kbapi.PackagePolicy{
			SecretReferences: &[]kbapi.PackagePolicySecretRef{
				{Id: "host-ref-1"}, {Id: "host-ref-2"}, {Id: "host-ref-3"},
			},
			Vars: &map[string]any{
				"hosts": map[string]any{
					"isSecretRef": true,
					"ids":         []any{"host-ref-1", "host-ref-2", "host-ref-3"},
				},
			},
		}

		private := privateData{}
		diags := integration_policy.HandleReqRespSecrets(ctx, req, resp, &private)
		require.Empty(t, diags)

		// Should replace list secret refs with original array
		expected := map[string]any{"hosts": []any{"host1.example.com", "host2.example.com", "host3.example.com"}}
		require.Equal(t, expected, *resp.Vars)

		// Should save individual mappings
		expectedPrivate := `{"host-ref-1":"host1.example.com","host-ref-2":"host2.example.com","host-ref-3":"host3.example.com"}`
		require.Equal(t, expectedPrivate, private["secrets"])
	})
}
