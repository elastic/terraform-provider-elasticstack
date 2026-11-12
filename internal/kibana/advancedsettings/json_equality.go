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

package advancedsettings

import (
	"encoding/json"
	"maps"
	"math/big"
	"slices"
)

func jsonValuesEqual(a, b any) bool {
	switch left := a.(type) {
	case json.Number:
		right, ok := b.(json.Number)
		return ok && numbersEqual(left, right)
	case []any:
		right, ok := b.([]any)
		return ok && slices.EqualFunc(left, right, jsonValuesEqual)
	case map[string]any:
		right, ok := b.(map[string]any)
		return ok && maps.EqualFunc(left, right, jsonValuesEqual)
	default:
		return a == b
	}
}

// Numbers are compared exactly rather than as float64 so that integers Kibana
// rounds beyond 2^53 still show as drift.
func numbersEqual(a, b json.Number) bool {
	left, leftOK := new(big.Rat).SetString(a.String())
	right, rightOK := new(big.Rat).SetString(b.String())
	return leftOK && rightOK && left.Cmp(right) == 0
}
