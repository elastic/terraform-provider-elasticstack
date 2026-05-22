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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Drilldown KBAPI union variants for `vis` by-reference config (`KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_1_Drilldowns_Item`).

const diagnosticSummaryDrilldownConv = "Structured drilldowns"

// VisDrilldownsFromAPI decodes vis by-reference drilldown union items.
func VisDrilldownsFromAPI(ctx context.Context, api *[]kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_1_Drilldowns_Item) (models.DrilldownsModel, diag.Diagnostics) {
	_ = ctx
	if api == nil {
		return nil, nil
	}
	out := make(models.DrilldownsModel, 0, len(*api))
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

// VisDrilldownsToAPI encodes structured drilldowns into vis by-reference drilldown unions.
func VisDrilldownsToAPI(items models.DrilldownsModel) (*[]kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_1_Drilldowns_Item, diag.Diagnostics) {
	if items == nil {
		return nil, nil
	}
	api := make([]kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_1_Drilldowns_Item, 0, len(items))
	var diags diag.Diagnostics
	for _, m := range items {
		u, itemDiags := drilldownModelToVisUnionItem(m)
		diags.Append(itemDiags...)
		if itemDiags.HasError() {
			return nil, diags
		}
		api = append(api, u)
	}
	return &api, diags
}

// DrilldownsFromVisByRefAPI translates API drilldowns on `vis.by_reference`.
func DrilldownsFromVisByRefAPI(ctx context.Context, api *[]kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_1_Drilldowns_Item) (models.DrilldownsModel, diag.Diagnostics) {
	_ = ctx
	if api == nil {
		return nil, nil
	}
	out := make(models.DrilldownsModel, 0, len(*api))
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

// DrilldownsToVisByRefAPI translates structured drilldowns for `vis.by_reference`.
func DrilldownsToVisByRefAPI(items models.DrilldownsModel) (*[]kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_1_Drilldowns_Item, diag.Diagnostics) {
	return VisDrilldownsToAPI(items)
}

type drilldownTypeJSONPeek struct {
	Type string `json:"type"`
}

func drilldownItemFromVisUnionRaw(item kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_1_Drilldowns_Item) (models.DrilldownItemModel, diag.Diagnostics) {
	raw, err := json.Marshal(item)
	if err != nil {
		return models.DrilldownItemModel{}, diag.Diagnostics{diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error())}
	}
	return decodePeekedVisDrilldownJSON(raw)
}

func decodePeekedVisDrilldownJSON(raw []byte) (models.DrilldownItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var peek drilldownTypeJSONPeek
	if err := json.Unmarshal(raw, &peek); err != nil {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
		return models.DrilldownItemModel{}, diags
	}
	switch peek.Type {
	case string(kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns0TypeDashboardDrilldown):
		return decodeDashboardBranchVis(raw)
	case string(kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns1TypeDiscoverDrilldown):
		return decodeDiscoverBranchVis(raw)
	case string(kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns2TypeUrlDrilldown):
		return decodeURLBranchVis(raw)
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
	return models.DrilldownItemModel{}, diags
}

func decodeDashboardBranchVis(raw []byte) (models.DrilldownItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var obj kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns0
	if err := json.Unmarshal(raw, &obj); err != nil {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
		return models.DrilldownItemModel{}, diags
	}
	wantType := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns0TypeDashboardDrilldown
	if obj.Type != wantType {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("Expected drilldown type %q but API returned %#q.", wantType, obj.Type)))
		return models.DrilldownItemModel{}, diags
	}
	wantTr := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns0TriggerOnApplyFilter
	if obj.Trigger != wantTr {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("Dashboard drilldown API `trigger` must be %#q for lossless import; API returned %#q.", wantTr, obj.Trigger)))
		return models.DrilldownItemModel{}, diags
	}
	dm := &models.DrilldownDashboardBlockModel{
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
	return models.DrilldownItemModel{Dashboard: dm}, diags
}

func decodeDiscoverBranchVis(raw []byte) (models.DrilldownItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var obj kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns1
	if err := json.Unmarshal(raw, &obj); err != nil {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
		return models.DrilldownItemModel{}, diags
	}
	wantType := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns1TypeDiscoverDrilldown
	if obj.Type != wantType {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("Expected drilldown type %q but API returned %#q.", wantType, obj.Type)))
		return models.DrilldownItemModel{}, diags
	}
	wantTr := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns1TriggerOnApplyFilter
	if obj.Trigger != wantTr {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("Discover drilldown API `trigger` must be %#q for lossless import; API returned %#q.", wantTr, obj.Trigger)))
		return models.DrilldownItemModel{}, diags
	}
	discover := &models.DrilldownDiscoverBlockModel{
		Label: types.StringValue(obj.Label),
	}
	if obj.OpenInNewTab != nil {
		discover.OpenInNewTab = types.BoolValue(*obj.OpenInNewTab)
	} else {
		discover.OpenInNewTab = types.BoolNull()
	}
	return models.DrilldownItemModel{Discover: discover}, diags
}

func decodeURLBranchVis(raw []byte) (models.DrilldownItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var obj kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns2
	if err := json.Unmarshal(raw, &obj); err != nil {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
		return models.DrilldownItemModel{}, diags
	}
	wantType := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns2TypeUrlDrilldown
	if obj.Type != wantType {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("Expected drilldown type %q but API returned %#q.", wantType, obj.Type)))
		return models.DrilldownItemModel{}, diags
	}
	triggerStr := string(obj.Trigger)
	if triggerStr == "" {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			"URL drilldown from the API omits required field `trigger`; structured drilldowns cannot represent this losslessly."))
		return models.DrilldownItemModel{}, diags
	}
	if !kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns2Trigger(triggerStr).Valid() {
		diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv,
			fmt.Sprintf("URL drilldown has unsupported API `trigger` %#q.", triggerStr)))
		return models.DrilldownItemModel{}, diags
	}
	url := &models.DrilldownURLBlockModel{
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
	return models.DrilldownItemModel{URL: url}, diags
}

func drilldownModelToVisUnionItem(m models.DrilldownItemModel) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_1_Drilldowns_Item, diag.Diagnostics) {
	var u kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_1_Drilldowns_Item
	var diags diag.Diagnostics
	switch {
	case m.Dashboard != nil:
		dd := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns0{
			DashboardId: m.Dashboard.DashboardID.ValueString(),
			Label:       m.Dashboard.Label.ValueString(),
			Trigger:     kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns0TriggerOnApplyFilter,
			Type:        kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns0TypeDashboardDrilldown,
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
		if err := u.FromKibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns0(dd); err != nil {
			diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
			return u, diags
		}
	case m.Discover != nil:
		dd := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns1{
			Label:   m.Discover.Label.ValueString(),
			Trigger: kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns1TriggerOnApplyFilter,
			Type:    kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns1TypeDiscoverDrilldown,
		}
		if typeutils.IsKnown(m.Discover.OpenInNewTab) {
			v := m.Discover.OpenInNewTab.ValueBool()
			dd.OpenInNewTab = &v
		}
		if err := u.FromKibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns1(dd); err != nil {
			diags.Append(diag.NewErrorDiagnostic(diagnosticSummaryDrilldownConv, err.Error()))
			return u, diags
		}
	case m.URL != nil:
		wire := map[string]any{
			"type":  string(kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns2TypeUrlDrilldown),
			"url":   m.URL.URL.ValueString(),
			"label": m.URL.Label.ValueString(),
		}
		if typeutils.IsKnown(m.URL.Trigger) {
			trigger := m.URL.Trigger.ValueString()
			if !kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1Drilldowns2Trigger(trigger).Valid() {
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

// ExplicitEmptyDrilldowns builds a slice that is intentionally non-nil with length 0 for distinguishing omission from explicit empty-clear in Terraform state.
func ExplicitEmptyDrilldowns() models.DrilldownsModel {
	return make(models.DrilldownsModel, 0)
}
