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

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
)

// alignDashboardStateFromPlanPanels preserves practitioner intent that depends on
// the full top-level panel slice. Common per-panel alignment already happens
// inside mapPanelFromAPI.
func alignDashboardStateFromPlanPanels(planPanels, statePanels []models.PanelModel) {
	lenscommon.ApplySliceAligners(planPanels, statePanels)
}

// suppressReadTopLevelPanelsWhenPlanEmpty clears echoed top-level panels on read/create/update
// when the practitioner set `panels = []` (explicit empty list). A nil plan slice means the
// attribute was omitted and must remain null in state, not an empty list.
func suppressReadTopLevelPanelsWhenPlanEmpty(planPanels []models.PanelModel, readModel *models.DashboardModel) {
	if readModel == nil || planPanels == nil || len(planPanels) != 0 {
		return
	}
	readModel.Panels = []models.PanelModel{}
}

func alignDashboardStateFromPlanSections(ctx context.Context, planSections, stateSections []models.SectionModel) {
	n := min(len(planSections), len(stateSections))
	for i := range n {
		alignDashboardStateFromPlanPanels(planSections[i].Panels, stateSections[i].Panels)
		panelCount := min(len(planSections[i].Panels), len(stateSections[i].Panels))
		for j := range panelCount {
			alignPanelStateFromPlan(ctx, &planSections[i].Panels[j], &stateSections[i].Panels[j])
		}
	}
}

func alignDashboardStateFromPlanPinnedPanels(ctx context.Context, planPins, statePins []models.PinnedPanelModel) {
	n := min(len(planPins), len(statePins))
	for i := range n {
		plan := planPins[i].SyntheticPanel()
		state := statePins[i].SyntheticPanel()
		alignPanelStateFromPlan(ctx, &plan, &state)
	}
}

func alignPanelStateFromPlan(ctx context.Context, plan, state *models.PanelModel) {
	if plan == nil || state == nil {
		return
	}

	lenscommon.PreservePlanJSONIfStateOmitsOptionalKeys(plan.ConfigJSON.Normalized, &state.ConfigJSON.Normalized, "filters", "query", "settings")
	planBlocks := visByValueChartBlocksFromPanel(plan)
	stateBlocks := visByValueChartBlocksFromPanel(state)
	if planBlocks != nil && stateBlocks != nil {
		for _, c := range lenscommon.All() {
			c.AlignStateFromPlan(ctx, planBlocks, stateBlocks)
		}
	}
	if h := LookupHandler(state.Type.ValueString()); h != nil {
		h.AlignStateFromPlan(ctx, plan, state)
	}
}

func visByValueChartBlocksFromPanel(pm *models.PanelModel) *models.LensByValueChartBlocks {
	if pm == nil || pm.VisConfig == nil || pm.VisConfig.ByValue == nil {
		return nil
	}
	return &pm.VisConfig.ByValue.LensByValueChartBlocks
}
