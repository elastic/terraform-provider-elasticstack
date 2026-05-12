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
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

// marshalAndClassifyMarkdownConfig is the shared prelude for the two populateMarkdownFromAPIAttempt* functions.
// It marshals config to JSON and, when enforceClassifier is true, verifies that the payload classifies as expectedBranch.
// Returns (raw JSON, true) on success, (nil, false) when marshalling fails or the classifier rejects the payload.
func marshalAndClassifyMarkdownConfig(config kbapi.KbnDashboardPanelTypeMarkdown_Config, enforceClassifier bool, expectedBranch markdownConfigBranch) ([]byte, bool) {
	raw, mErr := config.MarshalJSON()
	if mErr != nil {
		return nil, false
	}
	if enforceClassifier {
		branch, err := classifyMarkdownConfigFromRoot(raw)
		if err != nil || branch != expectedBranch {
			return nil, false
		}
	}
	return raw, true
}

// populateMarkdownFromAPIAttemptByValue decodes config as KbnDashboardPanelTypeMarkdownConfig0 (with JSON fallback).
// When enforceClassifier is true, raw JSON must classify as by-value (disambiguates the markdown union); when false,
// decoding is attempted for unknown-shaped payloads so REQ-010 can fall through to config_json when types fail.
func populateMarkdownFromAPIAttemptByValue(pm *panelModel, tfPanel *panelModel, config kbapi.KbnDashboardPanelTypeMarkdown_Config, enforceClassifier bool) bool {
	raw, ok := marshalAndClassifyMarkdownConfig(config, enforceClassifier, markdownConfigBranchByValue)
	if !ok {
		return false
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
	if _, ok := marshalAndClassifyMarkdownConfig(config, enforceClassifier, markdownConfigBranchByReference); !ok {
		return false
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
	byValue := func(m *markdownConfigModel) *markdownConfigByValueModel { return m.ByValue }
	settings := &markdownConfigSettingsModel{
		OpenLinksInNewTab: markdownByValueOpenLinksFromAPI(config.Settings.OpenLinksInNewTab, tfPanel),
	}
	pm.MarkdownConfig = &markdownConfigModel{
		ByValue: &markdownConfigByValueModel{
			Content:     types.StringValue(config.Content),
			Settings:    settings,
			Description: markdownStringFromAPI(config.Description, tfPanel, byValue, func(bv *markdownConfigByValueModel) types.String { return bv.Description }),
			HideTitle:   markdownBoolFromAPI(config.HideTitle, tfPanel, byValue, func(bv *markdownConfigByValueModel) types.Bool { return bv.HideTitle }),
			HideBorder:  markdownBoolFromAPI(config.HideBorder, tfPanel, byValue, func(bv *markdownConfigByValueModel) types.Bool { return bv.HideBorder }),
			Title:       markdownStringFromAPI(config.Title, tfPanel, byValue, func(bv *markdownConfigByValueModel) types.String { return bv.Title }),
		},
	}
}

// populateMarkdownFromAPIByReference maps API by-reference markdown config into Terraform state.
func populateMarkdownFromAPIByReference(pm *panelModel, tfPanel *panelModel, config kbapi.KbnDashboardPanelTypeMarkdownConfig1) {
	byReference := func(m *markdownConfigModel) *markdownConfigByReferenceModel { return m.ByReference }
	pm.MarkdownConfig = &markdownConfigModel{
		ByReference: &markdownConfigByReferenceModel{
			RefID:       types.StringValue(config.RefId),
			Description: markdownStringFromAPI(config.Description, tfPanel, byReference, func(br *markdownConfigByReferenceModel) types.String { return br.Description }),
			HideTitle:   markdownBoolFromAPI(config.HideTitle, tfPanel, byReference, func(br *markdownConfigByReferenceModel) types.Bool { return br.HideTitle }),
			HideBorder:  markdownBoolFromAPI(config.HideBorder, tfPanel, byReference, func(br *markdownConfigByReferenceModel) types.Bool { return br.HideBorder }),
			Title:       markdownStringFromAPI(config.Title, tfPanel, byReference, func(br *markdownConfigByReferenceModel) types.String { return br.Title }),
		},
	}
}

// markdownPriorKnown reports whether the prior TF state for a markdown branch field is a known value.
// M is the branch model type; V is an attr.Value field type (types.String or types.Bool).
func markdownPriorKnown[M any, V attr.Value](
	tfPanel *panelModel,
	branchOf func(*markdownConfigModel) *M,
	priorField func(*M) V,
) bool {
	if tfPanel == nil || tfPanel.MarkdownConfig == nil {
		return false
	}
	b := branchOf(tfPanel.MarkdownConfig)
	if b == nil {
		return false
	}
	return typeutils.IsKnown(priorField(b))
}

// markdownStringFromAPI maps an optional API string to a Terraform types.String,
// applying REQ-009 prior-state semantics for the given branch of the markdown config union.
func markdownStringFromAPI[M any](
	api *string,
	tfPanel *panelModel,
	branchOf func(*markdownConfigModel) *M,
	priorField func(*M) types.String,
) types.String {
	if (tfPanel == nil || markdownPriorKnown(tfPanel, branchOf, priorField)) && api != nil {
		return types.StringValue(*api)
	}
	if tfPanel != nil && tfPanel.MarkdownConfig != nil {
		b := branchOf(tfPanel.MarkdownConfig)
		if b != nil {
			p := priorField(b)
			if typeutils.IsKnown(p) {
				return p
			}
		}
	}
	return types.StringNull()
}

// markdownBoolFromAPI maps an optional API bool to a Terraform types.Bool,
// applying REQ-009 prior-state semantics for the given branch of the markdown config union.
func markdownBoolFromAPI[M any](
	api *bool,
	tfPanel *panelModel,
	branchOf func(*markdownConfigModel) *M,
	priorField func(*M) types.Bool,
) types.Bool {
	if (tfPanel == nil || markdownPriorKnown(tfPanel, branchOf, priorField)) && api != nil {
		return types.BoolValue(*api)
	}
	if tfPanel != nil && tfPanel.MarkdownConfig != nil {
		b := branchOf(tfPanel.MarkdownConfig)
		if b != nil {
			p := priorField(b)
			if typeutils.IsKnown(p) {
				return p
			}
		}
	}
	return types.BoolNull()
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

// markdownOptStringPtr returns a *string when v is a known value, nil otherwise.
func markdownOptStringPtr(v types.String) *string {
	if typeutils.IsKnown(v) {
		return v.ValueStringPointer()
	}
	return nil
}

// markdownOptBoolPtr returns a *bool when v is a known value, nil otherwise.
func markdownOptBoolPtr(v types.Bool) *bool {
	if typeutils.IsKnown(v) {
		return v.ValueBoolPointer()
	}
	return nil
}

// buildMarkdownConfig builds the API by-value markdown payload from Terraform.
func buildMarkdownConfig(pm panelModel) kbapi.KbnDashboardPanelTypeMarkdownConfig0 {
	if pm.MarkdownConfig == nil || pm.MarkdownConfig.ByValue == nil {
		return kbapi.KbnDashboardPanelTypeMarkdownConfig0{}
	}
	bv := pm.MarkdownConfig.ByValue
	config := kbapi.KbnDashboardPanelTypeMarkdownConfig0{
		Content:     bv.Content.ValueString(),
		Description: markdownOptStringPtr(bv.Description),
		HideTitle:   markdownOptBoolPtr(bv.HideTitle),
		HideBorder:  markdownOptBoolPtr(bv.HideBorder),
		Title:       markdownOptStringPtr(bv.Title),
	}
	if bv.Settings != nil && typeutils.IsKnown(bv.Settings.OpenLinksInNewTab) {
		config.Settings.OpenLinksInNewTab = bv.Settings.OpenLinksInNewTab.ValueBoolPointer()
	}
	return config
}

// buildMarkdownConfigByReference builds the API by-reference markdown payload from Terraform.
func buildMarkdownConfigByReference(pm panelModel) kbapi.KbnDashboardPanelTypeMarkdownConfig1 {
	if pm.MarkdownConfig == nil || pm.MarkdownConfig.ByReference == nil {
		return kbapi.KbnDashboardPanelTypeMarkdownConfig1{}
	}
	br := pm.MarkdownConfig.ByReference
	return kbapi.KbnDashboardPanelTypeMarkdownConfig1{
		RefId:       br.RefID.ValueString(),
		Description: markdownOptStringPtr(br.Description),
		HideTitle:   markdownOptBoolPtr(br.HideTitle),
		HideBorder:  markdownOptBoolPtr(br.HideBorder),
		Title:       markdownOptStringPtr(br.Title),
	}
}
