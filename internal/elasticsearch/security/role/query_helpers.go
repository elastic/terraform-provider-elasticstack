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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// marshalIndexQuery converts an Elasticsearch index query union value to a jsontypes.Normalized.
// String values are treated as pre-serialized JSON and passed through unchanged.
func marshalIndexQuery(query any) (jsontypes.Normalized, diag.Diagnostics) {
	var diags diag.Diagnostics
	if q, ok := query.(string); ok {
		return jsontypes.NewNormalizedValue(q), diags
	}
	return typeutils.MarshalToNormalized(query, path.Root("query"), &diags), diags
}
