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

// StringFromAny converts an any value to a Terraform types.String, returning
// types.StringNull() when v is nil or not a string.
func StringFromAny(v any) types.String {
	if str, ok := v.(string); ok {
		return types.StringValue(str)
	}
	return types.StringNull()
}

// BoolFromAny converts an any value to a Terraform types.Bool, returning
// types.BoolNull() when v is nil or not a bool.
func BoolFromAny(v any) types.Bool {
	if b, ok := v.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}

// Int32FromAnyFloat64 converts an any value holding a float64 (as produced by JSON
// decoding into an any-typed field) to a Terraform types.Int32, returning
// types.Int32Null() when v is nil or not a float64.
func Int32FromAnyFloat64(v any) types.Int32 {
	if f, ok := v.(float64); ok {
		return types.Int32Value(int32(f))
	}
	return types.Int32Null()
}

// Int64FromAnyFloat64 converts an any value holding a float64 (as produced by JSON
// decoding into an any-typed field) to a Terraform types.Int64, returning
// types.Int64Null() when v is nil or not a float64.
func Int64FromAnyFloat64(v any) types.Int64 {
	if f, ok := v.(float64); ok {
		return types.Int64Value(int64(f))
	}
	return types.Int64Null()
}
