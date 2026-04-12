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

type timeSliderControlConfigModel struct {
	StartPercentageOfTimeRange types.Float32 `tfsdk:"start_percentage_of_time_range"`
	EndPercentageOfTimeRange   types.Float32 `tfsdk:"end_percentage_of_time_range"`
	IsAnchored                 types.Bool    `tfsdk:"is_anchored"`
}

// populateTimeSliderControlFromAPI reads back a time slider control config from the API
// response and updates the panel model. Null-preservation semantics apply: if a field is
// null in the existing TF state (pm.TimeSliderControlConfig), we do not overwrite it with
// a Kibana-returned value. If there is no existing config block, and Kibana returns an
// empty/absent config, we leave TimeSliderControlConfig as nil.
//
// tfPanel is the prior TF state/plan panel, or nil on import. When nil, the function
// populates all API-returned fields unconditionally (no prior intent to preserve).
func populateTimeSliderControlFromAPI(pm *panelModel, tfPanel *panelModel, apiConfig struct {
	EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
	IsAnchored                 *bool    `json:"is_anchored,omitempty"`
	StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
}) {
	existing := pm.TimeSliderControlConfig

	// On import (tfPanel == nil) there is no prior intent. Populate from API if data exists.
	if tfPanel == nil {
		if apiConfig.StartPercentageOfTimeRange == nil &&
			apiConfig.EndPercentageOfTimeRange == nil &&
			apiConfig.IsAnchored == nil {
			return
		}
		pm.TimeSliderControlConfig = &timeSliderControlConfigModel{}
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

	// If the existing state has no config block, never create one here; always preserve nil intent.
	if existing == nil {
		// If Kibana returned nothing meaningful, there's nothing to sync, so keep the block nil.
		if apiConfig.StartPercentageOfTimeRange == nil &&
			apiConfig.EndPercentageOfTimeRange == nil &&
			apiConfig.IsAnchored == nil {
			return
		}
		// Kibana returned data but the practitioner didn't configure a block — still preserve nil intent.
		return
	}

	// Block exists in state — update only fields that are already known (non-null).
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

// buildTimeSliderControlConfig writes the TF model fields into the API panel struct.
func buildTimeSliderControlConfig(pm panelModel, tsPanel *kbapi.KbnDashboardPanelTypeTimeSliderControl) {
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
