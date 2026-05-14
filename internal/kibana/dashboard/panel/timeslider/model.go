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

package timeslider

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PopulateFromAPI reads back time slider control config from the dashboard API response into pm (null-preserving).
func PopulateFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, apiConfig struct {
	EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
	IsAnchored                 *bool    `json:"is_anchored,omitempty"`
	StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
}) {
	existing := pm.TimeSliderControlConfig

	if tfPanel == nil {
		if apiConfig.StartPercentageOfTimeRange == nil &&
			apiConfig.EndPercentageOfTimeRange == nil &&
			apiConfig.IsAnchored == nil {
			return
		}
		pm.TimeSliderControlConfig = &models.TimeSliderControlConfigModel{}
		existing = pm.TimeSliderControlConfig
		if apiConfig.StartPercentageOfTimeRange != nil {
			existing.StartPercentageOfTimeRange = types.Float32Value(*apiConfig.StartPercentageOfTimeRange)
		}
		if apiConfig.EndPercentageOfTimeRange != nil {
			existing.EndPercentageOfTimeRange = types.Float32Value(*apiConfig.EndPercentageOfTimeRange)
		}
		if apiConfig.IsAnchored != nil {
			existing.IsAnchored = types.BoolValue(*apiConfig.IsAnchored)
		}
		return
	}

	if existing == nil {
		if apiConfig.StartPercentageOfTimeRange == nil &&
			apiConfig.EndPercentageOfTimeRange == nil &&
			apiConfig.IsAnchored == nil {
			return
		}
		return
	}

	if typeutils.IsKnown(existing.StartPercentageOfTimeRange) && apiConfig.StartPercentageOfTimeRange != nil {
		existing.StartPercentageOfTimeRange = types.Float32Value(*apiConfig.StartPercentageOfTimeRange)
	}
	if typeutils.IsKnown(existing.EndPercentageOfTimeRange) && apiConfig.EndPercentageOfTimeRange != nil {
		existing.EndPercentageOfTimeRange = types.Float32Value(*apiConfig.EndPercentageOfTimeRange)
	}
	if typeutils.IsKnown(existing.IsAnchored) && apiConfig.IsAnchored != nil {
		existing.IsAnchored = types.BoolValue(*apiConfig.IsAnchored)
	}
}

// BuildConfig writes TF fields into tsPanel.Config.
func BuildConfig(pm models.PanelModel, tsPanel *kbapi.KbnDashboardPanelTypeTimeSliderControl) {
	cfg := pm.TimeSliderControlConfig
	if cfg == nil {
		return
	}
	if typeutils.IsKnown(cfg.StartPercentageOfTimeRange) {
		v := cfg.StartPercentageOfTimeRange.ValueFloat32()
		tsPanel.Config.StartPercentageOfTimeRange = &v
	}
	if typeutils.IsKnown(cfg.EndPercentageOfTimeRange) {
		v := cfg.EndPercentageOfTimeRange.ValueFloat32()
		tsPanel.Config.EndPercentageOfTimeRange = &v
	}
	if typeutils.IsKnown(cfg.IsAnchored) {
		tsPanel.Config.IsAnchored = cfg.IsAnchored.ValueBoolPointer()
	}
}
