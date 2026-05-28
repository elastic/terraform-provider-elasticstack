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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ByReferenceOptionalStringFromAPI returns the API value when present, falls back to the known prior TF value, otherwise null.
func ByReferenceOptionalStringFromAPI(
	api *string,
	prior *models.VisByReferenceModel,
	priorField func(*models.VisByReferenceModel) types.String,
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
	prior *models.VisByReferenceModel,
	priorField func(*models.VisByReferenceModel) types.Bool,
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

// VisByReferenceModelToAPIConfig1 maps Terraform `vis_config.by_reference` attributes to the generated Config1 shape.
func VisByReferenceModelToAPIConfig1(
	byRef models.VisByReferenceModel,
	referencesJSONFieldLabel string,
) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1, diag.Diagnostics) {
	var diags diag.Diagnostics
	api1 := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1{
		RefId: byRef.RefID.ValueString(),
		TimeRange: &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
			From: byRef.TimeRange.From.ValueString(),
			To:   byRef.TimeRange.To.ValueString(),
		},
	}
	if typeutils.IsKnown(byRef.TimeRange.Mode) {
		m := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaMode(byRef.TimeRange.Mode.ValueString())
		api1.TimeRange.Mode = &m
	}
	if typeutils.IsKnown(byRef.ReferencesJSON) {
		refs, d := JSONBytesFromOptionalNormalizedArray(byRef.ReferencesJSON, referencesJSONFieldLabel)
		diags.Append(d...)
		if d.HasError() {
			return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1{}, diags
		}
		if len(refs) > 0 {
			var out []kbapi.KibanaHTTPAPIsKbnContentManagementUtilsReferenceSchema
			if err := json.Unmarshal(refs, &out); err != nil {
				diags.AddError("Invalid `"+referencesJSONFieldLabel+"`", err.Error())
				return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1{}, diags
			}
			if out == nil {
				out = []kbapi.KibanaHTTPAPIsKbnContentManagementUtilsReferenceSchema{}
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
		dd, ddDiags := VisDrilldownsToAPI(byRef.Drilldowns)
		diags.Append(ddDiags...)
		if ddDiags.HasError() {
			return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1{}, diags
		}
		api1.Drilldowns = dd
	}
	return api1, diags
}

// HasLensByReferenceShapeAtRoot reports whether m has the by-reference shape: a non-empty ref_id
// and a time_range with non-empty from/to strings.
func HasLensByReferenceShapeAtRoot(m map[string]any) bool {
	if m == nil {
		return false
	}
	ref, ok := m["ref_id"]
	if !ok {
		return false
	}
	refS, ok := ref.(string)
	if !ok || refS == "" {
		return false
	}
	trAny, ok := m["time_range"]
	if !ok {
		return false
	}
	tr, ok := trAny.(map[string]any)
	if !ok {
		return false
	}
	from, fOK := tr["from"].(string)
	to, tOK := tr["to"].(string)
	return fOK && tOK && from != "" && to != ""
}

// LensByReferenceAttributes returns the shared Terraform schema attribute map for a by-reference
// lens panel config block (ref_id, references_json, title, description, hide_title, hide_border,
// drilldowns, time_range). Used by `vis_config.by_reference`.
func LensByReferenceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"ref_id": schema.StringAttribute{
			MarkdownDescription: "Reference name in the API `ref_id` field. When `references_json` is set, `ref_id` typically should match a `name` in that list so the link resolves as expected.",
			Required:            true,
		},
		"references_json": schema.StringAttribute{
			MarkdownDescription: "Optional normalized JSON array of `{ id, name, type }` saved-object references, matching the API `references` list (for example wiring a `lens` saved object to `ref_id`).",
			Optional:            true,
			CustomType:          jsontypes.NormalizedType{},
		},
		"title": schema.StringAttribute{
			MarkdownDescription: "Optional panel title.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Optional panel description.",
			Optional:            true,
		},
		"hide_title": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the panel title.",
			Optional:            true,
		},
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the panel border.",
			Optional:            true,
		},
		"drilldowns": panelkit.StructuredDrilldownsAttribute(),
		"time_range": schema.SingleNestedAttribute{
			MarkdownDescription: "Required time range for the by-reference panel config (`vis_config.by_reference`).",
			Required:            true,
			Attributes:          panelkit.TimeRangeAttributes(),
		},
	}
}

// PopulateVisByReferenceTFModelFromAPIConfig1 maps API Config1 fields into the Terraform `by_reference` model with REQ-009 preservation semantics.
func PopulateVisByReferenceTFModelFromAPIConfig1(
	ctx context.Context,
	cfg1 kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1,
	prior *models.VisByReferenceModel,
) (models.VisByReferenceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	tr := models.VisByReferenceTimeRangeModel{
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
	by := models.VisByReferenceModel{
		RefID:     types.StringValue(cfg1.RefId),
		TimeRange: tr,
	}

	by.Title = ByReferenceOptionalStringFromAPI(cfg1.Title, prior, func(br *models.VisByReferenceModel) types.String { return br.Title })
	by.Description = ByReferenceOptionalStringFromAPI(cfg1.Description, prior, func(br *models.VisByReferenceModel) types.String { return br.Description })
	by.HideTitle = ByReferenceOptionalBoolFromAPI(cfg1.HideTitle, prior, func(br *models.VisByReferenceModel) types.Bool { return br.HideTitle })
	by.HideBorder = ByReferenceOptionalBoolFromAPI(cfg1.HideBorder, prior, func(br *models.VisByReferenceModel) types.Bool { return br.HideBorder })

	switch {
	case cfg1.References != nil:
		b, err := json.Marshal(cfg1.References)
		if err != nil {
			diags.AddError("Failed to marshal references_json", err.Error())
			return models.VisByReferenceModel{}, diags
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
		items, drillDiags := VisDrilldownsFromAPI(ctx, cfg1.Drilldowns)
		diags.Append(drillDiags...)
		if drillDiags.HasError() {
			return models.VisByReferenceModel{}, diags
		}
		by.Drilldowns = items
	case prior != nil && prior.Drilldowns != nil:
		by.Drilldowns = prior.Drilldowns
	default:
		by.Drilldowns = nil
	}

	return by, diags
}
