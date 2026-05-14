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

func configPriorForVisRead(tfPanel, pm *panelModel) *visConfigModel {
	if tfPanel != nil && tfPanel.VisConfig != nil {
		return tfPanel.VisConfig
	}
	if pm != nil && pm.VisConfig != nil {
		return pm.VisConfig
	}
	return nil
}

// populateVisByReferenceFromAPI maps API vis config branch 1 (by-reference saved object panel).
func populateVisByReferenceFromAPI(
	ctx context.Context,
	prior *visConfigModel,
	pm *panelModel,
	cfg1 kbapi.KbnDashboardPanelTypeVisConfig1,
) diag.Diagnostics {
	var diags diag.Diagnostics
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
	var priorBR *lensDashboardAppByReferenceModel
	if prior != nil {
		priorBR = prior.ByReference
	}
	by.Title = byReferenceOptionalStringFromAPI(cfg1.Title, priorBR, func(br *lensDashboardAppByReferenceModel) types.String { return br.Title })
	by.Description = byReferenceOptionalStringFromAPI(cfg1.Description, priorBR, func(br *lensDashboardAppByReferenceModel) types.String { return br.Description })
	by.HideTitle = byReferenceOptionalBoolFromAPI(cfg1.HideTitle, priorBR, func(br *lensDashboardAppByReferenceModel) types.Bool { return br.HideTitle })
	by.HideBorder = byReferenceOptionalBoolFromAPI(cfg1.HideBorder, priorBR, func(br *lensDashboardAppByReferenceModel) types.Bool { return br.HideBorder })

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
		items, drillDiags := drilldownsFromVisByRefAPI(ctx, cfg1.Drilldowns)
		diags.Append(drillDiags...)
		if drillDiags.HasError() {
			return diags
		}
		by.Drilldowns = items
	case prior != nil && prior.ByReference != nil && prior.ByReference.Drilldowns != nil:
		by.Drilldowns = prior.ByReference.Drilldowns
	default:
		by.Drilldowns = nil
	}

	brCopy := by
	pm.VisConfig = &visConfigModel{
		ByReference: &brCopy,
	}
	return diags
}

func visByReferenceToAPI(
	byRef lensDashboardAppByReferenceModel,
	grid struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	},
	panelID *string,
) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	api1 := kbapi.KbnDashboardPanelTypeVisConfig1{
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
		refs, d := jsonBytesFromOptionalNormalizedArray(byRef.ReferencesJSON, "vis_config.by_reference.references_json")
		diags.Append(d...)
		if d.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		if len(refs) > 0 {
			var out []kbapi.KbnContentManagementUtilsReferenceSchema
			if err := json.Unmarshal(refs, &out); err != nil {
				diags.AddError("Invalid `vis_config.by_reference.references_json`", err.Error())
				return kbapi.DashboardPanelItem{}, diags
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
		dd, ddDiags := drilldownsToVisByRefAPI(byRef.Drilldowns)
		diags.Append(ddDiags...)
		if ddDiags.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		api1.Drilldowns = dd
	}
	var config kbapi.KbnDashboardPanelTypeVis_Config
	if err := config.FromKbnDashboardPanelTypeVisConfig1(api1); err != nil {
		diags.AddError("Failed to set vis by_reference config", err.Error())
		return kbapi.DashboardPanelItem{}, diags
	}
	visPanel := kbapi.KbnDashboardPanelTypeVis{
		Config: config,
		Grid:   grid,
		Id:     panelID,
		Type:   kbapi.Vis,
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeVis(visPanel); err != nil {
		diags.AddError("Failed to create visualization panel", err.Error())
	}
	return panelItem, diags
}

func visConfigToAPI(pm panelModel, dashboard *dashboardModel, grid struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
}, panelID *string) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := pm.VisConfig
	if cfg == nil {
		diags.AddError("Missing `vis_config`", "The `vis_config` block is required for typed `vis` panels.")
		return kbapi.DashboardPanelItem{}, diags
	}
	switch {
	case cfg.ByReference != nil:
		return visByReferenceToAPI(*cfg.ByReference, grid, panelID)
	case cfg.ByValue != nil:
		blocks := &cfg.ByValue.lensByValueChartBlocks
		conv, okConv := firstLensVisConverterForChartBlocks(blocks)
		if !okConv {
			diags.AddError("Invalid `vis_config.by_value`", "The typed chart block could not be resolved to a Lens visualization converter.")
			return kbapi.DashboardPanelItem{}, diags
		}
		config0, d := conv.buildAttributes(blocks, dashboard)
		diags.Append(d...)
		if d.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		var config kbapi.KbnDashboardPanelTypeVis_Config
		if err := config.FromKbnDashboardPanelTypeVisConfig0(config0); err != nil {
			diags.AddError("Failed to create visualization panel config", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		visPanel := kbapi.KbnDashboardPanelTypeVis{
			Config: config,
			Grid:   grid,
			Id:     panelID,
			Type:   kbapi.Vis,
		}
		var panelItem kbapi.DashboardPanelItem
		if err := panelItem.FromKbnDashboardPanelTypeVis(visPanel); err != nil {
			diags.AddError("Failed to create visualization panel", err.Error())
		}
		return panelItem, diags
	default:
		diags.AddError("Invalid `vis_config`", "Exactly one of `by_value` or `by_reference` must be set inside `vis_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}
}
