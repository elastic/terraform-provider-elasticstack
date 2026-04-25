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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func defaultOpaqueRootJSON(v any) any { return v }

// lensDashboardAPIGrid is the wire shape used by dashboard panel toAPI.
type lensDashboardAPIGrid struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
}

// lensDashboardAppToAPI converts a lens-dashboard-app panel to the Kibana API model.
func lensDashboardAppToAPI(pm panelModel, grid lensDashboardAPIGrid, panelID *string) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := pm.LensDashboardAppConfig
	if cfg == nil {
		diags.AddError("Missing `lens_dashboard_app_config`", "The `lens_dashboard_app_config` block is required for `lens-dashboard-app` panels.")
		return kbapi.DashboardPanelItem{}, diags
	}
	switch {
	case cfg.ByValue != nil:
		return lensDashboardAppByValueToAPI(*cfg.ByValue, grid, panelID)
	case cfg.ByReference != nil:
		return lensDashboardAppByReferenceToAPI(*cfg.ByReference, grid, panelID)
	default:
		diags.AddError("Invalid `lens_dashboard_app_config`", "Exactly one of `by_value` or `by_reference` must be set inside `lens_dashboard_app_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}
}

func lensDashboardAppByValueToAPI(
	byValue lensDashboardAppByValueModel,
	grid lensDashboardAPIGrid,
	panelID *string,
) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	var config kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	if err := config.UnmarshalJSON([]byte(byValue.ConfigJSON.ValueString())); err != nil {
		diags.AddError("Invalid `by_value.config_json` for lens-dashboard-app", err.Error())
		return kbapi.DashboardPanelItem{}, diags
	}
	ldPanel := kbapi.KbnDashboardPanelTypeLensDashboardApp{
		Config: config,
		Grid: kbapi.KbnDashboardPanelGrid{
			H: grid.H,
			W: grid.W,
			X: grid.X,
			Y: grid.Y,
		},
		Id: panelID,
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeLensDashboardApp(ldPanel); err != nil {
		diags.AddError("Failed to create lens-dashboard-app panel", err.Error())
	}
	return panelItem, diags
}

func lensDashboardAppByReferenceToAPI(
	byRef lensDashboardAppByReferenceModel,
	grid lensDashboardAPIGrid,
	panelID *string,
) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	api1 := kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1{
		RefId: byRef.RefID.ValueString(),
		TimeRange: kbapi.KbnEsQueryServerTimeRangeSchema{
			From: byRef.TimeRange.From.ValueString(),
			To:   byRef.TimeRange.To.ValueString(),
		},
	}
	if typeutils.IsKnown(byRef.TimeRange.Mode) {
		m := kbapi.KbnEsQueryServerTimeRangeSchemaMode(byRef.TimeRange.Mode.ValueString())
		api1.TimeRange.Mode = &m
	}
	if typeutils.IsKnown(byRef.ReferencesJSON) {
		refs, d := jsonBytesFromOptionalNormalizedArray(byRef.ReferencesJSON, "by_reference.references_json")
		diags.Append(d...)
		if d.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		if len(refs) > 0 {
			var out []kbapi.KbnContentManagementUtilsReferenceSchema
			if err := json.Unmarshal(refs, &out); err != nil {
				diags.AddError("Invalid `by_reference.references_json` for lens-dashboard-app", err.Error())
				return kbapi.DashboardPanelItem{}, diags
			}
			api1.References = &out
		}
	}
	if typeutils.IsKnown(byRef.Title) {
		t := byRef.Title.ValueString()
		api1.Title = &t
	}
	if typeutils.IsKnown(byRef.Description) {
		d := byRef.Description.ValueString()
		api1.Description = &d
	}
	if typeutils.IsKnown(byRef.HideTitle) {
		v := byRef.HideTitle.ValueBool()
		api1.HideTitle = &v
	}
	if typeutils.IsKnown(byRef.HideBorder) {
		v := byRef.HideBorder.ValueBool()
		api1.HideBorder = &v
	}
	if typeutils.IsKnown(byRef.DrilldownsJSON) {
		b, d := jsonBytesFromOptionalNormalizedArray(byRef.DrilldownsJSON, "by_reference.drilldowns_json")
		diags.Append(d...)
		if d.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		if len(b) > 0 {
			var items []kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item
			if err := json.Unmarshal(b, &items); err != nil {
				diags.AddError("Invalid `by_reference.drilldowns_json` for lens-dashboard-app", err.Error())
				return kbapi.DashboardPanelItem{}, diags
			}
			if len(items) > 0 {
				api1.Drilldowns = &items
			}
		}
	}
	var config kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	if err := config.FromKbnDashboardPanelTypeLensDashboardAppConfig1(api1); err != nil {
		diags.AddError("Failed to set lens-dashboard-app by_reference config", err.Error())
		return kbapi.DashboardPanelItem{}, diags
	}
	ldPanel := kbapi.KbnDashboardPanelTypeLensDashboardApp{
		Config: config,
		Grid: kbapi.KbnDashboardPanelGrid{
			H: grid.H,
			W: grid.W,
			X: grid.X,
			Y: grid.Y,
		},
		Id: panelID,
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeLensDashboardApp(ldPanel); err != nil {
		diags.AddError("Failed to create lens-dashboard-app panel", err.Error())
	}
	return panelItem, diags
}

func lensOptionalJSONBytes(n jsontypes.Normalized) []byte {
	if !typeutils.IsKnown(n) {
		return nil
	}
	return []byte(n.ValueString())
}

// jsonBytesFromOptionalNormalizedArray rejects JSON null and returns bytes for the array/object payload.
func jsonBytesFromOptionalNormalizedArray(n jsontypes.Normalized, field string) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	b := lensOptionalJSONBytes(n)
	if len(b) == 0 {
		return b, diags
	}
	if s := string(b); s == "null" {
		diags.AddError("Invalid JSON for "+field, "JSON `null` is not valid; omit the argument or use a JSON array.")
		return nil, diags
	}
	return b, diags
}

// populateLensDashboardAppFromAPI maps an API lens-dashboard-app panel to the TF model in pm.
// pm is seeded from the prior plan/state before this runs.
func populateLensDashboardAppFromAPI(
	ctx context.Context,
	pm *panelModel,
	_ *panelModel,
	api kbapi.KbnDashboardPanelTypeLensDashboardApp,
) diag.Diagnostics {
	var diags diag.Diagnostics
	prior := pm.LensDashboardAppConfig

	configBytes, err := api.Config.MarshalJSON()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	cfg1, err1 := api.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig1()
	byRef := err1 == nil && cfg1.RefId != "" && cfg1.TimeRange.From != "" && cfg1.TimeRange.To != ""

	if byRef {
		tr := lensDashboardAppTimeRangeModel{
			From: types.StringValue(cfg1.TimeRange.From),
			To:   types.StringValue(cfg1.TimeRange.To),
		}
		switch {
		case cfg1.TimeRange.Mode != nil:
			tr.Mode = types.StringValue(string(*cfg1.TimeRange.Mode))
		case prior != nil && prior.ByReference != nil && typeutils.IsKnown(prior.ByReference.TimeRange.Mode):
			tr.Mode = prior.ByReference.TimeRange.Mode
		default:
			tr.Mode = types.StringNull()
		}
		by := lensDashboardAppByReferenceModel{
			RefID:     types.StringValue(cfg1.RefId),
			TimeRange: tr,
		}
		by.Title = lensOptionalStringFromAPI(cfg1.Title, prior, func(br *lensDashboardAppByReferenceModel) types.String { return br.Title })
		by.Description = lensOptionalStringFromAPI(cfg1.Description, prior, func(br *lensDashboardAppByReferenceModel) types.String { return br.Description })
		by.HideTitle = lensOptionalBoolFromAPI(cfg1.HideTitle, prior, func(br *lensDashboardAppByReferenceModel) types.Bool { return br.HideTitle })
		by.HideBorder = lensOptionalBoolFromAPI(cfg1.HideBorder, prior, func(br *lensDashboardAppByReferenceModel) types.Bool { return br.HideBorder })

		switch {
		case cfg1.References != nil:
			b, err := json.Marshal(cfg1.References)
			if err != nil {
				return diagutil.FrameworkDiagFromError(err)
			}
			if norm, ok := marshalToNormalized(b, err, "references_json", &diags); ok {
				if prior != nil && prior.ByReference != nil {
					norm = preservePriorNormalizedWithDefaultsIfEquivalent(ctx, prior.ByReference.ReferencesJSON, norm, defaultOpaqueRootJSON, &diags)
				}
				by.ReferencesJSON = norm
			}
		case prior != nil && prior.ByReference != nil && typeutils.IsKnown(prior.ByReference.ReferencesJSON):
			by.ReferencesJSON = prior.ByReference.ReferencesJSON
		default:
			by.ReferencesJSON = jsontypes.NewNormalizedNull()
		}

		switch {
		case cfg1.Drilldowns != nil:
			b, err := json.Marshal(*cfg1.Drilldowns)
			if err != nil {
				return diagutil.FrameworkDiagFromError(err)
			}
			if norm, ok := marshalToNormalized(b, err, "drilldowns_json", &diags); ok {
				if prior != nil && prior.ByReference != nil {
					norm = preservePriorNormalizedWithDefaultsIfEquivalent(ctx, prior.ByReference.DrilldownsJSON, norm, defaultOpaqueRootJSON, &diags)
				}
				by.DrilldownsJSON = norm
			}
		case prior != nil && prior.ByReference != nil && typeutils.IsKnown(prior.ByReference.DrilldownsJSON):
			by.DrilldownsJSON = prior.ByReference.DrilldownsJSON
		default:
			by.DrilldownsJSON = jsontypes.NewNormalizedNull()
		}

		pm.LensDashboardAppConfig = &lensDashboardAppConfigModel{
			ByReference: &by,
		}
		return diags
	}

	// by_value: full API config JSON, normalized, with prior preservation for semantic stability.
	if norm, ok := marshalToNormalized(configBytes, nil, "by_value.config_json", &diags); ok {
		if prior != nil && prior.ByValue != nil {
			norm = preservePriorNormalizedWithDefaultsIfEquivalent(ctx, prior.ByValue.ConfigJSON, norm, defaultOpaqueRootJSON, &diags)
		}
		pm.LensDashboardAppConfig = &lensDashboardAppConfigModel{
			ByValue: &lensDashboardAppByValueModel{ConfigJSON: norm},
		}
	}
	return diags
}

func lensOptionalStringFromAPI(
	api *string,
	prior *lensDashboardAppConfigModel,
	priorField func(*lensDashboardAppByReferenceModel) types.String,
) types.String {
	if api != nil {
		return types.StringValue(*api)
	}
	if prior != nil && prior.ByReference != nil {
		p := priorField(prior.ByReference)
		if typeutils.IsKnown(p) {
			return p
		}
	}
	return types.StringNull()
}

func lensOptionalBoolFromAPI(
	api *bool,
	prior *lensDashboardAppConfigModel,
	priorField func(*lensDashboardAppByReferenceModel) types.Bool,
) types.Bool {
	if api != nil {
		return types.BoolValue(*api)
	}
	if prior != nil && prior.ByReference != nil {
		p := priorField(prior.ByReference)
		if typeutils.IsKnown(p) {
			return p
		}
	}
	return types.BoolNull()
}
