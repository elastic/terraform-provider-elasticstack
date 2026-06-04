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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createDashboard(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[models.DashboardModel],
) (entitycore.KibanaWriteResult[models.DashboardModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	planModel := req.Plan

	kibanaClient := client.GetKibanaOapiClient()
	spaceID := planModel.SpaceID.ValueString()

	apiReq := dashboardToAPICreateRequest(ctx, &planModel, &diags)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[models.DashboardModel]{}, diags
	}

	createResp, createDiags := kibanaoapi.CreateDashboard(ctx, kibanaClient, spaceID, apiReq)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[models.DashboardModel]{}, diags
	}

	if createResp.JSON201 == nil {
		diags.AddError("Dashboard create returned no body", "expected 201 response with dashboard id")
		return entitycore.KibanaWriteResult[models.DashboardModel]{}, diags
	}

	compID := clients.CompositeID{
		ClusterID:  spaceID,
		ResourceID: createResp.JSON201.Id,
	}
	planModel.ID = types.StringValue(compID.String())
	planModel.DashboardID = types.StringValue(createResp.JSON201.Id)

	return entitycore.KibanaWriteResult[models.DashboardModel]{Model: planModel}, diags
}
