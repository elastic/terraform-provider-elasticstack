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

package datastreamlifecycle

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// writeDataStreamLifecycle handles both Create and Update for the data
// stream lifecycle resource. The PUT API is idempotent, so the same body
// serves both lifecycle methods.
func writeDataStreamLifecycle(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[tfModel]) (entitycore.WriteResult[tfModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	resourceID := req.WriteID

	id, sdkDiags := client.ID(ctx, resourceID)
	if sdkDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	plan.ID = types.StringValue(id.String())

	apiModel, d := plan.toAPIModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	diags.Append(elasticsearch.PutDataStreamLifecycle(ctx, client, resourceID, plan.ExpandWildcards.ValueString(), apiModel)...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	return entitycore.WriteResult[tfModel]{Model: plan}, diags
}
