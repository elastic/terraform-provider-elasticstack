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

// NonEmptyStringPointerValue returns nil if the value is null, unknown, or empty string,
// otherwise returns a pointer to the string value.
func NonEmptyStringPointerValue(value types.String) *string {
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return nil
	}
	s := value.ValueString()
	return &s
}

// Float64PointerValue returns nil if unknown, otherwise the same as value.ValueFloat64Pointer().
func Float64PointerValue(value types.Float64) *float64 {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueFloat64Pointer()
}
