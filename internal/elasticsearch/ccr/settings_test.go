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

package ccr_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ccr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeFlatSettingsKeys(t *testing.T) {
	t.Parallel()

	t.Run("flat dotted key is expanded", func(t *testing.T) {
		t.Parallel()
		in := map[string]any{"index.refresh_interval": "30s"}
		out, changed := ccr.NormalizeFlatSettingsKeys(in)
		require.True(t, changed)
		assert.Equal(t, map[string]any{
			"index": map[string]any{"refresh_interval": "30s"},
		}, out)
	})

	t.Run("nested map unchanged, changed is false", func(t *testing.T) {
		t.Parallel()
		in := map[string]any{"index": map[string]any{"refresh_interval": "30s"}}
		out, changed := ccr.NormalizeFlatSettingsKeys(in)
		require.False(t, changed)
		assert.Equal(t, in, out)
	})

	t.Run("mixed flat and non-dotted keys merged", func(t *testing.T) {
		t.Parallel()
		in := map[string]any{
			"number_of_replicas":     "1",
			"index.refresh_interval": "30s",
		}
		out, changed := ccr.NormalizeFlatSettingsKeys(in)
		require.True(t, changed)
		assert.Equal(t, map[string]any{
			"number_of_replicas": "1",
			"index":              map[string]any{"refresh_interval": "30s"},
		}, out)
	})

	t.Run("empty map returns empty, changed is false", func(t *testing.T) {
		t.Parallel()
		in := map[string]any{}
		out, changed := ccr.NormalizeFlatSettingsKeys(in)
		require.False(t, changed)
		assert.Equal(t, in, out)
	})
}

func TestMergeSettingsMaps(t *testing.T) {
	t.Parallel()

	t.Run("empty base returns overlay", func(t *testing.T) {
		t.Parallel()
		overlay := map[string]any{"a": 1}
		got := ccr.MergeSettingsMaps(map[string]any{}, overlay)
		assert.Equal(t, overlay, got)
	})

	t.Run("empty overlay returns base", func(t *testing.T) {
		t.Parallel()
		base := map[string]any{"a": 1}
		got := ccr.MergeSettingsMaps(base, map[string]any{})
		assert.Equal(t, base, got)
	})

	t.Run("overlay leaf wins over base leaf", func(t *testing.T) {
		t.Parallel()
		base := map[string]any{"a": 1}
		overlay := map[string]any{"a": 2}
		got := ccr.MergeSettingsMaps(base, overlay)
		assert.Equal(t, map[string]any{"a": 2}, got)
	})

	t.Run("nested maps are recursively merged", func(t *testing.T) {
		t.Parallel()
		base := map[string]any{"index": map[string]any{"number_of_replicas": "1"}}
		overlay := map[string]any{"index": map[string]any{"refresh_interval": "30s"}}
		got := ccr.MergeSettingsMaps(base, overlay)
		assert.Equal(t, map[string]any{
			"index": map[string]any{
				"number_of_replicas": "1",
				"refresh_interval":   "30s",
			},
		}, got)
	})

	t.Run("disjoint keys from both maps are preserved", func(t *testing.T) {
		t.Parallel()
		base := map[string]any{"a": 1}
		overlay := map[string]any{"b": 2}
		got := ccr.MergeSettingsMaps(base, overlay)
		assert.Equal(t, map[string]any{"a": 1, "b": 2}, got)
	})
}
