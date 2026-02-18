package security_exception_item

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	entryTypeMatch    = "match"
	entryTypeWildcard = "wildcard"
	entryTypeMatchAny = "match_any"
	entryTypeList     = "list"
	entryTypeExists   = "exists"
	entryTypeNested   = "nested"
)

// ValidateConfig validates the configuration for an exception item resource.
// It ensures that entries are properly configured based on their type:
//
// - For "match" and "wildcard" types: 'value' must be set
// - For "match_any" type: 'values' must be set
// - For "list" type: 'list' object must be set with 'id' and 'type'
// - For "exists" type: only 'field' and 'operator' are required
// - For "nested" type: 'entries' must be set and validated recursively
// - The 'operator' field is required for all types except "nested"
//
// Validation only runs on known values. Values that are unknown (e.g., references to
// other resources that haven't been created yet) are skipped.
//
// The function adds appropriate error diagnostics if validation fails.
func (r *ExceptionItemResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ExceptionItemModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate entries
	if !utils.IsKnown(data.Entries) {
		return
	}

	var entries []EntryModel
	resp.Diagnostics.Append(data.Entries.ElementsAs(ctx, &entries, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i, entry := range entries {
		validateEntry(ctx, entry, i, &resp.Diagnostics, "entries")
	}
}

// validateEntry validates a single entry based on its type
func validateEntry(ctx context.Context, entry EntryModel, index int, diags *diag.Diagnostics, path string) {
	if !utils.IsKnown(entry.Type) {
		return
	}

	entryType := entry.Type.ValueString()
	entryPath := fmt.Sprintf("%s[%d]", path, index)

	switch entryType {
	case entryTypeMatch, entryTypeWildcard:
		// 'value' is required (only validate if not unknown)
		if entry.Value.IsNull() {
			diags.AddError(
				"Missing Required Field",
				fmt.Sprintf("Entry type '%s' requires 'value' to be set at %s.", entryType, entryPath),
			)
		}
		// 'operator' is required (only validate if not unknown)
		if entry.Operator.IsNull() {
			diags.AddError(
				"Missing Required Field",
				fmt.Sprintf("Entry type '%s' requires 'operator' to be set at %s.", entryType, entryPath),
			)
		}

	case entryTypeMatchAny:
		// 'values' is required (only validate if not unknown)
		if entry.Values.IsNull() {
			diags.AddError(
				"Missing Required Field",
				fmt.Sprintf("Entry type '%s' requires 'values' to be set at %s.", entryTypeMatchAny, entryPath),
			)
		}
		// 'operator' is required (only validate if not unknown)
		if entry.Operator.IsNull() {
			diags.AddError(
				"Missing Required Field",
				fmt.Sprintf("Entry type '%s' requires 'operator' to be set at %s.", entryTypeMatchAny, entryPath),
			)
		}

	case entryTypeList:
		// 'list' object is required (only validate if not unknown)
		if entry.List.IsNull() {
			diags.AddError(
				"Missing Required Field",
				fmt.Sprintf("Entry type '%s' requires 'list' object to be set at %s.", entryTypeList, entryPath),
			)
		} else if !entry.List.IsUnknown() {
			// Only validate list contents if the list object itself is known
			var listModel EntryListModel
			d := entry.List.As(ctx, &listModel, basetypes.ObjectAsOptions{})
			if d.HasError() {
				diags.Append(d...)
			} else {
				// Only validate if the values are not unknown
				if listModel.ID.IsNull() {
					diags.AddError(
						"Missing Required Field",
						fmt.Sprintf("Entry type '%s' requires 'list.id' to be set at %s.", entryTypeList, entryPath),
					)
				}

				if listModel.Type.IsNull() {
					diags.AddError(
						"Missing Required Field",
						fmt.Sprintf("Entry type '%s' requires 'list.type' to be set at %s.", entryTypeList, entryPath),
					)
				}
			}
		}
		// 'operator' is required (only validate if not unknown)
		if entry.Operator.IsNull() {
			diags.AddError(
				"Missing Required Field",
				fmt.Sprintf("Entry type '%s' requires 'operator' to be set at %s.", entryTypeList, entryPath),
			)
		}

	case entryTypeExists:
		// Only 'field' and 'operator' are required (already handled by schema)
		// 'operator' is required (only validate if not unknown)
		if entry.Operator.IsNull() {
			diags.AddError(
				"Missing Required Field",
				fmt.Sprintf("Entry type '%s' requires 'operator' to be set at %s.", entryTypeExists, entryPath),
			)
		}

	case entryTypeNested:
		// 'entries' is required for nested type (only validate if not unknown)
		if entry.Entries.IsNull() {
			diags.AddError(
				"Missing Required Field",
				fmt.Sprintf("Entry type '%s' requires 'entries' to be set at %s.", entryTypeNested, entryPath),
			)
			return
		}

		// Skip validation if entries are unknown
		if entry.Entries.IsUnknown() {
			return
		}

		// 'operator' should NOT be set for nested type
		if utils.IsKnown(entry.Operator) {
			diags.AddWarning(
				"Ignored Field",
				fmt.Sprintf("Entry type '%s' does not support 'operator'. This field will be ignored at %s.", entryTypeNested, entryPath),
			)
		}

		// Validate nested entries
		var nestedEntries []NestedEntryModel
		d := entry.Entries.ElementsAs(ctx, &nestedEntries, false)
		if d.HasError() {
			diags.Append(d...)
			return
		}

		for j, nestedEntry := range nestedEntries {
			validateNestedEntry(ctx, nestedEntry, j, diags, fmt.Sprintf("%s.entries", entryPath))
		}
	}
}

// validateNestedEntry validates a nested entry within a "nested" type entry
func validateNestedEntry(ctx context.Context, entry NestedEntryModel, index int, diags *diag.Diagnostics, path string) {
	if !utils.IsKnown(entry.Type) {
		return
	}

	entryType := entry.Type.ValueString()
	entryPath := fmt.Sprintf("%s[%d]", path, index)

	// Nested entries can only be: match, match_any, or exists
	switch entryType {
	case entryTypeMatch:
		// 'value' is required (only validate if not unknown)
		if entry.Value.IsNull() {
			diags.AddError(
				"Missing Required Field",
				fmt.Sprintf("Nested entry type '%s' requires 'value' to be set at %s.", entryTypeMatch, entryPath),
			)
		}

	case entryTypeMatchAny:
		// 'values' is required (only validate if not unknown)
		if entry.Values.IsNull() {
			diags.AddError(
				"Missing Required Field",
				fmt.Sprintf("Nested entry type '%s' requires 'values' to be set at %s.", entryTypeMatchAny, entryPath),
			)
		}

	case entryTypeExists:
		// Only 'field' and 'operator' are required (already handled by schema)
		// Nothing additional to validate

	default:
		diags.AddError(
			"Invalid Entry Type",
			fmt.Sprintf("Nested entry at %s has invalid type '%s'. Only 'match', 'match_any', and 'exists' are allowed for nested entries.", entryPath, entryType),
		)
	}

	// 'operator' is always required for nested entries (only validate if not unknown)
	if entry.Operator.IsNull() {
		diags.AddError(
			"Missing Required Field",
			fmt.Sprintf("Nested entry requires 'operator' to be set at %s.", entryPath),
		)
	}
}
