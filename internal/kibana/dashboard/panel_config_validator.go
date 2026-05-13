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
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = panelConfigValidator{}

// panelConfigValidator enforces panel-type-specific config requirements.
type panelConfigValidator struct{}

func (panelConfigValidator) Description(_ context.Context) string {
	return "Ensures markdown panels configure `markdown_config` or `config_json`, " +
		"`vis` panels configure exactly one of `vis_config` or `config_json` (typed chart blocks belong under `vis_config.by_value`), " +
		"`slo_burn_rate` panels configure `slo_burn_rate_config`, " +
		"`time_slider_control` panels use `time_slider_control_config` or omit config, " +
		"`image` panels configure `image_config`, " +
		"`slo_alerts` panels configure `slo_alerts_config`, " +
		"`discover_session` panels configure `discover_session_config`, " +
		"`slo_overview` panels configure `slo_overview_config`, " +
		"and `slo_error_budget` panels configure `slo_error_budget_config`. " +
		"`lens-dashboard-app` is validated by per-attribute validators on `lens_dashboard_app_config` " +
		"(e.g. type allowlist, required block, and conflicts with other panel config attributes); " +
		"this panel-level check does not duplicate those rules. " +
		"Practitioner-authored `config_json` for `time_slider_control` is rejected only by the `config_json` " +
		"attribute validator (type allowlist) to avoid duplicate diagnostics."
}

func (v panelConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

type panelConfigValueState struct {
	Set     bool
	Unknown bool
}

func panelConfigValueStateFromValue(value attr.Value) panelConfigValueState {
	if value == nil {
		return panelConfigValueState{}
	}
	if value.IsUnknown() {
		return panelConfigValueState{Unknown: true}
	}
	if value.IsNull() {
		return panelConfigValueState{}
	}
	return panelConfigValueState{Set: true}
}

func panelConfigSelectionList() string {
	options := []string{"`vis_config`", "`config_json`"}
	return strings.Join(options, ", ")
}

func panelConfigValidateDiags(
	panelType string,
	markdownConfig, configJSON, visConfig, sloBurnRateConfig, sloErrorBudgetConfig panelConfigValueState,
	sloOverviewConfig panelConfigValueState,
	imageConfig panelConfigValueState,
	sloAlertsConfig panelConfigValueState,
	discoverSessionConfig panelConfigValueState,
	attrPath *path.Path,
) diag.Diagnostics {
	var diags diag.Diagnostics
	add := func(summary, detail string) {
		if attrPath != nil {
			diags.AddAttributeError(*attrPath, summary, detail)
			return
		}
		diags.AddError(summary, detail)
	}

	switch panelType {
	case panelTypeDiscoverSession:
		if discoverSessionConfig.Set {
			return diags
		}
		if discoverSessionConfig.Unknown {
			return diags
		}
		add("Missing discover_session panel configuration", "Discover session panels require `discover_session_config`.")
	case panelTypeImage:
		if imageConfig.Set {
			return diags
		}
		if imageConfig.Unknown {
			return diags
		}
		add("Missing image panel configuration", "Image panels require `image_config`.")
	case panelTypeSloAlerts:
		if sloAlertsConfig.Set {
			return diags
		}
		if sloAlertsConfig.Unknown {
			return diags
		}
		add("Missing SLO alerts panel configuration", "SLO alerts panels require `slo_alerts_config`.")
	case panelTypeSloOverview:
		if sloOverviewConfig.Set {
			return diags
		}
		if sloOverviewConfig.Unknown {
			return diags
		}
		add("Missing SLO overview panel configuration", "SLO overview panels require `slo_overview_config`.")
	case panelTypeMarkdown:
		if markdownConfig.Set || configJSON.Set {
			return diags
		}
		if markdownConfig.Unknown || configJSON.Unknown {
			return diags
		}
		add("Missing markdown panel configuration", "Markdown panels require either `markdown_config` or `config_json`.")
	case panelTypeVis:
		setCount := 0
		hasUnknown := configJSON.Unknown || visConfig.Unknown
		if configJSON.Set {
			setCount++
		}
		if visConfig.Set {
			setCount++
		}

		if setCount == 1 {
			return diags
		}
		if setCount == 0 && hasUnknown {
			return diags
		}

		detail := fmt.Sprintf("Panels with `type = \"vis\"` require exactly one of %s.", panelConfigSelectionList())
		if setCount == 0 {
			add("Missing vis panel configuration", detail)
			return diags
		}
		add("Invalid vis panel configuration", detail)
	case panelTypeSloBurnRate:
		if sloBurnRateConfig.Set {
			return diags
		}
		if sloBurnRateConfig.Unknown {
			return diags
		}
		add("Missing SLO burn rate panel configuration", "SLO burn rate panels require `slo_burn_rate_config`.")
	case panelTypeSloErrorBudget:
		if sloErrorBudgetConfig.Set || sloErrorBudgetConfig.Unknown {
			return diags
		}
		add("Missing slo_error_budget panel configuration", "SLO error budget panels require `slo_error_budget_config`.")
	}

	return diags
}

func (v panelConfigValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	typeAttr := attrs["type"]
	if typeAttr == nil || typeAttr.IsNull() || typeAttr.IsUnknown() {
		return
	}

	typeValue, ok := typeAttr.(interface{ ValueString() string })
	if !ok {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid panel type", "The panel `type` must be a string value.")
		return
	}

	resp.Diagnostics.Append(panelConfigValidateDiags(
		typeValue.ValueString(),
		panelConfigValueStateFromValue(attrs["markdown_config"]),
		panelConfigValueStateFromValue(attrs["config_json"]),
		panelConfigValueStateFromValue(attrs["vis_config"]),
		panelConfigValueStateFromValue(attrs["slo_burn_rate_config"]),
		panelConfigValueStateFromValue(attrs["slo_error_budget_config"]),
		panelConfigValueStateFromValue(attrs["slo_overview_config"]),
		panelConfigValueStateFromValue(attrs["image_config"]),
		panelConfigValueStateFromValue(attrs["slo_alerts_config"]),
		panelConfigValueStateFromValue(attrs["discover_session_config"]),
		&req.Path,
	)...)
}

var _ validator.Object = pinnedPanelControlValidator{}

type pinnedPanelControlValidator struct{}

func (pinnedPanelControlValidator) Description(_ context.Context) string {
	return "Ensures each pinned panel entry sets exactly one `*_control_config` block matching `type` " +
		"(only the four dashboard control kinds allowed for `pinned_panels`)."
}

func (v pinnedPanelControlValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func pinnedPanelExpectedTypedControlAttr(panelType string) (attrName string, ok bool) {
	switch panelType {
	case panelTypeOptionsListControl:
		return "options_list_control_config", true
	case panelTypeRangeSlider:
		return "range_slider_control_config", true
	case panelTypeTimeSlider:
		return "time_slider_control_config", true
	case panelTypeEsqlControl:
		return "esql_control_config", true
	default:
		return "", false
	}
}

const pinnedPanelAllowedTypesDetail = "`options_list_control`, `range_slider_control`, `time_slider_control`, or `esql_control`"

func (v pinnedPanelControlValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	typeAttr := attrs["type"]

	var panelType string
	var typeKnown bool
	if typeAttr != nil && !typeAttr.IsNull() && !typeAttr.IsUnknown() {
		typeValue, ok := typeAttr.(interface{ ValueString() string })
		if !ok {
			resp.Diagnostics.AddAttributeError(req.Path.AtName("type"), "Invalid pinned panel entry type", "The `type` attribute must be a string value.")
			return
		}
		panelType = typeValue.ValueString()
		typeKnown = true
	}

	states := make(map[string]panelConfigValueState, len(pinnedPanelControlConfigNames))
	setAttrs := make([]string, 0, len(pinnedPanelControlConfigNames))
	anyUnknownSlot := false
	for _, name := range pinnedPanelControlConfigNames {
		st := panelConfigValueStateFromValue(attrs[name])
		states[name] = st
		if st.Unknown {
			anyUnknownSlot = true
		}
		if st.Set {
			setAttrs = append(setAttrs, name)
		}
	}
	setCount := len(setAttrs)

	if typeKnown {
		if _, valid := pinnedPanelExpectedTypedControlAttr(panelType); !valid {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("type"),
				"Invalid pinned panel entry type",
				fmt.Sprintf("Pinned panel entries only support dashboard controls in the control bar; `type` must be %s, got %q.",
					pinnedPanelAllowedTypesDetail,
					panelType,
				),
			)
			return
		}
	}

	if setCount >= 2 {
		quoted := make([]string, len(setAttrs))
		for i, name := range setAttrs {
			quoted[i] = "`" + name + "`"
		}
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid pinned panel entry configuration",
			fmt.Sprintf("Pinned panel entry must set exactly one typed control configuration block; found %s.", strings.Join(quoted, ", ")),
		)
		return
	}

	if setCount == 1 {
		if !typeKnown {
			return
		}
		expectedAttr, _ := pinnedPanelExpectedTypedControlAttr(panelType)
		if setAttrs[0] != expectedAttr {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName(setAttrs[0]),
				"Pinned panel control does not match type",
				fmt.Sprintf("Pinned panel entry has `type = %q` but sets `%s`; use `%s` instead.", panelType, setAttrs[0], expectedAttr),
			)
			return
		}
		for _, name := range pinnedPanelControlConfigNames {
			if name == expectedAttr {
				continue
			}
			if states[name].Unknown {
				return
			}
		}
		return
	}

	if anyUnknownSlot {
		return
	}
	if !typeKnown {
		return
	}

	expectedAttr, _ := pinnedPanelExpectedTypedControlAttr(panelType)
	resp.Diagnostics.AddAttributeError(
		req.Path.AtName(expectedAttr),
		"Missing pinned panel control configuration",
		fmt.Sprintf("Pinned panel entry with `type = %q` must set `%s`.", panelType, expectedAttr),
	)
}
