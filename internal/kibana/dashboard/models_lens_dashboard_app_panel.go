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
	SavedObjectID types.String `tfsdk:"saved_object_id"`
}

// lensDashboardAppTimeRangeModel is the time_range nested block model.
type lensDashboardAppTimeRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

// buildLensDashboardAppConfig writes the TF model into a KbnDashboardPanelTypeLensDashboardApp panel.
func buildLensDashboardAppConfig(pm panelModel, panel *kbapi.KbnDashboardPanelTypeLensDashboardApp) diag.Diagnostics {
	var diags diag.Diagnostics
	cfg := pm.LensDashboardAppConfig
	if cfg == nil {
		return diags
	}

	var apiConfig kbapi.KbnDashboardPanelTypeLensDashboardApp_Config

	if cfg.ByValue != nil {
		config0, d := buildLensDashboardAppByValueConfig(cfg)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if err := apiConfig.FromKbnDashboardPanelTypeLensDashboardAppConfig0(config0); err != nil {
			diags.AddError("Failed to build lens-dashboard-app by-value config", err.Error())
			return diags
		}
	} else if cfg.ByReference != nil {
		config1 := buildLensDashboardAppByReferenceConfig(cfg)
		if err := apiConfig.FromKbnDashboardPanelTypeLensDashboardAppConfig1(config1); err != nil {
			diags.AddError("Failed to build lens-dashboard-app by-reference config", err.Error())
			return diags
		}
	}

	panel.Config = apiConfig
	return diags
}

// buildLensDashboardAppByValueConfig constructs a Config0 (by-value) payload from the TF model.
//
// The new kbapi Config0 is a raw JSON union representing the chart configuration directly.
// attributes_json is treated as the base chart config; references_json is merged in under
// the "references" key if provided.
func buildLensDashboardAppByValueConfig(cfg *lensDashboardAppConfigModel) (kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	byVal := cfg.ByValue

	// Start from the attributes_json blob.
	attrsRaw := []byte(byVal.AttributesJSON.ValueString())

	// If references_json is set, merge the references into the config JSON.
	if typeutils.IsKnown(byVal.ReferencesJSON) && !byVal.ReferencesJSON.IsNull() {
		var refs []kbapi.KbnContentManagementUtilsReferenceSchema
		if err := json.Unmarshal([]byte(byVal.ReferencesJSON.ValueString()), &refs); err != nil {
			diags.AddError("Failed to parse references_json", err.Error())
			return kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0{}, diags
		}
		if len(refs) > 0 {
			attrsRaw, diags = mergeReferencesIntoJSON(attrsRaw, refs)
			if diags.HasError() {
				return kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0{}, diags
			}
		}
	}

	var config0 kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0
	if err := json.Unmarshal(attrsRaw, &config0); err != nil {
		diags.AddError("Failed to build lens-dashboard-app by-value config", err.Error())
		return kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0{}, diags
	}

	return config0, diags
}

// mergeReferencesIntoJSON injects a "references" array into a JSON object.
func mergeReferencesIntoJSON(raw []byte, refs []kbapi.KbnContentManagementUtilsReferenceSchema) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		diags.AddError("Failed to merge references into lens-dashboard-app config", err.Error())
		return raw, diags
	}
	refsBytes, err := json.Marshal(refs)
	if err != nil {
		diags.AddError("Failed to marshal references for lens-dashboard-app config", err.Error())
		return raw, diags
	}
	m["references"] = json.RawMessage(refsBytes)
	merged, err := json.Marshal(m)
	if err != nil {
		diags.AddError("Failed to marshal merged lens-dashboard-app config", err.Error())
		return raw, diags
	}
	return merged, diags
}

// buildLensDashboardAppByReferenceConfig constructs a Config1 (by-reference) payload from the TF model.
func buildLensDashboardAppByReferenceConfig(cfg *lensDashboardAppConfigModel) kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1 {
	byRef := cfg.ByReference

	config1 := kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1{
		RefId:     byRef.SavedObjectID.ValueString(),
		TimeRange: buildLensDashboardAppTimeRange(cfg),
	}

	applyLensDashboardAppSharedFields(cfg, &config1.Title, &config1.Description, &config1.HideTitle, &config1.HideBorder)

	return config1
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
// Mode detection: if the config unmarshals as Config1 (has a non-empty ref_id), it is
// by-reference; otherwise by-value.
//
// tfPanel is the prior TF state/plan panel (may be nil on import). When nil, all fields are
// populated unconditionally.
func populateLensDashboardAppFromAPI(pm *panelModel, tfPanel *panelModel, apiPanel kbapi.KbnDashboardPanelTypeLensDashboardApp) diag.Diagnostics {
	var diags diag.Diagnostics

	// Determine mode by probing the raw config JSON.
	rawBytes, err := apiPanel.Config.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to read lens-dashboard-app panel config", err.Error())
		return diags
	}

	// Use a minimal probe struct to detect mode: Config1 has a required "ref_id" field.
	var probe struct {
		RefID *string `json:"ref_id"`
	}
	if err := json.Unmarshal(rawBytes, &probe); err != nil {
		diags.AddError("Failed to probe lens-dashboard-app panel config mode", err.Error())
		return diags
	}

	var existing *lensDashboardAppConfigModel
	if tfPanel != nil {
		existing = tfPanel.LensDashboardAppConfig
	}

	if probe.RefID != nil && *probe.RefID != "" {
		// By-reference mode.
		config1, err := apiPanel.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig1()
		if err != nil {
			diags.AddError("Failed to read lens-dashboard-app by-reference config", err.Error())
			return diags
		}
		populateLensDashboardAppByReferenceFromAPI(pm, existing, config1)
	} else {
		// By-value mode.
		config0, err := apiPanel.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig0()
		if err != nil {
			diags.AddError("Failed to read lens-dashboard-app by-value config", err.Error())
			return diags
		}
		d := populateLensDashboardAppByValueFromAPI(pm, existing, config0)
		diags.Append(d...)
	}

	return diags
}

// populateLensDashboardAppByValueFromAPI populates pm.LensDashboardAppConfig from a Config0 response.
//
// attributes_json is preserved from the prior plan/state when available to avoid drift from
// API-injected defaults. references are extracted from the config JSON and stored separately.
func populateLensDashboardAppByValueFromAPI(pm *panelModel, existing *lensDashboardAppConfigModel, config0 kbapi.KbnDashboardPanelTypeLensDashboardAppConfig0) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the raw config JSON.
	rawBytes, err := config0.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to read lens-dashboard-app by-value config", err.Error())
		return diags
	}

	// Extract "references" from the JSON, leaving the remainder as attributes_json.
	attrsBytes, refsBytes, d := extractReferencesFromJSON(rawBytes)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Determine the attributes_json to store.
	// When a prior plan/state is available and it has a by_value config, preserve the
	// planned attributes_json value. The Kibana API may enrich the stored attributes with
	// default fields that were not in the user's plan, and storing the API-enriched form
	// back would cause a "Provider produced inconsistent result after apply" error. On
	// import (existing == nil or existing.ByValue == nil) we use the API response instead
	// so the full state is captured.
	var attributesJSON jsontypes.Normalized
	if existing != nil && existing.ByValue != nil && !existing.ByValue.AttributesJSON.IsNull() && !existing.ByValue.AttributesJSON.IsUnknown() {
		attributesJSON = existing.ByValue.AttributesJSON
	} else {
		attributesJSON = jsontypes.NewNormalizedValue(string(attrsBytes))
	}

	cfg := &lensDashboardAppConfigModel{
		ByValue: &lensDashboardAppByValueModel{
			AttributesJSON: attributesJSON,
		},
		ByReference: nil,
	}

	// references_json
	// Kibana may not preserve the "references" key in the stored Config0 JSON even when
	// it was present in the write request. When the API response does not include references,
	// fall back to the prior plan/state value so that Terraform's post-apply consistency
	// check does not fail. On import (existing == nil or existing.ByValue == nil) we use
	// the API response (which will be null when Kibana omits references).
	switch {
	case len(refsBytes) > 0:
		cfg.ByValue.ReferencesJSON = jsontypes.NewNormalizedValue(string(refsBytes))
	case existing != nil && existing.ByValue != nil:
		cfg.ByValue.ReferencesJSON = existing.ByValue.ReferencesJSON
	default:
		cfg.ByValue.ReferencesJSON = jsontypes.NewNormalizedNull()
	}

	// Shared optional fields are not part of the new by-value Config0 union; preserve
	// null for consistency.
	cfg.Title = types.StringNull()
	cfg.Description = types.StringNull()
	cfg.HideTitle = types.BoolNull()
	cfg.HideBorder = types.BoolNull()
	cfg.TimeRange = nil

	pm.LensDashboardAppConfig = cfg
	return diags
}

// extractReferencesFromJSON removes the "references" key from a JSON object and returns
// the stripped JSON and the references JSON separately.
func extractReferencesFromJSON(raw []byte) (attrsJSON []byte, refsJSON []byte, diags diag.Diagnostics) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		diags.AddError("Failed to parse lens-dashboard-app by-value config", err.Error())
		return raw, nil, diags
	}

	refsRaw, hasRefs := m["references"]
	if hasRefs {
		delete(m, "references")
		refsJSON = []byte(refsRaw)
	}

	stripped, err := json.Marshal(m)
	if err != nil {
		diags.AddError("Failed to re-marshal lens-dashboard-app by-value config", err.Error())
		return raw, nil, diags
	}
	return stripped, refsJSON, diags
}

// populateLensDashboardAppByReferenceFromAPI populates pm.LensDashboardAppConfig from a Config1 response.
func populateLensDashboardAppByReferenceFromAPI(pm *panelModel, existing *lensDashboardAppConfigModel, config1 kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1) {
	cfg := &lensDashboardAppConfigModel{
		ByValue: nil,
		ByReference: &lensDashboardAppByReferenceModel{
			SavedObjectID: types.StringValue(config1.RefId),
		},
	}

	populateLensDashboardAppSharedFromAPI(cfg, existing, config1.Title, config1.Description, config1.HideTitle, config1.HideBorder, config1.TimeRange)

	pm.LensDashboardAppConfig = cfg
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

	// time_range: reflect the API response to avoid preserving stale values.
	// Only preserve the null vs set intent from prior state: if the user did not configure
	// time_range (existing.TimeRange == nil), keep it nil regardless of what the API returns.
	if existing.TimeRange != nil {
		if apiTimeRange.From != "" || apiTimeRange.To != "" {
			cfg.TimeRange = &lensDashboardAppTimeRangeModel{
				From: types.StringValue(apiTimeRange.From),
				To:   types.StringValue(apiTimeRange.To),
			}
		} else {
			// API returned no time_range; set to nil to reflect remote state.
			cfg.TimeRange = nil
		}
	} else {
		cfg.TimeRange = nil
	}
}
