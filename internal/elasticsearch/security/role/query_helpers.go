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

package role

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// marshalIndexQuery converts an Elasticsearch index query union value to a jsontypes.Normalized.
func marshalIndexQuery(query any) (jsontypes.Normalized, diag.Diagnostics) {
	var diags diag.Diagnostics
	if query == nil {
		return jsontypes.NewNormalizedNull(), diags
	}
	switch q := query.(type) {
	case string:
		return jsontypes.NewNormalizedValue(q), diags
	default:
		b, err := json.Marshal(q)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling query: %s", err))
			return jsontypes.NewNormalizedNull(), diags
		}
		return jsontypes.NewNormalizedValue(string(b)), diags
	}
}
