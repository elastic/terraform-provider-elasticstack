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

package fleet

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testFleetItem struct {
	ID string
}

func TestHandleDeleteResponse(t *testing.T) {
	t.Run("no error on 200", func(t *testing.T) {
		diags := handleDeleteResponse(http.StatusOK, nil)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})

	t.Run("no error on 404 (idempotent delete)", func(t *testing.T) {
		diags := handleDeleteResponse(http.StatusNotFound, nil)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})

	t.Run("returns diagnostics for unexpected status", func(t *testing.T) {
		diags := handleDeleteResponse(http.StatusInternalServerError, []byte(`{"error":"boom"}`))
		require.True(t, diags.HasError())
		assert.Equal(t, "Unexpected status code from server: got HTTP 500", diags[0].Summary())
	})
}

func TestHandleGetItem(t *testing.T) {
	t.Run("returns extracted value on 200", func(t *testing.T) {
		item := testFleetItem{ID: "test-id"}
		result, diags := handleGetItem(http.StatusOK, nil, func() *testFleetItem { return &item })

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, result)
		assert.Equal(t, "test-id", result.ID)
	})

	t.Run("returns zero value on 404", func(t *testing.T) {
		result, diags := handleGetItem(http.StatusNotFound, nil, func() *testFleetItem { return nil })

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Nil(t, result)
	})

	t.Run("returns diagnostics for unexpected status", func(t *testing.T) {
		result, diags := handleGetItem(http.StatusInternalServerError, []byte(`{"error":"boom"}`), func() *testFleetItem { return nil })

		require.True(t, diags.HasError())
		assert.Nil(t, result)
		assert.Equal(t, "Unexpected status code from server: got HTTP 500", diags[0].Summary())
	})

	t.Run("works with slice types", func(t *testing.T) {
		items := []testFleetItem{{ID: "a"}, {ID: "b"}}
		result, diags := handleGetItem(http.StatusOK, nil, func() []testFleetItem { return items })

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Equal(t, items, result)
	})

	t.Run("returns nil slice on 404 for slice types", func(t *testing.T) {
		result, diags := handleGetItem(http.StatusNotFound, nil, func() []testFleetItem { return nil })

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Nil(t, result)
	})
}

func TestHandleMutateItem(t *testing.T) {
	t.Run("returns extracted value and status code on 200", func(t *testing.T) {
		item := testFleetItem{ID: "test-id"}
		result, statusCode, diags := handleMutateItem(http.StatusOK, nil, func() *testFleetItem { return &item })

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, result)
		assert.Equal(t, "test-id", result.ID)
		assert.Equal(t, http.StatusOK, statusCode)
	})

	t.Run("returns diagnostics and status code for unexpected status", func(t *testing.T) {
		result, statusCode, diags := handleMutateItem(http.StatusConflict, []byte(`{"error":"conflict"}`), func() *testFleetItem { return nil })

		require.True(t, diags.HasError())
		assert.Nil(t, result)
		assert.Equal(t, http.StatusConflict, statusCode)
		assert.Equal(t, "Unexpected status code from server: got HTTP 409", diags[0].Summary())
	})

	t.Run("returns diagnostics for internal server error", func(t *testing.T) {
		result, statusCode, diags := handleMutateItem(http.StatusInternalServerError, []byte(`{"error":"boom"}`), func() *testFleetItem { return nil })

		require.True(t, diags.HasError())
		assert.Nil(t, result)
		assert.Equal(t, http.StatusInternalServerError, statusCode)
	})
}
