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

package markdown

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const panelType = "markdown"

// Handler implements iface.Handler for markdown dashboard panels (typed `markdown_config` and/or panel `config_json`).
type Handler struct{}

func (Handler) PanelType() string                 { return panelType }
func (Handler) SchemaAttribute() schema.Attribute { return SchemaAttribute() }
func (Handler) ClassifyJSON(config map[string]any) bool {
	if config == nil {
		return false
	}
	content, has := config["content"]
	if !has {
		return false
	}
	_, ok := content.(string)
	return ok
}

func (Handler) PopulateJSONDefaults(config map[string]any) map[string]any {
	if config == nil {
		return config
	}
	if content, hasContent := config["content"]; hasContent {
		if _, ok := content.(string); ok {
			settings, _ := config["settings"].(map[string]any)
			if settings == nil {
				settings = map[string]any{}
			}
			if _, exists := settings["open_links_in_new_tab"]; !exists {
				settings["open_links_in_new_tab"] = true
			}
			config["settings"] = settings
		}
	}
	return config
}

func (Handler) PinnedHandler() iface.PinnedHandler { return nil }

func (Handler) AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel) {
	_, _, _ = ctx, plan, state
}

// FromAPI populates Terraform panel state from a markdown panel API item.
func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	markdownPanel, err := item.AsKbnDashboardPanelTypeMarkdown()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Dashboard panel decode", err.Error())
		return d
	}

	var diags diag.Diagnostics
	pm.Grid = panelkit.GridFromAPI(markdownPanel.Grid.X, markdownPanel.Grid.Y, markdownPanel.Grid.W, markdownPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(markdownPanel.Id)

	PopulateTypedConfigFromAPI(pm, prior, markdownPanel, &diags)

	configBytes, err := markdownPanel.Config.MarshalJSON()
	if err == nil {
		configJSON := newConfigJSON(panelJSONPopulateDefaults, string(configBytes))
		if prior != nil {
			configJSON = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, prior.ConfigJSON, configJSON, &diags)
		}
		pm.ConfigJSON = configJSON
	}

	return diags
}

// ToAPI serializes markdown panel Terraform state into kbapi union JSON.
func (Handler) ToAPI(pm models.PanelModel, _ *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var (
		diags     diag.Diagnostics
		panelItem kbapi.DashboardPanelItem
	)
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)

	switch {
	case pm.MarkdownConfig != nil:
		switch {
		case pm.MarkdownConfig.ByReference != nil:
			config1 := BuildConfigByReference(pm)
			var config kbapi.KbnDashboardPanelTypeMarkdown_Config
			if err := config.FromKbnDashboardPanelTypeMarkdownConfig1(config1); err != nil {
				return kbapi.DashboardPanelItem{}, diagutil.FrameworkDiagFromError(err)
			}
			panel := kbapi.KbnDashboardPanelTypeMarkdown{Config: config, Grid: grid, Id: id}
			if err := panelItem.FromKbnDashboardPanelTypeMarkdown(panel); err != nil {
				diags.AddError("Failed to create markdown panel", err.Error())
			}
			return panelItem, diags
		case pm.MarkdownConfig.ByValue != nil:
			config0 := BuildConfigByValue(pm)
			var config kbapi.KbnDashboardPanelTypeMarkdown_Config
			if err := config.FromKbnDashboardPanelTypeMarkdownConfig0(config0); err != nil {
				return kbapi.DashboardPanelItem{}, diagutil.FrameworkDiagFromError(err)
			}
			panel := kbapi.KbnDashboardPanelTypeMarkdown{Config: config, Grid: grid, Id: id}
			if err := panelItem.FromKbnDashboardPanelTypeMarkdown(panel); err != nil {
				diags.AddError("Failed to create markdown panel", err.Error())
			}
			return panelItem, diags
		default:
			diags.AddError(
				"Invalid markdown_config",
				"Set `markdown_config.by_value` or `markdown_config.by_reference` (exactly one).",
			)
			return kbapi.DashboardPanelItem{}, diags
		}

	case typeutils.IsKnown(pm.ConfigJSON) && !pm.ConfigJSON.IsNull():
		configJSONBytes := []byte(pm.ConfigJSON.ValueString())
		var config kbapi.KbnDashboardPanelTypeMarkdown_Config
		if err := config.UnmarshalJSON(configJSONBytes); err != nil {
			diags.AddError("Failed to unmarshal markdown panel config", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		panel := kbapi.KbnDashboardPanelTypeMarkdown{Config: config, Grid: grid, Id: id}
		if err := panelItem.FromKbnDashboardPanelTypeMarkdown(panel); err != nil {
			diags.AddError("Failed to create markdown panel", err.Error())
		}
		return panelItem, diags
	}

	diags.AddError("Unsupported markdown panel configuration", "No `markdown_config` block or panel-level `config_json` was provided.")
	return kbapi.DashboardPanelItem{}, diags
}

// ValidatePanelConfig enforces markdown panel presence/exclusion rules at the dashboard panel object scope.
func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	const mdBlock = panelType + "_config"
	md := attrs[mdBlock]
	cj := attrs["config_json"]
	mdSet := panelkit.AttrConcreteSet(md)
	mdUnk := panelkit.AttrUnknown(md)
	cjSet := panelkit.AttrConcreteSet(cj)
	cjUnk := panelkit.AttrUnknown(cj)
	if mdSet && cjSet {
		diags.AddAttributeError(
			attrPath,
			"Invalid markdown panel configuration",
			"Markdown panels cannot set both `markdown_config` and panel-level `config_json`; use exactly one.",
		)
		return diags
	}
	if mdSet || cjSet {
		return diags
	}
	if mdUnk || cjUnk {
		return diags
	}
	diags.AddAttributeError(
		attrPath,
		"Missing markdown panel configuration",
		"Markdown panels require either `markdown_config` or `config_json`.",
	)
	return diags
}
