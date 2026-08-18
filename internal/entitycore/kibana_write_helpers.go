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

// simpleKibanaWrite implements the build-request -> apiCall ->
// RequireNonNilKibanaWriteResponse -> populate shape shared by
// [SimpleKibanaCreate] and [SimpleKibanaUpdate].
func simpleKibanaWrite[T KibanaResourceModel, B, R any](
	verb, noun string,
	buildRequest func(ctx context.Context, req KibanaWriteRequest[T]) (*B, diag.Diagnostics),
	apiCall func(ctx context.Context, client *kibanaoapi.Client, spaceID string, body B) (*R, diag.Diagnostics),
	populate func(model *T, spaceID string, resp *R),
) KibanaWriteFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, req KibanaWriteRequest[T]) (KibanaWriteResult[T], diag.Diagnostics) {
		var diags diag.Diagnostics

		body, d := buildRequest(ctx, req)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		resp, d := apiCall(ctx, client.GetKibanaOapiClient(), req.SpaceID, *body)
		diags.Append(d...)
		if diags.HasError() {
			return KibanaWriteResult[T]{}, diags
		}

		if RequireNonNilKibanaWriteResponse(&diags, resp, verb, noun) {
			return KibanaWriteResult[T]{}, diags
		}

		m := req.Plan
		if populate != nil {
			populate(&m, req.SpaceID, resp)
		}

		return KibanaWriteResult[T]{Model: m}, diags
	}
}

// SimpleKibanaCreate returns a [KibanaWriteFunc] for the common
// build-request -> apiCreate -> RequireNonNilKibanaWriteResponse -> populate
// shape shared by the Kibana list and exception-list Create callbacks
// (securitylist, securitylistitem, securityexceptionlist,
// security_exception_item). buildRequest constructs the API request body
// from the plan (and, for Update-style rebuilds, req.WriteID); populate
// copies response fields such as server-assigned IDs back onto the model and
// may be nil when the response carries nothing the model needs. noun is the
// resource name used in the standard "Failed to create <noun>" diagnostic:
//
//	Create: entitycore.SimpleKibanaCreate(
//	    "security list",
//	    func(_ context.Context, req entitycore.KibanaWriteRequest[Model]) (*kbapi.CreateListJSONRequestBody, diag.Diagnostics) {
//	        return req.Plan.toCreateRequest()
//	    },
//	    kibanaoapi.CreateList,
//	    func(m *Model, spaceID string, resp *kbapi.SecurityListsAPIList) {
//	        m.ListID = typeutils.StringishValue(resp.Id)
//	        m.ID = entitycore.KibanaResourceID(spaceID, resp.Id)
//	    },
//	),
func SimpleKibanaCreate[T KibanaResourceModel, B, R any](
	noun string,
	buildRequest func(ctx context.Context, req KibanaWriteRequest[T]) (*B, diag.Diagnostics),
	apiCreate func(ctx context.Context, client *kibanaoapi.Client, spaceID string, body B) (*R, diag.Diagnostics),
	populate func(model *T, spaceID string, resp *R),
) KibanaWriteFunc[T] {
	return simpleKibanaWrite("create", noun, buildRequest, apiCreate, populate)
}

// SimpleKibanaUpdate mirrors [SimpleKibanaCreate] for Update callbacks.
func SimpleKibanaUpdate[T KibanaResourceModel, B, R any](
	noun string,
	buildRequest func(ctx context.Context, req KibanaWriteRequest[T]) (*B, diag.Diagnostics),
	apiUpdate func(ctx context.Context, client *kibanaoapi.Client, spaceID string, body B) (*R, diag.Diagnostics),
	populate func(model *T, spaceID string, resp *R),
) KibanaWriteFunc[T] {
	return simpleKibanaWrite("update", noun, buildRequest, apiUpdate, populate)
}
