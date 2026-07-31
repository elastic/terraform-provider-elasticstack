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

package elasticdefendintegrationpolicy

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = newElasticDefendIntegrationPolicyResource()
	_ resource.ResourceWithConfigure   = newElasticDefendIntegrationPolicyResource()
	_ resource.ResourceWithImportState = newElasticDefendIntegrationPolicyResource()
)

var (
	MinVersionPolicyIDs = version.Must(version.NewVersion("8.15.0"))
)

// checkAgentPolicyIDsVersionSupport returns an error diagnostic when
// agent_policy_ids is configured but the connected Kibana version predates
// MinVersionPolicyIDs. Both Create and Update must reject the request in
// that case before sending anything to the API.
func checkAgentPolicyIDsVersionSupport(ctx context.Context, client *clients.KibanaScopedClient, agentPolicyIDs types.List) diag.Diagnostics {
	if agentPolicyIDs.IsNull() || agentPolicyIDs.IsUnknown() {
		return nil
	}

	var diags diag.Diagnostics
	supported, d := client.EnforceMinVersion(ctx, MinVersionPolicyIDs)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	if !supported {
		diags.AddError(
			"Unsupported Elasticsearch version",
			fmt.Sprintf("agent_policy_ids requires Elastic Stack >= %s", MinVersionPolicyIDs.String()),
		)
	}
	return diags
}

type elasticDefendIntegrationPolicyResource struct {
	*entitycore.ResourceBase
	*entitycore.SpaceImporter
}

func newElasticDefendIntegrationPolicyResource() *elasticDefendIntegrationPolicyResource {
	return &elasticDefendIntegrationPolicyResource{
		ResourceBase:  entitycore.NewResourceBase(entitycore.ComponentFleet, "elastic_defend_integration_policy"),
		SpaceImporter: entitycore.NewSpaceImporter(path.Root("policy_id")),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newElasticDefendIntegrationPolicyResource()
}

func (r *elasticDefendIntegrationPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema()
}
