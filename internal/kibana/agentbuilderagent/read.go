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

package agentbuilderagent

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// agentReader abstracts the API calls made by readAgentWith so the read flow
// can be unit tested without a Kibana server.
type agentReader struct {
	getAgent              func(ctx context.Context) (*models.Agent, diag.Diagnostics)
	supportsAccessControl func(ctx context.Context) (bool, diag.Diagnostics)
	getAgentAccessControl func(ctx context.Context) (*models.AgentAccessControl, diag.Diagnostics)
}

// readAgent reads the agent and, on servers that support it (9.5.0+), its
// access_control from the dedicated sub-resource endpoint.
func readAgent(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, prior agentModel) (agentModel, bool, diag.Diagnostics) {
	oapi := client.GetKibanaOapiClient()
	return readAgentWith(ctx, agentReader{
		getAgent: func(ctx context.Context) (*models.Agent, diag.Diagnostics) {
			return kibanaoapi.GetAgent(ctx, oapi, spaceID, resourceID)
		},
		supportsAccessControl: func(ctx context.Context) (bool, diag.Diagnostics) {
			return client.EnforceMinVersion(ctx, agentbuilder.MinAccessControlAPIVersion)
		},
		getAgentAccessControl: func(ctx context.Context) (*models.AgentAccessControl, diag.Diagnostics) {
			return kibanaoapi.GetAgentAccessControl(ctx, oapi, spaceID, resourceID)
		},
	}, spaceID, prior)
}

// readAgentWith performs the read flow. A missing agent returns before any
// access-control call is attempted. Once the agent exists, the sub-resource
// call must succeed: any failure, including a 404, is surfaced as an error
// rather than silently nulling access_control.
func readAgentWith(ctx context.Context, r agentReader, spaceID string, prior agentModel) (agentModel, bool, diag.Diagnostics) {
	data, diags := r.getAgent(ctx)
	if diags.HasError() || data == nil {
		return prior, false, diags
	}

	diags.Append(prior.populateFromAPI(ctx, spaceID, data)...)

	supported, d := r.supportsAccessControl(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}
	if !supported {
		diags.Append(prior.populateAccessControl(ctx, nil)...)
		return prior, true, diags
	}

	accessControl, d := r.getAgentAccessControl(ctx)
	diags.Append(d...)
	if !diags.HasError() && accessControl == nil {
		diags.AddError("Agent access control not found",
			"The agent exists but its access-control sub-resource returned HTTP 404.")
	}
	if diags.HasError() {
		return prior, false, diags
	}
	diags.Append(prior.populateAccessControl(ctx, accessControl)...)
	return prior, true, diags
}
