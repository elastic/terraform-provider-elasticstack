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

func Test_populateMarkdownFromAPIByValue_mapsAllFields(t *testing.T) {
	content := "hello"
	description := "desc"
	hideTitle := true
	title := "panel title"
	hideFalse := false
	openLinks := false
	cfg := kbapi.KbnDashboardPanelTypeMarkdownConfig0{
		Content:     content,
		Description: &description,
		HideTitle:   &hideTitle,
		Title:       &title,
	}
	cfg.Settings.OpenLinksInNewTab = &openLinks
	cfg.HideBorder = &hideFalse

	pm := &panelModel{}
	populateMarkdownFromAPIByValue(pm, nil, cfg)
	require.NotNil(t, pm.MarkdownConfig)
	require.NotNil(t, pm.MarkdownConfig.ByValue)
	require.Nil(t, pm.MarkdownConfig.ByReference)
	bv := pm.MarkdownConfig.ByValue
	assert.Equal(t, types.StringValue(content), bv.Content)
	assert.Equal(t, types.StringValue(description), bv.Description)
	assert.Equal(t, types.BoolValue(hideTitle), bv.HideTitle)
	assert.Equal(t, types.StringValue(title), bv.Title)
	assert.Equal(t, types.BoolValue(false), bv.HideBorder)
	require.NotNil(t, bv.Settings)
	assert.Equal(t, types.BoolValue(openLinks), bv.Settings.OpenLinksInNewTab)
}

func Test_populateMarkdownFromAPIByValue_openLinksNullPreservedWhenAPIDefaultTrue(t *testing.T) {
	apiTrue := true
	cfg := kbapi.KbnDashboardPanelTypeMarkdownConfig0{Content: "hi"}
	cfg.Settings.OpenLinksInNewTab = &apiTrue

	tfPanel := &panelModel{
		MarkdownConfig: &markdownConfigModel{
			ByValue: &markdownConfigByValueModel{
				Content:  types.StringValue("hi"),
				Settings: &markdownConfigSettingsModel{OpenLinksInNewTab: types.BoolNull()},
			},
		},
	}
	pm := &panelModel{}
	populateMarkdownFromAPIByValue(pm, tfPanel, cfg)
	require.NotNil(t, pm.MarkdownConfig.ByValue.Settings)
	assert.True(t, pm.MarkdownConfig.ByValue.Settings.OpenLinksInNewTab.IsNull())
}

func Test_populateMarkdownFromAPIByValue_hideBorderNullPreservedWhenAPIEchoesFalse(t *testing.T) {
	apiFalse := false
	cfg := kbapi.KbnDashboardPanelTypeMarkdownConfig0{Content: "hi"}
	cfg.HideBorder = &apiFalse

	tfPanel := &panelModel{
		MarkdownConfig: &markdownConfigModel{
			ByValue: &markdownConfigByValueModel{
				Content:    types.StringValue("hi"),
				HideBorder: types.BoolNull(),
				Settings:   &markdownConfigSettingsModel{},
			},
		},
	}
	pm := &panelModel{}
	populateMarkdownFromAPIByValue(pm, tfPanel, cfg)
	assert.True(t, pm.MarkdownConfig.ByValue.HideBorder.IsNull())
}

func Test_populateMarkdownFromAPIByValue_hideBorderNullPreservedWhenAPIEchoesTrue(t *testing.T) {
	apiTrue := true
	cfg := kbapi.KbnDashboardPanelTypeMarkdownConfig0{Content: "hi"}
	cfg.HideBorder = &apiTrue

	tfPanel := &panelModel{
		MarkdownConfig: &markdownConfigModel{
			ByValue: &markdownConfigByValueModel{
				Content:    types.StringValue("hi"),
				HideBorder: types.BoolNull(),
				Settings:   &markdownConfigSettingsModel{},
			},
		},
	}
	pm := &panelModel{}
	populateMarkdownFromAPIByValue(pm, tfPanel, cfg)
	assert.True(t, pm.MarkdownConfig.ByValue.HideBorder.IsNull())
}

func Test_classifyMarkdownConfigFromRoot(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		raw  string
		want markdownConfigBranch
	}{
		{name: "by_value", raw: `{"content":"# a"}`, want: markdownConfigBranchByValue},
		{name: "by_reference", raw: `{"ref_id":"lib-1"}`, want: markdownConfigBranchByReference},
		{name: "unknown_both", raw: `{"ref_id":"lib-1","content":"# a"}`, want: markdownConfigBranchUnknown},
		{name: "unknown_neither", raw: `{"title":"t"}`, want: markdownConfigBranchUnknown},
		{name: "by_value_empty_content", raw: `{"content":""}`, want: markdownConfigBranchByValue},
		{name: "unknown_empty_ref_id", raw: `{"ref_id":""}`, want: markdownConfigBranchUnknown},
		{name: "by_reference_non_string_content_missing", raw: `{"ref_id":"r"}`, want: markdownConfigBranchByReference},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			br, err := classifyMarkdownConfigFromRoot([]byte(tc.raw))
			require.NoError(t, err)
			assert.Equal(t, tc.want, br)
		})
	}
	_, err := classifyMarkdownConfigFromRoot([]byte(`not json`))
	require.Error(t, err)
}

func Test_populateMarkdownFromAPIByReference_mapsPresentationFields(t *testing.T) {
	desc := "d"
	hide := true
	title := "overlay"
	border := false
	cfg := kbapi.KbnDashboardPanelTypeMarkdownConfig1{
		RefId:       "lib-md-1",
		Description: &desc,
		HideTitle:   &hide,
		Title:       &title,
		HideBorder:  &border,
	}
	pm := &panelModel{}
	populateMarkdownFromAPIByReference(pm, nil, cfg)
	require.NotNil(t, pm.MarkdownConfig.ByReference)
	require.Nil(t, pm.MarkdownConfig.ByValue)
	br := pm.MarkdownConfig.ByReference
	assert.Equal(t, types.StringValue("lib-md-1"), br.RefID)
	assert.Equal(t, types.StringValue("d"), br.Description)
	assert.Equal(t, types.BoolValue(true), br.HideTitle)
	assert.Equal(t, types.StringValue("overlay"), br.Title)
	assert.Equal(t, types.BoolValue(false), br.HideBorder)
}

func Test_populateMarkdownFromAPIByReference_hideBorderNullPreservedWhenAPIFalse(t *testing.T) {
	apiFalse := false
	cfg := kbapi.KbnDashboardPanelTypeMarkdownConfig1{RefId: "r1"}
	cfg.HideBorder = &apiFalse

	tfPanel := &panelModel{
		MarkdownConfig: &markdownConfigModel{
			ByReference: &markdownConfigByReferenceModel{
				RefID:      types.StringValue("r1"),
				HideBorder: types.BoolNull(),
			},
		},
	}
	pm := &panelModel{}
	populateMarkdownFromAPIByReference(pm, tfPanel, cfg)
	assert.True(t, pm.MarkdownConfig.ByReference.HideBorder.IsNull())
}

func Test_markdownByReferenceRoundTripViaUnion(t *testing.T) {
	pm := panelModel{
		MarkdownConfig: &markdownConfigModel{
			ByReference: &markdownConfigByReferenceModel{
				RefID: types.StringValue("lib-md-1"),
				Title: types.StringValue("shared"),
			},
		},
	}
	cfg1 := buildMarkdownConfigByReference(pm)
	var union kbapi.KbnDashboardPanelTypeMarkdown_Config
	require.NoError(t, union.FromKbnDashboardPanelTypeMarkdownConfig1(cfg1))
	decoded, err := union.AsKbnDashboardPanelTypeMarkdownConfig1()
	require.NoError(t, err)
	pm2 := &panelModel{}
	populateMarkdownFromAPIByReference(pm2, nil, decoded)
	require.NotNil(t, pm2.MarkdownConfig.ByReference)
	assert.Equal(t, "lib-md-1", pm2.MarkdownConfig.ByReference.RefID.ValueString())
	assert.Equal(t, "shared", pm2.MarkdownConfig.ByReference.Title.ValueString())
}

func Test_buildMarkdownConfig(t *testing.T) {
	pm := panelModel{
		MarkdownConfig: &markdownConfigModel{
			ByValue: &markdownConfigByValueModel{
				Content:     types.StringValue("hello"),
				Description: types.StringValue("desc"),
				HideTitle:   types.BoolValue(false),
				Title:       types.StringValue("panel title"),
				HideBorder:  types.BoolValue(true),
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
	require.NotNil(t, cfg.HideBorder)
	require.NotNil(t, cfg.Settings.OpenLinksInNewTab)
	assert.Equal(t, "hello", cfg.Content)
	assert.Equal(t, "desc", *cfg.Description)
	assert.False(t, *cfg.HideTitle)
	assert.Equal(t, "panel title", *cfg.Title)
	assert.True(t, *cfg.HideBorder)
	assert.True(t, *cfg.Settings.OpenLinksInNewTab)
}

func Test_buildMarkdownConfigByReference(t *testing.T) {
	pm := panelModel{
		MarkdownConfig: &markdownConfigModel{
			ByReference: &markdownConfigByReferenceModel{
				RefID:       types.StringValue("ref-99"),
				Description: types.StringValue("d"),
				HideTitle:   types.BoolValue(true),
				Title:       types.StringValue("t"),
				HideBorder:  types.BoolValue(false),
			},
		},
	}
	out := buildMarkdownConfigByReference(pm)
	assert.Equal(t, "ref-99", out.RefId)
	require.NotNil(t, out.Description)
	require.NotNil(t, out.HideTitle)
	require.NotNil(t, out.Title)
	require.NotNil(t, out.HideBorder)
	assert.Equal(t, "d", *out.Description)
	assert.True(t, *out.HideTitle)
	assert.Equal(t, "t", *out.Title)
	assert.False(t, *out.HideBorder)
}

func Test_markdownConfigByValueRoundTripViaUnion(t *testing.T) {
	pm := panelModel{
		MarkdownConfig: &markdownConfigModel{
			ByValue: &markdownConfigByValueModel{
				Content:    types.StringValue("round trip"),
				HideTitle:  types.BoolValue(true),
				HideBorder: types.BoolValue(false),
				Settings: &markdownConfigSettingsModel{
					OpenLinksInNewTab: types.BoolValue(false),
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
	populateMarkdownFromAPIByValue(pm2, nil, decoded)
	require.NotNil(t, pm2.MarkdownConfig)
	require.NotNil(t, pm2.MarkdownConfig.ByValue)
	assert.Equal(t, pm.MarkdownConfig.ByValue.Content, pm2.MarkdownConfig.ByValue.Content)
	assert.Equal(t, pm.MarkdownConfig.ByValue.HideTitle, pm2.MarkdownConfig.ByValue.HideTitle)
	assert.Equal(t, pm.MarkdownConfig.ByValue.HideBorder, pm2.MarkdownConfig.ByValue.HideBorder)
	require.NotNil(t, pm2.MarkdownConfig.ByValue.Settings)
	assert.Equal(t, types.BoolValue(false), pm2.MarkdownConfig.ByValue.Settings.OpenLinksInNewTab)
}
