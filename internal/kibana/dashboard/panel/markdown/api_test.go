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

package markdown_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/markdown"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestContract_byValue(t *testing.T) {
	t.Parallel()
	contracttest.Run(t, markdown.Handler{}, contracttest.Config{
		FullAPIResponse: `{
			"type": "markdown",
			"grid": {"x": 0, "y": 0, "w": 12, "h": 8},
			"id": "markdown-contract",
			"config": {
				"content": "# Title",
				"settings": {"open_links_in_new_tab": false},
				"hide_title": false,
				"hide_border": false,
				"description": "Unit test panel"
			}
		}`,
		OmitValidateRequiredZero: true,
		OmitRequiredLeafPresence: true,
		SkipFields: []string{
			"by_reference",
			"config.title",
			"by_value.title",
			"by_value.hide_title",
			"by_value.hide_border",
			"by_value.description",
			"by_value.settings.open_links_in_new_tab",
		},
	})
}

func TestClassifyJSON_byValueContentString(t *testing.T) {
	t.Parallel()
	h := markdown.Handler{}
	require.True(t, h.ClassifyJSON(map[string]any{"content": "# x"}))
	require.False(t, h.ClassifyJSON(map[string]any{"ref_id": "lib-1"}))
	require.False(t, h.ClassifyJSON(nil))
}

func TestPopulateJSONDefaults_openLinksWhenAbsent(t *testing.T) {
	t.Parallel()
	h := markdown.Handler{}
	cfg := map[string]any{"content": "hello"}
	got := h.PopulateJSONDefaults(cfg)
	settings := got["settings"].(map[string]any)
	require.Equal(t, true, settings["open_links_in_new_tab"])
}

func TestPopulateJSONDefaults_preservesExplicitOpenLinks(t *testing.T) {
	t.Parallel()
	h := markdown.Handler{}
	cfg := map[string]any{
		"content": "hello",
		"settings": map[string]any{
			"open_links_in_new_tab": false,
		},
	}
	got := h.PopulateJSONDefaults(cfg)
	settings := got["settings"].(map[string]any)
	require.Equal(t, false, settings["open_links_in_new_tab"])
}

func TestFromAPI_byReference_populatesRefID(t *testing.T) {
	t.Parallel()
	raw := `{
		"type": "markdown",
		"grid": {"x": 0, "y": 0, "w": 8, "h": 6},
		"id": "md-ref",
		"config": {"ref_id": "lib-md-99", "title": "From library"}
	}`
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.UnmarshalJSON([]byte(raw)))

	var pm models.PanelModel
	diags := markdown.Handler{}.FromAPI(context.Background(), &pm, nil, item)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, pm.MarkdownConfig)
	require.NotNil(t, pm.MarkdownConfig.ByReference)
	require.Equal(t, "lib-md-99", pm.MarkdownConfig.ByReference.RefID.ValueString())
}

func TestToAPI_configJSONPath(t *testing.T) {
	t.Parallel()
	payload, err := json.Marshal(map[string]any{
		"content": "## ok",
		"settings": map[string]any{
			"open_links_in_new_tab": true,
		},
	})
	require.NoError(t, err)

	pm := models.PanelModel{
		Type: types.StringValue("markdown"),
		Grid: models.PanelGridModel{X: types.Int64Value(0), Y: types.Int64Value(0)},
		ConfigJSON: customtypes.NewJSONWithDefaultsValue(string(payload), func(m map[string]any) map[string]any {
			return m
		}),
	}
	item, diags := markdown.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%v", diags)
	md, err := item.AsKbnDashboardPanelTypeMarkdown()
	require.NoError(t, err)
	cfg0, err := md.Config.AsKbnDashboardPanelTypeMarkdownConfig0()
	require.NoError(t, err)
	require.Equal(t, "## ok", cfg0.Content)
}
