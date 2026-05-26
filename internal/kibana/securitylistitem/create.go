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

package securitylistitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createSecurityListItem(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[Model]) (entitycore.KibanaWriteResult[Model], diag.Diagnostics) {
	m := req.Plan
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()

	createReq, d := m.toAPICreateModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	createdListItem, d := kibanaoapi.CreateListItem(ctx, oapiClient, req.SpaceID, *createReq)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	if createdListItem == nil {
		diags.AddError("Failed to create security list item", "API returned empty response")
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	m.ListItemID = typeutils.StringishValue(createdListItem.Id)
	m.ID = types.StringValue((&clients.CompositeID{
		ClusterID:  req.SpaceID,
		ResourceID: createdListItem.Id,
	}).String())

	return entitycore.KibanaWriteResult[Model]{Model: m}, diags
}
