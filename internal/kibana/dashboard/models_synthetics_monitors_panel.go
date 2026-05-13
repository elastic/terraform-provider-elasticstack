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

// syntheticsMonitorsConfigModel is the Terraform model for the synthetics_monitors_config block.
// All fields are optional; the block itself may be omitted for a bare panel with no filtering.
type syntheticsMonitorsConfigModel struct {
	Title       types.String                    `tfsdk:"title"`
	Description types.String                    `tfsdk:"description"`
	HideTitle   types.Bool                      `tfsdk:"hide_title"`
	HideBorder  types.Bool                      `tfsdk:"hide_border"`
	View        types.String                    `tfsdk:"view"`
	Filters     *syntheticsMonitorsFiltersModel `tfsdk:"filters"`
}

// syntheticsMonitorsFiltersModel holds the optional filter dimensions for a
// Synthetics monitors panel (projects, tags, monitor_ids, locations, monitor_types).
// Each dimension is a list of { label, value } pairs.
type syntheticsMonitorsFiltersModel struct {
	Projects     []syntheticsFilterItemModel `tfsdk:"projects"`
	Tags         []syntheticsFilterItemModel `tfsdk:"tags"`
	MonitorIDs   []syntheticsFilterItemModel `tfsdk:"monitor_ids"`
	Locations    []syntheticsFilterItemModel `tfsdk:"locations"`
	MonitorTypes []syntheticsFilterItemModel `tfsdk:"monitor_types"`
}

// syntheticsFilterItemModel is a single { label, value } filter entry.
type syntheticsFilterItemModel struct {
	Label types.String `tfsdk:"label"`
	Value types.String `tfsdk:"value"`
}

// buildSyntheticsMonitorsPanel converts the Terraform panel model to the Kibana API panel struct.
func buildSyntheticsMonitorsPanel(pm panelModel, grid struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
}, id *string) kbapi.KbnDashboardPanelTypeSyntheticsMonitors {
	panel := kbapi.KbnDashboardPanelTypeSyntheticsMonitors{
		Grid: kbapi.KbnDashboardPanelGrid{
			H: grid.H,
			W: grid.W,
			X: grid.X,
			Y: grid.Y,
		},
		Type: kbapi.SyntheticsMonitors,
		Id:   id,
	}

	cfg := pm.SyntheticsMonitorsConfig
	if cfg == nil {
		// No config configured — emit an empty config object (valid per API schema).
		return panel
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
		return panel
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

	return panel
}

// ensureSyntheticsAPIFilters initialises the Config.Filters pointer if it is nil.
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

// toSyntheticsFilterItems converts a slice of TF filter items to the anonymous API struct slice.
func toSyntheticsFilterItems(items []syntheticsFilterItemModel) []struct {
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

// populateSyntheticsMonitorsFromAPI reads back the Kibana API panel and updates the panel model.
// Implements null-preservation: when the prior TF state omitted the config block entirely, this
// call is a no-op to preserve practitioner intent.
//
// apiPanel is the panel returned from the API. tfPanel is the prior TF state/plan panel, or nil
// on import.
func populateSyntheticsMonitorsFromAPI(pm *panelModel, tfPanel *panelModel, apiPanel kbapi.KbnDashboardPanelTypeSyntheticsMonitors) {
	apiFilters := apiPanel.Config.Filters

	// On import (tfPanel == nil), populate config from API unconditionally.
	if tfPanel == nil {
		filters := fromSyntheticsAPIFilters(apiFilters)
		if apiPanel.Config.Title == nil &&
			apiPanel.Config.Description == nil &&
			apiPanel.Config.HideTitle == nil &&
			apiPanel.Config.HideBorder == nil &&
			apiPanel.Config.View == nil &&
			filters == nil {
			// API returned no meaningful config — keep config block null on import.
			return
		}
		pm.SyntheticsMonitorsConfig = &syntheticsMonitorsConfigModel{
			Title:       types.StringPointerValue(apiPanel.Config.Title),
			Description: types.StringPointerValue(apiPanel.Config.Description),
			HideTitle:   types.BoolPointerValue(apiPanel.Config.HideTitle),
			HideBorder:  types.BoolPointerValue(apiPanel.Config.HideBorder),
			View:        syntheticsMonitorsViewValue(apiPanel.Config.View),
			Filters:     filters,
		}
		return
	}

	existing := pm.SyntheticsMonitorsConfig

	// Prior state had no config block — preserve nil intent.
	if existing == nil {
		return
	}

	// Config block exists in state — update known attributes while preserving omitted/null intent.
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
		// API returned no filters; preserve prior filters block intent.
		return
	}

	filters := fromSyntheticsAPIFilters(apiFilters)
	if filters == nil {
		// API returned empty or absent filters.
		if existing.Filters == nil {
			// Both prior state and API have no meaningful filter content — keep null.
			return
		}
		// Prior state had an explicit (possibly empty) filters block.
		// Preserve the empty block to avoid a perpetual diff for `filters = {}`.
		return
	}
	existing.Filters = filters
}

// fromSyntheticsAPIFilters converts the API filters struct to the TF model.
// Returns nil when all filter dimensions are absent or empty (null-preservation).
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
}) *syntheticsMonitorsFiltersModel {
	if apiFilters == nil {
		return nil
	}

	projects := fromSyntheticsAPIItems(apiFilters.Projects)
	tags := fromSyntheticsAPIItems(apiFilters.Tags)
	monitorIDs := fromSyntheticsAPIItems(apiFilters.MonitorIds)
	locations := fromSyntheticsAPIItems(apiFilters.Locations)
	monitorTypes := fromSyntheticsAPIItems(apiFilters.MonitorTypes)

	// If all dimensions are nil (empty or absent), treat filters as null.
	if projects == nil && tags == nil && monitorIDs == nil && locations == nil && monitorTypes == nil {
		return nil
	}

	return &syntheticsMonitorsFiltersModel{
		Projects:     projects,
		Tags:         tags,
		MonitorIDs:   monitorIDs,
		Locations:    locations,
		MonitorTypes: monitorTypes,
	}
}

// fromSyntheticsAPIItems converts a pointer to an API filter item slice to a TF model slice.
// Returns nil when the slice is absent or empty.
func fromSyntheticsAPIItems(items *[]struct {
	Label string `json:"label"`
	Value string `json:"value"`
}) []syntheticsFilterItemModel {
	if items == nil || len(*items) == 0 {
		return nil
	}
	result := make([]syntheticsFilterItemModel, len(*items))
	for i, item := range *items {
		result[i] = syntheticsFilterItemModel{
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
