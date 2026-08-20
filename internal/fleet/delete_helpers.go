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

// ResolveFleetClientAndSpace resolves the Fleet client and operational space ID
// shared by the space-scoped Delete methods of the Fleet policy resources that
// override resource.Resource.Delete directly (agentpolicy,
// integration_policy, elastic_defend_integration_policy) instead of using the
// entitycore.KibanaResource[T] Delete callback: resolve the Kibana-scoped
// client from kibanaConnection, extract its Fleet client, then resolve the
// operational space from Terraform state via GetOperationalSpaceFromState.
//
// state must be the Delete request's state (not plan): GetOperationalSpaceFromState
// needs the space where the resource currently exists, which is why these
// resources reach for req.State via a raw Delete override rather than the
// KibanaResource[T] callback signature.
func ResolveFleetClientAndSpace(ctx context.Context, state tfsdk.State, clientFactory *clients.ProviderClientFactory, kibanaConnection types.List) (*fleetclient.Client, string, diag.Diagnostics) {
	client, diags := clientFactory.GetKibanaClient(ctx, kibanaConnection)
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
