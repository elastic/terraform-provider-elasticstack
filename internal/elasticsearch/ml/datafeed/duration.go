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

package datafeed

import (
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// durationPointerToString converts an Elasticsearch Duration pointer to types.String.
// Returns types.StringNull() when v is nil.
func durationPointerToString(v any) (types.String, error) {
	if v == nil {
		return types.StringNull(), nil
	}

	// When a typed nil pointer is passed as an interface value (e.g. (*types.Duration)(nil)),
	// the interface itself is non-nil. Treat these as null to avoid returning "null".
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return types.StringNull(), nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return types.StringNull(), err
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		s = string(b)
	}
	return types.StringValue(s), nil
}
