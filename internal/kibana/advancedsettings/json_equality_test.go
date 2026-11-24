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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONValuesEqual_onlyNumberFormattingIsIgnored(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		left  string
		right string
		want  bool
	}{
		"a trailing zero":                      {left: "1.50", right: "1.5", want: true},
		"a zero fraction":                      {left: "1.0", right: "1", want: true},
		"an exponent":                          {left: "1e2", right: "100", want: true},
		"a negative zero":                      {left: "-0", right: "0", want: true},
		"numbers nested in objects and arrays": {left: `{"a":[1.0,{"b":2.50}]}`, right: `{"a":[1,{"b":2.5}]}`, want: true},
		"object key order":                     {left: `{"a":1,"b":2}`, right: `{"b":2,"a":1}`, want: true},
		"different numbers":                    {left: "1.5", right: "1.6", want: false},
		"integers beyond 2^53 that differ":     {left: "9007199254740993", right: "9007199254740992", want: false},
		"a string and the number it spells":    {left: `"1"`, right: "1", want: false},
		"a string with different formatting":   {left: `"1.50"`, right: `"1.5"`, want: false},
		"different booleans":                   {left: "true", right: "false", want: false},
		"a boolean and a string":               {left: "true", right: `"true"`, want: false},
		"arrays in a different order":          {left: "[1,2]", right: "[2,1]", want: false},
		"arrays of different lengths":          {left: "[1,2]", right: "[1,2,3]", want: false},
		"objects with different keys":          {left: `{"a":1}`, right: `{"b":1}`, want: false},
		"objects with different sizes":         {left: `{"a":1}`, right: `{"a":1,"b":2}`, want: false},
		"a nested number that differs":         {left: `{"a":[1,2]}`, right: `{"a":[1,3]}`, want: false},
		"an array and an object":               {left: "[1]", right: `{"a":1}`, want: false},
		"a number and an array holding it":     {left: "1", right: "[1]", want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			left, leftErr := parseJSONKeepingNumbers(tt.left)
			right, rightErr := parseJSONKeepingNumbers(tt.right)
			require.NoError(t, leftErr)
			require.NoError(t, rightErr)

			assert.Equal(t, tt.want, jsonValuesEqual(left, right))
			assert.Equal(t, tt.want, jsonValuesEqual(right, left))
		})
	}
}
