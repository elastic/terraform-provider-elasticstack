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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

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

// appendHandlerDiagnosticsWithPinnedItemPath rewrites handler-emitted diagnostics onto `itemPath` so list-index
// context is preserved (iface.PinnedHandler implementations use diag.AddError/AddWarning without paths).
func appendHandlerDiagnosticsWithPinnedItemPath(dst *diag.Diagnostics, itemPath path.Path, src diag.Diagnostics) {
	for _, d := range src {
		switch d.Severity() {
		case diag.SeverityError:
			dst.AddAttributeError(itemPath, d.Summary(), d.Detail())
		case diag.SeverityWarning:
			dst.AddAttributeWarning(itemPath, d.Summary(), d.Detail())
		default:
			dst.Append(diag.WithPath(itemPath, d))
		}
	}
}

func dashboardPinnedPanelsToAPICreateItems(m *models.DashboardModel) (*kbapi.DashboardPinnedPanels, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.PinnedPanels == nil {
		return nil, diags
	}

	items := make(kbapi.DashboardPinnedPanels, 0, len(m.PinnedPanels))
	for i := range m.PinnedPanels {
		itemPath := path.Root("pinned_panels").AtListIndex(i)
		item, itemDiags := pinnedPanelToPinnedAPIItem(m.PinnedPanels[i], itemPath)
		diags.Append(itemDiags...)
		if diags.HasError() {
			return nil, diags
		}
		items = append(items, item)
	}
	return &items, diags
}

func pinnedPanelToPinnedAPIItem(pp models.PinnedPanelModel, itemPath path.Path) (kbapi.DashboardPinnedPanels_Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(pp.Type) {
		diags.AddAttributeError(itemPath, "Invalid pinned panel entry", "pinned panel `type` must be known.")
		return kbapi.DashboardPinnedPanels_Item{}, diags
	}

	h := LookupHandler(pp.Type.ValueString())
	if h == nil || h.PinnedHandler() == nil {
		diags.AddAttributeError(
			itemPath,
			"Unsupported pinned panel type",
			"pinned_panels entries must use one of the supported dashboard control types.",
		)
		return kbapi.DashboardPinnedPanels_Item{}, diags
	}

	ph := h.PinnedHandler()
	item, itemDiags := ph.ToAPI(pp)
	if itemDiags.HasError() {
		summary := "Invalid pinned panel configuration"
		if pp.Type.ValueString() == panelTypeEsqlControl {
			summary = "Invalid pinned ES|QL control configuration"
		}
		diags.AddAttributeError(itemPath, summary, pinnedPanelsDiagnosticsErrorsDetail(itemDiags))
		return kbapi.DashboardPinnedPanels_Item{}, diags
	}
	appendHandlerDiagnosticsWithPinnedItemPath(&diags, itemPath, itemDiags)
	return item, diags
}

func dashboardMapPinnedPanelsFromAPI(
	ctx context.Context,
	prior []models.PinnedPanelModel,
	api *[]kbapi.DashboardPinnedPanels_Item,
) ([]models.PinnedPanelModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if api == nil || len(*api) == 0 {
		if prior == nil {
			return nil, diags
		}
		return []models.PinnedPanelModel{}, diags
	}

	out := make([]models.PinnedPanelModel, 0, len(*api))
	for i, raw := range *api {
		itemPath := path.Root("pinned_panels").AtListIndex(i)

		var tf *models.PinnedPanelModel
		if i < len(prior) {
			tf = &prior[i]
		}

		discriminator, err := raw.Discriminator()
		if err != nil {
			diags.AddAttributeError(itemPath, "Failed to read pinned panel type", err.Error())
			return nil, diags
		}

		h := LookupHandler(discriminator)
		if h == nil || h.PinnedHandler() == nil {
			diags.AddAttributeError(
				itemPath,
				"Unsupported pinned panel type",
				"The dashboard API returned a pinned control type that is not supported by this resource.",
			)
			return nil, diags
		}

		ppm, d := h.PinnedHandler().FromAPI(ctx, tf, raw)
		appendHandlerDiagnosticsWithPinnedItemPath(&diags, itemPath, d)
		if diags.HasError() {
			return nil, diags
		}
		out = append(out, ppm)
	}

	return out, diags
}
