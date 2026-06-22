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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LensChartPresentationWrites holds normalized API write material for typed Lens chart roots.
type LensChartPresentationWrites struct {
	TimeRange     *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema
	HideTitle     *bool
	HideBorder    *bool
	References    *[]CMReferenceSchema
	DrilldownsRaw [][]byte
}

// LensChartPresentationWritesFor builds API presentation fields from Terraform chart-root attributes.
func LensChartPresentationWritesFor(in models.LensChartPresentationTFModel) (LensChartPresentationWrites, diag.Diagnostics) {
	var writes LensChartPresentationWrites
	var diags diag.Diagnostics

	writes.TimeRange = ResolveChartTimeRange(in.TimeRange)
	if typeutils.IsKnown(in.HideTitle) {
		v := in.HideTitle.ValueBool()
		writes.HideTitle = &v
	}
	if typeutils.IsKnown(in.HideBorder) {
		v := in.HideBorder.ValueBool()
		writes.HideBorder = &v
	}

	refs, refDiags := LensChartPresentationReferencesWrites(in.ReferencesJSON, "references_json")
	diags.Append(refDiags...)
	if refDiags.HasError() {
		return writes, diags
	}
	writes.References = refs

	if len(in.Drilldowns) > 0 {
		raw, ddDiags := LensDrilldownsToRawJSON(in.Drilldowns)
		diags.Append(ddDiags...)
		if ddDiags.HasError() {
			return writes, diags
		}
		writes.DrilldownsRaw = raw
	}

	return writes, diags
}

// LensChartPresentationReferencesWrites unmarshals references_json into API reference objects when present.
func LensChartPresentationReferencesWrites(referencesJSON jsontypes.Normalized, fieldLabel string) (*[]CMReferenceSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, d := JSONBytesFromOptionalNormalizedArray(referencesJSON, fieldLabel)
	diags.Append(d...)
	if d.HasError() || len(b) == 0 {
		return nil, diags
	}

	var refs []kbapi.KibanaHTTPAPIsKbnContentManagementUtilsReferenceSchema
	if err := json.Unmarshal(b, &refs); err != nil {
		diags.AddError("Invalid "+fieldLabel, err.Error())
		return nil, diags
	}
	return &refs, diags
}

// ApplyLensChartPresentationWrites assigns presentation fields from writes to the API struct pointer fields.
// Pass pointers to the target struct fields: &api.TimeRange, &api.HideTitle, etc.
// The DrilldownItem type parameter must match the Drilldowns_Item type of the target API struct.
func ApplyLensChartPresentationWrites[DrilldownItem any](
	writes LensChartPresentationWrites,
	timeRange **kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema,
	hideTitle **bool,
	hideBorder **bool,
	references **[]CMReferenceSchema,
	drilldowns **[]DrilldownItem,
) diag.Diagnostics {
	var diags diag.Diagnostics
	*timeRange = writes.TimeRange
	if writes.HideTitle != nil {
		*hideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		*hideBorder = writes.HideBorder
	}
	if writes.References != nil {
		*references = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := DecodeLensDrilldownSlice[DrilldownItem](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			*drilldowns = &items
		}
	}
	return diags
}

// DecodeLensDrilldownSlice unmarshals raw drilldown JSON produced by LensDrilldownsToRawJSON into generated union item types.
func DecodeLensDrilldownSlice[Item any](raw [][]byte) ([]Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(raw) == 0 {
		return nil, diags
	}
	out := make([]Item, len(raw))
	for i, b := range raw {
		if err := json.Unmarshal(b, &out[i]); err != nil {
			diags.AddError("Invalid drilldowns", fmt.Sprintf("drilldowns[%d]: %v", i, err))
			return nil, diags
		}
	}
	return out, diags
}

// LensDrilldownTriggerOnApplyFilter is the fixed trigger wire value for dashboard/discover drilldown variants on writes.
const LensDrilldownTriggerOnApplyFilter = "on_apply_filter"

// LensDrilldownsToRawJSON encodes Terraform drilldown list items to JSON payloads for API unions.
func LensDrilldownsToRawJSON(items []models.LensDrilldownItemTFModel) ([][]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(items) == 0 {
		return nil, diags
	}

	out := make([][]byte, 0, len(items))
	for i, item := range items {
		b, d := LensDrilldownItemToRawJSON(item, i)
		diags.Append(d...)
		if d.HasError() {
			return nil, diags
		}
		out = append(out, b)
	}
	return out, diags
}

// LensDrilldownItemToRawJSON encodes one Terraform drilldown list item to JSON for API unions.
func LensDrilldownItemToRawJSON(item models.LensDrilldownItemTFModel, index int) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	path := fmt.Sprintf("drilldowns[%d]", index)

	var variants int
	if item.DashboardDrilldown != nil {
		variants++
	}
	if item.DiscoverDrilldown != nil {
		variants++
	}
	if item.URLDrilldown != nil {
		variants++
	}
	if variants != 1 {
		diags.AddError("Invalid "+path, "Expected exactly one drilldown variant; this should have been caught by schema validation.")
		return nil, diags
	}

	switch {
	case item.DashboardDrilldown != nil:
		dd := item.DashboardDrilldown
		if !typeutils.IsKnown(dd.DashboardID) || !typeutils.IsKnown(dd.Label) {
			diags.AddError("Invalid "+path+".dashboard_drilldown", "`dashboard_id` and `label` are required.")
			return nil, diags
		}

		obj := map[string]any{
			attrType:         drilldownTypeDashboard,
			attrTrigger:      LensDrilldownTriggerOnApplyFilter,
			"dashboard_id":   dd.DashboardID.ValueString(),
			attrLabel:        dd.Label.ValueString(),
			"use_filters":    optionalBoolForDrilldownJSON(dd.UseFilters, true),
			"use_time_range": optionalBoolForDrilldownJSON(dd.UseTimeRange, true),
			attrOpenInNewTab: optionalBoolForDrilldownJSON(dd.OpenInNewTab, false),
		}
		b, err := json.Marshal(obj)
		if err != nil {
			diags.AddError("Invalid "+path+".dashboard_drilldown", err.Error())
			return nil, diags
		}
		return b, diags

	case item.DiscoverDrilldown != nil:
		dd := item.DiscoverDrilldown
		if !typeutils.IsKnown(dd.Label) {
			diags.AddError("Invalid "+path+".discover_drilldown", "`label` is required.")
			return nil, diags
		}

		obj := map[string]any{
			attrType:         drilldownTypeDiscover,
			attrTrigger:      LensDrilldownTriggerOnApplyFilter,
			attrLabel:        dd.Label.ValueString(),
			attrOpenInNewTab: optionalBoolForDrilldownJSON(dd.OpenInNewTab, true),
		}
		b, err := json.Marshal(obj)
		if err != nil {
			diags.AddError("Invalid "+path+".discover_drilldown", err.Error())
			return nil, diags
		}
		return b, diags

	default: // URL variant
		u := item.URLDrilldown
		if !typeutils.IsKnown(u.URL) || !typeutils.IsKnown(u.Label) || !typeutils.IsKnown(u.Trigger) {
			diags.AddError("Invalid "+path+".url_drilldown", "`url`, `label`, and `trigger` are required.")
			return nil, diags
		}
		obj := map[string]any{
			attrType:         drilldownTypeURL,
			attrURL:          u.URL.ValueString(),
			attrLabel:        u.Label.ValueString(),
			attrTrigger:      u.Trigger.ValueString(),
			"encode_url":     optionalBoolForDrilldownJSON(u.EncodeURL, true),
			attrOpenInNewTab: optionalBoolForDrilldownJSON(u.OpenInNewTab, true),
		}
		b, err := json.Marshal(obj)
		if err != nil {
			diags.AddError("Invalid "+path+".url_drilldown", err.Error())
			return nil, diags
		}
		return b, diags
	}
}

func optionalBoolForDrilldownJSON(b types.Bool, defaultIfUnknown bool) bool {
	if !typeutils.IsKnown(b) {
		return defaultIfUnknown
	}
	return b.ValueBool()
}

func lensTimeRangeModeString(mode *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaMode) string {
	if mode == nil {
		return ""
	}
	return string(*mode)
}

// chartTimeRangeFromAPI maps a chart-root API time range into Terraform state. It returns nil when the API omits the time range (REQ-040) and preserves null `mode` when the API omits mode (REQ-009).
func chartTimeRangeFromAPI(apiTimeRange *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema, priorState *models.TimeRangeModel) *models.TimeRangeModel {
	if apiTimeRange == nil {
		return nil
	}
	if apiTimeRange.From == "" && apiTimeRange.To == "" && (apiTimeRange.Mode == nil || lensTimeRangeModeString(apiTimeRange.Mode) == "") {
		return nil
	}

	return timeRangeModelFromAPIWithModePreservation(*apiTimeRange, priorState)
}

func timeRangeModelFromAPIWithModePreservation(api kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema, prior *models.TimeRangeModel) *models.TimeRangeModel {
	out := &models.TimeRangeModel{
		From: types.StringValue(api.From),
		To:   types.StringValue(api.To),
	}

	hasAPIMode := api.Mode != nil && lensTimeRangeModeString(api.Mode) != ""
	switch {
	case hasAPIMode:
		out.Mode = types.StringValue(lensTimeRangeModeString(api.Mode))
	case prior != nil && prior.Mode.IsNull():
		out.Mode = types.StringNull()
	case prior != nil && typeutils.IsKnown(prior.Mode):
		out.Mode = prior.Mode
	default:
		out.Mode = types.StringNull()
	}

	return out
}

func lensPresentationOptionalBoolRead(api *bool, prior types.Bool) types.Bool {
	if api != nil {
		return types.BoolValue(*api)
	}
	if prior.IsNull() {
		return types.BoolNull()
	}
	return prior
}

func lensPresentationReferencesJSONRead(ctx context.Context, prior jsontypes.Normalized, refs *[]CMReferenceSchema) (jsontypes.Normalized, diag.Diagnostics) {
	var diags diag.Diagnostics

	refsOmitted := refs == nil || len(*refs) == 0
	if refsOmitted {
		if prior.IsNull() {
			return jsontypes.NewNormalizedNull(), diags
		}

		if typeutils.IsKnown(prior) {
			return prior, diags
		}

		return jsontypes.NewNormalizedNull(), diags
	}

	b, err := json.Marshal(refs)
	if err != nil {
		diags.AddError("Failed to marshal references_json", err.Error())
		return jsontypes.NewNormalizedNull(), diags
	}

	if norm, ok := WrapNormalizedJSON(b, err, "references_json", &diags); ok {
		norm = panelkit.PreservePriorNormalizedWithDefaultsIfEquivalent(ctx, prior, norm, defaultOpaqueRootJSON, &diags)
		return norm, diags
	}

	return jsontypes.NewNormalizedNull(), diags
}

// LensDrilldownsAPIToWire re-marshals API drilldown union slices to raw JSON for reads.
func LensDrilldownsAPIToWire[Item any](items *[]Item) (wire [][]byte, omitted bool, diags diag.Diagnostics) {
	if items == nil {
		return nil, true, diags
	}

	out := make([][]byte, 0, len(*items))
	for i, it := range *items {
		b, err := json.Marshal(it)
		if err != nil {
			diags.AddError("Invalid drilldowns", fmt.Sprintf("drilldowns[%d]: %v", i, err))
			return nil, false, diags
		}
		out = append(out, b)
	}

	return out, false, diags
}

// drilldownsFromAPI decodes API drilldown payloads (JSON-encoded union items) into Terraform list items.
func drilldownsFromAPI(wire [][]byte) ([]models.LensDrilldownItemTFModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(wire) == 0 {
		return nil, diags
	}

	out := make([]models.LensDrilldownItemTFModel, 0, len(wire))
	for i, b := range wire {
		item, d := LensDrilldownItemFromAPIJSON(b, fmt.Sprintf("drilldowns[%d]", i))
		diags.Append(d...)
		if d.HasError() {
			return nil, diags
		}
		out = append(out, item)
	}

	return out, diags
}

// LensDrilldownItemFromAPIJSON decodes one drilldown union JSON blob into a Terraform list item.
// pathPrefix labels diagnostics (e.g. drilldownsFromAPI passes `fmt.Sprintf("drilldowns[%d]", i)`).
func LensDrilldownItemFromAPIJSON(raw []byte, pathPrefix string) (models.LensDrilldownItemTFModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		diags.AddError("Invalid "+pathPrefix, err.Error())
		return models.LensDrilldownItemTFModel{}, diags
	}

	switch head.Type {
	case "dashboard_drilldown":
		var body struct {
			DashboardID string `json:"dashboard_id"`
			Label       string `json:"label"`
			Trigger     string `json:"trigger"`

			UseFilters   *bool `json:"use_filters"`
			UseTimeRange *bool `json:"use_time_range"`
			OpenInNewTab *bool `json:"open_in_new_tab"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			diags.AddError("Invalid "+pathPrefix+".dashboard_drilldown", err.Error())
			return models.LensDrilldownItemTFModel{}, diags
		}

		trigger := typeutils.NonZero(body.Trigger, LensDrilldownTriggerOnApplyFilter)

		return models.LensDrilldownItemTFModel{
			DashboardDrilldown: &models.LensDashboardDrilldownTFModel{
				DashboardID:  types.StringValue(body.DashboardID),
				Label:        types.StringValue(body.Label),
				Trigger:      types.StringValue(trigger),
				UseFilters:   types.BoolPointerValue(body.UseFilters),
				UseTimeRange: types.BoolPointerValue(body.UseTimeRange),
				OpenInNewTab: types.BoolPointerValue(body.OpenInNewTab),
			},
		}, diags

	case "discover_drilldown":
		var body struct {
			Label        string `json:"label"`
			Trigger      string `json:"trigger"`
			OpenInNewTab *bool  `json:"open_in_new_tab"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			diags.AddError("Invalid "+pathPrefix+".discover_drilldown", err.Error())
			return models.LensDrilldownItemTFModel{}, diags
		}

		trigger := typeutils.NonZero(body.Trigger, LensDrilldownTriggerOnApplyFilter)

		return models.LensDrilldownItemTFModel{
			DiscoverDrilldown: &models.LensDiscoverDrilldownTFModel{
				Label:        types.StringValue(body.Label),
				Trigger:      types.StringValue(trigger),
				OpenInNewTab: types.BoolPointerValue(body.OpenInNewTab),
			},
		}, diags

	case "url_drilldown":
		var body struct {
			URL          string `json:"url"`
			Label        string `json:"label"`
			Trigger      string `json:"trigger"`
			EncodeURL    *bool  `json:"encode_url"`
			OpenInNewTab *bool  `json:"open_in_new_tab"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			diags.AddError("Invalid "+pathPrefix+".url_drilldown", err.Error())
			return models.LensDrilldownItemTFModel{}, diags
		}

		return models.LensDrilldownItemTFModel{
			URLDrilldown: &models.LensURLDrilldownTFModel{
				URL:          types.StringValue(body.URL),
				Label:        types.StringValue(body.Label),
				Trigger:      types.StringValue(body.Trigger),
				EncodeURL:    types.BoolPointerValue(body.EncodeURL),
				OpenInNewTab: types.BoolPointerValue(body.OpenInNewTab),
			},
		}, diags

	default:
		diags.AddError("Invalid "+pathPrefix, fmt.Sprintf("Unknown drilldown type %q", head.Type))
		return models.LensDrilldownItemTFModel{}, diags
	}
}

// PopulateLensChartPresentation consolidates the repeated drilldown-wire and presentation-read
// block found across all Lens panel FromAPI functions. It writes the result into *out.
// Returns false if any error occurs; the caller should immediately return diags.
func PopulateLensChartPresentation[Item any, Prior lensChartPresentationProvider](
	ctx context.Context,
	out *models.LensChartPresentationTFModel,
	prior Prior,
	apiTimeRange *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema,
	hideTitle, hideBorder *bool,
	refs *[]CMReferenceSchema,
	drilldowns *[]Item,
	diags *diag.Diagnostics,
) bool {
	ddWire, ddOmit, ddWireDiags := LensDrilldownsAPIToWire(drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return false
	}
	pres, presDiags := LensChartPresentationReadsFor(ctx, prior, apiTimeRange, hideTitle, hideBorder, refs, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return false
	}
	*out = pres
	return true
}

// LensChartPresentationReadsFor maps optional chart-root presentation API fields into Terraform state with REQ-009-style null preservation.
func LensChartPresentationReadsFor[Prior lensChartPresentationProvider](
	ctx context.Context,
	prior Prior,
	apiTimeRange *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema,
	hideTitle *bool,
	hideBorder *bool,
	refs *[]CMReferenceSchema,
	drilldownWire [][]byte,
	drilldownsOmitted bool,
) (models.LensChartPresentationTFModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	var priorPres *models.LensChartPresentationTFModel
	// Comparing against the zero value detects a nil typed pointer (e.g. *TreemapConfigModel)
	// because assigning it to an interface would not yield a nil interface.
	var zeroPrior Prior
	if prior != zeroPrior {
		priorPres = prior.GetLensChartPresentation()
	}

	var priorTime *models.TimeRangeModel
	var priorRefs jsontypes.Normalized
	var priorHideTitle types.Bool
	var priorHideBorder types.Bool
	var priorDrills []models.LensDrilldownItemTFModel

	if priorPres != nil {
		priorTime = priorPres.TimeRange
		priorRefs = priorPres.ReferencesJSON
		priorHideTitle = priorPres.HideTitle
		priorHideBorder = priorPres.HideBorder
		priorDrills = priorPres.Drilldowns
	} else {
		priorHideTitle = types.BoolNull()
		priorHideBorder = types.BoolNull()
		priorRefs = jsontypes.NewNormalizedNull()
	}

	var out models.LensChartPresentationTFModel
	out.TimeRange = chartTimeRangeFromAPI(apiTimeRange, priorTime)
	out.HideTitle = lensPresentationOptionalBoolRead(hideTitle, priorHideTitle)
	out.HideBorder = lensPresentationOptionalBoolRead(hideBorder, priorHideBorder)

	refNorm, refDiags := lensPresentationReferencesJSONRead(ctx, priorRefs, refs)
	diags.Append(refDiags...)
	if refDiags.HasError() {
		return models.LensChartPresentationTFModel{}, diags
	}
	out.ReferencesJSON = refNorm

	if !drilldownsOmitted {
		items, ddDiags := drilldownsFromAPI(drilldownWire)
		diags.Append(ddDiags...)
		if ddDiags.HasError() {
			return models.LensChartPresentationTFModel{}, diags
		}
		out.Drilldowns = items
	} else {
		out.Drilldowns = priorDrills
	}

	return out, diags
}

func defaultOpaqueRootJSON(v any) any { return v }

// JSONBytesFromOptionalNormalizedArray rejects JSON null and returns bytes for the array/object payload.
func JSONBytesFromOptionalNormalizedArray(n jsontypes.Normalized, field string) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	b := optionalLensNormalizedJSONBytes(n)
	if len(b) == 0 {
		return b, diags
	}
	if s := string(b); s == "null" {
		diags.AddError("Invalid JSON for "+field, "JSON `null` is not valid; omit the argument or use a JSON array.")
		return nil, diags
	}
	return b, diags
}

func optionalLensNormalizedJSONBytes(n jsontypes.Normalized) []byte {
	if !typeutils.IsKnown(n) {
		return nil
	}
	return []byte(n.ValueString())
}
