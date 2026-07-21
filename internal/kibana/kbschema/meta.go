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

package kbschema

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// MarshalMetaToNormalized marshals an arbitrary API meta value to a jsontypes.Normalized.
// Returns Null when meta is nil or marshals to JSON null (typed nil pointer); adds an error
// diagnostic and returns Null on marshal failure.
func MarshalMetaToNormalized(meta any, diags *diag.Diagnostics) jsontypes.Normalized {
	if meta == nil {
		return jsontypes.NewNormalizedNull()
	}
	b, err := json.Marshal(meta)
	if err != nil {
		diags.AddError("Failed to marshal meta field from API response to JSON", err.Error())
		return jsontypes.NewNormalizedNull()
	}
	// A typed nil pointer (e.g. *SomeMapType(nil)) satisfies meta != nil but marshals to "null".
	// Treat JSON null the same as a nil meta — both mean the field is absent.
	if string(b) == "null" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(b))
}
