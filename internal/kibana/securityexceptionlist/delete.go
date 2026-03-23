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

package securityexceptionlist

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *ExceptionListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ExceptionListModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Parse composite ID to get space_id and resource_id
	compID, compIDDiags := clients.CompositeIDFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(compIDDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete by resource ID from composite ID
	id := compID.ResourceID
	params := &kbapi.DeleteExceptionListParams{
		Id: &id,
	}

	// Include namespace_type if known (required for agnostic lists)
	if state.NamespaceType.ValueString() != "" {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(state.NamespaceType.ValueString())
		params.NamespaceType = &nsType
	}

	diags = kibanaoapi.DeleteExceptionList(ctx, client, compID.ClusterID, params)

	resp.Diagnostics.Append(diags...)
}
