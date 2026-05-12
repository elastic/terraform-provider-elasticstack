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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_populateMarkdownFromAPI(t *testing.T) {
	content := "hello"
	description := "desc"
	hideTitle := true
	title := "panel title"
	openLinks := false
	cfg := kbapi.KbnDashboardPanelTypeMarkdownConfig0{
		Content:     content,
		Description: &description,
		HideTitle:   &hideTitle,
		Title:       &title,
	}
	cfg.Settings.OpenLinksInNewTab = &openLinks

	pm := &panelModel{}
	populateMarkdownFromAPI(pm, cfg)
	require.NotNil(t, pm.MarkdownConfig)
	require.NotNil(t, pm.MarkdownConfig.ByValue)
	require.Nil(t, pm.MarkdownConfig.ByReference)
	bv := pm.MarkdownConfig.ByValue
	assert.Equal(t, types.StringValue(content), bv.Content)
	assert.Equal(t, types.StringValue(description), bv.Description)
	assert.Equal(t, types.BoolValue(hideTitle), bv.HideTitle)
	assert.Equal(t, types.StringValue(title), bv.Title)
	require.NotNil(t, bv.Settings)
	assert.Equal(t, types.BoolValue(openLinks), bv.Settings.OpenLinksInNewTab)
}

func Test_buildMarkdownConfig(t *testing.T) {
	pm := panelModel{
		MarkdownConfig: &markdownConfigModel{
			ByValue: &markdownConfigByValueModel{
				Content:     types.StringValue("hello"),
				Description: types.StringValue("desc"),
				HideTitle:   types.BoolValue(false),
				Title:       types.StringValue("panel title"),
				Settings: &markdownConfigSettingsModel{
					OpenLinksInNewTab: types.BoolValue(true),
				},
			},
		},
	}

	cfg := buildMarkdownConfig(pm)
	require.NotNil(t, cfg.Description)
	require.NotNil(t, cfg.HideTitle)
	require.NotNil(t, cfg.Title)
	require.NotNil(t, cfg.Settings.OpenLinksInNewTab)
	assert.Equal(t, "hello", cfg.Content)
	assert.Equal(t, "desc", *cfg.Description)
	assert.False(t, *cfg.HideTitle)
	assert.Equal(t, "panel title", *cfg.Title)
	assert.True(t, *cfg.Settings.OpenLinksInNewTab)
}

func Test_markdownConfigRoundTripViaUnion(t *testing.T) {
	pm := panelModel{
		MarkdownConfig: &markdownConfigModel{
			ByValue: &markdownConfigByValueModel{
				Content:   types.StringValue("round trip"),
				HideTitle: types.BoolValue(true),
				Settings: &markdownConfigSettingsModel{
					OpenLinksInNewTab: types.BoolNull(),
				},
			},
		},
	}

	cfg0 := buildMarkdownConfig(pm)
	var union kbapi.KbnDashboardPanelTypeMarkdown_Config
	require.NoError(t, union.FromKbnDashboardPanelTypeMarkdownConfig0(cfg0))

	decoded, err := union.AsKbnDashboardPanelTypeMarkdownConfig0()
	require.NoError(t, err)

	pm2 := &panelModel{}
	populateMarkdownFromAPI(pm2, decoded)
	require.NotNil(t, pm2.MarkdownConfig)
	require.NotNil(t, pm2.MarkdownConfig.ByValue)
	assert.Equal(t, pm.MarkdownConfig.ByValue.Content, pm2.MarkdownConfig.ByValue.Content)
	assert.Equal(t, pm.MarkdownConfig.ByValue.HideTitle, pm2.MarkdownConfig.ByValue.HideTitle)
}
