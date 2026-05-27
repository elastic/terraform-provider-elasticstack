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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrDiag(t *testing.T) {
	t.Run("returns single error diagnostic with summary and error detail", func(t *testing.T) {
		err := errors.New("something went wrong")
		diags := ErrDiag("operation failed", err)

		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, "operation failed", diags[0].Summary())
		assert.Equal(t, "something went wrong", diags[0].Detail())
	})

	t.Run("uses summary distinct from error text", func(t *testing.T) {
		err := errors.New("connection refused")
		diags := ErrDiag("Failed to read resource", err)

		require.True(t, diags.HasError())
		assert.Equal(t, "Failed to read resource", diags[0].Summary())
		assert.Equal(t, "connection refused", diags[0].Detail())
	})
}
