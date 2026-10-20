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
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// readAgent reads the agent and, on servers that support it (9.5.0+), its
// access_control from the dedicated sub-resource endpoint. A missing agent
// returns before any access-control call is attempted.
func readAgent(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, prior agentModel) (agentModel, bool, diag.Diagnostics) {
	oapi := client.GetKibanaOapiClient()
	data, diags := kibanaoapi.GetAgent(ctx, oapi, spaceID, resourceID)
	if diags.HasError() || data == nil {
		return prior, false, diags
	}

	diags.Append(prior.populateFromAPI(ctx, spaceID, data)...)

	supported, d := client.EnforceMinVersion(ctx, agentbuilder.MinAccessControlAPIVersion)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}
	if !supported {
		diags.Append(prior.populateAccessControl(nil)...)
		return prior, true, diags
	}

	accessControl, d := kibanaoapi.GetAgentAccessControl(ctx, oapi, spaceID, resourceID)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}
	diags.Append(prior.populateAccessControl(accessControl)...)
	return prior, true, diags
}
