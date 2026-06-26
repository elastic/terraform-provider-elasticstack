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

package tag

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateTag(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tagModel],
) (entitycore.KibanaWriteResult[tagModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()

	existing, getDiags := kibanaoapi.GetTag(ctx, oapiClient, req.SpaceID, req.WriteID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tagModel]{}, diags
	}

	if existing != nil {
		if managedDiags := checkManagedTag(existing); managedDiags.HasError() {
			return entitycore.KibanaWriteResult[tagModel]{}, managedDiags
		}
	}

	body := plan.toUpdateAPIModel(req.Prior)
	_, upsertDiags := kibanaoapi.UpsertTag(ctx, oapiClient, req.SpaceID, req.WriteID, body)
	diags.Append(upsertDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tagModel]{}, diags
	}

	plan.setCompositeIdentity(req.SpaceID, req.WriteID)
	return entitycore.KibanaWriteResult[tagModel]{Model: plan}, diags
}
