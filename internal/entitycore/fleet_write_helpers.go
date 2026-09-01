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

// SimpleFleetCreate is [SimpleKibanaCreate]'s counterpart for Fleet write
// callbacks that go through client.GetFleetClient() rather than
// client.GetKibanaOapiClient(). See [SimpleKibanaCreate] for the shared
// plan -> body -> API create -> populate shape and usage pattern.
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

// SimpleFleetUpdate is [SimpleKibanaUpdate]'s counterpart for Fleet write
// callbacks. It resolves the effective spaceID from req.Prior when present,
// falling back to req.SpaceID otherwise, before passing it to both toBody and
// apiUpdate. This centralizes the "spaceID := req.SpaceID; if req.Prior !=
// nil { spaceID = req.Prior.GetSpaceID().ValueString() }" fallback that
// space_ids-set resources (serverhost, output) need: on Update, the request
// must be routed using the space the resource actually lives in (the prior
// state), not the plan's possibly-changed space_ids. Resources with a plain,
// RequiresReplace-guarded space_id (for example proxy) are unaffected, since
// Prior's spaceID and the resolved plan spaceID are always equal there.
func SimpleFleetUpdate[T KibanaResourceModel, Body any, R any](
	toBody func(plan T, ctx context.Context, writeID string) (Body, diag.Diagnostics),
	apiUpdate func(ctx context.Context, client *fleetclient.Client, spaceID, writeID string, body Body) (*R, diag.Diagnostics),
	populate func(plan *T, ctx context.Context, spaceID string, resp *R) diag.Diagnostics,
) KibanaWriteFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, req KibanaWriteRequest[T]) (KibanaWriteResult[T], diag.Diagnostics) {
		plan := req.Plan
		var diags diag.Diagnostics

		spaceID := req.SpaceID
		if req.Prior != nil {
			spaceID = (*req.Prior).GetSpaceID().ValueString()
		}

		body, d := toBody(plan, ctx, req.WriteID)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		resp, d := apiUpdate(ctx, client.GetFleetClient(), spaceID, req.WriteID, body)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		diags.Append(populate(&plan, ctx, spaceID, resp)...)
		return KibanaWriteResult[T]{Model: plan}, diags
	}
}
