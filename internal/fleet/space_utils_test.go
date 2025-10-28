package fleet

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGetOperationalSpace(t *testing.T) {
	tests := []struct {
		name     string
		spaceIDs []string
		want     *string
		reason   string
	}{
		{
			name:     "empty list returns nil (implicit default)",
			spaceIDs: []string{},
			want:     nil,
			reason:   "Empty space_ids means use default space without /s/{spaceId} prefix",
		},
		{
			name:     "nil list returns nil (implicit default)",
			spaceIDs: nil,
			want:     nil,
			reason:   "Nil space_ids means use default space without /s/{spaceId} prefix",
		},
		{
			name:     "single default space",
			spaceIDs: []string{"default"},
			want:     stringPtr("default"),
			reason:   "Explicit default space",
		},
		{
			name:     "default space first in list",
			spaceIDs: []string{"default", "space-a"},
			want:     stringPtr("default"),
			reason:   "Default space is most stable, always prefer it",
		},
		{
			name:     "default space last in list",
			spaceIDs: []string{"space-a", "space-b", "default"},
			want:     stringPtr("default"),
			reason:   "CRITICAL: Prefer default over first - prevents orphaning on reorder",
		},
		{
			name:     "default space in middle of list",
			spaceIDs: []string{"space-a", "default", "space-b"},
			want:     stringPtr("default"),
			reason:   "CRITICAL: Prefer default regardless of position",
		},
		{
			name:     "no default space - use first",
			spaceIDs: []string{"space-a", "space-b"},
			want:     stringPtr("space-a"),
			reason:   "Fallback to first space when default not present",
		},
		{
			name:     "single custom space",
			spaceIDs: []string{"custom-space"},
			want:     stringPtr("custom-space"),
			reason:   "Single custom space (no default available)",
		},
		{
			name:     "empty string normalized to default",
			spaceIDs: []string{"", "space-a"},
			want:     stringPtr("default"),
			reason:   "Empty string should be treated as default space",
		},
		{
			name:     "empty string with other spaces",
			spaceIDs: []string{"space-a", "", "space-b"},
			want:     stringPtr("default"),
			reason:   "Empty string normalized to default takes precedence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetOperationalSpace(tt.spaceIDs)

			// Compare pointers and values
			if tt.want == nil {
				if got != nil {
					t.Errorf("GetOperationalSpace() = %v, want nil\nReason: %s", *got, tt.reason)
				}
			} else if got == nil {
				t.Errorf("GetOperationalSpace() = nil, want %v\nReason: %s", *tt.want, tt.reason)
			} else if *got != *tt.want {
				t.Errorf("GetOperationalSpace() = %v, want %v\nReason: %s", *got, *tt.want, tt.reason)
			}
		})
	}
}

// TestGetOperationalSpace_OrphaningPrevention validates the critical bug fix.
// This test ensures that reordering space_ids doesn't cause resource orphaning.
func TestGetOperationalSpace_OrphaningPrevention(t *testing.T) {
	scenarios := []struct {
		name         string
		initialState []string
		updatedState []string
		reason       string
	}{
		{
			name:         "prepend new space to list with default",
			initialState: []string{"default", "space-a"},
			updatedState: []string{"space-new", "default", "space-a"},
			reason:       "CRITICAL BUG FIX: Should use 'default' in both states, not first space",
		},
		{
			name:         "append new space to list with default",
			initialState: []string{"default", "space-a"},
			updatedState: []string{"default", "space-a", "space-new"},
			reason:       "Should use 'default' in both states (stable)",
		},
		{
			name:         "reorder spaces with default",
			initialState: []string{"space-a", "default"},
			updatedState: []string{"default", "space-a"},
			reason:       "Should use 'default' in both states (order independent)",
		},
		{
			name:         "move default from end to beginning",
			initialState: []string{"space-a", "space-b", "default"},
			updatedState: []string{"default", "space-a", "space-b"},
			reason:       "Should use 'default' in both states (position independent)",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			initialSpace := GetOperationalSpace(sc.initialState)
			updatedSpace := GetOperationalSpace(sc.updatedState)

			// Both should return "default"
			if initialSpace == nil || *initialSpace != "default" {
				t.Errorf("Initial state returned %v, want 'default'\nReason: %s",
					spaceToString(initialSpace), sc.reason)
			}

			if updatedSpace == nil || *updatedSpace != "default" {
				t.Errorf("Updated state returned %v, want 'default'\nReason: %s",
					spaceToString(updatedSpace), sc.reason)
			}

			// Most importantly: they should be EQUAL
			if !spacesEqual(initialSpace, updatedSpace) {
				t.Errorf("Operational space changed: %v → %v\nThis would cause resource orphaning!\nReason: %s",
					spaceToString(initialSpace), spaceToString(updatedSpace), sc.reason)
			}
		})
	}
}

// TestGetOperationalSpace_EdgeCases validates edge case handling.
func TestGetOperationalSpace_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		spaceIDs []string
		want     *string
	}{
		{
			name:     "multiple defaults (should handle gracefully)",
			spaceIDs: []string{"default", "default", "space-a"},
			want:     stringPtr("default"),
		},
		{
			name:     "spaces with special characters",
			spaceIDs: []string{"my-space", "another_space", "space.with.dots"},
			want:     stringPtr("my-space"),
		},
		{
			name:     "very long space list",
			spaceIDs: []string{"s1", "s2", "s3", "s4", "s5", "default", "s7", "s8"},
			want:     stringPtr("default"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetOperationalSpace(tt.spaceIDs)

			if tt.want == nil {
				if got != nil {
					t.Errorf("GetOperationalSpace() = %v, want nil", *got)
				}
			} else if got == nil {
				t.Errorf("GetOperationalSpace() = nil, want %v", *tt.want)
			} else if *got != *tt.want {
				t.Errorf("GetOperationalSpace() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func spaceToString(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func spacesEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// TestExtractSpaceIDs tests the ExtractSpaceIDs helper function that converts
// Terraform List types to Go string slices.
func TestExtractSpaceIDs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		input    types.List
		expected []string
	}{
		{
			name:     "null list returns empty slice",
			input:    types.ListNull(types.StringType),
			expected: []string{},
		},
		{
			name:     "unknown list returns empty slice",
			input:    types.ListUnknown(types.StringType),
			expected: []string{},
		},
		{
			name: "single space",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("default"),
			}),
			expected: []string{"default"},
		},
		{
			name: "multiple spaces",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("default"),
				types.StringValue("space-a"),
				types.StringValue("space-b"),
			}),
			expected: []string{"default", "space-a", "space-b"},
		},
		{
			name: "spaces with special characters",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("my-space"),
				types.StringValue("another_space"),
				types.StringValue("space.with.dots"),
			}),
			expected: []string{"my-space", "another_space", "space.with.dots"},
		},
		{
			name: "empty string in list",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue(""),
				types.StringValue("space-a"),
			}),
			expected: []string{"", "space-a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractSpaceIDs(ctx, tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("ExtractSpaceIDs() length = %v, want %v", len(got), len(tt.expected))
				return
			}

			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("ExtractSpaceIDs()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

// TestSpaceIDsToList tests the SpaceIDsToList helper function that converts
// Go string slices to Terraform List types.
func TestSpaceIDsToList(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		input       []string
		expectError bool
	}{
		{
			name:        "empty slice",
			input:       []string{},
			expectError: false,
		},
		{
			name:        "nil slice",
			input:       nil,
			expectError: false,
		},
		{
			name:        "single space",
			input:       []string{"default"},
			expectError: false,
		},
		{
			name:        "multiple spaces",
			input:       []string{"default", "space-a", "space-b"},
			expectError: false,
		},
		{
			name:        "spaces with special characters",
			input:       []string{"my-space", "another_space", "space.with.dots"},
			expectError: false,
		},
		{
			name:        "empty string in slice",
			input:       []string{"", "space-a"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, diags := SpaceIDsToList(ctx, tt.input)

			if tt.expectError {
				if !diags.HasError() {
					t.Error("SpaceIDsToList() expected error, got none")
				}
				return
			}

			if diags.HasError() {
				t.Errorf("SpaceIDsToList() unexpected error: %v", diags)
				return
			}

			// Verify the list was created correctly
			// Note: Empty slices return null lists (by design)
			if len(tt.input) == 0 {
				if !got.IsNull() {
					t.Error("SpaceIDsToList() should return null list for empty input")
				}
				return
			}

			if got.IsNull() {
				t.Error("SpaceIDsToList() returned null list for non-empty input")
				return
			}

			// Convert back to slice to verify round-trip
			var result []types.String
			diags = got.ElementsAs(ctx, &result, false)
			if diags.HasError() {
				t.Errorf("ElementsAs() error: %v", diags)
				return
			}

			if len(result) != len(tt.input) {
				t.Errorf("SpaceIDsToList() length = %v, want %v", len(result), len(tt.input))
				return
			}

			for i := range result {
				if result[i].ValueString() != tt.input[i] {
					t.Errorf("SpaceIDsToList()[%d] = %v, want %v", i, result[i].ValueString(), tt.input[i])
				}
			}
		})
	}
}

// TestExtractAndConvertRoundTrip tests that ExtractSpaceIDs and SpaceIDsToList
// are inverse operations (round-trip conversion works correctly).
func TestShouldPreserveSpaceIdsOrder(t *testing.T) {
	tests := []struct {
		name              string
		apiSpaceIds       *[]string
		originalSpaceIds  types.List
		populatedSpaceIds types.List
		expectedPreserve  bool
		description       string
	}{
		{
			name:              "user doesn't configure space_ids (null)",
			apiSpaceIds:       &[]string{"default"},
			originalSpaceIds:  types.ListNull(types.StringType),
			populatedSpaceIds: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default")}),
			expectedPreserve:  false,
			description:       "When user doesn't configure space_ids, use API's computed default",
		},
		{
			name:              "user configures space_ids, API sorts them (Kibana 9.1.3+)",
			apiSpaceIds:       &[]string{"default", "space-test-a"},
			originalSpaceIds:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("space-test-a"), types.StringValue("default")}),
			populatedSpaceIds: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default"), types.StringValue("space-test-a")}),
			expectedPreserve:  true,
			description:       "When user configures order and API sorts, preserve user's order",
		},
		{
			name:              "user configures space_ids on older Kibana (no support)",
			apiSpaceIds:       nil,
			originalSpaceIds:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default")}),
			populatedSpaceIds: types.ListNull(types.StringType),
			expectedPreserve:  false,
			description:       "When older Kibana doesn't support space_ids, don't preserve (feature not supported)",
		},
		{
			name:              "computed value (unknown in plan)",
			apiSpaceIds:       &[]string{"default"},
			originalSpaceIds:  types.ListUnknown(types.StringType),
			populatedSpaceIds: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default")}),
			expectedPreserve:  false,
			description:       "When value was computed (unknown), let provider compute, don't preserve",
		},
		{
			name:              "API returns null but user configured (edge case)",
			apiSpaceIds:       nil,
			originalSpaceIds:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("space-a")}),
			populatedSpaceIds: types.ListNull(types.StringType),
			expectedPreserve:  false,
			description:       "When API returns nil, don't preserve even if user configured (old Kibana)",
		},
		{
			name:              "populateFromAPI set null value",
			apiSpaceIds:       &[]string{},
			originalSpaceIds:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default")}),
			populatedSpaceIds: types.ListNull(types.StringType),
			expectedPreserve:  false,
			description:       "When populateFromAPI sets null, don't preserve (empty API response)",
		},
		{
			name:              "all conditions met - single space",
			apiSpaceIds:       &[]string{"default"},
			originalSpaceIds:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default")}),
			populatedSpaceIds: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default")}),
			expectedPreserve:  true,
			description:       "When all conditions met with single space, preserve (edge case where order doesn't matter but logic should still work)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldPreserveSpaceIdsOrder(tt.apiSpaceIds, tt.originalSpaceIds, tt.populatedSpaceIds)
			if got != tt.expectedPreserve {
				t.Errorf("ShouldPreserveSpaceIdsOrder() = %v, want %v\nDescription: %s", got, tt.expectedPreserve, tt.description)
			}
		})
	}
}

func TestExtractAndConvertRoundTrip(t *testing.T) {
	ctx := context.Background()

	testCases := [][]string{
		{},
		{"default"},
		{"default", "space-a"},
		{"space-a", "space-b", "space-c"},
		{"my-space", "another_space", "space.with.dots"},
	}

	for _, original := range testCases {
		t.Run(fmt.Sprintf("round_trip_%d_spaces", len(original)), func(t *testing.T) {
			// Convert to Terraform List
			list, diags := SpaceIDsToList(ctx, original)
			if diags.HasError() {
				t.Fatalf("SpaceIDsToList() error: %v", diags)
			}

			// Convert back to Go slice
			result := ExtractSpaceIDs(ctx, list)

			// Verify they match
			if len(result) != len(original) {
				t.Errorf("Round-trip length mismatch: got %v, want %v", len(result), len(original))
				return
			}

			for i := range result {
				if result[i] != original[i] {
					t.Errorf("Round-trip[%d] mismatch: got %v, want %v", i, result[i], original[i])
				}
			}
		})
	}
}
