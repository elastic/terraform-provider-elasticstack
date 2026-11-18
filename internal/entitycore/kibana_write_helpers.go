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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RequireNonNilKibanaWriteResponse appends the standard "Failed to <verb>
// <noun>" / "API returned empty response" error diagnostic when resp is nil.
// This wording and nil-check repeats verbatim across the Kibana list and
// exception-list Create/Update callbacks (securitylist, securitylistitem,
// securityexceptionlist, securityexceptionitem). It reports whether resp was
// nil so callers can return immediately:
//
//	createdList, d := kibanaoapi.CreateList(ctx, oapiClient, req.SpaceID, *createReq)
//	diags.Append(d...)
//	if diags.HasError() {
//	    return entitycore.KibanaWriteResult[Model]{}, diags
//	}
//	if entitycore.RequireNonNilKibanaWriteResponse(&diags, createdList, "create", "security list") {
//	    return entitycore.KibanaWriteResult[Model]{}, diags
//	}
func RequireNonNilKibanaWriteResponse[T any](diags *diag.Diagnostics, resp *T, verb, noun string) bool {
	if resp != nil {
		return false
	}
	diags.AddError("Failed to "+verb+" "+noun, "API returned empty response")
	return true
}

// KibanaResourceID builds the `<spaceID>/<resourceID>` composite Terraform
// resource ID shared by space-scoped Kibana list and exception-list
// resources, wrapping [clients.CompositeID] with the [types.StringValue]
// conversion every Create callback needs.
func KibanaResourceID(spaceID, resourceID string) types.String {
	return types.StringValue((&clients.CompositeID{
		ClusterID:  spaceID,
		ResourceID: resourceID,
	}).String())
}

// SimpleKibanaCreate returns a [KibanaWriteFunc] for the common plan -> body
// -> API create -> populate shape shared by simple Kibana write callbacks:
// convert the plan to Body via toBody, call apiCreate for req.SpaceID, then
// apply populate to the plan (typically setting SpaceID, plus any
// response-derived fields) before wrapping it in [KibanaWriteResult]. Use for
// resources whose create callback needs nothing beyond that shape; resources
// with extra steps (version gates, generated-ID capture) should wrap this in
// a small function instead of hand-rolling the tail, for example:
//
//	func createAgent(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[agentModel]) (entitycore.KibanaWriteResult[agentModel], diag.Diagnostics) {
//		supportsSkillIDs, diags := client.EnforceMinVersion(ctx, agentbuilder.MinExtendedAPIVersion)
//		if diags.HasError() {
//			return entitycore.KibanaWriteResult[agentModel]{}, diags
//		}
//		return entitycore.SimpleKibanaCreate[agentModel, kbapi.PostAgentBuilderAgentsJSONRequestBody, models.Agent](
//			func(plan agentModel, ctx context.Context) (kbapi.PostAgentBuilderAgentsJSONRequestBody, diag.Diagnostics) {
//				return plan.toAPICreateModel(ctx, supportsSkillIDs)
//			},
//			kibanaoapi.CreateAgent,
//			setAgentWriteSpaceID,
//		)(ctx, client, req)
//	}
//
// toBody is typically a toAPICreateModel method expression:
//
//	Skill.toAPICreateModel
func SimpleKibanaCreate[T KibanaResourceModel, Body any, R any](
	toBody func(plan T, ctx context.Context) (Body, diag.Diagnostics),
	apiCreate func(ctx context.Context, client *kibanaoapi.Client, spaceID string, body Body) (*R, diag.Diagnostics),
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

		resp, d := apiCreate(ctx, client.GetKibanaOapiClient(), req.SpaceID, body)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		diags.Append(populate(&plan, ctx, req.SpaceID, resp)...)
		return KibanaWriteResult[T]{Model: plan}, diags
	}
}

// SimpleKibanaUpdate is [SimpleKibanaCreate]'s counterpart for Update: it
// calls apiUpdate with req.SpaceID and req.WriteID instead of apiCreate with
// req.SpaceID alone, and it passes req.WriteID through to toBody as well,
// since update bodies commonly need to embed the resource ID being updated.
// See [SimpleKibanaCreate] for the shared shape and usage pattern.
func SimpleKibanaUpdate[T KibanaResourceModel, Body any, R any](
	toBody func(plan T, ctx context.Context, writeID string) (Body, diag.Diagnostics),
	apiUpdate func(ctx context.Context, client *kibanaoapi.Client, spaceID, writeID string, body Body) (*R, diag.Diagnostics),
	populate func(plan *T, ctx context.Context, spaceID string, resp *R) diag.Diagnostics,
) KibanaWriteFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, req KibanaWriteRequest[T]) (KibanaWriteResult[T], diag.Diagnostics) {
		plan := req.Plan
		var diags diag.Diagnostics

		body, d := toBody(plan, ctx, req.WriteID)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		resp, d := apiUpdate(ctx, client.GetKibanaOapiClient(), req.SpaceID, req.WriteID, body)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		diags.Append(populate(&plan, ctx, req.SpaceID, resp)...)
		return KibanaWriteResult[T]{Model: plan}, diags
	}
}
