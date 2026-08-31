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

package proxy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateProxy(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[proxyModel]) (entitycore.KibanaWriteResult[proxyModel], diag.Diagnostics) {
	return entitycore.SimpleFleetUpdate[proxyModel, kbapi.PutFleetProxiesItemidJSONRequestBody, kbapi.FleetProxyItem](
		func(plan proxyModel, _ context.Context) (kbapi.PutFleetProxiesItemidJSONRequestBody, diag.Diagnostics) {
			return plan.toAPIUpdateModel()
		},
		// fleetclient.UpdateProxy takes (spaceID, proxyID) rather than the
		// (writeID, spaceID) order SimpleFleetUpdate calls apiUpdate with.
		func(ctx context.Context, client *fleetclient.Client, writeID, spaceID string, body kbapi.PutFleetProxiesItemidJSONRequestBody) (*kbapi.FleetProxyItem, diag.Diagnostics) {
			return fleetclient.UpdateProxy(ctx, client, spaceID, writeID, body)
		},
		func(plan *proxyModel, _ context.Context, spaceID string, resp *kbapi.FleetProxyItem) diag.Diagnostics {
			return plan.populateFromAPI(spaceID, *resp)
		},
	)(ctx, client, req)
}
