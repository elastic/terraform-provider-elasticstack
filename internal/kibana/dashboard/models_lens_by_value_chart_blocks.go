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

import "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"

// lensByValueChartBlocksFromPanel returns the typed Lens chart block wrapper for the panel's
// active by-value path: either `vis_config.by_value` or `lens_dashboard_app_config.by_value`
// (design D3/D10). Attributes remain at the by_value object root on both Terraform models.
func lensByValueChartBlocksFromPanel(pm *models.PanelModel) *models.LensByValueChartBlocks {
	if pm == nil {
		return nil
	}
	if pm.VisConfig != nil && pm.VisConfig.ByValue != nil {
		return &pm.VisConfig.ByValue.LensByValueChartBlocks
	}
	if pm.LensDashboardAppConfig != nil && pm.LensDashboardAppConfig.ByValue != nil {
		blocks, ok := lensByValueChartBlocksForTypedLensApp(*pm.LensDashboardAppConfig.ByValue)
		if !ok {
			return nil
		}
		return blocks
	}
	return nil
}

func firstLensVisConverterForChartBlocks(blocks *models.LensByValueChartBlocks) (lensVisualizationConverter, bool) {
	for _, c := range lensVisConverters {
		if c.handlesTFConfigBlocks(blocks) {
			return c, true
		}
	}
	return nil, false
}

// seedWaffleLensByValueChartFromPriorPanel assigns the waffle chart pointer from practitioner plan/state
// into dest before vis read-mapping replaces blocks.WaffleConfig. The waffle converter keeps that pointer as
// `seed` across `populateFromAttributes` so mergeWaffleConfigFromPlanSeed can reconcile Kibana read omissions.
func seedWaffleLensByValueChartFromPriorPanel(dest *models.LensByValueChartBlocks, prior *models.PanelModel) {
	if dest == nil || prior == nil || prior.VisConfig == nil || prior.VisConfig.ByValue == nil {
		return
	}
	src := &prior.VisConfig.ByValue.LensByValueChartBlocks
	if src.WaffleConfig != nil {
		dest.WaffleConfig = src.WaffleConfig
	}
}
