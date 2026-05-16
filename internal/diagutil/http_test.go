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
