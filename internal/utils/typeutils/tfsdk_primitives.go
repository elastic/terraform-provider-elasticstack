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

import "github.com/hashicorp/terraform-plugin-framework/types"

// ValueStringPointer returns nil when the value is unknown or null, and &"" when empty.
// Use for computed optional fields that must distinguish null from empty string at the API level.
// For fields where empty string means "not set", prefer OptionalString instead.
func ValueStringPointer(value types.String) *string {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueStringPointer()
}

// Float64PointerValue returns nil if unknown, otherwise the same as value.ValueFloat64Pointer().
func Float64PointerValue(value types.Float64) *float64 {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueFloat64Pointer()
}

// OptionalBool returns a pointer to the bool value when set, or nil when null or unknown.
func OptionalBool(value types.Bool) *bool {
	if !IsKnown(value) {
		return nil
	}
	v := value.ValueBool()
	return &v
}

// OptionalString returns a non-nil pointer only when the value is known and non-empty.
// Returns nil when the value is null, unknown, or an empty string.
// Use for optional API string fields where empty string and absent are equivalent.
func OptionalString(value types.String) *string {
	if !IsKnown(value) || value.ValueString() == "" {
		return nil
	}
	v := value.ValueString()
	return &v
}

// OptionalInt returns a pointer to the int value when set, or nil when null or unknown.
// Use when the target API field expects *int rather than *int64.
func OptionalInt(value types.Int64) *int {
	if !IsKnown(value) {
		return nil
	}
	v := int(value.ValueInt64())
	return &v
}

// Int64Pointer returns a pointer to the int64 value when set, or nil when null or unknown.
func Int64Pointer(v types.Int64) *int64 {
	if !IsKnown(v) {
		return nil
	}
	val := v.ValueInt64()
	return &val
}

// Int64ToFloat32Ptr converts a types.Int64 to a *float32, returning nil when null or unknown.
func Int64ToFloat32Ptr(v types.Int64) *float32 {
	if p := Int64Pointer(v); p != nil {
		val := float32(*p)
		return &val
	}
	return nil
}

// IntPointerToInt64Value converts a *int to a types.Int64, returning types.Int64Null() when the pointer is nil.
func IntPointerToInt64Value(v *int) types.Int64 {
	return types.Int64PointerValue(Itol(v))
}

// Float32PointerToFloat64Value converts a *float32 to a types.Float64, returning types.Float64Null() when the pointer is nil.
func Float32PointerToFloat64Value(v *float32) types.Float64 {
	if v == nil {
		return types.Float64Null()
	}
	return types.Float64Value(float64(*v))
}

// Float32PointerToInt64Pointer converts a *float32 to a *int64, truncating any fractional part.
// Intended for numeric API fields that represent whole-number counts (e.g. a maximum series/points-to-plot value).
func Float32PointerToInt64Pointer(v *float32) *int64 {
	if v == nil {
		return nil
	}
	val := int64(*v)
	return &val
}

// NonEmptyStringOrNull returns types.StringValue(*s) when s is non-nil and non-empty,
// and types.StringNull() otherwise. Use for API fields that use an empty string to signal absence.
func NonEmptyStringOrNull(s *string) types.String {
	if s != nil && *s != "" {
		return types.StringValue(*s)
	}
	return types.StringNull()
}
