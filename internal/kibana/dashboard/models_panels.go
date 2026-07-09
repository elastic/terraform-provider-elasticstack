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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func dashboardMapPanelsFromAPI(ctx context.Context, m *models.DashboardModel, apiPanels *kbapi.DashboardPanels) ([]models.PanelModel, []models.SectionModel, diag.Diagnostics) {
	ctx = iface.WithEnclosingDashboard(ctx, m)
	var diags diag.Diagnostics
	if apiPanels == nil || len(*apiPanels) == 0 {
		return nil, nil, diags
	}

	var panels []models.PanelModel
	var sections []models.SectionModel

	for _, item := range *apiPanels {
		// Try section first; this avoids treating section items as panels.
		section, err := item.AsKibanaHTTPAPIsKbnDashboardSection()
		if err == nil && section.Title != "" {
			tfSectionIndex := len(sections)
			var tfSection *models.SectionModel
			if tfSectionIndex < len(m.Sections) {
				tfSection = &m.Sections[tfSectionIndex]
			}
			sectionModel, d := dashboardMapSectionFromAPI(ctx, m, tfSection, section)
			diags.Append(d...)
			if diags.HasError() {
				return nil, nil, diags
			}
			sections = append(sections, sectionModel)
			continue
		}

		panelItem, err := item.AsDashboardPanelItem()
		if err == nil {
			tfPanelIndex := len(panels)
			var tfPanel *models.PanelModel
			if tfPanelIndex < len(m.Panels) {
				tfPanel = &m.Panels[tfPanelIndex]
			}

			panel, d := dashboardMapPanelFromAPI(ctx, m, tfPanel, panelItem)
			diags.Append(d...)
			if diags.HasError() {
				return nil, nil, diags
			}

			panels = append(panels, panel)
		}
	}

	return panels, sections, diags
}

func dashboardMapSectionFromAPI(ctx context.Context, m *models.DashboardModel, tfSection *models.SectionModel, section Section) (models.SectionModel, diag.Diagnostics) {
	collapsed := types.BoolPointerValue(section.Collapsed)
	if tfSection != nil && !typeutils.IsKnown(tfSection.Collapsed) && section.Collapsed != nil && !*section.Collapsed {
		collapsed = types.BoolNull()
	}

	sm := models.SectionModel{
		Title:     types.StringValue(section.Title),
		Collapsed: collapsed,
		ID:        types.StringPointerValue(section.Id),
		Grid: models.SectionGridModel{
			Y: types.Int64Value(int64(section.Grid.Y)),
		},
	}

	// Map section panels
	var diags diag.Diagnostics
	if section.Panels != nil {
		var innerPanels []models.PanelModel
		for _, p := range *section.Panels {
			tfPanelIndex := len(innerPanels)
			var tfPanel *models.PanelModel
			if tfSection != nil && tfPanelIndex < len(tfSection.Panels) {
				tfPanel = &tfSection.Panels[tfPanelIndex]
			}

			pm, d := dashboardMapPanelFromAPI(ctx, m, tfPanel, p)
			diags.Append(d...)
			if diags.HasError() {
				return models.SectionModel{}, diags
			}

			innerPanels = append(innerPanels, pm)
		}
		sm.Panels = innerPanels
	}
	return sm, diags
}

func clearPanelConfigBlocks(pm *models.PanelModel) {
	pm.MarkdownConfig = nil
	pm.TimeSliderControlConfig = nil
	pm.SloBurnRateConfig = nil
	pm.SloOverviewConfig = nil
	pm.SloErrorBudgetConfig = nil
	pm.EsqlControlConfig = nil
	pm.OptionsListControlConfig = nil
	pm.RangeSliderControlConfig = nil
	pm.SyntheticsStatsOverviewConfig = nil
	pm.SyntheticsMonitorsConfig = nil
	pm.VisConfig = nil
	pm.ImageConfig = nil
	pm.SloAlertsConfig = nil
	pm.DiscoverSessionConfig = nil
	pm.LinksConfig = nil
	pm.FieldStatsTableConfig = nil
	pm.MlAnomalySwimlaneConfig = nil
	pm.MlAnomalyChartsConfig = nil
	pm.MlSingleMetricViewerConfig = nil
	pm.ApmServiceMapConfig = nil
	pm.FieldStatsTableConfig = nil
}

func dashboardMapPanelFromAPI(ctx context.Context, _ *models.DashboardModel, tfPanel *models.PanelModel, panelItem kbapi.DashboardPanelItem) (models.PanelModel, diag.Diagnostics) {
	// Build state from the API response. Do not shallow-copy tfPanel into pm: nested
	// pointers (e.g. vis_config.by_value) would be shared with the plan/prior model and
	// FromAPI mutations would corrupt the plan used for post-read alignment.
	var pm models.PanelModel

	discriminator, err := panelItem.Discriminator()
	if err != nil {
		return models.PanelModel{}, diagutil.FrameworkDiagFromError(err)
	}
	pm.Type = types.StringValue(discriminator)

	var diags diag.Diagnostics
	if h := LookupHandler(discriminator); h != nil {
		diags = h.FromAPI(ctx, &pm, tfPanel, panelItem)
		alignPanelStateFromPlan(ctx, tfPanel, &pm)
		return pm, diags
	}

	unknownPanelFromAPI(ctx, tfPanel, &pm, panelItem)
	alignPanelStateFromPlan(ctx, tfPanel, &pm)
	return pm, diags
}

func dashboardPanelsToAPI(ctx context.Context, m *models.DashboardModel) (*kbapi.DashboardPanels, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.Panels == nil && m.Sections == nil {
		return nil, diags
	}

	apiPanels := make(kbapi.DashboardPanels, 0, len(m.Panels)+len(m.Sections))

	// Process panels
	for _, pm := range m.Panels {
		panelItem, d := panelToAPI(ctx, pm, m)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var item kbapi.DashboardPanels_Item
		err := item.FromDashboardPanelItem(panelItem)
		if err != nil {
			diags.AddError("Failed to create dashboard panel item", err.Error())
		}

		apiPanels = append(apiPanels, item)
	}

	// Process sections
	for _, sm := range m.Sections {
		section := Section{
			Title: sm.Title.ValueString(),
			Grid: struct {
				Y float32 `json:"y"`
			}{
				Y: float32(sm.Grid.Y.ValueInt64()),
			},
		}

		if typeutils.IsKnown(sm.Collapsed) {
			section.Collapsed = new(sm.Collapsed.ValueBool())
		}
		if typeutils.IsKnown(sm.ID) {
			section.Id = new(sm.ID.ValueString())
		}

		if len(sm.Panels) > 0 {
			innerPanels := make([]kbapi.DashboardPanelItem, 0, len(sm.Panels))

			for _, pm := range sm.Panels {
				item, d := panelToAPI(ctx, pm, m)
				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}

				innerPanels = append(innerPanels, item)
			}
			section.Panels = &innerPanels
		}

		var item kbapi.DashboardPanels_Item
		err := item.FromKibanaHTTPAPIsKbnDashboardSection(section)
		if err != nil {
			diags.AddError("Failed to create dashboard section item", err.Error())
		}
		apiPanels = append(apiPanels, item)
	}

	return &apiPanels, diags
}

func panelToAPI(ctx context.Context, pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	if typeutils.IsKnown(pm.Type) {
		pt := pm.Type.ValueString()
		if h := LookupHandler(pt); h != nil {
			return h.ToAPI(pm, dashboard)
		}
	}

	for _, h := range AllHandlers() {
		if panelkit.HasConfig(&pm, h.PanelType()+"_config") {
			return h.ToAPI(pm, dashboard)
		}
	}

	return fallbackPanelToAPI(ctx, pm, dashboard)
}

func fallbackPanelToAPI(ctx context.Context, pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = ctx
	_ = dashboard

	var diags diag.Diagnostics

	grid := struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	}{
		X: float32(pm.Grid.X.ValueInt64()),
		Y: float32(pm.Grid.Y.ValueInt64()),
	}
	if typeutils.IsKnown(pm.Grid.W) {
		w := float32(pm.Grid.W.ValueInt64())
		grid.W = &w
	}
	if typeutils.IsKnown(pm.Grid.H) {
		h := float32(pm.Grid.H.ValueInt64())
		grid.H = &h
	}

	var panelID *string
	if typeutils.IsKnown(pm.ID) {
		panelID = new(pm.ID.ValueString())
	}

	var panelItem kbapi.DashboardPanelItem

	if typeutils.IsKnown(pm.ConfigJSON) && !pm.ConfigJSON.IsNull() {
		configJSON := []byte(pm.ConfigJSON.ValueString())
		fullPanel := map[string]any{
			attrPanelType: pm.Type.ValueString(),
			attrPanelGrid: grid,
			"config":      json.RawMessage(configJSON),
		}
		if panelID != nil {
			fullPanel[attrPanelID] = *panelID
		}
		rawBytes, mErr := json.Marshal(fullPanel)
		if mErr != nil {
			diags.AddError("Failed to marshal unknown panel", mErr.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		if err := panelItem.UnmarshalJSON(rawBytes); err != nil {
			diags.AddError("Failed to create unknown panel type", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		return panelItem, diags
	}

	if !typeutils.IsKnown(pm.Type) {
		diags.AddError("Unsupported panel configuration", "No panel configuration block was provided.")
		return kbapi.DashboardPanelItem{}, diags
	}

	panelType := pm.Type.ValueString()
	diags.AddError(
		"Unsupported panel type",
		fmt.Sprintf(
			"Panel type %q is not yet supported. This panel type was preserved from the API during read "+
				"but cannot be authored in configuration. To add support for this panel type, "+
				"wait for a provider update that includes a typed configuration block.",
			panelType,
		),
	)
	return kbapi.DashboardPanelItem{}, diags
}
