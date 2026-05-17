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

package visconfig

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
)

// lensConfigClass identifies how a vis panel API config JSON should be interpreted before trusting union helpers alone.
type lensConfigClass int

const (
	// lensConfigClassByValueChart means the payload has a non-empty string at top-level "type" (Lens chart discriminator).
	lensConfigClassByValueChart lensConfigClass = iota
	// lensConfigClassByReference means ref_id plus time_range.from/to (by-reference saved-object linkage shape).
	lensConfigClassByReference
	// lensConfigClassAmbiguous means neither a chart payload nor a complete by-reference shape.
	lensConfigClassAmbiguous
)

func classifyLensConfigFromRoot(root map[string]any) lensConfigClass {
	switch {
	case hasLensByValueChartTypeAtRoot(root):
		return lensConfigClassByValueChart
	case lenscommon.HasLensByReferenceShapeAtRoot(root):
		return lensConfigClassByReference
	default:
		return lensConfigClassAmbiguous
	}
}

func hasLensByValueChartTypeAtRoot(m map[string]any) bool {
	if m == nil {
		return false
	}
	v, ok := m["type"]
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s != ""
}

func panelHasTypedConfig(pm *models.PanelModel) bool {
	if pm == nil {
		return false
	}
	return pm.MarkdownConfig != nil ||
		pm.TimeSliderControlConfig != nil ||
		pm.SloBurnRateConfig != nil ||
		pm.SloOverviewConfig != nil ||
		pm.SloErrorBudgetConfig != nil ||
		pm.EsqlControlConfig != nil ||
		pm.OptionsListControlConfig != nil ||
		pm.RangeSliderControlConfig != nil ||
		pm.SyntheticsStatsOverviewConfig != nil ||
		pm.SyntheticsMonitorsConfig != nil ||
		pm.LensDashboardAppConfig != nil ||
		pm.VisConfig != nil ||
		pm.ImageConfig != nil ||
		pm.SloAlertsConfig != nil ||
		pm.DiscoverSessionConfig != nil
}

func priorPanelUsesConfigJSONOnly(prior *models.PanelModel) bool {
	if prior == nil || !typeutils.IsKnown(prior.ConfigJSON) {
		return false
	}
	return !panelHasTypedConfig(prior)
}

func configPriorForVisRead(tfPanel, pm *models.PanelModel) *models.VisConfigModel {
	if tfPanel != nil && tfPanel.VisConfig != nil {
		return tfPanel.VisConfig
	}
	if pm != nil && pm.VisConfig != nil {
		return pm.VisConfig
	}
	return nil
}
