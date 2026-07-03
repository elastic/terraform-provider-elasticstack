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

package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectEntityIDAndMarshal(t *testing.T) {
	t.Parallel()

	t.Run("adds entity map when absent", func(t *testing.T) {
		t.Parallel()
		body, diags := injectEntityIDAndMarshal(map[string]any{"tags": []string{"a"}}, "host-1")
		require.False(t, diags.HasError())

		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		entity, ok := got["entity"].(map[string]any)
		require.True(t, ok, "entity object should be present")
		assert.Equal(t, "host-1", entity["id"])
		assert.Contains(t, got, "tags")
	})

	t.Run("sets id on existing entity map without dropping fields", func(t *testing.T) {
		t.Parallel()
		body, diags := injectEntityIDAndMarshal(map[string]any{
			"entity": map[string]any{"name": "web01"},
		}, "host-2")
		require.False(t, diags.HasError())

		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		entity := got["entity"].(map[string]any)
		assert.Equal(t, "host-2", entity["id"])
		assert.Equal(t, "web01", entity["name"], "existing entity fields must be preserved")
	})

	t.Run("overwrites an existing id", func(t *testing.T) {
		t.Parallel()
		body, diags := injectEntityIDAndMarshal(map[string]any{
			"entity": map[string]any{"id": "old"},
		}, "new")
		require.False(t, diags.HasError())

		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "new", got["entity"].(map[string]any)["id"])
	})

	t.Run("returns diagnostics on unmarshalable payload", func(t *testing.T) {
		t.Parallel()
		// A channel cannot be JSON-marshaled, exercising the error path.
		_, diags := injectEntityIDAndMarshal(map[string]any{"bad": make(chan int)}, "host-3")
		require.True(t, diags.HasError())
		assert.Equal(t, "JSON marshal error", diags.Errors()[0].Summary())
	})
}
