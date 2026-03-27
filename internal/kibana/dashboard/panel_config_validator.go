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

var lensPanelConfigNames = []string{
	"xy_chart_config",
	"treemap_config",
	"mosaic_config",
	"datatable_config",
	"tagcloud_config",
	"heatmap_config",
	"waffle_config",
	"region_map_config",
	"gauge_config",
	"metric_chart_config",
	"pie_chart_config",
	"legacy_metric_config",
}

// panelConfigValidator enforces panel-type-specific config requirements.
type panelConfigValidator struct{}

func (panelConfigValidator) Description(_ context.Context) string {
	return "Ensures markdown panels configure `markdown_config` or `config_json`, " +
		"lens panels configure exactly one lens config block or `config_json`, " +
		"`slo_burn_rate` panels configure `slo_burn_rate_config`, " +
		"`time_slider_control` panels use `time_slider_control_config` or omit config, " +
		"and `slo_overview` panels configure `slo_overview_config`. " +
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
	options := make([]string, 0, len(lensPanelConfigNames)+1)
	options = append(options, "`config_json`")
	for _, name := range lensPanelConfigNames {
		options = append(options, fmt.Sprintf("`%s`", name))
	}
	return strings.Join(options, ", ")
}

func panelConfigValidateDiags(
	panelType string,
	markdownConfig, configJSON, sloBurnRateConfig panelConfigValueState,
	lensConfigs map[string]panelConfigValueState,
	sloOverviewConfig panelConfigValueState,
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
	case panelTypeLens:
		setCount := 0
		hasUnknown := configJSON.Unknown
		if configJSON.Set {
			setCount++
		}
		for _, name := range lensPanelConfigNames {
			state := lensConfigs[name]
			if state.Set {
				setCount++
			}
			hasUnknown = hasUnknown || state.Unknown
		}

		if setCount == 1 {
			return diags
		}
		if setCount == 0 && hasUnknown {
			return diags
		}

		detail := fmt.Sprintf("Lens panels require exactly one of %s.", panelConfigSelectionList())
		if setCount == 0 {
			add("Missing lens panel configuration", detail)
			return diags
		}
		add("Invalid lens panel configuration", detail)
	case panelTypeSloBurnRate:
		if sloBurnRateConfig.Set {
			return diags
		}
		if sloBurnRateConfig.Unknown {
			return diags
		}
		add("Missing SLO burn rate panel configuration", "SLO burn rate panels require `slo_burn_rate_config`.")
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

	lensConfigs := make(map[string]panelConfigValueState, len(lensPanelConfigNames))
	for _, name := range lensPanelConfigNames {
		lensConfigs[name] = panelConfigValueStateFromValue(attrs[name])
	}

	resp.Diagnostics.Append(panelConfigValidateDiags(
		typeValue.ValueString(),
		panelConfigValueStateFromValue(attrs["markdown_config"]),
		panelConfigValueStateFromValue(attrs["config_json"]),
		panelConfigValueStateFromValue(attrs["slo_burn_rate_config"]),
		lensConfigs,
		panelConfigValueStateFromValue(attrs["slo_overview_config"]),
		&req.Path,
	)...)
}
