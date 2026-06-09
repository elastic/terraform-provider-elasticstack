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

package templateutil

import (
	"testing"

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/stretchr/testify/assert"
)

func TestIsKnownSemanticallyEmpty(t *testing.T) {
	t.Run("mappings null returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(esindex.NewMappingsNull()))
	})

	t.Run("mappings unknown returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(esindex.NewMappingsUnknown()))
	})

	t.Run("mappings empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmpty(esindex.NewMappingsValue("{}")))
	})

	t.Run("mappings whitespace-padded empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmpty(esindex.NewMappingsValue("  {}  ")))
	})

	t.Run("mappings non-empty JSON object returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(esindex.NewMappingsValue(`{"properties":{}}`)))
	})

	t.Run("settings null returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(customtypes.NewIndexSettingsNull()))
	})

	t.Run("settings unknown returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(customtypes.NewIndexSettingsUnknown()))
	})

	t.Run("settings empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmpty(customtypes.NewIndexSettingsValue("{}")))
	})

	t.Run("settings whitespace-padded empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmpty(customtypes.NewIndexSettingsValue("  {}  ")))
	})

	t.Run("settings non-empty JSON object returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmpty(customtypes.NewIndexSettingsValue(`{"number_of_shards":1}`)))
	})
}
