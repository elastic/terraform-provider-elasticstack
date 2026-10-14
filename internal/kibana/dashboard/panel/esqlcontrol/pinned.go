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

package esqlcontrol

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type pinnedHandler = panelkit.ControlPinnedHandler[
	kbapi.KibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaEsqlControl,
	kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeEsqlControl,
]

func newPinnedHandler() pinnedHandler {
	return pinnedHandler{
		PanelTypeDiscriminator: panelType,
		AsGroup: func(raw kbapi.DashboardPinnedPanels_Item) (kbapi.KibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaEsqlControl, error) {
			return raw.AsKibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaEsqlControl()
		},
		BuildPanel: func() kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeEsqlControl {
			return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeEsqlControl{
				Grid: kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid{X: 0, Y: 0},
				Type: kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeEsqlControlTypeEsqlControl,
			}
		},
		PopulateFromAPI: func(_ context.Context, pm *models.PanelModel, prior *models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeEsqlControl) diag.Diagnostics {
			PopulateFromAPI(pm, prior, panel.Config)
			return nil
		},
		BuildConfig: BuildConfig,
		AfterRemapToGroup: func(group *kbapi.KibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaEsqlControl) {
			group.Type = kbapi.KibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaEsqlControlTypeEsqlControl
		},
		FromGroup: func(item *kbapi.DashboardPinnedPanels_Item, group kbapi.KibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaEsqlControl) error {
			return item.FromKibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaEsqlControl(group)
		},
		ParseErrSummary:     "Failed to parse pinned ES|QL control",
		RemapFromErrSummary: "Failed to remap pinned ES|QL control from API",
		RemapToErrSummary:   "Failed to remap pinned ES|QL control",
		FromGroupErrSummary: "Failed to build pinned ES|QL control payload",
	}
}
