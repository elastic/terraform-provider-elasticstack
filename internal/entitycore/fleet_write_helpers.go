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

// SimpleFleetCreate is [SimpleKibanaCreate]'s counterpart for simple Fleet
// write callbacks backed by *fleetclient.Client instead of
// *kibanaoapi.Client: convert the plan to Body via toBody, call apiCreate for
// req.SpaceID, then apply populate to the plan before wrapping it in
// [KibanaWriteResult]. See [SimpleKibanaCreate] for the shared shape and
// usage pattern, for example:
//
//	func createProxy(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[proxyModel]) (entitycore.KibanaWriteResult[proxyModel], diag.Diagnostics) {
//		return entitycore.SimpleFleetCreate[proxyModel, kbapi.PostFleetProxiesJSONRequestBody, kbapi.FleetProxyItem](
//			func(plan proxyModel, _ context.Context) (kbapi.PostFleetProxiesJSONRequestBody, diag.Diagnostics) {
//				return plan.toAPICreateModel()
//			},
//			fleetclient.CreateProxy,
//			func(plan *proxyModel, _ context.Context, spaceID string, resp *kbapi.FleetProxyItem) diag.Diagnostics {
//				return plan.populateFromAPI(spaceID, *resp)
//			},
//		)(ctx, client, req)
//	}
func SimpleFleetCreate[T KibanaResourceModel, Body any, R any](
	toBody func(plan T, ctx context.Context) (Body, diag.Diagnostics),
	apiCreate func(ctx context.Context, client *fleetclient.Client, spaceID string, body Body) (*R, diag.Diagnostics),
	populate func(plan *T, ctx context.Context, spaceID string, resp *R) diag.Diagnostics,
) KibanaWriteFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, req KibanaWriteRequest[T]) (KibanaWriteResult[T], diag.Diagnostics) {
		plan := req.Plan
		var diags diag.Diagnostics

		body, d := toBody(plan, ctx)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		resp, d := apiCreate(ctx, client.GetFleetClient(), req.SpaceID, body)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		diags.Append(populate(&plan, ctx, req.SpaceID, resp)...)
		return KibanaWriteResult[T]{Model: plan}, diags
	}
}

// SimpleFleetUpdate is [SimpleFleetCreate]'s counterpart for Update. It
// mirrors [SimpleKibanaUpdate] with one difference: the space passed to
// apiUpdate and populate is resolved via [ResolveUpdateSpaceID] rather than
// req.SpaceID directly, so Update always targets the space the resource
// currently lives in. apiUpdate takes the write ID before the space ID,
// matching most Fleet client Update functions; adapt call sites whose
// client function orders these the other way with a small wrapper closure.
func SimpleFleetUpdate[T KibanaResourceModel, Body any, R any](
	toBody func(plan T, ctx context.Context) (Body, diag.Diagnostics),
	apiUpdate func(ctx context.Context, client *fleetclient.Client, writeID, spaceID string, body Body) (*R, diag.Diagnostics),
	populate func(plan *T, ctx context.Context, spaceID string, resp *R) diag.Diagnostics,
) KibanaWriteFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, req KibanaWriteRequest[T]) (KibanaWriteResult[T], diag.Diagnostics) {
		plan := req.Plan
		var diags diag.Diagnostics

		body, d := toBody(plan, ctx)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		spaceID := ResolveUpdateSpaceID(req)
		resp, d := apiUpdate(ctx, client.GetFleetClient(), req.WriteID, spaceID, body)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		diags.Append(populate(&plan, ctx, spaceID, resp)...)
		return KibanaWriteResult[T]{Model: plan}, diags
	}
}
