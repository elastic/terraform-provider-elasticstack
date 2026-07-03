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

func TestReadURLDrilldownsFromAPI_empty(t *testing.T) {
	t.Parallel()
	require.Nil(t, panelkit.ReadURLDrilldownsFromAPI(nil, nil))
	require.Nil(t, panelkit.ReadURLDrilldownsFromAPI([]panelkit.URLDrilldownAPIItemData{}, nil))
}

func TestReadURLDrilldownsFromAPI_import_defaultsNulled(t *testing.T) {
	t.Parallel()
	// encode_url=true and open_in_new_tab=false are the server defaults — they should
	// come back as null on import (prior==nil) so practitioners can omit them.
	encTrue, encFalse := true, false
	items := []panelkit.URLDrilldownAPIItemData{
		{URL: "https://x", Label: "L", EncodeUrl: &encTrue, OpenInNewTab: &encFalse},
	}
	out := panelkit.ReadURLDrilldownsFromAPI(items, nil)
	require.Len(t, out, 1)
	assert.Equal(t, "https://x", out[0].URL.ValueString())
	assert.Equal(t, "L", out[0].Label.ValueString())
	assert.True(t, out[0].EncodeURL.IsNull(), "encode_url equals server default → null on import")
	assert.True(t, out[0].OpenInNewTab.IsNull(), "open_in_new_tab equals server default → null on import")
}

func TestReadURLDrilldownsFromAPI_import_nonDefaultSet(t *testing.T) {
	t.Parallel()
	// encode_url=false is non-default; open_in_new_tab=true is non-default — both should be kept.
	encFalse, openTrue := false, true
	items := []panelkit.URLDrilldownAPIItemData{
		{URL: "https://y", Label: "M", EncodeUrl: &encFalse, OpenInNewTab: &openTrue},
	}
	out := panelkit.ReadURLDrilldownsFromAPI(items, nil)
	require.Len(t, out, 1)
	assert.False(t, out[0].EncodeURL.IsNull())
	assert.False(t, out[0].EncodeURL.ValueBool())
	assert.False(t, out[0].OpenInNewTab.IsNull())
	assert.True(t, out[0].OpenInNewTab.ValueBool())
}

func TestReadURLDrilldownsFromAPI_import_omittedFields(t *testing.T) {
	t.Parallel()
	// API omits encode_url and open_in_new_tab entirely — both should be null.
	items := []panelkit.URLDrilldownAPIItemData{
		{URL: "https://z", Label: "N"},
	}
	out := panelkit.ReadURLDrilldownsFromAPI(items, nil)
	require.Len(t, out, 1)
	assert.True(t, out[0].EncodeURL.IsNull())
	assert.True(t, out[0].OpenInNewTab.IsNull())
}

func TestReadURLDrilldownsFromAPI_refresh_nullPreserved(t *testing.T) {
	t.Parallel()
	// Prior state has null for both optional fields — refresh should keep them null
	// even if the API returns non-nil values.
	prior := []models.URLDrilldownModel{
		{
			URL:          types.StringValue("https://a"),
			Label:        types.StringValue("A"),
			EncodeURL:    types.BoolNull(),
			OpenInNewTab: types.BoolNull(),
		},
	}
	encTrue, encFalse := true, false
	items := []panelkit.URLDrilldownAPIItemData{
		{URL: "https://a", Label: "A", EncodeUrl: &encTrue, OpenInNewTab: &encFalse},
	}
	out := panelkit.ReadURLDrilldownsFromAPI(items, prior)
	require.Len(t, out, 1)
	assert.True(t, out[0].EncodeURL.IsNull(), "prior null → stays null on refresh")
	assert.True(t, out[0].OpenInNewTab.IsNull(), "prior null → stays null on refresh")
}

func TestReadURLDrilldownsFromAPI_refresh_knownUpdated(t *testing.T) {
	t.Parallel()
	// Prior state has known values — they should be updated from the API.
	prior := []models.URLDrilldownModel{
		{
			URL:          types.StringValue("https://b"),
			Label:        types.StringValue("B"),
			EncodeURL:    types.BoolValue(false),
			OpenInNewTab: types.BoolValue(true),
		},
	}
	encTrue, encFalse := true, false
	items := []panelkit.URLDrilldownAPIItemData{
		{URL: "https://b", Label: "B", EncodeUrl: &encTrue, OpenInNewTab: &encFalse},
	}
	out := panelkit.ReadURLDrilldownsFromAPI(items, prior)
	require.Len(t, out, 1)
	assert.False(t, out[0].EncodeURL.IsNull())
	assert.True(t, out[0].EncodeURL.ValueBool())
	assert.False(t, out[0].OpenInNewTab.IsNull())
	assert.False(t, out[0].OpenInNewTab.ValueBool())
}

func TestReadDiscoverSessionDrilldownsFromAPI_import_defaultsNulled(t *testing.T) {
	t.Parallel()
	encTrue, encFalse := true, false
	items := []panelkit.URLDrilldownAPIItemData{
		{URL: "https://ds", Label: "DS", EncodeUrl: &encTrue, OpenInNewTab: &encFalse},
	}
	out := panelkit.ReadDiscoverSessionDrilldownsFromAPI(items, nil)
	require.Len(t, out, 1)
	assert.Equal(t, "https://ds", out[0].URL.ValueString())
	assert.True(t, out[0].EncodeURL.IsNull())
	assert.True(t, out[0].OpenInNewTab.IsNull())
}

func TestReadDiscoverSessionDrilldownsFromAPI_refresh_nullPreserved(t *testing.T) {
	t.Parallel()
	prior := []models.DiscoverSessionPanelDrilldown{
		{
			URL:          types.StringValue("https://ds"),
			Label:        types.StringValue("DS"),
			EncodeURL:    types.BoolNull(),
			OpenInNewTab: types.BoolNull(),
		},
	}
	encTrue, encFalse := true, false
	items := []panelkit.URLDrilldownAPIItemData{
		{URL: "https://ds", Label: "DS", EncodeUrl: &encTrue, OpenInNewTab: &encFalse},
	}
	out := panelkit.ReadDiscoverSessionDrilldownsFromAPI(items, prior)
	require.Len(t, out, 1)
	assert.True(t, out[0].EncodeURL.IsNull())
	assert.True(t, out[0].OpenInNewTab.IsNull())
}
