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

package lenscommon

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// MarshalToJSONWithDefaults stores already-marshaled bytes as JSONWithDefaultsValue,
// or adds an error to diags and returns (zero, false) on failure.
func MarshalToJSONWithDefaults[T any](bytes []byte, err error, fieldName string, defaults func(T) T, diags *diag.Diagnostics) (customtypes.JSONWithDefaultsValue[T], bool) {
	if err != nil {
		diags.AddError("Failed to marshal "+fieldName, err.Error())
		return customtypes.JSONWithDefaultsValue[T]{}, false
	}
	return customtypes.NewJSONWithDefaultsValue(string(bytes), defaults), true
}
