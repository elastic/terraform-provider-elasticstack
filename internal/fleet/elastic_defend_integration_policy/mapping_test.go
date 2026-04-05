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

package elasticdefendintegrationpolicy_test

import (
	"context"
	"testing"

	edip "github.com/elastic/terraform-provider-elasticstack/internal/fleet/elastic_defend_integration_policy"
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const testArtifactManifest = "WzEyMywxXQ=="

// TestPopulateModelFromAPIValidatesPackageName tests that populateModelFromAPI
// returns an error when the package name is not "endpoint" (REQ-005).
func TestPopulateModelFromAPIValidatesPackageName(t *testing.T) {
	ctx := context.Background()

	policy := &kbapi.DefendPackagePolicy{
		Id:      "policy-123",
		Name:    "wrong-package-policy",
		Enabled: true,
		Package: &struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    "not-endpoint",
			Version: "1.0.0",
		},
		Inputs: []kbapi.DefendPackagePolicyInput{},
	}

	model := &edip.ElasticDefendIntegrationPolicyModel{}
	diags := edip.PopulateModelFromAPI(ctx, model, policy)

	if !diags.HasError() {
		t.Error("expected error when package name is not 'endpoint', got no error")
	}
}

// TestPopulateModelFromAPIEndpointPackage tests that populateModelFromAPI
// succeeds and maps basic fields correctly for an endpoint package policy.
func TestPopulateModelFromAPIEndpointPackage(t *testing.T) {
	ctx := context.Background()

	agentPolicyID := "agent-policy-abc"
	namespace := "default"
	version := testArtifactManifest

	policy := &kbapi.DefendPackagePolicy{
		Id:       "policy-123",
		Name:     "my-endpoint-policy",
		Enabled:  true,
		Namespace: &namespace,
		PolicyId: &agentPolicyID,
		Version:  &version,
		Package: &struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    "endpoint",
			Version: "8.14.0",
		},
		Inputs: []kbapi.DefendPackagePolicyInput{
			{
				Type:    "endpoint",
				Enabled: true,
				Config: map[string]any{
					"integration_config": map[string]any{
						"value": map[string]any{
							"endpointConfig": map[string]any{
								"preset": "NGAv1",
							},
						},
					},
				},
			},
		},
	}

	model := &edip.ElasticDefendIntegrationPolicyModel{}
	diags := edip.PopulateModelFromAPI(ctx, model, policy)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if model.ID.ValueString() != "policy-123" {
		t.Errorf("expected ID=%q, got %q", "policy-123", model.ID.ValueString())
	}

	if model.PolicyID.ValueString() != "policy-123" {
		t.Errorf("expected PolicyID=%q, got %q", "policy-123", model.PolicyID.ValueString())
	}

	if model.Name.ValueString() != "my-endpoint-policy" {
		t.Errorf("expected Name=%q, got %q", "my-endpoint-policy", model.Name.ValueString())
	}

	if model.Enabled.ValueBool() != true {
		t.Errorf("expected Enabled=true, got %v", model.Enabled.ValueBool())
	}

	if model.IntegrationVersion.ValueString() != "8.14.0" {
		t.Errorf("expected IntegrationVersion=%q, got %q", "8.14.0", model.IntegrationVersion.ValueString())
	}

	if model.Preset.ValueString() != "NGAv1" {
		t.Errorf("expected Preset=%q, got %q", "NGAv1", model.Preset.ValueString())
	}
}

// TestPopulateModelFromAPINilPolicy tests that populateModelFromAPI handles nil
// gracefully.
func TestPopulateModelFromAPINilPolicy(t *testing.T) {
	ctx := context.Background()
	model := &edip.ElasticDefendIntegrationPolicyModel{}
	diags := edip.PopulateModelFromAPI(ctx, model, nil)
	if diags.HasError() {
		t.Errorf("expected no error for nil policy, got %v", diags)
	}
}

// TestBuildBootstrapRequest tests that the bootstrap request has the correct
// input type and preset path (REQ-008).
func TestBuildBootstrapRequest(t *testing.T) {
	model := &edip.ElasticDefendIntegrationPolicyModel{
		Name:               types.StringValue("my-endpoint"),
		Namespace:          types.StringValue("default"),
		AgentPolicyID:      types.StringValue("agent-123"),
		IntegrationVersion: types.StringValue("8.14.0"),
		Preset:             types.StringValue("NGAv1"),
	}

	req := edip.BuildBootstrapRequest(model)

	if req.Package.Name != "endpoint" {
		t.Errorf("expected package name %q, got %q", "endpoint", req.Package.Name)
	}

	if len(req.Inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(req.Inputs))
	}

	input := req.Inputs[0]
	if input.Type != "ENDPOINT_INTEGRATION_CONFIG" {
		t.Errorf("expected input type %q, got %q", "ENDPOINT_INTEGRATION_CONFIG", input.Type)
	}

	if !input.Enabled {
		t.Error("expected input enabled=true")
	}

	if input.Streams == nil {
		t.Error("expected input streams to be non-nil (empty list)")
	}

	// Verify preset is at config._config.value.endpointConfig.preset
	config, ok := input.Config["_config"]
	if !ok {
		t.Fatal("expected _config in bootstrap input config")
	}

	configMap, ok := config.(map[string]any)
	if !ok {
		t.Fatal("expected _config to be a map")
	}

	valueMap, ok := configMap["value"].(map[string]any)
	if !ok {
		t.Fatal("expected _config.value to be a map")
	}

	ecMap, ok := valueMap["endpointConfig"].(map[string]any)
	if !ok {
		t.Fatal("expected _config.value.endpointConfig to be a map")
	}

	if ecMap["preset"] != "NGAv1" {
		t.Errorf("expected preset=%q, got %v", "NGAv1", ecMap["preset"])
	}
}

// TestBuildBootstrapRequestNullPreset tests that a null preset omits the _config
// key from the bootstrap input config entirely, rather than sending an empty string.
func TestBuildBootstrapRequestNullPreset(t *testing.T) {
	model := &edip.ElasticDefendIntegrationPolicyModel{
		Name:               types.StringValue("my-endpoint"),
		Namespace:          types.StringValue("default"),
		AgentPolicyID:      types.StringValue("agent-123"),
		IntegrationVersion: types.StringValue("8.14.0"),
		Preset:             types.StringNull(),
	}

	req := edip.BuildBootstrapRequest(model)

	if len(req.Inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(req.Inputs))
	}

	if _, ok := req.Inputs[0].Config["_config"]; ok {
		t.Error("expected _config to be absent from bootstrap input config when preset is null")
	}
}

// TestExtractPrivateStateFromResponseNonEndpointFirst tests that
// extractPrivateStateFromResponse skips non-endpoint inputs and still finds
// the artifact_manifest from the endpoint input.
func TestExtractPrivateStateFromResponseNonEndpointFirst(t *testing.T) {
	version := testArtifactManifest
	policy := &kbapi.DefendPackagePolicy{
		Id:      "policy-123",
		Name:    "test",
		Enabled: true,
		Version: &version,
		Inputs: []kbapi.DefendPackagePolicyInput{
			{
				Type:    "some-other-type",
				Enabled: true,
				Config: map[string]any{
					"artifact_manifest": map[string]any{
						"artifacts": map[string]any{
							"wrong": "should not be picked up",
						},
					},
				},
			},
			{
				Type:    "endpoint",
				Enabled: true,
				Config: map[string]any{
					"artifact_manifest": map[string]any{
						"artifacts": map[string]any{
							"endpoint-exceptionlist-macos-v1": map[string]any{
								"sha256": "abc123",
							},
						},
					},
				},
			},
		},
	}

	ps := edip.ExtractPrivateStateFromResponse(policy)

	if ps.ArtifactManifest == nil {
		t.Fatal("expected ArtifactManifest to be non-nil")
	}

	artifacts, ok := ps.ArtifactManifest["artifacts"].(map[string]any)
	if !ok {
		t.Fatal("expected ArtifactManifest.artifacts to be a map")
	}

	if _, ok := artifacts["endpoint-exceptionlist-macos-v1"]; !ok {
		t.Error("expected artifact_manifest to come from the endpoint input, not the first non-endpoint input")
	}
}

// TestExtractPrivateStateFromResponse tests that private state extraction
// captures the version and artifact_manifest from an API response.
func TestExtractPrivateStateFromResponse(t *testing.T) {
	version := testArtifactManifest
	policy := &kbapi.DefendPackagePolicy{
		Id:      "policy-123",
		Name:    "test",
		Enabled: true,
		Version: &version,
		Inputs: []kbapi.DefendPackagePolicyInput{
			{
				Type:    "endpoint",
				Enabled: true,
				Config: map[string]any{
					"artifact_manifest": map[string]any{
						"artifacts": map[string]any{
							"endpoint-exceptionlist-macos-v1": map[string]any{
								"sha256": "abc123",
							},
						},
					},
				},
			},
		},
	}

	ps := edip.ExtractPrivateStateFromResponse(policy)

	if ps.Version != version {
		t.Errorf("expected Version=%q, got %q", version, ps.Version)
	}

	if ps.ArtifactManifest == nil {
		t.Error("expected ArtifactManifest to be non-nil")
	}
}
