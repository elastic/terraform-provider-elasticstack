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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	edip "github.com/elastic/terraform-provider-elasticstack/internal/fleet/elastic_defend_integration_policy"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const testArtifactManifest = "WzEyMywxXQ=="

// buildTestPackagePolicy constructs a kbapi.PackagePolicy with typed inputs
// suitable for Defend resource tests.
func buildTestPackagePolicy(id, name, pkgName, pkgVersion string, enabled bool, inputs kbapi.PackagePolicyTypedInputs) *kbapi.PackagePolicy {
	policy := &kbapi.PackagePolicy{
		Id:      id,
		Name:    name,
		Enabled: enabled,
		Package: &kbapi.KibanaHTTPAPIsPackagePolicyPackage{
			Name:    pkgName,
			Version: pkgVersion,
		},
	}
	if err := policy.Inputs.FromPackagePolicyTypedInputs(inputs); err != nil {
		panic("failed to set typed inputs: " + err.Error())
	}
	return policy
}

// buildTypedInputConfig constructs a PackagePolicyTypedInput.Config-compatible
// map entry from a raw value. The Defend API wraps config values in
// {frozen, type, value} structs.
func buildConfigEntry(value any) struct {
	Frozen *bool   `json:"frozen,omitempty"`
	Type   *string `json:"type,omitempty"`
	Value  any     `json:"value,omitempty"`
} {
	return struct {
		Frozen *bool   `json:"frozen,omitempty"`
		Type   *string `json:"type,omitempty"`
		Value  any     `json:"value,omitempty"`
	}{Value: value}
}

// TestPopulateModelFromAPIValidatesPackageName tests that populateModelFromAPI
// returns an error when the package name is not "endpoint" (REQ-005).
func TestPopulateModelFromAPIValidatesPackageName(t *testing.T) {
	ctx := context.Background()

	policy := buildTestPackagePolicy("wrong-policy-id", "wrong-package-policy", "not-endpoint", "1.0.0", false,
		kbapi.PackagePolicyTypedInputs{})

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

	icConfig := map[string]struct {
		Frozen *bool   `json:"frozen,omitempty"`
		Type   *string `json:"type,omitempty"`
		Value  any     `json:"value,omitempty"`
	}{
		"integration_config": buildConfigEntry(map[string]any{
			"endpointConfig": map[string]any{
				"preset": "NGAv1",
			},
		}),
	}

	inputs := kbapi.PackagePolicyTypedInputs{
		{
			Type:    "endpoint",
			Enabled: true,
			Config:  &icConfig,
			Streams: []kbapi.PackagePolicyTypedInputStream{},
		},
	}

	policy := buildTestPackagePolicy("policy-123", "my-endpoint-policy", "endpoint", "8.14.0", true, inputs)
	policy.Namespace = &namespace
	policy.PolicyId = &agentPolicyID
	policy.Version = &version

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

func TestPopulateModelFromAPIPreservesSpaceIDsWhenAPIOmitsThem(t *testing.T) {
	ctx := context.Background()

	inputs := kbapi.PackagePolicyTypedInputs{
		{
			Type:    "endpoint",
			Enabled: true,
			Streams: []kbapi.PackagePolicyTypedInputStream{},
		},
	}

	policy := buildTestPackagePolicy("policy-123", "my-endpoint-policy", "endpoint", "8.14.0", true, inputs)

	spaceIDs, diags := types.SetValueFrom(ctx, types.StringType, []string{"default"})
	if diags.HasError() {
		t.Fatalf("unexpected error creating space_ids set: %v", diags)
	}

	model := &edip.ElasticDefendIntegrationPolicyModel{
		SpaceIDs: spaceIDs,
	}
	diags = edip.PopulateModelFromAPI(ctx, model, policy)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	var actualSpaceIDs []string
	diags = model.SpaceIDs.ElementsAs(ctx, &actualSpaceIDs, false)
	if diags.HasError() {
		t.Fatalf("unexpected error reading space_ids set: %v", diags)
	}

	if len(actualSpaceIDs) != 1 || actualSpaceIDs[0] != "default" {
		t.Fatalf("expected preserved space_ids [default], got %v", actualSpaceIDs)
	}

	if model.ID.ValueString() != "default/policy-123" {
		t.Errorf("expected composite ID=%q, got %q", "default/policy-123", model.ID.ValueString())
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

// TestBuildBootstrapRequest tests that the bootstrap request uses the typed
// input format with ENDPOINT_INTEGRATION_CONFIG, and sends preset in
// _config (REQ-008).
func TestBuildBootstrapRequest(t *testing.T) {
	model := &edip.ElasticDefendIntegrationPolicyModel{
		Name:               types.StringValue("my-endpoint"),
		Namespace:          types.StringValue("default"),
		AgentPolicyID:      types.StringValue("agent-123"),
		IntegrationVersion: types.StringValue("8.14.0"),
		Preset:             types.StringValue("NGAv1"),
	}

	req, diags := edip.BuildBootstrapRequest(context.Background(), model)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if req.Package == nil || req.Package.Name != "endpoint" {
		name := "<nil>"
		if req.Package != nil {
			name = req.Package.Name
		}
		t.Errorf("expected package name %q, got %q", "endpoint", name)
	}

	if req.Inputs == nil || len(*req.Inputs) != 1 {
		count := 0
		if req.Inputs != nil {
			count = len(*req.Inputs)
		}
		t.Fatalf("expected 1 input, got %d", count)
	}

	input := (*req.Inputs)[0]
	if input.Type != "ENDPOINT_INTEGRATION_CONFIG" {
		t.Errorf("expected input type=%q, got %q", "ENDPOINT_INTEGRATION_CONFIG", input.Type)
	}

	if !input.Enabled {
		t.Error("expected input enabled=true")
	}

	if input.Streams == nil {
		t.Error("expected input streams to be non-nil (empty slice)")
	}

	// Verify the request serializes with inputs as a JSON array (typed format)
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	inputsRaw, ok := out["inputs"]
	if !ok {
		t.Fatal("expected inputs in request body")
	}

	// Typed format: inputs should be a JSON array
	if _, ok := inputsRaw.([]any); !ok {
		t.Errorf("expected inputs to be a JSON array (typed format), got %T", inputsRaw)
	}

	// Verify preset is in config. The request Config field is *map[string]interface{},
	// so config values are raw interface{} values (not the struct-wrapped type in responses).
	if input.Config == nil {
		t.Fatal("expected config to be non-nil when preset is set")
	}

	icEntry, ok := (*input.Config)["_config"]
	if !ok {
		t.Fatal("expected _config in bootstrap input config")
	}

	// The Config field is a typed map[string]struct{Frozen,Type,Value}; the
	// raw payload lives in Value.
	valueMap, ok := icEntry.Value.(map[string]any)
	if !ok {
		t.Fatalf("expected _config.value to be a map, got %T", icEntry.Value)
	}

	if valueMap["type"] != "endpoint" {
		t.Errorf("expected _config.value.type=%q, got %v", "endpoint", valueMap["type"])
	}

	ecMap, ok := valueMap["endpointConfig"].(map[string]any)
	if !ok {
		t.Fatal("expected endpointConfig to be a map")
	}

	if ecMap["preset"] != "NGAv1" {
		t.Errorf("expected preset=%q, got %v", "NGAv1", ecMap["preset"])
	}
}

// TestBuildBootstrapRequestNullPreset tests that a null preset omits the
// _config key from the bootstrap input config entirely, rather than
// sending an empty string.
func TestBuildBootstrapRequestNullPreset(t *testing.T) {
	model := &edip.ElasticDefendIntegrationPolicyModel{
		Name:               types.StringValue("my-endpoint"),
		Namespace:          types.StringValue("default"),
		AgentPolicyID:      types.StringValue("agent-123"),
		IntegrationVersion: types.StringValue("8.14.0"),
		Preset:             types.StringNull(),
	}

	req, diags := edip.BuildBootstrapRequest(context.Background(), model)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if req.Inputs == nil || len(*req.Inputs) != 1 {
		count := 0
		if req.Inputs != nil {
			count = len(*req.Inputs)
		}
		t.Fatalf("expected 1 input, got %d", count)
	}

	input := (*req.Inputs)[0]

	// When preset is null, Config should be nil (no _config)
	if input.Config != nil {
		if _, ok := (*input.Config)["_config"]; ok {
			t.Error("expected _config to be absent from bootstrap input config when preset is null")
		}
	}
}

func TestBuildBootstrapRequestWithAgentPolicyIDs(t *testing.T) {
	ctx := context.Background()
	model := &edip.ElasticDefendIntegrationPolicyModel{
		Name:               types.StringValue("my-endpoint"),
		Namespace:          types.StringValue("default"),
		AgentPolicyIDs:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("agent-123"), types.StringValue("agent-456")}),
		IntegrationVersion: types.StringValue("8.15.0"),
		Preset:             types.StringValue("NGAv1"),
	}

	req, diags := edip.BuildBootstrapRequest(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if req.PolicyIds == nil || len(*req.PolicyIds) != 2 {
		t.Fatalf("expected 2 policy IDs, got %#v", req.PolicyIds)
	}
	if (*req.PolicyIds)[0] != "agent-123" || (*req.PolicyIds)[1] != "agent-456" {
		t.Fatalf("unexpected policy IDs: %#v", *req.PolicyIds)
	}
	if req.PolicyId == nil || *req.PolicyId != "agent-123" {
		t.Fatalf("expected PolicyId compatibility value %q, got %#v", "agent-123", req.PolicyId)
	}
}

func TestPopulateModelFromAPIUsesAgentPolicyIDsWhenOriginallyConfigured(t *testing.T) {
	ctx := context.Background()
	policyIDs := []string{"agent-123", "agent-456"}
	inputs := kbapi.PackagePolicyTypedInputs{{Type: "endpoint", Enabled: true, Streams: []kbapi.PackagePolicyTypedInputStream{}}}
	policy := buildTestPackagePolicy("policy-123", "my-endpoint-policy", "endpoint", "8.15.0", true, inputs)
	policy.PolicyId = &policyIDs[0]
	policy.PolicyIds = &policyIDs

	model := &edip.ElasticDefendIntegrationPolicyModel{
		AgentPolicyIDs: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("placeholder")}),
	}
	diags := edip.PopulateModelFromAPI(ctx, model, policy)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if !model.AgentPolicyID.IsNull() && !model.AgentPolicyID.IsUnknown() {
		t.Fatalf("expected agent_policy_id to remain unset, got %v", model.AgentPolicyID)
	}
	var actual []string
	diags = model.AgentPolicyIDs.ElementsAs(ctx, &actual, false)
	if diags.HasError() {
		t.Fatalf("unexpected error reading agent_policy_ids: %v", diags)
	}
	if len(actual) != 2 || actual[0] != "agent-123" || actual[1] != "agent-456" {
		t.Fatalf("unexpected agent_policy_ids: %#v", actual)
	}
}

func TestPopulateModelFromAPIFallsBackToAgentPolicyIDsWhenMultiple(t *testing.T) {
	ctx := context.Background()
	policyIDs := []string{"agent-123", "agent-456"}
	inputs := kbapi.PackagePolicyTypedInputs{{Type: "endpoint", Enabled: true, Streams: []kbapi.PackagePolicyTypedInputStream{}}}
	policy := buildTestPackagePolicy("policy-123", "my-endpoint-policy", "endpoint", "8.15.0", true, inputs)
	policy.PolicyId = &policyIDs[0]
	policy.PolicyIds = &policyIDs

	// Neither field configured — simulates import with a fresh model
	model := &edip.ElasticDefendIntegrationPolicyModel{}
	diags := edip.PopulateModelFromAPI(ctx, model, policy)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if !model.AgentPolicyID.IsNull() && !model.AgentPolicyID.IsUnknown() {
		t.Fatalf("expected agent_policy_id to remain unset when API has multiple policy IDs, got %v", model.AgentPolicyID)
	}
	if model.AgentPolicyIDs.IsNull() || model.AgentPolicyIDs.IsUnknown() {
		t.Fatalf("expected agent_policy_ids to be populated when API has multiple policy IDs, got %v", model.AgentPolicyIDs)
	}
	var actual []string
	diags = model.AgentPolicyIDs.ElementsAs(ctx, &actual, false)
	if diags.HasError() {
		t.Fatalf("unexpected error reading agent_policy_ids: %v", diags)
	}
	if len(actual) != 2 || actual[0] != "agent-123" || actual[1] != "agent-456" {
		t.Fatalf("unexpected agent_policy_ids: %#v", actual)
	}
}

func TestPopulateModelFromAPIFallsBackToAgentPolicyIDWhenSingle(t *testing.T) {
	ctx := context.Background()
	policyID := "agent-123"
	inputs := kbapi.PackagePolicyTypedInputs{{Type: "endpoint", Enabled: true, Streams: []kbapi.PackagePolicyTypedInputStream{}}}
	policy := buildTestPackagePolicy("policy-123", "my-endpoint-policy", "endpoint", "8.15.0", true, inputs)
	policy.PolicyId = &policyID
	policy.PolicyIds = &[]string{policyID}

	// Neither field configured — simulates import with a fresh model
	model := &edip.ElasticDefendIntegrationPolicyModel{}
	diags := edip.PopulateModelFromAPI(ctx, model, policy)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if model.AgentPolicyID.IsNull() || model.AgentPolicyID.IsUnknown() {
		t.Fatalf("expected agent_policy_id to be populated when API has a single policy ID, got %v", model.AgentPolicyID)
	}
	if model.AgentPolicyID.ValueString() != policyID {
		t.Fatalf("expected agent_policy_id %q, got %q", policyID, model.AgentPolicyID.ValueString())
	}
	if !model.AgentPolicyIDs.IsNull() && !model.AgentPolicyIDs.IsUnknown() {
		t.Fatalf("expected agent_policy_ids to remain unset when API has a single policy ID, got %v", model.AgentPolicyIDs)
	}
}

func TestBuildFinalizeRequestIncludesArtifactManifest(t *testing.T) {
	ctx := context.Background()
	artifactManifest := map[string]any{
		"manifest_version": "1.0.0",
		"schema_version":   "v1",
		"artifacts":        map[string]any{},
	}
	endpointConfig := map[string]struct {
		Frozen *bool   `json:"frozen,omitempty"`
		Type   *string `json:"type,omitempty"`
		Value  any     `json:"value,omitempty"`
	}{
		"integration_config": buildConfigEntry(map[string]any{
			"endpointConfig": map[string]any{
				"preset": "EDRComplete",
			},
		}),
		"artifact_manifest": buildConfigEntry(artifactManifest),
		"policy": buildConfigEntry(map[string]any{
			"windows": map[string]any{"events": map[string]any{"process": true}},
			"mac":     map[string]any{"events": map[string]any{"process": true}},
			"linux":   map[string]any{"events": map[string]any{"process": true}},
		}),
	}
	inputs := kbapi.PackagePolicyTypedInputs{
		{
			Type:    "endpoint",
			Enabled: true,
			Config:  &endpointConfig,
			Streams: []kbapi.PackagePolicyTypedInputStream{},
		},
	}
	policy := buildTestPackagePolicy("policy-123", "my-endpoint", "endpoint", "8.14.0", true, inputs)
	policy.PolicyId = &[]string{"agent-123"}[0]

	model := &edip.ElasticDefendIntegrationPolicyModel{}
	diags := edip.PopulateModelFromAPI(ctx, model, policy)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics populating model, got %v", diags)
	}

	ps := edip.DefendPrivateState{
		Version:          "WzEyMywxXQ==",
		ArtifactManifest: artifactManifest,
	}

	req, diags := edip.BuildFinalizeRequest(ctx, model, nil, ps)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if req.Inputs == nil || len(*req.Inputs) != 1 {
		t.Fatalf("expected 1 input, got %v", req.Inputs)
	}

	input := (*req.Inputs)[0]
	if input.Config == nil {
		t.Fatal("expected config to be non-nil")
	}
	artifactManifestEntry, ok := (*input.Config)["artifact_manifest"]
	if !ok {
		t.Fatal("expected artifact_manifest in finalize input config")
	}
	valueMap, ok := artifactManifestEntry.Value.(map[string]any)
	if !ok {
		t.Fatalf("expected artifact_manifest.value to be a map, got %T", artifactManifestEntry.Value)
	}
	if valueMap["manifest_version"] != "1.0.0" {
		t.Errorf("expected manifest_version=%q, got %v", "1.0.0", valueMap["manifest_version"])
	}
}

func TestPopulateModelFromAPIAdvancedSettingsWhenConfigured(t *testing.T) {
	ctx := context.Background()

	endpointConfig := map[string]struct {
		Frozen *bool   `json:"frozen,omitempty"`
		Type   *string `json:"type,omitempty"`
		Value  any     `json:"value,omitempty"`
	}{
		"policy": buildConfigEntry(map[string]any{
			"linux": map[string]any{
				"advanced": map[string]any{
					"artifacts": map[string]any{
						"global": map[string]any{"base_url": "http://mirror.example.com"},
					},
				},
			},
		}),
	}
	inputs := kbapi.PackagePolicyTypedInputs{{
		Type:    "endpoint",
		Enabled: true,
		Config:  &endpointConfig,
		Streams: []kbapi.PackagePolicyTypedInputStream{},
	}}
	policy := buildTestPackagePolicy("policy-123", "my-endpoint", "endpoint", "8.14.0", true, inputs)

	model := &edip.ElasticDefendIntegrationPolicyModel{
		AdvancedSettings: types.MapValueMust(types.StringType, map[string]attr.Value{
			"linux.advanced.artifacts.global.base_url": types.StringValue("http://placeholder.example.com"),
		}),
	}
	diags := edip.PopulateModelFromAPI(ctx, model, policy)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}

	val, ok := model.AdvancedSettings.Elements()["linux.advanced.artifacts.global.base_url"]
	if !ok {
		t.Fatal("expected advanced setting in state")
	}
	if val.(types.String).ValueString() != "http://mirror.example.com" {
		t.Fatalf("unexpected value: %v", val)
	}
}

// TestExtractPrivateStateFromResponseNonEndpointFirst tests that
// extractPrivateStateFromResponse skips non-endpoint inputs and still finds
// the artifact_manifest from the endpoint input.
func TestExtractPrivateStateFromResponseNonEndpointFirst(t *testing.T) {
	version := testArtifactManifest

	amValue := map[string]any{
		"artifacts": map[string]any{
			"endpoint-exceptionlist-macos-v1": map[string]any{
				"sha256": "abc123",
			},
		},
	}
	wrongAmValue := map[string]any{
		"artifacts": map[string]any{
			"wrong": "should not be picked up",
		},
	}

	otherConfig := map[string]struct {
		Frozen *bool   `json:"frozen,omitempty"`
		Type   *string `json:"type,omitempty"`
		Value  any     `json:"value,omitempty"`
	}{
		"artifact_manifest": buildConfigEntry(wrongAmValue),
	}
	endpointConfig := map[string]struct {
		Frozen *bool   `json:"frozen,omitempty"`
		Type   *string `json:"type,omitempty"`
		Value  any     `json:"value,omitempty"`
	}{
		"artifact_manifest": buildConfigEntry(amValue),
	}

	inputs := kbapi.PackagePolicyTypedInputs{
		{
			Type:    "some-other-type",
			Enabled: true,
			Config:  &otherConfig,
			Streams: []kbapi.PackagePolicyTypedInputStream{},
		},
		{
			Type:    "endpoint",
			Enabled: true,
			Config:  &endpointConfig,
			Streams: []kbapi.PackagePolicyTypedInputStream{},
		},
	}

	policy := buildTestPackagePolicy("policy-123", "test", "endpoint", "8.14.0", true, inputs)
	policy.Version = &version

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

	amValue := map[string]any{
		"artifacts": map[string]any{
			"endpoint-exceptionlist-macos-v1": map[string]any{
				"sha256": "abc123",
			},
		},
	}
	endpointConfig := map[string]struct {
		Frozen *bool   `json:"frozen,omitempty"`
		Type   *string `json:"type,omitempty"`
		Value  any     `json:"value,omitempty"`
	}{
		"artifact_manifest": buildConfigEntry(amValue),
	}

	inputs := kbapi.PackagePolicyTypedInputs{
		{
			Type:    "endpoint",
			Enabled: true,
			Config:  &endpointConfig,
			Streams: []kbapi.PackagePolicyTypedInputStream{},
		},
	}

	policy := buildTestPackagePolicy("policy-123", "test", "endpoint", "8.14.0", true, inputs)
	policy.Version = &version

	ps := edip.ExtractPrivateStateFromResponse(policy)

	if ps.Version != version {
		t.Errorf("expected Version=%q, got %q", version, ps.Version)
	}

	if ps.ArtifactManifest == nil {
		t.Error("expected ArtifactManifest to be non-nil")
	}
}
