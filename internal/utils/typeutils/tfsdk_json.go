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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// MarshalToNormalized marshals v to JSON and returns a jsontypes.Normalized.
// Returns NewNormalizedNull() when v is nil, marshaling produces "null", or an error occurs.
func MarshalToNormalized(v any, fieldDesc string, diags *diag.Diagnostics) jsontypes.Normalized {
	if v == nil {
		return jsontypes.NewNormalizedNull()
	}
	b, err := json.Marshal(v)
	if err != nil {
		diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling %s: %s", fieldDesc, err))
		return jsontypes.NewNormalizedNull()
	}
	if bytes.Equal(b, []byte("null")) {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(b))
}
