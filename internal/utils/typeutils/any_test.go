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
	"github.com/stretchr/testify/require"
)

func TestStringFromAny(t *testing.T) {
	t.Parallel()

	t.Run("nil returns null", func(t *testing.T) {
		t.Parallel()
		require.True(t, typeutils.StringFromAny(nil).IsNull())
	})

	t.Run("string value returned", func(t *testing.T) {
		t.Parallel()
		got := typeutils.StringFromAny("hello")
		require.False(t, got.IsNull())
		require.Equal(t, "hello", got.ValueString())
	})

	t.Run("wrong type returns null", func(t *testing.T) {
		t.Parallel()
		require.True(t, typeutils.StringFromAny(42).IsNull())
	})
}

func TestBoolFromAny(t *testing.T) {
	t.Parallel()

	t.Run("nil returns null", func(t *testing.T) {
		t.Parallel()
		require.True(t, typeutils.BoolFromAny(nil).IsNull())
	})

	t.Run("bool value returned", func(t *testing.T) {
		t.Parallel()
		got := typeutils.BoolFromAny(true)
		require.False(t, got.IsNull())
		require.True(t, got.ValueBool())
	})

	t.Run("wrong type returns null", func(t *testing.T) {
		t.Parallel()
		require.True(t, typeutils.BoolFromAny("not-a-bool").IsNull())
	})
}

func TestInt32FromAnyFloat64(t *testing.T) {
	t.Parallel()

	t.Run("nil returns null", func(t *testing.T) {
		t.Parallel()
		require.True(t, typeutils.Int32FromAnyFloat64(nil).IsNull())
	})

	t.Run("float64 value coerced to int32", func(t *testing.T) {
		t.Parallel()
		got := typeutils.Int32FromAnyFloat64(float64(7))
		require.False(t, got.IsNull())
		require.Equal(t, int32(7), got.ValueInt32())
	})

	t.Run("wrong type returns null", func(t *testing.T) {
		t.Parallel()
		require.True(t, typeutils.Int32FromAnyFloat64("not-a-number").IsNull())
	})
}

func TestInt64FromAnyFloat64(t *testing.T) {
	t.Parallel()

	t.Run("nil returns null", func(t *testing.T) {
		t.Parallel()
		require.True(t, typeutils.Int64FromAnyFloat64(nil).IsNull())
	})

	t.Run("float64 value coerced to int64", func(t *testing.T) {
		t.Parallel()
		got := typeutils.Int64FromAnyFloat64(float64(10485760))
		require.False(t, got.IsNull())
		require.Equal(t, int64(10485760), got.ValueInt64())
	})

	t.Run("wrong type returns null", func(t *testing.T) {
		t.Parallel()
		require.True(t, typeutils.Int64FromAnyFloat64("not-a-number").IsNull())
	})
}
