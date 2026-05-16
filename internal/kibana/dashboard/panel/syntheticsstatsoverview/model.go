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

package syntheticsstatsoverview

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes the TF model into the API panel struct.
// When the config block is nil or entirely null, an empty config object is sent (valid: shows all monitors).
func BuildConfig(pm models.PanelModel, panel *kbapi.KbnDashboardPanelTypeSyntheticsStatsOverview) diag.Diagnostics {
	cfg := pm.SyntheticsStatsOverviewConfig
	if cfg == nil {
		return nil
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
			EncodeUrl    *bool                                                                     `json:"encode_url,omitempty"` //nolint:revive
			Label        string                                                                    `json:"label"`
			OpenInNewTab *bool                                                                     `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.KbnDashboardPanelTypeSyntheticsStatsOverviewConfigDrilldownsTrigger `json:"trigger"`
			Type         kbapi.KbnDashboardPanelTypeSyntheticsStatsOverviewConfigDrilldownsType    `json:"type"`
			Url          string                                                                    `json:"url"` //nolint:revive
		}, len(cfg.Drilldowns))

		for i, d := range cfg.Drilldowns {
			drilldowns[i].Url = d.URL.ValueString()
			drilldowns[i].Label = d.Label.ValueString()
			drilldowns[i].Trigger = kbapi.KbnDashboardPanelTypeSyntheticsStatsOverviewConfigDrilldownsTriggerOnOpenPanelMenu
			drilldowns[i].Type = kbapi.KbnDashboardPanelTypeSyntheticsStatsOverviewConfigDrilldownsTypeUrlDrilldown
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
		type apiFilterItem = struct {
			Label string `json:"label"`
			Value string `json:"value"`
		}

		toAPIItems := func(items []models.SyntheticsFilterItemModel) *[]apiFilterItem {
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
	return nil
}

// PopulateFromAPI reads back a synthetics stats overview panel from the API response.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiPanel kbapi.KbnDashboardPanelTypeSyntheticsStatsOverview) diag.Diagnostics {
	cfg := apiPanel.Config

	if prior == nil {
		if cfg.Title == nil && cfg.Description == nil && cfg.HideTitle == nil && cfg.HideBorder == nil &&
			(cfg.Drilldowns == nil || len(*cfg.Drilldowns) == 0) && !syntheticsFiltersHasAnyEntry(cfg.Filters) {
			return nil
		}
		pm.SyntheticsStatsOverviewConfig = &models.SyntheticsStatsOverviewConfigModel{
			Title:       types.StringPointerValue(cfg.Title),
			Description: types.StringPointerValue(cfg.Description),
			HideTitle:   types.BoolPointerValue(cfg.HideTitle),
			HideBorder:  types.BoolPointerValue(cfg.HideBorder),
			Drilldowns:  readSyntheticsStatsOverviewDrilldownsFromAPI(apiPanel, nil),
			Filters:     readSyntheticsStatsOverviewFiltersFromAPI(apiPanel, nil),
		}
		return nil
	}

	existing := pm.SyntheticsStatsOverviewConfig

	if existing == nil {
		return nil
	}

	if cfg.Title == nil && cfg.Description == nil && cfg.HideTitle == nil && cfg.HideBorder == nil &&
		(cfg.Drilldowns == nil || len(*cfg.Drilldowns) == 0) && cfg.Filters == nil {
		pm.SyntheticsStatsOverviewConfig = nil
		return nil
	}

	existing.Title = panelkit.PreserveString(existing.Title, cfg.Title)
	existing.Description = panelkit.PreserveString(existing.Description, cfg.Description)
	existing.HideTitle = panelkit.PreserveBool(existing.HideTitle, cfg.HideTitle)
	existing.HideBorder = panelkit.PreserveBool(existing.HideBorder, cfg.HideBorder)

	existing.Drilldowns = readSyntheticsStatsOverviewDrilldownsFromAPI(apiPanel, existing.Drilldowns)
	existing.Filters = readSyntheticsStatsOverviewFiltersFromAPI(apiPanel, existing.Filters)
	return nil
}

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

func readSyntheticsStatsOverviewDrilldownsFromAPI(
	apiPanel kbapi.KbnDashboardPanelTypeSyntheticsStatsOverview,
	priorDrilldowns []models.URLDrilldownModel,
) []models.URLDrilldownModel {
	apiDrilldowns := apiPanel.Config.Drilldowns
	if apiDrilldowns == nil || len(*apiDrilldowns) == 0 {
		return nil
	}

	result := make([]models.URLDrilldownModel, len(*apiDrilldowns))
	for i, d := range *apiDrilldowns {
		result[i] = models.URLDrilldownModel{
			URL:   types.StringValue(d.Url),
			Label: types.StringValue(d.Label),
		}

		var prior *models.URLDrilldownModel
		if i < len(priorDrilldowns) {
			prior = &priorDrilldowns[i]
		}

		switch {
		case prior != nil && prior.EncodeURL.IsNull():
			result[i].EncodeURL = types.BoolNull()
		case d.EncodeUrl != nil:
			result[i].EncodeURL = types.BoolValue(*d.EncodeUrl)
		default:
			result[i].EncodeURL = types.BoolNull()
		}

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

func readSyntheticsStatsOverviewFiltersFromAPI(
	apiPanel kbapi.KbnDashboardPanelTypeSyntheticsStatsOverview,
	priorFilters *models.SyntheticsFiltersModel,
) *models.SyntheticsFiltersModel {
	apiFilters := apiPanel.Config.Filters

	if apiFilters == nil {
		return priorFilters
	}

	if !syntheticsFiltersHasAnyEntry(apiFilters) {
		return nil
	}

	type apiFilterItem = struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	fromAPIItems := func(items *[]apiFilterItem) []models.SyntheticsFilterItemModel {
		if items == nil || len(*items) == 0 {
			return nil
		}
		out := make([]models.SyntheticsFilterItemModel, len(*items))
		for i, it := range *items {
			out[i] = models.SyntheticsFilterItemModel{
				Label: types.StringValue(it.Label),
				Value: types.StringValue(it.Value),
			}
		}
		return out
	}

	return &models.SyntheticsFiltersModel{
		Projects:     fromAPIItems(apiFilters.Projects),
		Tags:         fromAPIItems(apiFilters.Tags),
		MonitorIDs:   fromAPIItems(apiFilters.MonitorIds),
		Locations:    fromAPIItems(apiFilters.Locations),
		MonitorTypes: fromAPIItems(apiFilters.MonitorTypes),
	}
}
