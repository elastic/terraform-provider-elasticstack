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
			// Wrapped through an error-typed variable, not passed inline: vet's printf
			// check flags %w on a *types.ElasticsearchError because ElasticsearchError's
			// Error() has a value receiver, but the pointer is intentional here since
			// that's the type the client actually decodes errors into.
			name:     "nil *ElasticsearchError in chain returns false",
			err:      fmt.Errorf("wrap: %w", error((*types.ElasticsearchError)(nil))),
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
			err:      fmt.Errorf("wrapped: %w", error(&types.ElasticsearchError{Status: 404})),
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

func TestDiagsOrNotFound(t *testing.T) {
	t.Run("nil error returns no diagnostics", func(t *testing.T) {
		assert.False(t, DiagsOrNotFound(nil).HasError())
	})

	t.Run("404 error returns no diagnostics", func(t *testing.T) {
		assert.False(t, DiagsOrNotFound(&types.ElasticsearchError{Status: 404}).HasError())
	})

	t.Run("other error is wrapped into diagnostics", func(t *testing.T) {
		diags := DiagsOrNotFound(&types.ElasticsearchError{Status: 500})
		assert.True(t, diags.HasError())
	})
}

func TestDeleteWithNotFoundAsSuccess(t *testing.T) {
	t.Run("nil error returns no diagnostics", func(t *testing.T) {
		assert.False(t, DeleteWithNotFoundAsSuccess(nil, "Unable to delete thing").HasError())
	})

	t.Run("404 error returns no diagnostics", func(t *testing.T) {
		assert.False(t, DeleteWithNotFoundAsSuccess(&types.ElasticsearchError{Status: 404}, "Unable to delete thing").HasError())
	})

	t.Run("other error is wrapped into diagnostics using the given summary", func(t *testing.T) {
		diags := DeleteWithNotFoundAsSuccess(&types.ElasticsearchError{Status: 500}, "Unable to delete thing")
		assert.True(t, diags.HasError())
		assert.Equal(t, "Unable to delete thing", diags[0].Summary())
	})
}

func TestCallOrNotFound(t *testing.T) {
	t.Run("returns the result on success", func(t *testing.T) {
		result, diags := CallOrNotFound(func() (string, error) {
			return "value", nil
		})
		assert.False(t, diags.HasError())
		assert.Equal(t, "value", result)
	})

	t.Run("returns zero value and no diagnostics on 404", func(t *testing.T) {
		result, diags := CallOrNotFound(func() (string, error) {
			return "ignored", &types.ElasticsearchError{Status: 404}
		})
		assert.False(t, diags.HasError())
		assert.Empty(t, result)
	})

	t.Run("returns zero value and diagnostics on other errors", func(t *testing.T) {
		result, diags := CallOrNotFound(func() (string, error) {
			return "ignored", &types.ElasticsearchError{Status: 500}
		})
		assert.True(t, diags.HasError())
		assert.Empty(t, result)
	})
}
