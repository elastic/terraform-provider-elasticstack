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

package ml

import (
	"context"
	"errors"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteWithNotFoundAsSuccess(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		called := false
		diags := DeleteWithNotFoundAsSuccess(ctx, "ML thing", "thing-1", func() error {
			called = true
			return nil
		})
		assert.False(t, diags.HasError())
		assert.True(t, called)
	})

	t.Run("not found is treated as success", func(t *testing.T) {
		diags := DeleteWithNotFoundAsSuccess(ctx, "ML thing", "thing-1", func() error {
			return &types.ElasticsearchError{Status: 404}
		})
		assert.False(t, diags.HasError())
	})

	t.Run("other error produces a labelled diagnostic", func(t *testing.T) {
		diags := DeleteWithNotFoundAsSuccess(ctx, "ML thing", "thing-1", func() error {
			return errors.New("boom")
		})
		require.True(t, diags.HasError())
		assert.Equal(t, "Failed to delete ML thing", diags.Errors()[0].Summary())
		assert.Contains(t, diags.Errors()[0].Detail(), "thing-1")
		assert.Contains(t, diags.Errors()[0].Detail(), "boom")
	})
}

func TestReadWithNotFoundAsAbsent(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		result, found, diags := ReadWithNotFoundAsAbsent(ctx, "ML thing", "thing-1", func() (string, error) {
			return "value", nil
		})
		assert.False(t, diags.HasError())
		assert.True(t, found)
		assert.Equal(t, "value", result)
	})

	t.Run("not found is treated as absent", func(t *testing.T) {
		result, found, diags := ReadWithNotFoundAsAbsent(ctx, "ML thing", "thing-1", func() (string, error) {
			return "", &types.ElasticsearchError{Status: 404}
		})
		assert.False(t, diags.HasError())
		assert.False(t, found)
		assert.Empty(t, result)
	})

	t.Run("other error produces a labelled diagnostic", func(t *testing.T) {
		result, found, diags := ReadWithNotFoundAsAbsent(ctx, "ML thing", "thing-1", func() (string, error) {
			return "", errors.New("boom")
		})
		require.True(t, diags.HasError())
		assert.False(t, found)
		assert.Empty(t, result)
		assert.Equal(t, "Failed to get ML thing", diags.Errors()[0].Summary())
		assert.Contains(t, diags.Errors()[0].Detail(), "thing-1")
		assert.Contains(t, diags.Errors()[0].Detail(), "boom")
	})
}
