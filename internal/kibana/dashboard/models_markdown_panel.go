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

type markdownConfigModel struct {
	Content     types.String `tfsdk:"content"`
	Description types.String `tfsdk:"description"`
	HideTitle   types.Bool   `tfsdk:"hide_title"`
	Title       types.String `tfsdk:"title"`
}

func populateMarkdownFromAPI(pm *panelModel, config kbapi.KbnDashboardPanelTypeMarkdownConfig0) {
	pm.MarkdownConfig = &markdownConfigModel{
		Content:     types.StringPointerValue(config.Content),
		Description: types.StringPointerValue(config.Description),
		HideTitle:   types.BoolPointerValue(config.HideTitle),
		Title:       types.StringPointerValue(config.Title),
	}
}

func buildMarkdownConfig(pm panelModel) kbapi.KbnDashboardPanelTypeMarkdownConfig0 {
	config := kbapi.KbnDashboardPanelTypeMarkdownConfig0{
		Content: pm.MarkdownConfig.Content.ValueStringPointer(),
	}
	if typeutils.IsKnown(pm.MarkdownConfig.Description) {
		config.Description = pm.MarkdownConfig.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(pm.MarkdownConfig.HideTitle) {
		config.HideTitle = pm.MarkdownConfig.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(pm.MarkdownConfig.Title) {
		config.Title = pm.MarkdownConfig.Title.ValueStringPointer()
	}

	return config
}
