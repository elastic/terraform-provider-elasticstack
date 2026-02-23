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

package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringSliceOrCSV_UnmarshalJSON9(t *testing.T) {
	tests := []struct {
		name           string
		jsonString     string
		expectedResult StringSliceOrCSV
		expectedErr    error
	}{
		{
			name:           "should handle json arrays",
			jsonString:     `["a", "b", "c"]`,
			expectedResult: StringSliceOrCSV{"a", "b", "c"},
		},
		{
			name:           "should handle csv strings",
			jsonString:     `"a,b,c"`,
			expectedResult: StringSliceOrCSV{"a", "b", "c"},
		},
		{
			name:       "should handle explicit nulls",
			jsonString: `null`,
		},
		{
			name:       "should handle empty strings",
			jsonString: `""`,
		},
		{
			name:        "should fail on invalid data",
			jsonString:  "true",
			expectedErr: ErrInvalidStringSliceOrCSV,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualModel StringSliceOrCSV
			err := json.Unmarshal([]byte(tt.jsonString), &actualModel)

			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.expectedErr)
			}

			require.Equal(t, tt.expectedResult, actualModel)
		})
	}
}
