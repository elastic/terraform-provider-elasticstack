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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func readDashboard(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID string,
	spaceID string,
	model models.DashboardModel,
) (models.DashboardModel, bool, diag.Diagnostics) {
	kibanaClient := client.GetKibanaOapiClient()
	getResp, getDiags := kibanaoapi.GetDashboard(ctx, kibanaClient, spaceID, resourceID)
	var diags diag.Diagnostics
	diags.Append(getDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	if getResp == nil {
		return model, false, diags
	}

	_, unwrapDiags := diagutil.UnwrapJSON200(getResp.JSON200, "dashboard")
	diags.Append(unwrapDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	diags.Append(dashboardPopulateFromAPI(ctx, &model, getResp, resourceID, spaceID)...)
	return model, true, diags
}

func postReadDashboard(
	ctx context.Context,
	req entitycore.KibanaPostReadRequest[models.DashboardModel],
) (models.DashboardModel, diag.Diagnostics) {
	alignDashboardStateFromPlanPanels(req.Prior.Panels, req.State.Panels)
	suppressReadTopLevelPanelsWhenPlanEmpty(req.Prior.Panels, &req.State)
	alignDashboardStateFromPlanSections(ctx, req.Prior.Sections, req.State.Sections)
	alignDashboardStateFromPlanPinnedPanels(ctx, req.Prior.PinnedPanels, req.State.PinnedPanels)
	return req.State, nil
}
