package integration_policy_test

import (
	"context"
	"maps"
	"reflect"
	"testing"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

	secretRefs := &[]struct {
		Id *string `json:"id,omitempty"`
	}{
		{Id: utils.Pointer("known-secret")},
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
						Streams: &Map{
							"stream1": Map{
								"vars": maps.Clone(tt.input),
							},
						},
						Vars: utils.Pointer(maps.Clone(tt.input)),
					},
				},
				Vars: utils.Pointer(maps.Clone(tt.input)),
			}
			wants := fleetapi.PackagePolicy{
				Inputs: map[string]fleetapi.PackagePolicyInput{
					"input1": {
						Streams: &Map{
							"stream1": Map{
								"vars": tt.want,
							},
						},
						Vars: &tt.want,
					},
				},
				Vars: &tt.want,
			}

			diags := integration_policy.HandleRespSecrets(ctx, &resp, &private)
			// Policy vars
			if !reflect.DeepEqual(
				*resp.Vars,
				*wants.Vars,
			) {
				t.Errorf("HandleRespSecrets() policy-vars = %#v, want %#v",
					*resp.Vars,
					*wants.Vars,
				)
			}
			// Input vars
			if !reflect.DeepEqual(
				*resp.Inputs["input1"].Vars,
				*wants.Inputs["input1"].Vars,
			) {
				t.Errorf("HandleRespSecrets() input-vars = %#v, want %#v",
					*resp.Inputs["input1"].Vars,
					*wants.Inputs["input1"].Vars,
				)
			}
			// Stream vars
			if !reflect.DeepEqual(
				(*resp.Inputs["input1"].Streams)["stream1"].(Map)["vars"],
				(*wants.Inputs["input1"].Streams)["stream1"].(Map)["vars"],
			) {
				t.Errorf("HandleRespSecrets() stream-vars = %#v, want %#v",
					(*resp.Inputs["input1"].Streams)["stream1"].(Map)["vars"],
					(*wants.Inputs["input1"].Streams)["stream1"].(Map)["vars"],
				)
			}
			for _, d := range diags.Errors() {
				t.Errorf("HandleRespSecrets() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
			// privateData
			privateWants := privateData{"secrets": `{"known-secret":"secret"}`}
			if !reflect.DeepEqual(private, privateWants) {
				t.Errorf("HandleRespSecrets() privateData = %#v, want %#v", private, privateWants)
			}
		})
	}
}

func TestHandleReqRespSecrets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	secretRefs := &[]struct {
		Id *string `json:"id,omitempty"`
	}{
		{Id: utils.Pointer("known-secret")},
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
						Streams: &map[string]fleetapi.PackagePolicyRequestInputStream{
							"stream1": {
								Vars: utils.Pointer(maps.Clone(tt.reqInput)),
							},
						},
						Vars: utils.Pointer(maps.Clone(tt.reqInput)),
					},
				},
				Vars: utils.Pointer(maps.Clone(tt.reqInput)),
			}
			resp := fleetapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs: map[string]fleetapi.PackagePolicyInput{
					"input1": {
						Streams: &Map{
							"stream1": Map{
								"vars": maps.Clone(tt.respInput),
							},
						},
						Vars: utils.Pointer(maps.Clone(tt.respInput)),
					},
				},
				Vars: utils.Pointer(maps.Clone(tt.respInput)),
			}
			wants := fleetapi.PackagePolicy{
				SecretReferences: secretRefs,
				Inputs: map[string]fleetapi.PackagePolicyInput{
					"input1": {
						Streams: &Map{
							"stream1": Map{
								"vars": tt.want,
							},
						},
						Vars: &tt.want,
					},
				},
				Vars: &tt.want,
			}

			private := privateData{}
			diags := integration_policy.HandleReqRespSecrets(ctx, req, &resp, &private)
			// Policy vars
			if !reflect.DeepEqual(
				*resp.Vars,
				*wants.Vars,
			) {
				t.Errorf("HandleReqRespSecrets() policy-vars = %#v, want %#v",
					*resp.Vars,
					*wants.Vars,
				)
			}
			// Input vars
			if !reflect.DeepEqual(
				*resp.Inputs["input1"].Vars,
				*wants.Inputs["input1"].Vars,
			) {
				t.Errorf("HandleReqRespSecrets() input-vars = %#v, want %#v",
					*resp.Inputs["input1"].Vars,
					*wants.Inputs["input1"].Vars,
				)
			}
			// Stream vars
			if !reflect.DeepEqual(
				(*resp.Inputs["input1"].Streams)["stream1"].(Map)["vars"],
				(*wants.Inputs["input1"].Streams)["stream1"].(Map)["vars"],
			) {
				t.Errorf("HandleReqRespSecrets() stream-vars = %#v, want %#v",
					(*resp.Inputs["input1"].Streams)["stream1"].(Map)["vars"],
					(*wants.Inputs["input1"].Streams)["stream1"].(Map)["vars"],
				)
			}
			for _, d := range diags.Errors() {
				t.Errorf("HandleReqRespSecrets() diagnostic: %s: %s", d.Summary(), d.Detail())
			}

			if v, ok := (*req.Vars)["k"]; ok && v == "secret" {
				privateWants := privateData{"secrets": `{"known-secret":"secret"}`}
				if !reflect.DeepEqual(private, privateWants) {
					t.Errorf("HandleReqRespSecrets() privateData = %#v, want %#v", private, privateWants)
				}
			} else {
				privateWants := privateData{"secrets": `{}`}
				if !reflect.DeepEqual(private, privateWants) {
					t.Errorf("HandleReqRespSecrets() privateData = %#v, want %#v", private, privateWants)
				}
			}
		})
	}
}
