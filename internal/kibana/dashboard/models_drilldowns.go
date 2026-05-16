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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Backwards-compatible drilldown entry points (implementation is in lenscommon).

func fromAPI(ctx context.Context, api *[]kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item) (models.DrilldownsModel, diag.Diagnostics) {
	return lenscommon.LensDashboardAppDrilldownsFromAPI(ctx, api)
}

func toAPI(items models.DrilldownsModel) (*[]kbapi.KbnDashboardPanelTypeLensDashboardApp_Config_1_Drilldowns_Item, diag.Diagnostics) {
	return lenscommon.LensDashboardAppDrilldownsToAPI(items)
}

func drilldownsFromVisByRefAPI(ctx context.Context, api *[]kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item) (models.DrilldownsModel, diag.Diagnostics) {
	return lenscommon.DrilldownsFromVisByRefAPI(ctx, api)
}

func drilldownsToVisByRefAPI(items models.DrilldownsModel) (*[]kbapi.KbnDashboardPanelTypeVis_Config_1_Drilldowns_Item, diag.Diagnostics) {
	return lenscommon.DrilldownsToVisByRefAPI(items)
}

func explicitEmptyDrilldowns() models.DrilldownsModel {
	return lenscommon.ExplicitEmptyDrilldowns()
}
