package utils

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ListMeta struct {
	Context context.Context
	Index   int
	Path    path.Path
	Diags   *diag.Diagnostics
}

type MapMeta struct {
	Context context.Context
	Key     string
	Path    path.Path
	Diags   *diag.Diagnostics
}

type ObjectMeta struct {
	Context context.Context
	Path    path.Path
	Diags   *diag.Diagnostics
}

// ======================
// ===== Primitives =====
// ======================

// ValueStringPointer returns nil if unknown, otherwise the same as value.ValueStringPointer().
// Useful for computed optional fields without a default value, as these unknown values
// return a pointer to an empty string.
func ValueStringPointer(value types.String) *string {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueStringPointer()
}

// ================
// ===== Maps =====
// ================

// MapToNormalizedType marshals a map[string]T into a jsontypes.Normalized.
func MapToNormalizedType[T any](value map[string]T, p path.Path, diags *diag.Diagnostics) jsontypes.Normalized {
	if value == nil {
		return jsontypes.NewNormalizedNull()
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		diags.AddAttributeError(p, "marshal failure", err.Error())
	}

	return jsontypes.NewNormalizedValue(string(bytes))
}

// NormalizedTypeToMap unmarshals a jsontypes.Normalized to a map[string]T.
func NormalizedTypeToMap[T any](value jsontypes.Normalized, p path.Path, diags *diag.Diagnostics) map[string]T {
	if !IsKnown(value) {
		return nil
	}

	var dest map[string]T
	d := value.Unmarshal(&dest)
	diags.Append(ConvertToAttrDiags(d, p)...)
	return dest
}

// MapToMapType converts a tfsdk naive map[string]T1 into an types.Map of map[string]T2.
// This handles both structs and simple types to attr.Values.
func MapToMapType[T1 any, T2 any](ctx context.Context, value map[string]T1, elemType attr.Type, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta MapMeta) T2) types.Map {
	if value == nil {
		return types.MapNull(elemType)
	}

	elems := TransformMap(ctx, value, p, diags, iteratee)
	mapping, d := types.MapValueFrom(ctx, elemType, elems)
	diags.Append(ConvertToAttrDiags(d, p)...)

	return mapping
}

// MapTypeToMap converts a types.Map first into a tfsdk aware map[string]T1 and transforms
// the result into a map[string]T2.
func MapTypeToMap[T1 any, T2 any](ctx context.Context, value types.Map, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta MapMeta) T2) map[string]T2 {
	if !IsKnown(value) {
		return nil
	}

	elems := MapTypeAs[T1](ctx, value, p, diags)
	if diags.HasError() {
		return nil
	}

	return TransformMap(ctx, elems, p, diags, iteratee)
}

// MapTypeAs converts a types.Map into a tfsdk aware map[string]T.
func MapTypeAs[T any](ctx context.Context, value types.Map, p path.Path, diags *diag.Diagnostics) map[string]T {
	if !IsKnown(value) {
		return nil
	}

	var items map[string]T
	d := value.ElementsAs(ctx, &items, false)
	diags.Append(ConvertToAttrDiags(d, p)...)
	return items
}

// MapValueFrom converts a tfsdk aware map[string]T to a types.Map.
func MapValueFrom[T any](ctx context.Context, value map[string]T, elemType attr.Type, p path.Path, diags *diag.Diagnostics) types.Map {
	mapping, d := types.MapValueFrom(ctx, elemType, value)
	diags.Append(ConvertToAttrDiags(d, p)...)
	return mapping
}

// =================
// ===== Lists =====
// =================

// SliceToListType converts a tfsdk naive []T1 into an types.List of []T2.
// This handles both structs and simple types to attr.Values.
func SliceToListType[T1 any, T2 any](ctx context.Context, value []T1, elemType attr.Type, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta ListMeta) T2) types.List {
	if value == nil {
		return types.ListNull(elemType)
	}

	elems := TransformSlice(ctx, value, p, diags, iteratee)
	list, nd := types.ListValueFrom(ctx, elemType, elems)
	diags.Append(ConvertToAttrDiags(nd, p)...)

	return list
}

// SliceToListType_String converts a tfsdk naive []string into a types.List.
// This is a shorthand SliceToListType helper for strings.
func SliceToListType_String(ctx context.Context, value []string, p path.Path, diags *diag.Diagnostics) types.List {
	return SliceToListType(ctx, value, types.StringType, p, diags,
		func(item string, meta ListMeta) types.String {
			return types.StringValue(item)
		})
}

// ListTypeToMap converts a types.List first into a tfsdk aware map[string]T1
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

// ListTypeToSlice_String converts a types.List into a []string.
// This is a shorthand ListTypeToSlice helper for strings.
func ListTypeToSlice_String(ctx context.Context, value types.List, p path.Path, diags *diag.Diagnostics) []string {
	return ListTypeToSlice(ctx, value, p, diags, func(item types.String, meta ListMeta) string {
		return item.ValueString()
	})
}

// ListTypeAs converts a types.List into a tfsdk aware []T.
func ListTypeAs[T any](ctx context.Context, value types.List, p path.Path, diags *diag.Diagnostics) []T {
	if !IsKnown(value) {
		return nil
	}

	var items []T
	nd := value.ElementsAs(ctx, &items, false)
	diags.Append(ConvertToAttrDiags(nd, p)...)
	return items
}

// ListValueFrom converts a tfsdk aware []T to a types.List.
func ListValueFrom[T any](ctx context.Context, value []T, elemType attr.Type, p path.Path, diags *diag.Diagnostics) types.List {
	list, d := types.ListValueFrom(ctx, elemType, value)
	diags.Append(ConvertToAttrDiags(d, p)...)
	return list
}

// ===================
// ===== Objects =====
// ===================

// StructToObjectType converts a tfsdk naive T1 into an types.Object of T2.
func StructToObjectType[T1 any, T2 any](ctx context.Context, value *T1, attrTypes map[string]attr.Type, p path.Path, diags *diag.Diagnostics, transformee func(item T1, meta ObjectMeta) T2) types.Object {
	if value == nil {
		return types.ObjectNull(attrTypes)
	}

	item := TransformObject(ctx, value, p, diags, transformee)
	obj, d := types.ObjectValueFrom(ctx, attrTypes, item)
	diags.Append(ConvertToAttrDiags(d, p)...)

	return obj
}

// ObjectTypeToStruct converts a types.Object first into a tfsdk aware T1 and transforms
// the result into a T2.
func ObjectTypeToStruct[T1 any, T2 any](ctx context.Context, value types.Object, p path.Path, diags *diag.Diagnostics, transformee func(item T1, meta ObjectMeta) T2) *T2 {
	if !IsKnown(value) {
		return nil
	}

	item := ObjectTypeAs[T1](ctx, value, p, diags)
	if diags.HasError() {
		return nil
	}

	return TransformObject(ctx, item, p, diags, transformee)
}

// ObjectTypeAs converts a types.Object into a tfsdk aware T.
func ObjectTypeAs[T any](ctx context.Context, value types.Object, p path.Path, diags *diag.Diagnostics) *T {
	if !IsKnown(value) {
		return nil
	}

	var item T
	d := value.As(ctx, &item, basetypes.ObjectAsOptions{})
	diags.Append(ConvertToAttrDiags(d, p)...)
	return &item
}

// ObjectValueFrom converts a tfsdk aware T to a types.Object.
func ObjectValueFrom[T any](ctx context.Context, value *T, attrTypes map[string]attr.Type, p path.Path, diags *diag.Diagnostics) types.Object {
	obj, d := types.ObjectValueFrom(ctx, attrTypes, value)
	diags.Append(ConvertToAttrDiags(d, p)...)
	return obj
}

// ======================
// ===== Transforms =====
// ======================

// TransformObject converts T1 to T2 via the transformee.
func TransformObject[T1 any, T2 any](ctx context.Context, value *T1, p path.Path, diags *diag.Diagnostics, transformee func(item T1, meta ObjectMeta) T2) *T2 {
	if value == nil {
		return nil
	}

	result := transformee(*value, ObjectMeta{Context: ctx, Path: p, Diags: diags})
	return &result
}

// TransformMap converts map[string]T1 to map[string]T2 via the iteratee.
func TransformMap[T1 any, T2 any](ctx context.Context, value map[string]T1, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta MapMeta) T2) map[string]T2 {
	if value == nil {
		return nil
	}

	elems := make(map[string]T2, len(value))
	for k, v := range value {
		elems[k] = iteratee(v, MapMeta{Context: ctx, Key: k, Path: p.AtMapKey(k), Diags: diags})
	}

	return elems
}

// TransformSlice converts []T1 to []T2 via the iteratee.
func TransformSlice[T1 any, T2 any](ctx context.Context, value []T1, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta ListMeta) T2) []T2 {
	if value == nil {
		return nil
	}

	elems := make([]T2, len(value))
	for i, v := range value {
		elems[i] = iteratee(v, ListMeta{Context: ctx, Index: i, Path: p.AtListIndex(i), Diags: diags})
	}

	return elems
}

// TransformSliceToMap converts []T1 to map[string]]T2 via the iteratee.
func TransformSliceToMap[T1 any, T2 any](ctx context.Context, value []T1, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta ListMeta) (key string, elem T2)) map[string]T2 {
	if value == nil {
		return nil
	}

	elems := make(map[string]T2, len(value))
	for i, v := range value {
		k, v := iteratee(v, ListMeta{Context: ctx, Index: i, Path: p.AtListIndex(i), Diags: diags})
		elems[k] = v
	}

	return elems
}

// TransformSliceToMap converts []T1 to map[string]]T2 via the iteratee.
func TransformMapToSlice[T1 any, T2 any](ctx context.Context, value map[string]T1, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta MapMeta) T2) []T2 {
	if value == nil {
		return nil
	}

	elems := make([]T2, 0, len(value))
	for k, v := range value {
		v := iteratee(v, MapMeta{Context: ctx, Key: k, Path: p.AtMapKey(k), Diags: diags})
		elems = append(elems, v)
	}

	return elems
}
