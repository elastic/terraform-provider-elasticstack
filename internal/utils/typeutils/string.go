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

// StringishPointerValue converts a pointer to a string-like type to a Terraform types.String value.
func StringishPointerValue[T ~string](ptr *T) types.String {
	if ptr == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*ptr))
}

// StringishValue converts a value of any string-like type T to a Terraform types.String.
func StringishValue[T ~string](value T) types.String {
	return types.StringValue(string(value))
}

func NonEmptyStringishValue[T ~string](value T) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(string(value))
}

func NonEmptyStringishPointerValue[T ~string](ptr *T) types.String {
	if ptr == nil {
		return types.StringNull()
	}
	return NonEmptyStringishValue(*ptr)
}
