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

package osquerypack

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createOsqueryPack(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[osqueryPackModel],
) (entitycore.KibanaWriteResult[osqueryPackModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	body, bodyDiags := plan.toCreateRequestBody(ctx)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[osqueryPackModel]{}, diags
	}

	oapiClient := client.GetKibanaOapiClient()

	detail, createDiags := kibanaoapi.CreateOsqueryPack(ctx, oapiClient, req.SpaceID, body)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[osqueryPackModel]{}, diags
	}

	plan.PackID = types.StringValue(detail.SavedObjectId)
	plan.SpaceID = types.StringValue(req.SpaceID)
	plan.ID = types.StringValue((&clients.CompositeID{
		ClusterID:  req.SpaceID,
		ResourceID: detail.SavedObjectId,
	}).String())

	return entitycore.KibanaWriteResult[osqueryPackModel]{Model: plan}, diags
}
