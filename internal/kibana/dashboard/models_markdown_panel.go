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
	"encoding/json"

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

// markdownConfigBranch classifies raw markdown panel `config` JSON for union decode.
type markdownConfigBranch int

const (
	markdownConfigBranchUnknown markdownConfigBranch = iota
	markdownConfigBranchByValue
	markdownConfigBranchByReference
)

// classifyMarkdownConfigFromRoot inspects unmarshalled config JSON (see kbn-dashboard-panel-type-markdown):
// by-value carries string `content` and no library `ref_id`; by-reference carries non-empty `ref_id` and no `content`.
// Ambiguous or unparseable payloads return markdownConfigBranchUnknown and try-by-value-then-by-reference in the caller.
func classifyMarkdownConfigFromRoot(configBytes []byte) (markdownConfigBranch, error) {
	var root map[string]any
	if err := json.Unmarshal(configBytes, &root); err != nil {
		return markdownConfigBranchUnknown, err
	}
	refID, refOK := root["ref_id"].(string)
	hasRef := refOK && refID != ""
	_, hasContent := root["content"].(string)

	switch {
	case hasRef && !hasContent:
		return markdownConfigBranchByReference, nil
	case hasContent && !hasRef:
		return markdownConfigBranchByValue, nil
	default:
		return markdownConfigBranchUnknown, nil
	}
}

// populateMarkdownFromAPIAttemptByValue decodes config as KbnDashboardPanelTypeMarkdownConfig0 (with JSON fallback).
// When enforceClassifier is true, raw JSON must classify as by-value (disambiguates the markdown union); when false,
// decoding is attempted for unknown-shaped payloads so REQ-010 can fall through to config_json when types fail.
func populateMarkdownFromAPIAttemptByValue(pm *panelModel, tfPanel *panelModel, config kbapi.KbnDashboardPanelTypeMarkdown_Config, enforceClassifier bool) bool {
	raw, mErr := config.MarshalJSON()
	if mErr != nil {
		return false
	}
	if enforceClassifier {
		branch, err := classifyMarkdownConfigFromRoot(raw)
		if err != nil || branch != markdownConfigBranchByValue {
			return false
		}
	}
	config0, err := config.AsKbnDashboardPanelTypeMarkdownConfig0()
	if err != nil {
		var inline kbapi.KbnDashboardPanelTypeMarkdownConfig0
		if json.Unmarshal(raw, &inline) != nil {
			return false
		}
		config0 = inline
	}
	populateMarkdownFromAPIByValue(pm, tfPanel, config0)
	return true
}

// populateMarkdownFromAPIAttemptByReference decodes config as KbnDashboardPanelTypeMarkdownConfig1.
// When enforceClassifier is true, raw JSON must classify as by-reference; when false, decoding is only attempted
// when the payload is a valid by-reference shape (non-empty ref_id after parse).
func populateMarkdownFromAPIAttemptByReference(pm *panelModel, tfPanel *panelModel, config kbapi.KbnDashboardPanelTypeMarkdown_Config, enforceClassifier bool) bool {
	raw, mErr := config.MarshalJSON()
	if mErr != nil {
		return false
	}
	if enforceClassifier {
		branch, err := classifyMarkdownConfigFromRoot(raw)
		if err != nil || branch != markdownConfigBranchByReference {
			return false
		}
	}
	cfg1, err := config.AsKbnDashboardPanelTypeMarkdownConfig1()
	if err != nil {
		return false
	}
	if cfg1.RefId == "" {
		return false
	}
	populateMarkdownFromAPIByReference(pm, tfPanel, cfg1)
	return true
}

// populateMarkdownFromAPIByValue maps API by-value markdown config into Terraform state.
func populateMarkdownFromAPIByValue(pm *panelModel, tfPanel *panelModel, config kbapi.KbnDashboardPanelTypeMarkdownConfig0) {
	settings := &markdownConfigSettingsModel{
		OpenLinksInNewTab: markdownByValueOpenLinksFromAPI(config.Settings.OpenLinksInNewTab, tfPanel),
	}
	pm.MarkdownConfig = &markdownConfigModel{
		ByValue: &markdownConfigByValueModel{
			Content:     types.StringValue(config.Content),
			Settings:    settings,
			Description: markdownByValueStringFromAPI(config.Description, tfPanel, func(bv *markdownConfigByValueModel) types.String { return bv.Description }),
			HideTitle:   markdownByValueBoolFromAPI(config.HideTitle, tfPanel, func(bv *markdownConfigByValueModel) types.Bool { return bv.HideTitle }),
			HideBorder:  markdownByValueBoolFromAPI(config.HideBorder, tfPanel, func(bv *markdownConfigByValueModel) types.Bool { return bv.HideBorder }),
			Title:       markdownByValueStringFromAPI(config.Title, tfPanel, func(bv *markdownConfigByValueModel) types.String { return bv.Title }),
		},
	}
}

// populateMarkdownFromAPIByReference maps API by-reference markdown config into Terraform state.
func populateMarkdownFromAPIByReference(pm *panelModel, tfPanel *panelModel, config kbapi.KbnDashboardPanelTypeMarkdownConfig1) {
	pm.MarkdownConfig = &markdownConfigModel{
		ByReference: &markdownConfigByReferenceModel{
			RefID:       types.StringValue(config.RefId),
			Description: markdownByReferenceStringFromAPI(config.Description, tfPanel, func(br *markdownConfigByReferenceModel) types.String { return br.Description }),
			HideTitle:   markdownByReferenceBoolFromAPI(config.HideTitle, tfPanel, func(br *markdownConfigByReferenceModel) types.Bool { return br.HideTitle }),
			HideBorder:  markdownByReferenceBoolFromAPI(config.HideBorder, tfPanel, func(br *markdownConfigByReferenceModel) types.Bool { return br.HideBorder }),
			Title:       markdownByReferenceStringFromAPI(config.Title, tfPanel, func(br *markdownConfigByReferenceModel) types.String { return br.Title }),
		},
	}
}

// REQ-009 optional strings (see models_slo_error_budget_panel.go).
func markdownByValueStringFromAPI(
	api *string,
	tfPanel *panelModel,
	priorField func(*markdownConfigByValueModel) types.String,
) types.String {
	if (tfPanel == nil || markdownPriorKnownByValueString(tfPanel, priorField)) && api != nil {
		return types.StringValue(*api)
	}
	if tfPanel != nil && tfPanel.MarkdownConfig != nil && tfPanel.MarkdownConfig.ByValue != nil {
		p := priorField(tfPanel.MarkdownConfig.ByValue)
		if typeutils.IsKnown(p) {
			return p
		}
	}
	return types.StringNull()
}

func markdownByValueBoolFromAPI(
	api *bool,
	tfPanel *panelModel,
	priorField func(*markdownConfigByValueModel) types.Bool,
) types.Bool {
	if (tfPanel == nil || markdownPriorKnownByValueBool(tfPanel, priorField)) && api != nil {
		return types.BoolValue(*api)
	}
	if tfPanel != nil && tfPanel.MarkdownConfig != nil && tfPanel.MarkdownConfig.ByValue != nil {
		p := priorField(tfPanel.MarkdownConfig.ByValue)
		if typeutils.IsKnown(p) {
			return p
		}
	}
	return types.BoolNull()
}

func markdownPriorKnownByValueString(tfPanel *panelModel, priorField func(*markdownConfigByValueModel) types.String) bool {
	if tfPanel == nil || tfPanel.MarkdownConfig == nil || tfPanel.MarkdownConfig.ByValue == nil {
		return false
	}
	return typeutils.IsKnown(priorField(tfPanel.MarkdownConfig.ByValue))
}

func markdownPriorKnownByValueBool(tfPanel *panelModel, priorField func(*markdownConfigByValueModel) types.Bool) bool {
	if tfPanel == nil || tfPanel.MarkdownConfig == nil || tfPanel.MarkdownConfig.ByValue == nil {
		return false
	}
	return typeutils.IsKnown(priorField(tfPanel.MarkdownConfig.ByValue))
}

// markdownByValueOpenLinksFromAPI maps settings.open_links_in_new_tab with REQ-009 semantics:
// Kibana defaults this to true; when the practitioner left the attribute null, keep null after
// refresh even if the API echoes true (same pattern as SLO drilldown open_in_new_tab).
func markdownByValueOpenLinksFromAPI(api *bool, tfPanel *panelModel) types.Bool {
	if api != nil {
		var prior types.Bool
		if tfPanel != nil && tfPanel.MarkdownConfig != nil && tfPanel.MarkdownConfig.ByValue != nil && tfPanel.MarkdownConfig.ByValue.Settings != nil {
			prior = tfPanel.MarkdownConfig.ByValue.Settings.OpenLinksInNewTab
		} else {
			prior = types.BoolNull()
		}
		if typeutils.IsKnown(prior) || !*api {
			return types.BoolValue(*api)
		}
		return types.BoolNull()
	}
	if tfPanel != nil && tfPanel.MarkdownConfig != nil && tfPanel.MarkdownConfig.ByValue != nil && tfPanel.MarkdownConfig.ByValue.Settings != nil {
		p := tfPanel.MarkdownConfig.ByValue.Settings.OpenLinksInNewTab
		if typeutils.IsKnown(p) {
			return p
		}
	}
	return types.BoolNull()
}

func markdownByReferenceStringFromAPI(
	api *string,
	tfPanel *panelModel,
	priorField func(*markdownConfigByReferenceModel) types.String,
) types.String {
	if (tfPanel == nil || markdownPriorKnownByReferenceString(tfPanel, priorField)) && api != nil {
		return types.StringValue(*api)
	}
	if tfPanel != nil && tfPanel.MarkdownConfig != nil && tfPanel.MarkdownConfig.ByReference != nil {
		p := priorField(tfPanel.MarkdownConfig.ByReference)
		if typeutils.IsKnown(p) {
			return p
		}
	}
	return types.StringNull()
}

func markdownByReferenceBoolFromAPI(
	api *bool,
	tfPanel *panelModel,
	priorField func(*markdownConfigByReferenceModel) types.Bool,
) types.Bool {
	if (tfPanel == nil || markdownPriorKnownByReferenceBool(tfPanel, priorField)) && api != nil {
		return types.BoolValue(*api)
	}
	if tfPanel != nil && tfPanel.MarkdownConfig != nil && tfPanel.MarkdownConfig.ByReference != nil {
		p := priorField(tfPanel.MarkdownConfig.ByReference)
		if typeutils.IsKnown(p) {
			return p
		}
	}
	return types.BoolNull()
}

func markdownPriorKnownByReferenceString(tfPanel *panelModel, priorField func(*markdownConfigByReferenceModel) types.String) bool {
	if tfPanel == nil || tfPanel.MarkdownConfig == nil || tfPanel.MarkdownConfig.ByReference == nil {
		return false
	}
	return typeutils.IsKnown(priorField(tfPanel.MarkdownConfig.ByReference))
}

func markdownPriorKnownByReferenceBool(tfPanel *panelModel, priorField func(*markdownConfigByReferenceModel) types.Bool) bool {
	if tfPanel == nil || tfPanel.MarkdownConfig == nil || tfPanel.MarkdownConfig.ByReference == nil {
		return false
	}
	return typeutils.IsKnown(priorField(tfPanel.MarkdownConfig.ByReference))
}

// buildMarkdownConfig builds the API by-value markdown payload from Terraform.
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

func buildMarkdownConfigByReference(pm panelModel) kbapi.KbnDashboardPanelTypeMarkdownConfig1 {
	br := pm.MarkdownConfig.ByReference
	if br == nil {
		return kbapi.KbnDashboardPanelTypeMarkdownConfig1{}
	}
	config := kbapi.KbnDashboardPanelTypeMarkdownConfig1{
		RefId: br.RefID.ValueString(),
	}
	if typeutils.IsKnown(br.Description) {
		config.Description = br.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(br.HideTitle) {
		config.HideTitle = br.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(br.HideBorder) {
		config.HideBorder = br.HideBorder.ValueBoolPointer()
	}
	if typeutils.IsKnown(br.Title) {
		config.Title = br.Title.ValueStringPointer()
	}
	return config
}
