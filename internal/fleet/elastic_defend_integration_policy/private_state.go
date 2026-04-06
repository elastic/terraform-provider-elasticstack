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

package elasticdefendintegrationpolicy

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const privateStateKey = "defend_private_state"

// privateStateStorage is the interface for reading/writing private state,
// matching both resource.CreateResponse.Private, UpdateResponse.Private, etc.
type privateStateStorage interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

// savePrivateState encodes the given defendPrivateState into provider private
// state so it survives between Terraform operations.
func savePrivateState(ctx context.Context, storage privateStateStorage, ps defendPrivateState) diag.Diagnostics {
	data, err := json.Marshal(ps)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to encode Defend private state",
				err.Error(),
			),
		}
	}
	return storage.SetKey(ctx, privateStateKey, data)
}

// loadPrivateState decodes the defendPrivateState from provider private state.
// Returns a zero value if no private state has been saved yet.
func loadPrivateState(ctx context.Context, storage privateStateStorage) (defendPrivateState, diag.Diagnostics) {
	data, diags := storage.GetKey(ctx, privateStateKey)
	if diags.HasError() {
		return defendPrivateState{}, diags
	}
	if len(data) == 0 {
		return defendPrivateState{}, nil
	}
	var ps defendPrivateState
	if err := json.Unmarshal(data, &ps); err != nil {
		return defendPrivateState{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to decode Defend private state",
				err.Error(),
			),
		}
	}
	return ps, nil
}

// extractPrivateStateFromResponse extracts the opaque server-managed Defend
// payloads from an API response and builds a defendPrivateState ready to be
// persisted. It reads the artifact_manifest from the endpoint input config and
// the top-level version token.
func extractPrivateStateFromResponse(policy *kbapi.PackagePolicy) defendPrivateState {
	ps := defendPrivateState{}

	if policy.Version != nil {
		ps.Version = *policy.Version
	}

	// Extract artifact_manifest from the endpoint input's config.
	// The Inputs field is a union type; extract as typed inputs (non-simplified format).
	typedInputs, err := policy.Inputs.AsPackagePolicyTypedInputs()
	if err == nil {
		ps.ArtifactManifest = extractArtifactManifest(typedInputs)
	}

	return ps
}

// extractArtifactManifest returns the artifact_manifest config value from the
// first endpoint input in the policy, or nil if not found.
func extractArtifactManifest(inputs kbapi.PackagePolicyTypedInputs) map[string]any {
	for _, input := range inputs {
		if input.Type != endpointInputType {
			continue
		}
		if input.Config == nil {
			return nil
		}
		if amEntry, ok := (*input.Config)["artifact_manifest"]; ok {
			if manifest, ok := amEntry.Value.(map[string]any); ok {
				return manifest
			}
		}
		return nil
	}
	return nil
}
