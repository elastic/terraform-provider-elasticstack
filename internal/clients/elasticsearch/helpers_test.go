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

package elasticsearch

import (
	"errors"
	"fmt"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/stretchr/testify/assert"
)

func TestIsNotFoundElasticsearchError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error returns false",
			err:      nil,
			expected: false,
		},
		{
			name:     "non-elasticsearch error returns false",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "elasticsearch 404 error returns true",
			err:      &types.ElasticsearchError{Status: 404},
			expected: true,
		},
		{
			name:     "elasticsearch 404 wrapped in another error returns true",
			err:      fmt.Errorf("wrapped: %w", &types.ElasticsearchError{Status: 404}),
			expected: true,
		},
		{
			name:     "elasticsearch 500 error returns false",
			err:      &types.ElasticsearchError{Status: 500},
			expected: false,
		},
		{
			name:     "elasticsearch 403 error returns false",
			err:      &types.ElasticsearchError{Status: 403},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsNotFoundElasticsearchError(tc.err))
		})
	}
}
