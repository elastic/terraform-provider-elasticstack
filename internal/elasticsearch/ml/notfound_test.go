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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireNonEmptyID(t *testing.T) {
	t.Run("empty id produces a labelled diagnostic", func(t *testing.T) {
		diags := RequireNonEmptyID("", "calendar_id")
		require.True(t, diags.HasError())
		assert.Equal(t, "Invalid resource ID", diags.Errors()[0].Summary())
		assert.Equal(t, "calendar_id cannot be empty", diags.Errors()[0].Detail())
	})

	t.Run("non-empty id produces no diagnostics", func(t *testing.T) {
		diags := RequireNonEmptyID("thing-1", "calendar_id")
		assert.False(t, diags.HasError())
	})
}
