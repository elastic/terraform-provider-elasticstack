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

package fleet

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ResolveReadDeleteContext resolves the Fleet client and operational space ID shared by the
// Read and Delete preambles of policy-shaped Fleet resources (agent policy, integration
// policy, elastic defend integration policy): it resolves the Kibana/Fleet client for the
// given connection block and reads the operational space ID from state via
// [GetOperationalSpaceFromState].
//
// Callers still handle their own state.Get, policy ID extraction, and resource-specific API
// call/result handling; only the client and space resolution steps are shared here.
func ResolveReadDeleteContext(ctx context.Context, clientFactory *clients.ProviderClientFactory, kibanaConnection types.List, state tfsdk.State) (*fleetclient.Client, string, diag.Diagnostics) {
	var diags diag.Diagnostics

	client, clientDiags := clientFactory.GetKibanaClient(ctx, kibanaConnection)
	diags.Append(clientDiags...)
	if diags.HasError() {
		return nil, "", diags
	}

	spaceID, spaceDiags := GetOperationalSpaceFromState(ctx, state)
	diags.Append(spaceDiags...)
	if diags.HasError() {
		return nil, "", diags
	}

	return client.GetFleetClient(), spaceID, diags
}
