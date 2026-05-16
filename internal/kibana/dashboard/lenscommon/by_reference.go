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

package lenscommon

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ByReferenceOptionalStringFromAPI returns the API value when present, falls back to the known prior TF value, otherwise null.
func ByReferenceOptionalStringFromAPI(
	api *string,
	prior *models.LensDashboardAppByReferenceModel,
	priorField func(*models.LensDashboardAppByReferenceModel) types.String,
) types.String {
	if api != nil {
		return types.StringValue(*api)
	}
	if prior != nil {
		if p := priorField(prior); typeutils.IsKnown(p) {
			return p
		}
	}
	return types.StringNull()
}

// ByReferenceOptionalBoolFromAPI returns the API value when present, falls back to the known prior TF value, otherwise null.
func ByReferenceOptionalBoolFromAPI(
	api *bool,
	prior *models.LensDashboardAppByReferenceModel,
	priorField func(*models.LensDashboardAppByReferenceModel) types.Bool,
) types.Bool {
	if api != nil {
		return types.BoolValue(*api)
	}
	if prior != nil {
		if p := priorField(prior); typeutils.IsKnown(p) {
			return p
		}
	}
	return types.BoolNull()
}

// LensDashboardAppByReferenceModelToAPIConfig1 maps Terraform lens-dashboard-app `by_reference` attributes to the generated Config1 shape.
func LensDashboardAppByReferenceModelToAPIConfig1(
	byRef models.LensDashboardAppByReferenceModel,
	referencesJSONFieldLabel string,
) (kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1, diag.Diagnostics) {
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
		refs, d := JSONBytesFromOptionalNormalizedArray(byRef.ReferencesJSON, referencesJSONFieldLabel)
		diags.Append(d...)
		if d.HasError() {
			return kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1{}, diags
		}
		if len(refs) > 0 {
			var out []kbapi.KbnContentManagementUtilsReferenceSchema
			if err := json.Unmarshal(refs, &out); err != nil {
				diags.AddError("Invalid `"+referencesJSONFieldLabel+"`", err.Error())
				return kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1{}, diags
			}
			if out == nil {
				out = []kbapi.KbnContentManagementUtilsReferenceSchema{}
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
	if byRef.Drilldowns != nil {
		dd, ddDiags := LensDashboardAppDrilldownsToAPI(byRef.Drilldowns)
		diags.Append(ddDiags...)
		if ddDiags.HasError() {
			return kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1{}, diags
		}
		api1.Drilldowns = dd
	}
	return api1, diags
}

// VisByReferenceConfig1FromLens converts Lens-dashboard-app Config1 to Vis Config1 via JSON (wire-compatible shapes).
func VisByReferenceConfig1FromLens(in kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1) (kbapi.KbnDashboardPanelTypeVisConfig1, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, err := json.Marshal(in)
	if err != nil {
		diags.AddError("Failed to marshal lens-dashboard-app by_reference config", err.Error())
		return kbapi.KbnDashboardPanelTypeVisConfig1{}, diags
	}
	var out kbapi.KbnDashboardPanelTypeVisConfig1
	if err := json.Unmarshal(b, &out); err != nil {
		diags.AddError("Failed to convert lens-dashboard-app by_reference config to vis config", err.Error())
		return kbapi.KbnDashboardPanelTypeVisConfig1{}, diags
	}
	return out, diags
}

// PopulateLensByReferenceTFModelFromLensAppConfig1 maps API Config1 fields into the Terraform `by_reference` model with REQ-009 preservation semantics.
func PopulateLensByReferenceTFModelFromLensAppConfig1(
	ctx context.Context,
	cfg1 kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1,
	prior *models.LensDashboardAppByReferenceModel,
) (models.LensDashboardAppByReferenceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	tr := models.LensDashboardAppTimeRangeModel{
		From: types.StringValue(cfg1.TimeRange.From),
		To:   types.StringValue(cfg1.TimeRange.To),
	}
	switch {
	case cfg1.TimeRange.Mode != nil:
		tr.Mode = types.StringValue(string(*cfg1.TimeRange.Mode))
	case prior != nil && typeutils.IsKnown(prior.TimeRange.Mode):
		tr.Mode = prior.TimeRange.Mode
	default:
		tr.Mode = types.StringNull()
	}
	by := models.LensDashboardAppByReferenceModel{
		RefID:     types.StringValue(cfg1.RefId),
		TimeRange: tr,
	}

	by.Title = ByReferenceOptionalStringFromAPI(cfg1.Title, prior, func(br *models.LensDashboardAppByReferenceModel) types.String { return br.Title })
	by.Description = ByReferenceOptionalStringFromAPI(cfg1.Description, prior, func(br *models.LensDashboardAppByReferenceModel) types.String { return br.Description })
	by.HideTitle = ByReferenceOptionalBoolFromAPI(cfg1.HideTitle, prior, func(br *models.LensDashboardAppByReferenceModel) types.Bool { return br.HideTitle })
	by.HideBorder = ByReferenceOptionalBoolFromAPI(cfg1.HideBorder, prior, func(br *models.LensDashboardAppByReferenceModel) types.Bool { return br.HideBorder })

	switch {
	case cfg1.References != nil:
		b, err := json.Marshal(cfg1.References)
		if err != nil {
			diags.AddError("Failed to marshal references_json", err.Error())
			return models.LensDashboardAppByReferenceModel{}, diags
		}
		if norm, ok := MarshalToNormalized(b, err, "references_json", &diags); ok {
			if prior != nil {
				norm = panelkit.PreservePriorNormalizedWithDefaultsIfEquivalent(ctx, prior.ReferencesJSON, norm, defaultOpaqueRootJSON, &diags)
			}
			by.ReferencesJSON = norm
		}
	case prior != nil && typeutils.IsKnown(prior.ReferencesJSON):
		by.ReferencesJSON = prior.ReferencesJSON
	default:
		by.ReferencesJSON = jsontypes.NewNormalizedNull()
	}

	switch {
	case cfg1.Drilldowns != nil:
		items, drillDiags := LensDashboardAppDrilldownsFromAPI(ctx, cfg1.Drilldowns)
		diags.Append(drillDiags...)
		if drillDiags.HasError() {
			return models.LensDashboardAppByReferenceModel{}, diags
		}
		by.Drilldowns = items
	case prior != nil && prior.Drilldowns != nil:
		by.Drilldowns = prior.Drilldowns
	default:
		by.Drilldowns = nil
	}

	return by, diags
}
