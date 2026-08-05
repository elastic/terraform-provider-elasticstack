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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
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
