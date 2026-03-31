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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type lensVisualizationConverter interface {
	vizType() string
	handlesTFConfig(pm panelModel) bool
	populateFromAttributes(ctx context.Context, pm *panelModel, attrs kbapi.LensApiState) diag.Diagnostics
	buildAttributes(pm panelModel) (kbapi.LensApiState, diag.Diagnostics)
}

type lensVisualizationBase struct {
	visualizationType string
	hasTFPanelConfig  func(pm panelModel) bool
}

func (c lensVisualizationBase) vizType() string {
	return c.visualizationType
}

func (c lensVisualizationBase) handlesTFConfig(pm panelModel) bool {
	if c.hasTFPanelConfig == nil {
		return false
	}
	return c.hasTFPanelConfig(pm)
}

func detectLensVizType(attrs kbapi.LensApiState) string {
	chart, err := attrs.AsXyChart()
	if err != nil {
		return ""
	}
	return string(chart.Type)
}
