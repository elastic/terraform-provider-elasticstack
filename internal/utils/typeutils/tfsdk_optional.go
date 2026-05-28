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
