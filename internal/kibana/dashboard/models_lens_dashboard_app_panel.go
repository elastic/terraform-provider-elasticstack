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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// lensDashboardAppConfigModel is the Terraform model for the lens_dashboard_app_config block.
type lensDashboardAppConfigModel struct {
	ByValue     *lensDashboardAppByValueModel     `tfsdk:"by_value"`
	ByReference *lensDashboardAppByReferenceModel `tfsdk:"by_reference"`
	Title       types.String                      `tfsdk:"title"`
	Description types.String                      `tfsdk:"description"`
	HideTitle   types.Bool                        `tfsdk:"hide_title"`
	HideBorder  types.Bool                        `tfsdk:"hide_border"`
	TimeRange   *lensDashboardAppTimeRangeModel   `tfsdk:"time_range"`
}

// lensDashboardAppByValueModel is the by-value sub-block model.
type lensDashboardAppByValueModel struct {
	AttributesJSON jsontypes.Normalized `tfsdk:"attributes_json"`
	ReferencesJSON jsontypes.Normalized `tfsdk:"references_json"`
}

// lensDashboardAppByReferenceModel is the by-reference sub-block model.
type lensDashboardAppByReferenceModel struct {
	SavedObjectID types.String         `tfsdk:"saved_object_id"`
	OverridesJSON jsontypes.Normalized `tfsdk:"overrides_json"`
}

// lensDashboardAppTimeRangeModel is the time_range nested block model.
type lensDashboardAppTimeRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

// buildLensDashboardAppConfig writes the TF model into a KbnDashboardPanelLensDashboardApp panel.
func buildLensDashboardAppConfig(pm panelModel, panel *kbapi.KbnDashboardPanelLensDashboardApp) diag.Diagnostics {
	var diags diag.Diagnostics
	cfg := pm.LensDashboardAppConfig
	if cfg == nil {
		return diags
	}

	var apiConfig kbapi.KbnDashboardPanelLensDashboardApp_Config

	if cfg.ByValue != nil {
		config0, d := buildLensDashboardAppByValueConfig(cfg)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if err := apiConfig.FromKbnDashboardPanelLensDashboardAppConfig0(config0); err != nil {
			diags.AddError("Failed to build lens-dashboard-app by-value config", err.Error())
			return diags
		}
	} else if cfg.ByReference != nil {
		config1, d := buildLensDashboardAppByReferenceConfig(cfg)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if err := apiConfig.FromKbnDashboardPanelLensDashboardAppConfig1(config1); err != nil {
			diags.AddError("Failed to build lens-dashboard-app by-reference config", err.Error())
			return diags
		}
	}

	panel.Config = apiConfig
	return diags
}

// buildLensDashboardAppByValueConfig constructs a Config0 (by-value) payload from the TF model.
func buildLensDashboardAppByValueConfig(cfg *lensDashboardAppConfigModel) (kbapi.KbnDashboardPanelLensDashboardAppConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	byVal := cfg.ByValue

	// Unmarshal attributes_json into the LensApiState.
	var attrs kbapi.LensApiState
	if err := json.Unmarshal([]byte(byVal.AttributesJSON.ValueString()), &attrs); err != nil {
		diags.AddError("Failed to parse attributes_json", err.Error())
		return kbapi.KbnDashboardPanelLensDashboardAppConfig0{}, diags
	}

	config0 := kbapi.KbnDashboardPanelLensDashboardAppConfig0{
		Attributes: attrs,
		TimeRange:  buildLensDashboardAppTimeRange(cfg),
	}

	// references_json
	if typeutils.IsKnown(byVal.ReferencesJSON) {
		var refs []kbapi.KbnContentManagementUtilsReferenceSchema
		if err := json.Unmarshal([]byte(byVal.ReferencesJSON.ValueString()), &refs); err != nil {
			diags.AddError("Failed to parse references_json", err.Error())
			return kbapi.KbnDashboardPanelLensDashboardAppConfig0{}, diags
		}
		config0.References = &refs
	}

	applyLensDashboardAppSharedFields(cfg, &config0.Title, &config0.Description, &config0.HideTitle, &config0.HideBorder)

	return config0, diags
}

// buildLensDashboardAppByReferenceConfig constructs a Config1 (by-reference) payload from the TF model.
func buildLensDashboardAppByReferenceConfig(cfg *lensDashboardAppConfigModel) (kbapi.KbnDashboardPanelLensDashboardAppConfig1, diag.Diagnostics) {
	var diags diag.Diagnostics
	byRef := cfg.ByReference

	config1 := kbapi.KbnDashboardPanelLensDashboardAppConfig1{
		RefId:     byRef.SavedObjectID.ValueString(),
		TimeRange: buildLensDashboardAppTimeRange(cfg),
	}

	applyLensDashboardAppSharedFields(cfg, &config1.Title, &config1.Description, &config1.HideTitle, &config1.HideBorder)

	// overrides_json: the API Config1 does not have a dedicated Overrides field in the generated
	// struct, so we store it in the References field as a workaround — but first let us check
	// whether we need to handle it via raw JSON merging.
	// The API schema's Config1 has no "overrides" field in the generated Go struct, so
	// overrides_json is intentionally ignored on write for now (see design note on overrides_json).
	// TODO: if the API adds an Overrides field to Config1, wire it up here.
	_ = byRef.OverridesJSON

	return config1, diags
}

// buildLensDashboardAppTimeRange builds the API time range from the TF model.
// When time_range is not set, a zero-value KbnEsQueryServerTimeRangeSchema is returned
// (the API allows empty from/to strings).
func buildLensDashboardAppTimeRange(cfg *lensDashboardAppConfigModel) kbapi.KbnEsQueryServerTimeRangeSchema {
	if cfg.TimeRange == nil {
		return kbapi.KbnEsQueryServerTimeRangeSchema{}
	}
	return kbapi.KbnEsQueryServerTimeRangeSchema{
		From: cfg.TimeRange.From.ValueString(),
		To:   cfg.TimeRange.To.ValueString(),
	}
}

// applyLensDashboardAppSharedFields copies optional shared fields from the TF model to the API payload pointers.
func applyLensDashboardAppSharedFields(
	cfg *lensDashboardAppConfigModel,
	title, description **string,
	hideTitle, hideBorder **bool,
) {
	if typeutils.IsKnown(cfg.Title) {
		*title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Description) {
		*description = cfg.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.HideTitle) {
		*hideTitle = cfg.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.HideBorder) {
		*hideBorder = cfg.HideBorder.ValueBoolPointer()
	}
}

// populateLensDashboardAppFromAPI reads back a lens-dashboard-app panel config from the API
// response and updates the panel model.
//
// Mode detection: presence of a non-empty "attributes" key in the raw config JSON indicates
// by-value mode (Config0); otherwise by-reference mode (Config1).
//
// tfPanel is the prior TF state/plan panel (may be nil on import). When nil, all fields are
// populated unconditionally.
func populateLensDashboardAppFromAPI(pm *panelModel, tfPanel *panelModel, apiPanel kbapi.KbnDashboardPanelLensDashboardApp) diag.Diagnostics {
	var diags diag.Diagnostics

	// Determine mode by attempting to parse as Config0 (by-value) first.
	// Config0 has a required "attributes" field; Config1 has a required "ref_id" field.
	rawBytes, err := apiPanel.Config.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to read lens-dashboard-app panel config", err.Error())
		return diags
	}

	// Use a minimal probe struct to detect mode.
	var probe struct {
		Attributes *json.RawMessage `json:"attributes"`
		RefID      *string          `json:"ref_id"`
	}
	if err := json.Unmarshal(rawBytes, &probe); err != nil {
		diags.AddError("Failed to probe lens-dashboard-app panel config mode", err.Error())
		return diags
	}

	var existing *lensDashboardAppConfigModel
	if tfPanel != nil {
		existing = tfPanel.LensDashboardAppConfig
	}

	if probe.Attributes != nil {
		// By-value mode.
		config0, err := apiPanel.Config.AsKbnDashboardPanelLensDashboardAppConfig0()
		if err != nil {
			diags.AddError("Failed to read lens-dashboard-app by-value config", err.Error())
			return diags
		}
		d := populateLensDashboardAppByValueFromAPI(pm, existing, config0)
		diags.Append(d...)
	} else {
		// By-reference mode.
		config1, err := apiPanel.Config.AsKbnDashboardPanelLensDashboardAppConfig1()
		if err != nil {
			diags.AddError("Failed to read lens-dashboard-app by-reference config", err.Error())
			return diags
		}
		d := populateLensDashboardAppByReferenceFromAPI(pm, existing, config1)
		diags.Append(d...)
	}

	return diags
}

// populateLensDashboardAppByValueFromAPI populates pm.LensDashboardAppConfig from a Config0 response.
func populateLensDashboardAppByValueFromAPI(pm *panelModel, existing *lensDashboardAppConfigModel, config0 kbapi.KbnDashboardPanelLensDashboardAppConfig0) diag.Diagnostics {
	var diags diag.Diagnostics

	attrsBytes, err := json.Marshal(config0.Attributes)
	if err != nil {
		diags.AddError("Failed to marshal lens-dashboard-app attributes", err.Error())
		return diags
	}

	cfg := &lensDashboardAppConfigModel{
		ByValue: &lensDashboardAppByValueModel{
			AttributesJSON: jsontypes.NewNormalizedValue(string(attrsBytes)),
		},
		ByReference: nil,
	}

	// references_json
	if config0.References != nil && len(*config0.References) > 0 {
		refsBytes, err := json.Marshal(*config0.References)
		if err != nil {
			diags.AddError("Failed to marshal lens-dashboard-app references", err.Error())
			return diags
		}
		cfg.ByValue.ReferencesJSON = jsontypes.NewNormalizedValue(string(refsBytes))
	} else {
		// Preserve null if existing state had null; otherwise set null.
		if existing != nil && existing.ByValue != nil && typeutils.IsKnown(existing.ByValue.ReferencesJSON) {
			cfg.ByValue.ReferencesJSON = jsontypes.NewNormalizedNull()
		} else {
			cfg.ByValue.ReferencesJSON = jsontypes.NewNormalizedNull()
		}
	}

	populateLensDashboardAppSharedFromAPI(cfg, existing, config0.Title, config0.Description, config0.HideTitle, config0.HideBorder, config0.TimeRange)

	pm.LensDashboardAppConfig = cfg
	return diags
}

// populateLensDashboardAppByReferenceFromAPI populates pm.LensDashboardAppConfig from a Config1 response.
func populateLensDashboardAppByReferenceFromAPI(pm *panelModel, existing *lensDashboardAppConfigModel, config1 kbapi.KbnDashboardPanelLensDashboardAppConfig1) diag.Diagnostics {
	var diags diag.Diagnostics

	cfg := &lensDashboardAppConfigModel{
		ByValue: nil,
		ByReference: &lensDashboardAppByReferenceModel{
			SavedObjectID: types.StringValue(config1.RefId),
			// overrides_json: the API Config1 does not carry a typed Overrides field in the
			// generated struct. We preserve the existing state value if present.
			OverridesJSON: jsontypes.NewNormalizedNull(),
		},
	}

	// Preserve overrides_json from existing state if it was set.
	if existing != nil && existing.ByReference != nil && typeutils.IsKnown(existing.ByReference.OverridesJSON) {
		cfg.ByReference.OverridesJSON = existing.ByReference.OverridesJSON
	}

	populateLensDashboardAppSharedFromAPI(cfg, existing, config1.Title, config1.Description, config1.HideTitle, config1.HideBorder, config1.TimeRange)

	pm.LensDashboardAppConfig = cfg
	return diags
}

// populateLensDashboardAppSharedFromAPI populates shared optional fields on the config model.
// Null-preservation semantics: if the prior state had a null optional field, keep it null.
func populateLensDashboardAppSharedFromAPI(
	cfg *lensDashboardAppConfigModel,
	existing *lensDashboardAppConfigModel,
	apiTitle, apiDescription *string,
	apiHideTitle, apiHideBorder *bool,
	apiTimeRange kbapi.KbnEsQueryServerTimeRangeSchema,
) {
	// On import (existing == nil): populate from API unconditionally.
	if existing == nil {
		cfg.Title = types.StringPointerValue(apiTitle)
		cfg.Description = types.StringPointerValue(apiDescription)
		cfg.HideTitle = types.BoolPointerValue(apiHideTitle)
		cfg.HideBorder = types.BoolPointerValue(apiHideBorder)
		if apiTimeRange.From != "" || apiTimeRange.To != "" {
			cfg.TimeRange = &lensDashboardAppTimeRangeModel{
				From: types.StringValue(apiTimeRange.From),
				To:   types.StringValue(apiTimeRange.To),
			}
		}
		return
	}

	// Null-preservation for optional string fields.
	if typeutils.IsKnown(existing.Title) {
		cfg.Title = types.StringPointerValue(apiTitle)
	} else {
		cfg.Title = types.StringNull()
	}
	if typeutils.IsKnown(existing.Description) {
		cfg.Description = types.StringPointerValue(apiDescription)
	} else {
		cfg.Description = types.StringNull()
	}

	// Null-preservation for optional bool fields.
	if typeutils.IsKnown(existing.HideTitle) {
		cfg.HideTitle = types.BoolPointerValue(apiHideTitle)
	} else {
		cfg.HideTitle = types.BoolNull()
	}
	if typeutils.IsKnown(existing.HideBorder) {
		cfg.HideBorder = types.BoolPointerValue(apiHideBorder)
	} else {
		cfg.HideBorder = types.BoolNull()
	}

	// time_range: preserve null if not set in prior state.
	if existing.TimeRange != nil {
		if apiTimeRange.From != "" || apiTimeRange.To != "" {
			cfg.TimeRange = &lensDashboardAppTimeRangeModel{
				From: types.StringValue(apiTimeRange.From),
				To:   types.StringValue(apiTimeRange.To),
			}
		} else {
			cfg.TimeRange = existing.TimeRange
		}
	} else {
		cfg.TimeRange = nil
	}
}
