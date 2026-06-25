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

type pinnedHandler = panelkit.ControlPinnedHandler[
	kbapi.KibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaOptionsListControl,
	kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl,
]

func newPinnedHandler() pinnedHandler {
	return pinnedHandler{
		PanelTypeDiscriminator: panelType,
		AsGroup: func(raw kbapi.DashboardPinnedPanels_Item) (kbapi.KibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaOptionsListControl, error) {
			return raw.AsKibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaOptionsListControl()
		},
		BuildPanel: func() kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl {
			return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{
				Grid: kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid{X: 0, Y: 0},
			}
		},
		PopulateFromAPI: func(_ context.Context, pm *models.PanelModel, prior *models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl) diag.Diagnostics {
			return PopulateFromAPI(pm, prior, panel)
		},
		BuildConfig: BuildConfig,
		FromGroup: func(item *kbapi.DashboardPinnedPanels_Item, group kbapi.KibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaOptionsListControl) error {
			return item.FromKibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaOptionsListControl(group)
		},
		ParseErrSummary:     "Failed to parse pinned options list control",
		RemapFromErrSummary: "Failed to remap pinned options list control from API",
		RemapToErrSummary:   "Failed to remap pinned options list control",
		FromGroupErrSummary: "Failed to build pinned options list control payload",
	}
}
