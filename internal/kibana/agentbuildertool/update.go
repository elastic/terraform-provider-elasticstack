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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func updateTool(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[toolModel]) (entitycore.KibanaWriteResult[toolModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	plan.SpaceID = types.StringValue(req.SpaceID)

	body, d := plan.toAPIUpdateModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[toolModel]{}, diags
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError(err.Error(), "")
		return entitycore.KibanaWriteResult[toolModel]{}, diags
	}

	_, d = kibanaoapi.UpdateTool(ctx, oapiClient, req.SpaceID, req.WriteID, body)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[toolModel]{}, diags
	}

	return entitycore.KibanaWriteResult[toolModel]{Model: plan}, diags
}
