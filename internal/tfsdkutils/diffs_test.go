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

package tfsdkutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlattenMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  map[string]any
		out map[string]any
	}{
		{
			map[string]any{"key1": map[string]any{"key2": 1}},
			map[string]any{"key1.key2": 1},
		},
		{
			map[string]any{"key1": map[string]any{"key2": map[string]any{"key3": 1}}},
			map[string]any{"key1.key2.key3": 1},
		},
		{
			map[string]any{"key1": 1},
			map[string]any{"key1": 1},
		},
		{
			map[string]any{"key1": "test"},
			map[string]any{"key1": "test"},
		},
		{
			map[string]any{"key1": map[string]any{"key2": 1, "key3": "test", "key4": []int{1, 2, 3}}},
			map[string]any{"key1.key2": 1, "key1.key3": "test", "key1.key4": []int{1, 2, 3}},
		},
	}

	for _, tc := range tests {
		res := flattenMap(tc.in)
		require.Equal(t, tc.out, res)
	}
}
