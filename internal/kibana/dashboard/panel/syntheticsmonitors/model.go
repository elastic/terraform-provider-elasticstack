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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes Terraform monitors panel config onto the typed API panel (Grid/Id/Type must be set by the Handler).
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsMonitors) diag.Diagnostics {
	cfg := pm.SyntheticsMonitorsConfig
	if cfg == nil {
		return nil
	}

	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&panel.Config.Title, &panel.Config.Description, &panel.Config.HideTitle, &panel.Config.HideBorder)
	if typeutils.IsKnown(cfg.View) {
		view := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsMonitorsConfigView(cfg.View.ValueString())
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

// PopulateFromAPI maps the Kibana API panel config into Terraform panel state while preserving
// prior null intent (REQ-009). prior is the prior TF state/plan panel, or nil on import.
//
// pm always arrives with SyntheticsMonitorsConfig unset (callers build state from a zero-valued
// PanelModel to avoid aliasing plan pointers), so that field cannot be used to detect whether this
// panel was previously this same type. prior.SyntheticsMonitorsConfig is the only reliable signal:
// non-nil means the panel was already this type with a configured block and its null intent must
// be honored. Unlike most other panel config blocks, this entire block is optional even when the
// panel type matches, so prior.SyntheticsMonitorsConfig == nil is ambiguous (a type change, or a
// practitioner who simply never configured this block) and either way there is no null intent to
// honor and nothing should be conjured from API-only data.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsMonitors) diag.Diagnostics {
	if prior == nil {
		filters := fromSyntheticsAPIFilters(apiPanel.Config.Filters)
		if apiPanel.Config.Title == nil &&
			apiPanel.Config.Description == nil &&
			apiPanel.Config.HideTitle == nil &&
			apiPanel.Config.HideBorder == nil &&
			apiPanel.Config.View == nil &&
			filters == nil {
			return nil
		}
		pm.SyntheticsMonitorsConfig = syntheticsMonitorsConfigFromAPIImport(apiPanel)
		return nil
	}

	if prior.SyntheticsMonitorsConfig == nil {
		return nil
	}

	// Same-type update: rebuild from the API, then reapply the prior config's null intent for any
	// optional field the plan/state had not set (REQ-009 null-preservation).
	existing := syntheticsMonitorsConfigFromAPIImport(apiPanel)
	syntheticsMonitorsPreserveNullIntentFromPrior(prior.SyntheticsMonitorsConfig, existing)
	pm.SyntheticsMonitorsConfig = existing
	return nil
}

func syntheticsMonitorsPreserveNullIntentFromPrior(prior, existing *models.SyntheticsMonitorsConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	panelkit.NullPreservePresentationFromPrior(prior.Title, prior.Description, prior.HideTitle, prior.HideBorder,
		&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder)
	if !typeutils.IsKnown(prior.View) {
		existing.View = types.StringNull()
	}
	// The API config's filters sub-object is entirely optional; if the freshly imported filters
	// came back nil (nothing to show) but the practitioner had explicitly configured a (possibly
	// empty) filters block, keep that block rather than dropping it to avoid a perpetual diff.
	if existing.Filters == nil && prior.Filters != nil {
		existing.Filters = prior.Filters
	}
}

func syntheticsMonitorsConfigFromAPIImport(apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsMonitors) *models.SyntheticsMonitorsConfigModel {
	return &models.SyntheticsMonitorsConfigModel{
		Title:       types.StringPointerValue(apiPanel.Config.Title),
		Description: types.StringPointerValue(apiPanel.Config.Description),
		HideTitle:   types.BoolPointerValue(apiPanel.Config.HideTitle),
		HideBorder:  types.BoolPointerValue(apiPanel.Config.HideBorder),
		View:        syntheticsMonitorsViewValue(apiPanel.Config.View),
		Filters:     fromSyntheticsAPIFilters(apiPanel.Config.Filters),
	}
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

func syntheticsMonitorsViewValue(view *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSyntheticsMonitorsConfigView) types.String {
	if view == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*view))
}
