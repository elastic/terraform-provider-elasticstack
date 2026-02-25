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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type markdownConfigModel struct {
	Content     types.String `tfsdk:"content"`
	Description types.String `tfsdk:"description"`
	HideTitle   types.Bool   `tfsdk:"hide_title"`
	Title       types.String `tfsdk:"title"`
}

type markdownPanelConfigConverter struct{}

func (c markdownPanelConfigConverter) handlesAPIPanelConfig(pm *panelModel, panelType string, _ kbapi.DashboardPanelItem_Config) bool {
	return (pm == nil || pm.MarkdownConfig != nil) && panelType == "DASHBOARD_MARKDOWN"
}

func (c markdownPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.MarkdownConfig != nil
}

func (c markdownPanelConfigConverter) populateFromAPIPanel(_ context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	config0, err := config.AsDashboardPanelItemConfig0()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.MarkdownConfig = &markdownConfigModel{
		Content:     types.StringValue(config0.Content),
		Description: types.StringPointerValue(config0.Description),
		HideTitle:   types.BoolPointerValue(config0.HideTitle),
		Title:       types.StringPointerValue(config0.Title),
	}

	return nil
}

func (c markdownPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	config0 := kbapi.DashboardPanelItemConfig0{
		Content: pm.MarkdownConfig.Content.ValueString(),
	}
	if typeutils.IsKnown(pm.MarkdownConfig.Description) {
		config0.Description = new(pm.MarkdownConfig.Description.ValueString())
	}
	if typeutils.IsKnown(pm.MarkdownConfig.HideTitle) {
		config0.HideTitle = new(pm.MarkdownConfig.HideTitle.ValueBool())
	}
	if typeutils.IsKnown(pm.MarkdownConfig.Title) {
		config0.Title = new(pm.MarkdownConfig.Title.ValueString())
	}

	var diags diag.Diagnostics
	if err := apiConfig.FromDashboardPanelItemConfig0(config0); err != nil {
		diags.AddError("Failed to marshal panel config", err.Error())
	}

	return diags
}
