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

func TestHashedPackageID(t *testing.T) {
	t.Parallel()

	t.Run("matches hash of concatenated name and version", func(t *testing.T) {
		t.Parallel()
		want, err := typeutils.StringToHash("nginx" + "1.2.3")
		require.NoError(t, err)
		require.Equal(t, *want, typeutils.HashedPackageID("nginx", "1.2.3"))
	})

	t.Run("different versions produce different IDs", func(t *testing.T) {
		t.Parallel()
		require.NotEqual(t,
			typeutils.HashedPackageID("nginx", "1.2.3"),
			typeutils.HashedPackageID("nginx", "1.2.4"),
		)
	})

	t.Run("known SHA-1 hex for a fixed input", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "a9993e364706816aba3e25717850c26c9cd0d89d", typeutils.HashedPackageID("ab", "c"))
	})
}

func TestSlashJoinedPackageID(t *testing.T) {
	t.Parallel()

	t.Run("joins name and version with a slash", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "nginx/1.2.3", typeutils.SlashJoinedPackageID("nginx", "1.2.3"))
	})

	t.Run("does not escape slashes already present in the inputs", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "a/b//1.0", typeutils.SlashJoinedPackageID("a/b", "/1.0"))
	})
}
