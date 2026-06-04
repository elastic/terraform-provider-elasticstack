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

package entity

import (
	"context"
	"encoding/json"
	"testing"

	jsontypes "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCanonicalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:  "map with unsorted keys",
			input: map[string]any{"z": 1, "a": 2, "m": 3},
			want:  `{"a":2,"m":3,"z":1}`,
		},
		{
			name:  "nested map",
			input: map[string]any{"b": map[string]any{"z": 1, "a": 2}, "a": "hello"},
			want:  `{"a":"hello","b":{"a":2,"z":1}}`,
		},
		{
			name:  "array preserves order",
			input: map[string]any{"list": []any{3, 1, 2}},
			want:  `{"list":[3,1,2]}`,
		},
		{
			name:    "nil returns empty",
			input:   nil,
			want:    `null`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := canonicalJSON(tt.input)
			if tt.wantErr {
				if !diags.HasError() {
					t.Errorf("canonicalJSON(%v) expected error", tt.input)
				}
				return
			}
			if diags.HasError() {
				t.Fatalf("canonicalJSON(%v) unexpected error: %v", tt.input, diags)
			}
			if got != tt.want {
				t.Errorf("canonicalJSON(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestInjectEntityIDAndMarshal(t *testing.T) {
	tests := []struct {
		name     string
		bodyMap  map[string]any
		entityID string
		wantID   string
	}{
		{
			name:     "injects id when entity key absent",
			bodyMap:  map[string]any{"other": "value"},
			entityID: "abc",
			wantID:   "abc",
		},
		{
			name:     "merges id into existing entity map",
			bodyMap:  map[string]any{"entity": map[string]any{"name": "test"}},
			entityID: "xyz",
			wantID:   "xyz",
		},
		{
			name:     "overwrites existing id in entity map",
			bodyMap:  map[string]any{"entity": map[string]any{"id": "old", "name": "test"}},
			entityID: "new",
			wantID:   "new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, diags := injectEntityIDAndMarshal(tt.bodyMap, tt.entityID)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			var result map[string]any
			if err := json.Unmarshal(b, &result); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			entityMap, ok := result["entity"].(map[string]any)
			if !ok {
				t.Fatalf("entity key not a map: %T", result["entity"])
			}
			if got := entityMap["id"]; got != tt.wantID {
				t.Errorf("entity.id = %v, want %v", got, tt.wantID)
			}
		})
	}
}

func TestExtractEntitiesFromResponse(t *testing.T) {
	tests := []struct {
		name   string
		result map[string]any
		want   []any
	}{
		{
			name:   "entities key present",
			result: map[string]any{"entities": []any{"a", "b"}},
			want:   []any{"a", "b"},
		},
		{
			name:   "records key fallback",
			result: map[string]any{"records": []any{"c", "d"}},
			want:   []any{"c", "d"},
		},
		{
			name:   "entities takes precedence over records",
			result: map[string]any{"entities": []any{"x"}, "records": []any{"y"}},
			want:   []any{"x"},
		},
		{
			name:   "neither key present returns nil",
			result: map[string]any{"other": []any{"z"}},
			want:   nil,
		},
		{
			name:   "empty map returns nil",
			result: map[string]any{},
			want:   nil,
		},
		{
			name:   "entities key wrong type falls back to records",
			result: map[string]any{"entities": "not-a-slice", "records": []any{"r"}},
			want:   []any{"r"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractEntitiesFromResponse(tt.result)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractEntitiesFromResponse() len = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("ExtractEntitiesFromResponse()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCanonicalMapJSON(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		want  string
	}{
		{
			name:  "nil returns empty",
			input: nil,
			want:  "",
		},
		{
			name:  "empty map returns empty object",
			input: map[string]any{},
			want:  `{}`,
		},
		{
			name:  "sorted keys",
			input: map[string]any{"key_b": 2, "key_a": 1},
			want:  `{"key_a":1,"key_b":2}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canonicalMapJSON(tt.input)
			if got != tt.want {
				t.Errorf("canonicalMapJSON(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func newNullPlan(spaceID types.String, entityType, entityID string) tfModel {
	return tfModel{
		SpaceID:          spaceID,
		EntityType:       types.StringValue(entityType),
		EntityID:         types.StringValue(entityID),
		Entity:           types.ObjectNull(BlockAttrTypes()),
		Host:             types.ObjectNull(HostBlockAttrTypes()),
		User:             types.ObjectNull(UserBlockAttrTypes()),
		Service:          types.ObjectNull(ServiceBlockAttrTypes()),
		Cloud:            types.ObjectNull(CloudBlockAttrTypes()),
		Asset:            types.ObjectNull(AssetBlockAttrTypes()),
		Orchestrator:     types.ObjectNull(OrchestratorBlockAttrTypes()),
		Event:            types.ObjectNull(EventBlockAttrTypes()),
		Labels:           types.MapNull(types.StringType),
		Tags:             types.SetNull(types.StringType),
		EntityJSON:       jsontypes.NewNormalizedNull(),
		HostJSON:         jsontypes.NewNormalizedNull(),
		UserJSON:         jsontypes.NewNormalizedNull(),
		ServiceJSON:      jsontypes.NewNormalizedNull(),
		CloudJSON:        jsontypes.NewNormalizedNull(),
		AssetJSON:        jsontypes.NewNormalizedNull(),
		OrchestratorJSON: jsontypes.NewNormalizedNull(),
		EventJSON:        jsontypes.NewNormalizedNull(),
		LabelsJSON:       jsontypes.NewNormalizedNull(),
	}
}

func TestBuildEntityWriteBody(t *testing.T) {
	ctx := context.Background()

	t.Run("resolves spaceID and entityType from plan", func(t *testing.T) {
		plan := newNullPlan(types.StringValue("my-space"), "host", "host-1")

		spaceID, entityType, bodyBytes, diags := buildEntityWriteBody(ctx, plan)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if spaceID != "my-space" {
			t.Errorf("spaceID = %q, want %q", spaceID, "my-space")
		}
		if entityType != "host" {
			t.Errorf("entityType = %q, want %q", entityType, "host")
		}
		var result map[string]any
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		entityMap, ok := result["entity"].(map[string]any)
		if !ok {
			t.Fatalf("entity key missing or wrong type: %T", result["entity"])
		}
		if got := entityMap["id"]; got != "host-1" {
			t.Errorf("entity.id = %v, want %q", got, "host-1")
		}
	})

	t.Run("null space_id normalizes to default", func(t *testing.T) {
		plan := newNullPlan(types.StringNull(), "user", "u-1")

		spaceID, _, _, diags := buildEntityWriteBody(ctx, plan)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if spaceID != "default" {
			t.Errorf("spaceID = %q, want %q", spaceID, "default")
		}
	})

	t.Run("entity id is injected into body bytes", func(t *testing.T) {
		plan := newNullPlan(types.StringValue("default"), "service", "svc-42")

		_, _, bodyBytes, diags := buildEntityWriteBody(ctx, plan)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		var result map[string]any
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		entityMap, ok := result["entity"].(map[string]any)
		if !ok {
			t.Fatalf("entity key missing or wrong type")
		}
		if got := entityMap["id"]; got != "svc-42" {
			t.Errorf("entity.id = %v, want %q", got, "svc-42")
		}
	})
}
