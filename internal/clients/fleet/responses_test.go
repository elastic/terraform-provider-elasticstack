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
