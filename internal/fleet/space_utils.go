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
	"context"

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
