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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func pinnedPanelCreateItemsToPutItems(items []kbapi.KbnDashboardData_PinnedPanels_Item) ([]kbapi.PutDashboardsIdJSONBody_PinnedPanels_Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := make([]kbapi.PutDashboardsIdJSONBody_PinnedPanels_Item, len(items))
	for i := range items {
		b, err := json.Marshal(items[i])
		if err != nil {
			diags.AddError("Failed to marshal pinned panel", err.Error())
			return nil, diags
		}
		if err := json.Unmarshal(b, &out[i]); err != nil {
			diags.AddError("Failed to convert pinned panel for update", err.Error())
			return nil, diags
		}
	}
	return out, diags
}

func (m *dashboardModel) pinnedPanelsToAPICreateItems() (*[]kbapi.KbnDashboardData_PinnedPanels_Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.PinnedPanels == nil {
		return nil, diags
	}

	items := make([]kbapi.KbnDashboardData_PinnedPanels_Item, 0, len(m.PinnedPanels))
	for _, pp := range m.PinnedPanels {
		item, itemDiags := pp.toPinnedAPIItem()
		diags.Append(itemDiags...)
		if diags.HasError() {
			return nil, diags
		}
		items = append(items, item)
	}
	return &items, diags
}

func (m *dashboardModel) pinnedPanelsToAPIPutItems() (*[]kbapi.PutDashboardsIdJSONBody_PinnedPanels_Item, diag.Diagnostics) {
	createItems, diags := m.pinnedPanelsToAPICreateItems()
	if createItems == nil || diags.HasError() {
		return nil, diags
	}

	putItems, convDiags := pinnedPanelCreateItemsToPutItems(*createItems)
	diags.Append(convDiags...)
	if diags.HasError() {
		return nil, diags
	}
	return &putItems, diags
}

func (pp pinnedPanelModel) toPinnedAPIItem() (kbapi.KbnDashboardData_PinnedPanels_Item, diag.Diagnostics) {
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
			return kbapi.KbnDashboardData_PinnedPanels_Item{}, diagutil.FrameworkDiagFromError(err)
		}
		var item kbapi.KbnDashboardData_PinnedPanels_Item
		if err := item.FromKbnControlsSchemasControlsGroupSchemaOptionsListControl(group); err != nil {
			return kbapi.KbnDashboardData_PinnedPanels_Item{}, diagutil.FrameworkDiagFromError(err)
		}
		return item, diags

	case panelTypeRangeSlider:
		rsPanel := kbapi.KbnDashboardPanelTypeRangeSliderControl{
			Grid: kbapi.KbnDashboardPanelGrid{X: 0, Y: 0},
		}
		buildRangeSliderControlConfig(pm, &rsPanel)
		var group kbapi.KbnControlsSchemasControlsGroupSchemaRangeSliderControl
		if err := remapPinnedPanelJSON(rsPanel, &group); err != nil {
			return kbapi.KbnDashboardData_PinnedPanels_Item{}, diagutil.FrameworkDiagFromError(err)
		}
		var item kbapi.KbnDashboardData_PinnedPanels_Item
		if err := item.FromKbnControlsSchemasControlsGroupSchemaRangeSliderControl(group); err != nil {
			return kbapi.KbnDashboardData_PinnedPanels_Item{}, diagutil.FrameworkDiagFromError(err)
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
			return kbapi.KbnDashboardData_PinnedPanels_Item{}, diagutil.FrameworkDiagFromError(err)
		}
		var item kbapi.KbnDashboardData_PinnedPanels_Item
		if err := item.FromKbnControlsSchemasControlsGroupSchemaTimeSliderControl(group); err != nil {
			return kbapi.KbnDashboardData_PinnedPanels_Item{}, diagutil.FrameworkDiagFromError(err)
		}
		return item, diags

	case panelTypeEsqlControl:
		esqlPanel := kbapi.KbnDashboardPanelTypeEsqlControl{
			Grid: kbapi.KbnDashboardPanelGrid{X: 0, Y: 0},
		}
		diags.Append(buildEsqlControlConfig(pm, &esqlPanel)...)
		if diags.HasError() {
			return kbapi.KbnDashboardData_PinnedPanels_Item{}, diags
		}
		var group kbapi.KbnControlsSchemasControlsGroupSchemaEsqlControl
		if err := remapPinnedPanelJSON(esqlPanel, &group); err != nil {
			return kbapi.KbnDashboardData_PinnedPanels_Item{}, diagutil.FrameworkDiagFromError(err)
		}
		var item kbapi.KbnDashboardData_PinnedPanels_Item
		if err := item.FromKbnControlsSchemasControlsGroupSchemaEsqlControl(group); err != nil {
			return kbapi.KbnDashboardData_PinnedPanels_Item{}, diagutil.FrameworkDiagFromError(err)
		}
		return item, diags

	default:
		diags.AddError(
			"Unsupported pinned panel type",
			"pinned_panels entries must use one of the supported dashboard control types.",
		)
		return kbapi.KbnDashboardData_PinnedPanels_Item{}, diags
	}
}

func syntheticTfPanelFromPinned(tf *pinnedPanelModel) *panelModel {
	if tf == nil {
		return nil
	}
	pm := tf.syntheticPanelModel()
	return &pm
}

func (m *dashboardModel) mapPinnedPanelsFromAPI(ctx context.Context, prior []pinnedPanelModel, api *[]kbapi.KbnDashboardData_PinnedPanels_Item) ([]pinnedPanelModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if api == nil || len(*api) == 0 {
		if prior == nil {
			return nil, diags
		}
		return []pinnedPanelModel{}, diags
	}

	out := make([]pinnedPanelModel, 0, len(*api))
	for i, raw := range *api {
		var tf *pinnedPanelModel
		if i < len(prior) {
			tf = &prior[i]
		}

		discriminator, err := raw.Discriminator()
		if err != nil {
			diags.Append(diagutil.FrameworkDiagFromError(err)...)
			return nil, diags
		}

		switch discriminator {
		case panelTypeOptionsListControl:
			group, err := raw.AsKbnControlsSchemasControlsGroupSchemaOptionsListControl()
			if err != nil {
				diags.Append(diagutil.FrameworkDiagFromError(err)...)
				return nil, diags
			}
			var olPanel kbapi.KbnDashboardPanelTypeOptionsListControl
			if err := remapPinnedPanelJSON(group, &olPanel); err != nil {
				diags.Append(diagutil.FrameworkDiagFromError(err)...)
				return nil, diags
			}

			var ppm pinnedPanelModel
			if tf != nil {
				ppm = *tf
			}
			ppm.Type = types.StringValue(panelTypeOptionsListControl)

			pm := ppm.syntheticPanelModel()
			populateOptionsListControlFromAPI(&pm, syntheticTfPanelFromPinned(tf), &olPanel)

			ppm.OptionsListControlConfig = pm.OptionsListControlConfig
			ppm.RangeSliderControlConfig = nil
			ppm.TimeSliderControlConfig = nil
			ppm.EsqlControlConfig = nil
			out = append(out, ppm)

		case panelTypeRangeSlider:
			group, err := raw.AsKbnControlsSchemasControlsGroupSchemaRangeSliderControl()
			if err != nil {
				diags.Append(diagutil.FrameworkDiagFromError(err)...)
				return nil, diags
			}
			var rsPanel kbapi.KbnDashboardPanelTypeRangeSliderControl
			if err := remapPinnedPanelJSON(group, &rsPanel); err != nil {
				diags.Append(diagutil.FrameworkDiagFromError(err)...)
				return nil, diags
			}

			var ppm pinnedPanelModel
			if tf != nil {
				ppm = *tf
			}
			ppm.Type = types.StringValue(panelTypeRangeSlider)

			pm := ppm.syntheticPanelModel()
			populateRangeSliderControlFromAPI(ctx, &pm, syntheticTfPanelFromPinned(tf), &rsPanel)

			ppm.RangeSliderControlConfig = pm.RangeSliderControlConfig
			ppm.OptionsListControlConfig = nil
			ppm.TimeSliderControlConfig = nil
			ppm.EsqlControlConfig = nil
			out = append(out, ppm)

		case panelTypeTimeSlider:
			group, err := raw.AsKbnControlsSchemasControlsGroupSchemaTimeSliderControl()
			if err != nil {
				diags.Append(diagutil.FrameworkDiagFromError(err)...)
				return nil, diags
			}
			var tsPanel kbapi.KbnDashboardPanelTypeTimeSliderControl
			if err := remapPinnedPanelJSON(group, &tsPanel); err != nil {
				diags.Append(diagutil.FrameworkDiagFromError(err)...)
				return nil, diags
			}

			var ppm pinnedPanelModel
			if tf != nil {
				ppm = *tf
			}
			ppm.Type = types.StringValue(panelTypeTimeSlider)

			pm := ppm.syntheticPanelModel()
			populateTimeSliderControlFromAPI(&pm, syntheticTfPanelFromPinned(tf), tsPanel.Config)

			ppm.TimeSliderControlConfig = pm.TimeSliderControlConfig
			ppm.OptionsListControlConfig = nil
			ppm.RangeSliderControlConfig = nil
			ppm.EsqlControlConfig = nil
			out = append(out, ppm)

		case panelTypeEsqlControl:
			group, err := raw.AsKbnControlsSchemasControlsGroupSchemaEsqlControl()
			if err != nil {
				diags.Append(diagutil.FrameworkDiagFromError(err)...)
				return nil, diags
			}
			var esqlPanel kbapi.KbnDashboardPanelTypeEsqlControl
			if err := remapPinnedPanelJSON(group, &esqlPanel); err != nil {
				diags.Append(diagutil.FrameworkDiagFromError(err)...)
				return nil, diags
			}

			var ppm pinnedPanelModel
			if tf != nil {
				ppm = *tf
			}
			ppm.Type = types.StringValue(panelTypeEsqlControl)

			pm := ppm.syntheticPanelModel()
			populateEsqlControlFromAPI(&pm, syntheticTfPanelFromPinned(tf), esqlPanel.Config)

			ppm.EsqlControlConfig = pm.EsqlControlConfig
			ppm.OptionsListControlConfig = nil
			ppm.RangeSliderControlConfig = nil
			ppm.TimeSliderControlConfig = nil
			out = append(out, ppm)

		default:
			diags.AddError(
				"Unsupported pinned panel type",
				"The dashboard API returned a pinned control type that is not supported by this resource.",
			)
			return nil, diags
		}
	}

	return out, diags
}
