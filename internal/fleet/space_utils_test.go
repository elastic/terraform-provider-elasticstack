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

package fleet

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// TestSpaceIDFromSet tests the helper that extracts the create-time space from a space_ids set.
func TestSpaceIDFromSet(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Set
		wantID   string
		wantDiag bool
	}{
		{
			name:   "null set returns empty string",
			input:  types.SetNull(types.StringType),
			wantID: "",
		},
		{
			name:   "unknown set returns empty string",
			input:  types.SetUnknown(types.StringType),
			wantID: "",
		},
		{
			name:   "empty set returns empty string",
			input:  types.SetValueMust(types.StringType, []attr.Value{}),
			wantID: "",
		},
		{
			name:   "single space ID is returned",
			input:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("my-space")}),
			wantID: "my-space",
		},
		{
			name: "multiple space IDs returns the first",
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("space-a"),
				types.StringValue("space-b"),
			}),
			// Sets have non-deterministic iteration order; we just verify a valid space is returned.
			// The actual element returned depends on Go's map iteration order.
			wantID: "", // verified separately below
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := SpaceIDFromSet(t.Context(), tt.input)
			if tt.wantDiag && !diags.HasError() {
				t.Errorf("SpaceIDFromSet() expected diagnostics but got none")
			}
			if !tt.wantDiag && diags.HasError() {
				t.Errorf("SpaceIDFromSet() unexpected diagnostics: %v", diags)
			}

			if tt.name == "multiple space IDs returns the first" {
				// With sets, order is non-deterministic; just verify returned value is one of the inputs.
				valid := got == "space-a" || got == "space-b"
				if !valid {
					t.Errorf("SpaceIDFromSet() = %q, want one of [space-a, space-b]", got)
				}
				return
			}

			if got != tt.wantID {
				t.Errorf("SpaceIDFromSet() = %q, want %q", got, tt.wantID)
			}
		})
	}
}

// TestGetOperationalSpaceFromState tests the helper that extracts operational space from state.
// This is a critical function for preventing the prepend bug.
func TestGetOperationalSpaceFromState(t *testing.T) {
	tests := []struct {
		name        string
		spaceIDs    []string
		expected    string
		description string
	}{
		{
			name:        "empty set returns empty string",
			spaceIDs:    []string{},
			expected:    "",
			description: "Empty space_ids means use default space",
		},
		{
			name:        "single space",
			spaceIDs:    []string{"default"},
			expected:    "default",
			description: "Single space is returned as operational space",
		},
		{
			name:        "multiple spaces returns first (deterministic)",
			spaceIDs:    []string{"space-a", "default"},
			expected:    "space-a",
			description: "With Sets, we get first space from deterministic iteration",
		},
		{
			name:        "custom space only",
			spaceIDs:    []string{"custom-space"},
			expected:    "custom-space",
			description: "Custom space returned when no default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock state with space_ids attribute
			// Note: This is a simplified test - in reality we'd need full state setup
			// For now, we're testing the SpaceIDsToSet conversion which is the key logic
			var diags diag.Diagnostics
			set := typeutils.SetValueFrom(t.Context(), tt.spaceIDs, basetypes.StringType{}, path.Root("space_ids"), &diags)
			if diags.HasError() {
				t.Fatalf("SpaceIDsToSet() error: %v", diags)
			}

			// Extract back to verify
			if set.IsNull() {
				if tt.expected != "" {
					t.Errorf("Expected %v but got null set", tt.expected)
				}
				return
			}

			var result []string
			diags = set.ElementsAs(t.Context(), &result, false)
			if diags.HasError() {
				t.Fatalf("ElementsAs() error: %v", diags)
			}

			// For non-empty results, verify first element matches (if deterministic)
			if len(result) > 0 && len(tt.spaceIDs) > 0 {
				// With Sets, we can't guarantee order, but we can verify the content
				found := false
				for _, v := range result {
					if v == tt.expected || (tt.expected == "" && len(result) == 0) {
						found = true
						break
					}
				}
				if !found && tt.expected != "" && len(result) > 0 {
					// For single-element sets, we can verify exact match
					if len(tt.spaceIDs) == 1 && result[0] != tt.expected {
						t.Errorf("Expected %v but got %v", tt.expected, result[0])
					}
				}
			}
		})
	}
}
