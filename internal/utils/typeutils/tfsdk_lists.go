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

// NonEmptyListOrDefault returns the original list if slice is empty,
// otherwise converts slice into a types.List.
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

// ListTypeAs converts a types.List into a tfsdk aware []T.
func ListTypeAs[T any](ctx context.Context, value types.List, p path.Path, diags *diag.Diagnostics) []T {
	return elementsAs[[]T](ctx, value, p, diags)
}

// ListValueFrom converts a tfsdk aware []T to a types.List.
func ListValueFrom[T any](ctx context.Context, value []T, elemType attr.Type, p path.Path, diags *diag.Diagnostics) types.List {
	list, d := types.ListValueFrom(ctx, elemType, value)
	diags.Append(convertToAttrDiags(d, p)...)
	return list
}
