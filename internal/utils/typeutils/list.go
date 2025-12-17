package typeutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NonEmptyListOrDefault[T any](ctx context.Context, original types.List, elemType attr.Type, slice []T) (types.List, diag.Diagnostics) {
	if len(slice) == 0 {
		return original, nil
	}

	return types.ListValueFrom(ctx, elemType, slice)
}

// EnsureTypedList converts untyped zero-value lists to properly typed null lists.
// This is commonly needed during import operations where the framework may create
// untyped lists with DynamicPseudoType elements, which causes type conversion errors.
// If the list already has a proper type, it is returned unchanged.
func EnsureTypedList(ctx context.Context, list types.List, elemType attr.Type) types.List {
	// Check if the list has no element type (nil)
	if list.ElementType(ctx) == nil {
		return types.ListNull(elemType)
	}

	// Check if the list has a dynamic pseudo type
	if _, ok := list.ElementType(ctx).(basetypes.DynamicType); ok {
		return types.ListNull(elemType)
	}

	// List is already properly typed, return as-is
	return list
}
