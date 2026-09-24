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

package kbschema

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestAttrTypesCache_Get(t *testing.T) {
	var cache AttrTypesCache
	calls := 0

	populate := func(set Set) {
		calls++
		set("a", map[string]attr.Type{"foo": types.StringType})
		set("b", map[string]attr.Type{"bar": types.Int64Type})
	}

	require.Equal(t, map[string]attr.Type{"foo": types.StringType}, cache.Get("a", populate))
	require.Equal(t, map[string]attr.Type{"bar": types.Int64Type}, cache.Get("b", populate))
	require.Equal(t, 1, calls)
}

func TestAttrTypesCache_Get_PopulatesOnce(t *testing.T) {
	var cache AttrTypesCache
	calls := 0

	populate := func(set Set) {
		calls++
		set("key", map[string]attr.Type{"value": types.StringType})
	}

	for i := 0; i < 5; i++ {
		cache.Get("key", populate)
	}

	require.Equal(t, 1, calls)
}

func TestAttrTypesCache_Get_UnknownKeyReturnsNil(t *testing.T) {
	var cache AttrTypesCache

	populate := func(set Set) {
		set("known", map[string]attr.Type{"value": types.StringType})
	}

	require.Nil(t, cache.Get("missing", populate))
}
