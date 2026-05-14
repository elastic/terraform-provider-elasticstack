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

package models

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SyntheticPanel maps a pinned control-bar entry onto the subset of PanelModel shared with dashboard panels so the control
// FromAPI helpers can populate typed config without importing the dashboard package.
func (pp PinnedPanelModel) SyntheticPanel() PanelModel {
	return PanelModel{
		Type:                     pp.Type,
		TimeSliderControlConfig:  pp.TimeSliderControlConfig,
		EsqlControlConfig:        pp.EsqlControlConfig,
		OptionsListControlConfig: pp.OptionsListControlConfig,
		RangeSliderControlConfig: pp.RangeSliderControlConfig,
	}
}

// SeedPinnedPanelForRead seeds a pinned model with the discriminator from the API response, carrying forward prior TF state
// when `type` matches so null-preserving FromAPI behaves like dashboard panels matched by slice index + prior slice.
func SeedPinnedPanelForRead(prior *PinnedPanelModel, discriminator string) (PinnedPanelModel, *PanelModel) {
	var ppm PinnedPanelModel
	if prior != nil {
		ppm = *prior
	}
	ppm.Type = types.StringValue(discriminator)

	if prior != nil && typeutils.IsKnown(prior.Type) && prior.Type.ValueString() == discriminator {
		pm := ppm.SyntheticPanel()
		return ppm, &pm
	}
	ppm.OptionsListControlConfig = nil
	ppm.RangeSliderControlConfig = nil
	ppm.TimeSliderControlConfig = nil
	ppm.EsqlControlConfig = nil

	return ppm, nil
}

// ApplyPinnedSiblingControlConfig copies the active control config slot from synthetic pm onto ppm and clears mismatched sibling slots.
func ApplyPinnedSiblingControlConfig(pp *PinnedPanelModel, active string, pm *PanelModel) {
	pp.OptionsListControlConfig = nil
	pp.RangeSliderControlConfig = nil
	pp.TimeSliderControlConfig = nil
	pp.EsqlControlConfig = nil
	switch active {
	case "options_list_control":
		pp.OptionsListControlConfig = pm.OptionsListControlConfig
	case "range_slider_control":
		pp.RangeSliderControlConfig = pm.RangeSliderControlConfig
	case "time_slider_control":
		pp.TimeSliderControlConfig = pm.TimeSliderControlConfig
	case "esql_control":
		pp.EsqlControlConfig = pm.EsqlControlConfig
	}
}
