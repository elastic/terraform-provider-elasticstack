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

func TestMapRef(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		t.Parallel()
		var m map[string]int
		require.Nil(t, typeutils.MapRef[int](m))
	})

	t.Run("non-nil returns pointer to map", func(t *testing.T) {
		t.Parallel()
		m := map[string]int{"a": 1}
		p := typeutils.MapRef[int](m)
		require.NotNil(t, p)
		require.Equal(t, m, *p)
	})
}

func TestSliceRef(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		t.Parallel()
		var s []int
		require.Nil(t, typeutils.SliceRef[int](s))
	})

	t.Run("non-nil returns pointer to slice", func(t *testing.T) {
		t.Parallel()
		s := []int{1, 2, 3}
		p := typeutils.SliceRef[int](s)
		require.NotNil(t, p)
		require.Equal(t, s, *p)
	})
}

func TestDeref(t *testing.T) {
	t.Parallel()

	t.Run("nil returns zero value", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, 0, typeutils.Deref[int](nil))
		require.Empty(t, typeutils.Deref[string](nil))
	})

	t.Run("non-nil returns dereferenced value", func(t *testing.T) {
		t.Parallel()
		v := 42
		require.Equal(t, 42, typeutils.Deref(&v))
		s := "hello"
		require.Equal(t, "hello", typeutils.Deref(&s))
	})
}

func TestDefaultIfNil(t *testing.T) {
	t.Parallel()

	t.Run("nil returns zero value", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, 0, typeutils.DefaultIfNil[int](nil))
	})

	t.Run("non-nil returns value", func(t *testing.T) {
		t.Parallel()
		v := 7
		require.Equal(t, 7, typeutils.DefaultIfNil(&v))
	})
}

func TestNonNilSlice(t *testing.T) {
	t.Parallel()

	t.Run("nil becomes empty slice", func(t *testing.T) {
		t.Parallel()
		var s []int
		result := typeutils.NonNilSlice(s)
		require.NotNil(t, result)
		require.Empty(t, result)
	})

	t.Run("non-nil returned unchanged", func(t *testing.T) {
		t.Parallel()
		s := []int{1, 2}
		require.Equal(t, s, typeutils.NonNilSlice(s))
	})
}

func TestItol(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		t.Parallel()
		require.Nil(t, typeutils.Itol(nil))
	})

	t.Run("converts value", func(t *testing.T) {
		t.Parallel()
		v := 42
		result := typeutils.Itol(&v)
		require.NotNil(t, result)
		require.Equal(t, int64(42), *result)
	})
}

func TestLtoi(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil", func(t *testing.T) {
		t.Parallel()
		require.Nil(t, typeutils.Ltoi(nil))
	})

	t.Run("converts value", func(t *testing.T) {
		t.Parallel()
		var v int64 = 99
		result := typeutils.Ltoi(&v)
		require.NotNil(t, result)
		require.Equal(t, 99, *result)
	})
}
