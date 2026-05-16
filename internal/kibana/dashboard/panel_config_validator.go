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
	return "Delegates panel-type validation to registered iface.Handler implementations and retains legacy checks " +
		"for unmigrated dashboard panel kinds: `discover_session` (`discover_session_config`) " +
		"and `vis` (exactly one of `vis_config` or panel `config_json`; typed Lens charts belong under `vis_config.by_value`). " +
		"`lens-dashboard-app` relies on validators on `lens_dashboard_app_config`. " +
		"Optional control panels omit typed config blocks; practitioner-authored panel `config_json` on unsupported types " +
		"is enforced by the allowlist validator on `config_json` alone."
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
	configJSON, visConfig panelConfigValueState,
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
	}

	return diags
}

func (v panelConfigValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
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

	panel := typeValue.ValueString()
	if h := LookupHandler(panel); h != nil {
		if diags := h.ValidatePanelConfig(ctx, attrs, req.Path); len(diags) > 0 {
			resp.Diagnostics.Append(diags...)
		}
	}

	resp.Diagnostics.Append(panelConfigValidateDiags(
		panel,
		panelConfigValueStateFromValue(attrs["config_json"]),
		panelConfigValueStateFromValue(attrs["vis_config"]),
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

// pinnedPanelExpectedTypedControlAttr resolves the typed `*_control_config` block name for a pinned
// panel type by deferring to the registry. Returns ok=false for any panel type that is not registered
// or that does not implement PinnedHandler (i.e. not a pinned-capable control).
func pinnedPanelExpectedTypedControlAttr(panelType string) (attrName string, ok bool) {
	h := LookupHandler(panelType)
	if h == nil || h.PinnedHandler() == nil {
		return "", false
	}
	return h.PanelType() + "_config", true
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
