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

package maintenancewindow

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func updateMaintenanceWindow(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[Model]) (entitycore.KibanaWriteResult[Model], diag.Diagnostics) {
	return entitycore.SimpleKibanaUpdate[Model, kbapi.PatchMaintenanceWindowIdJSONRequestBody, struct{}](
		func(plan Model, ctx context.Context, _ string) (kbapi.PatchMaintenanceWindowIdJSONRequestBody, diag.Diagnostics) {
			return plan.toAPIUpdateRequest(ctx)
		},
		func(ctx context.Context, oapiClient *kibanaoapi.Client, spaceID, writeID string, body kbapi.PatchMaintenanceWindowIdJSONRequestBody) (*struct{}, diag.Diagnostics) {
			return nil, kibanaoapi.UpdateMaintenanceWindow(ctx, oapiClient, spaceID, writeID, body)
		},
		func(plan *Model, _ context.Context, spaceID string, _ *struct{}) diag.Diagnostics {
			plan.ID = types.StringValue(req.WriteID)
			plan.SpaceID = types.StringValue(spaceID)
			return nil
		},
	)(ctx, client, req)
}
