package fleet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
)

// getOperationalSpace determines which space to use for API operations.
//
// Fleet resources with space_ids support being visible in multiple spaces simultaneously.
// The resource has a single global ID, and space_ids controls visibility/access.
//
// DESIGN PRINCIPLE: Always prefer "default" space as the operational anchor.
// The "default" space cannot be deleted in Kibana, making it the most stable
// choice for API operations. This prevents resource orphaning when space_ids ordering changes.
//
// Selection Strategy:
//  1. If space_ids is empty/null → return nil (implicit default space)
//  2. If "default" is in space_ids → ALWAYS use "default" (most stable)
//  3. If empty string "" is in space_ids → use "default" (normalize empty to default)
//  4. Otherwise → use first space in list (fallback for custom-space-only resources)
//
// This ensures stable, predictable behavior that prevents resource orphaning
// when users reorder or modify space_ids.
//
// Example Scenarios:
//
//	[]                           → nil (default space)
//	["default"]                  → "default"
//	["space-a", "default"]       → "default" (prefer default over first)
//	["default", "space-a"]       → "default" (prefer default over first)
//	["space-a", "space-b"]       → "space-a" (no default, use first)
//	["", "space-a"]              → "default" (normalize empty string)
func GetOperationalSpace(spaceIDs []string) *string {
	if len(spaceIDs) == 0 {
		// Empty list means implicit default space
		// Return nil so API uses default space without /s/{spaceId} prefix
		return nil
	}

	// PRIORITY 1: Always prefer "default" space if present (most stable)
	// This prevents orphaning when users reorder space_ids
	for _, id := range spaceIDs {
		if id == "default" {
			defaultSpace := "default"
			return &defaultSpace
		}
		// Normalize empty string to "default"
		if id == "" {
			defaultSpace := "default"
			return &defaultSpace
		}
	}

	// PRIORITY 2: Fallback to first space (for custom-space-only resources)
	// This handles edge case where resource is intentionally not in default space
	return &spaceIDs[0]
}

// ExtractSpaceIDs converts a Terraform List of space IDs to a Go string slice.
// Returns empty slice if the list is null, unknown, or empty.
func ExtractSpaceIDs(ctx context.Context, spaceIDsList types.List) []string {
	if spaceIDsList.IsNull() || spaceIDsList.IsUnknown() {
		return []string{}
	}

	spaceIDTypes := utils.ListTypeAs[types.String](ctx, spaceIDsList, path.Root("space_ids"), &diag.Diagnostics{})
	if len(spaceIDTypes) == 0 {
		return []string{}
	}

	spaceIDs := make([]string, 0, len(spaceIDTypes))
	for _, idType := range spaceIDTypes {
		if !idType.IsNull() && !idType.IsUnknown() {
			spaceIDs = append(spaceIDs, idType.ValueString())
		}
	}

	return spaceIDs
}

// SpaceIDsToList converts a Go string slice to a Terraform List of strings.
func SpaceIDsToList(ctx context.Context, spaceIDs []string) (types.List, diag.Diagnostics) {
	if len(spaceIDs) == 0 {
		return types.ListNull(types.StringType), nil
	}

	spaceIDValues := make([]attr.Value, len(spaceIDs))
	for i, id := range spaceIDs {
		spaceIDValues[i] = types.StringValue(id)
	}

	return types.ListValue(types.StringType, spaceIDValues)
}

// ShouldPreserveSpaceIdsOrder determines whether to preserve the user's configured
// space_ids order instead of using the API response order.
//
// The Kibana Fleet API sorts space_ids alphabetically in responses (as of 9.1.3+),
// but Terraform expects the exact order from configuration to be preserved to avoid
// false drift detection.
//
// This function returns true ONLY when ALL of these conditions are met:
//
//  1. apiSpaceIds != nil: API response includes space_ids (Kibana 9.1+ feature)
//  2. !originalSpaceIds.IsNull(): User explicitly configured space_ids in plan/state
//  3. !originalSpaceIds.IsUnknown(): Value is known (not a computed value)
//  4. !populatedSpaceIds.IsNull(): populateFromAPI successfully set a non-null value
//
// Edge Cases Handled:
//   - User doesn't configure space_ids → Use API's computed default
//   - Older Kibana versions (no space_ids support) → Use API's null value
//   - User configures, but Kibana sorts → Preserve user's order
//   - Computed values → Let provider compute, don't preserve
//
// Example Truth Table:
//
//	User Config | API Response  | After Populate | Preserve? | Result
//	----------- | ------------- | -------------- | --------- | ------
//	null        | null          | null           | No        | null (old Kibana)
//	null        | ["default"]   | ["default"]    | No        | ["default"] (computed)
//	["a","b"]   | ["b","a"]     | ["b","a"]      | Yes       | ["a","b"] (user order)
//	unknown     | ["default"]   | ["default"]    | No        | ["default"] (computed)
//	["a"]       | null          | null           | No        | null (not supported)
func ShouldPreserveSpaceIdsOrder(apiSpaceIds *[]string, originalSpaceIds types.List, populatedSpaceIds types.List) bool {
	return apiSpaceIds != nil &&
		!originalSpaceIds.IsNull() &&
		!originalSpaceIds.IsUnknown() &&
		!populatedSpaceIds.IsNull()
}
