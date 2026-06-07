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

// BoolPointerValue converts a *bool to a types.Bool, returning types.BoolNull() when the pointer is nil.
func BoolPointerValue(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*v)
}

// Int64PointerValue converts a *int64 to a types.Int64, returning types.Int64Null() when the pointer is nil.
func Int64PointerValue(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

// NonEmptyStringOrNull returns types.StringValue(*s) when s is non-nil and non-empty,
// and types.StringNull() otherwise. Use for API fields that use an empty string to signal absence.
func NonEmptyStringOrNull(s *string) types.String {
	if s != nil && *s != "" {
		return types.StringValue(*s)
	}
	return types.StringNull()
}
