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

package dashboard

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_markdownPanelConfigConverter_handlesAPIPanelConfig(t *testing.T) {
	c := markdownPanelConfigConverter{}

	assert.True(t, c.handlesAPIPanelConfig(&panelModel{MarkdownConfig: &markdownConfigModel{}}, "DASHBOARD_MARKDOWN", apiPanelConfig{}))
	assert.True(t, c.handlesAPIPanelConfig(nil, "DASHBOARD_MARKDOWN", apiPanelConfig{}))
	assert.False(t, c.handlesAPIPanelConfig(&panelModel{}, "DASHBOARD_MARKDOWN", apiPanelConfig{}))
	assert.False(t, c.handlesAPIPanelConfig(&panelModel{MarkdownConfig: &markdownConfigModel{}}, "lens", apiPanelConfig{}))
}

func Test_markdownPanelConfigConverter_populateFromAPIPanel(t *testing.T) {
	content := "hello markdown"
	description := "desc"
	hideTitle := true
	title := "panel title"

	var cfg kbapi.KbnDashboardPanelDASHBOARDMARKDOWN_Config
	require.NoError(t, cfg.FromKbnDashboardPanelDASHBOARDMARKDOWNConfig0(kbapi.KbnDashboardPanelDASHBOARDMARKDOWNConfig0{
		Content:     &content,
		Description: &description,
		HideTitle:   &hideTitle,
		Title:       &title,
	}))
	pm := &panelModel{}
	diags := markdownPanelConfigConverter{}.populateFromAPIPanel(context.Background(), pm, apiPanelConfig{Markdown: &cfg})
	require.False(t, diags.HasError())
	require.NotNil(t, pm.MarkdownConfig)
	assert.Equal(t, types.StringValue(content), pm.MarkdownConfig.Content)
	assert.Equal(t, types.StringValue(description), pm.MarkdownConfig.Description)
	assert.Equal(t, types.BoolValue(hideTitle), pm.MarkdownConfig.HideTitle)
	assert.Equal(t, types.StringValue(title), pm.MarkdownConfig.Title)
}

func Test_markdownPanelConfigConverter_mapPanelToAPI(t *testing.T) {
	pm := panelModel{
		MarkdownConfig: &markdownConfigModel{
			Content:     types.StringValue("body"),
			Description: types.StringValue("desc"),
			HideTitle:   types.BoolValue(false),
			Title:       types.StringValue("title"),
		},
	}

	var out apiPanelConfig
	diags := markdownPanelConfigConverter{}.mapPanelToAPI(pm, &out)
	require.False(t, diags.HasError())

	require.NotNil(t, out.Markdown)
	cfg0, err := out.Markdown.AsKbnDashboardPanelDASHBOARDMARKDOWNConfig0()
	require.NoError(t, err)
	require.NotNil(t, cfg0.Content)
	assert.Equal(t, "body", *cfg0.Content)
	require.NotNil(t, cfg0.Description)
	assert.Equal(t, "desc", *cfg0.Description)
	require.NotNil(t, cfg0.HideTitle)
	assert.False(t, *cfg0.HideTitle)
	require.NotNil(t, cfg0.Title)
	assert.Equal(t, "title", *cfg0.Title)
}
