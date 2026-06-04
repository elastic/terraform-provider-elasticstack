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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateDashboard(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[models.DashboardModel],
) (entitycore.KibanaWriteResult[models.DashboardModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	planModel := req.Plan

	kibanaClient := client.GetKibanaOapiClient()

	// Resolve identity from plan (composite ID)
	composite, compositeDiags := clients.CompositeIDFromStr(planModel.ID.ValueString())
	diags.Append(compositeDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[models.DashboardModel]{}, diags
	}

	dashboardID := composite.ResourceID
	spaceID := composite.ClusterID

	apiReq := dashboardToAPIUpdateRequest(ctx, &planModel, &diags)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[models.DashboardModel]{}, diags
	}

	_, updateDiags := kibanaoapi.UpdateDashboard(ctx, kibanaClient, spaceID, dashboardID, apiReq)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[models.DashboardModel]{}, diags
	}

	return entitycore.KibanaWriteResult[models.DashboardModel]{Model: planModel}, diags
}
