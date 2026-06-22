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

package stateutil_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/stretchr/testify/require"
)

func TestCollapseListPath_key_absent(t *testing.T) {
	t.Parallel()
	m := map[string]any{"other": "value"}
	diags := stateutil.CollapseListPath(m, "missing", "missing")
	require.False(t, diags.HasError())
	_, ok := m["missing"]
	require.False(t, ok)
}

func TestCollapseListPath_value_nil(t *testing.T) {
	t.Parallel()
	m := map[string]any{"key": nil}
	diags := stateutil.CollapseListPath(m, "key", "key")
	require.False(t, diags.HasError())
	v, ok := m["key"]
	require.True(t, ok)
	require.Nil(t, v)
}

func TestCollapseListPath_non_list_passthrough(t *testing.T) {
	t.Parallel()
	obj := map[string]any{"field": "val"}
	m := map[string]any{"key": obj}
	diags := stateutil.CollapseListPath(m, "key", "key")
	require.False(t, diags.HasError())
	require.Equal(t, obj, m["key"])
}

func TestCollapseListPath_empty_list_to_nil(t *testing.T) {
	t.Parallel()
	m := map[string]any{"key": []any{}}
	diags := stateutil.CollapseListPath(m, "key", "key")
	require.False(t, diags.HasError())
	v, ok := m["key"]
	require.True(t, ok)
	require.Nil(t, v)
}

func TestCollapseListPath_singleton_list_to_element(t *testing.T) {
	t.Parallel()
	inner := map[string]any{"field": "value"}
	m := map[string]any{"key": []any{inner}}
	diags := stateutil.CollapseListPath(m, "key", "key")
	require.False(t, diags.HasError())
	require.Equal(t, inner, m["key"])
}

func TestCollapseListPath_multi_element_error(t *testing.T) {
	t.Parallel()
	m := map[string]any{"key": []any{map[string]any{}, map[string]any{}}}
	diags := stateutil.CollapseListPath(m, "key", "my.path.key")
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), `"my.path.key"`)
}
