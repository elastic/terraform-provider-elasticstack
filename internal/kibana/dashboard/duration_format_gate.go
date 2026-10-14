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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var durationFormatParamsRequirement = fmt.Sprintf(
	`format.type "duration" with "decimals"/"compact" requires Elastic Stack %s or later; remove these properties or upgrade Kibana.`,
	panelkit.MinKibanaAPISupportDurationFormatParams.Original(),
)

func enforceDurationFormatParamsVersion(ctx context.Context, client *clients.KibanaScopedClient, plan models.DashboardModel) diag.Diagnostics {
	location, found := firstDurationFormatParams(plan)
	if !found {
		return nil
	}
	supported, verDiags := client.EnforceMinVersion(ctx, panelkit.MinKibanaAPISupportDurationFormatParams)
	if verDiags.HasError() {
		return verDiags
	}
	if supported {
		return nil
	}
	var diags diag.Diagnostics
	diags.AddError(
		"Unsupported Lens XY duration format",
		location+": "+durationFormatParamsRequirement,
	)
	return diags
}

func firstDurationFormatParams(plan models.DashboardModel) (string, bool) {
	if location, found := firstDurationFormatParamsInPanels(plan.Panels, "panels"); found {
		return location, true
	}
	for sectionIdx, section := range plan.Sections {
		prefix := fmt.Sprintf("sections[%d].panels", sectionIdx)
		if location, found := firstDurationFormatParamsInPanels(section.Panels, prefix); found {
			return location, true
		}
	}
	return "", false
}

func firstDurationFormatParamsInPanels(panels []models.PanelModel, pathPrefix string) (string, bool) {
	for panelIdx, panel := range panels {
		blocks := visByValueChartBlocksFromPanel(&panel)
		if blocks == nil || blocks.XYChartConfig == nil {
			continue
		}
		chart := blocks.XYChartConfig
		for layerIdx, layer := range chart.Layers {
			if layer.DataLayer == nil {
				continue
			}
			for metricIdx, metric := range layer.DataLayer.Y {
				if !metricUsesUnsupportedDurationParams(metric.ConfigJSON) {
					continue
				}
				label := fmt.Sprintf("%s[%d]", pathPrefix, panelIdx)
				if panel.ID.ValueString() != "" {
					label += fmt.Sprintf(" (id %q)", panel.ID.ValueString())
				}
				return fmt.Sprintf("%s layer %d metric %d", label, layerIdx, metricIdx), true
			}
		}
	}
	return "", false
}

func metricUsesUnsupportedDurationParams(config jsontypes.Normalized) bool {
	if !typeutils.IsKnown(config) || config.ValueString() == "" {
		return false
	}
	var model map[string]any
	if err := json.Unmarshal([]byte(config.ValueString()), &model); err != nil {
		return false
	}
	return lenscommon.MetricFormatUsesUnsupportedDurationParams(model)
}
