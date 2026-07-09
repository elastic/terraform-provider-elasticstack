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

package panelkit_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDrilldownsFromWireJSON_nilInput(t *testing.T) {
	result := panelkit.ReadDrilldownsFromWireJSON(nil, nil)
	assert.Nil(t, result)
}

func TestReadDrilldownsFromWireJSON_invalidJSON(t *testing.T) {
	result := panelkit.ReadDrilldownsFromWireJSON([]byte(`not-json`), nil)
	assert.Nil(t, result)
}

func TestReadDrilldownsFromWireJSON_emptyArray(t *testing.T) {
	result := panelkit.ReadDrilldownsFromWireJSON([]byte(`[]`), nil)
	assert.Nil(t, result)
}

func TestReadDrilldownsFromWireJSON_noPrior_usesAPIValues(t *testing.T) {
	input := `[{"url":"https://example.com","label":"Go","trigger":"on_open_panel_menu","type":"url_drilldown","encode_url":true,"open_in_new_tab":false}]`
	result := panelkit.ReadDrilldownsFromWireJSON([]byte(input), nil)
	require.Len(t, result, 1)
	assert.Equal(t, "https://example.com", result[0].URL.ValueString())
	assert.Equal(t, "Go", result[0].Label.ValueString())
	assert.False(t, result[0].EncodeURL.IsNull())
	assert.True(t, result[0].EncodeURL.ValueBool())
	assert.False(t, result[0].OpenInNewTab.IsNull())
	assert.False(t, result[0].OpenInNewTab.ValueBool())
}

func TestReadDrilldownsFromWireJSON_noPrior_nilBoolsMapToNull(t *testing.T) {
	input := `[{"url":"https://example.com","label":"Go","trigger":"on_open_panel_menu","type":"url_drilldown"}]`
	result := panelkit.ReadDrilldownsFromWireJSON([]byte(input), nil)
	require.Len(t, result, 1)
	assert.True(t, result[0].EncodeURL.IsNull())
	assert.True(t, result[0].OpenInNewTab.IsNull())
}

func TestReadDrilldownsFromWireJSON_nullPreserve_priorNullKeptNull(t *testing.T) {
	// When prior booleans are null, they should remain null regardless of the API value.
	input := `[{"url":"https://example.com","label":"Go","trigger":"on_open_panel_menu","type":"url_drilldown","encode_url":true,"open_in_new_tab":true}]`
	prior := []models.URLDrilldownModel{
		{
			URL:          types.StringValue("https://example.com"),
			Label:        types.StringValue("Go"),
			EncodeURL:    types.BoolNull(),
			OpenInNewTab: types.BoolNull(),
		},
	}
	result := panelkit.ReadDrilldownsFromWireJSON([]byte(input), prior)
	require.Len(t, result, 1)
	assert.True(t, result[0].EncodeURL.IsNull(), "encode_url should remain null when prior was null")
	assert.True(t, result[0].OpenInNewTab.IsNull(), "open_in_new_tab should remain null when prior was null")
}

func TestReadDrilldownsFromWireJSON_nullPreserve_priorKnownUpdatedFromAPI(t *testing.T) {
	// When prior booleans are known, the API value replaces them.
	input := `[{"url":"https://example.com","label":"Go","trigger":"on_open_panel_menu","type":"url_drilldown","encode_url":false,"open_in_new_tab":true}]`
	prior := []models.URLDrilldownModel{
		{
			URL:          types.StringValue("https://example.com"),
			Label:        types.StringValue("Go"),
			EncodeURL:    types.BoolValue(true),
			OpenInNewTab: types.BoolValue(false),
		},
	}
	result := panelkit.ReadDrilldownsFromWireJSON([]byte(input), prior)
	require.Len(t, result, 1)
	assert.False(t, result[0].EncodeURL.IsNull())
	assert.False(t, result[0].EncodeURL.ValueBool())
	assert.False(t, result[0].OpenInNewTab.IsNull())
	assert.True(t, result[0].OpenInNewTab.ValueBool())
}

func TestReadDrilldownsFromWireJSON_newDrilldownBeyondPrior_usesAPIValues(t *testing.T) {
	// When the API returns more drilldowns than prior state, the extra ones use API values directly.
	input := `[` +
		`{"url":"https://a.com","label":"A","trigger":"on_open_panel_menu","type":"url_drilldown","encode_url":true},` +
		`{"url":"https://b.com","label":"B","trigger":"on_open_panel_menu","type":"url_drilldown","encode_url":false}` +
		`]`
	prior := []models.URLDrilldownModel{
		{URL: types.StringValue("https://a.com"), Label: types.StringValue("A"), EncodeURL: types.BoolNull()},
	}
	result := panelkit.ReadDrilldownsFromWireJSON([]byte(input), prior)
	require.Len(t, result, 2)
	// First drilldown: prior was null → null-preserve.
	assert.True(t, result[0].EncodeURL.IsNull())
	// Second drilldown: no prior at this index → use API value.
	assert.False(t, result[1].EncodeURL.IsNull())
	assert.False(t, result[1].EncodeURL.ValueBool())
}

func TestReadDrilldownsFromWireJSON_multipleDrilldowns(t *testing.T) {
	input := `[` +
		`{"url":"https://x.com","label":"X","trigger":"on_open_panel_menu","type":"url_drilldown","encode_url":true,"open_in_new_tab":false},` +
		`{"url":"https://y.com","label":"Y","trigger":"on_open_panel_menu","type":"url_drilldown"}` +
		`]`
	result := panelkit.ReadDrilldownsFromWireJSON([]byte(input), nil)
	require.Len(t, result, 2)
	assert.Equal(t, "https://x.com", result[0].URL.ValueString())
	assert.True(t, result[0].EncodeURL.ValueBool())
	assert.False(t, result[0].OpenInNewTab.ValueBool())
	assert.Equal(t, "https://y.com", result[1].URL.ValueString())
	assert.True(t, result[1].EncodeURL.IsNull())
	assert.True(t, result[1].OpenInNewTab.IsNull())
}
