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
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DecodeAnyToAttrValue converts a dynamic value decoded from JSON (string, bool,
// float64/int64/json.Number, []any, map[string]any, or nil) into the attr.Value
// that attr type t represents, recursing into List, Set and Object element types.
//
// The following scalar coercions are applied consistently regardless of caller:
//   - Int64Type accepts int/int64/float64/json.Number, or a base-10 numeric string.
//   - BoolType accepts a native bool, or a string parsed via strconv.ParseBool.
//   - StringType and jsontypes.NormalizedType require a native string; no coercion
//     is applied since callers rely on this to reject unexpected shapes.
func DecodeAnyToAttrValue(ctx context.Context, t attr.Type, raw any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	if raw == nil {
		return NullValueForType(t), diags
	}

	switch {
	case t.Equal(types.StringType):
		s, ok := raw.(string)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected string, got %T", raw))
			return types.StringUnknown(), diags
		}
		return types.StringValue(s), diags
	case t.Equal(types.BoolType):
		b, ok := coerceBool(raw)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected bool, got %T", raw))
			return types.BoolUnknown(), diags
		}
		return types.BoolValue(b), diags
	case t.Equal(types.Int64Type):
		n, ok := coerceInt64(raw)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected number, got %T", raw))
			return types.Int64Unknown(), diags
		}
		return types.Int64Value(n), diags
	}

	if _, ok := t.(jsontypes.NormalizedType); ok {
		s, ok := raw.(string)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected JSON string, got %T", raw))
			return jsontypes.NewNormalizedUnknown(), diags
		}
		return jsontypes.NewNormalizedValue(s), diags
	}
	if ty, ok := t.(types.ListType); ok {
		return decodeList(ctx, ty, raw)
	}
	if ty, ok := t.(types.SetType); ok {
		return decodeSet(ctx, ty, raw)
	}
	if ty, ok := t.(types.ObjectType); ok {
		return decodeObject(ctx, ty, raw)
	}

	diags.AddError("Internal error", fmt.Sprintf("unsupported attr type %T", t))
	return nil, diags
}

// NullValueForType returns the null attr.Value for t. Used when a decoded JSON map
// is missing a key that a target attr.Type expects.
func NullValueForType(t attr.Type) attr.Value {
	switch ty := t.(type) {
	case types.ListType:
		return types.ListNull(ty.ElemType)
	case types.SetType:
		return types.SetNull(ty.ElemType)
	case types.ObjectType:
		return types.ObjectNull(ty.AttrTypes)
	}
	if t.Equal(types.StringType) {
		return types.StringNull()
	}
	if t.Equal(types.Int64Type) {
		return types.Int64Null()
	}
	if t.Equal(types.BoolType) {
		return types.BoolNull()
	}
	if _, ok := t.(jsontypes.NormalizedType); ok {
		return jsontypes.NewNormalizedNull()
	}
	return types.StringNull()
}

// decodeList handles two distinct shapes used by callers:
//   - List of Object: raw is expected to be a single-element []any wrapping a
//     map[string]any (e.g. an ILM action toggle). A missing or non-array raw value
//     is treated as an absent/optional feature and decodes to null rather than an
//     error.
//   - List of anything else (e.g. scalar settings): raw must be a []any; every
//     element is decoded against ElemType and a non-array raw value is an error,
//     since callers rely on this to catch API/schema mismatches.
func decodeList(ctx context.Context, ty types.ListType, raw any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if elemOT, ok := ty.ElemType.(types.ObjectType); ok {
		slice, ok := raw.([]any)
		if !ok || len(slice) == 0 {
			return types.ListNull(ty.ElemType), diags
		}
		obj, d := decodeObject(ctx, elemOT, slice[0])
		diags.Append(d...)
		if diags.HasError() {
			return types.ListUnknown(ty.ElemType), diags
		}
		lv, d := types.ListValue(ty.ElemType, []attr.Value{obj})
		diags.Append(d...)
		return lv, diags
	}

	slice, ok := raw.([]any)
	if !ok {
		diags.AddError("Type mismatch", fmt.Sprintf("expected list, got %T", raw))
		return types.ListUnknown(ty.ElemType), diags
	}
	elems, d := decodeElements(ctx, ty.ElemType, slice)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListUnknown(ty.ElemType), diags
	}
	lv, d := types.ListValue(ty.ElemType, elems)
	diags.Append(d...)
	return lv, diags
}

// decodeSet decodes a []any into a Set, applying the same per-element decoding as
// the scalar branch of decodeList. A non-array raw value is an error.
func decodeSet(ctx context.Context, ty types.SetType, raw any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	slice, ok := raw.([]any)
	if !ok {
		diags.AddError("Type mismatch", fmt.Sprintf("expected list, got %T", raw))
		return types.SetUnknown(ty.ElemType), diags
	}
	elems, d := decodeElements(ctx, ty.ElemType, slice)
	diags.Append(d...)
	if diags.HasError() {
		return types.SetUnknown(ty.ElemType), diags
	}
	sv, d := types.SetValue(ty.ElemType, elems)
	diags.Append(d...)
	return sv, diags
}

func decodeElements(ctx context.Context, elemType attr.Type, slice []any) ([]attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	elems := make([]attr.Value, 0, len(slice))
	for _, item := range slice {
		v, d := DecodeAnyToAttrValue(ctx, elemType, item)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		elems = append(elems, v)
	}
	return elems, diags
}

// decodeObject decodes raw into an Object with the given AttrTypes. raw is normally
// a map[string]any, but a single-element []any wrapping a map is also accepted
// (e.g. an ILM action represented as a one-item list of its attributes). Keys
// absent from raw decode to their type's null value rather than an error.
func decodeObject(ctx context.Context, ty types.ObjectType, raw any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if slice, ok := raw.([]any); ok {
		if len(slice) == 0 {
			return types.ObjectNull(ty.AttrTypes), diags
		}
		raw = slice[0]
	}

	m, ok := raw.(map[string]any)
	if !ok {
		diags.AddError("Type mismatch", fmt.Sprintf("expected object map, got %T", raw))
		return types.ObjectUnknown(ty.AttrTypes), diags
	}

	attrs := make(map[string]attr.Value, len(ty.AttrTypes))
	for k, elemT := range ty.AttrTypes {
		rawVal, ok := m[k]
		if !ok || rawVal == nil {
			attrs[k] = NullValueForType(elemT)
			continue
		}
		v, d := DecodeAnyToAttrValue(ctx, elemT, rawVal)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectUnknown(ty.AttrTypes), diags
		}
		attrs[k] = v
	}

	ov, d := types.ObjectValue(ty.AttrTypes, attrs)
	diags.Append(d...)
	return ov, diags
}

// coerceInt64 converts int/int64/float64/json.Number, or a base-10 numeric string,
// to int64.
func coerceInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
case float64:
		if n < float64(-1<<63) || n >= float64(1<<63) || n != float64(int64(n)) {
			return 0, false
		}
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		return i, err == nil
	default:
		return 0, false
	}
}

// coerceBool converts a native bool, or a string parsed via strconv.ParseBool, to bool.
func coerceBool(v any) (bool, bool) {
	switch b := v.(type) {
	case bool:
		return b, true
	case string:
		parsed, err := strconv.ParseBool(b)
		return parsed, err == nil
	default:
		return false, false
	}
}
