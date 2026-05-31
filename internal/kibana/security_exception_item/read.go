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

package securityexceptionitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *ExceptionItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ExceptionItemModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetKibanaClient(ctx, state.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse composite ID to get space_id and resource_id
	compID, compIDDiags := clients.CompositeIDFromStr(state.ID.ValueString())
	resp.Diagnostics.Append(compIDDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SpaceID = types.StringValue(compID.ClusterID)

	oapiClient := client.GetKibanaOapiClient()

	readResp, diags := kibanaoapi.GetExceptionListItem(ctx, oapiClient, state.SpaceID.ValueString(), exceptionItemReadParams(compID.ResourceID, state.NamespaceType))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If namespace_type was not known (e.g., during import) and the item was not found,
	// try reading with namespace_type=agnostic
	if readResp == nil && !typeutils.IsKnown(state.NamespaceType) {
		agnosticNsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType("agnostic")
		id := compID.ResourceID
		retryParams := &kbapi.ReadExceptionListItemParams{
			Id:            &id,
			NamespaceType: &agnosticNsType,
		}
		readResp, diags = kibanaoapi.GetExceptionListItem(ctx, oapiClient, state.SpaceID.ValueString(), retryParams)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if readResp == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with response using model method
	diags = state.fromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func exceptionItemReadParams(itemID string, namespaceType types.String) *kbapi.ReadExceptionListItemParams {
	id := itemID
	params := &kbapi.ReadExceptionListItemParams{Id: &id}
	if typeutils.IsKnown(namespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(namespaceType.ValueString())
		params.NamespaceType = &nsType
	}
	return params
}

// refreshExceptionItemState reads an item after create/update so state matches the API.
// The second return value is true when the resource should be removed from state (not found).
func refreshExceptionItemState(
	ctx context.Context,
	oapiClient *kibanaoapi.Client,
	spaceID string,
	plan ExceptionItemModel,
	itemID string,
) (ExceptionItemModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	readResp, d := kibanaoapi.GetExceptionListItem(ctx, oapiClient, spaceID, exceptionItemReadParams(itemID, plan.NamespaceType))
	diags.Append(d...)
	if diags.HasError() {
		return plan, false, diags
	}
	if readResp == nil {
		return plan, true, diags
	}

	diags.Append(plan.fromAPI(ctx, readResp)...)
	return plan, false, diags
}
