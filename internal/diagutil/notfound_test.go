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
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type notFoundTestState struct {
	Name string
}

func TestWarnNotFoundAndKeepState(t *testing.T) {
	t.Run("returns the passed-in state unchanged with found=false", func(t *testing.T) {
		state := notFoundTestState{Name: "prior"}

		got, found, diags := WarnNotFoundAndKeepState(context.Background(), "Widget", "my-id", state, diag.Diagnostics{})

		assert.False(t, found)
		assert.Equal(t, state, got)
		assert.False(t, diags.HasError())
	})

	t.Run("preserves diagnostics already accumulated by the caller", func(t *testing.T) {
		var diags diag.Diagnostics
		diags.AddWarning("prior warning", "detail")

		_, found, gotDiags := WarnNotFoundAndKeepState(context.Background(), "Widget", "my-id", notFoundTestState{}, diags)

		assert.False(t, found)
		require.Len(t, gotDiags, 1)
		assert.Equal(t, "prior warning", gotDiags[0].Summary())
	})

	t.Run("works with generic state types other than structs", func(t *testing.T) {
		got, found, _ := WarnNotFoundAndKeepState(context.Background(), "Widget", "my-id", "prior-state", diag.Diagnostics{})

		assert.False(t, found)
		assert.Equal(t, "prior-state", got)
	})
}
