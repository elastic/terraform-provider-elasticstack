package fleet

import (
	"testing"
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
