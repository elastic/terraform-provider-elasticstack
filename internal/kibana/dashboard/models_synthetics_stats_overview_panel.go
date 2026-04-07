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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// syntheticsStatsOverviewConfigModel is the Terraform model for the synthetics_stats_overview_config block.
type syntheticsStatsOverviewConfigModel struct {
	Title       types.String                            `tfsdk:"title"`
	Description types.String                            `tfsdk:"description"`
	HideTitle   types.Bool                              `tfsdk:"hide_title"`
	HideBorder  types.Bool                              `tfsdk:"hide_border"`
	Drilldowns  []syntheticsStatsOverviewDrilldownModel `tfsdk:"drilldowns"`
	Filters     *syntheticsStatsOverviewFiltersModel    `tfsdk:"filters"`
}

// syntheticsStatsOverviewDrilldownModel holds one URL drilldown entry.
type syntheticsStatsOverviewDrilldownModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	Type         types.String `tfsdk:"type"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

// syntheticsStatsOverviewFiltersModel holds per-category Synthetics monitor filter constraints.
type syntheticsStatsOverviewFiltersModel struct {
	Projects     []syntheticsFilterItemModel `tfsdk:"projects"`
	Tags         []syntheticsFilterItemModel `tfsdk:"tags"`
	MonitorIDs   []syntheticsFilterItemModel `tfsdk:"monitor_ids"`
	Locations    []syntheticsFilterItemModel `tfsdk:"locations"`
	MonitorTypes []syntheticsFilterItemModel `tfsdk:"monitor_types"`
}

// syntheticsFilterItemModel holds a single { label, value } filter option.
type syntheticsFilterItemModel struct {
	Label types.String `tfsdk:"label"`
	Value types.String `tfsdk:"value"`
}

// buildSyntheticsStatsOverviewConfig writes the TF model into the API panel struct.
// When the config block is nil or entirely null, an empty config object is sent (valid: shows all monitors).
func buildSyntheticsStatsOverviewConfig(pm panelModel, panel *kbapi.KbnDashboardPanelSyntheticsStatsOverview) {
	cfg := pm.SyntheticsStatsOverviewConfig
	if cfg == nil {
		return
	}

	if typeutils.IsKnown(cfg.Title) {
		panel.Config.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Description) {
		panel.Config.Description = cfg.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.HideTitle) {
		panel.Config.HideTitle = cfg.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.HideBorder) {
		panel.Config.HideBorder = cfg.HideBorder.ValueBoolPointer()
	}

	if len(cfg.Drilldowns) > 0 {
		drilldowns := make([]struct {
			EncodeUrl    *bool                                                                 `json:"encode_url,omitempty"` //nolint:revive
			Label        string                                                                `json:"label"`
			OpenInNewTab *bool                                                                 `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTrigger `json:"trigger"`
			Type         kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsType    `json:"type"`
			Url          string                                                                `json:"url"` //nolint:revive
		}, len(cfg.Drilldowns))

		for i, d := range cfg.Drilldowns {
			drilldowns[i].Url = d.URL.ValueString()
			drilldowns[i].Label = d.Label.ValueString()
			drilldowns[i].Trigger = kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTrigger(d.Trigger.ValueString())
			drilldowns[i].Type = kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsType(d.Type.ValueString())
			if typeutils.IsKnown(d.EncodeURL) {
				drilldowns[i].EncodeUrl = d.EncodeURL.ValueBoolPointer()
			}
			if typeutils.IsKnown(d.OpenInNewTab) {
				drilldowns[i].OpenInNewTab = d.OpenInNewTab.ValueBoolPointer()
			}
		}
		panel.Config.Drilldowns = &drilldowns
	}

	if cfg.Filters != nil {
		// apiFilterItem is a type alias for the anonymous filter-entry struct shared by all
		// filter categories in the Kibana API. The alias lets us write a single inline helper
		// without repeating the struct literal for each category.
		type apiFilterItem = struct {
			Label string `json:"label"`
			Value string `json:"value"`
		}

		toAPIItems := func(items []syntheticsFilterItemModel) *[]apiFilterItem {
			if len(items) == 0 {
				return nil
			}
			out := make([]apiFilterItem, len(items))
			for i, it := range items {
				out[i] = apiFilterItem{Label: it.Label.ValueString(), Value: it.Value.ValueString()}
			}
			return &out
		}

		projects := toAPIItems(cfg.Filters.Projects)
		tags := toAPIItems(cfg.Filters.Tags)
		monitorIDs := toAPIItems(cfg.Filters.MonitorIDs)
		locations := toAPIItems(cfg.Filters.Locations)
		monitorTypes := toAPIItems(cfg.Filters.MonitorTypes)

		// Only set the filters struct when at least one category is non-empty.
		if projects != nil || tags != nil || monitorIDs != nil || locations != nil || monitorTypes != nil {
			panel.Config.Filters = &struct {
				Locations *[]struct {
					Label string `json:"label"`
					Value string `json:"value"`
				} `json:"locations,omitempty"`
				MonitorIds *[]struct { //nolint:revive
					Label string `json:"label"`
					Value string `json:"value"`
				} `json:"monitor_ids,omitempty"`
				MonitorTypes *[]struct {
					Label string `json:"label"`
					Value string `json:"value"`
				} `json:"monitor_types,omitempty"`
				Projects *[]struct {
					Label string `json:"label"`
					Value string `json:"value"`
				} `json:"projects,omitempty"`
				Tags *[]struct {
					Label string `json:"label"`
					Value string `json:"value"`
				} `json:"tags,omitempty"`
			}{
				Projects:     projects,
				Tags:         tags,
				MonitorIds:   monitorIDs,
				Locations:    locations,
				MonitorTypes: monitorTypes,
			}
		}
	}
}

// populateSyntheticsStatsOverviewFromAPI reads back a synthetics stats overview panel from the
// API response and updates the panel model. Null-preservation semantics apply.
//
// tfPanel is the prior TF state/plan panel, or nil on import. When nil, all API-returned
// fields are populated unconditionally (no prior intent to preserve).
func populateSyntheticsStatsOverviewFromAPI(pm *panelModel, tfPanel *panelModel, apiPanel kbapi.KbnDashboardPanelSyntheticsStatsOverview) {
	cfg := apiPanel.Config

	// On import (tfPanel == nil), populate unconditionally when at least one API field is set.
	if tfPanel == nil {
		if cfg.Title == nil && cfg.Description == nil && cfg.HideTitle == nil && cfg.HideBorder == nil &&
			(cfg.Drilldowns == nil || len(*cfg.Drilldowns) == 0) && !syntheticsFiltersHasAnyEntry(cfg.Filters) {
			// Empty config — keep block null.
			return
		}
		pm.SyntheticsStatsOverviewConfig = &syntheticsStatsOverviewConfigModel{
			Title:       types.StringPointerValue(cfg.Title),
			Description: types.StringPointerValue(cfg.Description),
			HideTitle:   types.BoolPointerValue(cfg.HideTitle),
			HideBorder:  types.BoolPointerValue(cfg.HideBorder),
			Drilldowns:  readSyntheticsStatsOverviewDrilldownsFromAPI(apiPanel, nil),
			Filters:     readSyntheticsStatsOverviewFiltersFromAPI(apiPanel, nil),
		}
		return
	}

	existing := pm.SyntheticsStatsOverviewConfig

	// If prior state had no config block, preserve nil intent.
	if existing == nil {
		return
	}

	// Block exists in state — apply null-preservation per field.
	if typeutils.IsKnown(existing.Title) {
		existing.Title = types.StringPointerValue(cfg.Title)
	}
	if typeutils.IsKnown(existing.Description) {
		existing.Description = types.StringPointerValue(cfg.Description)
	}
	if typeutils.IsKnown(existing.HideTitle) {
		existing.HideTitle = types.BoolPointerValue(cfg.HideTitle)
	}
	if typeutils.IsKnown(existing.HideBorder) {
		existing.HideBorder = types.BoolPointerValue(cfg.HideBorder)
	}

	existing.Drilldowns = readSyntheticsStatsOverviewDrilldownsFromAPI(apiPanel, existing.Drilldowns)
	existing.Filters = readSyntheticsStatsOverviewFiltersFromAPI(apiPanel, existing.Filters)
}

// syntheticsFiltersHasAnyEntry returns true if the filters object has at least one non-empty category.
func syntheticsFiltersHasAnyEntry(f *struct {
	Locations *[]struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"locations,omitempty"`
	MonitorIds *[]struct { //nolint:revive
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"monitor_ids,omitempty"`
	MonitorTypes *[]struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"monitor_types,omitempty"`
	Projects *[]struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"projects,omitempty"`
	Tags *[]struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"tags,omitempty"`
}) bool {
	if f == nil {
		return false
	}
	return (f.Projects != nil && len(*f.Projects) > 0) ||
		(f.Tags != nil && len(*f.Tags) > 0) ||
		(f.MonitorIds != nil && len(*f.MonitorIds) > 0) ||
		(f.Locations != nil && len(*f.Locations) > 0) ||
		(f.MonitorTypes != nil && len(*f.MonitorTypes) > 0)
}

// readSyntheticsStatsOverviewDrilldownsFromAPI converts API drilldowns to TF models.
// priorDrilldowns is the existing TF state (may be nil on import).
// Optional bool fields (encode_url, open_in_new_tab) use null-preservation when prior state is available.
func readSyntheticsStatsOverviewDrilldownsFromAPI(
	apiPanel kbapi.KbnDashboardPanelSyntheticsStatsOverview,
	priorDrilldowns []syntheticsStatsOverviewDrilldownModel,
) []syntheticsStatsOverviewDrilldownModel {
	apiDrilldowns := apiPanel.Config.Drilldowns
	if apiDrilldowns == nil || len(*apiDrilldowns) == 0 {
		return nil
	}

	result := make([]syntheticsStatsOverviewDrilldownModel, len(*apiDrilldowns))
	for i, d := range *apiDrilldowns {
		result[i] = syntheticsStatsOverviewDrilldownModel{
			URL:     types.StringValue(d.Url),
			Label:   types.StringValue(d.Label),
			Trigger: types.StringValue(string(d.Trigger)),
			Type:    types.StringValue(string(d.Type)),
		}

		var prior *syntheticsStatsOverviewDrilldownModel
		if i < len(priorDrilldowns) {
			prior = &priorDrilldowns[i]
		}

		// encode_url: null-preserve if prior was null; otherwise populate from API.
		switch {
		case prior != nil && prior.EncodeURL.IsNull():
			result[i].EncodeURL = types.BoolNull()
		case d.EncodeUrl != nil:
			result[i].EncodeURL = types.BoolValue(*d.EncodeUrl)
		default:
			result[i].EncodeURL = types.BoolNull()
		}

		// open_in_new_tab: null-preserve if prior was null; otherwise populate from API.
		switch {
		case prior != nil && prior.OpenInNewTab.IsNull():
			result[i].OpenInNewTab = types.BoolNull()
		case d.OpenInNewTab != nil:
			result[i].OpenInNewTab = types.BoolValue(*d.OpenInNewTab)
		default:
			result[i].OpenInNewTab = types.BoolNull()
		}
	}
	return result
}

// readSyntheticsStatsOverviewFiltersFromAPI converts API filters to TF model.
// Treats a nil or empty filters object as equivalent to an absent block.
// priorFilters is the existing TF state (may be nil on import).
func readSyntheticsStatsOverviewFiltersFromAPI(
	apiPanel kbapi.KbnDashboardPanelSyntheticsStatsOverview,
	priorFilters *syntheticsStatsOverviewFiltersModel,
) *syntheticsStatsOverviewFiltersModel {
	apiFilters := apiPanel.Config.Filters

	// Treat nil or empty filters as absent block.
	if !syntheticsFiltersHasAnyEntry(apiFilters) {
		// If prior state had a filters block, preserve it (avoids drift from API not returning filters).
		if priorFilters != nil {
			return priorFilters
		}
		return nil
	}

	// apiFilterItem is a type alias for the anonymous filter-entry struct shared by all categories.
	type apiFilterItem = struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	fromAPIItems := func(items *[]apiFilterItem) []syntheticsFilterItemModel {
		if items == nil || len(*items) == 0 {
			return nil
		}
		out := make([]syntheticsFilterItemModel, len(*items))
		for i, it := range *items {
			out[i] = syntheticsFilterItemModel{
				Label: types.StringValue(it.Label),
				Value: types.StringValue(it.Value),
			}
		}
		return out
	}

	return &syntheticsStatsOverviewFiltersModel{
		Projects:     fromAPIItems(apiFilters.Projects),
		Tags:         fromAPIItems(apiFilters.Tags),
		MonitorIDs:   fromAPIItems(apiFilters.MonitorIds),
		Locations:    fromAPIItems(apiFilters.Locations),
		MonitorTypes: fromAPIItems(apiFilters.MonitorTypes),
	}
}
