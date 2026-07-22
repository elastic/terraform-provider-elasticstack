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

package typeutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// --- []types.String <-> []string helpers (no context required) ---

// ValueStringSlice converts []types.String to []string.
func ValueStringSlice(v []types.String) []string {
	var res []string
	for _, s := range v {
		res = append(res, s.ValueString())
	}
	return res
}

// StringSliceValue converts []string to []types.String.
func StringSliceValue(v []string) []types.String {
	var res []types.String
	for _, s := range v {
		res = append(res, types.StringValue(s))
	}
	return res
}

// --- Must-style helpers (panic on failure, no context required) ---

// StringsToListMust converts a []string to a types.List of string elements.
// The result is always valid; an empty or nil slice produces an empty list.
// This is a Must-style variant that panics on error; use SliceToListTypeString
// for safe, diagnostic-based conversion.
func StringsToListMust(strs []string) types.List {
	if len(strs) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}
	vals := make([]attr.Value, len(strs))
	for i, s := range strs {
		vals[i] = types.StringValue(s)
	}
	return types.ListValueMust(types.StringType, vals)
}

// ListToStringsMust extracts a []string from a types.List of string elements.
// Null or unknown lists return nil.
// This is a Must-style variant that panics on non-string elements; use
// ListTypeToSliceString for safe, diagnostic-based extraction.
func ListToStringsMust(list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	elems := list.Elements()
	strs := make([]string, len(elems))
	for i, v := range elems {
		strs[i] = v.(types.String).ValueString()
	}
	return strs
}

// NonEmptyListOrDefault returns the original list if slice is empty,
// otherwise converts slice into a types.List.
func NonEmptyListOrDefault[T any](ctx context.Context, original types.List, elemType attr.Type, slice []T) (types.List, diag.Diagnostics) {
	return nonEmptyCollectionOrDefault(ctx, original, elemType, slice, types.ListValueFrom)
}

// EnsureTypedList converts untyped zero-value lists to properly typed null lists.
// This is commonly needed during import operations where the framework may create
// untyped lists with DynamicPseudoType elements, which causes type conversion errors.
// If the list already has a proper type, it is returned unchanged.
func EnsureTypedList(ctx context.Context, list types.List, elemType attr.Type) types.List {
	if list.ElementType(ctx) == nil {
		return types.ListNull(elemType)
	}

	if _, ok := list.ElementType(ctx).(basetypes.DynamicType); ok {
		return types.ListNull(elemType)
	}

	return list
}

// SliceToListType converts a tfsdk naive []T1 into an types.List of []T2.
// This handles both structs and simple types to attr.Values.
func SliceToListType[T1 any, T2 any](ctx context.Context, value []T1, elemType attr.Type, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta ListMeta) T2) types.List {
	if value == nil {
		return types.ListNull(elemType)
	}

	elems := TransformSlice(ctx, value, p, diags, iteratee)
	list, nd := types.ListValueFrom(ctx, elemType, elems)
	diags.Append(convertToAttrDiags(nd, p)...)

	return list
}

// SliceToListTypeString converts a tfsdk naive []string into a types.List.
// This is a shorthand SliceToListType helper for strings.
func SliceToListTypeString(ctx context.Context, value []string, p path.Path, diags *diag.Diagnostics) types.List {
	return ListValueFrom(ctx, value, types.StringType, p, diags)
}

// ListTypeToMap converts a types.List first into a tfsdk aware []T1
// and transforms the result into a map[string]T2.
func ListTypeToMap[T1 any, T2 any](ctx context.Context, value types.List, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta ListMeta) (key string, elem T2)) map[string]T2 {
	if !IsKnown(value) {
		return nil
	}

	items := ListTypeAs[T1](ctx, value, p, diags)
	if diags.HasError() {
		return nil
	}

	return TransformSliceToMap(ctx, items, p, diags, iteratee)
}

// ListTypeToSlice converts a types.List first into a tfsdk aware []T1 and transforms
// the result into a []T2.
func ListTypeToSlice[T1 any, T2 any](ctx context.Context, value types.List, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta ListMeta) T2) []T2 {
	if !IsKnown(value) {
		return nil
	}

	elems := ListTypeAs[T1](ctx, value, p, diags)
	if diags.HasError() {
		return nil
	}

	return TransformSlice(ctx, elems, p, diags, iteratee)
}

// ListTypeToSliceString converts a types.List into a []string.
// This is a shorthand ListTypeToSlice helper for strings.
func ListTypeToSliceString(ctx context.Context, value types.List, p path.Path, diags *diag.Diagnostics) []string {
	return ListTypeAs[string](ctx, value, p, diags)
}

// ListTypeToSliceStringPtr extracts a *[]string from an optional list attribute,
// returning nil when the list is null or unknown.
func ListTypeToSliceStringPtr(ctx context.Context, l types.List, p path.Path, diags *diag.Diagnostics) *[]string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	result := ListTypeToSliceString(ctx, l, p, diags)
	if diags.HasError() {
		return nil
	}
	return &result
}

// ListTypeAs converts a types.List into a tfsdk aware []T.
func ListTypeAs[T any](ctx context.Context, value types.List, p path.Path, diags *diag.Diagnostics) []T {
	return elementsAs[[]T](ctx, value, p, diags)
}

// ListValueFrom converts a tfsdk aware []T to a types.List.
func ListValueFrom[T any](ctx context.Context, value []T, elemType attr.Type, p path.Path, diags *diag.Diagnostics) types.List {
	return collectionValueFrom(ctx, value, elemType, p, diags, types.ListValueFrom)
}
