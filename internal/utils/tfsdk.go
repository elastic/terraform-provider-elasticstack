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

// SliceToListType_String converts a tfsdk naive []string into a types.List.
// This is a shorthand SliceToListType helper for strings.
func SliceToListType_String(ctx context.Context, value []string, path path.Path, diags diag.Diagnostics) types.List {
	return SliceToListType(ctx, value, types.StringType, path, diags, types.StringValue)
}

// ListTypeToSlice converts a types.List first into a tfsdk aware []T1 and transforms
// the result into a []T2.
func ListTypeToSlice[T1 any, T2 any](ctx context.Context, value types.List, path path.Path, diags diag.Diagnostics, iteratee func(item T1) T2) []T2 {
	if !IsKnown(value) {
		return nil
	}

	elems := ListTypeAs[T1](ctx, value, path, diags)
	if diags.HasError() {
		return nil
	}

	return TransformSlice(elems, iteratee)
}

// ListTypeToSlice_String converts a types.List into a []string.
// This is a shorthand ListTypeToSlice helper for strings.
func ListTypeToSlice_String(ctx context.Context, value types.List, path path.Path, diags diag.Diagnostics) []string {
	return ListTypeToSlice(ctx, value, path, diags, func(item types.String) string {
		return item.ValueString()
	})
}

// ListTypeAs converts a types.List into a tfsdk aware []T.
func ListTypeAs[T any](ctx context.Context, value types.List, path path.Path, diags diag.Diagnostics) []T {
	if !IsKnown(value) {
		return nil
	}

	var items []T
	nd := value.ElementsAs(ctx, &items, false)
	diags.Append(ConvertToAttrDiags(nd, path)...)
	return items
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
