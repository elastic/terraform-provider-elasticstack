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

package index

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stringIsJSONObject(t *testing.T) {
	tests := []struct {
		name                  string
		fieldVal              any
		expectedErrsToContain []string
	}{
		{
			name:     "should not return an error for a valid json object",
			fieldVal: "{}",
		},
		{
			name:     "should not return an error for a null",
			fieldVal: "null",
		},

		{
			name:     "should return an error if the field is not a string",
			fieldVal: true,
			expectedErrsToContain: []string{
				"expected type of field-name to be string",
			},
		},
		{
			name:     "should return an error if the field is valid json, but not an object",
			fieldVal: "[]",
			expectedErrsToContain: []string{
				"expected field-name to be a JSON object. Check the documentation for the expected format.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, errors := stringIsJSONObject(tt.fieldVal, "field-name")
			require.Empty(t, warnings)

			require.Len(t, errors, len(tt.expectedErrsToContain))
			for i, err := range errors {
				require.ErrorContains(t, err, tt.expectedErrsToContain[i])
			}
		})
	}
}
