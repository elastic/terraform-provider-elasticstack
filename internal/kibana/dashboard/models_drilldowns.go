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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Drilldown KBAPI union variants for `lens-dashboard-app` / `vis` by-reference config:
//
//	KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item (canonical encoder path)
//	KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item (decoded via the same discriminator fields)
//
// Concrete union members share wire shape with:
//
//	KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns0 (dashboard_drilldown)
//	KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns1 (discover_drilldown)
//	KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns2 (url_drilldown)
//
// (Vis equivalents: KbnDashboardPanelTypeVisConfig1Drilldowns0/1/2.)

// drilldownsModel is Terraform state for REQ-039 structured drilldown lists.
type drilldownsModel []drilldownItemModel

type drilldownItemModel struct {
	Dashboard *drilldownDashboardBlockModel `tfsdk:"dashboard"`
	Discover  *drilldownDiscoverBlockModel  `tfsdk:"discover"`
	URL       *drilldownURLBlockModel       `tfsdk:"url"`
}

type drilldownDashboardBlockModel struct {
	DashboardID  types.String `tfsdk:"dashboard_id"`
	Label        types.String `tfsdk:"label"`
	UseFilters   types.Bool   `tfsdk:"use_filters"`
	UseTimeRange types.Bool   `tfsdk:"use_time_range"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type drilldownDiscoverBlockModel struct {
	Label        types.String `tfsdk:"label"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type drilldownURLBlockModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

// fromAPI decodes Lens by-reference drilldown union items (`KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item`).
func fromAPI(ctx context.Context, api *[]kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item) (drilldownsModel, diag.Diagnostics) {
	_ = ctx
	if api == nil {
		return nil, nil
	}
	out := make(drilldownsModel, 0, len(*api))
	var diags diag.Diagnostics
	for _, item := range *api {
		m, itemDiags := drilldownItemFromLensUnionRaw(item)
		diags.Append(itemDiags...)
		if !itemDiags.HasError() {
			out = append(out, m)
		}
	}
	if diags.HasError() {
		return nil, diags
	}
	return out, diags
}

// toAPI encodes structured drilldowns into Lens by-reference drilldown unions.
func toAPI(items drilldownsModel) (*[]kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item, diag.Diagnostics) {
	if items == nil {
		return nil, nil
	}
	api := make([]kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item, 0, len(items))
	var diags diag.Diagnostics
	for _, m := range items {
		u, itemDiags := drilldownModelToLensUnionItem(m)
		diags.Append(itemDiags...)
		if itemDiags.HasError() {
			return nil, diags
		}
		api = append(api, u)
	}
	return &api, diags
}

// drilldownsFromVisByRefAPI translates API drilldowns on `vis.by_reference`.
func drilldownsFromVisByRefAPI(ctx context.Context, api *[]kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item) (drilldownsModel, diag.Diagnostics) {
	_ = ctx
	if api == nil {
		return nil, nil
	}
	out := make(drilldownsModel, 0, len(*api))
	var diags diag.Diagnostics
	for _, item := range *api {
		m, itemDiags := drilldownItemFromVisUnionRaw(item)
		diags.Append(itemDiags...)
		if !itemDiags.HasError() {
			out = append(out, m)
		}
	}
	if diags.HasError() {
		return nil, diags
	}
	return out, diags
}

// drilldownsToVisByRefAPI translates structured drilldowns for `vis.by_reference`.
func drilldownsToVisByRefAPI(items drilldownsModel) (*[]kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item, diag.Diagnostics) {
	lensSlice, diags := toAPI(items)
	if diags.HasError() || lensSlice == nil {
		return nil, diags
	}
	// Vis and Lens drilldown unions share identical JSON wire shapes in kbapi codegen.
	// Round-trip each item through JSON to convert union types without duplicating field mapping.
	// If code generation ever diverges between the two, UnmarshalJSON will fail with a clear error.
	visSlice := make([]kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item, 0, len(*lensSlice))
	for _, li := range *lensSlice {
		payload, err := json.Marshal(li)
		if err != nil {
			diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
			return nil, diags
		}
		var vi kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item
		if err := vi.UnmarshalJSON(payload); err != nil {
			diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
			return nil, diags
		}
		visSlice = append(visSlice, vi)
	}
	return &visSlice, diags
}

type drilldownTypeJSONPeek struct {
	Type string `json:"type"`
}

const diagnosticSummaryDrilldownConv = "Structured drilldowns"

func drilldownItemFromLensUnionRaw(item kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item) (drilldownItemModel, diag.Diagnostics) {
	raw, err := json.Marshal(item)
	if err != nil {
		return drilldownItemModel{}, diag.Diagnostics{diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error())}
	}
	return decodePeekedLensDrilldownJSON(raw)
}

func drilldownItemFromVisUnionRaw(item kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item) (drilldownItemModel, diag.Diagnostics) {
	raw, err := json.Marshal(item)
	if err != nil {
		return drilldownItemModel{}, diag.Diagnostics{diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error())}
	}
	return decodePeekedLensDrilldownJSON(raw)
}

// Shared JSON discriminator branch for Lens and Vis by-reference unions (matching wire shape).
func decodePeekedLensDrilldownJSON(raw []byte) (drilldownItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var peek drilldownTypeJSONPeek
	if err := json.Unmarshal(raw, &peek); err != nil {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
		return drilldownItemModel{}, diags
	}
	switch peek.Type {
	case string(kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns0TypeDashboardDrilldown):
		return decodeDashboardBranchLens(raw)
	case string(kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns1TypeDiscoverDrilldown):
		return decodeDiscoverBranchLens(raw)
	case string(kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns2TypeUrlDrilldown):
		return decodeURLBranchLens(raw)
	case "":
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			"API drilldown is missing required discriminator field `type` "+
				"(expected dashboard_drilldown, discover_drilldown, or url_drilldown)."))
	default:
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf(
				"Unsupported API drilldown `type` %#q; cannot represent it using structured terraform drilldown blocks.",
				peek.Type)))
	}
	return drilldownItemModel{}, diags
}

func decodeDashboardBranchLens(raw []byte) (drilldownItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var obj kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns0
	if err := json.Unmarshal(raw, &obj); err != nil {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
		return drilldownItemModel{}, diags
	}
	wantType := kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns0TypeDashboardDrilldown
	if obj.Type != wantType {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("Expected drilldown type %q but API returned %#q.", wantType, obj.Type)))
		return drilldownItemModel{}, diags
	}
	wantTr := kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns0TriggerOnApplyFilter
	if obj.Trigger != wantTr {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("Dashboard drilldown API `trigger` must be %#q for lossless import; API returned %#q.", wantTr, obj.Trigger)))
		return drilldownItemModel{}, diags
	}
	dm := &drilldownDashboardBlockModel{
		DashboardID: types.StringValue(obj.DashboardId),
		Label:       types.StringValue(obj.Label),
	}
	if obj.UseFilters != nil {
		dm.UseFilters = types.BoolValue(*obj.UseFilters)
	} else {
		dm.UseFilters = types.BoolNull()
	}
	if obj.UseTimeRange != nil {
		dm.UseTimeRange = types.BoolValue(*obj.UseTimeRange)
	} else {
		dm.UseTimeRange = types.BoolNull()
	}
	if obj.OpenInNewTab != nil {
		dm.OpenInNewTab = types.BoolValue(*obj.OpenInNewTab)
	} else {
		dm.OpenInNewTab = types.BoolNull()
	}
	return drilldownItemModel{Dashboard: dm}, diags
}

func decodeDiscoverBranchLens(raw []byte) (drilldownItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var obj kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns1
	if err := json.Unmarshal(raw, &obj); err != nil {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
		return drilldownItemModel{}, diags
	}
	wantType := kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns1TypeDiscoverDrilldown
	if obj.Type != wantType {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("Expected drilldown type %q but API returned %#q.", wantType, obj.Type)))
		return drilldownItemModel{}, diags
	}
	wantTr := kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns1TriggerOnApplyFilter
	if obj.Trigger != wantTr {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("Discover drilldown API `trigger` must be %#q for lossless import; API returned %#q.", wantTr, obj.Trigger)))
		return drilldownItemModel{}, diags
	}
	discover := &drilldownDiscoverBlockModel{
		Label: types.StringValue(obj.Label),
	}
	if obj.OpenInNewTab != nil {
		discover.OpenInNewTab = types.BoolValue(*obj.OpenInNewTab)
	} else {
		discover.OpenInNewTab = types.BoolNull()
	}
	return drilldownItemModel{Discover: discover}, diags
}

func decodeURLBranchLens(raw []byte) (drilldownItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var obj kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns2
	if err := json.Unmarshal(raw, &obj); err != nil {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
		return drilldownItemModel{}, diags
	}
	wantType := kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns2TypeUrlDrilldown
	if obj.Type != wantType {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("Expected drilldown type %q but API returned %#q.", wantType, obj.Type)))
		return drilldownItemModel{}, diags
	}
	triggerStr := string(obj.Trigger)
	if triggerStr == "" {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			"URL drilldown from the API omits required field `trigger`; structured drilldowns cannot represent this losslessly."))
		return drilldownItemModel{}, diags
	}
	if !kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns2Trigger(triggerStr).Valid() {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("URL drilldown has unsupported API `trigger` %#q.", triggerStr)))
		return drilldownItemModel{}, diags
	}
	url := &drilldownURLBlockModel{
		URL:     types.StringValue(obj.Url),
		Label:   types.StringValue(obj.Label),
		Trigger: types.StringValue(triggerStr),
	}
	if obj.EncodeUrl != nil {
		url.EncodeURL = types.BoolValue(*obj.EncodeUrl)
	} else {
		url.EncodeURL = types.BoolNull()
	}
	if obj.OpenInNewTab != nil {
		url.OpenInNewTab = types.BoolValue(*obj.OpenInNewTab)
	} else {
		url.OpenInNewTab = types.BoolNull()
	}
	return drilldownItemModel{URL: url}, diags
}

func drilldownModelToLensUnionItem(m drilldownItemModel) (kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item, diag.Diagnostics) {
	var u kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item
	var diags diag.Diagnostics
	switch {
	case m.Dashboard != nil:
		dd := kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns0{
			DashboardId: m.Dashboard.DashboardID.ValueString(),
			Label:       m.Dashboard.Label.ValueString(),
			Trigger:     kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns0TriggerOnApplyFilter,
			Type:        kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns0TypeDashboardDrilldown,
		}
		if typeutils.IsKnown(m.Dashboard.UseFilters) {
			v := m.Dashboard.UseFilters.ValueBool()
			dd.UseFilters = &v
		}
		if typeutils.IsKnown(m.Dashboard.UseTimeRange) {
			v := m.Dashboard.UseTimeRange.ValueBool()
			dd.UseTimeRange = &v
		}
		if typeutils.IsKnown(m.Dashboard.OpenInNewTab) {
			v := m.Dashboard.OpenInNewTab.ValueBool()
			dd.OpenInNewTab = &v
		}
		if err := u.FromKbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns0(dd); err != nil {
			diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
			return u, diags
		}
	case m.Discover != nil:
		dd := kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns1{
			Label:   m.Discover.Label.ValueString(),
			Trigger: kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns1TriggerOnApplyFilter,
			Type:    kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns1TypeDiscoverDrilldown,
		}
		if typeutils.IsKnown(m.Discover.OpenInNewTab) {
			v := m.Discover.OpenInNewTab.ValueBool()
			dd.OpenInNewTab = &v
		}
		if err := u.FromKbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns1(dd); err != nil {
			diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
			return u, diags
		}
	case m.URL != nil:
		wire := map[string]any{
			"type":  string(kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns2TypeUrlDrilldown),
			"url":   m.URL.URL.ValueString(),
			"label": m.URL.Label.ValueString(),
		}
		// During plan refinement `trigger` may be unknown; omit it from the wire map until known.
		// Terraform schema marks `trigger` required, so finalized applies always emit it.
		if typeutils.IsKnown(m.URL.Trigger) {
			trigger := m.URL.Trigger.ValueString()
			if !kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1Drilldowns2Trigger(trigger).Valid() {
				diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
					fmt.Sprintf("Unsupported URL drilldown `trigger` %#q.", trigger)))
				return u, diags
			}
			wire["trigger"] = trigger
		}
		if typeutils.IsKnown(m.URL.EncodeURL) {
			v := m.URL.EncodeURL.ValueBool()
			wire["encode_url"] = v
		}
		if typeutils.IsKnown(m.URL.OpenInNewTab) {
			v := m.URL.OpenInNewTab.ValueBool()
			wire["open_in_new_tab"] = v
		}
		b, err := json.Marshal(wire)
		if err != nil {
			diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
			return u, diags
		}
		if err := u.UnmarshalJSON(b); err != nil {
			diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
			return u, diags
		}
	default:
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			"Internal error: drilldown model has no branch set."))
	}
	return u, diags
}

// explicitEmptyDrilldowns builds a slice that is intentionally non-nil with length 0. Terraform Plugin Framework decodes:
// - a null/absent optional `drilldowns` attribute into a nil Go slice (`reflect.Zero` on slice type — see terraform-plugin-framework `internal/reflect/into.go`);
// - practitioner `drilldowns = []` into reflect.MakeSlice(..., 0, 0), i.e. a non-nil empty slice (`internal/reflect/slice.go`).
// lensDashboardAppByReferenceToAPI therefore uses `byRef.Drilldowns != nil` to distinguish omission from explicit empty-clear.
func explicitEmptyDrilldowns() drilldownsModel {
	return make(drilldownsModel, 0)
}
