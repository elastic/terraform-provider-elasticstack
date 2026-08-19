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

package entitycore

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// SimpleFleetRead returns a [KibanaReadFunc] for the common
// fetch-then-populate shape shared by simple Fleet resources: call apiGet for
// resourceID in spaceID, treat a nil result as "not found", and otherwise
// apply populate to the prior model. Use for resources whose read callback
// needs nothing beyond that shape; resources with extra logic (retries,
// derived fields not sourced from the API payload) should keep a hand-written
// read callback. populate is typically a populateFromAPI method expression:
//
//	Read: entitycore.SimpleFleetRead[proxyModel, kbapi.FleetProxyItem](fleetclient.GetProxy, (*proxyModel).populateFromAPIPtr),
func SimpleFleetRead[T KibanaResourceModel, R any](
	apiGet func(ctx context.Context, client *fleetclient.Client, spaceID, resourceID string) (*R, diag.Diagnostics),
	populate func(model *T, ctx context.Context, spaceID string, data *R) diag.Diagnostics,
) KibanaReadFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, prior T) (T, bool, diag.Diagnostics) {
		var diags diag.Diagnostics

		fleetClient := client.GetFleetClient()

		data, d := apiGet(ctx, fleetClient, spaceID, resourceID)
		diags.Append(d...)
		if diags.HasError() {
			return prior, false, diags
		}

		if data == nil {
			return prior, false, diags
		}

		diags.Append(populate(&prior, ctx, spaceID, data)...)
		return prior, true, diags
	}
}
