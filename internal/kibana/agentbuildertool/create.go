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

func createTool(ctx context.Context, client *clients.KibanaScopedClient, spaceID string, plan toolModel) (toolModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	plan.SpaceID = types.StringValue(spaceID)

	body, d := plan.toAPICreateModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return plan, diags
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError(err.Error(), "")
		return plan, diags
	}

	created, d := kibanaoapi.CreateTool(ctx, oapiClient, spaceID, body)
	diags.Append(d...)
	if diags.HasError() {
		return plan, diags
	}

	tool, d := kibanaoapi.GetTool(ctx, oapiClient, spaceID, created.ID)
	diags.Append(d...)
	if diags.HasError() {
		return plan, diags
	}

	diags.Append(plan.populateFromAPI(ctx, tool)...)
	return plan, diags
}
