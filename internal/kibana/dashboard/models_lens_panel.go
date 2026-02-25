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

import "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"

type lensPanelConfigConverter struct {
	visualizationType string
	hasTFPanelConfig  func(pm panelModel) bool
}

func (c lensPanelConfigConverter) handlesAPIPanelConfig(pm *panelModel, panelType string, cfg kbapi.DashboardPanelItem_Config) bool {
	if c.hasTFPanelConfig != nil && pm != nil && !c.hasTFPanelConfig(*pm) {
		return false
	}

	if panelType != "lens" {
		return false
	}

	cfgMap, err := cfg.AsDashboardPanelItemConfig2()
	if err != nil {
		return false
	}

	return c.hasExpectedVisualizationType(cfgMap)
}

func (c lensPanelConfigConverter) hasExpectedVisualizationType(cfgMap map[string]any) bool {
	attrs, ok := cfgMap["attributes"]
	if !ok {
		return false
	}

	attrsMap, ok := attrs.(map[string]any)
	if !ok {
		return false
	}

	vizType, ok := attrsMap["type"]
	if !ok {
		return false
	}

	return vizType == c.visualizationType
}
