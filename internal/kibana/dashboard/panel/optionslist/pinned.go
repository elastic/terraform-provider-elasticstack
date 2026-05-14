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

package optionslist

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type pinnedHandler struct{}

func (pinnedHandler) FromAPI(ctx context.Context, prior *models.PinnedPanelModel, raw kbapi.DashboardPinnedPanels_Item) (models.PinnedPanelModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	group, err := raw.AsKbnControlsSchemasControlsGroupSchemaOptionsListControl()
	if err != nil {
		diags.AddError("Failed to parse pinned options list control", err.Error())
		return models.PinnedPanelModel{}, diags
	}
	var olPanel kbapi.KbnDashboardPanelTypeOptionsListControl
	if err := panelkit.RemapViaJSON(group, &olPanel); err != nil {
		diags.AddError("Failed to remap pinned options list control from API", err.Error())
		return models.PinnedPanelModel{}, diags
	}

	ppm, populateTf := models.SeedPinnedPanelForRead(prior, panelType)
	pm := ppm.SyntheticPanel()
	PopulateFromAPI(&pm, populateTf, &olPanel)
	models.ApplyPinnedSiblingControlConfig(&ppm, panelType, &pm)
	_ = ctx
	return ppm, diags
}

func (pinnedHandler) ToAPI(ppm models.PinnedPanelModel) (kbapi.DashboardPinnedPanels_Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	pm := ppm.SyntheticPanel()
	olPanel := kbapi.KbnDashboardPanelTypeOptionsListControl{
		Grid: kbapi.KbnDashboardPanelGrid{X: 0, Y: 0},
	}
	BuildConfig(pm, &olPanel)
	var group kbapi.KbnControlsSchemasControlsGroupSchemaOptionsListControl
	if err := panelkit.RemapViaJSON(olPanel, &group); err != nil {
		diags.AddError("Failed to remap pinned options list control", err.Error())
		return kbapi.DashboardPinnedPanels_Item{}, diags
	}
	var item kbapi.DashboardPinnedPanels_Item
	if err := item.FromKbnControlsSchemasControlsGroupSchemaOptionsListControl(group); err != nil {
		diags.AddError("Failed to build pinned options list control payload", err.Error())
		return kbapi.DashboardPinnedPanels_Item{}, diags
	}
	return item, diags
}
