package utils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SliceToListType converts a tfsdk naive []T1 into an types.List of []T2.
// This handles both structs and simple types to attr.Values.
func SliceToListType[T1 any, T2 any](ctx context.Context, value []T1, elemType attr.Type, path path.Path, diags diag.Diagnostics, iteratee func(item T1) T2) types.List {
	if value == nil {
		return types.ListNull(elemType)
	}

	elems := TransformSlice(value, iteratee)
	list, nd := types.ListValueFrom(ctx, elemType, elems)
	diags.Append(ConvertToAttrDiags(nd, path)...)

	return list
}

// TransformSlice converts []T1 to []T2 via the iteratee.
func TransformSlice[T1 any, T2 any](value []T1, iteratee func(item T1) T2) []T2 {
	if value == nil {
		return nil
	}

	elems := make([]T2, len(value))
	for i, v := range value {
		elems[i] = iteratee(v)
	}

	return elems
}
