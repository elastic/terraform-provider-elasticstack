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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	panelTypeSloOverview = "slo_overview"
)

// sloOverviewConfigModel is the top-level typed block for SLO overview panels.
// Exactly one of Single or Groups must be set.
type sloOverviewConfigModel struct {
	Single *sloSingleConfigModel `tfsdk:"single"`
	Groups *sloGroupsConfigModel `tfsdk:"groups"`
}

// sloSingleConfigModel holds single-SLO overview configuration.
type sloSingleConfigModel struct {
	SloID         types.String        `tfsdk:"slo_id"`
	SloInstanceID types.String        `tfsdk:"slo_instance_id"`
	RemoteName    types.String        `tfsdk:"remote_name"`
	Title         types.String        `tfsdk:"title"`
	Description   types.String        `tfsdk:"description"`
	HideTitle     types.Bool          `tfsdk:"hide_title"`
	HideBorder    types.Bool          `tfsdk:"hide_border"`
	Drilldowns    []sloDrilldownModel `tfsdk:"drilldowns"`
}

// sloGroupsConfigModel holds groups SLO overview configuration.
type sloGroupsConfigModel struct {
	Title        types.String          `tfsdk:"title"`
	Description  types.String          `tfsdk:"description"`
	HideTitle    types.Bool            `tfsdk:"hide_title"`
	HideBorder   types.Bool            `tfsdk:"hide_border"`
	Drilldowns   []sloDrilldownModel   `tfsdk:"drilldowns"`
	GroupFilters *sloGroupFiltersModel `tfsdk:"group_filters"`
}

// sloGroupFiltersModel holds the group_filters block for groups mode.
type sloGroupFiltersModel struct {
	GroupBy     types.String         `tfsdk:"group_by"`
	Groups      []types.String       `tfsdk:"groups"`
	KQLQuery    types.String         `tfsdk:"kql_query"`
	FiltersJSON jsontypes.Normalized `tfsdk:"filters_json"`
}

// sloDrilldownModel holds one URL drilldown entry.
type sloDrilldownModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	Type         types.String `tfsdk:"type"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

// sloDrilldownWireJSON is the JSON wire format for drilldown entries.
// It uses snake_case JSON tags matching the Kibana API.
type sloDrilldownWireJSON struct {
	EncodeURL    *bool  `json:"encode_url,omitempty"`
	Label        string `json:"label"`
	OpenInNewTab *bool  `json:"open_in_new_tab,omitempty"`
	Trigger      string `json:"trigger"`
	Type         string `json:"type"`
	URL          string `json:"url"`
}

// sloOverviewToAPI converts TF model to Kibana API panel item.
func sloOverviewToAPI(pm panelModel, grid struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
}, uid *string) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := pm.SloOverviewConfig

	var config kbapi.KbnDashboardPanelSloOverview_Config

	if cfg.Single != nil {
		single, d := singleToAPI(cfg.Single)
		diags.Append(d...)
		if diags.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		// NOTE: config.FromSloSingleOverviewEmbeddable overwrites OverviewMode to the
		// discriminator string "slo-single-overview-embeddable"; bypass it and marshal
		// directly so that overview_mode = "single" is preserved in the payload.
		b, err := json.Marshal(single)
		if err != nil {
			diags.AddError("Failed to marshal SLO single overview config", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		if err := config.UnmarshalJSON(b); err != nil {
			diags.AddError("Failed to set SLO single overview config", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
	} else if cfg.Groups != nil {
		groups, d := groupsToAPI(cfg.Groups)
		diags.Append(d...)
		if diags.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		// Same workaround: FromSloGroupOverviewEmbeddable would overwrite overview_mode
		// to "slo-group-overview-embeddable"; marshal directly to preserve "groups".
		b, err := json.Marshal(groups)
		if err != nil {
			diags.AddError("Failed to marshal SLO groups overview config", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		if err := config.UnmarshalJSON(b); err != nil {
			diags.AddError("Failed to set SLO groups overview config", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
	}

	panel := kbapi.KbnDashboardPanelSloOverview{
		Config: config,
		Grid: struct {
			H *float32 `json:"h,omitempty"`
			W *float32 `json:"w,omitempty"`
			X float32  `json:"x"`
			Y float32  `json:"y"`
		}{
			H: grid.H,
			W: grid.W,
			X: grid.X,
			Y: grid.Y,
		},
		Type: kbapi.SloOverview,
		Uid:  uid,
	}

	var item kbapi.DashboardPanelItem
	if err := item.FromKbnDashboardPanelSloOverview(panel); err != nil {
		diags.AddError("Failed to create SLO overview panel item", err.Error())
	}
	return item, diags
}

// singleToAPI converts a sloSingleConfigModel to the kbapi SloSingleOverviewEmbeddable type.
// Drilldowns are serialized via JSON to avoid referencing the anonymous field type.
func singleToAPI(m *sloSingleConfigModel) (kbapi.SloSingleOverviewEmbeddable, diag.Diagnostics) {
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

// setDrilldownsOnSingle sets the Drilldowns field on a SloSingleOverviewEmbeddable.
func setDrilldownsOnSingle(api *kbapi.SloSingleOverviewEmbeddable, drilldowns []sloDrilldownModel) diag.Diagnostics {
	return injectDrilldownsJSON(api, drilldowns)
}

// groupsToAPI converts a sloGroupsConfigModel to the kbapi SloGroupOverviewEmbeddable type.
func groupsToAPI(m *sloGroupsConfigModel) (kbapi.SloGroupOverviewEmbeddable, diag.Diagnostics) {
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

// setDrilldownsOnGroups sets the Drilldowns field on a SloGroupOverviewEmbeddable.
func setDrilldownsOnGroups(api *kbapi.SloGroupOverviewEmbeddable, drilldowns []sloDrilldownModel) diag.Diagnostics {
	return injectDrilldownsJSON(api, drilldowns)
}

// injectDrilldownsJSON marshals api to JSON, overlays the "drilldowns" key with the
// provided drilldown models, and unmarshals back. Used for both single and groups SLO
// overview embeddables whose Drilldowns field has an anonymous struct type that cannot
// be referenced directly in typed Go code.
func injectDrilldownsJSON(api any, drilldowns []sloDrilldownModel) diag.Diagnostics {
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

// buildDrilldownsWire converts TF drilldown models to JSON wire format.
func buildDrilldownsWire(drilldowns []sloDrilldownModel) []sloDrilldownWireJSON {
	result := make([]sloDrilldownWireJSON, len(drilldowns))
	for i, dd := range drilldowns {
		result[i] = sloDrilldownWireJSON{
			URL:     dd.URL.ValueString(),
			Label:   dd.Label.ValueString(),
			Trigger: dd.Trigger.ValueString(),
			Type:    dd.Type.ValueString(),
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

func groupFiltersToAPI(m *sloGroupFiltersModel) (*struct {
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

// sloOverviewFromAPI reads a KbnDashboardPanelSloOverview from the API response and
// populates the panel model. tfPanel is the prior TF state/plan (may be nil on import).
func sloOverviewFromAPI(pm *panelModel, tfPanel *panelModel, panel kbapi.KbnDashboardPanelSloOverview) diag.Diagnostics {
	var diags diag.Diagnostics

	discriminator, err := panel.Config.Discriminator()
	if err != nil {
		diags.AddError("Failed to determine SLO overview mode", err.Error())
		return diags
	}

	// The generated kbapi code sets overview_mode to the embeddable type name when writing
	// (e.g. "slo-single-overview-embeddable"), so discriminate on that value.
	switch discriminator {
	case "slo-single-overview-embeddable", "single":
		single, err := panel.Config.AsSloSingleOverviewEmbeddable()
		if err != nil {
			diags.AddError("Failed to read SLO single overview config", err.Error())
			return diags
		}
		return sloSingleFromAPI(pm, tfPanel, single)
	case "slo-group-overview-embeddable", "groups":
		groups, err := panel.Config.AsSloGroupOverviewEmbeddable()
		if err != nil {
			diags.AddError("Failed to read SLO groups overview config", err.Error())
			return diags
		}
		return sloGroupsFromAPI(pm, tfPanel, groups)
	default:
		diags.AddError("Unknown SLO overview mode", "Expected 'slo-single-overview-embeddable' or 'slo-group-overview-embeddable', got: "+discriminator)
		return diags
	}
}

// sloStringFromAPIOrPrior returns the API string value if set, or the prior state value if known,
// or null if neither is available.
func sloStringFromAPIOrPrior(apiVal *string, priorVal types.String) types.String {
	if apiVal != nil {
		return types.StringPointerValue(apiVal)
	}
	if typeutils.IsKnown(priorVal) {
		return priorVal
	}
	return types.StringNull()
}

// sloBoolFromAPIOrPrior returns the API bool value if set, or the prior state value if known,
// or null if neither is available.
func sloBoolFromAPIOrPrior(apiVal *bool, priorVal types.Bool) types.Bool {
	if apiVal != nil {
		return types.BoolPointerValue(apiVal)
	}
	if typeutils.IsKnown(priorVal) {
		return priorVal
	}
	return types.BoolNull()
}

func sloSingleFromAPI(pm *panelModel, tfPanel *panelModel, api kbapi.SloSingleOverviewEmbeddable) diag.Diagnostics {
	var diags diag.Diagnostics

	// Determine prior state for null-preservation
	var priorSingle *sloSingleConfigModel
	if tfPanel != nil && tfPanel.SloOverviewConfig != nil {
		priorSingle = tfPanel.SloOverviewConfig.Single
	}

	m := &sloSingleConfigModel{}

	m.SloID = types.StringValue(api.SloId)

	// slo_instance_id null-preservation: if prior state was null and API returns "*", preserve null
	if api.SloInstanceId != nil {
		if priorSingle != nil && priorSingle.SloInstanceID.IsNull() && *api.SloInstanceId == "*" {
			m.SloInstanceID = types.StringNull()
		} else {
			m.SloInstanceID = types.StringPointerValue(api.SloInstanceId)
		}
	} else {
		// API omitted it — preserve prior state if known, else null
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

	pm.SloOverviewConfig = &sloOverviewConfigModel{Single: m}
	return diags
}

func sloGroupsFromAPI(pm *panelModel, tfPanel *panelModel, api kbapi.SloGroupOverviewEmbeddable) diag.Diagnostics {
	var diags diag.Diagnostics

	var priorGroups *sloGroupsConfigModel
	if tfPanel != nil && tfPanel.SloOverviewConfig != nil {
		priorGroups = tfPanel.SloOverviewConfig.Groups
	}

	m := &sloGroupsConfigModel{}

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
		// Null-preservation: if prior state had a group_filters block (or this is an import with
		// no prior state), populate from API. If prior state had no group_filters block, keep null
		// even when the API echoes back defaults (e.g. group_by="status").
		gf := &sloGroupFiltersModel{}

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
		// Prior state had group_filters but API returned none — preserve prior block to avoid
		// spurious diffs (nil API response means no change).
		m.GroupFilters = priorGroups.GroupFilters
	}

	pm.SloOverviewConfig = &sloOverviewConfigModel{Groups: m}
	return diags
}

// drilldownsFromWireJSON decodes a JSON array of drilldown objects into TF models.
func drilldownsFromWireJSON(b []byte) []sloDrilldownModel {
	var wire []sloDrilldownWireJSON
	if err := json.Unmarshal(b, &wire); err != nil {
		return nil
	}
	result := make([]sloDrilldownModel, len(wire))
	for i, dd := range wire {
		result[i] = sloDrilldownModel{
			URL:     types.StringValue(dd.URL),
			Label:   types.StringValue(dd.Label),
			Trigger: types.StringValue(dd.Trigger),
			Type:    types.StringValue(dd.Type),
		}
		result[i].EncodeURL = types.BoolPointerValue(dd.EncodeURL)
		result[i].OpenInNewTab = types.BoolPointerValue(dd.OpenInNewTab)
	}
	return result
}

// populateFiltersJSONFromAPI marshals the API filter items to normalized JSON.
func populateFiltersJSONFromAPI(filters []kbapi.SloGroupOverviewEmbeddable_GroupFilters_Filters_Item, out *jsontypes.Normalized) diag.Diagnostics {
	return populateFilterJSONFromMarshaled(filters, out)
}
