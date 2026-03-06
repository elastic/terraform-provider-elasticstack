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
	"encoding/json"

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

func (c markdownPanelConfigConverter) handlesAPIPanelConfig(pm *panelModel, panelType string, _ json.RawMessage) bool {
	return (pm == nil || pm.MarkdownConfig != nil) && panelType == "DASHBOARD_MARKDOWN"
}

func (c markdownPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.MarkdownConfig != nil
}

func (c markdownPanelConfigConverter) populateFromAPIPanel(_ context.Context, pm *panelModel, config json.RawMessage) diag.Diagnostics {
	var cfg kbapi.KbnDashboardPanelDASHBOARDMARKDOWN_Config
	if len(config) > 0 {
		if err := cfg.UnmarshalJSON(config); err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
	}

	config0, err := cfg.AsKbnDashboardPanelDASHBOARDMARKDOWNConfig0()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.MarkdownConfig = &markdownConfigModel{
		Content:     types.StringPointerValue(config0.Content),
		Description: types.StringPointerValue(config0.Description),
		HideTitle:   types.BoolPointerValue(config0.HideTitle),
		Title:       types.StringPointerValue(config0.Title),
	}
	if pm.MarkdownConfig.Content.IsNull() {
		pm.MarkdownConfig.Content = types.StringValue("")
	}

	return nil
}

func (c markdownPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *json.RawMessage) diag.Diagnostics {
	content := pm.MarkdownConfig.Content.ValueString()
	config0 := kbapi.KbnDashboardPanelDASHBOARDMARKDOWNConfig0{
		Content: &content,
	}
	if typeutils.IsKnown(pm.MarkdownConfig.Description) {
		config0.Description = pm.MarkdownConfig.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(pm.MarkdownConfig.HideTitle) {
		config0.HideTitle = pm.MarkdownConfig.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(pm.MarkdownConfig.Title) {
		config0.Title = pm.MarkdownConfig.Title.ValueStringPointer()
	}

	var cfg kbapi.KbnDashboardPanelDASHBOARDMARKDOWN_Config
	var diags diag.Diagnostics
	if err := cfg.FromKbnDashboardPanelDASHBOARDMARKDOWNConfig0(config0); err != nil {
		diags.AddError("Failed to build markdown panel config", err.Error())
		return diags
	}

	raw, err := cfg.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal panel config", err.Error())
		return diags
	}

	*apiConfig = json.RawMessage(raw)
	if len(*apiConfig) == 0 {
		diags.AddError("Failed to marshal panel config", "Generated markdown panel config was empty")
	}

	return diags
}
