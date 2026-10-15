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

package rolloverindexseries

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEffectiveWriter_explicitFlag(t *testing.T) {
	name, ok := effectiveWriter([]aliasMember{
		{Name: "orders-000001", IsWriteIndex: new(false)},
		{Name: "orders-000003", IsWriteIndex: new(true)},
	})
	require.True(t, ok)
	require.Equal(t, "orders-000003", name)
}

func TestEffectiveWriter_soleUnsetIsImplicit(t *testing.T) {
	name, ok := effectiveWriter([]aliasMember{{Name: "orders-000001"}})
	require.True(t, ok)
	require.Equal(t, "orders-000001", name)
}

func TestEffectiveWriter_soleExplicitFalseHasNoWriter(t *testing.T) {
	_, ok := effectiveWriter([]aliasMember{{Name: "orders-000001", IsWriteIndex: new(false)}})
	require.False(t, ok)
}

func TestEffectiveWriter_multipleMembersWithoutDesignatedWriter(t *testing.T) {
	_, unsetOK := effectiveWriter([]aliasMember{
		{Name: "orders-000002"},
		{Name: "orders-000001"},
	})
	require.False(t, unsetOK)

	_, falseOK := effectiveWriter([]aliasMember{
		{Name: "orders-000002", IsWriteIndex: new(false)},
		{Name: "orders-000001", IsWriteIndex: new(false)},
	})
	require.False(t, falseOK)
}

func TestEffectiveWriter_doesNotPreferNumericSuffix(t *testing.T) {
	name, ok := effectiveWriter([]aliasMember{
		{Name: "orders-000009"},
		{Name: "orders-current", IsWriteIndex: new(true)},
	})
	require.True(t, ok)
	require.Equal(t, "orders-current", name)
	require.False(t, hasGenerationSuffix("orders-current"))
	require.True(t, hasGenerationSuffix("orders-000003"))
	require.True(t, hasGenerationSuffix("unrelated-index-1"))
}
