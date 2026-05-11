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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (pp pinnedPanelModel) syntheticPanelModel() panelModel {
	return panelModel{
		Type:                     pp.Type,
		TimeSliderControlConfig:  pp.TimeSliderControlConfig,
		EsqlControlConfig:        pp.EsqlControlConfig,
		OptionsListControlConfig: pp.OptionsListControlConfig,
		RangeSliderControlConfig: pp.RangeSliderControlConfig,
	}
}

func remapPinnedPanelJSON[A any, B any](in A, out *B) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func pinnedPanelsDiagnosticsErrorsDetail(d diag.Diagnostics) string {
	var parts []string
	for _, x := range d {
		if x.Severity() != diag.SeverityError {
			continue
		}
		parts = append(parts, strings.TrimSpace(x.Summary()+": "+x.Detail()))
	}
	return strings.Join(parts, "; ")
}

func (m *dashboardModel) pinnedPanelsToAPICreateItems() (*kbapi.DashboardPinnedPanels, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.PinnedPanels == nil {
		return nil, diags
	}

	items := make(kbapi.DashboardPinnedPanels, 0, len(m.PinnedPanels))
	for i := range m.PinnedPanels {
		itemPath := path.Root("pinned_panels").AtListIndex(i)
		item, itemDiags := m.PinnedPanels[i].toPinnedAPIItem(itemPath)
		diags.Append(itemDiags...)
		if diags.HasError() {
			return nil, diags
		}
		items = append(items, item)
	}
	return &items, diags
}

func (pp pinnedPanelModel) toPinnedAPIItem(itemPath path.Path) (kbapi.DashboardPinnedPanels_Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	pm := pp.syntheticPanelModel()

	switch pm.Type.ValueString() {
	case panelTypeOptionsListControl:
		olPanel := kbapi.KbnDashboardPanelTypeOptionsListControl{
			Grid: kbapi.KbnDashboardPanelGrid{X: 0, Y: 0},
		}
		buildOptionsListControlConfig(pm, &olPanel)
		var group kbapi.KbnControlsSchemasControlsGroupSchemaOptionsListControl
		if err := remapPinnedPanelJSON(olPanel, &group); err != nil {
			diags.AddAttributeError(itemPath, "Failed to remap pinned options list control", err.Error())
			return kbapi.DashboardPinnedPanels_Item{}, diags
		}
		var item kbapi.DashboardPinnedPanels_Item
		if err := item.FromKbnControlsSchemasControlsGroupSchemaOptionsListControl(group); err != nil {
			diags.AddAttributeError(itemPath, "Failed to build pinned options list control payload", err.Error())
			return kbapi.DashboardPinnedPanels_Item{}, diags
		}
		return item, diags

	case panelTypeRangeSlider:
		rsPanel := kbapi.KbnDashboardPanelTypeRangeSliderControl{
			Grid: kbapi.KbnDashboardPanelGrid{X: 0, Y: 0},
		}
		buildRangeSliderControlConfig(pm, &rsPanel)
		var group kbapi.KbnControlsSchemasControlsGroupSchemaRangeSliderControl
		if err := remapPinnedPanelJSON(rsPanel, &group); err != nil {
			diags.AddAttributeError(itemPath, "Failed to remap pinned range slider control", err.Error())
			return kbapi.DashboardPinnedPanels_Item{}, diags
		}
		var item kbapi.DashboardPinnedPanels_Item
		if err := item.FromKbnControlsSchemasControlsGroupSchemaRangeSliderControl(group); err != nil {
			diags.AddAttributeError(itemPath, "Failed to build pinned range slider control payload", err.Error())
			return kbapi.DashboardPinnedPanels_Item{}, diags
		}
		return item, diags

	case panelTypeTimeSlider:
		tsPanel := kbapi.KbnDashboardPanelTypeTimeSliderControl{
			Grid: kbapi.KbnDashboardPanelGrid{X: 0, Y: 0},
			Config: struct {
				EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
				IsAnchored                 *bool    `json:"is_anchored,omitempty"`
				StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
			}{},
		}
		buildTimeSliderControlConfig(pm, &tsPanel)
		var group kbapi.KbnControlsSchemasControlsGroupSchemaTimeSliderControl
		if err := remapPinnedPanelJSON(tsPanel, &group); err != nil {
			diags.AddAttributeError(itemPath, "Failed to remap pinned time slider control", err.Error())
			return kbapi.DashboardPinnedPanels_Item{}, diags
		}
		var item kbapi.DashboardPinnedPanels_Item
		if err := item.FromKbnControlsSchemasControlsGroupSchemaTimeSliderControl(group); err != nil {
			diags.AddAttributeError(itemPath, "Failed to build pinned time slider control payload", err.Error())
			return kbapi.DashboardPinnedPanels_Item{}, diags
		}
		return item, diags

	case panelTypeEsqlControl:
		esqlPanel := kbapi.KbnDashboardPanelTypeEsqlControl{
			Grid: kbapi.KbnDashboardPanelGrid{X: 0, Y: 0},
		}
		esqlDiags := buildEsqlControlConfig(pm, &esqlPanel)
		if esqlDiags.HasError() {
			diags.AddAttributeError(itemPath, "Invalid pinned ES|QL control configuration", pinnedPanelsDiagnosticsErrorsDetail(esqlDiags))
			return kbapi.DashboardPinnedPanels_Item{}, diags
		}
		diags.Append(esqlDiags...)
		var group kbapi.KbnControlsSchemasControlsGroupSchemaEsqlControl
		if err := remapPinnedPanelJSON(esqlPanel, &group); err != nil {
			diags.AddAttributeError(itemPath, "Failed to remap pinned ES|QL control", err.Error())
			return kbapi.DashboardPinnedPanels_Item{}, diags
		}
		var item kbapi.DashboardPinnedPanels_Item
		if err := item.FromKbnControlsSchemasControlsGroupSchemaEsqlControl(group); err != nil {
			diags.AddAttributeError(itemPath, "Failed to build pinned ES|QL control payload", err.Error())
			return kbapi.DashboardPinnedPanels_Item{}, diags
		}
		return item, diags

	default:
		diags.AddAttributeError(
			itemPath,
			"Unsupported pinned panel type",
			"pinned_panels entries must use one of the supported dashboard control types.",
		)
		return kbapi.DashboardPinnedPanels_Item{}, diags
	}
}

// seedPinnedPanelModelForRead seeds a pinned panel model with the discriminator
// from the API response, carrying prior TF state forward only when its `type`
// matches. populateTf is non-nil only when the prior TF state can be reused for
// drift preservation in populate*FromAPI helpers.
func seedPinnedPanelModelForRead(tf *pinnedPanelModel, discriminator string) (ppm pinnedPanelModel, populateTf *panelModel) {
	if tf != nil {
		ppm = *tf
	}
	ppm.Type = types.StringValue(discriminator)

	if tf != nil && typeutils.IsKnown(tf.Type) && tf.Type.ValueString() == discriminator {
		pm := tf.syntheticPanelModel()
		populateTf = &pm
	} else {
		ppm.OptionsListControlConfig = nil
		ppm.RangeSliderControlConfig = nil
		ppm.TimeSliderControlConfig = nil
		ppm.EsqlControlConfig = nil
	}

	return ppm, populateTf
}

// applyPinnedControlConfig assigns the active control config from a synthetic
// panelModel onto ppm and clears the other three sibling slots so each pinned
// entry only carries the discriminator-matching block.
func (pp *pinnedPanelModel) applyPinnedControlConfig(active string, pm *panelModel) {
	pp.OptionsListControlConfig = nil
	pp.RangeSliderControlConfig = nil
	pp.TimeSliderControlConfig = nil
	pp.EsqlControlConfig = nil
	switch active {
	case panelTypeOptionsListControl:
		pp.OptionsListControlConfig = pm.OptionsListControlConfig
	case panelTypeRangeSlider:
		pp.RangeSliderControlConfig = pm.RangeSliderControlConfig
	case panelTypeTimeSlider:
		pp.TimeSliderControlConfig = pm.TimeSliderControlConfig
	case panelTypeEsqlControl:
		pp.EsqlControlConfig = pm.EsqlControlConfig
	}
}

func (m *dashboardModel) mapPinnedPanelsFromAPI(ctx context.Context, prior []pinnedPanelModel, api *[]kbapi.DashboardPinnedPanels_Item) ([]pinnedPanelModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if api == nil || len(*api) == 0 {
		if prior == nil {
			return nil, diags
		}
		return []pinnedPanelModel{}, diags
	}

	out := make([]pinnedPanelModel, 0, len(*api))
	for i, raw := range *api {
		itemPath := path.Root("pinned_panels").AtListIndex(i)

		var tf *pinnedPanelModel
		if i < len(prior) {
			tf = &prior[i]
		}

		discriminator, err := raw.Discriminator()
		if err != nil {
			diags.AddAttributeError(itemPath, "Failed to read pinned panel type", err.Error())
			return nil, diags
		}

		switch discriminator {
		case panelTypeOptionsListControl:
			group, err := raw.AsKbnControlsSchemasControlsGroupSchemaOptionsListControl()
			if err != nil {
				diags.AddAttributeError(itemPath, "Failed to parse pinned options list control", err.Error())
				return nil, diags
			}
			var olPanel kbapi.KbnDashboardPanelTypeOptionsListControl
			if err := remapPinnedPanelJSON(group, &olPanel); err != nil {
				diags.AddAttributeError(itemPath, "Failed to remap pinned options list control from API", err.Error())
				return nil, diags
			}

			ppm, populateTf := seedPinnedPanelModelForRead(tf, panelTypeOptionsListControl)

			pm := ppm.syntheticPanelModel()
			populateOptionsListControlFromAPI(&pm, populateTf, &olPanel)

			ppm.applyPinnedControlConfig(panelTypeOptionsListControl, &pm)
			out = append(out, ppm)

		case panelTypeRangeSlider:
			group, err := raw.AsKbnControlsSchemasControlsGroupSchemaRangeSliderControl()
			if err != nil {
				diags.AddAttributeError(itemPath, "Failed to parse pinned range slider control", err.Error())
				return nil, diags
			}
			var rsPanel kbapi.KbnDashboardPanelTypeRangeSliderControl
			if err := remapPinnedPanelJSON(group, &rsPanel); err != nil {
				diags.AddAttributeError(itemPath, "Failed to remap pinned range slider control from API", err.Error())
				return nil, diags
			}

			ppm, populateTf := seedPinnedPanelModelForRead(tf, panelTypeRangeSlider)

			pm := ppm.syntheticPanelModel()
			populateRangeSliderControlFromAPI(ctx, &pm, populateTf, &rsPanel)

			ppm.applyPinnedControlConfig(panelTypeRangeSlider, &pm)
			out = append(out, ppm)

		case panelTypeTimeSlider:
			group, err := raw.AsKbnControlsSchemasControlsGroupSchemaTimeSliderControl()
			if err != nil {
				diags.AddAttributeError(itemPath, "Failed to parse pinned time slider control", err.Error())
				return nil, diags
			}
			var tsPanel kbapi.KbnDashboardPanelTypeTimeSliderControl
			if err := remapPinnedPanelJSON(group, &tsPanel); err != nil {
				diags.AddAttributeError(itemPath, "Failed to remap pinned time slider control from API", err.Error())
				return nil, diags
			}

			ppm, populateTf := seedPinnedPanelModelForRead(tf, panelTypeTimeSlider)

			pm := ppm.syntheticPanelModel()
			populateTimeSliderControlFromAPI(&pm, populateTf, tsPanel.Config)

			ppm.applyPinnedControlConfig(panelTypeTimeSlider, &pm)
			out = append(out, ppm)

		case panelTypeEsqlControl:
			group, err := raw.AsKbnControlsSchemasControlsGroupSchemaEsqlControl()
			if err != nil {
				diags.AddAttributeError(itemPath, "Failed to parse pinned ES|QL control", err.Error())
				return nil, diags
			}
			var esqlPanel kbapi.KbnDashboardPanelTypeEsqlControl
			if err := remapPinnedPanelJSON(group, &esqlPanel); err != nil {
				diags.AddAttributeError(itemPath, "Failed to remap pinned ES|QL control from API", err.Error())
				return nil, diags
			}

			ppm, populateTf := seedPinnedPanelModelForRead(tf, panelTypeEsqlControl)

			pm := ppm.syntheticPanelModel()
			populateEsqlControlFromAPI(&pm, populateTf, esqlPanel.Config)

			ppm.applyPinnedControlConfig(panelTypeEsqlControl, &pm)
			out = append(out, ppm)

		default:
			diags.AddAttributeError(
				itemPath,
				"Unsupported pinned panel type",
				"The dashboard API returned a pinned control type that is not supported by this resource.",
			)
			return nil, diags
		}
	}

	return out, diags
}
