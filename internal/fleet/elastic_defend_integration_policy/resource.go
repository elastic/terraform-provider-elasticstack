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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/resourcecore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newElasticDefendIntegrationPolicyResource()
	_ resource.ResourceWithConfigure   = newElasticDefendIntegrationPolicyResource()
	_ resource.ResourceWithImportState = newElasticDefendIntegrationPolicyResource()
)

type elasticDefendIntegrationPolicyResource struct {
	*resourcecore.Core
}

func newElasticDefendIntegrationPolicyResource() *elasticDefendIntegrationPolicyResource {
	return &elasticDefendIntegrationPolicyResource{
		Core: resourcecore.New(resourcecore.ComponentFleet, "elastic_defend_integration_policy"),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newElasticDefendIntegrationPolicyResource()
}

func (r *elasticDefendIntegrationPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema()
}

func (r *elasticDefendIntegrationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Accept both "<space_id>/<policy_id>" and plain "<policy_id>" import IDs.
	// Pass the raw import ID through to the "id" attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Parse the composite ID to seed policy_id and, if present, space_ids.
	compID, diags := clients.CompositeIDFromStrFw(req.ID)
	if diags.HasError() {
		// Plain policy_id: no space prefix
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), req.ID)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), compID.ResourceID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_ids"), []string{compID.ClusterID})...)
	}
}
