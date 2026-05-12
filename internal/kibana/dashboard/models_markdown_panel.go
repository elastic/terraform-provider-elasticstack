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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// markdownConfigModel is the TF model for markdown_config (by-value / by-reference union).
type markdownConfigModel struct {
	ByValue     *markdownConfigByValueModel     `tfsdk:"by_value"`
	ByReference *markdownConfigByReferenceModel `tfsdk:"by_reference"`
}

type markdownConfigByValueModel struct {
	Content     types.String                 `tfsdk:"content"`
	Settings    *markdownConfigSettingsModel `tfsdk:"settings"`
	Description types.String                 `tfsdk:"description"`
	HideTitle   types.Bool                   `tfsdk:"hide_title"`
	Title       types.String                 `tfsdk:"title"`
	HideBorder  types.Bool                   `tfsdk:"hide_border"`
}

type markdownConfigSettingsModel struct {
	OpenLinksInNewTab types.Bool `tfsdk:"open_links_in_new_tab"`
}

type markdownConfigByReferenceModel struct {
	RefID       types.String `tfsdk:"ref_id"`
	Description types.String `tfsdk:"description"`
	HideTitle   types.Bool   `tfsdk:"hide_title"`
	Title       types.String `tfsdk:"title"`
	HideBorder  types.Bool   `tfsdk:"hide_border"`
}

// populateMarkdownFromAPI maps API by-value markdown config into Terraform state.
//
// TODO(markdown-panel-gaps task 2): detect and populate markdown_config.by_reference from
// KbnDashboardPanelTypeMarkdownConfig1; preserve REQ-009 null semantics across all new fields.
func populateMarkdownFromAPI(pm *panelModel, config kbapi.KbnDashboardPanelTypeMarkdownConfig0) {
	settings := &markdownConfigSettingsModel{
		OpenLinksInNewTab: types.BoolPointerValue(config.Settings.OpenLinksInNewTab),
	}
	pm.MarkdownConfig = &markdownConfigModel{
		ByValue: &markdownConfigByValueModel{
			Content:     types.StringValue(config.Content),
			Settings:    settings,
			Description: types.StringPointerValue(config.Description),
			HideTitle:   types.BoolPointerValue(config.HideTitle),
			HideBorder:  types.BoolPointerValue(config.HideBorder),
			Title:       types.StringPointerValue(config.Title),
		},
	}
}

// buildMarkdownConfig builds the API by-value markdown payload from Terraform.
//
// TODO(markdown-panel-gaps task 2): support by_reference (Config1) in panelModel.toAPI and tighten
// validation when MarkdownConfig is present but branches are inconsistent.
func buildMarkdownConfig(pm panelModel) kbapi.KbnDashboardPanelTypeMarkdownConfig0 {
	if pm.MarkdownConfig == nil || pm.MarkdownConfig.ByValue == nil {
		return kbapi.KbnDashboardPanelTypeMarkdownConfig0{}
	}
	bv := pm.MarkdownConfig.ByValue
	config := kbapi.KbnDashboardPanelTypeMarkdownConfig0{
		Content: bv.Content.ValueString(),
	}
	if typeutils.IsKnown(bv.Description) {
		config.Description = bv.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(bv.HideTitle) {
		config.HideTitle = bv.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(bv.HideBorder) {
		config.HideBorder = bv.HideBorder.ValueBoolPointer()
	}
	if typeutils.IsKnown(bv.Title) {
		config.Title = bv.Title.ValueStringPointer()
	}
	if bv.Settings != nil && typeutils.IsKnown(bv.Settings.OpenLinksInNewTab) {
		config.Settings.OpenLinksInNewTab = bv.Settings.OpenLinksInNewTab.ValueBoolPointer()
	}
	return config
}
