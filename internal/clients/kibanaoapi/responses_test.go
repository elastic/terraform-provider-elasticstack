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

package kibanaoapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testAgentBuilderModel struct {
	ID string `json:"id"`
}

func TestHandleGetResponse(t *testing.T) {
	t.Run("returns decoded model on 200", func(t *testing.T) {
		result, diags := handleGetResponse[testAgentBuilderModel](http.StatusOK, []byte(`{"id":"test-id"}`))

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, result)
		assert.Equal(t, "test-id", result.ID)
	})

	t.Run("returns nil result on 404", func(t *testing.T) {
		result, diags := handleGetResponse[testAgentBuilderModel](http.StatusNotFound, nil)

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Nil(t, result)
	})

	t.Run("returns diagnostics for malformed json", func(t *testing.T) {
		result, diags := handleGetResponse[testAgentBuilderModel](http.StatusOK, []byte(`{"id":`))

		require.True(t, diags.HasError())
		assert.Nil(t, result)
	})

	t.Run("returns diagnostics for unexpected status", func(t *testing.T) {
		result, diags := handleGetResponse[testAgentBuilderModel](http.StatusInternalServerError, []byte(`{"error":"boom"}`))

		require.True(t, diags.HasError())
		assert.Nil(t, result)
		assert.Equal(t, "Unexpected status code from server: got HTTP 500", diags[0].Summary())
		assert.JSONEq(t, `{"error":"boom"}`, diags[0].Detail())
	})
}

func TestHandleMutateResponse(t *testing.T) {
	t.Run("returns decoded model on 200", func(t *testing.T) {
		result, diags := handleMutateResponse[testAgentBuilderModel](http.StatusOK, []byte(`{"id":"test-id"}`))

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, result)
		assert.Equal(t, "test-id", result.ID)
	})

	t.Run("returns diagnostics for unexpected status", func(t *testing.T) {
		result, diags := handleMutateResponse[testAgentBuilderModel](http.StatusNotFound, []byte(`{"error":"missing"}`))

		require.True(t, diags.HasError())
		assert.Nil(t, result)
		assert.Equal(t, "Unexpected status code from server: got HTTP 404", diags[0].Summary())
		assert.JSONEq(t, `{"error":"missing"}`, diags[0].Detail())
	})
}

func TestHandleGetTypedResponse(t *testing.T) {
	t.Run("returns extracted value on 200", func(t *testing.T) {
		model := testAgentBuilderModel{ID: "typed-id"}
		result, diags := handleGetTypedResponse(http.StatusOK, nil, func() *testAgentBuilderModel { return &model })

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, result)
		assert.Equal(t, "typed-id", result.ID)
	})

	t.Run("returns nil result on 404", func(t *testing.T) {
		result, diags := handleGetTypedResponse(http.StatusNotFound, nil, func() *testAgentBuilderModel { return nil })

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Nil(t, result)
	})

	t.Run("returns diagnostics for unexpected status", func(t *testing.T) {
		result, diags := handleGetTypedResponse(http.StatusInternalServerError, []byte(`{"error":"boom"}`), func() *testAgentBuilderModel { return nil })

		require.True(t, diags.HasError())
		assert.Nil(t, result)
		assert.Equal(t, "Unexpected status code from server: got HTTP 500", diags[0].Summary())
	})

	t.Run("returns diagnostics when extracted value is nil on 200", func(t *testing.T) {
		result, diags := handleGetTypedResponse(http.StatusOK, nil, func() *testAgentBuilderModel { return nil })

		require.True(t, diags.HasError())
		assert.Nil(t, result)
		assert.Equal(t, "Failed to parse response", diags[0].Summary())
	})
}

func TestHandleMutateTypedResponse(t *testing.T) {
	t.Run("returns extracted value on 200", func(t *testing.T) {
		model := testAgentBuilderModel{ID: "typed-id"}
		result, diags := handleMutateTypedResponse(http.StatusOK, nil, func() *testAgentBuilderModel { return &model })

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, result)
		assert.Equal(t, "typed-id", result.ID)
	})

	t.Run("returns diagnostics for unexpected status", func(t *testing.T) {
		result, diags := handleMutateTypedResponse(http.StatusNotFound, []byte(`{"error":"missing"}`), func() *testAgentBuilderModel { return nil })

		require.True(t, diags.HasError())
		assert.Nil(t, result)
		assert.Equal(t, "Unexpected status code from server: got HTTP 404", diags[0].Summary())
	})

	t.Run("returns extracted value on custom success status", func(t *testing.T) {
		model := testAgentBuilderModel{ID: "typed-id"}
		result, diags := handleMutateTypedResponse(http.StatusCreated, nil, func() *testAgentBuilderModel { return &model }, http.StatusOK, http.StatusCreated)

		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		require.NotNil(t, result)
		assert.Equal(t, "typed-id", result.ID)
	})

	t.Run("returns diagnostics when extracted value is nil on success", func(t *testing.T) {
		result, diags := handleMutateTypedResponse(http.StatusOK, nil, func() *testAgentBuilderModel { return nil })

		require.True(t, diags.HasError())
		assert.Nil(t, result)
		assert.Equal(t, "Failed to parse response", diags[0].Summary())
	})
}
