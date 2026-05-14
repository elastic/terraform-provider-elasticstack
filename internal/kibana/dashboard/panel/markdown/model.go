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
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// configBranch classifies raw markdown panel `config` JSON for union decode.
type configBranch int

const (
	configBranchUnknown configBranch = iota
	configBranchByValue
	configBranchByReference
)

// classifyConfigFromRoot inspects unmarshalled config JSON (see kbn-dashboard-panel-type-markdown):
// by-value carries string `content` and no library `ref_id`; by-reference carries non-empty `ref_id` and no `content`.
// Ambiguous or unparseable payloads return configBranchUnknown and try-by-value-then-by-reference in the caller.
func classifyConfigFromRoot(configBytes []byte) (configBranch, error) {
	var root map[string]any
	if err := json.Unmarshal(configBytes, &root); err != nil {
		return configBranchUnknown, err
	}
	refID, refOK := root["ref_id"].(string)
	hasRef := refOK && refID != ""
	_, hasContent := root["content"].(string)

	switch {
	case hasRef && !hasContent:
		return configBranchByReference, nil
	case hasContent && !hasRef:
		return configBranchByValue, nil
	default:
		return configBranchUnknown, nil
	}
}

func marshalAndClassifyMarkdownConfig(config kbapi.KbnDashboardPanelTypeMarkdown_Config, enforceClassifier bool, expectedBranch configBranch) ([]byte, bool) {
	raw, mErr := config.MarshalJSON()
	if mErr != nil {
		return nil, false
	}
	if enforceClassifier {
		branch, err := classifyConfigFromRoot(raw)
		if err != nil || branch != expectedBranch {
			return nil, false
		}
	}
	return raw, true
}

func populateFromAPIAttemptByValue(pm *models.PanelModel, tfPanel *models.PanelModel, config kbapi.KbnDashboardPanelTypeMarkdown_Config, enforceClassifier bool) bool {
	raw, ok := marshalAndClassifyMarkdownConfig(config, enforceClassifier, configBranchByValue)
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
	populateFromAPIByValue(pm, tfPanel, config0)
	return true
}

func populateFromAPIAttemptByReference(pm *models.PanelModel, tfPanel *models.PanelModel, config kbapi.KbnDashboardPanelTypeMarkdown_Config, enforceClassifier bool) bool {
	if _, ok := marshalAndClassifyMarkdownConfig(config, enforceClassifier, configBranchByReference); !ok {
		return false
	}
	cfg1, err := config.AsKbnDashboardPanelTypeMarkdownConfig1()
	if err != nil {
		return false
	}
	if cfg1.RefId == "" {
		return false
	}
	populateFromAPIByReference(pm, tfPanel, cfg1)
	return true
}

// PopulateTypedConfigFromAPI maps typed markdown branches when Terraform is not using config-json-only authoring.
//
// Mirrors legacy `dashboardMapPanelFromAPI` markdown branch sequencing (ported from models_markdown_panel.go /
// models_panels.go).
func PopulateTypedConfigFromAPI(pm *models.PanelModel, prior *models.PanelModel, markdownPanel kbapi.KbnDashboardPanelTypeMarkdown, diags *diag.Diagnostics) {
	if panelUsesConfigJSONOnly(prior) {
		return
	}

	rawConfig, rawErr := markdownPanel.Config.MarshalJSON()
	branch := configBranchUnknown
	if rawErr != nil {
		diags.AddWarning(
			"Markdown panel configuration",
			fmt.Sprintf(
				"Could not marshal panel config for markdown branch classification: %v. Using union decode fallback.",
				rawErr,
			),
		)
	} else {
		var err error
		branch, err = classifyConfigFromRoot(rawConfig)
		if err != nil {
			diags.AddWarning(
				"Markdown panel configuration",
				fmt.Sprintf(
					"Could not parse panel config JSON for markdown branch classification: %v. Using union decode fallback.",
					err,
				),
			)
			branch = configBranchUnknown
		}
	}

	decodeMarkdownFails := func() {
		diags.AddError(
			"Invalid markdown panel config",
			"Could not decode markdown panel config as by-value or by-reference.",
		)
	}

	switch branch {
	case configBranchByReference:
		if populateFromAPIAttemptByReference(pm, prior, markdownPanel.Config, true) {
			return
		}
		if !populateFromAPIAttemptByValue(pm, prior, markdownPanel.Config, true) {
			decodeMarkdownFails()
		}
	case configBranchByValue:
		if populateFromAPIAttemptByValue(pm, prior, markdownPanel.Config, true) {
			return
		}
		if !populateFromAPIAttemptByReference(pm, prior, markdownPanel.Config, true) {
			decodeMarkdownFails()
		}
	default:
		if populateFromAPIAttemptByValue(pm, prior, markdownPanel.Config, false) {
			return
		}
		if !populateFromAPIAttemptByReference(pm, prior, markdownPanel.Config, false) {
			decodeMarkdownFails()
		}
	}
}

// panelUsesConfigJSONOnly must stay aligned with dashboard.panelUsesConfigJSONOnly (models_panels.go).
func panelUsesConfigJSONOnly(pm *models.PanelModel) bool {
	if pm == nil || !typeutils.IsKnown(pm.ConfigJSON) {
		return false
	}
	return !panelHasTypedConfig(pm)
}

func panelHasTypedConfig(pm *models.PanelModel) bool {
	return pm.MarkdownConfig != nil ||
		pm.TimeSliderControlConfig != nil ||
		pm.SloBurnRateConfig != nil ||
		pm.SloOverviewConfig != nil ||
		pm.SloErrorBudgetConfig != nil ||
		pm.EsqlControlConfig != nil ||
		pm.OptionsListControlConfig != nil ||
		pm.RangeSliderControlConfig != nil ||
		pm.SyntheticsStatsOverviewConfig != nil ||
		pm.SyntheticsMonitorsConfig != nil ||
		pm.LensDashboardAppConfig != nil ||
		pm.VisConfig != nil ||
		pm.ImageConfig != nil ||
		pm.SloAlertsConfig != nil ||
		pm.DiscoverSessionConfig != nil
}

func populateFromAPIByValue(pm *models.PanelModel, tfPanel *models.PanelModel, config kbapi.KbnDashboardPanelTypeMarkdownConfig0) {
	byValue := func(m *models.MarkdownConfigModel) *models.MarkdownConfigByValueModel { return m.ByValue }
	settings := &models.MarkdownConfigSettingsModel{
		OpenLinksInNewTab: byValueOpenLinksFromAPI(config.Settings.OpenLinksInNewTab, tfPanel),
	}
	pm.MarkdownConfig = &models.MarkdownConfigModel{
		ByValue: &models.MarkdownConfigByValueModel{
			Content:     types.StringValue(config.Content),
			Settings:    settings,
			Description: stringFromAPI(config.Description, tfPanel, byValue, func(bv *models.MarkdownConfigByValueModel) types.String { return bv.Description }),
			HideTitle:   boolFromAPI(config.HideTitle, tfPanel, byValue, func(bv *models.MarkdownConfigByValueModel) types.Bool { return bv.HideTitle }),
			HideBorder:  boolFromAPI(config.HideBorder, tfPanel, byValue, func(bv *models.MarkdownConfigByValueModel) types.Bool { return bv.HideBorder }),
			Title:       stringFromAPI(config.Title, tfPanel, byValue, func(bv *models.MarkdownConfigByValueModel) types.String { return bv.Title }),
		},
	}
}

func populateFromAPIByReference(pm *models.PanelModel, tfPanel *models.PanelModel, config kbapi.KbnDashboardPanelTypeMarkdownConfig1) {
	byReference := func(m *models.MarkdownConfigModel) *models.MarkdownConfigByReferenceModel { return m.ByReference }
	pm.MarkdownConfig = &models.MarkdownConfigModel{
		ByReference: &models.MarkdownConfigByReferenceModel{
			RefID:       types.StringValue(config.RefId),
			Description: stringFromAPI(config.Description, tfPanel, byReference, func(br *models.MarkdownConfigByReferenceModel) types.String { return br.Description }),
			HideTitle:   boolFromAPI(config.HideTitle, tfPanel, byReference, func(br *models.MarkdownConfigByReferenceModel) types.Bool { return br.HideTitle }),
			HideBorder:  boolFromAPI(config.HideBorder, tfPanel, byReference, func(br *models.MarkdownConfigByReferenceModel) types.Bool { return br.HideBorder }),
			Title:       stringFromAPI(config.Title, tfPanel, byReference, func(br *models.MarkdownConfigByReferenceModel) types.String { return br.Title }),
		},
	}
}

func markdownPriorKnown[M any, V attr.Value](
	tfPanel *models.PanelModel,
	branchOf func(*models.MarkdownConfigModel) *M,
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

func stringFromAPI[M any](
	api *string,
	tfPanel *models.PanelModel,
	branchOf func(*models.MarkdownConfigModel) *M,
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

func boolFromAPI[M any](
	api *bool,
	tfPanel *models.PanelModel,
	branchOf func(*models.MarkdownConfigModel) *M,
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

func byValueOpenLinksFromAPI(api *bool, tfPanel *models.PanelModel) types.Bool {
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

func optStringPtr(v types.String) *string {
	if typeutils.IsKnown(v) {
		return v.ValueStringPointer()
	}
	return nil
}

func optBoolPtr(v types.Bool) *bool {
	if typeutils.IsKnown(v) {
		return v.ValueBoolPointer()
	}
	return nil
}

// BuildConfigByValue builds the API by-value markdown payload from Terraform state.
func BuildConfigByValue(pm models.PanelModel) kbapi.KbnDashboardPanelTypeMarkdownConfig0 {
	if pm.MarkdownConfig == nil || pm.MarkdownConfig.ByValue == nil {
		return kbapi.KbnDashboardPanelTypeMarkdownConfig0{}
	}
	bv := pm.MarkdownConfig.ByValue
	config := kbapi.KbnDashboardPanelTypeMarkdownConfig0{
		Content:     bv.Content.ValueString(),
		Description: optStringPtr(bv.Description),
		HideTitle:   optBoolPtr(bv.HideTitle),
		HideBorder:  optBoolPtr(bv.HideBorder),
		Title:       optStringPtr(bv.Title),
	}
	if bv.Settings != nil && typeutils.IsKnown(bv.Settings.OpenLinksInNewTab) {
		config.Settings.OpenLinksInNewTab = bv.Settings.OpenLinksInNewTab.ValueBoolPointer()
	}
	return config
}

// BuildConfigByReference builds the API by-reference markdown payload from Terraform state.
func BuildConfigByReference(pm models.PanelModel) kbapi.KbnDashboardPanelTypeMarkdownConfig1 {
	if pm.MarkdownConfig == nil || pm.MarkdownConfig.ByReference == nil {
		return kbapi.KbnDashboardPanelTypeMarkdownConfig1{}
	}
	br := pm.MarkdownConfig.ByReference
	return kbapi.KbnDashboardPanelTypeMarkdownConfig1{
		RefId:       br.RefID.ValueString(),
		Description: optStringPtr(br.Description),
		HideTitle:   optBoolPtr(br.HideTitle),
		HideBorder:  optBoolPtr(br.HideBorder),
		Title:       optStringPtr(br.Title),
	}
}
