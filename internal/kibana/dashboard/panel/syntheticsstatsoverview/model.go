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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes the TF model into the API panel struct.
// When the config block is nil or entirely null, an empty config object is sent (valid: shows all monitors).
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsStatsOverview) diag.Diagnostics {
	cfg := pm.SyntheticsStatsOverviewConfig
	if cfg == nil {
		return nil
	}

	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&panel.Config.Title, &panel.Config.Description, &panel.Config.HideTitle, &panel.Config.HideBorder)

	var diags diag.Diagnostics
	if len(cfg.Drilldowns) > 0 {
		diags.Append(panelkit.InjectDrilldownsJSON(&panel.Config, cfg.Drilldowns)...)
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
	return diags
}

// PopulateFromAPI reads back a synthetics stats overview panel from the API response.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsStatsOverview) diag.Diagnostics {
	cfg := apiPanel.Config

	if prior == nil {
		if cfg.Title == nil && cfg.Description == nil && cfg.HideTitle == nil && cfg.HideBorder == nil &&
			(cfg.Drilldowns == nil || len(*cfg.Drilldowns) == 0) && !syntheticsFiltersHasAnyEntry(cfg.Filters) {
			return nil
		}
		pm.SyntheticsStatsOverviewConfig = syntheticsStatsOverviewConfigFromAPIImport(apiPanel)
		return nil
	}

	if pm.SyntheticsStatsOverviewConfig == nil && prior.SyntheticsStatsOverviewConfig != nil {
		pm.SyntheticsStatsOverviewConfig = syntheticsStatsOverviewConfigFromAPIImport(apiPanel)
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

	panelkit.ApplyPresentationFromAPI(&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder,
		cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder)

	var priorDrilldowns []models.URLDrilldownModel
	if prior.SyntheticsStatsOverviewConfig != nil {
		priorDrilldowns = prior.SyntheticsStatsOverviewConfig.Drilldowns
	}
	existing.Drilldowns = readSyntheticsStatsOverviewDrilldownsFromAPI(apiPanel, priorDrilldowns)
	existing.Filters = readSyntheticsStatsOverviewFiltersFromAPI(apiPanel, existing.Filters)
	return nil
}

func syntheticsStatsOverviewConfigFromAPIImport(apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsStatsOverview) *models.SyntheticsStatsOverviewConfigModel {
	cfg := apiPanel.Config
	return &models.SyntheticsStatsOverviewConfigModel{
		Title:       types.StringPointerValue(cfg.Title),
		Description: types.StringPointerValue(cfg.Description),
		HideTitle:   types.BoolPointerValue(cfg.HideTitle),
		HideBorder:  types.BoolPointerValue(cfg.HideBorder),
		Drilldowns:  readSyntheticsStatsOverviewDrilldownsFromAPI(apiPanel, nil),
		Filters:     readSyntheticsStatsOverviewFiltersFromAPI(apiPanel, nil),
	}
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
	apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsStatsOverview,
	priorDrilldowns []models.URLDrilldownModel,
) []models.URLDrilldownModel {
	return panelkit.MapURLDrilldownsFromAPI(apiPanel.Config.Drilldowns, func(d struct {
		EncodeUrl    *bool                                                                                   `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                                                  `json:"label"`
		OpenInNewTab *bool                                                                                   `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsStatsOverviewConfigDrilldownsTrigger `json:"trigger"`
		Type         kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsStatsOverviewConfigDrilldownsType    `json:"type"`
		Url          string                                                                                  `json:"url"` //nolint:revive
	}) panelkit.URLDrilldownAPIItemData {
		return panelkit.URLDrilldownAPIItemData{URL: d.Url, Label: d.Label, EncodeUrl: d.EncodeUrl, OpenInNewTab: d.OpenInNewTab}
	}, priorDrilldowns)
}

func readSyntheticsStatsOverviewFiltersFromAPI(
	apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsStatsOverview,
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
