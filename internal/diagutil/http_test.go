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

package diagutil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnwrapJSON200(t *testing.T) {
	t.Run("returns value when non-nil", func(t *testing.T) {
		val := "hello"
		got, diags := UnwrapJSON200(&val, "thing")
		require.False(t, diags.HasError())
		assert.Equal(t, &val, got)
	})

	t.Run("returns error diagnostic when nil", func(t *testing.T) {
		got, diags := UnwrapJSON200[string](nil, "list item")
		assert.Nil(t, got)
		require.True(t, diags.HasError())
		assert.Equal(t, "Failed to parse list item response", diags[0].Summary())
		assert.Equal(t, "API returned 200 but response body was nil", diags[0].Detail())
	})
}

func TestHandleStatusResponse(t *testing.T) {
	t.Run("returns nil for accepted status codes", func(t *testing.T) {
		assert.False(t, HandleStatusResponse(http.StatusOK, nil, http.StatusOK, http.StatusNotFound).HasError())
		assert.False(t, HandleStatusResponse(http.StatusNotFound, nil, http.StatusOK, http.StatusNotFound).HasError())
	})

	t.Run("returns diagnostics for unexpected status", func(t *testing.T) {
		diags := HandleStatusResponse(http.StatusBadRequest, []byte(`bad request`), http.StatusOK)

		require.True(t, diags.HasError())
		assert.Equal(t, "Unexpected status code from server: got HTTP 400", diags[0].Summary())
		assert.Equal(t, "bad request", diags[0].Detail())
	})
}

func TestReportKibanaBoomHTTPError(t *testing.T) {
	const summary = "failed to import saved objects"

	t.Run("uses Boom message as detail when envelope is valid", func(t *testing.T) {
		body := []byte(`{"statusCode":422,"error":"Unprocessable Entity","message":"Doc belongs to newer Kibana [10.3.0]"}`)

		diags := ReportKibanaBoomHTTPError(http.StatusUnprocessableEntity, summary, body)

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, summary, diags[0].Summary())
		assert.Equal(t, "Doc belongs to newer Kibana [10.3.0]", diags[0].Detail())
	})

	t.Run("falls back to ReportUnknownHTTPError when body is not JSON", func(t *testing.T) {
		body := []byte(`not json`)

		diags := ReportKibanaBoomHTTPError(http.StatusUnprocessableEntity, summary, body)

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, "Unexpected status code from server: got HTTP 422", diags[0].Summary())
		assert.Equal(t, "not json", diags[0].Detail())
	})

	t.Run("falls back to ReportUnknownHTTPError when message is empty", func(t *testing.T) {
		body := []byte(`{"statusCode":422,"error":"Unprocessable Entity","message":""}`)

		diags := ReportKibanaBoomHTTPError(http.StatusUnprocessableEntity, summary, body)

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, "Unexpected status code from server: got HTTP 422", diags[0].Summary())
		assert.JSONEq(t, `{"statusCode":422,"error":"Unprocessable Entity","message":""}`, diags[0].Detail())
	})

	t.Run("falls back when body is empty", func(t *testing.T) {
		diags := ReportKibanaBoomHTTPError(http.StatusUnprocessableEntity, summary, []byte{})

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, "Unexpected status code from server: got HTTP 422", diags[0].Summary())
		assert.Empty(t, diags[0].Detail())
	})

	t.Run("falls back when body is JSON null", func(t *testing.T) {
		diags := ReportKibanaBoomHTTPError(http.StatusUnprocessableEntity, summary, []byte(`null`))

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, "Unexpected status code from server: got HTTP 422", diags[0].Summary())
		assert.Equal(t, "null", diags[0].Detail())
	})

	t.Run("falls back when message field is absent", func(t *testing.T) {
		body := []byte(`{"statusCode":422,"error":"Unprocessable Entity"}`)

		diags := ReportKibanaBoomHTTPError(http.StatusUnprocessableEntity, summary, body)

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, "Unexpected status code from server: got HTTP 422", diags[0].Summary())
		assert.JSONEq(t, `{"statusCode":422,"error":"Unprocessable Entity"}`, diags[0].Detail())
	})

	t.Run("falls back when error field is absent", func(t *testing.T) {
		body := []byte(`{"statusCode":422,"message":"some message"}`)

		diags := ReportKibanaBoomHTTPError(http.StatusUnprocessableEntity, summary, body)

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, "Unexpected status code from server: got HTTP 422", diags[0].Summary())
		assert.JSONEq(t, `{"statusCode":422,"message":"some message"}`, diags[0].Detail())
	})

	t.Run("falls back when statusCode does not match HTTP status", func(t *testing.T) {
		body := []byte(`{"statusCode":500,"error":"Internal Server Error","message":"boom"}`)

		diags := ReportKibanaBoomHTTPError(http.StatusUnprocessableEntity, summary, body)

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, "Unexpected status code from server: got HTTP 422", diags[0].Summary())
		assert.JSONEq(t, `{"statusCode":500,"error":"Internal Server Error","message":"boom"}`, diags[0].Detail())
	})

	t.Run("falls back for non-Boom JSON that includes a message field", func(t *testing.T) {
		body := []byte(`{"message":"not a boom envelope"}`)

		diags := ReportKibanaBoomHTTPError(http.StatusUnprocessableEntity, summary, body)

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, "Unexpected status code from server: got HTTP 422", diags[0].Summary())
		assert.JSONEq(t, `{"message":"not a boom envelope"}`, diags[0].Detail())
	})

	t.Run("trims whitespace from Boom message", func(t *testing.T) {
		body := []byte(`{"statusCode":422,"error":"Unprocessable Entity","message":"  Doc belongs to newer Kibana [10.3.0]  "}`)

		diags := ReportKibanaBoomHTTPError(http.StatusUnprocessableEntity, summary, body)

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, summary, diags[0].Summary())
		assert.Equal(t, "Doc belongs to newer Kibana [10.3.0]", diags[0].Detail())
	})
}
