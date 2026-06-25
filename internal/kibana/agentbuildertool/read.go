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

package agentbuildertool

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readTool(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, prior toolModel) (toolModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// SpaceID must be set before populateFromAPI so it can build the composite ID.
	prior.SpaceID = types.StringValue(spaceID)

	oapiClient := client.GetKibanaOapiClient()

	tool, d := kibanaoapi.GetTool(ctx, oapiClient, spaceID, resourceID)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}

	if tool == nil {
		return prior, false, diags
	}

	diags.Append(prior.populateFromAPI(ctx, tool)...)
	return prior, true, diags
}
