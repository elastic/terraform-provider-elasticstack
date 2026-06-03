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
	"encoding/json"
	"testing"
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
