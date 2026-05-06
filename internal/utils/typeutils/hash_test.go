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

func TestStringToHash(t *testing.T) {
	t.Parallel()

	t.Run("same input produces same hash", func(t *testing.T) {
		t.Parallel()
		h1, err1 := typeutils.StringToHash("hello")
		h2, err2 := typeutils.StringToHash("hello")
		require.NoError(t, err1)
		require.NoError(t, err2)
		require.Equal(t, *h1, *h2)
	})

	t.Run("different input produces different hash", func(t *testing.T) {
		t.Parallel()
		h1, _ := typeutils.StringToHash("hello")
		h2, _ := typeutils.StringToHash("world")
		require.NotEqual(t, *h1, *h2)
	})

	t.Run("returns known SHA-1 hex", func(t *testing.T) {
		t.Parallel()
		h, err := typeutils.StringToHash("abc")
		require.NoError(t, err)
		require.Equal(t, "a9993e364706816aba3e25717850c26c9cd0d89d", *h)
	})
}
