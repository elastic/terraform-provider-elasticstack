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

// SimpleFromAPI handles the common FromAPI boilerplate shared by simple panel types.
//
// asPanel decodes the raw DashboardPanelItem into the typed panel T; a decode error causes
// an immediate return with a "Dashboard panel decode" diagnostic.
// gridID extracts the panel's Grid and Id fields from T.
// populate writes the panel-type-specific config from T into pm (with prior for plan alignment).
func SimpleFromAPI[T any](
	_ context.Context,
	pm, prior *models.PanelModel,
	item kbapi.DashboardPanelItem,
	asPanel func() (T, error),
	gridID func(T) (kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, *string),
	populate func(pm *models.PanelModel, prior *models.PanelModel, panel T) diag.Diagnostics,
) diag.Diagnostics {
	apiPanel, err := asPanel()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Dashboard panel decode", err.Error())
		return d
	}

	grid, id := gridID(apiPanel)
	pm.Grid = GridFromAPI(grid.X, grid.Y, grid.W, grid.H)
	pm.ID = IDFromAPI(id)
	pm.ConfigJSON = PanelConfigJSONNull()
	return populate(pm, prior, apiPanel)
}

// SimpleToAPI handles the common ToAPI boilerplate shared by simple panel types.
//
// mkPanel receives the pre-computed grid and id and returns the fully configured typed panel T
// along with any diagnostics from populating its config.
// fromPanel serializes T into a DashboardPanelItem; a serialization error adds an error diagnostic
// using errorMsg as the summary.
func SimpleToAPI[T any](
	pm models.PanelModel,
	mkPanel func(grid kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, id *string) (T, diag.Diagnostics),
	fromPanel func(*kbapi.DashboardPanelItem, T) error,
	errorMsg string,
) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	grid := GridToAPI(pm.Grid)
	id := IDToAPI(pm.ID)

	// GridToAPI returns an anonymous struct that is identical to kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid.
	apiGrid := kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid(grid)
	panel, diags := mkPanel(apiGrid, id)

	var panelItem kbapi.DashboardPanelItem
	if err := fromPanel(&panelItem, panel); err != nil {
		diags.AddError(errorMsg, err.Error())
	}
	return panelItem, diags
}
