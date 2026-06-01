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

// ValueStringPointer returns nil if unknown, otherwise the same as value.ValueStringPointer().
// Useful for computed optional fields without a default value, as these unknown values
// return a pointer to an empty string.
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

// OptStringPtr returns nil if the value is null or unknown, otherwise returns a pointer to the string value.
func OptStringPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueStringPointer()
}

// OptionalBool returns a pointer to the bool value when set, or nil when null or unknown.
func OptionalBool(value types.Bool) *bool {
	if !IsKnown(value) {
		return nil
	}
	v := value.ValueBool()
	return &v
}

// OptionalString returns a pointer to the string value when set and non-empty, or nil otherwise.
func OptionalString(value types.String) *string {
	if !IsKnown(value) || value.ValueString() == "" {
		return nil
	}
	v := value.ValueString()
	return &v
}

// BoolPointerValue converts a *bool to a types.Bool, returning types.BoolNull() when the pointer is nil.
func BoolPointerValue(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*v)
}

// NonEmptyStringOrNull returns types.StringValue(*s) when s is non-nil and non-empty,
// and types.StringNull() otherwise. Use for API fields that use an empty string to signal absence.
func NonEmptyStringOrNull(s *string) types.String {
	if s != nil && *s != "" {
		return types.StringValue(*s)
	}
	return types.StringNull()
}
