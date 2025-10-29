package fleet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetOperationalSpaceFromState extracts the operational space ID from Terraform state.
// This helper reads space_ids from state (not plan) to determine which space to use
// for API operations, preventing errors when space_ids changes (e.g., prepending a new space).
//
// **Why read from STATE not PLAN:**
// When updating space_ids = ["space-a"] → ["space-b", "space-a"], we need to query
// the policy in a space where it currently EXISTS (space-a from STATE), not where it
// WILL exist (space-b from PLAN). Otherwise, the API call fails with 404.
//
// Selection Strategy:
//  1. Extract space_ids from state
//  2. If empty/null → return "" (uses default space without /s/{spaceId} prefix)
//  3. Otherwise → return first space from state (where resource currently exists)
//
// Note: With Sets, there's no inherent ordering, but we can rely on deterministic
// iteration to get a consistent space for API operations.
func GetOperationalSpaceFromState(ctx context.Context, state tfsdk.State) (string, diag.Diagnostics) {
	var stateSpaces types.Set
	diags := state.GetAttribute(ctx, path.Root("space_ids"), &stateSpaces)
	if diags.HasError() {
		return "", diags
	}

	// If null/unknown, use default space (empty string)
	if stateSpaces.IsNull() || stateSpaces.IsUnknown() {
		return "", nil
	}

	// Extract space IDs from the Set
	var spaceIDs []string
	diags.Append(stateSpaces.ElementsAs(ctx, &spaceIDs, false)...)
	if diags.HasError() {
		return "", diags
	}

	// If empty, use default space
	if len(spaceIDs) == 0 {
		return "", nil
	}

	// Return first space (deterministic due to Set iteration)
	// This is where the resource currently exists in the API
	return spaceIDs[0], nil
}

// SpaceIDsToSet converts a Go string slice to a Terraform Set of strings.
func SpaceIDsToSet(ctx context.Context, spaceIDs []string) (types.Set, diag.Diagnostics) {
	if len(spaceIDs) == 0 {
		return types.SetNull(types.StringType), nil
	}

	spaceIDValues := make([]attr.Value, len(spaceIDs))
	for i, id := range spaceIDs {
		spaceIDValues[i] = types.StringValue(id)
	}

	return types.SetValue(types.StringType, spaceIDValues)
}
