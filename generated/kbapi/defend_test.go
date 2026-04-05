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

// Tests for the hand-authored defend.go types in the kbapi package.
// These verify that the typed-input encoding types are distinct from the
// generic mapped-input types and that JSON serialisation behaves correctly.

package kbapi_test

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

// TestDefendRequestInputEncoding verifies that DefendPackagePolicyRequestInput
// serialises correctly without a "type" field (the type is the map key in
// DefendPackagePolicyRequest.Inputs).
func TestDefendRequestInputEncoding(t *testing.T) {
	input := kbapi.DefendPackagePolicyRequestInput{
		Enabled: true,
		Streams: []interface{}{},
		Config: map[string]interface{}{
			"integration_config": map[string]interface{}{
				"value": map[string]interface{}{
					"endpointConfig": map[string]interface{}{
						"preset": "NGAv1",
					},
				},
			},
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// The "type" field is now the map key, not a field in the input struct
	if _, ok := out["type"]; ok {
		t.Errorf("expected no 'type' field in input struct (type is now the map key), got: %v", out["type"])
	}

	if _, ok := out["streams"]; !ok {
		t.Errorf("expected streams field to be present")
	}

	if _, ok := out["config"]; !ok {
		t.Errorf("expected config field to be present")
	}
}

// TestDefendRequestVersionField verifies that the top-level "version" field in
// DefendPackagePolicyRequest is serialised correctly, which is required for
// Defend update requests (optimistic concurrency control).
func TestDefendRequestVersionField(t *testing.T) {
	version := "WzEyMywxXQ=="
	req := kbapi.DefendPackagePolicyRequest{
		Name:    "my-endpoint-policy",
		Version: &version,
		Package: kbapi.PackagePolicyRequestPackage{
			Name:    "endpoint",
			Version: "8.14.0",
		},
		Inputs: map[string]kbapi.DefendPackagePolicyRequestInput{},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got, ok := out["version"]; !ok || got != version {
		t.Errorf("expected version=%q, got %v", version, out["version"])
	}
}

// TestDefendResponseTypedInputShape verifies that DefendPackagePolicy deserialises
// the typed inputs list correctly.
func TestDefendResponseTypedInputShape(t *testing.T) {
	jsonData := `{
		"id": "policy-123",
		"name": "my-endpoint-policy",
		"enabled": true,
		"version": "WzEyMywxXQ==",
		"inputs": [
			{
				"type": "endpoint",
				"enabled": true,
				"config": {
					"integration_config": {
						"value": {
							"endpointConfig": {
								"preset": "NGAv1"
							}
						}
					}
				}
			}
		]
	}`

	var policy kbapi.DefendPackagePolicy
	if err := json.Unmarshal([]byte(jsonData), &policy); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if policy.Id != "policy-123" {
		t.Errorf("expected Id=%q, got %q", "policy-123", policy.Id)
	}

	if policy.Version == nil || *policy.Version != "WzEyMywxXQ==" {
		t.Errorf("expected Version=%q, got %v", "WzEyMywxXQ==", policy.Version)
	}

	if len(policy.Inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(policy.Inputs))
	}

	input := policy.Inputs[0]
	if input.Type != "endpoint" {
		t.Errorf("expected input type=%q, got %q", "endpoint", input.Type)
	}

	if input.Config == nil {
		t.Fatal("expected input config to be non-nil")
	}

	ic, ok := input.Config["integration_config"]
	if !ok {
		t.Fatal("expected integration_config in input config")
	}
	_ = ic
}

// TestMappedVsTypedInputSameShape verifies that both the generic PackagePolicyRequest
// and DefendPackagePolicyRequest use a map-keyed inputs shape (JSON object), since
// the Fleet API requires inputs as an object for all package policies on 8.14+.
func TestMappedVsTypedInputSameShape(t *testing.T) {
	// Generic (mapped) request: inputs is *map[string]PackagePolicyRequestInput
	mappedReq := kbapi.PackagePolicyRequest{
		Name: "generic-policy",
		Package: kbapi.PackagePolicyRequestPackage{
			Name:    "some-integration",
			Version: "1.0.0",
		},
		Inputs: func() *map[string]kbapi.PackagePolicyRequestInput {
			m := map[string]kbapi.PackagePolicyRequestInput{
				"some-input": {
					Enabled: ptr(true),
				},
			}
			return &m
		}(),
	}

	mappedData, err := json.Marshal(mappedReq)
	if err != nil {
		t.Fatalf("json.Marshal mappedReq failed: %v", err)
	}

	var mappedOut map[string]interface{}
	if err := json.Unmarshal(mappedData, &mappedOut); err != nil {
		t.Fatalf("json.Unmarshal mappedData failed: %v", err)
	}

	// For mapped requests, inputs should be an object (map)
	inputs, ok := mappedOut["inputs"]
	if !ok {
		t.Fatal("expected inputs in mapped request")
	}

	if _, ok := inputs.(map[string]interface{}); !ok {
		t.Errorf("expected mapped inputs to be an object, got %T", inputs)
	}

	// Defend request: inputs is map[string]DefendPackagePolicyRequestInput
	// This matches the Fleet API's simplified format (map keyed by input type)
	defendReq := kbapi.DefendPackagePolicyRequest{
		Name: "endpoint-policy",
		Package: kbapi.PackagePolicyRequestPackage{
			Name:    "endpoint",
			Version: "8.14.0",
		},
		Inputs: map[string]kbapi.DefendPackagePolicyRequestInput{
			"endpoint": {
				Enabled: true,
				Streams: []interface{}{},
			},
		},
	}

	defendData, err := json.Marshal(defendReq)
	if err != nil {
		t.Fatalf("json.Marshal defendReq failed: %v", err)
	}

	var defendOut map[string]interface{}
	if err := json.Unmarshal(defendData, &defendOut); err != nil {
		t.Fatalf("json.Unmarshal defendData failed: %v", err)
	}

	// Defend inputs should also be an object (map), not an array
	defendInputs, ok := defendOut["inputs"]
	if !ok {
		t.Fatal("expected inputs in Defend request")
	}

	if _, ok := defendInputs.(map[string]interface{}); !ok {
		t.Errorf("expected Defend inputs to be an object (map), got %T", defendInputs)
	}
}

func ptr[T any](v T) *T { return &v }
