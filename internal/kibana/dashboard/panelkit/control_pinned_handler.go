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

package panelkit

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ControlPinnedHandler is a reusable generic implementation of iface.PinnedHandler for
// dashboard control-bar panels. Each control-panel package instantiates one with its
// own type parameters and callbacks, eliminating the near-identical boilerplate that
// was duplicated across esqlcontrol, optionslist, rangeslider, and timeslider.
//
// G is the kbapi "group schema" type (ControlsGroupSchema* variants).
// P is the kbapi "dashboard panel type" (DashboardPanelType* variants).
type ControlPinnedHandler[G any, P any] struct {
	// PanelTypeDiscriminator is the panel type string (e.g. "range_slider_control").
	PanelTypeDiscriminator string

	// AsGroup extracts the G value from a DashboardPinnedPanels_Item union.
	AsGroup func(raw kbapi.DashboardPinnedPanels_Item) (G, error)

	// BuildPanel returns a zero-initialized P ready to be populated.
	BuildPanel func() P

	// PopulateFromAPI populates pm from the P value and returns any diagnostics.
	// The ctx and prior arguments are passed through unchanged from the outer call.
	// Implementations that do not need ctx or do not return diags should wrap the
	// local package function in a small closure.
	PopulateFromAPI func(ctx context.Context, pm *models.PanelModel, prior *models.PanelModel, panel *P) diag.Diagnostics

	// BuildConfig writes pm fields into panel and returns any diagnostics.
	// Implementations that do not return diags should return nil.
	BuildConfig func(pm models.PanelModel, panel *P) diag.Diagnostics

	// AfterRemapToGroup is called after panel is remapped back to a G value, allowing
	// callers to set extra discriminator fields (e.g. group.Type). May be nil.
	AfterRemapToGroup func(group *G)

	// FromGroup encodes a G into a DashboardPinnedPanels_Item union.
	FromGroup func(item *kbapi.DashboardPinnedPanels_Item, group G) error

	// Error message strings used in diagnostics.
	ParseErrSummary     string
	RemapFromErrSummary string
	RemapToErrSummary   string
	FromGroupErrSummary string
}

// FromAPI implements iface.PinnedHandler.
func (h ControlPinnedHandler[G, P]) FromAPI(ctx context.Context, prior *models.PinnedPanelModel, raw kbapi.DashboardPinnedPanels_Item) (models.PinnedPanelModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if h.PanelTypeDiscriminator == "" || h.AsGroup == nil || h.BuildPanel == nil || h.PopulateFromAPI == nil {
		diags.AddError("Internal error", "ControlPinnedHandler is misconfigured (missing required callback or panel type discriminator)")
		return models.PinnedPanelModel{}, diags
	}

	group, err := h.AsGroup(raw)
		diags.AddError(h.ParseErrSummary, err.Error())
		return models.PinnedPanelModel{}, diags
	}

	panel := h.BuildPanel()
	if err := RemapViaJSON(group, &panel); err != nil {
		diags.AddError(h.RemapFromErrSummary, err.Error())
		return models.PinnedPanelModel{}, diags
	}

	ppm, populateTf := models.SeedPinnedPanelForRead(prior, h.PanelTypeDiscriminator)
	pm := ppm.SyntheticPanel()
	populateDiags := h.PopulateFromAPI(ctx, &pm, populateTf, &panel)
	diags.Append(populateDiags...)
	if diags.HasError() {
		return models.PinnedPanelModel{}, diags
	}
	models.ApplyPinnedSiblingControlConfig(&ppm, h.PanelTypeDiscriminator, &pm)
	return ppm, diags
}

// ToAPI implements iface.PinnedHandler.
func (h ControlPinnedHandler[G, P]) ToAPI(ppm models.PinnedPanelModel) (kbapi.DashboardPinnedPanels_Item, diag.Diagnostics) {
	var diags diag.Diagnostics

	if h.BuildPanel == nil || h.BuildConfig == nil || h.FromGroup == nil {
		diags.AddError("Internal error", "ControlPinnedHandler is misconfigured (missing required callback)")
		return kbapi.DashboardPinnedPanels_Item{}, diags
	}

	pm := ppm.SyntheticPanel()
	panel := h.BuildPanel()
	if diags.HasError() {
		return kbapi.DashboardPinnedPanels_Item{}, diags
	}

	var group G
	if err := RemapViaJSON(panel, &group); err != nil {
		diags.AddError(h.RemapToErrSummary, err.Error())
		return kbapi.DashboardPinnedPanels_Item{}, diags
	}
	if h.AfterRemapToGroup != nil {
		h.AfterRemapToGroup(&group)
	}

	var item kbapi.DashboardPinnedPanels_Item
	if err := h.FromGroup(&item, group); err != nil {
		diags.AddError(h.FromGroupErrSummary, err.Error())
		return kbapi.DashboardPinnedPanels_Item{}, diags
	}
	return item, diags
}
