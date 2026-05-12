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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// lensChartPresentationTFModel mirrors optional chart-root presentation fields on typed Lens configs.
type lensChartPresentationTFModel struct {
	TimeRange      *timeRangeModel            `tfsdk:"time_range"`
	HideTitle      types.Bool                 `tfsdk:"hide_title"`
	HideBorder     types.Bool                 `tfsdk:"hide_border"`
	ReferencesJSON jsontypes.Normalized       `tfsdk:"references_json"`
	Drilldowns     []lensDrilldownItemTFModel `tfsdk:"drilldowns"`
}

func newNullLensChartPresentationTFModel() lensChartPresentationTFModel {
	return lensChartPresentationTFModel{
		TimeRange:      nil,
		HideTitle:      types.BoolNull(),
		HideBorder:     types.BoolNull(),
		ReferencesJSON: jsontypes.NewNormalizedNull(),
		Drilldowns:     nil,
	}
}

// lensChartPresentationWrites holds normalized API write material for chart roots.
type lensChartPresentationWrites struct {
	TimeRange     kbapi.KbnEsQueryServerTimeRangeSchema
	HideTitle     *bool
	HideBorder    *bool
	References    *[]kbapi.KbnContentManagementUtilsReferenceSchema
	DrilldownsRaw [][]byte
}

func lensChartPresentationWritesFor(dashboard *dashboardModel, in lensChartPresentationTFModel) (lensChartPresentationWrites, diag.Diagnostics) {
	var writes lensChartPresentationWrites
	var diags diag.Diagnostics

	writes.TimeRange = resolveChartTimeRange(dashboard, in.TimeRange)
	if typeutils.IsKnown(in.HideTitle) {
		v := in.HideTitle.ValueBool()
		writes.HideTitle = &v
	}
	if typeutils.IsKnown(in.HideBorder) {
		v := in.HideBorder.ValueBool()
		writes.HideBorder = &v
	}

	refs, refDiags := lensChartPresentationReferencesWrites(in.ReferencesJSON, "references_json")
	diags.Append(refDiags...)
	if refDiags.HasError() {
		return writes, diags
	}
	writes.References = refs

	if len(in.Drilldowns) > 0 {
		raw, ddDiags := lensDrilldownsToRawJSON(in.Drilldowns)
		diags.Append(ddDiags...)
		if ddDiags.HasError() {
			return writes, diags
		}
		writes.DrilldownsRaw = raw
	}

	return writes, diags
}

func lensChartPresentationReferencesWrites(referencesJSON jsontypes.Normalized, fieldLabel string) (*[]kbapi.KbnContentManagementUtilsReferenceSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, d := jsonBytesFromOptionalNormalizedArray(referencesJSON, fieldLabel)
	diags.Append(d...)
	if d.HasError() || len(b) == 0 {
		return nil, diags
	}

	var refs []kbapi.KbnContentManagementUtilsReferenceSchema
	if err := json.Unmarshal(b, &refs); err != nil {
		diags.AddError("Invalid "+fieldLabel, err.Error())
		return nil, diags
	}
	return &refs, diags
}

// decodeLensDrilldownSlice unmarshals raw drilldown JSON produced by lensDrilldownsToRawJSON into generated union item types.
func decodeLensDrilldownSlice[Item any](raw [][]byte) ([]Item, diag.Diagnostics) {
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

// lensDrilldownItemTFModel is one drilldown entry; exactly one nested variant is set after validation.
type lensDrilldownItemTFModel struct {
	DashboardDrilldown *lensDashboardDrilldownTFModel `tfsdk:"dashboard_drilldown"`
	DiscoverDrilldown  *lensDiscoverDrilldownTFModel  `tfsdk:"discover_drilldown"`
	URLDrilldown       *lensURLDrilldownTFModel       `tfsdk:"url_drilldown"`
}

type lensDashboardDrilldownTFModel struct {
	DashboardID  types.String `tfsdk:"dashboard_id"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	UseFilters   types.Bool   `tfsdk:"use_filters"`
	UseTimeRange types.Bool   `tfsdk:"use_time_range"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type lensDiscoverDrilldownTFModel struct {
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type lensURLDrilldownTFModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

const lensDrilldownTriggerOnApplyFilter = "on_apply_filter"

func lensDrilldownsToRawJSON(items []lensDrilldownItemTFModel) ([][]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(items) == 0 {
		return nil, diags
	}

	out := make([][]byte, 0, len(items))
	for i, item := range items {
		b, d := lensDrilldownItemToRawJSON(item, i)
		diags.Append(d...)
		if d.HasError() {
			return nil, diags
		}
		out = append(out, b)
	}
	return out, diags
}

func lensDrilldownItemToRawJSON(item lensDrilldownItemTFModel, index int) ([]byte, diag.Diagnostics) {
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

		// Kibana only supports a single trigger value for dashboard drilldowns; keep the wire value fixed
		// even when the computed TF attribute is unknown during early plan evaluation.
		obj := map[string]any{
			"type":            "dashboard_drilldown",
			"trigger":         lensDrilldownTriggerOnApplyFilter,
			"dashboard_id":    dd.DashboardID.ValueString(),
			"label":           dd.Label.ValueString(),
			"use_filters":     optionalBoolForDrilldownJSON(dd.UseFilters, true),
			"use_time_range":  optionalBoolForDrilldownJSON(dd.UseTimeRange, true),
			"open_in_new_tab": optionalBoolForDrilldownJSON(dd.OpenInNewTab, false),
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

		// Mirror dashboard drilldown: API trigger is fixed; do not rely on computed TF state for writes.
		obj := map[string]any{
			"type":            "discover_drilldown",
			"trigger":         lensDrilldownTriggerOnApplyFilter,
			"label":           dd.Label.ValueString(),
			"open_in_new_tab": optionalBoolForDrilldownJSON(dd.OpenInNewTab, true),
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
			"type":            "url_drilldown",
			"url":             u.URL.ValueString(),
			"label":           u.Label.ValueString(),
			"trigger":         u.Trigger.ValueString(),
			"encode_url":      optionalBoolForDrilldownJSON(u.EncodeURL, true),
			"open_in_new_tab": optionalBoolForDrilldownJSON(u.OpenInNewTab, true),
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

func lensTimeRangeModeString(mode *kbapi.KbnEsQueryServerTimeRangeSchemaMode) string {
	if mode == nil {
		return ""
	}
	return string(*mode)
}

func lensTimeRangesAPILiteralEqual(a, b kbapi.KbnEsQueryServerTimeRangeSchema) bool {
	if a.From != b.From || a.To != b.To {
		return false
	}
	return lensTimeRangeModeString(a.Mode) == lensTimeRangeModeString(b.Mode)
}

func dashboardLensComparableTimeRange(dashboard *dashboardModel) (kbapi.KbnEsQueryServerTimeRangeSchema, bool) {
	if dashboard == nil || dashboard.TimeRange == nil {
		return kbapi.KbnEsQueryServerTimeRangeSchema{}, false
	}
	return timeRangeModelToAPI(dashboard.TimeRange), true
}

// chartTimeRangeFromAPI maps a chart-root API time range into Terraform state with REQ-038/REQ-009 null-preservation semantics.
func chartTimeRangeFromAPI(dashboard *dashboardModel, apiTimeRange kbapi.KbnEsQueryServerTimeRangeSchema, priorState *timeRangeModel) *timeRangeModel {
	// unmarshals can yield a zero-valued time_range when the wire JSON omits the object.
	// Treat that as "no chart-level time_range" so TF state preserves null while the write path still inherits dashboard time.
	if apiTimeRange.From == "" && apiTimeRange.To == "" && (apiTimeRange.Mode == nil || lensTimeRangeModeString(apiTimeRange.Mode) == "") {
		return nil
	}

	priorWasNil := priorState == nil

	dashTR, dashOK := dashboardLensComparableTimeRange(dashboard)
	if priorWasNil && dashOK && lensTimeRangesAPILiteralEqual(apiTimeRange, dashTR) {
		return nil
	}

	return timeRangeModelFromAPIWithModePreservation(apiTimeRange, priorState)
}

func timeRangeModelFromAPIWithModePreservation(api kbapi.KbnEsQueryServerTimeRangeSchema, prior *timeRangeModel) *timeRangeModel {
	out := &timeRangeModel{
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

func lensPresentationReferencesJSONRead(ctx context.Context, prior jsontypes.Normalized, refs *[]kbapi.KbnContentManagementUtilsReferenceSchema) (jsontypes.Normalized, diag.Diagnostics) {
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

	if norm, ok := marshalToNormalized(b, err, "references_json", &diags); ok {
		norm = preservePriorNormalizedWithDefaultsIfEquivalent(ctx, prior, norm, defaultOpaqueRootJSON, &diags)
		return norm, diags
	}

	return jsontypes.NewNormalizedNull(), diags
}

func lensDrilldownsAPIToWire[Item any](items *[]Item) (wire [][]byte, omitted bool, diags diag.Diagnostics) {
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
func drilldownsFromAPI(wire [][]byte) ([]lensDrilldownItemTFModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(wire) == 0 {
		return nil, diags
	}

	out := make([]lensDrilldownItemTFModel, 0, len(wire))
	for i, b := range wire {
		item, d := lensDrilldownItemFromAPIJSON(b, i)
		diags.Append(d...)
		if d.HasError() {
			return nil, diags
		}
		out = append(out, item)
	}

	return out, diags
}

func lensDrilldownItemFromAPIJSON(raw []byte, index int) (lensDrilldownItemTFModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	path := fmt.Sprintf("drilldowns[%d]", index)

	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		diags.AddError("Invalid "+path, err.Error())
		return lensDrilldownItemTFModel{}, diags
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
			diags.AddError("Invalid "+path+".dashboard_drilldown", err.Error())
			return lensDrilldownItemTFModel{}, diags
		}

		trigger := body.Trigger
		if trigger == "" {
			trigger = lensDrilldownTriggerOnApplyFilter
		}

		return lensDrilldownItemTFModel{
			DashboardDrilldown: &lensDashboardDrilldownTFModel{
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
			diags.AddError("Invalid "+path+".discover_drilldown", err.Error())
			return lensDrilldownItemTFModel{}, diags
		}

		trigger := body.Trigger
		if trigger == "" {
			trigger = lensDrilldownTriggerOnApplyFilter
		}

		return lensDrilldownItemTFModel{
			DiscoverDrilldown: &lensDiscoverDrilldownTFModel{
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
			diags.AddError("Invalid "+path+".url_drilldown", err.Error())
			return lensDrilldownItemTFModel{}, diags
		}

		return lensDrilldownItemTFModel{
			URLDrilldown: &lensURLDrilldownTFModel{
				URL:          types.StringValue(body.URL),
				Label:        types.StringValue(body.Label),
				Trigger:      types.StringValue(body.Trigger),
				EncodeURL:    types.BoolPointerValue(body.EncodeURL),
				OpenInNewTab: types.BoolPointerValue(body.OpenInNewTab),
			},
		}, diags

	default:
		diags.AddError("Invalid "+path, fmt.Sprintf("Unknown drilldown type %q", head.Type))
		return lensDrilldownItemTFModel{}, diags
	}
}

// lensChartPresentationReadsFor maps optional chart-root presentation API fields into Terraform state with REQ-009-style null preservation.
func lensChartPresentationReadsFor(
	ctx context.Context,
	dashboard *dashboardModel,
	prior *lensChartPresentationTFModel,
	apiTimeRange kbapi.KbnEsQueryServerTimeRangeSchema,
	hideTitle *bool,
	hideBorder *bool,
	refs *[]kbapi.KbnContentManagementUtilsReferenceSchema,
	drilldownWire [][]byte,
	drilldownsOmitted bool,
) (lensChartPresentationTFModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	var priorTime *timeRangeModel
	var priorRefs jsontypes.Normalized
	var priorHideTitle types.Bool
	var priorHideBorder types.Bool
	var priorDrills []lensDrilldownItemTFModel

	if prior != nil {
		priorTime = prior.TimeRange
		priorRefs = prior.ReferencesJSON
		priorHideTitle = prior.HideTitle
		priorHideBorder = prior.HideBorder
		priorDrills = prior.Drilldowns
	} else {
		priorHideTitle = types.BoolNull()
		priorHideBorder = types.BoolNull()
		priorRefs = jsontypes.NewNormalizedNull()
	}

	var out lensChartPresentationTFModel
	out.TimeRange = chartTimeRangeFromAPI(dashboard, apiTimeRange, priorTime)
	out.HideTitle = lensPresentationOptionalBoolRead(hideTitle, priorHideTitle)
	out.HideBorder = lensPresentationOptionalBoolRead(hideBorder, priorHideBorder)

	refNorm, refDiags := lensPresentationReferencesJSONRead(ctx, priorRefs, refs)
	diags.Append(refDiags...)
	if refDiags.HasError() {
		return lensChartPresentationTFModel{}, diags
	}
	out.ReferencesJSON = refNorm

	if !drilldownsOmitted {
		items, ddDiags := drilldownsFromAPI(drilldownWire)
		diags.Append(ddDiags...)
		if ddDiags.HasError() {
			return lensChartPresentationTFModel{}, diags
		}
		out.Drilldowns = items
	} else {
		out.Drilldowns = priorDrills
	}

	return out, diags
}
