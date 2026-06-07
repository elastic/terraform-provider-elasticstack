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

func TestIsKnownSemanticallyEmptyMappings(t *testing.T) {
	t.Run("null returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmptyMappings(esindex.NewMappingsNull()))
	})

	t.Run("unknown returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmptyMappings(esindex.NewMappingsUnknown()))
	})

	t.Run("empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmptyMappings(esindex.NewMappingsValue("{}")))
	})

	t.Run("whitespace-padded empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmptyMappings(esindex.NewMappingsValue("  {}  ")))
	})

	t.Run("non-empty JSON object returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmptyMappings(esindex.NewMappingsValue(`{"properties":{}}`)))
	})
}

func TestIsKnownSemanticallyEmptySettings(t *testing.T) {
	t.Run("null returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmptySettings(customtypes.NewIndexSettingsNull()))
	})

	t.Run("unknown returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmptySettings(customtypes.NewIndexSettingsUnknown()))
	})

	t.Run("empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmptySettings(customtypes.NewIndexSettingsValue("{}")))
	})

	t.Run("whitespace-padded empty JSON object returns true", func(t *testing.T) {
		assert.True(t, IsKnownSemanticallyEmptySettings(customtypes.NewIndexSettingsValue("  {}  ")))
	})

	t.Run("non-empty JSON object returns false", func(t *testing.T) {
		assert.False(t, IsKnownSemanticallyEmptySettings(customtypes.NewIndexSettingsValue(`{"number_of_shards":1}`)))
	})
}
