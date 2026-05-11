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

// lensChartPresentationInput carries chart-root presentation TF fields for shared write-path helpers.
type lensChartPresentationInput struct {
	TimeRange      *timeRangeModel
	HideTitle      types.Bool
	HideBorder     types.Bool
	ReferencesJSON jsontypes.Normalized
	Drilldowns     []lensDrilldownItemTFModel
}

// lensChartPresentationWrites holds normalized API write material for chart roots.
type lensChartPresentationWrites struct {
	TimeRange     kbapi.KbnEsQueryServerTimeRangeSchema
	HideTitle     *bool
	HideBorder    *bool
	References    *[]kbapi.KbnContentManagementUtilsReferenceSchema
	DrilldownsRaw [][]byte
}

func lensChartPresentationWritesFor(dashboard *dashboardModel, in lensChartPresentationInput) (lensChartPresentationWrites, diag.Diagnostics) {
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
