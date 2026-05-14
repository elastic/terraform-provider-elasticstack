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

package slooverview

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Exactly one of Single or Groups must be set.
//
// trigger and type are always hardcoded to "on_open_panel_menu" / "url_drilldown" — they
// are not exposed to users (matching the slo_burn_rate_config drilldowns approach).

type sloDrilldownWireJSON struct {
	EncodeURL    *bool  `json:"encode_url,omitempty"`
	Label        string `json:"label"`
	OpenInNewTab *bool  `json:"open_in_new_tab,omitempty"`
	Trigger      string `json:"trigger"`
	Type         string `json:"type"`
	URL          string `json:"url"`
}

// BuildConfig writes Terraform panel state into the typed API panel's config union (Grid/Id are set separately).
func BuildConfig(pm models.PanelModel, panel *kbapi.KbnDashboardPanelTypeSloOverview) diag.Diagnostics {
	var diags diag.Diagnostics
	cfg := pm.SloOverviewConfig
	if cfg == nil {
		return nil
	}

	var config kbapi.KbnDashboardPanelTypeSloOverview_Config

	if cfg.Single != nil {
		single, d := singleToAPI(cfg.Single)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		b, err := json.Marshal(single)
		if err != nil {
			diags.AddError("Failed to marshal SLO single overview config", err.Error())
			return diags
		}
		if err := config.UnmarshalJSON(b); err != nil {
			diags.AddError("Failed to set SLO single overview config", err.Error())
			return diags
		}
	} else if cfg.Groups != nil {
		groups, d := groupsToAPI(cfg.Groups)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		b, err := json.Marshal(groups)
		if err != nil {
			diags.AddError("Failed to marshal SLO groups overview config", err.Error())
			return diags
		}
		if err := config.UnmarshalJSON(b); err != nil {
			diags.AddError("Failed to set SLO groups overview config", err.Error())
			return diags
		}
	}

	panel.Config = config
	return diags
}

func singleToAPI(m *models.SloOverviewSingleModel) (kbapi.SloSingleOverviewEmbeddable, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.SloSingleOverviewEmbeddable{
		OverviewMode: kbapi.SloSingleOverviewEmbeddableOverviewModeSingle,
		SloId:        m.SloID.ValueString(),
	}

	if typeutils.IsKnown(m.SloInstanceID) {
		api.SloInstanceId = m.SloInstanceID.ValueStringPointer()
	}
	if typeutils.IsKnown(m.RemoteName) {
		api.RemoteName = m.RemoteName.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Title) {
		api.Title = m.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = m.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(m.HideTitle) {
		api.HideTitle = m.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.HideBorder) {
		api.HideBorder = m.HideBorder.ValueBoolPointer()
	}

	if len(m.Drilldowns) > 0 {
		d := setDrilldownsOnSingle(&api, m.Drilldowns)
		diags.Append(d...)
	}

	return api, diags
}

func setDrilldownsOnSingle(api *kbapi.SloSingleOverviewEmbeddable, drilldowns []models.URLDrilldownModel) diag.Diagnostics {
	return injectDrilldownsJSON(api, drilldowns)
}

func groupsToAPI(m *models.SloOverviewGroupsModel) (kbapi.SloGroupOverviewEmbeddable, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.SloGroupOverviewEmbeddable{
		OverviewMode: kbapi.Groups,
	}

	if typeutils.IsKnown(m.Title) {
		api.Title = m.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = m.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(m.HideTitle) {
		api.HideTitle = m.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.HideBorder) {
		api.HideBorder = m.HideBorder.ValueBoolPointer()
	}

	if len(m.Drilldowns) > 0 {
		d := setDrilldownsOnGroups(&api, m.Drilldowns)
		diags.Append(d...)
	}

	if m.GroupFilters != nil {
		gf, d := groupFiltersToAPI(m.GroupFilters)
		diags.Append(d...)
		if !diags.HasError() {
			api.GroupFilters = gf
		}
	}

	return api, diags
}

func setDrilldownsOnGroups(api *kbapi.SloGroupOverviewEmbeddable, drilldowns []models.URLDrilldownModel) diag.Diagnostics {
	return injectDrilldownsJSON(api, drilldowns)
}

func injectDrilldownsJSON(api any, drilldowns []models.URLDrilldownModel) diag.Diagnostics {
	ddsJSON, err := json.Marshal(buildDrilldownsWire(drilldowns))
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to marshal drilldowns", err.Error())}
	}
	base, err := json.Marshal(api)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to marshal SLO config", err.Error())}
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to unmarshal SLO config", err.Error())}
	}
	m["drilldowns"] = ddsJSON
	merged, err := json.Marshal(m)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to re-marshal SLO config", err.Error())}
	}
	if err := json.Unmarshal(merged, api); err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to apply drilldowns to SLO config", err.Error())}
	}
	return nil
}

func buildDrilldownsWire(drilldowns []models.URLDrilldownModel) []sloDrilldownWireJSON {
	result := make([]sloDrilldownWireJSON, len(drilldowns))
	for i, dd := range drilldowns {
		result[i] = sloDrilldownWireJSON{
			URL:     dd.URL.ValueString(),
			Label:   dd.Label.ValueString(),
			Trigger: "on_open_panel_menu",
			Type:    "url_drilldown",
		}
		if typeutils.IsKnown(dd.EncodeURL) {
			result[i].EncodeURL = dd.EncodeURL.ValueBoolPointer()
		}
		if typeutils.IsKnown(dd.OpenInNewTab) {
			result[i].OpenInNewTab = dd.OpenInNewTab.ValueBoolPointer()
		}
	}
	return result
}

func groupFiltersToAPI(m *models.SloGroupFiltersModel) (*struct {
	Filters  *[]kbapi.SloGroupOverviewEmbeddable_GroupFilters_Filters_Item `json:"filters,omitempty"`
	GroupBy  *kbapi.SloGroupOverviewEmbeddableGroupFiltersGroupBy          `json:"group_by,omitempty"`
	Groups   *[]string                                                     `json:"groups,omitempty"`
	KqlQuery *string                                                       `json:"kql_query,omitempty"`
}, diag.Diagnostics) {
	var diags diag.Diagnostics
	gf := &struct {
		Filters  *[]kbapi.SloGroupOverviewEmbeddable_GroupFilters_Filters_Item `json:"filters,omitempty"`
		GroupBy  *kbapi.SloGroupOverviewEmbeddableGroupFiltersGroupBy          `json:"group_by,omitempty"`
		Groups   *[]string                                                     `json:"groups,omitempty"`
		KqlQuery *string                                                       `json:"kql_query,omitempty"`
	}{}

	if typeutils.IsKnown(m.GroupBy) {
		gb := kbapi.SloGroupOverviewEmbeddableGroupFiltersGroupBy(m.GroupBy.ValueString())
		gf.GroupBy = &gb
	}

	if len(m.Groups) > 0 {
		groups := make([]string, len(m.Groups))
		for i, g := range m.Groups {
			groups[i] = g.ValueString()
		}
		gf.Groups = &groups
	}

	if typeutils.IsKnown(m.KQLQuery) {
		gf.KqlQuery = m.KQLQuery.ValueStringPointer()
	}

	if typeutils.IsKnown(m.FiltersJSON) && !m.FiltersJSON.IsNull() {
		var filters []kbapi.SloGroupOverviewEmbeddable_GroupFilters_Filters_Item
		if err := json.Unmarshal([]byte(m.FiltersJSON.ValueString()), &filters); err != nil {
			diags.AddError("Failed to unmarshal filters_json", err.Error())
			return nil, diags
		}
		gf.Filters = &filters
	}

	return gf, diags
}

// PopulateFromAPI maps an SLO overview API panel into Terraform panel state. prior is TF plan/state (nil on import).
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, panel kbapi.KbnDashboardPanelTypeSloOverview) diag.Diagnostics {
	var diags diag.Diagnostics

	discriminator, err := panel.Config.Discriminator()
	if err != nil {
		diags.AddError("Failed to determine SLO overview mode", err.Error())
		return diags
	}

	switch discriminator {
	case "slo-single-overview-embeddable", "single":
		single, err := panel.Config.AsSloSingleOverviewEmbeddable()
		if err != nil {
			diags.AddError("Failed to read SLO single overview config", err.Error())
			return diags
		}
		return sloSingleFromAPI(pm, prior, single)
	case "slo-group-overview-embeddable", "groups":
		groups, err := panel.Config.AsSloGroupOverviewEmbeddable()
		if err != nil {
			diags.AddError("Failed to read SLO groups overview config", err.Error())
			return diags
		}
		return sloGroupsFromAPI(pm, prior, groups)
	default:
		diags.AddError("Unknown SLO overview mode", "Expected 'slo-single-overview-embeddable' or 'slo-group-overview-embeddable', got: "+discriminator)
		return diags
	}
}

func sloStringFromAPIOrPrior(apiVal *string, priorVal types.String) types.String {
	if apiVal != nil {
		return types.StringPointerValue(apiVal)
	}
	if typeutils.IsKnown(priorVal) {
		return priorVal
	}
	return types.StringNull()
}

func sloBoolFromAPIOrPrior(apiVal *bool, priorVal types.Bool) types.Bool {
	if apiVal != nil {
		return types.BoolPointerValue(apiVal)
	}
	if typeutils.IsKnown(priorVal) {
		return priorVal
	}
	return types.BoolNull()
}

func sloSingleFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, api kbapi.SloSingleOverviewEmbeddable) diag.Diagnostics {
	var diags diag.Diagnostics

	var priorSingle *models.SloOverviewSingleModel
	if tfPanel != nil && tfPanel.SloOverviewConfig != nil {
		priorSingle = tfPanel.SloOverviewConfig.Single
	}

	m := &models.SloOverviewSingleModel{}

	m.SloID = types.StringValue(api.SloId)

	if api.SloInstanceId != nil {
		switch {
		case priorSingle == nil && *api.SloInstanceId == "*":
			m.SloInstanceID = types.StringNull()
		case priorSingle != nil && priorSingle.SloInstanceID.IsNull() && *api.SloInstanceId == "*":
			m.SloInstanceID = types.StringNull()
		default:
			m.SloInstanceID = types.StringPointerValue(api.SloInstanceId)
		}
	} else {
		if priorSingle != nil && typeutils.IsKnown(priorSingle.SloInstanceID) {
			m.SloInstanceID = priorSingle.SloInstanceID
		} else {
			m.SloInstanceID = types.StringNull()
		}
	}

	var priorRemoteName types.String
	if priorSingle != nil {
		priorRemoteName = priorSingle.RemoteName
	}
	m.RemoteName = sloStringFromAPIOrPrior(api.RemoteName, priorRemoteName)

	var priorSingleTitle types.String
	if priorSingle != nil {
		priorSingleTitle = priorSingle.Title
	}
	m.Title = sloStringFromAPIOrPrior(api.Title, priorSingleTitle)

	var priorSingleDesc types.String
	if priorSingle != nil {
		priorSingleDesc = priorSingle.Description
	}
	m.Description = sloStringFromAPIOrPrior(api.Description, priorSingleDesc)

	var priorSingleHideTitle types.Bool
	if priorSingle != nil {
		priorSingleHideTitle = priorSingle.HideTitle
	}
	m.HideTitle = sloBoolFromAPIOrPrior(api.HideTitle, priorSingleHideTitle)

	var priorSingleHideBorder types.Bool
	if priorSingle != nil {
		priorSingleHideBorder = priorSingle.HideBorder
	}
	m.HideBorder = sloBoolFromAPIOrPrior(api.HideBorder, priorSingleHideBorder)

	if api.Drilldowns != nil {
		dds, err := json.Marshal(*api.Drilldowns)
		if err == nil {
			m.Drilldowns = drilldownsFromWireJSON(dds)
		}
	}

	pm.SloOverviewConfig = &models.SloOverviewConfigModel{Single: m}
	return diags
}

func sloGroupsFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, api kbapi.SloGroupOverviewEmbeddable) diag.Diagnostics {
	var diags diag.Diagnostics

	var priorGroups *models.SloOverviewGroupsModel
	if tfPanel != nil && tfPanel.SloOverviewConfig != nil {
		priorGroups = tfPanel.SloOverviewConfig.Groups
	}

	m := &models.SloOverviewGroupsModel{}

	var priorGroupsTitle types.String
	if priorGroups != nil {
		priorGroupsTitle = priorGroups.Title
	}
	m.Title = sloStringFromAPIOrPrior(api.Title, priorGroupsTitle)

	var priorGroupsDesc types.String
	if priorGroups != nil {
		priorGroupsDesc = priorGroups.Description
	}
	m.Description = sloStringFromAPIOrPrior(api.Description, priorGroupsDesc)

	var priorGroupsHideTitle types.Bool
	if priorGroups != nil {
		priorGroupsHideTitle = priorGroups.HideTitle
	}
	m.HideTitle = sloBoolFromAPIOrPrior(api.HideTitle, priorGroupsHideTitle)

	var priorGroupsHideBorder types.Bool
	if priorGroups != nil {
		priorGroupsHideBorder = priorGroups.HideBorder
	}
	m.HideBorder = sloBoolFromAPIOrPrior(api.HideBorder, priorGroupsHideBorder)

	if api.Drilldowns != nil {
		dds, err := json.Marshal(*api.Drilldowns)
		if err == nil {
			m.Drilldowns = drilldownsFromWireJSON(dds)
		}
	}

	if api.GroupFilters != nil && (priorGroups == nil || priorGroups.GroupFilters != nil) {
		gf := &models.SloGroupFiltersModel{}

		var priorGroupBy types.String
		if priorGroups != nil && priorGroups.GroupFilters != nil {
			priorGroupBy = priorGroups.GroupFilters.GroupBy
		}
		if api.GroupFilters.GroupBy != nil {
			gf.GroupBy = types.StringValue(string(*api.GroupFilters.GroupBy))
		} else {
			gf.GroupBy = sloStringFromAPIOrPrior(nil, priorGroupBy)
		}

		if api.GroupFilters.Groups != nil && len(*api.GroupFilters.Groups) > 0 {
			gf.Groups = make([]types.String, len(*api.GroupFilters.Groups))
			for i, g := range *api.GroupFilters.Groups {
				gf.Groups[i] = types.StringValue(g)
			}
		}

		var priorKQLQuery types.String
		if priorGroups != nil && priorGroups.GroupFilters != nil {
			priorKQLQuery = priorGroups.GroupFilters.KQLQuery
		}
		gf.KQLQuery = sloStringFromAPIOrPrior(api.GroupFilters.KqlQuery, priorKQLQuery)

		switch {
		case api.GroupFilters.Filters != nil && len(*api.GroupFilters.Filters) > 0:
			d := populateFiltersJSONFromAPI(*api.GroupFilters.Filters, &gf.FiltersJSON)
			diags.Append(d...)
		case priorGroups != nil && priorGroups.GroupFilters != nil && typeutils.IsKnown(priorGroups.GroupFilters.FiltersJSON):
			gf.FiltersJSON = priorGroups.GroupFilters.FiltersJSON
		default:
			gf.FiltersJSON = jsontypes.NewNormalizedNull()
		}

		m.GroupFilters = gf
	} else if priorGroups != nil && priorGroups.GroupFilters != nil {
		m.GroupFilters = priorGroups.GroupFilters
	}

	pm.SloOverviewConfig = &models.SloOverviewConfigModel{Groups: m}
	return diags
}

func drilldownsFromWireJSON(b []byte) []models.URLDrilldownModel {
	var wire []sloDrilldownWireJSON
	if err := json.Unmarshal(b, &wire); err != nil {
		return nil
	}
	result := make([]models.URLDrilldownModel, len(wire))
	for i, dd := range wire {
		result[i] = models.URLDrilldownModel{
			URL:   types.StringValue(dd.URL),
			Label: types.StringValue(dd.Label),
		}
		result[i].EncodeURL = types.BoolPointerValue(dd.EncodeURL)
		result[i].OpenInNewTab = types.BoolPointerValue(dd.OpenInNewTab)
	}
	return result
}

func populateFiltersJSONFromAPI(filters []kbapi.SloGroupOverviewEmbeddable_GroupFilters_Filters_Item, out *jsontypes.Normalized) diag.Diagnostics {
	return populateFilterJSONFromMarshaled(filters, out)
}
