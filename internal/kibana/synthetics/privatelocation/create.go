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

package privatelocation

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func createPrivateLocation(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[Model]) (entitycore.KibanaWriteResult[Model], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	oapiClient, d := client.GetKibanaOapiClient()
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	// Preserve the planned geo values before the API call. The Kibana API stores geo
	// coordinates as float32 and returns float32-precision values on read (e.g.
	// 42.42 → 42.41999816894531). If we blindly set state from the API response, the
	// state value differs from the plan value, which Terraform rejects as an
	// inconsistent result. We use the planned values (from config) so state matches
	// the plan. The Float32PrecisionType custom type handles subsequent semantic
	// equality checks so that subsequent plans detect no diff.
	plannedGeo := plan.Geo

	body := plan.toCreateBody()
	result, dg := kibanaoapi.CreatePrivateLocation(ctx, oapiClient, req.SpaceID, body)
	diags.Append(dg...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	model := modelFromAPI(*result, req.SpaceID)
	model.KibanaConnection = plan.KibanaConnection
	model.Geo = plannedGeo

	return entitycore.KibanaWriteResult[Model]{Model: model}, diags
}
