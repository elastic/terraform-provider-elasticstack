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

package syntheticsmonitors

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes Terraform monitors panel config onto the typed API panel (Grid/Id/Type must be set by the Handler).
func BuildConfig(pm models.PanelModel, panel *kbapi.KbnDashboardPanelTypeSyntheticsMonitors) diag.Diagnostics {
	cfg := pm.SyntheticsMonitorsConfig
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
	if typeutils.IsKnown(cfg.View) {
		view := kbapi.KbnDashboardPanelTypeSyntheticsMonitorsConfigView(cfg.View.ValueString())
		panel.Config.View = &view
	}

	if cfg.Filters == nil {
		return nil
	}

	if len(cfg.Filters.Projects) > 0 {
		items := toSyntheticsFilterItems(cfg.Filters.Projects)
		panel.Config.Filters = ensureSyntheticsAPIFilters(panel.Config.Filters)
		panel.Config.Filters.Projects = &items
	}
	if len(cfg.Filters.Tags) > 0 {
		items := toSyntheticsFilterItems(cfg.Filters.Tags)
		panel.Config.Filters = ensureSyntheticsAPIFilters(panel.Config.Filters)
		panel.Config.Filters.Tags = &items
	}
	if len(cfg.Filters.MonitorIDs) > 0 {
		items := toSyntheticsFilterItems(cfg.Filters.MonitorIDs)
		panel.Config.Filters = ensureSyntheticsAPIFilters(panel.Config.Filters)
		panel.Config.Filters.MonitorIds = &items
	}
	if len(cfg.Filters.Locations) > 0 {
		items := toSyntheticsFilterItems(cfg.Filters.Locations)
		panel.Config.Filters = ensureSyntheticsAPIFilters(panel.Config.Filters)
		panel.Config.Filters.Locations = &items
	}
	if len(cfg.Filters.MonitorTypes) > 0 {
		items := toSyntheticsFilterItems(cfg.Filters.MonitorTypes)
		panel.Config.Filters = ensureSyntheticsAPIFilters(panel.Config.Filters)
		panel.Config.Filters.MonitorTypes = &items
	}

	return nil
}

func ensureSyntheticsAPIFilters(f *struct {
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
}) *struct {
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
} {
	if f != nil {
		return f
	}
	return &struct {
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
	}{}
}

func toSyntheticsFilterItems(items []models.SyntheticsFilterItemModel) []struct {
	Label string `json:"label"`
	Value string `json:"value"`
} {
	result := make([]struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}, len(items))
	for i, item := range items {
		result[i].Label = item.Label.ValueString()
		result[i].Value = item.Value.ValueString()
	}
	return result
}

// PopulateFromAPI reads the Kibana API panel into Terraform panel state.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiPanel kbapi.KbnDashboardPanelTypeSyntheticsMonitors) diag.Diagnostics {
	apiFilters := apiPanel.Config.Filters

	if prior == nil {
		filters := fromSyntheticsAPIFilters(apiFilters)
		if apiPanel.Config.Title == nil &&
			apiPanel.Config.Description == nil &&
			apiPanel.Config.HideTitle == nil &&
			apiPanel.Config.HideBorder == nil &&
			apiPanel.Config.View == nil &&
			filters == nil {
			return nil
		}
		pm.SyntheticsMonitorsConfig = &models.SyntheticsMonitorsConfigModel{
			Title:       types.StringPointerValue(apiPanel.Config.Title),
			Description: types.StringPointerValue(apiPanel.Config.Description),
			HideTitle:   types.BoolPointerValue(apiPanel.Config.HideTitle),
			HideBorder:  types.BoolPointerValue(apiPanel.Config.HideBorder),
			View:        syntheticsMonitorsViewValue(apiPanel.Config.View),
			Filters:     filters,
		}
		return nil
	}

	existing := pm.SyntheticsMonitorsConfig

	if existing == nil {
		return nil
	}

	if typeutils.IsKnown(existing.Title) {
		existing.Title = types.StringPointerValue(apiPanel.Config.Title)
	}
	if typeutils.IsKnown(existing.Description) {
		existing.Description = types.StringPointerValue(apiPanel.Config.Description)
	}
	if typeutils.IsKnown(existing.HideTitle) {
		existing.HideTitle = types.BoolPointerValue(apiPanel.Config.HideTitle)
	}
	if typeutils.IsKnown(existing.HideBorder) {
		existing.HideBorder = types.BoolPointerValue(apiPanel.Config.HideBorder)
	}
	if typeutils.IsKnown(existing.View) {
		existing.View = syntheticsMonitorsViewValue(apiPanel.Config.View)
	}

	if apiFilters == nil {
		return nil
	}

	filters := fromSyntheticsAPIFilters(apiFilters)
	if filters == nil {
		if existing.Filters == nil {
			return nil
		}
		return nil
	}
	existing.Filters = filters
	return nil
}

func fromSyntheticsAPIFilters(apiFilters *struct {
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
}) *models.SyntheticsFiltersModel {
	if apiFilters == nil {
		return nil
	}

	projects := fromSyntheticsAPIItems(apiFilters.Projects)
	tags := fromSyntheticsAPIItems(apiFilters.Tags)
	monitorIDs := fromSyntheticsAPIItems(apiFilters.MonitorIds)
	locations := fromSyntheticsAPIItems(apiFilters.Locations)
	monitorTypes := fromSyntheticsAPIItems(apiFilters.MonitorTypes)

	if projects == nil && tags == nil && monitorIDs == nil && locations == nil && monitorTypes == nil {
		return nil
	}

	return &models.SyntheticsFiltersModel{
		Projects:     projects,
		Tags:         tags,
		MonitorIDs:   monitorIDs,
		Locations:    locations,
		MonitorTypes: monitorTypes,
	}
}

func fromSyntheticsAPIItems(items *[]struct {
	Label string `json:"label"`
	Value string `json:"value"`
}) []models.SyntheticsFilterItemModel {
	if items == nil || len(*items) == 0 {
		return nil
	}
	result := make([]models.SyntheticsFilterItemModel, len(*items))
	for i, item := range *items {
		result[i] = models.SyntheticsFilterItemModel{
			Label: types.StringValue(item.Label),
			Value: types.StringValue(item.Value),
		}
	}
	return result
}

func syntheticsMonitorsViewValue(view *kbapi.KbnDashboardPanelTypeSyntheticsMonitorsConfigView) types.String {
	if view == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*view))
}
