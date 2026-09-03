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

package typeutils_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestAssertSameType(t *testing.T) {
	t.Run("same type", func(t *testing.T) {
		expected := types.StringValue("hello")
		other := types.StringValue("world")

		got, ok, diags := typeutils.AssertSameType(expected, other)

		require.True(t, ok)
		require.False(t, diags.HasError())
		require.Equal(t, other, got)
	})

	t.Run("different type", func(t *testing.T) {
		expected := types.StringValue("hello")
		other := types.BoolValue(true)

		_, ok, diags := typeutils.AssertSameType[types.String](expected, other)

		require.False(t, ok)
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary(), "Semantic Equality Check Error")
		require.Contains(t, diags[0].Detail(), "types.String")
		require.Contains(t, diags[0].Detail(), "types.Bool")
	})
}

func TestUnmarshalJSONForSemanticEquals(t *testing.T) {
	t.Run("valid json", func(t *testing.T) {
		got, diags := typeutils.UnmarshalJSONForSemanticEquals[map[string]any](`{"a":1}`)

		require.False(t, diags.HasError())
		require.Equal(t, map[string]any{"a": float64(1)}, got)
	})

	t.Run("invalid json", func(t *testing.T) {
		_, diags := typeutils.UnmarshalJSONForSemanticEquals[map[string]any](`{invalid`)

		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary(), "Semantic Equality Check Error")
	})
}
