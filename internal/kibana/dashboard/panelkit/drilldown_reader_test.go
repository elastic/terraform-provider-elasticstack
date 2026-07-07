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

package panelkit

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boolPtr(b bool) *bool { return &b }

func TestReadURLDrilldownsFromAPI_nilEntries_returnsNil(t *testing.T) {
	t.Parallel()
	result := ReadURLDrilldownsFromAPI(nil, nil, true, false)
	assert.Nil(t, result)
}

func TestReadURLDrilldownsFromAPI_emptyEntries_returnsNil(t *testing.T) {
	t.Parallel()
	result := ReadURLDrilldownsFromAPI([]URLDrilldownAPIEntry{}, nil, true, false)
	assert.Nil(t, result)
}

func TestReadURLDrilldownsFromAPI_importMode_defaultValues_produceNull(t *testing.T) {
	t.Parallel()
	// When prior==nil (import) and API values equal the server defaults, both fields should be null
	// so practitioners can omit them in their config.
	entries := []URLDrilldownAPIEntry{
		{URL: "https://example.com", Label: "open", EncodeURL: boolPtr(true), OpenInNewTab: boolPtr(false)},
	}
	result := ReadURLDrilldownsFromAPI(entries, nil, true, false)
	require.Len(t, result, 1)
	assert.Equal(t, "https://example.com", result[0].URL.ValueString())
	assert.Equal(t, "open", result[0].Label.ValueString())
	assert.True(t, result[0].EncodeURL.IsNull(), "encode_url matching default should be null on import")
	assert.True(t, result[0].OpenInNewTab.IsNull(), "open_in_new_tab matching default should be null on import")
}

func TestReadURLDrilldownsFromAPI_importMode_nonDefaultValues_stored(t *testing.T) {
	t.Parallel()
	// When prior==nil (import) and API values differ from defaults, they should be stored.
	entries := []URLDrilldownAPIEntry{
		{URL: "https://example.com", Label: "open", EncodeURL: boolPtr(false), OpenInNewTab: boolPtr(true)},
	}
	result := ReadURLDrilldownsFromAPI(entries, nil, true, false)
	require.Len(t, result, 1)
	assert.False(t, result[0].EncodeURL.IsNull())
	assert.Equal(t, false, result[0].EncodeURL.ValueBool())
	assert.False(t, result[0].OpenInNewTab.IsNull())
	assert.Equal(t, true, result[0].OpenInNewTab.ValueBool())
}

func TestReadURLDrilldownsFromAPI_importMode_nilAPIBools_produceNull(t *testing.T) {
	t.Parallel()
	entries := []URLDrilldownAPIEntry{
		{URL: "https://example.com", Label: "open", EncodeURL: nil, OpenInNewTab: nil},
	}
	result := ReadURLDrilldownsFromAPI(entries, nil, true, false)
	require.Len(t, result, 1)
	assert.True(t, result[0].EncodeURL.IsNull())
	assert.True(t, result[0].OpenInNewTab.IsNull())
}

func TestReadURLDrilldownsFromAPI_refreshMode_priorNullPreserved(t *testing.T) {
	t.Parallel()
	// When prior is non-nil with null fields, null must be preserved regardless of API value.
	prior := []models.URLDrilldownModel{
		{
			URL:          types.StringValue("https://example.com"),
			Label:        types.StringValue("open"),
			EncodeURL:    types.BoolNull(),
			OpenInNewTab: types.BoolNull(),
		},
	}
	entries := []URLDrilldownAPIEntry{
		{URL: "https://example.com", Label: "open", EncodeURL: boolPtr(true), OpenInNewTab: boolPtr(false)},
	}
	result := ReadURLDrilldownsFromAPI(entries, prior, true, false)
	require.Len(t, result, 1)
	assert.True(t, result[0].EncodeURL.IsNull(), "prior null encode_url should be preserved")
	assert.True(t, result[0].OpenInNewTab.IsNull(), "prior null open_in_new_tab should be preserved")
}

func TestReadURLDrilldownsFromAPI_refreshMode_knownPriorUpdatedFromAPI(t *testing.T) {
	t.Parallel()
	// When prior is non-nil with known fields, API values replace them.
	prior := []models.URLDrilldownModel{
		{
			URL:          types.StringValue("https://example.com"),
			Label:        types.StringValue("open"),
			EncodeURL:    types.BoolValue(false),
			OpenInNewTab: types.BoolValue(true),
		},
	}
	entries := []URLDrilldownAPIEntry{
		{URL: "https://example.com", Label: "open", EncodeURL: boolPtr(true), OpenInNewTab: boolPtr(false)},
	}
	result := ReadURLDrilldownsFromAPI(entries, prior, true, false)
	require.Len(t, result, 1)
	assert.Equal(t, true, result[0].EncodeURL.ValueBool())
	assert.Equal(t, false, result[0].OpenInNewTab.ValueBool())
}

func TestReadURLDrilldownsFromAPI_refreshMode_knownPriorAPIReturnsNil_goesNull(t *testing.T) {
	t.Parallel()
	// When prior is known but API omits the field, result should be null.
	prior := []models.URLDrilldownModel{
		{
			URL:          types.StringValue("https://example.com"),
			Label:        types.StringValue("open"),
			EncodeURL:    types.BoolValue(true),
			OpenInNewTab: types.BoolValue(false),
		},
	}
	entries := []URLDrilldownAPIEntry{
		{URL: "https://example.com", Label: "open", EncodeURL: nil, OpenInNewTab: nil},
	}
	result := ReadURLDrilldownsFromAPI(entries, prior, true, false)
	require.Len(t, result, 1)
	assert.True(t, result[0].EncodeURL.IsNull(), "known prior with nil API should produce null")
	assert.True(t, result[0].OpenInNewTab.IsNull())
}

func TestReadURLDrilldownsFromAPI_moreEntriesThanPrior_newItemsUseImportMode(t *testing.T) {
	t.Parallel()
	prior := []models.URLDrilldownModel{
		{
			URL:          types.StringValue("https://a.com"),
			Label:        types.StringValue("a"),
			EncodeURL:    types.BoolNull(),
			OpenInNewTab: types.BoolNull(),
		},
	}
	entries := []URLDrilldownAPIEntry{
		{URL: "https://a.com", Label: "a", EncodeURL: boolPtr(true), OpenInNewTab: boolPtr(false)},
		{URL: "https://b.com", Label: "b", EncodeURL: boolPtr(false), OpenInNewTab: boolPtr(true)},
	}
	result := ReadURLDrilldownsFromAPI(entries, prior, true, false)
	require.Len(t, result, 2)
	// First entry: prior is null → null preserved
	assert.True(t, result[0].EncodeURL.IsNull())
	assert.True(t, result[0].OpenInNewTab.IsNull())
	// Second entry: no prior → import mode; non-default values stored
	assert.False(t, result[1].EncodeURL.IsNull())
	assert.Equal(t, false, result[1].EncodeURL.ValueBool())
	assert.False(t, result[1].OpenInNewTab.IsNull())
	assert.Equal(t, true, result[1].OpenInNewTab.ValueBool())
}
