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
	Title       types.String                        `tfsdk:"title"`
	Description types.String                        `tfsdk:"description"`
	HideTitle   types.Bool                          `tfsdk:"hide_title"`
	HideBorder  types.Bool                          `tfsdk:"hide_border"`
	Drilldowns  []syntheticsStatsOverviewDrilldownModel  `tfsdk:"drilldowns"`
	Filters     *syntheticsStatsOverviewFiltersModel `tfsdk:"filters"`
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
			EncodeUrl    *bool                                                           `json:"encode_url,omitempty"`    //nolint:revive
			Label        string                                                          `json:"label"`
			OpenInNewTab *bool                                                           `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTrigger `json:"trigger"`
			Type         kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsType    `json:"type"`
			Url          string                                                          `json:"url"` //nolint:revive
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
		filters := &struct {
			Locations *[]struct {
				Label string `json:"label"`
				Value string `json:"value"`
			} `json:"locations,omitempty"`
			MonitorIds *[]struct {
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
		}{}

		if len(cfg.Filters.Projects) > 0 {
			items := make([]struct {
				Label string `json:"label"`
				Value string `json:"value"`
			}, len(cfg.Filters.Projects))
			for i, p := range cfg.Filters.Projects {
				items[i] = struct {
					Label string `json:"label"`
					Value string `json:"value"`
				}{Label: p.Label.ValueString(), Value: p.Value.ValueString()}
			}
			filters.Projects = &items
		}

		if len(cfg.Filters.Tags) > 0 {
			items := make([]struct {
				Label string `json:"label"`
				Value string `json:"value"`
			}, len(cfg.Filters.Tags))
			for i, t := range cfg.Filters.Tags {
				items[i] = struct {
					Label string `json:"label"`
					Value string `json:"value"`
				}{Label: t.Label.ValueString(), Value: t.Value.ValueString()}
			}
			filters.Tags = &items
		}

		if len(cfg.Filters.MonitorIDs) > 0 {
			items := make([]struct {
				Label string `json:"label"`
				Value string `json:"value"`
			}, len(cfg.Filters.MonitorIDs))
			for i, m := range cfg.Filters.MonitorIDs {
				items[i] = struct {
					Label string `json:"label"`
					Value string `json:"value"`
				}{Label: m.Label.ValueString(), Value: m.Value.ValueString()}
			}
			filters.MonitorIds = &items
		}

		if len(cfg.Filters.Locations) > 0 {
			items := make([]struct {
				Label string `json:"label"`
				Value string `json:"value"`
			}, len(cfg.Filters.Locations))
			for i, l := range cfg.Filters.Locations {
				items[i] = struct {
					Label string `json:"label"`
					Value string `json:"value"`
				}{Label: l.Label.ValueString(), Value: l.Value.ValueString()}
			}
			filters.Locations = &items
		}

		if len(cfg.Filters.MonitorTypes) > 0 {
			items := make([]struct {
				Label string `json:"label"`
				Value string `json:"value"`
			}, len(cfg.Filters.MonitorTypes))
			for i, mt := range cfg.Filters.MonitorTypes {
				items[i] = struct {
					Label string `json:"label"`
					Value string `json:"value"`
				}{Label: mt.Label.ValueString(), Value: mt.Value.ValueString()}
			}
			filters.MonitorTypes = &items
		}

		// Only set the filters struct if at least one category is non-empty.
		if filters.Projects != nil || filters.Tags != nil || filters.MonitorIds != nil ||
			filters.Locations != nil || filters.MonitorTypes != nil {
			panel.Config.Filters = filters
		}
	}
}

// populateSyntheticsStatsOverviewFromAPI reads back a synthetics stats overview config from the
// API response and updates the panel model. Null-preservation semantics apply.
//
// tfPanel is the prior TF state/plan panel, or nil on import. When nil, all API-returned
// fields are populated unconditionally.
func populateSyntheticsStatsOverviewFromAPI(pm *panelModel, tfPanel *panelModel, apiConfig struct {
	Description *string `json:"description,omitempty"`
	Drilldowns  *[]struct {
		EncodeUrl    *bool                                                           `json:"encode_url,omitempty"`    //nolint:revive
		Label        string                                                          `json:"label"`
		OpenInNewTab *bool                                                           `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTrigger `json:"trigger"`
		Type         kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsType    `json:"type"`
		Url          string                                                          `json:"url"` //nolint:revive
	} `json:"drilldowns,omitempty"`
	Filters *struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct {
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
	} `json:"filters,omitempty"`
	HideBorder *bool   `json:"hide_border,omitempty"`
	HideTitle  *bool   `json:"hide_title,omitempty"`
	Title      *string `json:"title,omitempty"`
}) {
	// On import (tfPanel == nil), populate unconditionally when at least one API field is set.
	if tfPanel == nil {
		if !syntheticsConfigHasAnyField(apiConfig) {
			// Empty config — keep block null.
			return
		}
		cfg := syntheticsFromAPIUnconditional(apiConfig)
		pm.SyntheticsStatsOverviewConfig = cfg
		return
	}

	existing := pm.SyntheticsStatsOverviewConfig

	// If prior state had no config block, preserve nil intent.
	if existing == nil {
		return
	}

	// Block exists in state — apply null-preservation per field.
	if typeutils.IsKnown(existing.Title) {
		existing.Title = types.StringPointerValue(apiConfig.Title)
	}
	if typeutils.IsKnown(existing.Description) {
		existing.Description = types.StringPointerValue(apiConfig.Description)
	}
	if typeutils.IsKnown(existing.HideTitle) {
		existing.HideTitle = types.BoolPointerValue(apiConfig.HideTitle)
	}
	if typeutils.IsKnown(existing.HideBorder) {
		existing.HideBorder = types.BoolPointerValue(apiConfig.HideBorder)
	}

	existing.Drilldowns = readSyntheticsStatsOverviewDrilldownsFromAPI(apiConfig.Drilldowns, existing.Drilldowns)
	existing.Filters = readSyntheticsStatsOverviewFiltersFromAPI(apiConfig.Filters, existing.Filters)
}

// syntheticsConfigHasAnyField returns true if the API config has at least one non-nil field.
func syntheticsConfigHasAnyField(apiConfig struct {
	Description *string `json:"description,omitempty"`
	Drilldowns  *[]struct {
		EncodeUrl    *bool                                                           `json:"encode_url,omitempty"`    //nolint:revive
		Label        string                                                          `json:"label"`
		OpenInNewTab *bool                                                           `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTrigger `json:"trigger"`
		Type         kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsType    `json:"type"`
		Url          string                                                          `json:"url"` //nolint:revive
	} `json:"drilldowns,omitempty"`
	Filters *struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct {
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
	} `json:"filters,omitempty"`
	HideBorder *bool   `json:"hide_border,omitempty"`
	HideTitle  *bool   `json:"hide_title,omitempty"`
	Title      *string `json:"title,omitempty"`
}) bool {
	return apiConfig.Title != nil ||
		apiConfig.Description != nil ||
		apiConfig.HideTitle != nil ||
		apiConfig.HideBorder != nil ||
		(apiConfig.Drilldowns != nil && len(*apiConfig.Drilldowns) > 0) ||
		(apiConfig.Filters != nil && syntheticsFiltersHasAnyEntry(apiConfig.Filters))
}

// syntheticsFiltersHasAnyEntry returns true if the filters object has at least one non-empty category.
func syntheticsFiltersHasAnyEntry(f *struct {
	Locations *[]struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"locations,omitempty"`
	MonitorIds *[]struct {
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

// syntheticsFromAPIUnconditional populates a config model from the API without null-preservation.
// Used on import when there is no prior TF state.
func syntheticsFromAPIUnconditional(apiConfig struct {
	Description *string `json:"description,omitempty"`
	Drilldowns  *[]struct {
		EncodeUrl    *bool                                                           `json:"encode_url,omitempty"`    //nolint:revive
		Label        string                                                          `json:"label"`
		OpenInNewTab *bool                                                           `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTrigger `json:"trigger"`
		Type         kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsType    `json:"type"`
		Url          string                                                          `json:"url"` //nolint:revive
	} `json:"drilldowns,omitempty"`
	Filters *struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct {
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
	} `json:"filters,omitempty"`
	HideBorder *bool   `json:"hide_border,omitempty"`
	HideTitle  *bool   `json:"hide_title,omitempty"`
	Title      *string `json:"title,omitempty"`
}) *syntheticsStatsOverviewConfigModel {
	cfg := &syntheticsStatsOverviewConfigModel{
		Title:       types.StringPointerValue(apiConfig.Title),
		Description: types.StringPointerValue(apiConfig.Description),
		HideTitle:   types.BoolPointerValue(apiConfig.HideTitle),
		HideBorder:  types.BoolPointerValue(apiConfig.HideBorder),
	}
	cfg.Drilldowns = readSyntheticsStatsOverviewDrilldownsFromAPI(apiConfig.Drilldowns, nil)
	cfg.Filters = readSyntheticsStatsOverviewFiltersFromAPI(apiConfig.Filters, nil)
	return cfg
}

// readSyntheticsStatsOverviewDrilldownsFromAPI converts API drilldowns to TF models.
// priorDrilldowns is the existing TF state (may be nil on import).
// Optional bool fields (encode_url, open_in_new_tab) use null-preservation when prior state is available.
func readSyntheticsStatsOverviewDrilldownsFromAPI(
	apiDrilldowns *[]struct {
		EncodeUrl    *bool                                                           `json:"encode_url,omitempty"`    //nolint:revive
		Label        string                                                          `json:"label"`
		OpenInNewTab *bool                                                           `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsTrigger `json:"trigger"`
		Type         kbapi.KbnDashboardPanelSyntheticsStatsOverviewConfigDrilldownsType    `json:"type"`
		Url          string                                                          `json:"url"` //nolint:revive
	},
	priorDrilldowns []syntheticsStatsOverviewDrilldownModel,
) []syntheticsStatsOverviewDrilldownModel {
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
	apiFilters *struct {
		Locations *[]struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"locations,omitempty"`
		MonitorIds *[]struct {
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
	},
	priorFilters *syntheticsStatsOverviewFiltersModel,
) *syntheticsStatsOverviewFiltersModel {
	// Treat nil or empty filters as absent block.
	if !syntheticsFiltersHasAnyEntry(apiFilters) {
		// If prior state had a filters block, preserve it (avoids drift from API not returning filters).
		if priorFilters != nil {
			return priorFilters
		}
		return nil
	}

	f := &syntheticsStatsOverviewFiltersModel{}

	if apiFilters.Projects != nil && len(*apiFilters.Projects) > 0 {
		f.Projects = make([]syntheticsFilterItemModel, len(*apiFilters.Projects))
		for i, p := range *apiFilters.Projects {
			f.Projects[i] = syntheticsFilterItemModel{
				Label: types.StringValue(p.Label),
				Value: types.StringValue(p.Value),
			}
		}
	}

	if apiFilters.Tags != nil && len(*apiFilters.Tags) > 0 {
		f.Tags = make([]syntheticsFilterItemModel, len(*apiFilters.Tags))
		for i, t := range *apiFilters.Tags {
			f.Tags[i] = syntheticsFilterItemModel{
				Label: types.StringValue(t.Label),
				Value: types.StringValue(t.Value),
			}
		}
	}

	if apiFilters.MonitorIds != nil && len(*apiFilters.MonitorIds) > 0 {
		f.MonitorIDs = make([]syntheticsFilterItemModel, len(*apiFilters.MonitorIds))
		for i, m := range *apiFilters.MonitorIds {
			f.MonitorIDs[i] = syntheticsFilterItemModel{
				Label: types.StringValue(m.Label),
				Value: types.StringValue(m.Value),
			}
		}
	}

	if apiFilters.Locations != nil && len(*apiFilters.Locations) > 0 {
		f.Locations = make([]syntheticsFilterItemModel, len(*apiFilters.Locations))
		for i, l := range *apiFilters.Locations {
			f.Locations[i] = syntheticsFilterItemModel{
				Label: types.StringValue(l.Label),
				Value: types.StringValue(l.Value),
			}
		}
	}

	if apiFilters.MonitorTypes != nil && len(*apiFilters.MonitorTypes) > 0 {
		f.MonitorTypes = make([]syntheticsFilterItemModel, len(*apiFilters.MonitorTypes))
		for i, mt := range *apiFilters.MonitorTypes {
			f.MonitorTypes[i] = syntheticsFilterItemModel{
				Label: types.StringValue(mt.Label),
				Value: types.StringValue(mt.Value),
			}
		}
	}

	return f
}
