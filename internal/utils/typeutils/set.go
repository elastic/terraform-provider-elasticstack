package typeutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NonEmptySetOrDefault[T any](ctx context.Context, original types.Set, elemType attr.Type, slice []T) (types.Set, diag.Diagnostics) {
	if len(slice) == 0 {
		// If the slice is empty, we need to decide whether to return an empty set or null/original
		// If original is a known empty set (user explicitly set []), we should return an empty set
		// If original is null (user didn't set it), we should keep it null
		if !original.IsNull() && !original.IsUnknown() {
			// Original is a known value (could be empty set), return an empty set to maintain consistency
			return types.SetValueFrom(ctx, elemType, slice)
		}
		// Original is null or unknown, preserve it
		return original, nil
	}

	return types.SetValueFrom(ctx, elemType, slice)
}

// SetValueFromOptionalComputed handles optional and computed set attributes.
// For these fields, we need to preserve user-specified empty sets while allowing
// null for unspecified fields. This function checks the original value to determine
// the correct behavior.
func SetValueFromOptionalComputed[T any](ctx context.Context, original types.Set, elemType attr.Type, slice []T) (types.Set, diag.Diagnostics) {
	// If slice has values, always return them as a set
	if len(slice) > 0 {
		return types.SetValueFrom(ctx, elemType, slice)
	}

	// Slice is empty or nil. Check the original value:
	// - If original was explicitly set (known, possibly empty), return empty set
	// - If original was null/unknown (not set by user), preserve it
	if !original.IsNull() && !original.IsUnknown() {
		// Original is a known value (user set it), return empty set
		return types.SetValueFrom(ctx, elemType, []T{})
	}

	// Original is null or unknown, preserve it
	return original, nil
}
